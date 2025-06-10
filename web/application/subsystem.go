package application

import (
	"bytes"
	"context"
	"encoding/json"
	foundationstatus "foundation/foundation-status"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	cfr2 "meta/cf-r2"
	"meta/engine"
	metaerror "meta/meta-error"
	metasystem "meta/meta-system"
	"meta/subsystem"
	"time"
)

type Subsystem struct {
	subsystem.Subsystem
}

func GetSubsystem() *Subsystem {
	if thisSubsystem := engine.GetSubsystem[*Subsystem](); thisSubsystem != nil {
		return thisSubsystem.(*Subsystem)
	}
	return nil
}

func (s *Subsystem) GetName() string {
	return "Migrate"
}

func (s *Subsystem) Start() error {
	err := s.startSubSystem()
	if err != nil {
		return err
	}
	return nil
}

func (s *Subsystem) startSubSystem() error {

	//var err error
	//
	//c := cron.NewWithSeconds()
	//// 每3秒运行一次任务
	//_, err = c.AddFunc(
	//	"0/3 * * * * ?", func() {
	//		err := s.handleStart()
	//		if err != nil {
	//			metapanic.ProcessError(err)
	//			return
	//		}
	//	},
	//)
	//if err != nil {
	//	return metaerror.Wrap(err, "error adding function to cron")
	//}
	//
	//c.Start()

	return nil
}

func (s *Subsystem) handleStart() error {

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
	statusJsonData := foundationstatus.NewWeberStatusBuilder().
		Name("DidaOJ").
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
	key := "status/web.json"
	_, err = r2Client.PutObjectWithContext(
		ctx, &s3.PutObjectInput{
			Bucket:      aws.String("didapipa-oj"),
			Key:         aws.String(key),
			Body:        bytes.NewReader(statusBytes),
			ContentType: aws.String("application/json"),
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "put object error, key: %s", key)
	}
	return nil
}
