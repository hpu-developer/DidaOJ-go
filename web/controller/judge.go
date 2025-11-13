package controller

import (
	"encoding/json"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	foundationuser "foundation/foundation-user"
	foundationview "foundation/foundation-view"
	"log/slog"
	metacontroller "meta/controller"
	metaerrorcode "meta/error-code"
	metapanic "meta/meta-panic"
	metaredis "meta/meta-redis"
	metaresponse "meta/meta-response"
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
			if judgeJob.Score < 1000 {
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
	list, contestInserter, err := judgeService.GetJudgeList(
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
		HasAuth bool `json:"has_auth"`
		Contest struct {
			Inserter int `json:"inserter"`
		} `json:"contest"`
		List []*foundationview.JudgeJob `json:"list"`
	}{
		HasAuth: true,
		Contest: struct {
			Inserter int `json:"inserter"`
		}{
			Inserter: contestInserter,
		},
		List: list,
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
	if int(language) < 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if len(code) < 10 {
		metaresponse.NewResponse(ctx, weberrorcode.JudgeApproveCodeTooShort, nil)
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

	errorCode := foundationservice.GetJudgeService().IsEnableRemoteJudge(
		problem.OriginOj,
		problem.OriginId,
		language,
		code,
	)
	if errorCode != int(metaerrorcode.Success) {
		metaresponse.NewResponse(ctx, errorCode, nil)
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
		Private(judgeApprove.IsPrivate).
		Status(foundationjudge.JudgeStatusInit).
		Build()
	err = judgeService.InsertJudgeJob(ctx, judgeJob)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, judgeJob)
}

func (c *JudgeController) PostPrivate(ctx *gin.Context) {
	var requestData struct {
		Id        int  `json:"id"`
		IsPrivate bool `json:"is_private"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	id := requestData.Id
	isPrivate := requestData.IsPrivate
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
		// 判断是否为提交者本人
		inserter, err := foundationservice.GetJudgeService().GetJudgeInserter(ctx, id)
		if err != nil {
			metaresponse.NewResponseError(ctx, err)
			return
		}
		if inserter != userId {
			metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
			return
		}
	}
	err = foundationservice.GetJudgeService().SetJudgeJobPrivate(ctx, id, isPrivate)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
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

// GetJudgeReward 获取用户已通过但未获得经验的题目列表
func (c *JudgeController) GetReward(ctx *gin.Context) {
	// 获取当前用户ID
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	// 获取用户尚未获得经验的AC题目（优化为单次数据库查询）
	rewardProblems, err := foundationservice.GetUserService().GetUserUnrewardedACProblems(ctx, userId)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, rewardProblems)
}

// PostReward 为用户发放奖励经验值（按问题）
func (c *JudgeController) PostReward(ctx *gin.Context) {
	// 获取当前用户ID
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	// 获取问题ID参数
	var request struct {
		ProblemId int `json:"id" binding:"required"`
	}
	if err = ctx.ShouldBindJSON(&request); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	// 检查用户是否通过了该问题
	ac, err := foundationservice.GetJudgeService().CheckUserProblemAC(ctx, userId, request.ProblemId)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	if !ac {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	// 获取当前时间
	nowTime := metatime.GetTimeNow()

	// 添加奖励经验值（每个问题只能领取一次）
	hasDuplicate, level, experience, err := foundationservice.GetUserService().AddUserRewardExperience(ctx, userId, request.ProblemId, nowTime)
	if err != nil {
		metaresponse.NewResponseError(ctx, err, nil)
		return
	}

	// 如果已经领取过奖励
	if hasDuplicate {
		metaresponse.NewResponse(ctx, weberrorcode.UserRewardAlreadyDone, nil)
		return
	}

	// 计算当前等级的经验值（当前总经验 - 上一等级的总经验）
	// 计算当前等级升级所需的经验值
	experienceUpgrade := foundationuser.GetExperienceForUpgrade(level)

	// 计算当前等级已获得的经验值
	experienceCurrentLevel := experience
	for i := 1; i < level; i++ {
		experienceCurrentLevel -= foundationuser.GetExperienceForUpgrade(i)
	}

	// 构建响应数据
	responseData := map[string]interface{}{
		"has_duplicate":            hasDuplicate,
		"level":                    level,
		"experience_current_level": experienceCurrentLevel,
		"experience_upgrade":       experienceUpgrade,
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}
