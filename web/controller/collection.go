package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationenum "foundation/foundation-enum"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	foundationview "foundation/foundation-view"
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	"meta/error-code"
	metapanic "meta/meta-panic"
	"meta/meta-response"
	metaslice "meta/meta-slice"
	metatime "meta/meta-time"
	"meta/set"
	"strconv"
	"time"
	weberrorcode "web/error-code"
	"web/request"
)

type CollectionController struct {
	metacontroller.Controller
}

func (c *CollectionController) Get(ctx *gin.Context) {
	collectionService := foundationservice.GetCollectionService()
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
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	collection, joined, problems, tags, attemptStatus, err := collectionService.GetCollection(ctx, id, userId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if collection == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	responseData := struct {
		Collection    *foundationview.CollectionDetail            `json:"collection"`
		Joined        bool                                        `json:"joined"`                   // 是否已加入
		Problems      []*foundationview.ProblemViewList           `json:"problems"`                 // 题目列表
		Tags          []*foundationmodel.Tag                      `json:"tags,omitempty"`           // 题目标签
		AttemptStatus map[int]foundationenum.ProblemAttemptStatus `json:"attempt_status,omitempty"` // 尝试状态，如果已加入则返回
	}{
		Collection:    collection,
		Joined:        joined,
		Problems:      problems,
		Tags:          tags,
		AttemptStatus: attemptStatus,
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *CollectionController) GetEdit(ctx *gin.Context) {
	collectionService := foundationservice.GetCollectionService()
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
	_, hasAuth, err := collectionService.CheckEditAuth(ctx, id)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	collection, err := collectionService.GetCollectionEdit(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if collection == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	responseData := struct {
		Collection *foundationview.CollectionDetail `json:"collection"`
	}{
		Collection: collection,
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *CollectionController) GetList(ctx *gin.Context) {
	collectionService := foundationservice.GetCollectionService()
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
	var list []*foundationview.CollectionList
	var totalCount int
	list, totalCount, err = collectionService.GetCollectionList(ctx, page, pageSize)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Time       time.Time                        `json:"time"`
		TotalCount int                              `json:"total_count"`
		List       []*foundationview.CollectionList `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *CollectionController) GetRank(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	collectionId, err := strconv.Atoi(id)
	if err != nil || collectionId <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	startTime, endTime, problems, ranks, err := foundationservice.GetCollectionService().GetCollectionRanks(
		ctx,
		collectionId,
	)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	responseData := struct {
		StartTime *time.Time                       `json:"start_time"`
		EndTime   *time.Time                       `json:"end_time"` // 结束时间
		Problems  int                              `json:"problem"`  // 题目数量
		Ranks     []*foundationview.CollectionRank `json:"ranks"`
	}{
		StartTime: startTime,
		EndTime:   endTime,
		Problems:  problems,
		Ranks:     ranks,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *CollectionController) PostCreate(ctx *gin.Context) {
	var requestData request.CollectionEdit
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
	collectionService := foundationservice.GetCollectionService()
	// 控制创建时的标题唯一，一定程度上防止误重复创建
	ok, err := collectionService.HasCollectionTitle(ctx, userId, requestData.Title)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if ok {
		metaresponse.NewResponse(ctx, weberrorcode.CollectionTitleDuplicate, nil)
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
	startTime := requestData.StartTime
	endTime := requestData.EndTime

	nowTime := metatime.GetTimeNow()

	private := requestData.Private

	collection := foundationmodel.NewCollectionBuilder().
		Title(requestData.Title).
		Description(requestData.Description).
		StartTime(startTime).
		EndTime(endTime).
		Inserter(userId).
		Private(private).
		InsertTime(nowTime).
		ModifyTime(nowTime).
		Build()

	err = collectionService.InsertCollection(ctx, collection, realProblemIds, memberIds)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, collection.Id)
}

func (c *CollectionController) PostEdit(ctx *gin.Context) {
	var requestData request.CollectionEdit
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
	collectionService := foundationservice.GetCollectionService()
	// 控制创建时的标题唯一，一定程度上防止误重复创建
	ok, err := collectionService.HasCollectionTitle(ctx, userId, requestData.Title)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if ok {
		metaresponse.NewResponse(ctx, weberrorcode.CollectionTitleDuplicate, nil)
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
	startTime := requestData.StartTime
	endTime := requestData.EndTime

	nowTime := metatime.GetTimeNow()

	private := requestData.Private

	collection := foundationmodel.NewCollectionBuilder().
		Id(requestData.Id).
		Title(requestData.Title).
		Description(requestData.Description).
		StartTime(startTime).
		EndTime(endTime).
		Inserter(userId).
		Private(private).
		InsertTime(nowTime).
		ModifyTime(nowTime).
		Build()

	err = collectionService.UpdateCollection(ctx, collection, realProblemIds, memberIds)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, collection.ModifyTime)
}

func (c *CollectionController) PostJoin(ctx *gin.Context) {
	var requestData struct {
		Id int `json:"id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	collectionId := requestData.Id
	collectionService := foundationservice.GetCollectionService()
	userId, hasAuth, err := collectionService.CheckJoinAuth(ctx, collectionId)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	err = collectionService.PostJoin(ctx, collectionId, userId)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, nil)
}

func (c *CollectionController) PostQuit(ctx *gin.Context) {
	var requestData struct {
		Id int `json:"id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NeedLogin, nil)
		return
	}
	if userId <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.NeedLogin, nil)
		return
	}
	collectionId := requestData.Id
	collectionService := foundationservice.GetCollectionService()
	err = collectionService.PostQuit(ctx, collectionId, userId)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, nil)
}
