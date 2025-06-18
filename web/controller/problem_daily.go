package controller

import (
	"fmt"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	foundationr2 "foundation/foundation-r2"
	foundationservice "foundation/foundation-service"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	cfr2 "meta/cf-r2"
	"meta/error-code"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	"meta/meta-response"
	metatime "meta/meta-time"
	"strconv"
	"time"
	weberrorcode "web/error-code"
	"web/request"
	"web/service"
)

func (c *ProblemController) GetDailyImageToken(ctx *gin.Context) {
	id := ctx.Query("id")
	_, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblemDaily)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	// 获取 R2 客户端
	r2Client := cfr2.GetSubsystem().GetClient("didapipa-oj")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	// 配置参数
	bucketName := "didapipa-oj"
	objectKey := metahttp.UrlJoin(
		"uploading/problem-daily",
		id,
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

func (c *ProblemController) GetDaily(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemService := foundationservice.GetProblemService()
	problemDaily, err := problemService.GetProblemDaily(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problemDaily == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	responseData := struct {
		Time         time.Time                     `json:"time"`
		ProblemDaily *foundationmodel.ProblemDaily `json:"problem_daily"`
	}{
		Time:         metatime.GetTimeNow(),
		ProblemDaily: problemDaily,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetDailyEdit(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	_, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblemDaily)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	problemService := foundationservice.GetProblemService()
	problemDaily, err := problemService.GetProblemDailyEdit(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problemDaily == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, problemDaily)
}

func (c *ProblemController) GetDailyId(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemService := foundationservice.GetProblemService()
	problemId, err := problemService.GetProblemIdByDaily(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, problemId)
}

func (c *ProblemController) GetDailyList(ctx *gin.Context) {
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
	var startDate *string
	startDateStr := ctx.Query("start_date")
	if startDateStr != "" {
		if _, err := time.Parse("2006-01-02", startDateStr); err != nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		startDate = &startDateStr
	}
	var endDate *string
	endDateStr := ctx.Query("end_date")
	if endDateStr != "" {
		if _, err := time.Parse("2006-01-02", endDateStr); err != nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		endDate = &endDateStr
	}

	problemService := foundationservice.GetProblemService()
	problemId := ctx.Query("problem_id")

	userId, err := foundationauth.GetUserIdFromContext(ctx)

	list, totalCount, tags, problemStatus, err := problemService.GetDailyList(
		ctx,
		userId,
		startDate,
		endDate,
		problemId,
		page,
		pageSize,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Time          time.Time                                       `json:"time"`
		TotalCount    int                                             `json:"total_count"`
		List          []*foundationmodel.ProblemDaily                 `json:"list"`
		Tags          []*foundationmodel.ProblemTag                   `json:"tags,omitempty"`
		AttemptStatus map[string]foundationmodel.ProblemAttemptStatus `json:"attempt_status,omitempty"`
	}{
		Time:          metatime.GetTimeNow(),
		TotalCount:    totalCount,
		List:          list,
		Tags:          tags,
		AttemptStatus: problemStatus,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetDailyRecently(ctx *gin.Context) {

	userId, err := foundationauth.GetUserIdFromContext(ctx)

	list, problemStatus, err := foundationservice.GetProblemService().GetDailyRecently(ctx, userId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	responseData := struct {
		Time          time.Time                                       `json:"time"`
		List          []*foundationmodel.ProblemDaily                 `json:"list"`
		AttemptStatus map[string]foundationmodel.ProblemAttemptStatus `json:"attempt_status,omitempty"`
	}{
		Time:          metatime.GetTimeNow(),
		List:          list,
		AttemptStatus: problemStatus,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) PostDailyCreate(ctx *gin.Context) {
	var requestData request.ProblemDailyEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	id := requestData.Id
	const layout = "2006-01-02"
	t, err := time.Parse(layout, id)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	year := t.Year()
	if year < 2010 || year > 2100 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblemDaily)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	ok, err = foundationservice.GetProblemService().HasProblemDaily(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if ok {
		metaresponse.NewResponse(ctx, weberrorcode.ProblemDailyAlreadyExists, nil)
		return
	}
	ok, err = foundationservice.GetProblemService().HasProblemDailyProblem(ctx, requestData.ProblemId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if ok {
		metaresponse.NewResponse(ctx, weberrorcode.ProblemDailyProblemAlreadyExists, nil)
		return
	}
	ok, err = foundationservice.GetProblemService().HasProblem(ctx, requestData.ProblemId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, weberrorcode.ProblemNotFound, nil)
		return
	}

	var finalNeedUpdateUrls []*foundationr2.R2ImageUrl
	requestData.Solution, finalNeedUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		requestData.Solution,
		nil,
		metahttp.UrlJoin("problem-daily"),
		metahttp.UrlJoin("problem-daily", id),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	var codeNeedUpdateUrls []*foundationr2.R2ImageUrl
	requestData.Code, codeNeedUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		requestData.Code,
		nil,
		metahttp.UrlJoin("problem-daily"),
		metahttp.UrlJoin("problem-daily", id),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	problemDaily := foundationmodel.NewProblemDailyBuilder().
		Id(id).
		ProblemId(requestData.ProblemId).
		Solution(requestData.Solution).
		Code(requestData.Code).
		CreateTime(metatime.GetTimeNow()).
		UpdateTime(metatime.GetTimeNow()).
		CreatorId(userId).
		UpdaterId(userId).
		Build()
	err = foundationservice.GetProblemService().PostDailyCreate(ctx, problemDaily)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			metaresponse.NewResponse(ctx, weberrorcode.ProblemDailyAlreadyExists, nil)
			return
		}
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	finalNeedUpdateUrls = append(finalNeedUpdateUrls, codeNeedUpdateUrls...)
	// 因为数据库已经保存了，因此即使图片失败这里也返回成功
	err = service.GetR2ImageService().MoveImageAfterSave(finalNeedUpdateUrls)
	if err != nil {
		metapanic.ProcessError(err)
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *ProblemController) PostDailyEdit(ctx *gin.Context) {
	var requestData request.ProblemDailyEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	id := requestData.Id
	const layout = "2006-01-02"
	t, err := time.Parse(layout, id)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	year := t.Year()
	if year < 2010 || year > 2100 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblemDaily)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	ok, err = foundationservice.GetProblemService().HasProblem(ctx, requestData.ProblemId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, weberrorcode.ProblemNotFound, nil)
		return
	}
	oldDaily, err := foundationservice.GetProblemService().GetProblemDailyEdit(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	var finalNeedUpdateUrls []*foundationr2.R2ImageUrl
	requestData.Solution, finalNeedUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		requestData.Solution,
		&oldDaily.Solution,
		metahttp.UrlJoin("problem-daily", id),
		metahttp.UrlJoin("problem-daily", id),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	var codeNeedUpdateUrls []*foundationr2.R2ImageUrl
	requestData.Code, codeNeedUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		requestData.Code,
		&oldDaily.Code,
		metahttp.UrlJoin("problem-daily", id),
		metahttp.UrlJoin("problem-daily", id),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	problemDaily := foundationmodel.NewProblemDailyBuilder().
		ProblemId(requestData.ProblemId).
		Solution(requestData.Solution).
		Code(requestData.Code).
		UpdateTime(metatime.GetTimeNow()).
		UpdaterId(userId).
		Build()
	err = foundationservice.GetProblemService().PostDailyEdit(ctx, id, problemDaily)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	finalNeedUpdateUrls = append(finalNeedUpdateUrls, codeNeedUpdateUrls...)
	// 因为数据库已经保存了，因此即使图片失败这里也返回成功
	err = service.GetR2ImageService().MoveImageAfterSave(finalNeedUpdateUrls)
	if err != nil {
		metapanic.ProcessError(err)
	}

	if problemDaily.CreatorId > 0 {
		user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, problemDaily.CreatorId)
		if err == nil && user != nil {
			problemDaily.UpdaterUsername = &user.Username
			problemDaily.UpdaterNickname = &user.Nickname
		}
	}
	if problemDaily.UpdaterId > 0 {
		if problemDaily.UpdaterId == problemDaily.CreatorId {
			problemDaily.UpdaterUsername = problemDaily.CreatorUsername
			problemDaily.UpdaterNickname = problemDaily.CreatorNickname
		} else {
			user, err := foundationservice.GetUserService().GetUserAccountInfo(ctx, problemDaily.UpdaterId)
			if err == nil && user != nil {
				problemDaily.UpdaterUsername = &user.Username
				problemDaily.UpdaterNickname = &user.Nickname
			}
		}
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, problemDaily)
}
