package foundationmodel

import (
	foundationenum "foundation/foundation-enum"
	"time"
)

type Contest struct {
	Id               int                               `json:"id" gorm:"primaryKey;autoIncrement"`
	Title            string                            `json:"title" gorm:"type:varchar(75);not null"`
	Description      *string                           `json:"description,omitempty" gorm:"type:text"`
	Notification     *string                           `json:"notification,omitempty" gorm:"type:varchar(100)"`
	StartTime        time.Time                         `json:"start_time,omitempty" gorm:"type:datetime"`
	EndTime          time.Time                         `json:"end_time,omitempty" gorm:"type:datetime"`
	Inserter         int                               `json:"inserter,omitempty"`
	InsertTime       time.Time                         `json:"insert_time,omitempty" gorm:"type:datetime"`
	Modifier         int                               `json:"modifier,omitempty"`
	ModifyTime       time.Time                         `json:"modify_time,omitempty" gorm:"type:datetime"`
	Private          bool                              `json:"private,omitempty" gorm:"type:tinyint(1)"`
	Password         *string                           `json:"password,omitempty" gorm:"type:varchar(35)"`
	SubmitAnytime    bool                              `json:"submit_anytime,omitempty" gorm:"type:tinyint(1)"`
	Type             foundationenum.ContestType        `json:"type,omitempty" gorm:"type:tinyint"`
	ScoreType        foundationenum.ContestScoreType   `json:"score_type,omitempty" gorm:"type:tinyint"`
	LockRankDuration *time.Duration                    `json:"lock_rank_duration,omitempty" gorm:"type:bigint"`
	AlwaysLock       bool                              `json:"always_lock,omitempty" gorm:"type:tinyint(1)"`
	DiscussType      foundationenum.ContestDiscussType `json:"discuss_type,omitempty" gorm:"type:tinyint;comment:'讨论类型，0正常讨论，1仅查看自己的讨论'"`
}

func (*Contest) TableName() string {
	return "contest"
}

type ContestBuilder struct {
	item *Contest
}

func NewContestBuilder() *ContestBuilder {
	return &ContestBuilder{item: &Contest{}}
}

func (b *ContestBuilder) Id(id int) *ContestBuilder {
	b.item.Id = id
	return b
}

func (b *ContestBuilder) Title(title string) *ContestBuilder {
	b.item.Title = title
	return b
}

func (b *ContestBuilder) Description(description *string) *ContestBuilder {
	b.item.Description = description
	return b
}

func (b *ContestBuilder) Notification(notification *string) *ContestBuilder {
	b.item.Notification = notification
	return b
}

func (b *ContestBuilder) StartTime(startTime time.Time) *ContestBuilder {
	b.item.StartTime = startTime
	return b
}

func (b *ContestBuilder) EndTime(endTime time.Time) *ContestBuilder {
	b.item.EndTime = endTime
	return b
}

func (b *ContestBuilder) Inserter(inserter int) *ContestBuilder {
	b.item.Inserter = inserter
	return b
}

func (b *ContestBuilder) InsertTime(insertTime time.Time) *ContestBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *ContestBuilder) Modifier(modifier int) *ContestBuilder {
	b.item.Modifier = modifier
	return b
}

func (b *ContestBuilder) ModifyTime(modifyTime time.Time) *ContestBuilder {
	b.item.ModifyTime = modifyTime
	return b
}

func (b *ContestBuilder) Private(private bool) *ContestBuilder {
	b.item.Private = private
	return b
}

func (b *ContestBuilder) Password(password *string) *ContestBuilder {
	b.item.Password = password
	return b
}

func (b *ContestBuilder) SubmitAnytime(submitAnytime bool) *ContestBuilder {
	b.item.SubmitAnytime = submitAnytime
	return b
}

func (b *ContestBuilder) Type(typ foundationenum.ContestType) *ContestBuilder {
	b.item.Type = typ
	return b
}

func (b *ContestBuilder) ScoreType(scoreType foundationenum.ContestScoreType) *ContestBuilder {
	b.item.ScoreType = scoreType
	return b
}

func (b *ContestBuilder) LockRankDuration(lockRankDuration *time.Duration) *ContestBuilder {
	b.item.LockRankDuration = lockRankDuration
	return b
}

func (b *ContestBuilder) AlwaysLock(alwaysLock bool) *ContestBuilder {
	b.item.AlwaysLock = alwaysLock
	return b
}

func (b *ContestBuilder) DiscussType(discussType foundationenum.ContestDiscussType) *ContestBuilder {
	b.item.DiscussType = discussType
	return b
}

func (b *ContestBuilder) Build() *Contest {
	return b.item
}
