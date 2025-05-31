package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	"meta/error-code"
	"meta/meta-response"
	metatime "meta/meta-time"
	"strconv"
	"time"
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
		Problems  []string                          `json:"problems"` // 题目索引列表
		Ranks     []*foundationmodel.CollectionRank `json:"ranks"`
	}{
		StartTime: startTime,
		EndTime:   endTime,
		Problems:  problems,
		Ranks:     ranks,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *CollectionController) PostCreate(ctx *gin.Context) {
	var requestData request.CollectionCreate
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
	startTime, err := metatime.GetTimeByDateString(requestData.StartTime)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	endTime, err := metatime.GetTimeByDateString(requestData.EndTime)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	collectionService := foundationservice.GetCollectionService()

	collection := foundationmodel.NewCollectionBuilder().
		Title(requestData.Title).
		Description(requestData.Description).
		StartTime(startTime).
		EndTime(endTime).
		OwnerId(userId).
		CreateTime(metatime.GetTimeNow()).
		Build()

	err = collectionService.InsertCollection(ctx, collection)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, collection)
}
