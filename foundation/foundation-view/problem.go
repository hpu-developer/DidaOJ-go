package foundationview

import (
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"time"
)

type Problem struct {
	foundationmodel.Problem

	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	ModifierUsername string `json:"modifier_username"`
	ModifierNickname string `json:"modifier_nickname"`

	// Remote 信息
	OriginOj     *string `json:"origin_oj,omitempty"`
	OriginId     *string `json:"origin_id,omitempty"`
	OriginUrl    *string `json:"origin_url,omitempty"`
	OriginAuthor *string `json:"origin_author,omitempty"`
}

type ProblemJudgeData struct {
	Id  int    `json:"id"`
	Key string `json:"key"`

	Title string `json:"title"`

	JudgeType foundationjudge.JudgeType `json:"judge_type"`

	Inserter         int       `json:"inserter"`
	InserterUsername string    `json:"inserter_username"`
	InserterNickname string    `json:"inserter_nickname"`
	InsertTime       time.Time `json:"insert_time"`
	Modifier         int       `json:"modifier"`
	ModifierUsername string    `json:"modifier_username"`
	ModifierNickname string    `json:"modifier_nickname"`

	ModifyTime time.Time `json:"modify_time"`

	JudgeMd5 *string `json:"judge_md5"`
}

type ProblemViewApproveJudge struct {
	Id       int    `json:"id"`
	OriginOj string `json:"origin_oj"` // 题目来源的OJ
	OriginId string `json:"origin_id"` // 题目来源的Id
}

func (p *ProblemViewApproveJudge) TableName() string {
	return "problem_remote"
}

type ProblemViewList struct {
	Id      int    `json:"id"`
	Key     string `json:"key"`
	Title   string `json:"title"`
	Accept  int    `json:"accept"`
	Attempt int    `json:"attempt"`

	Tags []int `json:"tags"` // 题目标签列表
}

type ProblemViewTitle struct {
	Id    int    `json:"id"`
	Key   string `json:"key"`
	Title string `json:"title"`
}

type ProblemViewAuth struct {
	Id          string `json:"id" bson:"_id"`
	CreatorId   int    `json:"creator_id" bson:"creator_id"`     // 创建者Id
	Private     bool   `json:"private" bson:"private"`           // 是否私有
	Members     []int  `json:"members" bson:"members"`           // 访问权限用户列表，只有在私有题目时才有意义
	AuthMembers []int  `json:"auth_members" bson:"auth_members"` // 题目管理员，对题目有编辑权限
}

type ProblemViewAttempt struct {
	Id      string `json:"id" bson:"_id"`
	Accept  int    `json:"accept" bson:"accept"`
	Attempt int    `json:"attempt" bson:"attempt"`
}
