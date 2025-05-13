package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	cfr2 "meta/cf-r2"
	metacontroller "meta/controller"
	"meta/error-code"
	metapanic "meta/meta-panic"
	"meta/meta-response"
	metatime "meta/meta-time"
	"strconv"
	"strings"
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
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problem, err := problemService.GetProblem(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problem == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	var tags []*foundationmodel.ProblemTag
	if problem.Tags != nil {
		tags, err = problemService.GetProblemTagByIds(ctx, problem.Tags)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
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
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetList(ctx *gin.Context) {
	problemService := foundationservice.GetProblemService()
	title := ctx.Query("title")
	tag := ctx.Query("tag")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
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
		metaresponse.NewResponseError(ctx, err)
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
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetTagList(ctx *gin.Context) {
	problemService := foundationservice.GetProblemService()
	maxCountStr := ctx.DefaultQuery("max_count", "-1")
	maxCount, err := strconv.Atoi(maxCountStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	list, totalCount, err := problemService.GetProblemTagList(ctx, maxCount)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
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
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetJudge(ctx *gin.Context) {
	problemService := foundationservice.GetProblemService()
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problem, err := problemService.GetProblemJudge(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problem == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	var judges []string

	r2Client := cfr2.GetSubsystem().GetClient("judge-data")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	problemId := problem.Id

	prefixKey := problemId + "/"

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("didaoj-judge"),
		Prefix: aws.String(prefixKey),
	}
	err = r2Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			judges = append(judges, strings.TrimPrefix(*obj.Key, prefixKey))
		}
		return true
	})
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	responseData := struct {
		Problem *foundationmodel.Problem `json:"problem"`
		Judges  []string                 `json:"judges"`
	}{
		Problem: problem,
		Judges:  judges,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *ProblemController) PostEdit(ctx *gin.Context) {
	var requestData request.ProblemEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.Title == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.TimeLimit <= 0 || requestData.MemoryLimit <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	userId, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	hasProblem, err := foundationservice.GetProblemService().HasProblem(ctx, requestData.Id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !hasProblem {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	err = foundationservice.GetProblemService().PostEdit(ctx, userId, &requestData)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, nil)
}
