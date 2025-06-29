package foundationview

import foundationmodel "foundation/foundation-model"

type ProblemDaily struct {
	foundationmodel.ProblemDaily

	ProblemKey     string `json:"problem_key"`
	ProblemTitle   string `json:"problem_title"`
	ProblemTags    []int  `json:"problem_tags,omitempty" gorm:"-"`
	ProblemAccept  int    `json:"problem_accept"`
	ProblemAttempt int    `json:"problem_attempt"`
}

type ProblemDailyList struct {
	Key        string `json:"key"`
	Title      string `json:"title"`
	Accept     int    `json:"accept,omitempty"`
	Attempt    int    `json:"attempt,omitempty"`
	ProblemId  int    `json:"problem_id"`
	ProblemKey string `json:"problem_key"`

	Tags []int `json:"tags,omitempty" gorm:"-"`
}

type ProblemDailyEdit struct {
	foundationmodel.ProblemDaily

	ProblemKey   string `json:"problem_key"`
	ProblemTitle string `json:"problem_title"`

	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	ModifierUsername string `json:"modifier_username"`
	ModifierNickname string `json:"modifier_nickname"`
}
