package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	foundationdao "foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"io"
	"judge/config"
	gojudge "judge/go-judge"
	"log/slog"
	"meta/cron"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	metapath "meta/meta-path"
	metastring "meta/meta-string"
	"meta/routine"
	"meta/singleton"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// 需要保证只有一个goroutine在处理判题数据
type judgeDataDownloadEntry struct {
	mu  sync.Mutex
	ref int32
}

type JudgeService struct {
	runningTasks      atomic.Int32
	judgeDataDownload sync.Map
}

var singletonJudgeService = singleton.Singleton[JudgeService]{}

func GetJudgeService() *JudgeService {
	return singletonJudgeService.GetInstance(
		func() *JudgeService {
			return &JudgeService{}
		},
	)
}

func (s *JudgeService) Start() error {
	c := cron.NewWithSeconds()
	_, err := c.AddFunc(
		"* * * * * ?", func() {
			// 每秒运行一次任务
			err := s.handleStart()
			if err != nil {
				metapanic.ProcessError(err)
				return
			}
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "error adding function to cron")
	}

	c.Start()

	return nil
}

func (s *JudgeService) handleStart() error {
	maxJob := config.GetConfig().MaxJob
	if int(s.runningTasks.Load()) >= maxJob {
		return nil
	}
	ctx := context.Background()
	jobs, err := foundationdao.GetJudgeJobDao().GetJudgeJobListPendingJudge(ctx, maxJob)
	if err != nil {
		return metaerror.Wrap(err, "failed to get judge job list")
	}
	s.runningTasks.Add(int32(len(jobs)))
	for _, job := range jobs {
		routine.SafeGo(fmt.Sprintf("RunningJudgeJob_%d", job.Id), func() error {
			defer s.runningTasks.Add(-1)
			slog.Info(fmt.Sprintf("JudgeTask_%d start", job.Id))
			err = s.startJudgeTask(job)
			slog.Info(fmt.Sprintf("JudgeTask_%d end", job.Id))
			if err != nil {
				err := foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(ctx, job.Id, foundationjudge.JudgeStatusJudgeFail)
				if err != nil {
					metapanic.ProcessError(err)
				}
				return err
			}
			return nil
		})
	}
	return nil
}

func (s *JudgeService) startJudgeTask(job *foundationmodel.JudgeJob) error {
	ctx := context.Background()

	err := foundationdao.GetJudgeJobDao().StartProcessJudgeJob(ctx, job.Id, config.GetConfig().Judger)
	if err != nil {
		return metaerror.Wrap(err, "failed to start process judge job")
	}
	problem, err := foundationdao.GetProblemDao().GetProblem(ctx, job.ProblemId)
	if err != nil {
		return metaerror.Wrap(err, "failed to get problem")
	}
	err = s.updateJudgeData(ctx, problem.Id, problem.JudgeMd5)
	if err != nil {
		return metaerror.Wrap(err, "failed to update judge data")
	}
	execFileId, extraMessage, compileStatus, err := s.compileCode(job)
	if extraMessage != "" {
		markErr := foundationdao.GetJudgeJobDao().MarkJudgeJobCompileMessage(ctx, job.Id, extraMessage)
		if markErr != nil {
			metapanic.ProcessError(markErr)
		}
	}
	if err != nil {
		return err
	}
	if compileStatus != foundationjudge.JudgeStatusAccept {
		err := foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(ctx, job.Id, compileStatus)
		if err != nil {
			metapanic.ProcessError(err)
		}
		return nil
	}
	slog.Info("compile code success", "job", job.Id, "execFileId", execFileId)
	err = foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(ctx, job.Id, foundationjudge.JudgeStatusRunning)
	if err != nil {
		metapanic.ProcessError(err)
	}
	err = s.runJudgeTask(ctx, job, problem.TimeLimit, problem.MemoryLimit, execFileId)
	return err
}

