package foundationview

import (
	foundationmodel "foundation/foundation-model"
)

type JudgeJob struct {
	foundationmodel.JudgeJob

	ContestProblemIndex int `json:"contest_problem_index"`

	JudgerName       string `json:"judger_name"`
	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
}
