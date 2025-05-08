package foundationmodel

import (
	"time"
)

type ProblemAttemptStatus int

var (
	ProblemAttemptStatusNone      ProblemAttemptStatus = 0
	ProblemAttemptStatusAttempt   ProblemAttemptStatus = 1
	ProblemAttemptStatusWAccepted ProblemAttemptStatus = 2
)

type ProblemAuth int

var (
	ProblemAuthPublic   ProblemAuth = 0 // 公开
	ProblemAuthPassword ProblemAuth = 1 // 密码，输入密码可以访问
	ProblemAuthPrivate  ProblemAuth = 2 // 私有，指定用户可以访问
)

type Problem struct {
	Id          string    `json:"id" bson:"_id"`
	Sort        int       `json:"sort" bson:"sort"` // 排序
	Title       string    `json:"title" bson:"title"`
	Description string    `json:"description" bson:"description"`
	Source      string    `json:"source" bson:"source"`
	Creator     string    `json:"creator" bson:"creator"`
	Privilege   int       `json:"privilege" bson:"privilege"`
	TimeLimit   int       `json:"time_limit" bson:"time_limit"`
	MemoryLimit int       `json:"memory_limit" bson:"memory_limit"`
	JudgeType   JudgeType `json:"judge_type" bson:"judge_type"`
	Accept      int       `json:"accept" bson:"accept"`
	Attempt     int       `json:"attempt" bson:"attempt"`
	Tags        []int     `json:"tags" bson:"tags"`
	InsertTime  time.Time `json:"insert_time" bson:"insert_time"`
	UpdateTime  time.Time `json:"update_time" bson:"update_time"`
	JudgeMd5    string    `json:"judge_md5" bson:"judge_md5"` // 判题数据的Md5标识
}

type ProblemBuilder struct {
	item *Problem
}

func NewProblemBuilder() *ProblemBuilder {
	return &ProblemBuilder{item: &Problem{}}
}

func (b *ProblemBuilder) Id(id string) *ProblemBuilder {
	b.item.Id = id
	return b
}

func (b *ProblemBuilder) Sort(sort int) *ProblemBuilder {
	b.item.Sort = sort
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

func (b *ProblemBuilder) Source(source string) *ProblemBuilder {
	b.item.Source = source
	return b
}

func (b *ProblemBuilder) Creator(creator string) *ProblemBuilder {
	b.item.Creator = creator
	return b
}

func (b *ProblemBuilder) Privilege(privilege int) *ProblemBuilder {
	b.item.Privilege = privilege
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

func (b *ProblemBuilder) JudgeType(judgeType JudgeType) *ProblemBuilder {
	b.item.JudgeType = judgeType
	return b
}

func (b *ProblemBuilder) Tags(tags []int) *ProblemBuilder {
	b.item.Tags = tags
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

func (b *ProblemBuilder) InsertTime(insertTime time.Time) *ProblemBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *ProblemBuilder) UpdateTime(updateTime time.Time) *ProblemBuilder {
	b.item.UpdateTime = updateTime
	return b
}

func (b *ProblemBuilder) Build() *Problem {
	return b.item
}