func (s *JudgeService) updateJudgeData(ctx context.Context, problemId string, md5 string) error {
	val, _ := s.judgeDataDownload.LoadOrStore(problemId, &judgeDataDownloadEntry{})
	e := val.(*judgeDataDownloadEntry)
	atomic.AddInt32(&e.ref, 1)
	defer func() {
		if atomic.AddInt32(&e.ref, -1) == 0 {
			s.judgeDataDownload.Delete(problemId)
		}
	}()
	e.mu.Lock()
	defer e.mu.Unlock()
	judgeMd5FilePath := path.Join(".judge_data", problemId, "md5.txt")
	content, err := metastring.GetStringFromOpenFile(judgeMd5FilePath)
	if err != nil || strings.TrimSpace(content) != strings.TrimSpace(md5) {
		return s.downloadJudgeData(ctx, problemId)
	}
	return nil
}

func (s *JudgeService) downloadJudgeData(ctx context.Context, problemId string) error {

	slog.Info("downloading judge data", "problemId", problemId)

	// 删除旧的判题数据
	judgeDataDir := path.Join(".judge_data", problemId)
	err := os.RemoveAll(judgeDataDir)

	// 初始化 R2 连接（这里用 AWS SDK）
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("auto"),                           // R2一般写 auto
		Endpoint:         aws.String(config.GetConfig().JudgeData.Url), // 替换成你的 R2 Endpoint
		S3ForcePathStyle: aws.Bool(true),                               // R2要求这个必须 true
		Credentials: credentials.NewStaticCredentials(config.GetConfig().JudgeData.Key,
			config.GetConfig().JudgeData.Secret,
			config.GetConfig().JudgeData.Token),
	})
	if err != nil {
		return metaerror.Wrap(err, "failed to create session")
	}

	s3Client := s3.New(sess)

	// 1. 列出 problemId 目录下的所有对象
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("didaoj-judge"),
		Prefix: aws.String(problemId + "/"), // 确保带 `/`，只列出这个目录下的
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var downloadErr error

	err = s3Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			wg.Add(1)
			go func(obj *s3.Object) {
				defer wg.Done()
				localPath := path.Join(".judge_data", *obj.Key)
				err := s.downloadObject(ctx, s3Client, "didaoj-judge", *obj.Key, localPath)
				if err != nil {
					mu.Lock()
					if downloadErr == nil {
						downloadErr = err
					}
					mu.Unlock()
				}
			}(obj)
		}
		return true
	})
	if err != nil {
		return metaerror.Wrap(err, "failed to list objects")
	}

	// 等待所有下载完成
	wg.Wait()

	// 如果有任何错误，返回
	if downloadErr != nil {
		return downloadErr
	}

	return nil
}

