package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	"meta/error-code"
	metapanic "meta/meta-panic"
	metatime "meta/meta-time"
	"meta/response"
	"strconv"
	"time"
	"web/request"
)

type JudgeController struct {
	metacontroller.Controller
}

func (c *JudgeController) PostApprove(ctx *gin.Context) {
	var judgeApprove request.JudgeApprove
	err := ctx.ShouldBindJSON(&judgeApprove)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemId := judgeApprove.ProblemId
	language := judgeApprove.Language
	code := judgeApprove.Code
	if problemId == "" || int(language) < 0 || code == "" {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	problem, err := foundationservice.GetProblemService().GetProblem(ctx, problemId)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if problem == nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	judgeService := foundationservice.GetJudgeService()

	nowTime := metatime.GetTimeNow()
	codeLength := len(code)
	judgeJob := foundationmodel.NewJudgeJobBuilder().
		ProblemId(problemId).
		Author(userId).
		ApproveTime(nowTime).
		Language(language).
		Code(code).
		CodeLength(codeLength).
		Status(foundationjudge.JudgeStatusInit).
		Build()
	err = judgeService.InsertJudgeJob(ctx, judgeJob)
	if err != nil {
		response.NewResponseError(ctx, err)
		return
	}
	response.NewResponse(ctx, metaerrorcode.Success, judgeJob)
}

func (c *JudgeController) Get(ctx *gin.Context) {
	judgeService := foundationservice.GetJudgeService()
	idStr := ctx.Query("id")
	if idStr == "" {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	judgeJob, err := judgeService.GetJudge(ctx, id)
	if err != nil {
		response.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if judgeJob == nil {
		response.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	response.NewResponse(ctx, metaerrorcode.Success, judgeJob)
}

func (c *JudgeController) GetList(ctx *gin.Context) {
	judgeService := foundationservice.GetJudgeService()
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if pageSize != 50 && pageSize != 100 {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemId := ctx.Query("problem_id")
	username := ctx.Query("username")
	languageStr := ctx.Query("language")
	language := foundationjudge.JudgeLanguageUnknown
	if languageStr != "" {
		languageInt, err := strconv.Atoi(languageStr)
		if err == nil && foundationjudge.IsValidJudgeLanguage(languageInt) {
			language = foundationjudge.JudgeLanguage(languageInt)
		}
	}
	statusStr := ctx.Query("status")
	status := foundationjudge.JudgeStatusUnknown
	if statusStr != "" {
		statusInt, err := strconv.Atoi(statusStr)
		if err == nil && foundationjudge.IsValidJudgeStatus(statusInt) {
			status = foundationjudge.JudgeStatus(statusInt)
		}
	}
	list, totalCount, err := judgeService.GetJudgeList(ctx, problemId, username, language, status, page, pageSize)
	if err != nil {
		response.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Time       time.Time                   `json:"time"`
		TotalCount int                         `json:"total_count"`
		List       []*foundationmodel.JudgeJob `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	response.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *JudgeController) PostRejudgeRecently(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	ok, err := foundationservice.GetUserService().CheckUserAuths(ctx, userId, []string{"i-manage-judge"})
	if err != nil {
		metapanic.ProcessError(err)
		response.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		response.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	err = foundationservice.GetJudgeService().RejudgeRecently(ctx)
	if err != nil {
		response.NewResponseError(ctx, err)
		return
	}

	response.NewResponse(ctx, metaerrorcode.Success)
}
