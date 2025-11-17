package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	foundationdao "foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationrun "foundation/foundation-run"
	"judge/config"
	gojudge "judge/go-judge"
	"log/slog"
	"meta/cron"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	"meta/routine"
	"meta/singleton"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// 需要保证只有一个goroutine在处理判题数据
type runMutexEntry struct {
	mu  sync.Mutex
	ref int32
}

type RunService struct {
	requestMutex sync.Mutex
	runningTasks atomic.Int32

	// 防止因重判等情况多次获取到了同一个判题任务（不过多个判题机则靠key来忽略）
	runJobMutexMap sync.Map

	goJudgeClient *http.Client
}

var singletonRunService = singleton.Singleton[RunService]{}

func GetRunService() *RunService {
	return singletonRunService.GetInstance(
		func() *RunService {
			s := &RunService{}
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

func (s *RunService) Start() error {

	c := cron.NewWithSeconds()
	_, err := c.AddFunc(
		"* * * * * ?", func() {
			// 每秒检查一次任务是否能运行
			s.checkStartJob()
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "error adding function to cron")
	}

	c.Start()

	return nil
}

func (s *RunService) checkStartJob() {
	err := s.handleStart()
	if err != nil {
		metapanic.ProcessError(err)
	}
}

func (s *RunService) handleStart() error {

	// 如果没开启评测，停止判题
	if !GetStatusService().IsEnableJudge() {
		return nil
	}
	// 如果上报状态报错，停止判题
	if GetStatusService().IsReportError() {
		return nil
	}

	// 保证同时只有一个handleStart
	if !s.requestMutex.TryLock() {
		return nil
	}
	defer s.requestMutex.Unlock()

	maxJob := config.GetConfig().MaxJobRun
	runningCount := int(s.runningTasks.Load())
	if runningCount >= maxJob {
		return nil
	}
	ctx := context.Background()
	jobs, err := foundationdao.GetRunJobDao().RequestRunJobListPending(
		ctx,
		maxJob-runningCount,
		config.GetConfig().Judger.Key,
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to get run job list")
	}
	jobsCount := len(jobs)
	if jobsCount == 0 {
		return nil
	}

	slog.Info("get run job list", "runningCount", runningCount, "maxJob", maxJob, "count", jobsCount)

	s.runningTasks.Add(int32(jobsCount))

	for _, job := range jobs {
		routine.SafeGo(
			fmt.Sprintf("RunningRunJob_%d", job.Id), func() error {
				// 执行完本Job后再尝试启动一次任务
				defer s.checkStartJob()

				defer func() {
					slog.Info(fmt.Sprintf("RunJob_%d end", job.Id))
					s.runningTasks.Add(-1)
				}()
				val, _ := s.runJobMutexMap.LoadOrStore(job.Id, &runMutexEntry{})
				e := val.(*runMutexEntry)
				atomic.AddInt32(&e.ref, 1)

				defer func() {
					if atomic.AddInt32(&e.ref, -1) == 0 {
						s.runJobMutexMap.Delete(job.Id)
					}
				}()
				e.mu.Lock()
				defer e.mu.Unlock()

				slog.Info(fmt.Sprintf("RunJob_%d start", job.Id))
				err = s.startRunJob(job)
				if err != nil {
					markErr := foundationdao.GetRunJobDao().MarkRunJobRunStatus(
						ctx,
						job.Id,
						config.GetConfig().Judger.Key,
						foundationrun.RunStatusRunFail,
						err.Error(),
						0, 0,
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

func (s *RunService) runJudgeTask(
	ctx context.Context,
	job *foundationmodel.RunJob,
	cpuLimit int, memoryLimit int,
	execFileIds map[string]string,
) (foundationrun.RunStatus, int, int, string, error) {

	var args []string
	var copyIns map[string]interface{}
	switch job.Language {
	case foundationjudge.JudgeLanguageC, foundationjudge.JudgeLanguageCpp,
		foundationjudge.JudgeLanguagePascal, foundationjudge.JudgeLanguageGolang:
		args = []string{"a"}
		fileId, ok := execFileIds["a"]
		if !ok {
			return foundationrun.RunStatusRunFail, 0, 0, "", metaerror.New("fileId not found")
		}
		copyIns = map[string]interface{}{
			"a": map[string]interface{}{
				"fileId": fileId,
			},
		}
	case foundationjudge.JudgeLanguageJava:
		className := foundationjudge.GetJavaClass(job.Code)
		if className == "" {
			return foundationrun.RunStatusCE, 0, 0, "", metaerror.New("class name not found")
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
			return foundationrun.RunStatusRunFail, 0, 0, "", metaerror.New("fileId not found")
		}
		copyIns = map[string]interface{}{
			jarFileName: map[string]interface{}{
				"fileId": fileId,
			},
		}
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
			return foundationrun.RunStatusRunFail, 0, 0, "", metaerror.New("fileId not found")
		}
		copyIns = map[string]interface{}{
			"a.js": map[string]interface{}{
				"fileId": fileId,
			},
		}
	default:
		return foundationrun.RunStatusRunFail, 0, 0, "", metaerror.New("language not support: %d", job.Language)
	}

	data := map[string]interface{}{
		"cmd": []map[string]interface{}{
			{
				"args": args,
				"env":  []string{"PATH=/usr/bin:/bin"},
				"files": []map[string]interface{}{
					{"content": job.Input},
					{"name": "stdout", "max": 10240},
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
		return foundationrun.RunStatusRunFail, 0, 0, "", metaerror.Wrap(err)
	}

	goJudgeUrl := config.GetConfig().GoJudge.Url
	runUrl := metahttp.UrlJoin(goJudgeUrl, "run")

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
		slog.Warn("runRunTask err", "jsonData", data)
		return foundationrun.RunStatusRunFail, 0, 0, "", metaerror.Wrap(err)
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
		return foundationrun.RunStatusRunFail, 0, 0, "", metaerror.Wrap(err, "failed to decode response")
	}
	if len(responseDataList) != 1 {
		return foundationrun.RunStatusRunFail, 0, 0, "", metaerror.New(
			"unexpected response length: %d",
			len(responseDataList),
		)
	}
	responseData := responseDataList[0]

	status := foundationrun.RunStatusFinish

	if responseData.Status != gojudge.StatusAccepted {
		switch responseData.Status {
		case gojudge.StatusSignalled:
			status = foundationrun.RunStatusRE
		case gojudge.StatusNonzeroExit:
			status = foundationrun.RunStatusRE
		case gojudge.StatusInternalError:
			slog.Warn("internal error", "job", job.Id, "responseData", responseData)
			status = foundationrun.RunStatusRunFail
		case gojudge.StatusOutputLimit:
			status = foundationrun.RunStatusOLE
		case gojudge.StatusFileError:
			status = foundationrun.RunStatusOLE
		case gojudge.StatusMemoryLimit:
			status = foundationrun.RunStatusMLE
		case gojudge.StatusTimeLimit:
			status = foundationrun.RunStatusTLE
		default:
			slog.Warn("status error", "job", job.Id, "responseData", responseData)
			status = foundationrun.RunStatusRunFail
		}
	}

	finalTime := responseData.Time
	finalMemory := responseData.Memory

	content := responseData.Files.Stdout
	if len(content) > 0 {
		content += "\n"
	}
	content += responseData.Files.Stderr

	return status, finalTime, finalMemory, content, nil
}

func (s *RunService) startRunJob(job *foundationmodel.RunJob) error {
	ctx := context.Background()
	jobId := job.Id
	ok, err := foundationdao.GetRunJobDao().StartProcessRunJob(
		ctx,
		job.Id,
		config.GetConfig().Judger.Key,
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to start process run job")
	}
	if !ok {
		// 如果没有成功处理，可以认为是中途已经被别的判题机处理了
		return nil
	}

	var execFileIds map[string]string
	var extraMessage string
	var compileStatus foundationjudge.JudgeStatus

	goJudgeUrl := config.GetConfig().GoJudge.Url
	runUrl := metahttp.UrlJoin(goJudgeUrl, "run")

	if foundationjudge.IsLanguageNeedCompile(job.Language) {
		execFileIds, extraMessage, compileStatus, err = foundationjudge.CompileCode(
			s.goJudgeClient,
			strconv.Itoa(job.Id),
			runUrl,
			job.Language,
			job.Code,
			GetJudgeService().configFileIds,
			false,
		)
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
		realStatus := foundationrun.RunStatusRunning
		if compileStatus != foundationjudge.JudgeStatusAC {
			realStatus = foundationrun.RunStatusCE
			if compileStatus != foundationjudge.JudgeStatusCE {
				realStatus = foundationrun.RunStatusCLE
			}
		}
		slog.Info("compile code success", "run job", job.Id, "execFileIds", execFileIds)
		err = foundationdao.GetRunJobDao().MarkRunJobRunStatus(
			ctx,
			job.Id,
			config.GetConfig().Judger.Key,
			realStatus,
			extraMessage,
			0, 0,
		)
		if err != nil {
			return err
		}
		if realStatus != foundationrun.RunStatusRunning {
			return nil
		}
	}

	// 写死1秒限制
	timeLimit := 1000
	memoryLimit := 65536

	cpuLimit := timeLimit * 1000000
	memoryLimit = memoryLimit * 1024
	if job.Language == foundationjudge.JudgeLanguageJava {
		cpuLimit = cpuLimit + 2000*1000000
		memoryLimit = memoryLimit + 1024*1024*64
	}

	finalStatus, sumTime, sumMemory, content, err := s.runJudgeTask(
		ctx,
		job,
		cpuLimit,
		memoryLimit,
		execFileIds,
	)

	if err != nil {
		return metaerror.Wrap(err, "failed to run task")
	}

	return foundationdao.GetRunJobDao().MarkRunJobRunStatus(ctx, job.Id, config.GetConfig().Judger.Key, finalStatus, content, sumTime, sumMemory)
}
