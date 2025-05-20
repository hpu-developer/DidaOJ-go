package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	metaerrorcode "meta/error-code"
	metaresponse "meta/meta-response"
	metatime "meta/meta-time"
	"strconv"
	"time"
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
		Time       time.Time                   `json:"time"`
		TotalCount int                         `json:"total_count"`
		List       []*foundationmodel.UserRank `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}
