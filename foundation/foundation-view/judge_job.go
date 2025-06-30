package foundationview

import (
	foundationmodel "foundation/foundation-model"
)

type JudgeJob struct {
	foundationmodel.JudgeJob

	ContestProblemIndex int `json:"contest_problem_index,omitempty"`

	JudgerName       string `json:"judger_name,omitempty"`
	InserterUsername string `json:"inserter_username,omitempty"`
	InserterNickname string `json:"inserter_nickname,omitempty"`

	CompileMessage *string `json:"compile_message,omitempty"`

	Task []*foundationmodel.JudgeTask `json:"task,omitempty" gorm:"-"`
}
