package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	foundationdao "foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"gopkg.in/yaml.v3"
	"io"
	"judge/config"
	gojudge "judge/go-judge"
	"log/slog"
	cfr2 "meta/cf-r2"
	"meta/cron"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	metamath "meta/meta-math"
	metapanic "meta/meta-panic"
	metapath "meta/meta-path"
	metastring "meta/meta-string"
	"meta/routine"
	"meta/singleton"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// 需要保证只有一个goroutine在处理判题数据
type judgeDataDownloadEntry struct {
	mu  sync.Mutex
	ref int32
}

type JudgeService struct {
	runningTasks atomic.Int32

	// 有些时候同一个问题只能有一个逻辑去处理
	problemMutexMap sync.Map

	// 题目号对应的特判程序文件ID
	specialFileIds map[string]string
	configFileIds  map[string]string
}

var singletonJudgeService = singleton.Singleton[JudgeService]{}

func GetJudgeService() *JudgeService {
	return singletonJudgeService.GetInstance(
		func() *JudgeService {
			s := &JudgeService{}
			return s
		},
	)
}

func (s *JudgeService) Start() error {

	err := s.cleanGoJudge()
	if err != nil {
		return err
	}

	err = s.uploadFiles()
	if err != nil {
		return metaerror.Wrap(err, "error uploading files")
	}

	c := cron.NewWithSeconds()
	_, err = c.AddFunc(
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

func (s *JudgeService) getSpecialFileId(problemId string) string {
	if s.specialFileIds == nil {
		return ""
	}
	fileId, ok := s.specialFileIds[problemId]
	if !ok {
		return ""
	}
	return fileId
}

func (s *JudgeService) GetConfigFileId(fileKey string) string {
	if s.configFileIds == nil {
		return ""
	}
	fileId, ok := s.configFileIds[fileKey]
	if !ok {
		return ""
	}
	return fileId
}

func (s *JudgeService) cleanGoJudge() error {
	goJudgeUrl := config.GetConfig().GoJudge.Url
	goJudgeFileUrl := metahttp.UrlJoin(goJudgeUrl, "file")
	fileListResp, err := http.Get(goJudgeFileUrl)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(fileListResp.Body)
	var fileList map[string]string
	err = json.NewDecoder(fileListResp.Body).Decode(&fileList)
	if err != nil {
		return metaerror.Wrap(err, "failed to decode file list")
	}
	client := &http.Client{}
	for fileId, _ := range fileList {
		deleteUrl := metahttp.UrlJoin(goJudgeUrl, "file", fileId)
		request, err := http.NewRequest(http.MethodDelete, deleteUrl, nil)
		if err != nil {
			return err
		}
		_, err = client.Do(request)
		if err != nil {
			return metaerror.Wrap(err, "failed to delete file")
		}
	}

	return nil
}

func (s *JudgeService) uploadFiles() error {
	client := &http.Client{}
	filesConfig := config.GetFilesConfig()

	for fileKey, filePath := range filesConfig {
		fileId, err := s.uploadFile(client, filePath)
		if err != nil {
			return metaerror.Wrap(err, "failed to upload file: %s", filePath)
		}
		slog.Info("file uploaded successfully", "fileId", fileId)
		if s.configFileIds == nil {
			s.configFileIds = make(map[string]string)
		}
		s.configFileIds[fileKey] = *fileId
	}

	return nil
}

func (s *JudgeService) uploadFile(client *http.Client, filePath string) (*string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to open file: %s", filePath)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close file: %s", filePath))
		}
	}(file)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to create form file part")
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to copy file content")
	}
	err = writer.Close()
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to close multipart writer")
	}
	goJudgeUrl := config.GetConfig().GoJudge.Url
	uploadUrl := metahttp.UrlJoin(goJudgeUrl, "file")
	request, err := http.NewRequest(http.MethodPost, uploadUrl, body)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to create upload request for file: %s", filePath)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := client.Do(request)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to upload file: %s", filePath)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close response body for file: %s", filePath))
		}
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		return nil, metaerror.New("failed to upload file: %s, status code: %d", filePath, response.StatusCode)
	}
	var fileId string
	err = json.NewDecoder(response.Body).Decode(&fileId)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to decode upload response for file: %s", filePath)
	}
	return &fileId, nil
}

