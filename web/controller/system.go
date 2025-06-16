package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log"
	cfr2 "meta/cf-r2"
	metacontroller "meta/controller"
	"meta/error-code"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	"meta/meta-response"
	metasystem "meta/meta-system"
	"net/http"
	"time"
	"web/service"
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

func (c *SystemController) GetImageToken(ctx *gin.Context) {
	_, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageWeb)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	// 获取 R2 客户端
	r2Client := cfr2.GetSubsystem().GetClient("didapipa-oj")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	// 配置参数
	bucketName := "didapipa-oj"
	objectKey := fmt.Sprintf("uploading/system/%d_%s", time.Now().Unix(), uuid.New().String())

	// 设置 URL 有效期
	req, _ := r2Client.PutObjectRequest(
		&s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
		},
	)

	// 设置 URL 有效时间，例如 15 分钟
	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {
		log.Printf("Failed to sign request: %v", err)
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	// 返回上传信息
	metaresponse.NewResponse(
		ctx, metaerrorcode.Success, gin.H{
			"upload_url":  urlStr,
			"preview_url": metahttp.UrlJoin("https://r2-oj.didapipa.com", objectKey),
		},
	)
}

func (c *SystemController) PostNotification(ctx *gin.Context) {
	var requestData struct {
		Theme   string `json:"theme" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError)
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

func (c *SystemController) PostAnnouncement(ctx *gin.Context) {
	var requestData struct {
		Title   string `json:"title"`
		Content string `json:"content" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError)
		return
	}

	var oldDescription string
	r2Url := "https://r2-oj.didapipa.com/system/notification.json" + "?" + time.Now().Format("20060102150405")
	resp, err := http.Get(r2Url)
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "close response body failed"))
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError)
			return
		}
	}(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		oldDescription = ""
	} else if resp.StatusCode == http.StatusOK {
		oldDescription = string(resp.Body)
	} else {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError)
		return
	}

	description, needUpdateUrls, err := service.GetR2ImageService().ProcessContentFromMarkdown(
		requestData.Content,
		oldDescription,
		metahttp.UrlJoin("problem", problemId),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	r2Client := cfr2.GetSubsystem().GetClient("didapipa-oj")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError)
		return
	}
	// 构建 Judger 状态 JSON 数据
	statusJsonData := struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}{
		Title:   requestData.Title,
		Content: requestData.Content,
	}
	statusBytes, err := json.Marshal(statusJsonData)
	if err != nil {
		metapanic.ProcessError(metaerror.Wrap(err, "marshal status data failed"))
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError)
		return
	}
	key := "system/announcement.json"
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
