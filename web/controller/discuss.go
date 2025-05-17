package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationcontest "foundation/foundation-contest"
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

type DiscussController struct {
	metacontroller.Controller
}

func (c *DiscussController) Get(ctx *gin.Context) {
	discussService := foundationservice.GetDiscussService()
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
	discuss, err := discussService.GetDiscuss(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if discuss == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	var tags []*foundationmodel.DiscussTag
	if discuss.Tags != nil {
		tags, err = discussService.GetDiscussTagByIds(ctx, discuss.Tags)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
	}
	responseData := struct {
		Discuss *foundationmodel.Discuss      `json:"discuss"`
		Tags    []*foundationmodel.DiscussTag `json:"tags,omitempty"`
	}{
		Discuss: discuss,
		Tags:    tags,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *DiscussController) GetList(ctx *gin.Context) {
	discussService := foundationservice.GetDiscussService()
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
	problemId := ctx.Query("problem_id")
	var contestId, constProblemIndex int
	contestIdStr := ctx.Query("contest_id")
	if contestIdStr != "" {
		contestId, err = strconv.Atoi(contestIdStr)
		constProblemIndex = foundationcontest.GetContestProblemIndex(problemId)
	}
	title := ctx.Query("title")
	username := ctx.Query("username")

	var list []*foundationmodel.Discuss
	var totalCount int
	list, totalCount, err = discussService.GetDiscussList(ctx,
		contestId, constProblemIndex, problemId,
		title, username,
		page, pageSize)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Time       time.Time                  `json:"time"`
		TotalCount int                        `json:"total_count"`
		List       []*foundationmodel.Discuss `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *DiscussController) GetCommentList(ctx *gin.Context) {
	discussService := foundationservice.GetDiscussService()
	discussIdStr := ctx.Query("id")
	if discussIdStr == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	discussId, err := strconv.Atoi(discussIdStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
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
	if pageSize != 20 && pageSize != 50 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	var list []*foundationmodel.DiscussComment
	var totalCount int
	list, totalCount, err = discussService.GetDiscussCommentList(ctx, discussId, page, pageSize)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Time       time.Time                         `json:"time"`
		TotalCount int                               `json:"total_count"`
		List       []*foundationmodel.DiscussComment `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *DiscussController) PostCreate(ctx *gin.Context) {
	var requestData request.DiscussCreate
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

	discussService := foundationservice.GetDiscussService()

	discuss := foundationmodel.NewDiscussBuilder().
		Title(requestData.Title).
		Content(requestData.Content).
		AuthorId(userId).
		InsertTime(metatime.GetTimeNow()).
		ModifyTime(metatime.GetTimeNow()).
		UpdateTime(metatime.GetTimeNow()).
		Build()

	err = discussService.InsertDiscuss(ctx, discuss)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, discuss)
}
