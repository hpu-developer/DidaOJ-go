package service

import (
	"context"
	"fmt"
	foundationdao "foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationremote "foundation/foundation-remote"
	"judge/config"
	"log/slog"
	"meta/cron"
	metaerror "meta/meta-error"
	metapanic "meta/meta-panic"
	"meta/routine"
	"meta/singleton"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// RemoteMutexEntry 需要保证只有一个goroutine在处理判题数据
type RemoteMutexEntry struct {
	mu  sync.Mutex
	ref int32
}

type RemoteService struct {
	requestMutex sync.Mutex
	runningTasks atomic.Int32

	// 防止因重判等情况多次获取到了同一个判题任务（不过多个判题机则靠key来忽略）
	JudgeJobMutexMap sync.Map
	// 有些时候同一个问题只能有一个逻辑去处理
	problemMutexMap sync.Map

	// 题目号对应的特判程序文件ID
	specialFileIds map[int]string
	// 配置静态文件标识与文件ID的映射
	configFileIds map[string]string

	goJudgeClient *http.Client
}

var singletonRemoteService = singleton.Singleton[RemoteService]{}

func GetRemoteService() *RemoteService {
	return singletonRemoteService.GetInstance(
		func() *RemoteService {
			s := &RemoteService{}
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

func (s *RemoteService) Start() error {

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

func (s *RemoteService) handleStart() error {

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

	maxJob := config.GetConfig().MaxJobRemote
	runningCount := int(s.runningTasks.Load())
	if runningCount >= maxJob {
		return nil
	}
	ctx := context.Background()
	jobs, err := foundationdao.GetJudgeJobDao().RequestRemoteJudgeJobListPendingJudge(
		ctx,
		maxJob-runningCount,
		config.GetConfig().Judger.Key,
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to get Remote job list")
	}
	jobsCount := len(jobs)
	if jobsCount == 0 {
		return nil
	}

	slog.Info("get Remote job list", "runningCount", runningCount, "maxJob", maxJob, "count", jobsCount)

	s.runningTasks.Add(int32(jobsCount))

	for _, job := range jobs {
		routine.SafeGo(
			fmt.Sprintf("RunningJudgeJob_%d", job.Id), func() error {
				defer func() {
					slog.Info(fmt.Sprintf("RemoteTask_%d end", job.Id))
					s.runningTasks.Add(-1)
				}()
				val, _ := s.JudgeJobMutexMap.LoadOrStore(job.Id, &RemoteMutexEntry{})
				e := val.(*RemoteMutexEntry)
				atomic.AddInt32(&e.ref, 1)
				defer func() {
					if atomic.AddInt32(&e.ref, -1) == 0 {
						s.JudgeJobMutexMap.Delete(job.Id)
					}
				}()
				e.mu.Lock()
				defer e.mu.Unlock()

				slog.Info(fmt.Sprintf("RemoteTask_%d start", job.Id))
				err = s.startRemoteTask(job)
				if err != nil {
					markErr := foundationdao.GetJudgeJobCompileDao().MarkJudgeJobCompileMessage(
						ctx, job.Id, config.GetConfig().Judger.Key,
						err.Error(),
					)
					if markErr != nil {
						metapanic.ProcessError(markErr)
					}
					markErr = foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(
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

func (s *RemoteService) startRemoteTask(job *foundationmodel.JudgeJob) error {
	ctx := context.Background()
	ok, err := foundationdao.GetJudgeJobDao().StartProcessJudgeJob(ctx, job.Id, config.GetConfig().Judger.Key)
	if err != nil {
		return metaerror.Wrap(err, "failed to start process Remote job")
	}
	if !ok {
		// 如果没有成功处理，可以认为是中途已经被别的判题机处理了
		return nil
	}
	problem, err := foundationdao.GetProblemDao().GetProblemViewForRemoteJudge(ctx, job.ProblemId)
	if err != nil {
		return metaerror.Wrap(err, "failed to get problem")
	}
	if problem == nil {
		return metaerror.New("problem not found: %s", job.ProblemId)
	}
	oj := foundationremote.GetRemoteTypeByString(problem.OriginOj)
	agent := foundationremote.GetRemoteAgent(oj)
	remoteId, remoteAccount, err := agent.PostSubmitJudgeJob(ctx, problem.OriginId, job.Language, job.Code)
	if err != nil {
		markErr := foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeFinalStatus(
			ctx, job.Id, config.GetConfig().Judger.Key,
			foundationjudge.JudgeStatusSubmitFail,
			problem.Id,
			job.Inserter,
			0,
			0,
			0,
		)
		if markErr != nil {
			metapanic.ProcessError(markErr)
		}
		return nil
	}
	err = foundationdao.GetJudgeJobDao().MarkJudgeJobRemoteSubmit(
		ctx, job.Id, config.GetConfig().Judger.Key,
		remoteId,
		remoteAccount,
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to mark Remote judge job submit")
	}

	// 每隔3秒更新一次结果
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	currentStatus := foundationjudge.JudgeStatusQueuing

	for {
		select {
		case <-ticker.C:
			status, score, finalTime, finalMemory, err := agent.GetJudgeJobStatus(ctx, remoteId)
			if err != nil {
				return metaerror.Wrap(err, "failed to get Remote judge job status")
			}
			if foundationjudge.IsJudgeStatusRunning(status) {
				if currentStatus != status {
					currentStatus = status
					err := foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(
						ctx,
						job.Id,
						config.GetConfig().Judger.Key,
						currentStatus,
					)
					if err != nil {
						return metaerror.Wrap(err, "failed to mark Remote judge job running status")
					}
				}
				continue
			}
			extraMessage, err := agent.GetJudgeJobExtraMessage(ctx, remoteId, status)
			if err != nil {
				return metaerror.Wrap(err, "failed to get Remote judge job extra message")
			}
			if extraMessage != "" {
				markErr := foundationdao.GetJudgeJobCompileDao().MarkJudgeJobCompileMessage(
					ctx, job.Id, config.GetConfig().Judger.Key,
					extraMessage,
				)
				if markErr != nil {
					return metaerror.Wrap(markErr, "failed to mark Remote judge job extra message")
				}
			}
			err = foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeFinalStatus(
				ctx, job.Id, config.GetConfig().Judger.Key,
				status,
				problem.Id,
				job.Inserter,
				score,
				finalTime,
				finalMemory,
			)
			if err != nil {
				return metaerror.Wrap(err, "failed to mark Remote judge job final status")
			}
			return nil
		}
	}
}
