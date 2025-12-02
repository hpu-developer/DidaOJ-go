package controller

import (
	"strconv"

	foundationerrorcode "foundation/error-code"
	foundationservice "foundation/foundation-service"
	metacontroller "meta/controller"
	metaerrorcode "meta/error-code"
	metaresponse "meta/meta-response"

	"github.com/gin-gonic/gin"
)

type BotController struct {
	metacontroller.Controller
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
