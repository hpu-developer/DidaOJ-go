package foundationview

import (
	foundationjudge "foundation/foundation-judge"
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
	InserterEmail    string `json:"inserter_email,omitempty"`

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

type JudgeJobRank struct {
	Id               int                           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Language         foundationjudge.JudgeLanguage `json:"language" gorm:"column:language;not null"`
	CodeLength       int                           `json:"code_length" gorm:"column:code_length;not null"`
	Time             int                           `json:"time,omitempty" gorm:"column:time"`
	Memory           int                           `json:"memory,omitempty" gorm:"column:memory"`
	Inserter         int                           `json:"inserter" gorm:"column:inserter;not null"`
	InserterUsername string                        `json:"inserter_username,omitempty"`
	InserterNickname string                        `json:"inserter_nickname,omitempty"`
	InserterEmail    string                        `json:"inserter_email,omitempty"`
	InsertTime       time.Time                     `json:"insert_time" gorm:"column:insert_time;not null"`
	Private          bool                          `json:"private,omitempty" gorm:"column:private;not null"`
}
