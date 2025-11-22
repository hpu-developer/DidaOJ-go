package foundationservice

import (
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
