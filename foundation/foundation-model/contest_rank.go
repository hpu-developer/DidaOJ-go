package foundationmodel

import (
	"time"
)

type ContestRank struct {
	AuthorId       int     `json:"author_id" bson:"author_id"`                                 // 提交者UserId
	AuthorUsername *string `json:"author_username,omitempty" bson:"author_username,omitempty"` // 提交者用户名
	AuthorNickname *string `json:"author_nickname,omitempty" bson:"author_nickname,omitempty"` // 提交者昵称

	Problems map[string]struct {
		FirstAcTime *time.Time `json:"first_ac_time,omitempty" bson:"first_ac_time,omitempty"` // 首次AC时间
		SubmitCount int        `json:"submit_count" bson:"submit_count"`                       // 提交次数
	} `json:"problems" bson:"problems"` // 题目提交情况
}
