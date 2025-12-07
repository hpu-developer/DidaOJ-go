package foundationview

import foundationjudge "foundation/foundation-judge"

type BotCodeView struct {
	Id       int                           `json:"id"`
	Language foundationjudge.JudgeLanguage `json:"language"`
	Code     string                        `json:"code"`
	Version  int                           `json:"version"`
	Inserter int                           `json:"inserter"`
}

type BotCodePlayerView struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
