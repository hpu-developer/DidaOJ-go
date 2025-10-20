package service

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"judge/config"
	"log/slog"
	"meta/cron"
	metaerror "meta/meta-error"
	metaformat "meta/meta-format"
	metapanic "meta/meta-panic"
	metasystem "meta/meta-system"
	"meta/singleton"
	"time"
)

type StatusService struct {
	isReportError bool
	isEnableJudge bool
}

var singletonStatusService = singleton.Singleton[StatusService]{}

func GetStatusService() *StatusService {
	return singletonStatusService.GetInstance(
		func() *StatusService {
			s := &StatusService{}
			return s
		},
	)
}

func (s *StatusService) Start() error {

	s.isEnableJudge = false
	s.isReportError = false

	c := cron.NewWithSeconds()
	// 每3秒运行一次任务
	_, err := c.AddFunc(
		"0/3 * * * * ?", func() {
			err := s.handleStart()
			if err != nil {
				s.isReportError = true
				metapanic.ProcessError(err)
				return
			}
			s.isReportError = false
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "error adding function to cron")
	}

	c.Start()

	return nil
}

func (s *StatusService) handleStart() error {

	slog.Info("status service start", "judger", metaformat.StringByJson(config.GetConfig().Judger))

	ctx := context.Background()

	nowTime := time.Now()

	cpuUsage, err := metasystem.GetCpuUsage()
	if err != nil {
		return metaerror.Wrap(err, "get cpu usage failed")
	}
	memoryUsed, memoryTotal, err := metasystem.GetVirtualMemory()
	if err != nil {
		return metaerror.Wrap(err, "get memory usage failed")
	}
	avgMessage, err := metasystem.GetAvgMessage()
	if err != nil {
		return metaerror.Wrap(err, "get avg message failed")
	}

	// 构建 Judger 状态 JSON 数据
	judgerData := foundationmodel.NewJudgerBuilder().
		Key(config.GetConfig().Judger.Key).
		Name(config.GetConfig().Judger.Name).
		MaxJob(config.GetConfig().MaxJob).
		CpuUsage(cpuUsage).
		MemUsage(memoryUsed).
		MemTotal(memoryTotal).
		AvgMessage(avgMessage).
		ModifyTime(nowTime).
		Build()

	err = foundationdao.GetJudgerDao().UpdateJudger(ctx, judgerData)
	if err != nil {
		return err
	}

	s.isEnableJudge, err = foundationdao.GetJudgerDao().IsEnableJudge(ctx, config.GetConfig().Judger.Key)
	if err != nil {
		return metaerror.Wrap(err, "get is enable judge failed")
	}
	return nil
}

func (s *StatusService) IsReportError() bool {
	return s.isReportError
}

func (s *StatusService) IsEnableJudge() bool {
	return s.isEnableJudge
}
