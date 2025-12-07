package foundationview

import (
	foundationbot "foundation/foundation-bot"
	foundationmodel "foundation/foundation-model"
)

type BotReplayView struct {
	foundationmodel.BotReplay

	Bots []*BotCodePlayerView `json:"bots"`

	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	InserterEmail    string `json:"inserter_email"`
}

type BotReplayParamView struct {
	Status  foundationbot.BotGameStatus `json:"status"`
	Param   string                      `json:"param"`
	Message string                      `json:"message"`
}
