package foundationview

import (
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
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

type ContestDetailClone struct {
	Title            string                            `json:"title" gorm:"type:varchar(75);not null"`
	Type             foundationenum.ContestType        `json:"type" gorm:"type:tinyint(1) unsigned;"`
	ScoreType        foundationenum.ContestScoreType   `json:"score_type,omitempty" gorm:"type:tinyint"`
	LockRankDuration *time.Duration                    `json:"lock_rank_duration,omitempty" gorm:"type:bigint"`
	AlwaysLock       bool                              `json:"always_lock,omitempty" gorm:"type:tinyint(1)"`
	DiscussType      foundationenum.ContestDiscussType `json:"discuss_type,omitempty" gorm:"type:tinyint;comment:'讨论类型，0正常讨论，1仅查看自己的讨论'"`

	Problems []int `json:"problems" gorm:"-"` // 比赛题目列表
	Members  []int `json:"members" gorm:"-"`  // 比赛成员列表
}

type ContestDetailEdit struct {
	foundationmodel.Contest

	Problems []int            `json:"problems" gorm:"-"` // 比赛题目列表
	Members  []*ContestMember `json:"members" gorm:"-"`  // 比赛成员列表

	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	ModifierUsername string `json:"modifier_username"`
	ModifierNickname string `json:"modifier_nickname"`
}

type ContestProblemStatistics struct {
	ProblemId  int                                 `json:"problem_id,omitempty" gorm:"column:problem_id;primaryKey"`
	Index      uint8                               `json:"index" gorm:"column:index;type:tinyint(1) unsigned;"`
	Statistics map[foundationjudge.JudgeStatus]int `json:"statistics,omitempty" gorm:"-"`
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
