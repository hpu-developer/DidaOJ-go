package controller

import (
	"encoding/json"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	foundationview "foundation/foundation-view"
	"log/slog"
	metacontroller "meta/controller"
	"meta/error-code"
	metapanic "meta/meta-panic"
	metaredis "meta/meta-redis"
	"meta/meta-response"
	metatime "meta/meta-time"
	"strconv"
	"time"
	weberrorcode "web/error-code"
	"web/request"

	"github.com/gin-gonic/gin"
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
	_, hasAuth, hasTaskAuth, contest, err := foundationservice.GetJudgeService().CheckJudgeViewAuth(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if contest != nil {
		contestIdStr := ctx.Query("contest_id")
		if contestIdStr == "" {
			metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
			return
		}
		contestId, err := strconv.Atoi(contestIdStr)
		if err != nil || contestId != contest.Id {
			metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
			return
		}
	}
	fields := []string{
		"id",
		"problem_id",
		"inserter",
		"insert_time",
		"language",
		"code",
		"code_length",
		"status",
		"judger",
		"judge_time",
		"task_current",
		"task_total",
		"score",
		"time",
		"memory",
		"private",
	}
	judgeJob, err := judgeService.GetJudge(ctx, id, fields)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if judgeJob == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	if hasTaskAuth {
		judgeJob.Task, err = foundationservice.GetJudgeService().GetJudgeTaskList(ctx, id)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
	}
	if contest != nil {
		judgeJob.ContestProblemIndex, err = foundationservice.GetContestService().GetContestProblemIndexById(
			ctx,
			contest.Id,
			judgeJob.ProblemId,
		)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
		judgeJob.ProblemId = 0
		if contest.Type == foundationenum.ContestTypeAcm {
			// IOI模式之外隐藏分数信息
			if judgeJob.Score < 100 {
				judgeJob.Score = 0
			}
		}
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
	_, hasAuth, _, _, err := foundationservice.GetJudgeService().CheckJudgeViewAuth(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
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
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
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
	if page < 1 || pageSize != 20 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if page > 10 {
		metaresponse.NewResponse(ctx, weberrorcode.JudgeListTooManySkip, nil)
		return
	}
	if page*pageSize > 500 {
		metaresponse.NewResponse(ctx, weberrorcode.JudgeListTooManySkip, nil)
		return
	}
	contestIdStr := ctx.Query("contest_id")
	var problemKey string
	var contestId int
	if contestIdStr != "" {
		contestId, err = strconv.Atoi(contestIdStr)
		if err != nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		_, hasAuth, err := foundationservice.GetContestService().CheckViewAuth(ctx, contestId)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
		if !hasAuth {
			responseData := struct {
				HasAuth bool `json:"has_auth"`
			}{
				HasAuth: false,
			}
			metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
			return
		}
	}
	problemKey = ctx.Query("problem_key")
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

	userId, err := foundationauth.GetUserIdFromContext(ctx)
	list, err := judgeService.GetJudgeList(
		ctx, userId,
		problemKey, contestId,
		username, language, status,
		page, pageSize,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		HasAuth bool                       `json:"has_auth"`
		List    []*foundationview.JudgeJob `json:"list"`
	}{
		HasAuth: true,
		List:    list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *JudgeController) GetStaticsRecently(ctx *gin.Context) {
	codeKey := "judge_statics_recently"
	redisClient := metaredis.GetSubsystem().GetClient()
	cached, err := redisClient.Get(ctx, codeKey).Result()
	if err == nil && cached != "" {
		var statics interface{}
		if err := json.Unmarshal([]byte(cached), &statics); err == nil {
			metaresponse.NewResponse(ctx, metaerrorcode.Success, statics)
			return
		}
	}
	judgeService := foundationservice.GetJudgeService()
	statics, err := judgeService.GetJudgeJobCountStaticsRecently(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	// 缓存数据（序列化为 JSON）并设置 1 分钟过期
	bytes, err := json.Marshal(statics)
	if err == nil {
		redisClient.Set(ctx, codeKey, string(bytes), time.Minute)
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, statics)
}

func (c *JudgeController) PostApprove(ctx *gin.Context) {
	var judgeApprove request.JudgeApprove
	err := ctx.ShouldBindJSON(&judgeApprove)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	language := judgeApprove.Language
	code := judgeApprove.Code
	if int(language) < 0 || code == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemId := judgeApprove.ProblemId
	contestId := judgeApprove.ContestId
	problemIndex := judgeApprove.ProblemIndex
	if problemId == 0 && (contestId <= 0 || problemIndex <= 0) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	var userId int
	var hasAuth bool
	if problemId > 0 {
		userId, hasAuth, err = foundationservice.GetProblemService().CheckSubmitAuth(ctx, problemId)
		if err != nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
			return
		}
		if userId <= 0 {
			metaresponse.NewResponse(ctx, foundationerrorcode.NeedLogin, nil)
			return
		}
		contestId = 0
		problemIndex = 0
	} else {
		userId, err = foundationauth.GetUserIdFromContext(ctx)
		if err != nil || userId <= 0 {
			metaresponse.NewResponse(ctx, foundationerrorcode.NeedLogin, nil)
			return
		}
		problemId, err = foundationservice.GetContestService().GetProblemIdByContestIndex(
			ctx,
			contestId,
			problemIndex,
		)
		if err != nil {
			metaresponse.NewResponseError(ctx, err)
			return
		}
		if problemId <= 0 {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		userId, hasAuth, err = foundationservice.GetContestService().CheckSubmitAuth(
			ctx,
			contestId,
		)
		if err != nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
			return
		}
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, weberrorcode.JudgeJobCannotApprove, nil)
		return
	}

	problem, err := foundationservice.GetProblemService().GetProblemViewApproveJudge(ctx, problemId)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	if problem == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	if !foundationservice.GetJudgeService().IsEnableRemoteJudge(problem.OriginOj, problem.OriginId, language) {
		metaresponse.NewResponse(ctx, weberrorcode.JudgeApproveCannotOriginOj, nil)
		return
	}

	judgeService := foundationservice.GetJudgeService()

	nowTime := metatime.GetTimeNow()
	codeLength := len(code)
	judgeJob := foundationmodel.NewJudgeJobBuilder().
		ProblemId(problemId).
		ContestId(contestId).
		Inserter(userId).
		InsertTime(nowTime).
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
		ProblemKey string                        `json:"problem_key"`
		Language   foundationjudge.JudgeLanguage `json:"language"`
		Status     foundationjudge.JudgeStatus   `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	ProblemKey := requestData.ProblemKey
	language := requestData.Language
	status := requestData.Status
	problemId := 0
	if ProblemKey == "" &&
		!foundationjudge.IsValidJudgeLanguage(int(language)) &&
		!foundationjudge.IsValidJudgeStatus(int(status)) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	} else {
		var err error
		problemId, err = foundationservice.GetProblemService().GetProblemIdByKey(ctx, ProblemKey)
		if err != nil {
			metaresponse.NewResponseError(ctx, err)
			return
		}
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
	slog.Info(
		"Rejudge search judge jobs",
		"userId",
		userId,
		"problemId",
		problemId,
		"language",
		language,
		"status",
		status,
	)
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
