package foundationview

import (
	foundationmodel "foundation/foundation-model"
)

type BotReplayView struct {
	foundationmodel.BotReplay

	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	InserterEmail    string `json:"inserter_email"`
}
