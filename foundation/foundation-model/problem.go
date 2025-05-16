package foundationmodel

import (
	foundationjudge "foundation/foundation-judge"
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
	Id   string `json:"id" bson:"_id"`
	Sort int    `json:"sort" bson:"sort"` // 排序

	OriginOj  *string `json:"origin_oj,omitempty" bson:"origin_oj,omitempty"`   // 题目来源的OJ
	OriginId  *string `json:"origin_id,omitempty" bson:"origin_id,omitempty"`   // 题目来源的Id
	OriginUrl *string `json:"origin_url,omitempty" bson:"origin_url,omitempty"` // 题目来源的Url

	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	Source      string `json:"source" bson:"source"`

	CreatorId       int     `json:"creator_id,omitempty" bson:"creator_id,omitempty"`
	CreatorUsername *string `json:"creator_username,omitempty" bson:"creator_username,omitempty"`
	CreatorNickname *string `json:"creator_nickname,omitempty" bson:"creator_nickname,omitempty"`

	Privilege   int                       `json:"privilege" bson:"privilege"`
	TimeLimit   int                       `json:"time_limit" bson:"time_limit"`
	MemoryLimit int                       `json:"memory_limit" bson:"memory_limit"` // 题目内存限制，单位为KB
	Tags        []int                     `json:"tags,omitempty" bson:"tags,omitempty"`
	JudgeType   foundationjudge.JudgeType `json:"judge_type" bson:"judge_type"`
	JudgeMd5    *string                   `json:"judge_md5,omitempty" bson:"judge_md5,omitempty"` // 判题数据的Md5标识
	InsertTime  time.Time                 `json:"insert_time" bson:"insert_time"`
	UpdateTime  time.Time                 `json:"update_time" bson:"update_time"`

	Accept  int `json:"accept" bson:"accept"`
	Attempt int `json:"attempt" bson:"attempt"`
}

type ProblemViewTitle struct {
	Id    string `json:"id" bson:"_id"`
	Title string `json:"title" bson:"title"`
}

type ProblemViewAttempt struct {
	Id      string `json:"id" bson:"_id"`
	Accept  int    `json:"accept" bson:"accept"`
	Attempt int    `json:"attempt" bson:"attempt"`
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

func (b *ProblemBuilder) OriginOj(originOj string) *ProblemBuilder {
	b.item.OriginOj = &originOj
	return b
}

func (b *ProblemBuilder) OriginId(originId string) *ProblemBuilder {
	b.item.OriginId = &originId
	return b
}

func (b *ProblemBuilder) OriginUrl(originUrl string) *ProblemBuilder {
	b.item.OriginUrl = &originUrl
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

func (b *ProblemBuilder) CreatorId(creatorId int) *ProblemBuilder {
	b.item.CreatorId = creatorId
	return b
}

func (b *ProblemBuilder) CreatorUsername(creatorUsername string) *ProblemBuilder {
	b.item.CreatorUsername = &creatorUsername
	return b
}

func (b *ProblemBuilder) CreatorNickname(creatorNickname string) *ProblemBuilder {
	b.item.CreatorNickname = &creatorNickname
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

func (b *ProblemBuilder) JudgeType(judgeType foundationjudge.JudgeType) *ProblemBuilder {
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
