package controller

import (
	"errors"
	"fmt"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationmodel "foundation/foundation-model"
	foundationoj "foundation/foundation-oj"
	foundationr2 "foundation/foundation-r2"
	foundationservice "foundation/foundation-service"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	cfr2 "meta/cf-r2"
	metacontroller "meta/controller"
	"meta/error-code"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	"meta/meta-response"
	metatime "meta/meta-time"
	metazip "meta/meta-zip"
	"meta/set"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"web/config"
	weberrorcode "web/error-code"
	"web/request"
	"web/service"
)

type ProblemJudgeData struct {
	Key          string     `json:"key"`
	Size         *int64     `json:"size"`
	LastModified *time.Time `json:"last_modified"`
}

type ProblemController struct {
	metacontroller.Controller
}

func (c *ProblemController) Get(ctx *gin.Context) {
	problemService := foundationservice.GetProblemService()
	id := ctx.Query("id")
	isContest := false
	if id == "" {
		contestIdStr := ctx.Query("contest_id")
		if contestIdStr == "" {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		contestId, err := strconv.Atoi(contestIdStr)
		if err != nil || contestId <= 0 {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		problemIndexStr := ctx.Query("problem_index")
		if problemIndexStr == "" {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		problemIndex, err := strconv.Atoi(problemIndexStr)
		if err != nil || problemIndex <= 0 {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		idPtr, err := problemService.GetProblemIdByContest(ctx, contestId, problemIndex)
		if err != nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		if idPtr == nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
			return
		}
		id = *idPtr
		isContest = true
	}

	userId, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)

	problem, err := problemService.GetProblemView(ctx, id, userId, ok)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problem == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	var tags []*foundationmodel.ProblemTag
	if isContest {
		// 比赛时隐藏一些信息
		problem.Id = ""
		problem.Source = ""
		problem.Accept = 0
		problem.Attempt = 0
		problem.OriginOj = nil
		problem.OriginId = nil
		problem.OriginUrl = nil
	} else {
		if problem.Tags != nil {
			tags, err = problemService.GetProblemTagByIds(ctx, problem.Tags)
			if err != nil {
				metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
				return
			}
		}
	}
	responseData := struct {
		Problem *foundationmodel.Problem      `json:"problem"`
		Tags    []*foundationmodel.ProblemTag `json:"tags,omitempty"`
	}{
		Problem: problem,
		Tags:    tags,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetList(ctx *gin.Context) {
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
	problemService := foundationservice.GetProblemService()
	oj := ctx.Query("oj")
	if oj != "" {
		oj = foundationoj.GetOriginOjKey(oj)
		if oj == "" {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
	}
	title := ctx.Query("title")
	tag := ctx.Query("tag")
	var list []*foundationmodel.Problem
	var totalCount int
	var problemStatus map[string]foundationmodel.ProblemAttemptStatus
	userId, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if userId > 0 {
		private := ctx.Query("private") != "0"
		list, totalCount, problemStatus, err = problemService.GetProblemListWithUser(
			ctx,
			userId,
			ok,
			oj,
			title,
			tag,
			private,
			page,
			pageSize,
		)
	} else {
		list, totalCount, err = problemService.GetProblemList(ctx, oj, title, tag, page, pageSize)
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

func (c *ProblemController) GetAttemptStatus(ctx *gin.Context) {
	idsStr := ctx.Query("ids")
	if idsStr == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	ids := strings.Split(idsStr, ",")
	validProblemIds, err := foundationservice.GetProblemService().FilterValidProblemIds(ctx, ids)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	if len(validProblemIds) == 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	problemStatus, err := foundationservice.GetJudgeService().GetProblemAttemptStatus(
		ctx,
		validProblemIds,
		userId,
		-1,
		nil,
		nil,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, problemStatus)
}

func (c *ProblemController) GetRecommend(ctx *gin.Context) {
	userId, hasAuth, err := foundationservice.GetUserService().CheckUserAuth(
		ctx,
		foundationauth.AuthTypeManageProblem,
	)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if userId <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	problemService := foundationservice.GetProblemService()
	problemId := ctx.Query("problem_id")
	list, err := problemService.GetProblemRecommend(ctx, userId, hasAuth, problemId)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	var tags []*foundationmodel.ProblemTag
	if len(list) > 0 {
		tagIdSet := set.New[int]()
		for _, problem := range list {
			if problem.Tags != nil {
				for _, tagId := range problem.Tags {
					tagIdSet.Add(tagId)
				}
			}
		}
		var tagIds []int
		tagIdSet.Foreach(
			func(tagId *int) bool {
				tagIds = append(tagIds, *tagId)
				return true
			},
		)
		tags, err = foundationservice.GetProblemService().GetProblemTagByIds(ctx, tagIds)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
	}
	responseData := struct {
		List []*foundationmodel.Problem    `json:"list"`
		Tags []*foundationmodel.ProblemTag `json:"tags,omitempty"`
	}{
		List: list,
		Tags: tags,
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
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	_, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
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
	problem, err := problemService.GetProblemViewJudgeData(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problem == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	r2Client := cfr2.GetSubsystem().GetClient("judge-data")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	problemId := problem.Id
	prefixKey := filepath.ToSlash(problemId + "/")
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("didaoj-judge"),
		Prefix: aws.String(prefixKey),
	}

	var judges []*ProblemJudgeData

	err = r2Client.ListObjectsV2PagesWithContext(
		ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				judgeData := &ProblemJudgeData{
					Key:          strings.TrimPrefix(*obj.Key, prefixKey),
					Size:         obj.Size,
					LastModified: obj.LastModified,
				}
				judges = append(judges, judgeData)
			}
			return true
		},
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	responseData := struct {
		Problem *foundationmodel.Problem `json:"problem"`
		Judges  []*ProblemJudgeData      `json:"judges"`
	}{
		Problem: problem,
		Judges:  judges,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}
func (c *ProblemController) GetJudgeDataDownload(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	key := ctx.Query("key")
	if key == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 鉴权
	_, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	// 获取题目信息
	problemService := foundationservice.GetProblemService()
	problem, err := problemService.GetProblemViewJudgeData(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problem == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	// 获取 R2 客户端
	r2Client := cfr2.GetSubsystem().GetClient("judge-data")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	// 生成预签名链接
	objectKey := filepath.ToSlash(path.Join(id, key))
	req, _ := r2Client.GetObjectRequest(
		&s3.GetObjectInput{
			Bucket: aws.String("didaoj-judge"),
			Key:    aws.String(objectKey),
		},
	)
	expire := 10 * time.Minute
	urlStr, err := req.Presign(expire)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, urlStr)
}

func (c *ProblemController) GetImageToken(ctx *gin.Context) {
	id := ctx.Query("id")
	_, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
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
		"uploading/problem",
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

func (c *ProblemController) PostParse(ctx *gin.Context) {
	var requestData struct {
		Problems []string `json:"problems" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemList := requestData.Problems
	userId, hasAuth, err := foundationservice.GetUserService().CheckUserAuth(
		ctx,
		foundationauth.AuthTypeManageProblem,
	)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	problemTitles, err := foundationservice.GetProblemService().GetProblemTitles(ctx, userId, hasAuth, problemList)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Problems []*foundationmodel.ProblemViewTitle `json:"problems"`
	}{
		Problems: problemTitles,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) PostCrawl(ctx *gin.Context) {
	var requestData struct {
		OJ string `json:"oj",binding:"required"`
		Id string `json:"id",binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.OJ == "" || requestData.Id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	oj := strings.ToLower(requestData.OJ)
	id := strings.TrimSpace(requestData.Id)
	if oj == "didaoj" {
		ok, err := foundationservice.GetProblemService().HasProblem(ctx, id)
		if err != nil {
			return
		}
		if !ok {
			metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
			return
		}
		metaresponse.NewResponse(ctx, metaerrorcode.Success, id)
		return
	}
	newId, err := service.GetProblemCrawlService().PostCrawlProblem(ctx, oj, id)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	if newId == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, newId)
}

func (c *ProblemController) PostJudgeData(ctx *gin.Context) {
	problemId := ctx.PostForm("id")
	if problemId == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	_, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
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
	problem, err := problemService.GetProblemViewJudgeData(ctx, problemId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problem == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	file, err := ctx.FormFile("zip")
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError)
		return
	}
	tempDir, err := os.MkdirTemp("", "didaoj-judge-data-*")
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError)
		return
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "<UNK>: "+path))
		}
	}(tempDir)
	uploadedPath := filepath.Join(tempDir, file.Filename)
	if err := ctx.SaveUploadedFile(file, uploadedPath); err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError)
		return
	}
	unzipDir := filepath.Join(tempDir, "unzipped")
	if err := metazip.UzipFile(uploadedPath, unzipDir); err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.ProblemJudgeDataMustZip, nil)
		return
	}

	err = problemService.PostJudgeData(
		ctx,
		problemId,
		unzipDir,
		problem.JudgeMd5,
		config.GetConfig().GoJudge.Url,
		false,
	)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, nil)
}

