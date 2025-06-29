package foundationview

import "time"

type ContestRankDetail struct {
	Id               int            `json:"id" bson:"_id"` // 比赛Id
	Title            string         `json:"title" bson:"title"`
	StartTime        time.Time      `json:"start_time,omitempty" bson:"start_time,omitempty"`
	EndTime          time.Time      `json:"end_time,omitempty" bson:"end_time,omitempty"`                     // 结束时间
	LockRankDuration *time.Duration `json:"lock_rank_duration,omitempty" bson:"lock_rank_duration,omitempty"` // 比赛结束前锁定排名的时长，空则不锁榜，锁榜期间ACM模式下只可以查看自己的提交结果，OI模式下无法查看所有的提交结果
	AlwaysLock       bool           `json:"always_lock" bson:"always_lock"`                                   // 比赛结束后是否锁定排名，如果锁定则需要手动关闭（关闭时此值设为false）

	Problems []int `json:"problems,omitempty"` // 题目Id列表

	MembersIgnore []int `json:"members_ignore,omitempty"` // 忽略排名成员列表
}

type ContestRankProblem struct {
	Id      int        `json:"id,omitempty"`
	Index   uint8      `json:"index,omitempty"`   // 题目索引
	Attempt int        `json:"attempt,omitempty"` // 尝试次数（截止到首次AC）
	Ac      *time.Time `json:"ac,omitempty"`      // 首次AC时间
	Lock    int        `json:"lock,omitempty"`    // 未知的尝试次数（可能是锁榜期间的尝试次数）
}

type ContestRank struct {
	AuthorId       int     `json:"author_id"`                 // 提交者UserId
	AuthorUsername *string `json:"author_username,omitempty"` // 提交者用户名
	AuthorNickname *string `json:"author_nickname,omitempty"` // 提交者昵称

	Problems []*ContestRankProblem `json:"problems,omitempty"` // 题目提交情况
}