// 单独抽一个下载单个对象的方法
func (s *JudgeService) downloadObject(ctx context.Context, s3Client *s3.S3, bucket, key string, localPath string) error {
	getObjInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	result, err := s3Client.GetObjectWithContext(ctx, getObjInput)
	if err != nil {
		return fmt.Errorf("failed to get object %s: %w", key, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(result.Body)
	err = os.MkdirAll(filepath.Dir(localPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", localPath, err)
	}
	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", localPath, err)
	}
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(outFile)
	_, err = io.Copy(outFile, result.Body)
	if err != nil {
		return metaerror.Wrap(err, "failed to save object %s", key)
	}
	return nil
}

func (s *JudgeService) compileCode(job *foundationmodel.JudgeJob) (map[string]string, string, foundationjudge.JudgeStatus, error) {
	slog.Info("compile code", "job", job.Id)
	runUrl := metahttp.UrlJoin(config.GetConfig().GoJudgeUrl, "run")

	var args []string
	var copyIns map[string]interface{}
	var copyOutCached []string

	switch job.Language {
	case foundationjudge.JudgeLanguageC:
		args = []string{"gcc", "-fno-asm", "-fmax-errors=10", "-Wall", "--static", "-DONLINE_JUDGE", "-o", "a", "a.c", "-lm"}
		copyIns = map[string]interface{}{
			"a.c": map[string]interface{}{
				"content": job.Code,
			},
		}
		copyOutCached = []string{"a"}
		break
	case foundationjudge.JudgeLanguageCpp:
		args = []string{"g++", "-fno-asm", "-fmax-errors=10", "-Wall", "--static", "-DONLINE_JUDGE", "-o", "a", "a.cc"}
		copyIns = map[string]interface{}{
			"a.cc": map[string]interface{}{
				"content": job.Code,
			},
		}
		copyOutCached = []string{"a"}
		break
	case foundationjudge.JudgeLanguageJava:
		args = []string{"javac", "-J-Xms128m", "-J-Xmx512m", "-encoding", "UTF-8", "Main.java"}
		copyIns = map[string]interface{}{
			"Main.java": map[string]interface{}{
				"content": job.Code,
			},
		}
		copyOutCached = []string{"Main.class"}
		break
	case foundationjudge.JudgeLanguagePython:
		args = []string{"python3", "-c", "import py_compile; py_compile.compile(r'a.py')"}
		copyIns = map[string]interface{}{
			"a.py": map[string]interface{}{
				"content": job.Code,
			},
		}
		copyOutCached = nil
	default:
		return nil, "compile failed, language not support.",
			foundationjudge.JudgeStatusJudgeFail,
			metaerror.New("language not support: %d",
				job.Language,
			)
	}

	// 准备请求数据
	data := map[string]interface{}{
		"cmd": []map[string]interface{}{
			{
				"args": args,
				"env":  []string{"PATH=/usr/bin:/bin"},
				"files": []map[string]interface{}{
					{"content": ""},
					{"name": "stdout", "max": 10240},
					{"name": "stderr", "max": 10240},
				},
				"cpuLimit":      10000000000,
				"memoryLimit":   1048576 * 500, // 500MB
				"procLimit":     50,
				"copyIn":        copyIns,
				"copyOut":       []string{"stdout", "stderr"},
				"copyOutCached": copyOutCached,
			},
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, "compile failed, system error.", foundationjudge.JudgeStatusJudgeFail, err
	}
	resp, err := http.Post(runUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "compile failed, upload file error.", foundationjudge.JudgeStatusJudgeFail, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, "compile failed, upload file response error.", foundationjudge.JudgeStatusJudgeFail, metaerror.New("unexpected status code: %d", resp.StatusCode)
	}
	var responseDataList []struct {
		Status gojudge.Status `json:"status"`
		Error  string         `json:"error"`
		Files  struct {
			Stderr string `json:"stderr"`
			Stdout string `json:"stdout"`
		} `json:"files"`
		FileIds map[string]string `json:"fileIds"`
	}
	err = json.NewDecoder(resp.Body).Decode(&responseDataList)
	if err != nil {
		return nil, fmt.Sprintf("compile failed, upload file response parse error."), foundationjudge.JudgeStatusJudgeFail, metaerror.Wrap(err, "failed to decode response")
	}
	if len(responseDataList) != 1 {
		return nil, "compile failed, compile response data error.", foundationjudge.JudgeStatusJudgeFail, metaerror.New("unexpected response length: %d", len(responseDataList))
	}
	responseData := responseDataList[0]
	errorMessage := responseData.Error
	if responseData.Files.Stderr != "" {
		if errorMessage != "" {
			errorMessage += "\n"
		}
		errorMessage += responseData.Files.Stderr
	}
	if responseData.Files.Stdout != "" {
		if errorMessage != "" {
			errorMessage += "\n"
		}
		errorMessage += responseData.Files.Stdout
	}
	if responseData.Status != gojudge.StatusAccepted {
		if responseData.Status != gojudge.StatusNonzeroExit {
			return nil, errorMessage, foundationjudge.JudgeStatusCLE, nil
		} else {
			return nil, errorMessage, foundationjudge.JudgeStatusCE, nil
		}
	}
	return responseData.FileIds, errorMessage, foundationjudge.JudgeStatusAccept, nil
}

func (s *JudgeService) runJudgeTask(ctx context.Context, job *foundationmodel.JudgeJob, timeLimit int, memoryLimit int, execFileId map[string]string) error {
	problemId := job.ProblemId
	val, _ := s.judgeDataDownload.LoadOrStore(problemId, &judgeDataDownloadEntry{})
	e := val.(*judgeDataDownloadEntry)
	atomic.AddInt32(&e.ref, 1)
	defer func() {
		if atomic.AddInt32(&e.ref, -1) == 0 {
			s.judgeDataDownload.Delete(problemId)
		}
	}()
	e.mu.Lock()
	defer e.mu.Unlock()

	judgeDataDir := path.Join(".judge_data", problemId)
	files, err := os.ReadDir(judgeDataDir)
	if err != nil {
		return metaerror.Wrap(err, "failed to read judge data dir")
	}
	// TODO获取rule.yaml文件

	enableRule := false

	hasInFiles := make(map[string]bool)
	var Files []string

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".in") {
			hasInFiles[metapath.GetBaseName(file.Name())] = true
		} else if strings.HasSuffix(file.Name(), ".out") {
			Files = append(Files, metapath.GetBaseName(file.Name()))
		}
	}
	taskCount := 0
	sort.Slice(Files, func(i, j int) bool {
		return Files[i] < Files[j]
	})
	for _, file := range Files {
		if !hasInFiles[file] {
			continue
		}
		taskCount++
	}
	err = foundationdao.GetJudgeJobDao().MarkJudgeJobTaskTotal(ctx, job.Id, taskCount)
	if err != nil {
		metapanic.ProcessError(err)
	}

	acTask := 0
	finalStatus := foundationjudge.JudgeStatusAccept
	sumTime := 0
	sumMemory := 0

	cpuLimit := timeLimit * 1000000
	memoryLimit = memoryLimit * 1024
	if job.Language == foundationjudge.JudgeLanguageJava {
		cpuLimit = cpuLimit + 2000*1000000
		memoryLimit = memoryLimit + 1024*1024*64
	}

	for _, file := range Files {
		task := foundationmodel.NewJudgeTaskBuilder().
			TaskId(file).
			Status(foundationjudge.JudgeStatusJudgeFail).
			Time(0).
			Memory(0).
			Content("").
			WaHint("").
			Build()

		runUrl := metahttp.UrlJoin(config.GetConfig().GoJudgeUrl, "run")
		inContent, err := metastring.GetStringFromOpenFile(path.Join(judgeDataDir, file+".in"))
		if err != nil {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return err
		}

		var args []string
		var copyIns map[string]interface{}
		switch job.Language {
		case foundationjudge.JudgeLanguageC:
			fallthrough
		case foundationjudge.JudgeLanguageCpp:
			args = []string{"a"}
			fileId, ok := execFileId["a"]
			if !ok {
				markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
				if markErr != nil {
					metapanic.ProcessError(markErr)
				}
				return metaerror.New("fileId not found")
			}
			copyIns = map[string]interface{}{
				"a": map[string]interface{}{
					"fileId": fileId,
				},
			}
		case foundationjudge.JudgeLanguageJava:
			args = []string{"java", "-Djava.security.manager", "-Djava.security.policy=./java.policy", "Main"}
			fileId, ok := execFileId["Main.class"]
			if !ok {
				markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
				if markErr != nil {
					metapanic.ProcessError(markErr)
				}
				return metaerror.New("fileId not found")
			}
			copyIns = map[string]interface{}{
				"Main.class": map[string]interface{}{
					"fileId": fileId,
				},
			}
			break
		case foundationjudge.JudgeLanguagePython:
			args = []string{"python3", "a.py"}
			copyIns = map[string]interface{}{
				"a.py": map[string]interface{}{
					"content": job.Code,
				},
			}
		default:
			return metaerror.New("language not support: %d", job.Language)
		}

		data := map[string]interface{}{
			"cmd": []map[string]interface{}{
				{
					"args": args,
					"env":  []string{"PATH=/usr/bin:/bin"},
					"files": []map[string]interface{}{
						{"content": inContent},
						{"name": "stdout", "max": 10240},
						{"name": "stderr", "max": 10240},
					},
					"cpuLimit":    cpuLimit,
					"memoryLimit": memoryLimit,
					"procLimit":   50,
					"copyIn":      copyIns,
				},
			},
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return metaerror.Wrap(err)
		}
		resp, err := http.Post(runUrl, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return metaerror.Wrap(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				metapanic.ProcessError(err)
			}
		}(resp.Body)
		if resp.StatusCode != http.StatusOK {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return metaerror.New("unexpected status code: %d", resp.StatusCode)
		}
		var responseDataList []struct {
			Status gojudge.Status `json:"status"`
			Files  struct {
				Stderr string `json:"stderr"`
				Stdout string `json:"stdout"`
			} `json:"files"`
			Time   int `json:"time"`
			Memory int `json:"memory"`
		}
		err = json.NewDecoder(resp.Body).Decode(&responseDataList)
		if err != nil {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return metaerror.Wrap(err, "failed to decode response")
		}
		if len(responseDataList) != 1 {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return metaerror.New("unexpected response length: %d", len(responseDataList))
		}
		responseData := responseDataList[0]
		if responseData.Status != gojudge.StatusAccepted {
			switch responseData.Status {
			case gojudge.StatusSignalled:
				task.Status = foundationjudge.JudgeStatusRE
			case gojudge.StatusNonzeroExit:
				task.Status = foundationjudge.JudgeStatusRE
			case gojudge.StatusInternalError:
				task.Status = foundationjudge.JudgeStatusJudgeFail
			case gojudge.StatusOutputLimit:
				task.Status = foundationjudge.JudgeStatusOLE
			case gojudge.StatusFileError:
				task.Status = foundationjudge.JudgeStatusOLE
			case gojudge.StatusMemoryLimit:
				task.Status = foundationjudge.JudgeStatusMLE
			case gojudge.StatusTimeLimit:
				task.Status = foundationjudge.JudgeStatusTLE
			default:
				task.Status = foundationjudge.JudgeStatusJudgeFail
			}
			finalStatus = foundationjudge.GetFinalStatus(finalStatus, task.Status)
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			continue
		}
		rightOutContent, err := metastring.GetStringFromOpenFile(path.Join(judgeDataDir, file+".out"))
		if err != nil {
			return err
		}

		task.Time = responseData.Time
		task.Memory = responseData.Memory

		sumTime += responseData.Time
		sumMemory += responseData.Memory

		ansContent := responseData.Files.Stdout

		// 移除所有空行和每行前后的空格
		outContentMyPe := strings.Fields(rightOutContent)
		ansContentMyPe := strings.Fields(ansContent)
		WaHint := ""
		for i := 0; i < len(outContentMyPe); i++ {
			if i < len(ansContentMyPe) {
				if outContentMyPe[i] != ansContentMyPe[i] {
					WaHint = fmt.Sprintf("%s != %s", outContentMyPe[i], ansContentMyPe[i])
				}
			} else {
				WaHint = fmt.Sprintf("%s not found", outContentMyPe[i])
				break
			}
		}
		if WaHint != "" {
			task.Status = foundationjudge.JudgeStatusWA
			task.WaHint = WaHint
		} else {
			if rightOutContent == ansContent {
				acTask++
				task.Status = foundationjudge.JudgeStatusAccept
			} else {
				task.Status = foundationjudge.JudgeStatusPE
			}
		}
		finalStatus = foundationjudge.GetFinalStatus(finalStatus, task.Status)
		err = foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
		if err != nil {
			return err
		}
	}
	score := 0
	// 更新任务状态
	if !enableRule {
		if acTask == taskCount {
			score = 100
		} else {
			score = int(float64(acTask) / float64(taskCount) * 100)
		}
	}

	finalTime := sumTime / taskCount
	finalMemory := sumMemory / taskCount

	err = foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeFinalStatus(ctx, job.Id, finalStatus, score, finalTime, finalMemory)

	return err
}
