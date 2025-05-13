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
	metatime "meta/meta-time"
	"meta/response"
	"strconv"
	"time"
	"web/request"
)

type ProblemController struct {
	metacontroller.Controller
}

func (c *ProblemController) Get(ctx *gin.Context) {
	problemService := foundationservice.GetProblemService()
	id := ctx.Query("id")
	if id == "" {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problem, err := problemService.GetProblem(ctx, id)
	if err != nil {
		response.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problem == nil {
		response.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	var tags []*foundationmodel.ProblemTag
	if problem.Tags != nil {
		tags, err = problemService.GetProblemTagByIds(ctx, problem.Tags)
		if err != nil {
			response.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
	}
	responseData := struct {
		Problem *foundationmodel.Problem      `json:"problem"`
		Tags    []*foundationmodel.ProblemTag `json:"tags"`
	}{
		Problem: problem,
		Tags:    tags,
	}
	response.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetList(ctx *gin.Context) {
	problemService := foundationservice.GetProblemService()
	title := ctx.Query("title")
	tag := ctx.Query("tag")
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
	var list []*foundationmodel.Problem
	var totalCount int
	var problemStatus map[string]foundationmodel.ProblemAttemptStatus
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err == nil {
		list, totalCount, problemStatus, err = problemService.GetProblemListWithUser(ctx, userId, title, tag, page, pageSize)
	} else {
		list, totalCount, err = problemService.GetProblemList(ctx, title, tag, page, pageSize)
	}
	if err != nil {
		response.NewResponseError(ctx, err)
		return
	}
	responseData := struct {
		Time                 time.Time                                       `json:"time"`
		TotalCount           int                                             `json:"total_count"`
		List                 []*foundationmodel.Problem                      `json:"list"`
		ProblemAttemptStatus map[string]foundationmodel.ProblemAttemptStatus `json:"problem_attempt_status,omitempty"`
	}{
		Time:                 metatime.GetTimeNow(),
		TotalCount:           totalCount,
		List:                 list,
		ProblemAttemptStatus: problemStatus,
	}
	response.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetTagList(ctx *gin.Context) {
	problemService := foundationservice.GetProblemService()
	maxCountStr := ctx.DefaultQuery("max_count", "-1")
	maxCount, err := strconv.Atoi(maxCountStr)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	list, totalCount, err := problemService.GetProblemTagList(ctx, maxCount)
	if err != nil {
		response.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Time       time.Time                     `json:"time"`
		TotalCount int                           `json:"total_count"`
		List       []*foundationmodel.ProblemTag `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	response.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) PostEdit(ctx *gin.Context) {
	var requestData request.ProblemEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.Title == "" {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.TimeLimit <= 0 || requestData.MemoryLimit <= 0 {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		response.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, userId, foundationauth.AuthTypeManageProblem)
	if err != nil {
		metapanic.ProcessError(err)
		response.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		response.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	hasProblem, err := foundationservice.GetProblemService().HasProblem(ctx, requestData.Id)
	if err != nil {
		response.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !hasProblem {
		response.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	err = foundationservice.GetProblemService().PostEdit(ctx, userId, &requestData)
	if err != nil {
		response.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	response.NewResponse(ctx, metaerrorcode.Success, nil)
}
