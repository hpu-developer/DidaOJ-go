package foundationview

import (
	foundationmodel "foundation/foundation-model"
)

type BotGameView struct {
	foundationmodel.BotGame
}

type BotGameListView struct {
	Id           int    `json:"id"`
	GameKey      string `json:"game_key"`
	Title        string `json:"title"`
	Introduction string `json:"introduction"`
	PlayerMin    int    `json:"player_min"`
	PlayerMax    int    `json:"player_max"`
}
