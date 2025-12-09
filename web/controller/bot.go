package controller

import (
	"strconv"
	"time"

	foundationerrorcode "foundation/error-code"
	foundationmodel "foundation/foundation-model"
	foundationr2 "foundation/foundation-r2"
	foundationservice "foundation/foundation-service"
	foundationview "foundation/foundation-view"
	metacontroller "meta/controller"
	metaerrorcode "meta/error-code"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	metaresponse "meta/meta-response"
	metatime "meta/meta-time"
	"web/request"
	"web/service"

	"github.com/gin-gonic/gin"
)

type BotController struct {
	metacontroller.Controller
}

func (c *BotController) GetGame(ctx *gin.Context) {
	// 获取参数
	gameKey := ctx.Query("game_key")

	// 参数校验
	if gameKey == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	// 获取bot服务
	botService := foundationservice.GetBotService()

	// 获取游戏信息
	gameView, err := botService.GetGameByKey(ctx, gameKey)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if gameView == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	// 返回结果
	metaresponse.NewResponse(ctx, metaerrorcode.Success, gameView)
}

func (c *BotController) GetGameList(ctx *gin.Context) {
	// 获取bot服务
	botService := foundationservice.GetBotService()

	// 获取游戏列表
	gameListView, err := botService.GetGameList(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	// 返回结果
	metaresponse.NewResponse(ctx, metaerrorcode.Success, gameListView)
}

func (c *BotController) GetReplay(ctx *gin.Context) {
	// 获取参数
	gameKey := ctx.Query("game_key")
	replayIdStr := ctx.Query("replay_id")

	// 参数校验
	if gameKey == "" || replayIdStr == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	replayId, err := strconv.Atoi(replayIdStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	// 获取bot服务
	botService := foundationservice.GetBotService()

	// 获取完整的replay信息
	botReplayView, err := botService.GetBotReplayById(ctx, replayId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	if botReplayView == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	// 返回结果
	metaresponse.NewResponse(ctx, metaerrorcode.Success, botReplayView)
}

func (c *BotController) GetReplayParam(ctx *gin.Context) {
	// 获取参数
	gameKey := ctx.Query("game_key")
	replayIdStr := ctx.Query("replay_id")

	// 参数校验
	if gameKey == "" || replayIdStr == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	replayId, err := strconv.Atoi(replayIdStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	// 获取bot服务
	botService := foundationservice.GetBotService()

	// 只获取需要的字段：status、param 和 message，避免查询不必要的数据
	param, err := botService.GetBotReplayParamById(ctx, replayId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if param == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	// 返回结果
	metaresponse.NewResponse(ctx, metaerrorcode.Success, param)
}

func (c *BotController) GetReplayList(ctx *gin.Context) {
	// 获取参数
	gameKey := ctx.Query("game_key")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")

	// 转换参数
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取bot服务
	botService := foundationservice.GetBotService()

	// 获取回放列表
	replayList, total, err := botService.GetBotReplayList(ctx, gameKey, page, pageSize)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	// 返回结果
	responseData := struct {
		List  []*foundationview.BotReplayView `json:"list"`
		Total int64                           `json:"total"`
	}{
		List:  replayList,
		Total: total,
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

// PostGameEdit 编辑Bot游戏信息
func (c *BotController) PostGameEdit(ctx *gin.Context) {
	var requestData request.BotGameEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	// 验证请求数据
	ok, errorCode := requestData.CheckRequest()
	if !ok {
		metaresponse.NewResponse(ctx, errorCode, nil)
		return
	}

	// 获取bot服务
	botService := foundationservice.GetBotService()

	botGameId := requestData.GameId

	userId, ok, err := botService.CheckGameEditAuth(ctx, botGameId)
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

	oldDescription, err := foundationservice.GetBotService().GetGameDescription(ctx, botGameId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	var needUpdateUrls []*foundationr2.R2ImageUrl
	description, needUpdateUrls, err = service.GetR2ImageService().ProcessContentFromMarkdown(
		description,
		oldDescription,
		metahttp.UrlJoin("bot_game", strconv.Itoa(botGameId)),
		metahttp.UrlJoin("bot_game", strconv.Itoa(botGameId)),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	requestData.Description = description

	nowTime := metatime.GetTimeNow()

	botGame := foundationmodel.NewBotGameBuilder().
		Title(requestData.Title).
		Description(requestData.Description).
		JudgeCode(requestData.JudgeCode).
		Modifier(userId).
		ModifyTime(nowTime).
		Build()

	err = foundationservice.GetBotService().UpdateBotGame(ctx, botGameId, botGame)
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
		ModifyTime  time.Time `json:"modify_time"`
	}{
		Description: description,
		ModifyTime:  nowTime,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}
