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

	Private  bool    `json:"private"`
	Password *string `json:"password,omitempty"` // 比赛密码，私有比赛时需要
	Members  []int   `json:"members"`            // 成员列表，逗号分隔的用户Id列表

	LockRankDuration int64 `json:"lock_rank_duration,omitempty"` // 锁榜时长，空则不锁榜（单位秒）
	AlwaysLock       bool  `json:"always_lock"`                  // 比赛结束后是否锁定排名，如果锁定则需要手动关闭（关闭时此值设为false）

	SubmitAnytime bool `json:"submit_anytime,omitempty"`
}
