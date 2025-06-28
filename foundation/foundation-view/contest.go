package foundationview

import (
	foundationenum "foundation/foundation-enum"
	"time"
)

type ContestViewLock struct {
	Id int `json:"id" bson:"_id"` // 比赛Id

	StartTime time.Time `json:"start_time,omitempty" bson:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty" bson:"end_time,omitempty"` // 结束时间

	Type foundationenum.ContestType `json:"type" bson:"type"` // 比赛类型

	AlwaysLock       bool           `json:"always_lock" bson:"always_lock"`                                   // 比赛结束后是否锁定排名，如果锁定则需要手动关闭（关闭时此值设为false）
	LockRankDuration *time.Duration `json:"lock_rank_duration,omitempty" bson:"lock_rank_duration,omitempty"` // 比赛结束前锁定排名的时长，空则不锁榜，锁榜期间榜单仅展示尝试次数，ACM模式下只可以查看自己的提交结果，OI模式下无法查看所有的提交结果
}
