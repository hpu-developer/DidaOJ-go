package foundationmodel

import (
	"time"
)

type CollectionRankView struct {
	Id        int        `json:"id" bson:"_id"` // 比赛Id
	StartTime *time.Time `json:"start_time,omitempty" bson:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty" bson:"end_time,omitempty"` // 结束时间
	Problems  []string   `json:"problems,omitempty" bson:"problems,omitempty"` // 题目Id列表
	Members   []int      `json:"members,omitempty" bson:"members,omitempty"`
}

type CollectionRankProblem struct {
	Id      string     `json:"id" bson:"id"`                     // 题目索引
	Attempt int        `json:"attempt" bson:"attempt"`           // 尝试次数（截止到首次AC）
	Ac      *time.Time `json:"ac,omitempty" bson:"ac,omitempty"` // 首次AC时间
}

type CollectionRank struct {
	AuthorId       int     `json:"author_id" bson:"author_id"`                                 // 提交者UserId
	AuthorUsername *string `json:"author_username,omitempty" bson:"author_username,omitempty"` // 提交者用户名
	AuthorNickname *string `json:"author_nickname,omitempty" bson:"author_nickname,omitempty"` // 提交者昵称

	Accept int `json:"accept" bson:"accept"` // 通过数
}