func (s *JudgeService) handleStart() error {

	// 如果上报状态报错，停止判题
	if GetStatusService().IsReportError() {
		return nil
	}

	maxJob := config.GetConfig().MaxJob
	if int(s.runningTasks.Load()) >= maxJob {
		return nil
	}
	ctx := context.Background()
	jobs, err := foundationdao.GetJudgeJobDao().RequestJudgeJobListPendingJudge(ctx, maxJob, config.GetConfig().Judger.Key)
	if err != nil {
		return metaerror.Wrap(err, "failed to get judge job list")
	}
	jobsCount := len(jobs)
	if jobsCount == 0 {
		return nil
	}

	slog.Info("get judge job list", "count", jobsCount, "maxJob", maxJob)

	s.runningTasks.Add(int32(jobsCount))

	for _, job := range jobs {
		routine.SafeGo(fmt.Sprintf("RunningJudgeJob_%d", job.Id), func() error {
			defer func() {
				slog.Info(fmt.Sprintf("JudgeTask_%d end", job.Id))
				s.runningTasks.Add(-1)
			}()
			slog.Info(fmt.Sprintf("JudgeTask_%d start", job.Id))
			err = s.startJudgeTask(job)
			if err != nil {
				markErr := foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(ctx, job.Id, foundationjudge.JudgeStatusJudgeFail)
				if markErr != nil {
					metapanic.ProcessError(markErr)
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

	err := foundationdao.GetJudgeJobDao().StartProcessJudgeJob(ctx, job.Id, config.GetConfig().Judger.Key)
	if err != nil {
		return metaerror.Wrap(err, "failed to start process judge job")
	}
	problem, err := foundationdao.GetProblemDao().GetProblem(ctx, job.ProblemId)
	if err != nil {
		return metaerror.Wrap(err, "failed to get problem")
	}
	if problem == nil {
		return metaerror.New("problem not found: %s", job.ProblemId)
	}
	if problem.JudgeMd5 == nil {
		return metaerror.New("problem judge md5 is nil: %s", job.ProblemId)
	}
	err = s.updateJudgeData(ctx, problem.Id, *problem.JudgeMd5)
	if err != nil {
		return metaerror.Wrap(err, "failed to update judge data")
	}

	var execFileIds map[string]string
	var extraMessage string
	var compileStatus foundationjudge.JudgeStatus
	if foundationjudge.IsLanguageNeedCompile(job.Language) {
		execFileIds, extraMessage, compileStatus, err = s.compileCode(job)
		if extraMessage != "" {
			markErr := foundationdao.GetJudgeJobDao().MarkJudgeJobCompileMessage(ctx, job.Id, extraMessage)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
		}
		defer func() {
			client := &http.Client{}
			for _, fileId := range execFileIds {
				goJudgeUrl := config.GetConfig().GoJudge.Url
				deleteUrl := metahttp.UrlJoin(goJudgeUrl, "file", fileId)
				request, err := http.NewRequest(http.MethodDelete, deleteUrl, nil)
				if err != nil {
					metapanic.ProcessError(metaerror.Wrap(err, "failed to create delete request"))
					continue
				}
				_, err = client.Do(request)
				if err != nil {
					metapanic.ProcessError(metaerror.Wrap(err, "failed to delete file"))
					continue
				}
			}
		}()
		if err != nil {
			return err
		}
		if compileStatus != foundationjudge.JudgeStatusAC {
			err := foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(ctx, job.Id, compileStatus)
			if err != nil {
				metapanic.ProcessError(err)
			}
			return nil
		}
		slog.Info("compile code success", "job", job.Id, "execFileIds", execFileIds)
		err = foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(ctx, job.Id, foundationjudge.JudgeStatusRunning)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}
	err = s.runJudgeJob(ctx, job, problem, execFileIds)
	return err
}

func (s *JudgeService) updateJudgeData(ctx context.Context, problemId string, md5 string) error {
	val, _ := s.problemMutexMap.LoadOrStore(problemId, &judgeDataDownloadEntry{})
	e := val.(*judgeDataDownloadEntry)
	atomic.AddInt32(&e.ref, 1)
	defer func() {
		if atomic.AddInt32(&e.ref, -1) == 0 {
			s.problemMutexMap.Delete(problemId)
		}
	}()
	e.mu.Lock()
	defer e.mu.Unlock()
	judgeMd5FilePath := path.Join(".judge_data", problemId, md5)
	// 判断 judgeMd5FilePath 是否存在
	_, err := os.Stat(judgeMd5FilePath)
	if err == nil {
		// 文件存在，直接返回
		return nil
	} else if !os.IsNotExist(err) {
		// 其他错误，返回报错
		return metaerror.Wrap(err, "failed to stat judge md5 file")
	}
	return s.downloadJudgeData(ctx, problemId, md5)
}

func (s *JudgeService) downloadJudgeData(ctx context.Context, problemId string, md5 string) error {

	slog.Info("downloading judge data", "problemId", problemId)

	r2Client := cfr2.GetSubsystem().GetClient("judge-data")
	if r2Client == nil {
		return metaerror.New("r2Client is nil")
	}

	// 删除旧的判题数据
	judgeDataDir := path.Join(".judge_data", problemId)
	err := os.RemoveAll(judgeDataDir)

	// 1. 列出 problemId 目录下的所有对象
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("didaoj-judge"),
		Prefix: aws.String(problemId + "/"), // 确保带 `/`，只列出这个目录下的
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var downloadErr error

	err = r2Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			wg.Add(1)
			go func(obj *s3.Object) {
				defer wg.Done()
				localPath := path.Join(".judge_data", *obj.Key)
				err := s.downloadObject(ctx, r2Client, "didaoj-judge", *obj.Key, localPath)
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

func (s *JudgeService) compileSpecialJudge(job *foundationmodel.JudgeJob, md5 string, jobConfig *foundationjudge.JudgeJobConfig) (string, error) {
	problemId := job.ProblemId

	specialFileId := s.getSpecialFileId(problemId)
	if specialFileId != "" {
		return specialFileId, nil
	}

	val, _ := s.problemMutexMap.LoadOrStore(problemId, &judgeDataDownloadEntry{})
	e := val.(*judgeDataDownloadEntry)
	atomic.AddInt32(&e.ref, 1)
	defer func() {
		if atomic.AddInt32(&e.ref, -1) == 0 {
			s.problemMutexMap.Delete(problemId)
		}
	}()
	e.mu.Lock()
	defer e.mu.Unlock()

	runUrl := metahttp.UrlJoin(config.GetConfig().GoJudge.Url, "run")

	language := foundationjudge.GetLanguageByKey(jobConfig.SpecialJudge.Language)
	if !foundationjudge.IsValidJudgeLanguage(int(language)) {
		return "", metaerror.New("invalid language: %s", jobConfig.SpecialJudge.Language)
	}

	// 考虑编译机性能影响，暂时仅允许部分语言
	if !foundationjudge.IsValidSpecialJudgeLanguage(language) {
		return "", metaerror.New("language %s not valid special language", jobConfig.SpecialJudge.Language)
	}

	judgeDataDir := path.Join(".judge_data", problemId, md5)

	codeFilePath := filepath.Join(judgeDataDir, jobConfig.SpecialJudge.Source)
	codeContent, err := metastring.GetStringFromOpenFile(codeFilePath)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to read special judge code file")
	}

	execFileIds, extraMessage, compileStatus, err := foundationjudge.CompileCode(strconv.Itoa(job.Id), runUrl, language, codeContent, nil)
	if extraMessage != "" {
		slog.Warn("judge compile", "extraMessage", extraMessage, "compileStatus", compileStatus)
	}
	if compileStatus != foundationjudge.JudgeStatusAC {
		return "", metaerror.New("compile special judge failed: %s", extraMessage)
	}
	if err != nil {
		return "", metaerror.Wrap(err, "failed to compile special judge")
	}

	var ok bool
	specialFileId, ok = execFileIds["a"]
	if !ok {
		return "", metaerror.New("special judge compile failed, fileId not found")
	}
	if s.specialFileIds == nil {
		s.specialFileIds = make(map[string]string)
	}
	s.specialFileIds[problemId] = specialFileId
	return specialFileId, nil
}

func (s *JudgeService) compileCode(job *foundationmodel.JudgeJob) (map[string]string, string, foundationjudge.JudgeStatus, error) {
	goJudgeUrl := config.GetConfig().GoJudge.Url
	runUrl := metahttp.UrlJoin(goJudgeUrl, "run")
	return foundationjudge.CompileCode(strconv.Itoa(job.Id), runUrl, job.Language, job.Code, s.configFileIds)
}

func (s *JudgeService) runJudgeJob(ctx context.Context, job *foundationmodel.JudgeJob,
	problem *foundationmodel.Problem,
	execFileIds map[string]string) error {
	problemId := job.ProblemId

	timeLimit := problem.TimeLimit
	memoryLimit := problem.MemoryLimit
	md5 := *problem.JudgeMd5

	judgeDataDir := path.Join(".judge_data", problemId, md5)

	taskCount := 0

	var jobConfig foundationjudge.JudgeJobConfig

	// 获取rule.yaml文件并解析
	ruleFilePath := path.Join(judgeDataDir, "rule.yaml")
	yamlFile, err := os.ReadFile(ruleFilePath)
	if err == nil {
		err = yaml.Unmarshal(yamlFile, &jobConfig)
		if err != nil {
			return metaerror.Wrap(err, "Unmarshal config file error")
		}
	}

	if problem.JudgeType == foundationjudge.JudgeTypeSpecial {
		if jobConfig.SpecialJudge == nil {
			specialFiles := map[string]string{
				"spj.c":   "c",
				"spj.cc":  "cpp",
				"spj.cpp": "cpp",
			}
			// 判断是否存在对应文件
			for fileName, language := range specialFiles {
				filePath := path.Join(judgeDataDir, fileName)
				_, err := os.Stat(filePath)
				if err == nil {
					jobConfig.SpecialJudge = &foundationjudge.SpecialJudgeConfig{}
					jobConfig.SpecialJudge.Language = language
					jobConfig.SpecialJudge.Source = fileName
					break
				}
			}
		}
	}

	var specialFileId string
	if jobConfig.SpecialJudge != nil {
		specialFileId, err = s.compileSpecialJudge(job, md5, &jobConfig)
		if err != nil {
			return metaerror.Wrap(err, "failed to compile special judge")
		}
		if specialFileId == "" {
			return metaerror.New("special judge compile failed")
		}
	}

	if len(jobConfig.Tasks) <= 0 {
		// 如果没有rule.yaml文件，则根据文件生成Config信息
		files, err := os.ReadDir(judgeDataDir)
		if err != nil {
			return metaerror.Wrap(err, "failed to read judge data dir")
		}
		var outFileNames []string
		hasInFiles := make(map[string]bool)
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".out") {
				outFileNames = append(outFileNames, metapath.GetBaseName(file.Name()))
			} else if strings.HasSuffix(file.Name(), ".in") {
				hasInFiles[metapath.GetBaseName(file.Name())] = true
			}
		}
		sort.Slice(outFileNames, func(i, j int) bool {
			return outFileNames[i] < outFileNames[j]
		})
		totalScore := 100
		averageScore := totalScore / len(outFileNames)
		for i, file := range outFileNames {
			outFile, err := os.Stat(path.Join(judgeDataDir, file+".out"))
			if err != nil {
				continue
			}
			judgeTaskConfig := &foundationjudge.JudgeTaskConfig{
				Key:      file,
				OutFile:  file + ".out",
				OutLimit: metamath.Max(outFile.Size()*2, 1024),
			}
			if hasInFiles[file] {
				judgeTaskConfig.InFile = file + ".in"
			}
			if i == len(outFileNames)-1 {
				judgeTaskConfig.Score = totalScore - averageScore*(len(outFileNames)-1)
			} else {
				judgeTaskConfig.Score = averageScore
			}
			jobConfig.Tasks = append(jobConfig.Tasks, judgeTaskConfig)
		}
	}

	taskCount = len(jobConfig.Tasks)

	if taskCount == 0 {
		return metaerror.New("no job task found")
	}

	err = foundationdao.GetJudgeJobDao().MarkJudgeJobTaskTotal(ctx, job.Id, taskCount)
	if err != nil {
		metapanic.ProcessError(err)
	}

	finalStatus := foundationjudge.JudgeStatusAC
	sumTime := 0
	sumMemory := 0

	cpuLimit := timeLimit * 1000000
	memoryLimit = memoryLimit * 1024
	if job.Language == foundationjudge.JudgeLanguageJava {
		cpuLimit = cpuLimit + 2000*1000000
		memoryLimit = memoryLimit + 1024*1024*64
	}

	finalScore := 0

	for _, taskConfig := range jobConfig.Tasks {
		finalStatus, sumTime, sumMemory, finalScore, err = s.runJudgeTask(ctx, job, taskConfig, cpuLimit, memoryLimit, sumTime, sumMemory, finalScore, specialFileId, judgeDataDir, execFileIds, finalStatus)
		if err != nil {
			return metaerror.Wrap(err, "failed to run task")
		}
	}

	var finalTime, finalMemory int

	if finalStatus == foundationjudge.JudgeStatusAC ||
		finalStatus == foundationjudge.JudgeStatusWA ||
		finalStatus == foundationjudge.JudgeStatusPE {
		finalTime = sumTime / taskCount
		finalMemory = sumMemory / taskCount
	}

	err = foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeFinalStatus(ctx, job.Id,
		finalStatus,
		problemId,
		job.AuthorId,
		finalScore,
		finalTime,
		finalMemory,
	)

	return err
}

func (s *JudgeService) runJudgeTask(ctx context.Context,
	job *foundationmodel.JudgeJob,
	taskConfig *foundationjudge.JudgeTaskConfig,
	cpuLimit int, memoryLimit int,
	sumTime int, sumMemory int,
	finalScore int,
	specialFileId string,
	judgeDataDir string,
	execFileIds map[string]string,
	finalStatus foundationjudge.JudgeStatus,
) (foundationjudge.JudgeStatus, int, int, int, error) {

	var err error

	key := taskConfig.Key
	task := foundationmodel.NewJudgeTaskBuilder().
		TaskId(key).
		Status(foundationjudge.JudgeStatusJudgeFail).
		Time(0).
		Memory(0).
		Score(0).
		Content("").
		WaHint("").
		Build()

	goJudgeUrl := config.GetConfig().GoJudge.Url
	runUrl := metahttp.UrlJoin(goJudgeUrl, "run")

	var inContent string

	if taskConfig.InFile != "" {
		inContent, err = metastring.GetStringFromOpenFile(path.Join(judgeDataDir, taskConfig.InFile))
		if err != nil {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, err
		}
	}

	var args []string
	var copyIns map[string]interface{}
	switch job.Language {
	case foundationjudge.JudgeLanguageC, foundationjudge.JudgeLanguageCpp,
		foundationjudge.JudgeLanguagePascal, foundationjudge.JudgeLanguageGolang:
		args = []string{"a"}
		fileId, ok := execFileIds["a"]
		if !ok {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.New("fileId not found")
		}
		copyIns = map[string]interface{}{
			"a": map[string]interface{}{
				"fileId": fileId,
			},
		}
	case foundationjudge.JudgeLanguageJava:
		args = []string{"java", "Main"}
		fileId, ok := execFileIds["Main.class"]
		if !ok {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.New("fileId not found")
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
	case foundationjudge.JudgeLanguageLua:
		args = []string{"luajit", "a.lua"}
		copyIns = map[string]interface{}{
			"a.lua": map[string]interface{}{
				"content": job.Code,
			},
		}
	case foundationjudge.JudgeLanguageTypeScript:
		args = []string{"node", "a.js"}
		fileId, ok := execFileIds["a.js"]
		if !ok {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.New("fileId not found")
		}
		copyIns = map[string]interface{}{
			"a.js": map[string]interface{}{
				"fileId": fileId,
			},
		}
	default:
		return finalStatus, sumTime, sumMemory,
			finalScore, metaerror.New("language not support: %d", job.Language)
	}

	data := map[string]interface{}{
		"cmd": []map[string]interface{}{
			{
				"args": args,
				"env":  []string{"PATH=/usr/bin:/bin"},
				"files": []map[string]interface{}{
					{"content": inContent},
					{"name": "stdout", "max": taskConfig.OutLimit},
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
		return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err)
	}
	resp, err := http.Post(runUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err)
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
		return finalStatus, sumTime, sumMemory, finalScore, metaerror.New("unexpected status code: %d", resp.StatusCode)
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
		return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err, "failed to decode response")
	}
	if len(responseDataList) != 1 {
		markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
		if markErr != nil {
			metapanic.ProcessError(markErr)
		}
		return finalStatus, sumTime, sumMemory, finalScore, metaerror.New("unexpected response length: %d", len(responseDataList))
	}
	responseData := responseDataList[0]

	task.Content = metastring.GetTextEllipsis(responseData.Files.Stderr, 1000)

	if responseData.Status != gojudge.StatusAccepted {
		switch responseData.Status {
		case gojudge.StatusSignalled:
			task.Status = foundationjudge.JudgeStatusRE
		case gojudge.StatusNonzeroExit:
			task.Status = foundationjudge.JudgeStatusRE
		case gojudge.StatusInternalError:
			slog.Warn("internal error", "job", job.Id, "responseData", responseData)
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
			slog.Warn("status error", "job", job.Id, "responseData", responseData)
			task.Status = foundationjudge.JudgeStatusJudgeFail
		}
		finalStatus = foundationjudge.GetFinalStatus(finalStatus, task.Status)
		markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
		if markErr != nil {
			metapanic.ProcessError(markErr)
		}
		return finalStatus, sumTime, sumMemory, finalScore, nil
	}
	var rightOutContent string
	if taskConfig.OutFile != "" {
		rightOutContent, err = metastring.GetStringFromOpenFile(path.Join(judgeDataDir, taskConfig.OutFile))
		if err != nil {
			return finalStatus, sumTime, sumMemory, finalScore, err
		}
	}

	task.Time = responseData.Time
	task.Memory = responseData.Memory

	sumTime += responseData.Time
	sumMemory += responseData.Memory

	userAnsContent := responseData.Files.Stdout

	if specialFileId == "" {
		// 移除所有空行和每行前后的空格
		outContentMyPe := strings.Fields(rightOutContent)
		ansContentMyPe := strings.Fields(userAnsContent)
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
			task.WaHint = metastring.GetTextEllipsis(WaHint, 1000)
		} else {
			//各自删除最后的换行符，避免最后的换行与测试数据不同带来没必要的误差
			rightOutContent = strings.TrimSuffix(rightOutContent, "\n")
			userAnsContent = strings.TrimSuffix(userAnsContent, "\n")
			if rightOutContent == userAnsContent {
				task.Score = taskConfig.Score
				finalScore += taskConfig.Score
				task.Status = foundationjudge.JudgeStatusAC
			} else {
				task.Status = foundationjudge.JudgeStatusPE
			}
		}
	} else {
		specialData := map[string]interface{}{
			"cmd": []map[string]interface{}{
				{
					"args": []string{"spj", "test.in", "test.out", "user.out"},
					"env":  []string{"PATH=/usr/bin:/bin"},
					"files": []map[string]interface{}{
						{"content": inContent},
						{"name": "stdout", "max": 10240},
						{"name": "stderr", "max": 10240},
					},
					"cpuLimit":    cpuLimit,
					"memoryLimit": memoryLimit,
					"procLimit":   50,
					"copyIn": map[string]interface{}{
						"spj": map[string]interface{}{
							"fileId": specialFileId,
						},
						"test.in": map[string]interface{}{
							"content": inContent,
						},
						"test.out": map[string]interface{}{
							"content": rightOutContent,
						},
						"user.out": map[string]interface{}{
							"content": userAnsContent,
						},
					},
				},
			},
		}
		specialJsonData, err := json.Marshal(specialData)
		if err != nil {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err)
		}
		specialResp, err := http.Post(runUrl, "application/json", bytes.NewBuffer(specialJsonData))
		if err != nil {
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				metapanic.ProcessError(err)
			}
		}(specialResp.Body)
		if specialResp.StatusCode != http.StatusOK {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.New("unexpected status code: %d", specialResp.StatusCode)
		}
		var specialRespDataList []struct {
			Status     gojudge.Status `json:"status"`
			ExitStatus int            `json:"exitStatus"`
			Files      struct {
				Stderr string `json:"stderr"`
				Stdout string `json:"stdout"`
			} `json:"files"`
			Time   int `json:"time"`
			Memory int `json:"memory"`
		}
		err = json.NewDecoder(specialResp.Body).Decode(&specialRespDataList)
		if err != nil {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err, "failed to decode response")
		}
		if len(specialRespDataList) != 1 {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.New("unexpected response length: %d", len(specialRespDataList))
		}
		specialRespData := specialRespDataList[0]

		if task.Content != "" {
			task.Content = task.Content + "\n"
		}
		task.Content = task.Content + specialRespData.Files.Stderr
		task.WaHint = metastring.GetTextEllipsis(specialRespData.Files.Stdout, 1000)

		if specialRespData.Status == gojudge.StatusAccepted {
			task.Score = taskConfig.Score
			finalScore += taskConfig.Score
			task.Status = foundationjudge.JudgeStatusAC
		} else {
			if specialRespData.Status == gojudge.StatusNonzeroExit {
				switch specialRespData.ExitStatus {
				case int(foundationjudge.SpecialJudgeExitCodeWA):
					task.Status = foundationjudge.JudgeStatusWA
				case int(foundationjudge.SpecialJudgeExitCodePE):
					task.Status = foundationjudge.JudgeStatusPE
				default:
					task.Status = foundationjudge.JudgeStatusRE
				}
			} else {
				task.Status = foundationjudge.JudgeStatusJudgeFail
			}
		}
	}

	finalStatus = foundationjudge.GetFinalStatus(finalStatus, task.Status)
	err = foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, task)
	if err != nil {
		return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err, "failed to add judge job task")
	}
	return finalStatus, sumTime, sumMemory, finalScore, nil
}
