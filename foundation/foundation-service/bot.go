package foundationservice

import (
	"context"
	foundationauth "foundation/foundation-auth"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"meta/singleton"

	"github.com/gin-gonic/gin"
)

type BotService struct {
}

var singletonBotService = singleton.Singleton[BotService]{}

func GetBotService() *BotService {
	return singletonBotService.GetInstance(
		func() *BotService {
			return &BotService{}
		},
	)
}

func (s *BotService) CheckGameEditAuth(ctx *gin.Context, id int) (
	int,
	bool,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		hasAuth, err = foundationdao.GetBotGameDao().CheckGameEditAuth(ctx, id, userId)
		if err != nil {
			return userId, false, err
		}
		if !hasAuth {
			return userId, false, nil
		}
	}
	return userId, true, nil
}

func (s *BotService) GetGameByKey(ctx context.Context, key string) (*foundationview.BotGameView, error) {
	// 获取bot game信息
	botGame, err := foundationdao.GetBotGameDao().GetBotGameByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if botGame == nil {
		return nil, nil
	}
	// 构建返回视图
	botGameView := &foundationview.BotGameView{
		BotGame: *botGame,
	}
	return botGameView, nil
}

// GetGameDescription 获取游戏描述
func (s *BotService) GetGameDescription(ctx context.Context, id int) (*string, error) {
	return foundationdao.GetBotGameDao().GetBotGameDescription(ctx, id)
}

// GetBotReplayById 根据ID获取BotReplay信息
func (s *BotService) GetBotReplayById(ctx context.Context, id int) (*foundationview.BotReplayView, error) {
	// 获取bot replay信息
	botReplay, err := foundationdao.GetBotReplayDao().GetBotReplayById(ctx, id)
	if err != nil {
		return nil, err
	}
	if botReplay == nil {
		return nil, nil
	}

	// 构建返回视图
	botReplayView := &foundationview.BotReplayView{
		BotReplay: *botReplay,
	}

	// 获取用户信息
	if botReplay.Inserter > 0 {
		user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, botReplay.Inserter)
		if err == nil && user != nil {
			botReplayView.InserterUsername = user.Username
			botReplayView.InserterNickname = user.Nickname
			botReplayView.InserterEmail = user.Email
		}
	}

	return botReplayView, nil
}

// GetBotReplayParamById 根据ID获取BotReplay的状态、参数和消息（只查询需要的字段）
func (s *BotService) GetBotReplayParamById(ctx context.Context, id int) (*foundationview.BotReplayParamView, error) {
	return foundationdao.GetBotReplayDao().GetBotReplayParamById(ctx, id)
}

func (s *BotService) UpdateBotGame(
	ctx context.Context,
	botGameId int,
	botGame *foundationmodel.BotGame,
) error {
	return foundationdao.GetBotGameDao().UpdateBotGame(ctx, botGameId, botGame)
}
