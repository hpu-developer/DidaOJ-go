package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	foundationdao "foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
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
	"meta/retry"
	"meta/routine"
	"meta/singleton"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// 需要保证只有一个goroutine在处理判题数据
type judgeMutexEntry struct {
	mu  sync.Mutex
	ref int32
}

type JudgeService struct {
	requestMutex sync.Mutex
	runningTasks atomic.Int32

	// 防止因重判等情况多次获取到了同一个判题任务（不过多个判题机则靠key来忽略）
	judgeJobMutexMap sync.Map
	// 有些时候同一个问题只能有一个逻辑去处理
	problemMutexMap sync.Map

	// 题目号对应的特判程序文件ID
	specialFileIds map[int]string
	// 配置静态文件标识与文件ID的映射
	configFileIds map[string]string

	goJudgeClient *http.Client
}

var singletonJudgeService = singleton.Singleton[JudgeService]{}

func GetJudgeService() *JudgeService {
	return singletonJudgeService.GetInstance(
		func() *JudgeService {
			s := &JudgeService{}
			s.goJudgeClient = &http.Client{
				Transport: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 100,
					MaxConnsPerHost:     100,
					IdleConnTimeout:     90 * time.Second,
				},
				Timeout: 60 * time.Second, // 请求整体超时
			}
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

func (s *JudgeService) getSpecialFileId(problemId int) string {
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

	_, respBody, err := metahttp.SendRequestRetry(
		s.goJudgeClient,
		"cleanGoJudge",
		6,
		time.Second*10,
		http.MethodGet, goJudgeFileUrl,
		nil,
		nil,
		true,
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to get file list from GoJudge")
	}
	var fileList map[string]string
	err = json.Unmarshal(respBody, &fileList)
	if err != nil {
		return metaerror.Wrap(err, "failed to decode file list")
	}
	for fileId, _ := range fileList {
		deleteUrl := metahttp.UrlJoin(goJudgeUrl, "file", fileId)
		err := foundationjudge.DeleteFile(s.goJudgeClient, "cleanGoJudge", deleteUrl)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *JudgeService) uploadFiles() error {
	filesConfig := config.GetFilesConfig()

	for fileKey, filePath := range filesConfig {
		fileId, err := s.uploadFile(filePath)
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

func (s *JudgeService) uploadFile(filePath string) (*string, error) {
	return foundationjudge.UploadFile(s.goJudgeClient, config.GetConfig().GoJudge.Url, filePath)
}

func (s *JudgeService) handleStart() error {

	// 如果上报状态报错，停止判题
	if GetStatusService().IsReportError() {
		return nil
	}

	// 保证同时只有一个handleStart
	if !s.requestMutex.TryLock() {
		return nil
	}
	defer s.requestMutex.Unlock()

	maxJob := config.GetConfig().MaxJob
	runningCount := int(s.runningTasks.Load())
	if runningCount >= maxJob {
		return nil
	}
	ctx := context.Background()
	jobs, err := foundationdao.GetJudgeJobDao().RequestLocalJudgeJobListPendingJudge(
		ctx,
		maxJob-runningCount,
		config.GetConfig().Judger.Key,
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to get judge job list")
	}
	jobsCount := len(jobs)
	if jobsCount == 0 {
		return nil
	}

	slog.Info("get judge job list", "runningCount", runningCount, "maxJob", maxJob, "count", jobsCount)

	s.runningTasks.Add(int32(jobsCount))

	for _, job := range jobs {
		routine.SafeGo(
			fmt.Sprintf("RunningJudgeJob_%d", job.Id), func() error {
				defer func() {
					slog.Info(fmt.Sprintf("JudgeTask_%d end", job.Id))
					s.runningTasks.Add(-1)
				}()
				val, _ := s.judgeJobMutexMap.LoadOrStore(job.Id, &judgeMutexEntry{})
				e := val.(*judgeMutexEntry)
				atomic.AddInt32(&e.ref, 1)
				defer func() {
					if atomic.AddInt32(&e.ref, -1) == 0 {
						s.judgeJobMutexMap.Delete(job.Id)
					}
				}()
				e.mu.Lock()
				defer e.mu.Unlock()

				slog.Info(fmt.Sprintf("JudgeTask_%d start", job.Id))
				err = s.startJudgeTask(job)
				if err != nil {
					markErr := foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(
						ctx,
						job.Id,
						config.GetConfig().Judger.Key,
						foundationjudge.JudgeStatusJudgeFail,
					)
					if markErr != nil {
						metapanic.ProcessError(markErr)
					}
					return err
				}
				return nil
			},
		)
	}
	return nil
}

func (s *JudgeService) startJudgeTask(job *foundationmodel.JudgeJob) error {
	ctx := context.Background()
	jobId := job.Id
	ok, err := foundationdao.GetJudgeJobDao().StartProcessJudgeJob(ctx, job.Id, config.GetConfig().Judger.Key)
	if err != nil {
		return metaerror.Wrap(err, "failed to start process judge job")
	}
	if !ok {
		// 如果没有成功处理，可以认为是中途已经被别的判题机处理了
		return nil
	}
	problem, err := foundationdao.GetProblemDao().GetProblemViewForJudge(ctx, job.ProblemId)
	if err != nil {
		return metaerror.Wrap(err, "failed to get problem")
	}
	if problem == nil {
		return metaerror.New("problem not found: %s", job.ProblemId)
	}
	if problem.JudgeMd5 == nil {
		return metaerror.New("problem judge md5 is nil: %d", job.ProblemId)
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
			markErr := foundationdao.GetJudgeJobCompileDao().MarkJudgeJobCompileMessage(
				ctx, job.Id, config.GetConfig().Judger.Key,
				extraMessage,
			)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
		}
		defer func() {
			for _, fileId := range execFileIds {
				goJudgeUrl := config.GetConfig().GoJudge.Url
				deleteUrl := metahttp.UrlJoin(goJudgeUrl, "file", fileId)
				err := foundationjudge.DeleteFile(s.goJudgeClient, strconv.Itoa(jobId), deleteUrl)
				if err != nil {
					return
				}
			}
		}()
		if err != nil {
			return err
		}
		if compileStatus != foundationjudge.JudgeStatusAC {
			err := foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(
				ctx,
				job.Id,
				config.GetConfig().Judger.Key,
				compileStatus,
			)
			if err != nil {
				metapanic.ProcessError(err)
			}
			return nil
		}
		slog.Info("compile code success", "job", job.Id, "execFileIds", execFileIds)
		err = foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(
			ctx,
			job.Id,
			config.GetConfig().Judger.Key,
			foundationjudge.JudgeStatusRunning,
		)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}
	err = s.runJudgeJob(ctx, job, problem, execFileIds)
	return err
}

func (s *JudgeService) updateJudgeData(ctx context.Context, problemId int, md5 string) error {
	val, _ := s.problemMutexMap.LoadOrStore(problemId, &judgeMutexEntry{})
	e := val.(*judgeMutexEntry)
	atomic.AddInt32(&e.ref, 1)
	defer func() {
		if atomic.AddInt32(&e.ref, -1) == 0 {
			s.problemMutexMap.Delete(problemId)
		}
	}()
	e.mu.Lock()
	defer e.mu.Unlock()
	judgeMd5FilePath := path.Join(".judge_data", strconv.Itoa(problemId), md5)
	// 判断 judgeMd5FilePath 是否存在
	_, err := os.Stat(judgeMd5FilePath)
	if err == nil {
		// 文件存在，直接返回
		return nil
	} else if !os.IsNotExist(err) {
		// 其他错误，返回报错
		return metaerror.Wrap(err, "failed to stat judge md5 file")
	}
	err = s.downloadJudgeData(ctx, problemId, md5)
	if err != nil {
		// 有可能下载了一半，因此删除文件夹
		slog.Error("download judge data failed", "problemId", problemId, "error", err)
		judgeDataDir := path.Join(".judge_data", strconv.Itoa(problemId))
		removeErr := os.RemoveAll(judgeDataDir)
		if err != nil {
			err = metaerror.Join(err, removeErr)
		}
		return metaerror.Wrap(err, "failed to download judge data")
	}
	return err
}

func (s *JudgeService) downloadJudgeData(ctx context.Context, problemId int, md5 string) error {

	slog.Info("downloading judge data", "problemId", problemId)

	r2Client := cfr2.GetSubsystem().GetClient("judge-data")
	if r2Client == nil {
		return metaerror.New("r2Client is nil")
	}

	// 删除旧的判题数据
	judgeDataDir := path.Join(".judge_data", strconv.Itoa(problemId))
	err := os.RemoveAll(judgeDataDir)

	// 删除旧缓存的spj
	specialFileId := s.getSpecialFileId(problemId)
	if specialFileId != "" {
		goJudgeUrl := config.GetConfig().GoJudge.Url
		deleteUrl := metahttp.UrlJoin(goJudgeUrl, "file", specialFileId)
		err = foundationjudge.DeleteFile(s.goJudgeClient, "delete special judge file", deleteUrl)
		if err != nil {
			slog.Error("delete special judge file failed", "problemId", problemId, "error", err)
			return metaerror.Wrap(err, "failed to delete special judge file")
		}
		slog.Info("delete special judge file success", "problemId", problemId, "fileId", specialFileId)
		// 清除缓存
		delete(s.specialFileIds, problemId)
	}

	// 1. 列出 problemId 目录下的所有对象
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("didaoj-judge"),
		Prefix: aws.String(strconv.Itoa(problemId) + "/"), // 确保带 `/`，只列出这个目录下的
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var downloadErr error

	err = r2Client.ListObjectsV2PagesWithContext(
		ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				if strings.HasSuffix(*obj.Key, ".zip") {
					continue
				}
				wg.Add(1)
				routine.SafeGo(
					"download judge data", func() error {
						defer wg.Done()
						localPath := path.Join(".judge_data", *obj.Key)
						var finalErr error
						_ = retry.TryRetrySleep(
							"download judge data", 6, time.Second*10, func(i int) bool {
								err := s.downloadObject(ctx, r2Client, "didaoj-judge", *obj.Key, localPath)
								if err != nil {
									finalErr = err
									return false
								}
								finalErr = nil
								return true
							},
						)
						if finalErr != nil {
							mu.Lock()
							defer mu.Unlock()
							if downloadErr == nil {
								downloadErr = finalErr
							}
						}
						return nil
					},
				)
			}
			return true
		},
	)
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
func (s *JudgeService) downloadObject(
	ctx context.Context,
	s3Client *s3.S3,
	bucket, key string,
	localPath string,
) error {
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

func (s *JudgeService) compileSpecialJudge(
	job *foundationmodel.JudgeJob,
	md5 string,
	jobConfig *foundationjudge.JudgeJobConfig,
) (string, error) {
	problemId := job.ProblemId

	specialFileId := s.getSpecialFileId(problemId)
	if specialFileId != "" {
		return specialFileId, nil
	}

	val, _ := s.problemMutexMap.LoadOrStore(problemId, &judgeMutexEntry{})
	e := val.(*judgeMutexEntry)
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

	judgeDataDir := path.Join(".judge_data", strconv.Itoa(problemId), md5)

	codeFilePath := filepath.Join(judgeDataDir, jobConfig.SpecialJudge.Source)
	codeContent, err := metastring.GetStringFromOpenFile(codeFilePath)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to read special judge code file")
	}

	execFileIds, extraMessage, compileStatus, err := foundationjudge.CompileCode(
		s.goJudgeClient,
		strconv.Itoa(job.Id),
		runUrl,
		language,
		codeContent,
		s.configFileIds,
		true,
	)
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
		s.specialFileIds = make(map[int]string)
	}
	s.specialFileIds[problemId] = specialFileId
	return specialFileId, nil
}

func (s *JudgeService) compileCode(job *foundationmodel.JudgeJob) (
	map[string]string,
	string,
	foundationjudge.JudgeStatus,
	error,
) {
	goJudgeUrl := config.GetConfig().GoJudge.Url
	runUrl := metahttp.UrlJoin(goJudgeUrl, "run")
	return foundationjudge.CompileCode(
		s.goJudgeClient,
		strconv.Itoa(job.Id),
		runUrl,
		job.Language,
		job.Code,
		s.configFileIds,
		false,
	)
}

func (s *JudgeService) runJudgeJob(
	ctx context.Context,
	job *foundationmodel.JudgeJob,
	problem *foundationview.ProblemForJudge,
	execFileIds map[string]string,
) error {
	problemId := job.ProblemId

	timeLimit := problem.TimeLimit
	memoryLimit := problem.MemoryLimit
	md5 := *problem.JudgeMd5

	judgeDataDir := path.Join(".judge_data", strconv.Itoa(problemId), md5)

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
		sort.Slice(
			outFileNames, func(i, j int) bool {
				return outFileNames[i] < outFileNames[j]
			},
		)
		totalScore := 1000
		taskCount = len(outFileNames)
		averageScore := totalScore / taskCount
		for _, file := range outFileNames {
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
			judgeTaskConfig.Score = averageScore
			jobConfig.Tasks = append(jobConfig.Tasks, judgeTaskConfig)
		}
		leftScore := totalScore % taskCount
		for i := taskCount - 1; i >= 0 && leftScore > 0; i-- {
			jobConfig.Tasks[i].Score += 1
			leftScore--
		}
	}

	taskCount = len(jobConfig.Tasks)

	if taskCount == 0 {
		return metaerror.New("no job task found")
	}

	// 把配置的分数转换为总值为1000
	sumScore := 0
	for _, taskConfig := range jobConfig.Tasks {
		sumScore += taskConfig.Score
	}
	scoreRate := 1000.0 / sumScore
	sumScore = 0
	for _, taskConfig := range jobConfig.Tasks {
		taskConfig.Score = taskConfig.Score * scoreRate
		sumScore += taskConfig.Score
	}
	leftScore := 1000 - sumScore
	for i := taskCount - 1; i >= 0 && leftScore > 0; i-- {
		jobConfig.Tasks[i].Score += 1
		leftScore--
	}

	err = foundationdao.GetJudgeJobDao().MarkJudgeJobTaskTotal(ctx, job.Id, config.GetConfig().Judger.Key, taskCount)
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
		finalStatus, sumTime, sumMemory, finalScore, err = s.runJudgeTask(
			ctx,
			job,
			taskConfig,
			cpuLimit,
			memoryLimit,
			sumTime,
			sumMemory,
			finalScore,
			specialFileId,
			judgeDataDir,
			execFileIds,
			finalStatus,
		)
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

	err = foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeFinalStatus(
		ctx, job.Id, config.GetConfig().Judger.Key,
		finalStatus,
		problemId,
		job.Inserter,
		finalScore,
		finalTime,
		finalMemory,
	)

	return err
}

func (s *JudgeService) runJudgeTask(
	ctx context.Context,
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
		Id(job.Id).
		TaskId(key).
		Status(foundationjudge.JudgeStatusJudgeFail).
		Time(0).
		Memory(0).
		Score(0).
		Content("").
		Hint("").
		Build()

	goJudgeUrl := config.GetConfig().GoJudge.Url
	runUrl := metahttp.UrlJoin(goJudgeUrl, "run")

	var inContent string

	if taskConfig.InFile != "" {
		inContent, err = metastring.GetStringFromOpenFile(path.Join(judgeDataDir, taskConfig.InFile))
		if err != nil {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(
				ctx,
				job.Id,
				config.GetConfig().Judger.Key,
				task,
			)
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
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(
				ctx,
				job.Id,
				config.GetConfig().Judger.Key,
				task,
			)
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
		className := foundationjudge.GetJavaClass(job.Code)
		if className == "" {
			return foundationjudge.JudgeStatusCE, 0, 0, 0, err
		}
		packageName := foundationjudge.GetJavaPackage(job.Code)
		qualifiedName := className
		if packageName != "" {
			qualifiedName = packageName + "." + className
		}
		jarFileName := className + ".jar"
		args = []string{
			"java",
			"-Dfile.encoding=UTF-8",
			"-cp",
			jarFileName,
			qualifiedName,
		}
		fileId, ok := execFileIds[jarFileName]
		if !ok {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(
				ctx,
				job.Id,
				config.GetConfig().Judger.Key,
				task,
			)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.New("fileId not found")
		}
		copyIns = map[string]interface{}{
			jarFileName: map[string]interface{}{
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
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(
				ctx,
				job.Id,
				config.GetConfig().Judger.Key,
				task,
			)
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
				//"dataSegmentLimit":  true,
				//"addressSpaceLimit": true,
				"copyIn": copyIns,
			},
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(
			ctx,
			job.Id,
			config.GetConfig().Judger.Key,
			task,
		)
		if markErr != nil {
			metapanic.ProcessError(markErr)
		}
		return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err)
	}

	_, respBody, err := metahttp.SendRequestRetry(
		s.goJudgeClient,
		strconv.Itoa(job.Id),
		6,
		time.Second*10,
		http.MethodPost, runUrl,
		nil,
		bytes.NewBuffer(jsonData),
		true,
	)
	if err != nil {
		slog.Warn("runJudgeTask err", "jsonData", data)
		return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err, "failed to send request to GoJudge")
	}
	var responseDataList []struct {
		Status gojudge.Status `json:"status"`
		Files  struct {
			Stderr string `json:"stderr"`
			Stdout string `json:"stdout"`
		} `json:"files"`
		Error  string `json:"error"`
		Time   int    `json:"time"`
		Memory int    `json:"memory"`
	}
	err = json.Unmarshal(respBody, &responseDataList)
	if err != nil {
		markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(
			ctx,
			job.Id,
			config.GetConfig().Judger.Key,
			task,
		)
		if markErr != nil {
			metapanic.ProcessError(markErr)
		}
		return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err, "failed to decode response")
	}
	if len(responseDataList) != 1 {
		markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(
			ctx,
			job.Id,
			config.GetConfig().Judger.Key,
			task,
		)
		if markErr != nil {
			metapanic.ProcessError(markErr)
		}
		return finalStatus, sumTime, sumMemory, finalScore, metaerror.New(
			"unexpected response length: %d",
			len(responseDataList),
		)
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
		markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(
			ctx,
			job.Id,
			config.GetConfig().Judger.Key,
			task,
		)
		if markErr != nil {
			metapanic.ProcessError(
				metaerror.Wrap(
					markErr,
					"failed to add judge job task current:%d task:%s",
					job.Id,
					task.TaskId,
				),
			)
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
					WaHint = fmt.Sprintf("#%d %s != %s", i+1, outContentMyPe[i], ansContentMyPe[i])
				}
			} else {
				WaHint = fmt.Sprintf("#%d %s not found", i+1, outContentMyPe[i])
				break
			}
		}
		if WaHint != "" {
			task.Status = foundationjudge.JudgeStatusWA
			task.Hint = metastring.GetTextEllipsis(WaHint, 1000)
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
					"args": []string{"spj", "test.in", "user.out", "test.out"},
					"env":  []string{"PATH=/usr/bin:/bin"},
					"files": []map[string]interface{}{
						{"content": inContent},
						{"name": "stdout", "max": 10240},
						{"name": "stderr", "max": 10240},
					},
					"cpuLimit":    30000000,          // 提供30秒给spj
					"memoryLimit": 512 * 1024 * 1024, // 提供512MB秒给spj
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
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(
				ctx,
				job.Id,
				config.GetConfig().Judger.Key,
				task,
			)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err)
		}
		_, specialResp, err := metahttp.SendRequestRetry(
			s.goJudgeClient,
			strconv.Itoa(job.Id),
			6,
			time.Second*10,
			http.MethodPost, runUrl,
			nil,
			bytes.NewBuffer(specialJsonData),
			true,
		)
		if err != nil {
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err)
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
		err = json.Unmarshal(specialResp, &specialRespDataList)
		if err != nil {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(
				ctx,
				job.Id,
				config.GetConfig().Judger.Key,
				task,
			)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err, "failed to decode response")
		}
		if len(specialRespDataList) != 1 {
			markErr := foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(
				ctx,
				job.Id,
				config.GetConfig().Judger.Key,
				task,
			)
			if markErr != nil {
				metapanic.ProcessError(markErr)
			}
			return finalStatus, sumTime, sumMemory, finalScore, metaerror.New(
				"unexpected response length: %d",
				len(specialRespDataList),
			)
		}
		specialRespData := specialRespDataList[0]

		if task.Content != "" {
			task.Content = task.Content + "\n"
		}
		task.Content = task.Content + specialRespData.Files.Stdout

		if task.Hint != "" {
			task.Hint = task.Hint + "\n"
		}
		task.Hint = task.Hint + specialRespData.Files.Stderr
		if specialRespData.Status == gojudge.StatusAccepted {
			task.Score = taskConfig.Score
			finalScore += taskConfig.Score
			task.Status = foundationjudge.JudgeStatusAC
		} else if specialRespData.Status == gojudge.StatusTimeLimit {
			task.Status = foundationjudge.JudgeStatusTLE
			if task.Content != "" {
				task.Content = task.Content + "\n"
			}
			task.Content = task.Content + "spj Time Limit Exceeded"
		} else if specialRespData.Status == gojudge.StatusMemoryLimit {
			task.Status = foundationjudge.JudgeStatusMLE
			if task.Content != "" {
				task.Content = task.Content + "\n"
			}
			task.Content = task.Content + "spj Memory Limit Exceeded"
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
	err = foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id, config.GetConfig().Judger.Key, task)
	if err != nil {
		return finalStatus, sumTime, sumMemory, finalScore, metaerror.Wrap(err, "failed to add judge job task")
	}
	return finalStatus, sumTime, sumMemory, finalScore, nil
}
