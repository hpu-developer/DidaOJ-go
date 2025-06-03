package request

import "time"

type ContestEdit struct {
	Id           int       `json:"id"`                        // 比赛Id
	Title        string    `json:"title" validate:"required"` // 比赛标题
	Description  string    `json:"description"`
	Notification string    `json:"notification"`                   // 比赛通知
	StartTime    time.Time `json:"start_time" validate:"required"` // 比赛开启时间
	EndTime      time.Time `json:"end_time" validate:"required"`   // 比赛结束时间
	Problems     []string  `json:"problems" validate:"required"`   // 题目列表，逗号分隔的题目Id列表

	Members []int `json:"members"` // 成员列表，逗号分隔的用户Id列表
}
