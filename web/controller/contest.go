package controller

import (
	"errors"
	"fmt"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationmodel "foundation/foundation-model"
	foundationr2 "foundation/foundation-r2"
	foundationservice "foundation/foundation-service"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	cfr2 "meta/cf-r2"
	metacontroller "meta/controller"
	"meta/error-code"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	"meta/meta-response"
	metastring "meta/meta-string"
	metatime "meta/meta-time"
	"meta/set"
	"strconv"
	"time"
	weberrorcode "web/error-code"
	"web/request"
	"web/service"
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
	nowTime := metatime.GetTimeNow()
	contest, hasAuth, needPassword, attemptStatus, err := contestService.GetContest(ctx, id, nowTime)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if contest == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	responseData := struct {
		Now           time.Time                                    `json:"now"`
		HasAuth       bool                                         `json:"has_auth"`
		NeedPassword  bool                                         `json:"need_password,omitempty"` // 是否需要密码
		Contest       *foundationmodel.Contest                     `json:"contest"`
		AttemptStatus map[int]foundationmodel.ProblemAttemptStatus `json:"attempt_status,omitempty"`
	}{
		Now:           nowTime,
		HasAuth:       hasAuth,
		NeedPassword:  needPassword,
		Contest:       contest,
		AttemptStatus: attemptStatus,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
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
	_, hasAuth, err := foundationservice.GetContestService().CheckEditAuth(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
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
	title := ctx.Query("title")
	username := ctx.Query("username")
	var list []*foundationmodel.Contest
	var totalCount int
	list, totalCount, err = contestService.GetContestList(ctx, title, username, page, pageSize)
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
	startTime, err := foundationservice.GetContestService().GetContestStartTime(ctx, contestId)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	if startTime == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	nowTime := metatime.GetTimeNow()
	if startTime.After(nowTime) {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	_, hasAuth, err := foundationservice.GetContestService().CheckViewAuth(ctx, contestId)
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
	_, hasAuth, err := foundationservice.GetContestService().CheckViewAuth(ctx, contestId)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
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

	nowTime := metatime.GetTimeNow()

	var contest *foundationmodel.ContestViewRank
	var problems []int
	var isLocked bool
	var ranks []*foundationmodel.ContestRank
	contest, problems, ranks, isLocked, err = foundationservice.GetContestService().GetContestRanks(
		ctx,
		contestId,
		nowTime,
	)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	if contest == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	responseData := struct {
		HasAuth  bool                             `json:"has_auth"`
		Now      time.Time                        `json:"now"`
		IsLocked bool                             `json:"is_locked"` // 是否锁榜状态
		Contest  *foundationmodel.ContestViewRank `json:"contest"`
		Problems []int                            `json:"problems"` // 题目索引列表
		Ranks    []*foundationmodel.ContestRank   `json:"ranks"`
	}{
		HasAuth:  true,
		Now:      nowTime,
		IsLocked: isLocked,
		Contest:  contest,
		Problems: problems,
		Ranks:    ranks,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ContestController) GetImageToken(ctx *gin.Context) {
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
	_, err = foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	// 获取 R2 客户端
	r2Client := cfr2.GetSubsystem().GetClient("didapipa-oj")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	var key string
	if id > 0 {
		key = strconv.Itoa(id)
	}

	// 配置参数
	bucketName := "didapipa-oj"
	objectKey := metahttp.UrlJoin(
		"uploading/contest",
		key,
		fmt.Sprintf("%d_%s", time.Now().Unix(), uuid.New().String()),
	)

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
				Weight(0). // 分数默认为0
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

	var password *string
	if !metastring.IsEmpty(requestData.Password) {
		password = requestData.Password
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
		Password(password).
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

	var description string
	var needUpdateUrls []*foundationr2.R2ImageUrl
	description, needUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		requestData.Description,
		nil,
		metahttp.UrlJoin("contest"),
		metahttp.UrlJoin("contest", strconv.Itoa(contest.Id)),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if len(needUpdateUrls) > 0 {
		err := foundationservice.GetContestService().UpdateDescription(ctx, contest.Id, description)
		if err != nil {
			return
		}
		err = service.GetR2ImageService().MoveImageAfterSave(needUpdateUrls)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}
	contest.Description = description

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

	contestId := requestData.Id

	oldDescription, err := contestService.GetContestDescription(ctx, contestId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
			return
		}
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	var needUpdateUrls []*foundationr2.R2ImageUrl
	requestData.Description, needUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		requestData.Description,
		oldDescription,
		metahttp.UrlJoin("contest", strconv.Itoa(contestId)),
		metahttp.UrlJoin("contest", strconv.Itoa(contestId)),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	var problems []*foundationmodel.ContestProblem
	for _, problemId := range realProblemIds {
		problems = append(
			problems, foundationmodel.NewContestProblemBuilder().
				ProblemId(problemId).
				ViewId(nil). // 题目描述Id，默认为nil
				Weight(0). // 分数默认为0
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

	var password *string
	if !metastring.IsEmpty(requestData.Password) {
		password = requestData.Password
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
		Password(password).
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

	// 因为数据库已经保存了，因此即使图片失败这里也返回成功
	err = service.GetR2ImageService().MoveImageAfterSave(needUpdateUrls)
	if err != nil {
		metapanic.ProcessError(err)
	}

	responseData := struct {
		Description string     `json:"description"`
		UpdateTime  *time.Time `json:"update_time"`
	}{
		Description: requestData.Description,
		UpdateTime:  &contest.UpdateTime,
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ContestController) PostDolos(ctx *gin.Context) {
	var requestData struct {
		Id int `json:"id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	_, hasAuth, err := foundationservice.GetContestService().CheckEditAuth(ctx, requestData.Id)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	url, err := foundationservice.GetContestService().DolosContest(ctx, requestData.Id)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, url)
}

func (c *ContestController) PostPassword(ctx *gin.Context) {
	var requestData struct {
		ContestId int    `json:"id" binding:"required"`
		Password  string `json:"password" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, hasAuth, err := foundationservice.GetContestService().CheckViewAuth(ctx, requestData.ContestId)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if hasAuth {
		metaresponse.NewResponse(ctx, metaerrorcode.Success, nil)
		return
	}
	if userId <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.NeedLogin, nil)
		return
	}
	success, err := foundationservice.GetContestService().PostPassword(
		ctx,
		userId,
		requestData.ContestId,
		requestData.Password,
	)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	if !success {
		metaresponse.NewResponse(ctx, weberrorcode.ContestPostPasswordError, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, nil)
}
