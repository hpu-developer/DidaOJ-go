package controller

import (
	"errors"
	"fmt"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationcontest "foundation/foundation-contest"
	foundationmodel "foundation/foundation-model-mongo"
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

func (c *DiscussController) GetEdit(ctx *gin.Context) {
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
	_, hasAuth, err := discussService.CheckEditAuth(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
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

func (c *DiscussController) GetImageToken(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	id := ctx.Query("id")
	// 获取 R2 客户端
	r2Client := cfr2.GetSubsystem().GetClient("didapipa-oj")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	// 配置参数
	bucketName := "didapipa-oj"
	objectKey := metahttp.UrlJoin(
		"uploading/discuss",
		id,
		strconv.Itoa(userId),
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

func (c *DiscussController) GetCommentImageToken(ctx *gin.Context) {
	idStr := ctx.Query("id")
	commentIdStr := ctx.Query("comment_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, hasAuth, err := foundationservice.GetDiscussService().CheckViewAuth(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	commentId, err := strconv.Atoi(commentIdStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if commentId > 0 {
		_, hasAuth, discussComment, err := foundationservice.GetDiscussService().CheckEditCommentAuth(ctx, commentId)
		if err != nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
			return
		}
		if !hasAuth {
			metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
			return
		}
		idStr = strconv.Itoa(discussComment.DiscussId)
		commentIdStr = strconv.Itoa(commentId)
	} else {
		idStr = strconv.Itoa(id)
		commentIdStr = ""
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
		"uploading/discuss",
		idStr,
		commentIdStr,
		strconv.Itoa(userId),
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
	onlyProblemStr := ctx.Query("only_problem")
	onlyProblem := false
	if onlyProblemStr == "1" {
		onlyProblem = true
	}
	problemId := ctx.Query("problem_id")
	var contestId, constProblemIndex int
	contestIdStr := ctx.Query("contest_id")
	if contestIdStr != "" {
		contestId, err = strconv.Atoi(contestIdStr)
		_, hasAuth, err := foundationservice.GetContestService().CheckViewAuth(ctx, contestId)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
		if !hasAuth {
			responseData := struct {
				HasAuth bool `json:"has_auth"`
			}{
				HasAuth: false,
			}
			metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
			return
		}
		constProblemIndex = foundationcontest.GetContestProblemIndex(problemId)
	}
	title := ctx.Query("title")
	username := ctx.Query("username")

	var list []*foundationmodel.Discuss
	var totalCount int
	list, totalCount, err = discussService.GetDiscussList(
		ctx,
		onlyProblem,
		contestId, constProblemIndex, problemId,
		title, username,
		page, pageSize,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		HasAuth    bool                       `json:"has_auth"`
		Time       time.Time                  `json:"time"`
		TotalCount int                        `json:"total_count"`
		List       []*foundationmodel.Discuss `json:"list"`
	}{
		HasAuth:    true,
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
	var requestData request.DiscussEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.Title == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	if requestData.ProblemId > 0 {
		ok, err := foundationservice.GetProblemService().HasProblem(ctx, requestData.ProblemId)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
		if !ok {
			metaresponse.NewResponse(ctx, weberrorcode.ProblemNotFound, nil)
			return
		}
	}

	discussService := foundationservice.GetDiscussService()

	timeNow := metatime.GetTimeNow()

	discuss := foundationmodel.NewDiscussBuilder().
		Title(requestData.Title).
		Content(requestData.Content).
		AuthorId(userId).
		InsertTime(timeNow).
		ModifyTime(timeNow).
		UpdateTime(timeNow).
		Build()

	err = discussService.InsertDiscuss(ctx, discuss)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	var description string
	var needUpdateUrls []*foundationr2.R2ImageUrl
	description, needUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		requestData.Content,
		nil,
		metahttp.UrlJoin("discuss"),
		metahttp.UrlJoin("discuss", strconv.Itoa(discuss.Id)),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if len(needUpdateUrls) > 0 {
		err := discussService.UpdateContent(ctx, discuss.Id, description)
		if err != nil {
			return
		}
		err = service.GetR2ImageService().MoveImageAfterSave(needUpdateUrls)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, discuss.Id)
}

func (c *DiscussController) PostEdit(ctx *gin.Context) {
	var requestData request.DiscussEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.Title == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	discussService := foundationservice.GetDiscussService()

	_, hasAuth, err := discussService.CheckEditAuth(ctx, requestData.Id)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if requestData.ProblemId > 0 {
		ok, err := foundationservice.GetProblemService().HasProblem(ctx, requestData.ProblemId)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
		if !ok {
			metaresponse.NewResponse(ctx, weberrorcode.ProblemNotFound, nil)
			return
		}
	}

	discussId := requestData.Id
	content := requestData.Content

	oldDescription, err := discussService.GetContent(ctx, discussId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
			return
		}
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	var needUpdateUrls []*foundationr2.R2ImageUrl
	content, needUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		content,
		oldDescription,
		metahttp.UrlJoin("discuss", strconv.Itoa(discussId)),
		metahttp.UrlJoin("discuss", strconv.Itoa(discussId)),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	requestData.Content = content

	nowTime := metatime.GetTimeNow()

	discuss := foundationmodel.NewDiscussBuilder().
		Id(discussId).
		Title(requestData.Title).
		//ProblemId(requestData.ProblemId).
		Content(requestData.Content).
		ModifyTime(nowTime).
		UpdateTime(nowTime).
		Build()

	err = discussService.PostEdit(ctx, discussId, discuss)
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
		Content    string    `json:"content"`
		ModifyTime time.Time `json:"modify_time"`
	}{
		Content:    content,
		ModifyTime: nowTime,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}
func (c *DiscussController) PostCommentCreate(ctx *gin.Context) {
	var requestData request.DiscussCommentEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.Content == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	discussService := foundationservice.GetDiscussService()

	discussId := requestData.DiscussId

	userId, hasAuth, err := discussService.CheckViewAuth(ctx, discussId)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	timeNow := metatime.GetTimeNow()

	discussComment := foundationmodel.NewDiscussCommentBuilder().
		DiscussId(discussId).
		Content(requestData.Content).
		AuthorId(userId).
		InsertTime(timeNow).
		UpdateTime(timeNow).
		Build()

	err = discussService.InsertDiscussComment(ctx, discussComment)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}

	var description string
	var needUpdateUrls []*foundationr2.R2ImageUrl
	description, needUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		requestData.Content,
		nil,
		metahttp.UrlJoin("discuss", strconv.Itoa(discussId)),
		metahttp.UrlJoin("discuss", strconv.Itoa(discussId), strconv.Itoa(discussComment.Id)),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if len(needUpdateUrls) > 0 {
		err := discussService.UpdateContent(ctx, discussComment.Id, description)
		if err != nil {
			return
		}
		err = service.GetR2ImageService().MoveImageAfterSave(needUpdateUrls)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *DiscussController) PostCommentEdit(ctx *gin.Context) {
	var requestData request.DiscussCommentEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.Content == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	discussService := foundationservice.GetDiscussService()

	// 校验是否有编辑权限（对评论）
	_, hasAuth, discussComment, err := discussService.CheckEditCommentAuth(ctx, requestData.Id)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !hasAuth {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	if discussComment.DiscussId != requestData.DiscussId {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	// 获取旧内容用于图片清理
	// 处理 markdown 内容和图片链接
	var needUpdateUrls []*foundationr2.R2ImageUrl
	content, needUpdateUrls, err := service.GetR2ImageService().ProcessContentFromMarkdown(
		requestData.Content,
		&discussComment.Content,
		metahttp.UrlJoin("discuss", strconv.Itoa(discussComment.DiscussId), strconv.Itoa(requestData.Id)),
		metahttp.UrlJoin("discuss", strconv.Itoa(discussComment.DiscussId), strconv.Itoa(requestData.Id)),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	nowTime := metatime.GetTimeNow()

	err = discussService.UpdateCommentContent(ctx, requestData.Id, discussComment.DiscussId, content, nowTime)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	err = service.GetR2ImageService().MoveImageAfterSave(needUpdateUrls)
	if err != nil {
		metapanic.ProcessError(err)
	}

	responseData := struct {
		Content    string    `json:"content"`
		UpdateTime time.Time `json:"update_time"`
	}{
		Content:    content,
		UpdateTime: nowTime,
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}
