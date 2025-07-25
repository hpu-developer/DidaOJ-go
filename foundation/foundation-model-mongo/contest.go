package foundationmodelmongo

import (
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	"time"
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
	Id            int       `json:"id" bson:"_id"`      // 数据库索引时真正的Id
	Title         string    `json:"title" bson:"title"` // 比赛标题
	Description   string    `json:"description" bson:"description"`
	Notification  string    `json:"notification,omitempty" bson:"notification,omitempty"`     // 比赛通知，会醒目的出现在大部分页面
	StartTime     time.Time `json:"start_time" bson:"start_time"`                             // 比赛开始时间
	EndTime       time.Time `json:"end_time" bson:"end_time"`                                 // 比赛结束时间
	OwnerId       int       `json:"owner_id" bson:"owner_id"`                                 // 比赛组织者
	OwnerUsername *string   `json:"owner_username,omitempty" bson:"owner_username,omitempty"` // 比赛组织者用户名
	OwnerNickname *string   `json:"owner_nickname,omitempty" bson:"owner_nickname,omitempty"` // 比赛组织者昵称
	Languages     []string  `json:"languages,omitempty" bson:"languages,omitempty"`           // 允许的语言

	CreateTime time.Time `json:"create_time" bson:"create_time"` // 创建时间
	UpdateTime time.Time `json:"update_time" bson:"update_time"` // 更新时间

	// 权限相关
	Private  bool    `json:"private,omitempty" bson:"private,omitempty"`   // 比赛权限
	Password *string `json:"password,omitempty" bson:"password,omitempty"` // 比赛密码
	Members  []int   `json:"members,omitempty" bson:"members,omitempty"`   // 比赛成员，只有在私有比赛时才会使用

	SubmitAnytime bool `json:"submit_anytime,omitempty" bson:"submit_anytime,omitempty"` // 是否允许在比赛结束后提交，默认为false

	Authors     []int `json:"authors,omitempty" bson:"authors,omitempty"`           // 作者列表，用于展示出题人
	AuthMembers []int `json:"auth_members,omitempty" bson:"auth_members,omitempty"` // 管理员，可以对本比赛进行编辑与查看

	Volunteers []int `json:"volunteers,omitempty" bson:"volunteers,omitempty"` // 志愿者列表，可以对本比赛的进度查看，方便发气球等工作

	// 排名相关
	Type          foundationenum.ContestType      `json:"type" bson:"type"`                                         // 比赛类型
	ScoreType     foundationenum.ContestScoreType `json:"score_type" bson:"score_type"`                             // 分数类型
	VirtualReplay *VirtualReplay                  `json:"virtual_replay,omitempty" bson:"virtual_replay,omitempty"` // 虚拟赛信息

	VMembers []int `json:"v_members,omitempty" bson:"v_members,omitempty"` // 忽略排名的成员列表

	LockRankDuration *time.Duration `json:"lock_rank_duration,omitempty" bson:"lock_rank_duration,omitempty"` // 比赛结束前锁定排名的时长，空则不锁榜，锁榜期间ACM模式下只可以查看自己的提交结果，榜单仅展示尝试次数，OI模式下无法查看所有的提交结果，榜单保持锁榜前的状态
	AlwaysLock       bool           `json:"always_lock" bson:"always_lock"`                                   // 比赛结束后是否锁定排名，如果锁定则需要手动关闭（关闭时此值设为false）

	// 题目相关
	Problems []*ContestProblem `json:"problems,omitempty" bson:"problems,omitempty"` // 题目Id列表

	DiscussType foundationenum.ContestDiscussType `json:"discuss_type,omitempty" bson:"discuss_type,omitempty"` // 讨论类型，0表示不启用讨论，1表示启用讨论

	// Migrate相关
	MigrateJolId  int `json:"-" bson:"-"` // Jol中的Id
	MigrateVhojId int `json:"-" bson:"-"` // Vhoj中的Id
}

type ContestViewLock struct {
	Id int `json:"id" bson:"_id"` // 比赛Id

	OwnerId     int   `json:"owner_id" bson:"owner_id"`
	AuthMembers []int `json:"auth_members" bson:"auth_members"`

	StartTime time.Time `json:"start_time,omitempty" bson:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty" bson:"end_time,omitempty"` // 结束时间

	Type foundationenum.ContestType `json:"type" bson:"type"` // 比赛类型

	AlwaysLock       bool           `json:"always_lock" bson:"always_lock"`                                   // 比赛结束后是否锁定排名，如果锁定则需要手动关闭（关闭时此值设为false）
	LockRankDuration *time.Duration `json:"lock_rank_duration,omitempty" bson:"lock_rank_duration,omitempty"` // 比赛结束前锁定排名的时长，空则不锁榜，锁榜期间榜单仅展示尝试次数，ACM模式下只可以查看自己的提交结果，OI模式下无法查看所有的提交结果
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

func (b *ContestBuilder) Description(description string) *ContestBuilder {
	b.item.Description = description
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

func (b *ContestBuilder) UpdateTime(updateTime time.Time) *ContestBuilder {
	b.item.UpdateTime = updateTime
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

func (b *ContestBuilder) Problems(problems []*ContestProblem) *ContestBuilder {
	b.item.Problems = problems
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

func (b *ContestBuilder) Members(members []int) *ContestBuilder {
	b.item.Members = members
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

func (b *ContestBuilder) VirtualReplay(virtualReplay *VirtualReplay) *ContestBuilder {
	b.item.VirtualReplay = virtualReplay
	return b
}

func (b *ContestBuilder) VMembers(vMembers []int) *ContestBuilder {
	b.item.VMembers = vMembers
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

func (b *ContestBuilder) SubmitAnytime(submitAnytime bool) *ContestBuilder {
	b.item.SubmitAnytime = submitAnytime
	return b
}

func (b *ContestBuilder) AuthMembers(authMembers []int) *ContestBuilder {
	b.item.AuthMembers = authMembers
	return b
}

func (b *ContestBuilder) Authors(authors []int) *ContestBuilder {
	b.item.Authors = authors
	return b
}

func (b *ContestBuilder) Volunteers(volunteers []int) *ContestBuilder {
	b.item.Volunteers = volunteers
	return b
}

func (b *ContestBuilder) Build() *Contest {
	return b.item
}
