package foundationmodel

import (
	foundationjudge "foundation/foundation-judge"
	"time"
)

type ContestType int

var (
	ContestTypeAcm       ContestType = 0 // ACM模式比赛=
	ContestTypeOiHighest ContestType = 1 // OI模式比赛，以最高分提交为准
	ContestTypeOiLast    ContestType = 2 // OI模式比赛，以最后一次提交为准
)

type ContestAuth int

var (
	ContestAuthPublic   ContestAuth = 0 // 公开
	ContestAuthPassword ContestAuth = 1 // 密码，输入密码可以访问
	ContestAuthPrivate  ContestAuth = 2 // 私有，指定用户可以访问
)

type ContestScoreType int

var (
	ContestScoreTypeNone     ContestScoreType = 0 // 不启用分数排名，一般用于ACM模式，ACM启用则仅用于展示
	ContestScoreTypeAccepted ContestScoreType = 1 // 题目Accepted才认为得分
	ContestScoreTypePartial  ContestScoreType = 2 // 题目部分得分也按比例得分
)

type VirtualReplay struct {
	RankList []struct {
		Nickname         string        `json:"nickname" bson:"nickname"`                     // 昵称
		ContestProblemId string        `json:"contest_problem_id" bson:"contest_problem_id"` // 比赛题目Id
		AcceptedTime     time.Duration `json:"accepted_time" bson:"accepted_time"`           // 成功时间（距开始时间）
		AttemptedCount   int           `json:"attempted_count" bson:"attempted_count"`       // 尝试次数
	} `json:"rank_list" bson:"rank_list"` // 关注排名列表
	JudgeList []struct {
		Nickname         string                      `json:"nickname" bson:"nickname"`                     // 昵称
		ContestProblemId string                      `json:"contest_problem_id" bson:"contest_problem_id"` // 比赛题目Id
		JudgeTime        time.Duration               `json:"judge_time" bson:"judge_time"`                 // 判题时间（距开始时间）
		JudgeStatus      foundationjudge.JudgeStatus `json:"judge_status" bson:"judge_status"`             // 判题状态
	} `json:"judge_list" bson:"judge_list"` // 评测列表
}

type Contest struct {
	Id            int                   `json:"id" bson:"_id"`                                            // 数据库索引时真正的Id
	Title         string                `json:"title" bson:"title"`                                       // 比赛标题
	Descriptions  []*ContestDescription `json:"descriptions,omitempty" bson:"descriptions,omitempty"`     // 比赛描述
	Notification  string                `json:"notification,omitempty" bson:"notification,omitempty"`     // 比赛通知，会醒目的出现在大部分页面
	StartTime     time.Time             `json:"start_time" bson:"start_time"`                             // 比赛开始时间
	EndTime       time.Time             `json:"end_time" bson:"end_time"`                                 // 比赛结束时间
	OwnerId       int                   `json:"owner_id" bson:"owner_id"`                                 // 比赛组织者
	OwnerUsername *string               `json:"owner_username,omitempty" bson:"owner_username,omitempty"` // 比赛组织者用户名
	OwnerNickname *string               `json:"owner_nickname,omitempty" bson:"owner_nickname,omitempty"` // 比赛组织者昵称
	Languages     []string              `json:"languages,omitempty" bson:"languages,omitempty"`           // 允许的语言

	CreateTime time.Time `json:"create_time" bson:"create_time"` // 创建时间

	// 权限相关
	Auth     ContestAuth `json:"auth" bson:"auth"`                             // 比赛权限
	Password *string     `json:"password,omitempty" bson:"password,omitempty"` // 比赛密码
	Members  []int       `json:"members,omitempty" bson:"members,omitempty"`   // 比赛成员，只有在私有比赛时才会使用

	// 排名相关
	Type          ContestType      `json:"type" bson:"type"`                                         // 比赛类型
	ScoreType     ContestScoreType `json:"score_type" bson:"score_type"`                             // 分数类型
	VirtualReplay *VirtualReplay   `json:"virtual_replay,omitempty" bson:"virtual_replay,omitempty"` // 虚拟赛信息

	AlwaysLock       bool           `json:"always_lock" bson:"always_lock"`                                   // 比赛结束后是否锁定排名，如果锁定则需要手动关闭（关闭时此值设为false）
	LockRankDuration *time.Duration `json:"lock_rank_duration,omitempty" bson:"lock_rank_duration,omitempty"` // 比赛结束前锁定排名的时长，空则不锁榜，锁榜期间榜单仅展示尝试次数，ACM模式下只可以查看自己的提交结果，OI模式下无法查看所有的提交结果

	// 题目相关
	Problems []*ContestProblem `json:"problems,omitempty" bson:"problems,omitempty"` // 题目Id列表

	// Migrate相关
	MigrateJolId  int `json:"-" bson:"-"` // Jol中的Id
	MigrateVhojId int `json:"-" bson:"-"` // Vhoj中的Id
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

func (b *ContestBuilder) Descriptions(descriptions []*ContestDescription) *ContestBuilder {
	b.item.Descriptions = descriptions
	return b
}

func (b *ContestBuilder) Notification(notification string) *ContestBuilder {
	b.item.Notification = notification
	return b
}

func (b *ContestBuilder) CreateTime(createTime time.Time) *ContestBuilder {
	b.item.CreateTime = createTime
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

func (b *ContestBuilder) OwnerId(ownerId int) *ContestBuilder {
	b.item.OwnerId = ownerId
	return b
}

func (b *ContestBuilder) OwnerUsername(ownerUsername string) *ContestBuilder {
	b.item.OwnerUsername = &ownerUsername
	return b
}

func (b *ContestBuilder) OwnerNickname(ownerNickname string) *ContestBuilder {
	b.item.OwnerNickname = &ownerNickname
	return b
}

func (b *ContestBuilder) Languages(languages []string) *ContestBuilder {
	b.item.Languages = languages
	return b
}

func (b *ContestBuilder) Auth(auth ContestAuth) *ContestBuilder {
	b.item.Auth = auth
	return b
}

func (b *ContestBuilder) Password(password *string) *ContestBuilder {
	b.item.Password = password
	return b
}

func (b *ContestBuilder) Members(members []int) *ContestBuilder {
	b.item.Members = members
	return b
}

func (b *ContestBuilder) Type(typ ContestType) *ContestBuilder {
	b.item.Type = typ
	return b
}

func (b *ContestBuilder) ScoreType(scoreType ContestScoreType) *ContestBuilder {
	b.item.ScoreType = scoreType
	return b
}

func (b *ContestBuilder) Build() *Contest {
	return b.item
}
