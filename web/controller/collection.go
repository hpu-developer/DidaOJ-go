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
	collection, problems, err := collectionService.GetCollection(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if collection == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	responseData := struct {
		Collection *foundationmodel.Collection          `json:"collection"`
		Problems   []*foundationmodel.CollectionProblem `json:"problems"` // 题目列表
	}{
		Collection: collection,
		Problems:   problems,
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
	_, hasAuth, err := collectionService.CheckUserAuth(ctx, id)
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
		Collection *foundationmodel.Collection `json:"collection"`
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
	var list []*foundationmodel.Collection
	var totalCount int
	list, totalCount, err = collectionService.GetCollectionList(ctx, page, pageSize)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Time       time.Time                     `json:"time"`
		TotalCount int                           `json:"total_count"`
		List       []*foundationmodel.Collection `json:"list"`
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
		StartTime *time.Time                        `json:"start_time"`
		EndTime   *time.Time                        `json:"end_time"` // 结束时间
		Problem   int                               `json:"problem"`  // 题目数量
		Ranks     []*foundationmodel.CollectionRank `json:"ranks"`
	}{
		StartTime: startTime,
		EndTime:   endTime,
		Problem:   problems,
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
	startTime := requestData.StartTime
	endTime := requestData.EndTime

	nowTime := metatime.GetTimeNow()

	private := requestData.Private

	collection := foundationmodel.NewCollectionBuilder().
		Title(requestData.Title).
		Description(requestData.Description).
		StartTime(startTime).
		EndTime(endTime).
		OwnerId(userId).
		Problems(realProblemIds).
		Private(private).
		Members(memberIds).
		CreateTime(nowTime).
		UpdateTime(nowTime).
		Build()

	err = collectionService.InsertCollection(ctx, collection)
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
	_, hasAuth, err := foundationservice.GetCollectionService().CheckUserAuth(ctx, requestData.Id)
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

	collectionService := foundationservice.GetCollectionService()

	nowTime := metatime.GetTimeNow()

	startTime := requestData.StartTime
	endTime := requestData.EndTime

	private := requestData.Private

	collection := foundationmodel.NewCollectionBuilder().
		Title(requestData.Title).
		Description(requestData.Description).
		StartTime(startTime).
		EndTime(endTime).
		OwnerId(userId).
		Problems(realProblemIds).
		Private(private).
		Members(memberIds).
		UpdateTime(nowTime).
		Build()

	err = collectionService.UpdateCollection(ctx, requestData.Id, collection)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, collection.UpdateTime)
}
