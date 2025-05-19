package service

import (
	"bytes"
	"context"
	"encoding/json"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	foundationstatus "foundation/foundation-status"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"judge/config"
	"log/slog"
	cfr2 "meta/cf-r2"
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
	ctx := context.Background()

	err := s.registerJudger(ctx)
	if err != nil {
		return err
	}

	c := cron.NewWithSeconds()
	// 每3秒运行一次任务
	_, err = c.AddFunc(
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

	r2Client := cfr2.GetSubsystem().GetClient("didapipa-oj")
	if r2Client == nil {
		return metaerror.New("r2Client is nil")
	}
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
	statusJsonData := foundationstatus.NewJudgerStatusBuilder().
		Name(config.GetConfig().Judger.Name).
		CpuUsage(cpuUsage).
		MemUsage(memoryUsed).
		MemTotal(memoryTotal).
		AvgMessage(avgMessage).
		UpdateTime(nowTime).
		Build()
	statusBytes, err := json.Marshal(statusJsonData)
	if err != nil {
		return metaerror.Wrap(err, "marshal status json failed")
	}
	key := "status/judger/" + config.GetConfig().Judger.Key + ".json"
	_, err = r2Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String("didapipa-oj"),
		Key:         aws.String(key),
		Body:        bytes.NewReader(statusBytes),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return metaerror.Wrap(err, "put object error, key: %s", key)
	}
	return nil
}

func (s *StatusService) IsReportError() bool {
	return s.isReportError
}

func (s *StatusService) registerJudger(ctx context.Context) error {
	judger := foundationmodel.NewJudgerBuilder().
		Key(config.GetConfig().Judger.Key).
		Name(config.GetConfig().Judger.Name).
		Build()
	err := foundationdao.GetJudgerDao().UpdateJudger(ctx, judger)
	if err != nil {
		return err
	}
	judgers, err := foundationdao.GetJudgerDao().GetJudgers(ctx)
	if err != nil {
		return err
	}
	judgerJsonData := struct {
		Judgers []*foundationmodel.Judger `json:"judgers"`
	}{
		Judgers: judgers,
	}
	// 上传到R2
	r2Client := cfr2.GetSubsystem().GetClient("didapipa-oj")
	if r2Client == nil {
		return metaerror.New("r2Client is nil")
	}
	statusBytes, err := json.Marshal(judgerJsonData)
	if err != nil {
		return metaerror.Wrap(err, "marshal status json failed")
	}
	key := "status/judger.json"
	_, err = r2Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String("didapipa-oj"),
		Key:         aws.String(key),
		Body:        bytes.NewReader(statusBytes),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return metaerror.Wrap(err, "put object error, key: %s", key)
	}
	return nil
}
