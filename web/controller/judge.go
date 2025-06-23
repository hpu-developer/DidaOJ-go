package controller

import (
	"encoding/json"
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
	metaredis "meta/meta-redis"
	"meta/meta-response"
	metatime "meta/meta-time"
	"strconv"
	"time"
	weberrorcode "web/error-code"
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
		"_id",
		"approve_time",
		"language",
		"score",
		"status",
		"time",
		"memory",
		"author_id",
		"code",
		"code_length",
	}
	if contest == nil {
		fields = append(fields, "problem_id")
	} else {
		fields = append(fields, "contest_id", "contest_problem_index")
	}
	if hasTaskAuth {
		fields = append(fields, "task")
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
	if contest != nil {
		if contest.Type == foundationmodel.ContestTypeAcm {
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
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	list, totalCount, err := judgeService.GetJudgeList(
		ctx, userId,
		contestId, constProblemIndex,
		problemId, username, language, status,
		page, pageSize,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		HasAuth    bool                        `json:"has_auth"`
		TotalCount int                         `json:"total_count"`
		List       []*foundationmodel.JudgeJob `json:"list"`
	}{
		HasAuth:    true,
		TotalCount: totalCount,
		List:       list,
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
	if problemId == "" && (contestId <= 0 || problemIndex <= 0) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	var userId int
	var hasAuth bool
	if problemId != "" {
		userId, hasAuth, err = foundationservice.GetProblemService().CheckSubmitAuth(ctx, problemId)
		if err != nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
			return
		}
		contestId = 0
		problemIndex = 0
	} else {
		problemIdPtr, err := foundationservice.GetContestService().GetProblemIdByContest(ctx, contestId, problemIndex)
		if err != nil {
			metaresponse.NewResponseError(ctx, err)
			return
		}
		if problemIdPtr == nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
			return
		}
		problemId = *problemIdPtr
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

	if problem.OriginId != nil {
		metaresponse.NewResponse(ctx, weberrorcode.JudgeApproveCannotOriginOj, nil)
		return
	}

	judgeService := foundationservice.GetJudgeService()

	nowTime := metatime.GetTimeNow()
	codeLength := len(code)
	judgeJob := foundationmodel.NewJudgeJobBuilder().
		ProblemId(problemId).
		ContestId(contestId).
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
