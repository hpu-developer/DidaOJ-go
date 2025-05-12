package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	"meta/error-code"
	metatime "meta/meta-time"
	"meta/response"
	"strconv"
	"time"
	"web/request"
)

type ContestController struct {
	metacontroller.Controller
}

func (c *ContestController) Get(ctx *gin.Context) {
	contestService := foundationservice.GetContestService()
	id := ctx.Query("id")
	if id == "" {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	contest, err := contestService.GetContest(ctx, id)
	if err != nil {
		response.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if contest == nil {
		response.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	response.NewResponse(ctx, metaerrorcode.Success, contest)
}

func (c *ContestController) GetList(ctx *gin.Context) {
	contestService := foundationservice.GetContestService()
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if pageSize != 50 && pageSize != 100 {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	var list []*foundationmodel.Contest
	var totalCount int
	list, totalCount, err = contestService.GetContestList(ctx, page, pageSize)
	if err != nil {
		response.NewResponse(ctx, metaerrorcode.CommonError, nil)
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
	response.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ContestController) PostCreate(ctx *gin.Context) {
	var requestData request.ContestCreate
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if requestData.Title == "" {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	startTime, err := metatime.GetTimeByDateString(requestData.OpenTime[0])
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	endTime, err := metatime.GetTimeByDateString(requestData.OpenTime[1])
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	contestService := foundationservice.GetContestService()

	contest := foundationmodel.NewContestBuilder().
		Title(requestData.Title).
		Description(requestData.Description).
		StartTime(*startTime).
		EndTime(*endTime).
		OwnerId(userId).
		CreateTime(metatime.GetTimeNow()).
		Build()

	err = contestService.InsertContest(ctx, contest)
	if err != nil {
		response.NewResponseError(ctx, err)
		return
	}

	response.NewResponse(ctx, metaerrorcode.Success, contest)
}
