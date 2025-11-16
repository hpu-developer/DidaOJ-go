package foundationmodel

import (
	foundationjudge "foundation/foundation-judge"
	foundationrun "foundation/foundation-run"
	"time"
)

type RunJob struct {
	Id         int                           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Inserter   int                           `json:"inserter" gorm:"column:inserter;not null"`
	InsertTime time.Time                     `json:"insert_time" gorm:"column:insert_time;not null"`
	Language   foundationjudge.JudgeLanguage `json:"language" gorm:"column:language;not null"`
	Code       string                        `json:"code" gorm:"column:code;type:text;not null"`
	Input      string                        `json:"input,omitempty" gorm:"column:input;type:text"`
	Status     foundationrun.RunStatus       `json:"status" gorm:"column:status;not null"`
	Judger     string                        `json:"judger" gorm:"column:judger;not null;type:varchar(10)"`
	Time       int                           `json:"time,omitempty" gorm:"column:time"`
	Memory     int                           `json:"memory,omitempty" gorm:"column:memory"`
	Content    string                        `json:"content,omitempty" gorm:"column:content;type:text"`
}

// TableName 重写表名
func (RunJob) TableName() string {
	return "run_job"
}

type RunJobBuilder struct {
	item *RunJob
}

func NewRunJobBuilder() *RunJobBuilder {
	return &RunJobBuilder{
		item: &RunJob{},
	}
}

func (b *RunJobBuilder) Id(id int) *RunJobBuilder {
	b.item.Id = id
	return b
}

func (b *RunJobBuilder) Inserter(inserter int) *RunJobBuilder {
	b.item.Inserter = inserter
	return b
}

func (b *RunJobBuilder) Code(code string) *RunJobBuilder {
	b.item.Code = code
	return b
}

func (b *RunJobBuilder) Input(input string) *RunJobBuilder {
	b.item.Input = input
	return b
}

func (b *RunJobBuilder) Language(language foundationjudge.JudgeLanguage) *RunJobBuilder {
	b.item.Language = language
	return b
}

func (b *RunJobBuilder) Error(content string) *RunJobBuilder {
	b.item.Content = content
	return b
}

func (b *RunJobBuilder) Status(status foundationrun.RunStatus) *RunJobBuilder {
	b.item.Status = status
	return b
}

func (b *RunJobBuilder) Judger(judger string) *RunJobBuilder {
	b.item.Judger = judger
	return b
}

func (b *RunJobBuilder) Time(time int) *RunJobBuilder {
	b.item.Time = time
	return b
}

func (b *RunJobBuilder) Memory(memory int) *RunJobBuilder {
	b.item.Memory = memory
	return b
}

func (b *RunJobBuilder) InsertTime(insertTime time.Time) *RunJobBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *RunJobBuilder) Build() *RunJob {
	return b.item
}
