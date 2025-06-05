package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	"meta/error-code"
	metapanic "meta/meta-panic"
	"meta/meta-response"
	metastring "meta/meta-string"
	metatime "meta/meta-time"
	"meta/set"
	"strconv"
	"time"
	weberrorcode "web/error-code"
	"web/request"
)

type ContestController struct {
	metacontroller.Controller
}

func (c *ContestController) Get(ctx *gin.Context) {
	contestService := foundationservice.GetContestService()
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
	// TODO 判断权限
	contest, err := contestService.GetContest(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if contest == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, contest)
}

func (c *ContestController) GetEdit(ctx *gin.Context) {
	contestService := foundationservice.GetContestService()
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
	// TODO 判断权限
	contest, problems, err := contestService.GetContestEdit(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if contest == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	responseData := struct {
		Contest  *foundationmodel.Contest `json:"contest"`
		Problems []string                 `json:"problems"` // 题目索引列表
	}{
		Contest:  contest,
		Problems: problems,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ContestController) GetList(ctx *gin.Context) {
	contestService := foundationservice.GetContestService()
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
	var list []*foundationmodel.Contest
	var totalCount int
	list, totalCount, err = contestService.GetContestList(ctx, page, pageSize)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Time       time.Time                  `json:"time"`
		TotalCount int                        `json:"total_count"`
		List       []*foundationmodel.Contest `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ContestController) GetProblem(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	contestId, err := strconv.Atoi(id)
	if err != nil || contestId <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemIndexStr := ctx.Query("problem_index")
	if problemIndexStr == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemIndex, err := strconv.Atoi(problemIndexStr)
	if err != nil || problemIndex <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	_, hasAuth, err := foundationservice.GetContestService().CheckEditAuth(ctx, contestId)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	problemId, err := foundationservice.GetContestService().GetProblemIdByContest(ctx, contestId, problemIndex)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, problemId)
}

func (c *ContestController) GetProblemList(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	contestId, err := strconv.Atoi(id)
	if err != nil || contestId <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	_, hasAuth, err := foundationservice.GetContestService().CheckSubmitAuth(ctx, contestId)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	problems, err := foundationservice.GetContestService().GetContestProblems(
		ctx,
		contestId,
	)
	requestData := struct {
		Problems []int `json:"problems"` // 题目索引
	}{
		Problems: problems,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, requestData)
}

func (c *ContestController) GetRank(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	contestId, err := strconv.Atoi(id)
	if err != nil || contestId <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	contest, problems, ranks, err := foundationservice.GetContestService().GetContestRanks(ctx, contestId)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	if contest == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	responseData := struct {
		Contest  *foundationmodel.ContestViewRank `json:"contest"`
		Problems []int                            `json:"problems"` // 题目索引列表
		Ranks    []*foundationmodel.ContestRank   `json:"ranks"`
	}{
		Contest:  contest,
		Problems: problems,
		Ranks:    ranks,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ContestController) PostCreate(ctx *gin.Context) {
	var requestData request.ContestEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if requestData.Title == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	startTime := requestData.StartTime
	endTime := requestData.EndTime
	if endTime.Before(startTime) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	nowTime := metatime.GetTimeNow()

	contestService := foundationservice.GetContestService()
	// 控制创建时的标题唯一，一定程度上防止误重复创建
	ok, err := contestService.HasContestTitle(ctx, userId, requestData.Title)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if ok {
		metaresponse.NewResponse(ctx, weberrorcode.ContestTitleDuplicate, nil)
		return
	}
	memberIds, err := foundationservice.GetUserService().FilterValidUserIds(ctx, requestData.Members)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	//requestData.Problems去重
	if len(requestData.Problems) == 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 去重
	problemIds := metastring.RemoveDuplicate(requestData.Problems)
	validProblemIds, err := foundationservice.GetProblemService().FilterValidProblemIds(ctx, problemIds)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	validProblemIdSet := set.FromSlice(validProblemIds)
	var realProblemIds []string
	// 保持输入的problemIds顺序
	for _, problemId := range problemIds {
		if validProblemIdSet.Contains(problemId) {
			realProblemIds = append(realProblemIds, problemId)
		}
	}
	if len(realProblemIds) < 1 {
		metaresponse.NewResponse(ctx, weberrorcode.ContestNotFoundProblem)
		return
	}
	if len(realProblemIds) > 52 {
		metaresponse.NewResponse(ctx, weberrorcode.ContestTooManyProblem)
		return
	}

	var problems []*foundationmodel.ContestProblem
	for _, problemId := range realProblemIds {
		problems = append(
			problems, foundationmodel.NewContestProblemBuilder().
				ProblemId(problemId).
				ViewId(nil). // 题目描述Id，默认为nil
				Score(0). // 分数默认为0
				Index(len(problems)+1). // 索引从1开始
				Build(),
		)
	}

	private := requestData.Private

	var lockRankDuration *time.Duration
	if requestData.LockRankDuration > 0 {
		lockRankDurationPtr := time.Duration(requestData.LockRankDuration) * time.Second
		lockRankDuration = &lockRankDurationPtr
	}

	contest := foundationmodel.NewContestBuilder().
		Title(requestData.Title).
		Description(requestData.Description).
		Notification(requestData.Notification).
		StartTime(startTime).
		EndTime(endTime).
		OwnerId(userId).
		CreateTime(nowTime).
		UpdateTime(nowTime).
		Problems(problems).
		Private(private).
		Members(memberIds).
		LockRankDuration(lockRankDuration).
		AlwaysLock(requestData.AlwaysLock).
		SubmitAnytime(requestData.SubmitAnytime).
		Build()

	err = contestService.InsertContest(ctx, contest)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, contest)
}

func (c *ContestController) PostEdit(ctx *gin.Context) {
	var requestData request.ContestEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if requestData.Title == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	startTime := requestData.StartTime
	endTime := requestData.EndTime
	if endTime.Before(startTime) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	nowTime := metatime.GetTimeNow()

	_, hasAuth, err := foundationservice.GetContestService().CheckEditAuth(ctx, requestData.Id)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	memberIds, err := foundationservice.GetUserService().FilterValidUserIds(ctx, requestData.Members)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	//requestData.Problems去重
	if len(requestData.Problems) == 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 去重
	problemIds := metastring.RemoveDuplicate(requestData.Problems)
	validProblemIds, err := foundationservice.GetProblemService().FilterValidProblemIds(ctx, problemIds)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	validProblemIdSet := set.FromSlice(validProblemIds)
	var realProblemIds []string
	// 保持输入的problemIds顺序
	for _, problemId := range problemIds {
		if validProblemIdSet.Contains(problemId) {
			realProblemIds = append(realProblemIds, problemId)
		}
	}
	if len(realProblemIds) < 1 {
		metaresponse.NewResponse(ctx, weberrorcode.ContestNotFoundProblem)
		return
	}
	if len(realProblemIds) > 52 {
		metaresponse.NewResponse(ctx, weberrorcode.ContestTooManyProblem)
		return
	}

	contestService := foundationservice.GetContestService()

	var problems []*foundationmodel.ContestProblem
	for _, problemId := range realProblemIds {
		problems = append(
			problems, foundationmodel.NewContestProblemBuilder().
				ProblemId(problemId).
				ViewId(nil). // 题目描述Id，默认为nil
				Score(0). // 分数默认为0
				Index(len(problems)+1). // 索引从1开始
				Build(),
		)
	}

	private := requestData.Private

	var lockRankDuration *time.Duration
	if requestData.LockRankDuration > 0 {
		lockRankDurationPtr := time.Duration(requestData.LockRankDuration) * time.Second
		lockRankDuration = &lockRankDurationPtr
	}

	contest := foundationmodel.NewContestBuilder().
		Title(requestData.Title).
		Description(requestData.Description).
		Notification(requestData.Notification).
		StartTime(startTime).
		EndTime(endTime).
		OwnerId(userId).
		CreateTime(nowTime).
		UpdateTime(nowTime).
		Problems(problems).
		Private(private).
		Members(memberIds).
		LockRankDuration(lockRankDuration).
		AlwaysLock(requestData.AlwaysLock).
		SubmitAnytime(requestData.SubmitAnytime).
		Build()

	err = contestService.UpdateContest(ctx, requestData.Id, contest)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, contest.UpdateTime)
}
