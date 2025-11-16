package controller

import (
	"strconv"

	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationdao "foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationrun "foundation/foundation-run"
	foundationservice "foundation/foundation-service"
	metacontroller "meta/controller"
	metaerrorcode "meta/error-code"
	metaresponse "meta/meta-response"
	metatime "meta/meta-time"
	weberrorcode "web/error-code"
	"web/request"

	"github.com/gin-gonic/gin"
)

type RunController struct {
	metacontroller.Controller
}

func (c *RunController) Post(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	// 绑定请求参数
	var req request.RunCode
	if err := ctx.ShouldBindJSON(&req); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	// 参数验证
	if len(req.Code) < 10 {
		metaresponse.NewResponse(ctx, weberrorcode.JudgeApproveCodeTooShort, nil)
		return
	}

	language := foundationjudge.JudgeLanguage(req.Language)

	// 检查语言是否有效
	if !foundationjudge.IsValidJudgeLanguage(int(language)) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	nowTime := metatime.GetTimeNow()

	// 创建运行任务记录
	runJob := foundationmodel.NewRunJobBuilder().
		Inserter(userId).
		InsertTime(nowTime).
		Language(language).
		Code(req.Code).
		Input(req.Input).
		Status(foundationrun.RunStatusInit). // 等待状态
		Build()

	// 保存到数据库
	runJobService := foundationservice.GetRunJobService()
	if err := runJobService.AddRunJob(ctx, runJob); err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, runJob.Id)
}

// GetStatus 获取运行状态
func (c *RunController) GetStatus(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	idStr := ctx.Query("id")
	if idStr == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	// 获取运行任务记录
	runJob, err := foundationdao.GetRunJobDao().GetRunJob(ctx, id, userId)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, runJob)
}
