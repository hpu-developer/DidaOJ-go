package foundationview

import foundationjudge "foundation/foundation-judge"

type BotAgentView struct {
	Id       int                           `json:"id"`
	Language foundationjudge.JudgeLanguage `json:"language"`
	Code     string                        `json:"code"`
	Version  int                           `json:"version"`
	Name     string                        `json:"name"`

	Inserter         int    `json:"inserter"`
	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	InserterEmail    string `json:"inserter_email"`
}

type BotAgentPlayerView struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
