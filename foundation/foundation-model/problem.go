package foundationmodel

import (
	foundationjudge "foundation/foundation-judge"
	"time"
)

type Problem struct {
	Id  int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`           // 问题ID
	Key string `json:"key" gorm:"column:key;type:varchar(18);unique;not null"` // 问题对外标识

	Title       string  `json:"title" gorm:"column:title;type:varchar(50);not null"`      // 标题
	Description string  `json:"description" gorm:"column:description;type:text;not null"` // 描述
	Source      *string `json:"source,omitempty" gorm:"column:source;type:varchar(120)"`  // 来源

	TimeLimit   int                       `json:"time_limit" gorm:"column:time_limit;not null"`     // 时间限制，单位ms
	MemoryLimit int                       `json:"memory_limit" gorm:"column:memory_limit;not null"` // 内存限制，单位KB
	JudgeType   foundationjudge.JudgeType `json:"judge_type" gorm:"column:judge_type;not null"`     // 判题类型

	Private bool `json:"private,omitempty" gorm:"column:private;not null"` // 是否私有

	Inserter   int       `json:"inserter" gorm:"column:inserter;not null"`       // 创建者ID
	InsertTime time.Time `json:"insert_time" gorm:"column:insert_time;not null"` // 创建时间
	Modifier   int       `json:"modifier" gorm:"column:modifier;not null"`       // 修改者ID
	ModifyTime time.Time `json:"modify_time" gorm:"column:modify_time;not null"` // 修改时间

	Accept  int `json:"accept,omitempty" gorm:"column:accept;not null"`   // 通过人数
	Attempt int `json:"attempt,omitempty" gorm:"column:attempt;not null"` // 尝试人数
}

// TableName 指定表名
func (Problem) TableName() string {
	return "problem"
}

type ProblemBuilder struct {
	item *Problem
}

func NewProblemBuilder() *ProblemBuilder {
	return &ProblemBuilder{item: &Problem{}}
}

func (b *ProblemBuilder) Id(id int) *ProblemBuilder {
	b.item.Id = id
	return b
}

func (b *ProblemBuilder) Key(key string) *ProblemBuilder {
	b.item.Key = key
	return b
}

func (b *ProblemBuilder) Title(title string) *ProblemBuilder {
	b.item.Title = title
	return b
}

func (b *ProblemBuilder) Description(description string) *ProblemBuilder {
	b.item.Description = description
	return b
}

func (b *ProblemBuilder) Source(source *string) *ProblemBuilder {
	b.item.Source = source
	return b
}

func (b *ProblemBuilder) Private(private bool) *ProblemBuilder {
	b.item.Private = private
	return b
}

func (b *ProblemBuilder) TimeLimit(timeLimit int) *ProblemBuilder {
	b.item.TimeLimit = timeLimit
	return b
}

func (b *ProblemBuilder) MemoryLimit(memoryLimit int) *ProblemBuilder {
	b.item.MemoryLimit = memoryLimit
	return b
}

func (b *ProblemBuilder) JudgeType(judgeType foundationjudge.JudgeType) *ProblemBuilder {
	b.item.JudgeType = judgeType
	return b
}

func (b *ProblemBuilder) Inserter(inserter int) *ProblemBuilder {
	b.item.Inserter = inserter
	return b
}

func (b *ProblemBuilder) InsertTime(insertTime time.Time) *ProblemBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *ProblemBuilder) Modifier(modifier int) *ProblemBuilder {
	b.item.Modifier = modifier
	return b
}

func (b *ProblemBuilder) ModifyTime(modifyTime time.Time) *ProblemBuilder {
	b.item.ModifyTime = modifyTime
	return b
}

func (b *ProblemBuilder) Accept(accept int) *ProblemBuilder {
	b.item.Accept = accept
	return b
}

func (b *ProblemBuilder) Attempt(attempt int) *ProblemBuilder {
	b.item.Attempt = attempt
	return b
}

func (b *ProblemBuilder) Build() *Problem {
	return b.item
}
