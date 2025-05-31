package request

import "time"

type CollectionEdit struct {
	Id          int        `json:"id"`                        // 题集Id
	Title       string     `json:"title" validate:"required"` // 题集标题
	Description string     `json:"description"`
	StartTime   *time.Time `json:"start_time"` // 题集开启时间
	EndTime     *time.Time `json:"end_time"`   // 题集结束时间
	Users       []int      `json:"users"`      // 题集成员列表，逗号分隔的用户Id列表
	Problems    []string   `json:"problems"`   // 题目列表，逗号分隔的题目Id列表
}
