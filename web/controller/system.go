package controller

import (
	"bytes"
	"encoding/json"
	foundationerrorcode "foundation/error-code"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	cfr2 "meta/cf-r2"
	metacontroller "meta/controller"
	"meta/error-code"
	metaerror "meta/meta-error"
	metapanic "meta/meta-panic"
	"meta/meta-response"
	metasystem "meta/meta-system"
	"time"
)

type SystemController struct {
	metacontroller.Controller
}

func (c *SystemController) GetStatus(ctx *gin.Context) {

	nowTime := time.Now()

	cpuUsage, err := metasystem.GetCpuUsage()
	if err != nil {
		metapanic.ProcessError(metaerror.Wrap(err, "get cpu usage failed"))
		return
	}
	memoryUsed, memoryTotal, err := metasystem.GetVirtualMemory()
	if err != nil {
		metapanic.ProcessError(metaerror.Wrap(err, "get virtual memory failed"))
		return
	}
	avgMessage, err := metasystem.GetAvgMessage()
	if err != nil {
		metapanic.ProcessError(metaerror.Wrap(err, "get avg message failed"))
		return
	}

	judgers, err := foundationservice.GetJudgerService().GetJudgerList(ctx)
	if err != nil {
		metapanic.ProcessError(metaerror.Wrap(err, "get judger list failed"))
		return
	}

	webStatus := foundationmodel.NewWebStatusBuilder().
		Name("DidaOJ").
		CpuUsage(cpuUsage).
		MemUsage(memoryUsed).
		MemTotal(memoryTotal).
		AvgMessage(avgMessage).
		UpdateTime(nowTime).
		Build()

	responseData := struct {
		Web    *foundationmodel.WebStatus `json:"web"`
		Judger []*foundationmodel.Judger  `json:"judger,omitempty"`
	}{
		Web:    webStatus,
		Judger: judgers,
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *SystemController) PostNotification(ctx *gin.Context) {
	var requestData struct {
		Theme   string `json:"theme" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		return
	}
	if requestData.Theme != "success" && requestData.Theme != "info" &&
		requestData.Theme != "warning" && requestData.Theme != "error" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError)
		return
	}
	r2Client := cfr2.GetSubsystem().GetClient("didapipa-oj")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError)
		return
	}
	// 构建 Judger 状态 JSON 数据
	statusJsonData := struct {
		Theme   string `json:"theme"`
		Content string `json:"content"`
	}{
		Theme:   requestData.Theme,
		Content: requestData.Content,
	}
	statusBytes, err := json.Marshal(statusJsonData)
	if err != nil {
		metapanic.ProcessError(metaerror.Wrap(err, "marshal status data failed"))
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError)
		return
	}
	key := "system/notification.json"
	_, err = r2Client.PutObjectWithContext(
		ctx, &s3.PutObjectInput{
			Bucket:      aws.String("didapipa-oj"),
			Key:         aws.String(key),
			Body:        bytes.NewReader(statusBytes),
			ContentType: aws.String("application/json"),
		},
	)
	if err != nil {
		metapanic.ProcessError(metaerror.Wrap(err, "put object to r2 failed"))
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}
