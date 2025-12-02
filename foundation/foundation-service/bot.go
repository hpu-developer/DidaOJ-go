package foundationservice

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationview "foundation/foundation-view"
	"meta/singleton"
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
