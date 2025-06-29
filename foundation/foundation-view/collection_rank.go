package foundationview

import "time"

type CollectionRankDetail struct {
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`

	Problems []int `json:"problems"` // 题目数量
}

type CollectionRank struct {
	Inserter         int     `json:"inserter"`                    // 提交者UserId
	InserterUsername *string `json:"inserter_username,omitempty"` // 提交者用户名
	InserterNickname *string `json:"inserter_nickname,omitempty"` // 提交者昵称

	Accept int `json:"accept" bson:"accept"` // 通过数
}
