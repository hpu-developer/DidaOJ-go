package foundationview

import (
	foundationenum "foundation/foundation-enum"
	foundationmodel "foundation/foundation-model"
	"time"
)

type ContestDetail struct {
	foundationmodel.Contest

	Problems []*ContestProblemDetail `json:"problems" gorm:"-"` // 比赛题目列表

	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	ModifierUsername string `json:"modifier_username"`
	ModifierNickname string `json:"modifier_nickname"`
}

type ContestViewLock struct {
	Id int `json:"id"` // 比赛ID

	Inserter int `json:"inserter"`

	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"` // 结束时间

	Type foundationenum.ContestType `json:"type"` // 比赛类型

	AlwaysLock       bool           `json:"always_lock"`                  // 比赛结束后是否锁定排名，如果锁定则需要手动关闭（关闭时此值设为false）
	LockRankDuration *time.Duration `json:"lock_rank_duration,omitempty"` // 比赛结束前锁定排名的时长，空则不锁榜，锁榜期间榜单仅展示尝试次数，ACM模式下只可以查看自己的提交结果，OI模式下无法查看所有的提交结果
}

type ContestList struct {
	Id               int        `json:"id" gorm:"primaryKey;autoIncrement"`
	Title            string     `json:"title" gorm:"type:varchar(30)"`
	StartTime        *time.Time `json:"start_time"`
	EndTime          *time.Time `json:"end_time"`
	Private          bool       `json:"private"` // 是否私有比赛
	Inserter         int        `json:"inserter"`
	InserterUsername string     `json:"inserter_username"`
	InserterNickname string     `json:"inserter_nickname"`
}
