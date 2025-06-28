package foundationview

import foundationmodel "foundation/foundation-model"

type ProblemDailyEdit struct {
	foundationmodel.ProblemDaily

	ProblemKey   string `json:"problem_key"`
	ProblemTitle string `json:"problem_title"`

	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	ModifierUsername string `json:"modifier_username"`
	ModifierNickname string `json:"modifier_nickname"`
}
