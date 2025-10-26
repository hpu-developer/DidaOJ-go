package foundationview

import (
	foundationmodel "foundation/foundation-model"
	"time"
)

type DiscussDetail struct {
	foundationmodel.Discuss

	ProblemKey   *string `json:"problem_key,omitempty"`   // 题目Key
	ProblemTitle *string `json:"problem_title,omitempty"` // 题目标题

	ContestTitle        *string `json:"contest_title,omitempty"`         // 比赛标题
	ContestProblemIndex int     `json:"contest_problem_index,omitempty"` // 比赛题目索引

	Tags []*foundationmodel.DiscussTag `json:"tags" gorm:"-"` // 比赛题目列表

	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	InserterEmail    string `json:"inserter_email"`
	ModifierUsername string `json:"modifier_username"`
	ModifierNickname string `json:"modifier_nickname"`
}

type DiscussCommentViewEdit struct {
	Id        int    `json:"id"`                   // 数据库索引时的Id
	DiscussId int    `json:"discuss_id,omitempty"` // 讨论的Id
	Content   string `json:"content"`              // 讨论内容
	Inserter  int    `json:"inserter"`             // 讨论作者Id
}

type DiscussList struct {
	Id         int       `json:"id"`    // 数据库索引时的Id
	Title      string    `json:"title"` // 讨论标题
	Inserter   int       `json:"inserter" gorm:"column:inserter;not null"`
	InsertTime time.Time `json:"insert_time,omitempty" gorm:"column:insert_time"`
	Modifier   int       `json:"modifier" gorm:"column:modifier;not null"`
	ModifyTime time.Time `json:"modify_time,omitempty" gorm:"column:modify_time"`
	Updater    int       `json:"updater" gorm:"column:updater;not null"`
	UpdateTime time.Time `json:"update_time,omitempty" gorm:"column:update_time"`

	ViewCount int  `json:"view_count,omitempty" gorm:"column:view_count"`
	ContestId *int `json:"contest_id,omitempty" gorm:"column:contest_id"` // 比赛Id
	ProblemId *int `json:"problem_id,omitempty" gorm:"column:problem_id"` // 题目Id

	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	ModifierUsername string `json:"modifier_username"`
	ModifierNickname string `json:"modifier_nickname"`

	ContestProblemIndex *int    `json:"contest_problem_index,omitempty"` // 题目索引
	ProblemKey          *string `json:"problem_key,omitempty"`
}
