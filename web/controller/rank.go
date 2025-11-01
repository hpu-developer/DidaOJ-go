package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationservice "foundation/foundation-service"
	foundationview "foundation/foundation-view"
	metacontroller "meta/controller"
	metaerrorcode "meta/error-code"
	metaresponse "meta/meta-response"
	metatime "meta/meta-time"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type RankController struct {
	metacontroller.Controller
}

func (c *RankController) GetAcAll(ctx *gin.Context) {
	userService := foundationservice.GetUserService()
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
	if pageSize != 50 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	list, totalCount, err := userService.GetRankAcAll(ctx, page, pageSize)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	responseData := struct {
		Time       time.Time                  `json:"time"`
		TotalCount int                        `json:"total_count"`
		List       []*foundationview.UserRank `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *RankController) GetAcProblem(ctx *gin.Context) {
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
	if pageSize != 50 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	list, totalCount, err := foundationservice.GetJudgeService().GetRankAcProblem(ctx, nil, nil, page, pageSize)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	responseData := struct {
		Time       time.Time                  `json:"time"`
		TotalCount int                        `json:"total_count"`
		List       []*foundationview.UserRank `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *RankController) GetAcProblemToday(ctx *gin.Context) {
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
	if pageSize != 50 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 获取进入凌晨的时间
	startTime := metatime.GetTimeTodayStart()
	endTime := metatime.GetTimeTodayEnd()
	list, totalCount, err := foundationservice.GetJudgeService().GetRankAcProblem(
		ctx,
		&startTime,
		&endTime,
		page,
		pageSize,
	)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	responseData := struct {
		Time       time.Time                  `json:"time"`
		TotalCount int                        `json:"total_count"`
		List       []*foundationview.UserRank `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *RankController) GetAcProblemDay7(ctx *gin.Context) {
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
	if pageSize != 50 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 获取7日之前的凌晨0点
	startTime := metatime.GetTimeDayStart(-7 * 24 * time.Hour)
	endTime := metatime.GetTimeNow()
	list, totalCount, err := foundationservice.GetJudgeService().GetRankAcProblem(
		ctx,
		&startTime,
		&endTime,
		page,
		pageSize,
	)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	responseData := struct {
		Time       time.Time                  `json:"time"`
		TotalCount int                        `json:"total_count"`
		List       []*foundationview.UserRank `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *RankController) GetAcProblemDay30(ctx *gin.Context) {
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
	if pageSize != 50 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 获取30日之前的凌晨0点
	startTime := metatime.GetTimeDayStart(-30 * 24 * time.Hour)
	endTime := metatime.GetTimeNow()
	list, totalCount, err := foundationservice.GetJudgeService().GetRankAcProblem(
		ctx,
		&startTime,
		&endTime,
		page,
		pageSize,
	)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	responseData := struct {
		Time       time.Time                  `json:"time"`
		TotalCount int                        `json:"total_count"`
		List       []*foundationview.UserRank `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *RankController) GetAcProblemYear(ctx *gin.Context) {
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
	if pageSize != 50 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 获取365日之前的凌晨0点
	startTime := metatime.GetTimeDayStart(-365 * 24 * time.Hour)
	endTime := metatime.GetTimeNow()
	list, totalCount, err := foundationservice.GetJudgeService().GetRankAcProblem(
		ctx,
		&startTime,
		&endTime,
		page,
		pageSize,
	)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	responseData := struct {
		Time       time.Time                  `json:"time"`
		TotalCount int                        `json:"total_count"`
		List       []*foundationview.UserRank `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}
