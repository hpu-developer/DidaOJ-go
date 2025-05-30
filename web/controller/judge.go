package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationcontest "foundation/foundation-contest"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/gin-gonic/gin"
	"log/slog"
	metacontroller "meta/controller"
	"meta/error-code"
	metapanic "meta/meta-panic"
	"meta/meta-response"
	metatime "meta/meta-time"
	"strconv"
	"time"
	"web/request"
)

type JudgeController struct {
	metacontroller.Controller
}

func (c *JudgeController) Get(ctx *gin.Context) {
	judgeService := foundationservice.GetJudgeService()
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
	judgeJob, err := judgeService.GetJudge(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if judgeJob == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, judgeJob)
}

func (c *JudgeController) GetCode(ctx *gin.Context) {
	judgeService := foundationservice.GetJudgeService()
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
	language, jobCode, err := judgeService.GetJudgeCode(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Language foundationjudge.JudgeLanguage `json:"language"`
		Code     *string                       `json:"code"`
	}{
		Language: language,
		Code:     jobCode,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *JudgeController) GetList(ctx *gin.Context) {
	judgeService := foundationservice.GetJudgeService()
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "50")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if pageSize != 50 && pageSize != 100 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemId := ctx.Query("problem_id")
	var contestId, constProblemIndex int
	contestIdStr := ctx.Query("contest_id")
	if contestIdStr != "" {
		contestId, err = strconv.Atoi(contestIdStr)
		constProblemIndex = foundationcontest.GetContestProblemIndex(problemId)
	}
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
	status := foundationjudge.JudgeStatusMax
	if statusStr != "" {
		statusInt, err := strconv.Atoi(statusStr)
		if err == nil && foundationjudge.IsValidJudgeStatus(statusInt) {
			status = foundationjudge.JudgeStatus(statusInt)
		}
	}
	list, totalCount, err := judgeService.GetJudgeList(
		ctx,
		contestId, constProblemIndex,
		problemId, username, language, status,
		page, pageSize,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
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
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *JudgeController) PostApprove(ctx *gin.Context) {
	var judgeApprove request.JudgeApprove
	err := ctx.ShouldBindJSON(&judgeApprove)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemId := judgeApprove.ProblemId
	language := judgeApprove.Language
	code := judgeApprove.Code
	if problemId == "" || int(language) < 0 || code == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	ok, err := foundationservice.GetProblemService().HasProblem(ctx, problemId)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	judgeService := foundationservice.GetJudgeService()

	nowTime := metatime.GetTimeNow()
	codeLength := len(code)
	judgeJob := foundationmodel.NewJudgeJobBuilder().
		ProblemId(problemId).
		AuthorId(userId).
		ApproveTime(nowTime).
		Language(language).
		Code(code).
		CodeLength(codeLength).
		Status(foundationjudge.JudgeStatusInit).
		Build()
	err = judgeService.InsertJudgeJob(ctx, judgeJob)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, judgeJob)
}

func (c *JudgeController) PostRejudge(ctx *gin.Context) {
	var requestData struct {
		Id int `json:"id"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	id := requestData.Id
	if id <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	ok, err := foundationservice.GetUserService().CheckUserAuthByUserId(ctx, userId, foundationauth.AuthTypeManageJudge)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	err = foundationservice.GetJudgeService().RejudgeJob(ctx, id)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *JudgeController) PostRejudgeSearch(ctx *gin.Context) {
	var requestData struct {
		ProblemId string                        `json:"problem_id"`
		Language  foundationjudge.JudgeLanguage `json:"language"`
		Status    foundationjudge.JudgeStatus   `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemId := requestData.ProblemId
	language := requestData.Language
	status := requestData.Status
	if problemId == "" &&
		!foundationjudge.IsValidJudgeLanguage(int(language)) &&
		!foundationjudge.IsValidJudgeStatus(int(status)) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	ok, err := foundationservice.GetUserService().CheckUserAuthByUserId(ctx, userId, foundationauth.AuthTypeManageJudge)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	err = foundationservice.GetJudgeService().PostRejudgeSearch(ctx, problemId, language, status)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *JudgeController) PostRejudgeRecently(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	ok, err := foundationservice.GetUserService().CheckUserAuthByUserId(ctx, userId, foundationauth.AuthTypeManageJudge)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	err = foundationservice.GetJudgeService().RejudgeRecently(ctx)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *JudgeController) PostRejudgeAll(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	ok, err := foundationservice.GetUserService().CheckUserAuthByUserId(ctx, userId, foundationauth.AuthTypeManageJudge)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	slog.Warn("Rejudge all judge jobs", "userId", userId)

	err = foundationservice.GetJudgeService().RejudgeAll(ctx)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}
