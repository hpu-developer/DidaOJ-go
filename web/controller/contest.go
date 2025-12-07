package controller

import (
	"encoding/json"
	"fmt"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationr2 "foundation/foundation-r2"
	foundationservice "foundation/foundation-service"
	foundationview "foundation/foundation-view"
	"io"
	"log"
	cfr2 "meta/cf-r2"
	metacontroller "meta/controller"
	metaerrorcode "meta/error-code"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	metaresponse "meta/meta-response"
	metaslice "meta/meta-slice"
	metastring "meta/meta-string"
	metatime "meta/meta-time"
	"meta/set"
	"net/http"
	"sort"
	"strconv"
	"time"
	weberrorcode "web/error-code"
	"web/request"
	"web/service"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		Now           time.Time                                   `json:"now"`
		HasAuth       bool                                        `json:"has_auth"`
		NeedPassword  bool                                        `json:"need_password,omitempty"` // 是否需要密码
		Contest       *foundationview.ContestDetail               `json:"contest"`
		AttemptStatus map[int]foundationenum.ProblemAttemptStatus `json:"attempt_status,omitempty"`
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
	contest, err := contestService.GetContestEdit(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if contest == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	responseData := struct {
		Contest *foundationview.ContestDetailEdit `json:"contest"`
	}{
		Contest: contest,
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
	var list []*foundationview.ContestList
	var totalCount int
	list, totalCount, err = contestService.GetContestList(ctx, title, username, page, pageSize)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Time       time.Time                     `json:"time"`
		TotalCount int                           `json:"total_count"`
		List       []*foundationview.ContestList `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ContestController) GetRecently(ctx *gin.Context) {
	contestList := []*foundationview.ContestRemoteList{}

	codeforceUrl := "https://codeforces.com/api/contest.list?gym=false"
	resp, err := http.Get(codeforceUrl)
	if err == nil {
		defer resp.Body.Close()
		// 把codeforce的json解析到contestList
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			metaresponse.NewResponseError(ctx, err)
			return
		}
		var codeforceResp struct {
			Status string `json:"status"`
			Result []struct {
				Id                  int    `json:"id"`
				Name                string `json:"name"`
				Type                string `json:"type"`
				Phase               string `json:"phase"`
				Frozen              bool   `json:"frozen"`
				DurationSeconds     int    `json:"durationSeconds"`
				StartTimeSeconds    int64  `json:"startTimeSeconds"`
				RelativeTimeSeconds int64  `json:"relativeTimeSeconds"`
			} `json:"result"`
		}
		if err := json.Unmarshal(body, &codeforceResp); err == nil {
			for _, item := range codeforceResp.Result {
				contestList = append(contestList, &foundationview.ContestRemoteList{
					Title:     item.Name,
					StartTime: metatime.GetTimeByTimestamp(item.StartTimeSeconds),
					EndTime:   metatime.GetTimeByTimestamp(item.StartTimeSeconds + int64(item.DurationSeconds)),
					Status:    item.Phase,
					Type:      item.Type,
					Source:    "codeforce",
					Link:      fmt.Sprintf("https://codeforces.com/contests/%d", item.Id),
				})
			}
		}
	}

	algcontestUrl := "https://algcontest.rainng.com/contests"
	resp, err = http.Get(algcontestUrl)
	if err == nil {
		defer resp.Body.Close()
		// 把algcontest的json解析到contestList
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			metaresponse.NewResponseError(ctx, err)
			return
		}
		var algcontestResp []struct {
			Oj             string `json:"oj"`
			Name           string `json:"name"`
			StartTimeStamp int64  `json:"startTimeStamp"`
			EndTimeStamp   int64  `json:"endTimeStamp"`
			Status         string `json:"status"`
			OiContest      bool   `json:"oiContest"`
			Link           string `json:"link"`
		}
		if err := json.Unmarshal(body, &algcontestResp); err == nil {
			for _, item := range algcontestResp {
				contest := &foundationview.ContestRemoteList{
					Title:     item.Name,
					StartTime: metatime.GetTimeByTimestamp(item.StartTimeStamp),
					EndTime:   metatime.GetTimeByTimestamp(item.EndTimeStamp),
					Status:    item.Status,
					Source:    item.Oj,
					Link:      item.Link,
				}
				if item.OiContest {
					contest.Type = "oi"
				} else {
					contest.Type = "acm"
				}
				contestList = append(contestList, contest)
			}
		}
	}

	// 对contestList根据start_time排序
	sort.Slice(contestList, func(i, j int) bool {
		return contestList[i].StartTime.After(contestList[j].StartTime)
	})

	// 仅保留前20个
	if len(contestList) > 20 {
		contestList = contestList[:20]
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, contestList)
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
	problemId, err := foundationservice.GetContestService().GetProblemKeyByContestIndex(ctx, contestId, problemIndex)
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

	var contest *foundationview.ContestRankDetail
	var ranks []*foundationview.ContestRank
	var isLocked bool
	contest, ranks, isLocked, err = foundationservice.GetContestService().GetContestRanks(
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
		HasAuth  bool                              `json:"has_auth"`
		Now      time.Time                         `json:"now"`
		IsLocked bool                              `json:"is_locked"` // 是否锁榜状态
		Contest  *foundationview.ContestRankDetail `json:"contest"`
		Ranks    []*foundationview.ContestRank     `json:"ranks"` // 排行榜
	}{
		HasAuth:  true,
		Now:      nowTime,
		IsLocked: isLocked,
		Contest:  contest,
		Ranks:    ranks,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ContestController) GetStatistics(ctx *gin.Context) {
	contestService := foundationservice.GetContestService()
	idStr := ctx.Query("id")
	if idStr == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	contestId, err := strconv.Atoi(idStr)
	if err != nil {
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
	languageStr := ctx.Query("language")
	language := foundationjudge.JudgeLanguageUnknown
	if languageStr != "" {
		languageInt, err := strconv.Atoi(languageStr)
		if err == nil && foundationjudge.IsValidJudgeLanguage(languageInt) {
			language = foundationjudge.JudgeLanguage(languageInt)
		}
	}

	countStatics, err := foundationservice.GetJudgeService().GetContestCountStatics(
		ctx,
		contestId,
		language,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	languages, err := foundationservice.GetJudgeService().GetContestLanguageStatics(ctx, contestId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	statistics, err := contestService.GetContestStatistics(ctx, contestId, language)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	resp := struct {
		HasAuth    bool                                       `json:"has_auth"`
		Count      []*foundationview.JudgeJobCountStatics     `json:"count"`
		Language   map[foundationjudge.JudgeLanguage]int      `json:"language"`
		Statistics []*foundationview.ContestProblemStatistics `json:"statistics"`
	}{
		HasAuth:    true,
		Count:      countStatics,
		Language:   languages,
		Statistics: statistics,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, resp)
}

func (c *ContestController) GetMemberSelf(ctx *gin.Context) {
	contestId, err := strconv.Atoi(ctx.Query("id"))
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	member, err := foundationservice.GetContestService().GetContestMember(
		ctx,
		contestId,
		userId,
	)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, member)
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
	ok, errorCode := requestData.CheckRequest()
	if !ok {
		metaresponse.NewResponse(ctx, errorCode, nil)
		return
	}
	startTime := requestData.StartTime
	endTime := requestData.EndTime

	nowTime := metatime.GetTimeNow()

	contestService := foundationservice.GetContestService()
	// 控制创建时的标题唯一，一定程度上防止误重复创建
	ok, err = contestService.HasContestTitle(ctx, userId, requestData.Title)
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
	problemIds := metaslice.RemoveDuplicate(requestData.Problems)
	validProblemIds, err := foundationservice.GetProblemService().FilterValidProblemIds(ctx, problemIds)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	validProblemIdSet := set.FromSlice(validProblemIds)
	var realProblemIds []int
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
				ViewId(nil).                   // 题目描述Id，默认为nil
				Score(0).                      // 分数默认为0
				Index(uint8(len(problems)+1)). // 索引从1开始
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
		Inserter(userId).
		InsertTime(nowTime).
		Modifier(userId).
		ModifyTime(nowTime).
		Private(private).
		Password(password).
		LockRankDuration(lockRankDuration).
		AlwaysLock(requestData.AlwaysLock).
		SubmitAnytime(requestData.SubmitAnytime).
		Build()

	err = contestService.InsertContest(ctx, contest, problems, nil, nil, memberIds, nil, nil, nil)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	if requestData.Description != nil {
		var description string
		var needUpdateUrls []*foundationr2.R2ImageUrl
		description, needUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
			*requestData.Description,
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
		contest.Description = &description
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
	ok, errorCode := requestData.CheckRequest()
	if !ok {
		metaresponse.NewResponse(ctx, errorCode, nil)
		return
	}
	startTime := requestData.StartTime
	endTime := requestData.EndTime
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
	problemIds := metaslice.RemoveDuplicate(requestData.Problems)
	validProblemIds, err := foundationservice.GetProblemService().FilterValidProblemIds(ctx, problemIds)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	validProblemIdSet := set.FromSlice(validProblemIds)
	var realProblemIds []int
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
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	var description string
	if requestData.Description != nil {
		description = *requestData.Description
	}

	var needUpdateUrls []*foundationr2.R2ImageUrl
	description, needUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		description,
		oldDescription,
		metahttp.UrlJoin("contest", strconv.Itoa(contestId)),
		metahttp.UrlJoin("contest", strconv.Itoa(contestId)),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	if description != "" {
		requestData.Description = &description
	} else {
		requestData.Description = nil
	}

	var problems []*foundationmodel.ContestProblem
	for _, problemId := range realProblemIds {
		problems = append(
			problems, foundationmodel.NewContestProblemBuilder().
				ProblemId(problemId).
				ViewId(nil).                   // 题目描述Id，默认为nil
				Score(0).                      // 分数默认为0
				Index(uint8(len(problems)+1)). // 索引从1开始
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
		Id(contestId).
		Title(requestData.Title).
		Description(requestData.Description).
		Notification(requestData.Notification).
		StartTime(startTime).
		EndTime(endTime).
		Modifier(userId).
		ModifyTime(nowTime).
		Private(private).
		Password(password).
		LockRankDuration(lockRankDuration).
		AlwaysLock(requestData.AlwaysLock).
		SubmitAnytime(requestData.SubmitAnytime).
		Build()

	members := make([]*foundationmodel.ContestMember, 0, len(memberIds))
	for _, uid := range memberIds {
		members = append(members, foundationmodel.NewContestMemberBuilder().
			ContestName("").
			UserId(uid).
			Build())
	}

	err = contestService.UpdateContest(ctx, contest, problems, nil, nil, members, nil, nil, nil)
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
		Description *string    `json:"description"`
		UpdateTime  *time.Time `json:"update_time"`
	}{
		Description: requestData.Description,
		UpdateTime:  &contest.ModifyTime,
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
		metaresponse.NewResponseError(ctx, err)
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

func (c *ContestController) PostMemberEditSelf(ctx *gin.Context) {
	var requestData struct {
		Id          int    `json:"id" binding:"required"`
		ContestName string `json:"contest_name"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, hasAuth, err := foundationservice.GetContestService().CheckViewAuthWithoutStartTime(ctx, requestData.Id)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	err = foundationservice.GetContestService().PostContestMemberName(
		ctx,
		userId,
		requestData.Id,
		requestData.ContestName,
	)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, nil)
}