func (c *ProblemController) PostCreate(ctx *gin.Context) {
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
	ok, err = foundationservice.GetProblemService().HasProblemTitle(ctx, requestData.Title)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if ok {
		metaresponse.NewResponse(ctx, weberrorcode.ProblemTitleDuplicate, nil)
		return
	}

	nowTime := metatime.GetTimeNow()

	problem := foundationmodel.NewProblemBuilder().
		Title(requestData.Title).
		Description(requestData.Description).
		Source(requestData.Source).
		TimeLimit(requestData.TimeLimit).
		MemoryLimit(requestData.MemoryLimit).
		InsertTime(nowTime).
		UpdateTime(nowTime).
		CreatorId(userId).
		Private(requestData.Private).
		Build()
	problemId, err := foundationservice.GetProblemService().InsertProblem(ctx, problem, requestData.Tags)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	var description string
	var needUpdateUrls []*foundationr2.R2ImageUrl
	description, needUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		requestData.Description,
		nil,
		metahttp.UrlJoin("problem"),
		metahttp.UrlJoin("problem", *problemId),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if len(needUpdateUrls) > 0 {
		err := foundationservice.GetProblemService().UpdateProblemDescription(ctx, *problemId, description)
		if err != nil {
			return
		}
		err = service.GetR2ImageService().MoveImageAfterSave(needUpdateUrls)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, problemId)
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
	problemId := requestData.Id

	_, ok, err := foundationservice.GetProblemService().CheckEditAuth(ctx, problemId)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	description := requestData.Description

	oldDescription, err := foundationservice.GetProblemService().GetProblemDescription(ctx, problemId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
			return
		}
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	var needUpdateUrls []*foundationr2.R2ImageUrl
	description, needUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		description,
		oldDescription,
		metahttp.UrlJoin("problem", problemId),
		metahttp.UrlJoin("problem", problemId),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	requestData.Description = description

	nowTime := metatime.GetTimeNow()

	problem := foundationmodel.NewProblemBuilder().
		Title(requestData.Title).
		Description(requestData.Description).
		Source(requestData.Source).
		TimeLimit(requestData.TimeLimit).
		MemoryLimit(requestData.MemoryLimit).
		Private(requestData.Private).
		UpdateTime(nowTime).
		Build()

	err = foundationservice.GetProblemService().UpdateProblem(ctx, problemId, problem, requestData.Tags)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	// 因为数据库已经保存了，因此即使图片失败这里也返回成功
	err = service.GetR2ImageService().MoveImageAfterSave(needUpdateUrls)
	if err != nil {
		metapanic.ProcessError(err)
	}

	responseData := struct {
		Description string    `json:"description"`
		UpdateTime  time.Time `json:"update_time"`
	}{
		Description: description,
		UpdateTime:  nowTime,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}
