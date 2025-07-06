package foundationview

import (
	foundationmodel "foundation/foundation-model"
	"time"
)

type JudgeJob struct {
	foundationmodel.JudgeJob

	ProblemKey          string `json:"problem_key,omitempty"`           // 题目Key
	ContestProblemIndex int    `json:"contest_problem_index,omitempty"` // 比赛题目索引

	JudgerName       string `json:"judger_name,omitempty"`
	InserterUsername string `json:"inserter_username,omitempty"`
	InserterNickname string `json:"inserter_nickname,omitempty"`

	CompileMessage *string `json:"compile_message,omitempty"`

	Task []*foundationmodel.JudgeTask `json:"task,omitempty" gorm:"-"`
}

type JudgeJobViewAuth struct {
	Id         int       `json:"id"`
	ContestId  int       `json:"contest_id,omitempty"`                       // 比赛ID
	Inserter   int       `json:"inserter_id" bson:"inserter_id"`             // 提交者UserId
	InsertTime time.Time `json:"inserter_time" bson:"inserter_time"`         // 申请时间
	Private    bool      `json:"private,omitempty" bson:"private,omitempty"` // 是否隐藏源码
}
