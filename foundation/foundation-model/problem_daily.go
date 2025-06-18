package foundationmodel

import "time"

type ProblemDaily struct {
	Id        string `json:"id" bson:"_id"`
	ProblemId string `json:"problem_id" bson:"problem_id"`
	Solution  string `json:"solution" bson:"solution"`
	Code      string `json:"code" bson:"code"`

	CreateTime time.Time `json:"create_time" bson:"create_time"`
	UpdateTime time.Time `json:"update_time" bson:"update_time"`

	CreatorId       int     `json:"creator_id" bson:"creator_id"`
	CreatorUsername *string `json:"creator_username,omitempty" bson:"creator_username,omitempty"`
	CreatorNickname *string `json:"creator_nickname,omitempty" bson:"creator_nickname,omitempty"`
	UpdaterId       int     `json:"updater_id" bson:"updater_id"`
	UpdaterUsername *string `json:"updater_username,omitempty" bson:"updater_username,omitempty"`
	UpdaterNickname *string `json:"updater_nickname,omitempty" bson:"updater_nickname,omitempty"`

	Title   *string `json:"title,omitempty" bson:"title,omitempty"`
	Tags    []int   `json:"tags,omitempty" bson:"tags,omitempty"`
	Accept  int     `json:"accept,omitempty" bson:"accept,omitempty"`
	Attempt int     `json:"attempt,omitempty" bson:"attempt,omitempty"`
}

type ProblemDailyBuilder struct {
	item *ProblemDaily
}

func NewProblemDailyBuilder() *ProblemDailyBuilder {
	return &ProblemDailyBuilder{item: &ProblemDaily{}}
}

func (b *ProblemDailyBuilder) Id(id string) *ProblemDailyBuilder {
	b.item.Id = id
	return b
}

func (b *ProblemDailyBuilder) ProblemId(problemId string) *ProblemDailyBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *ProblemDailyBuilder) Solution(solution string) *ProblemDailyBuilder {
	b.item.Solution = solution
	return b
}

func (b *ProblemDailyBuilder) Code(code string) *ProblemDailyBuilder {
	b.item.Code = code
	return b
}

func (b *ProblemDailyBuilder) CreateTime(createTime time.Time) *ProblemDailyBuilder {
	b.item.CreateTime = createTime
	return b
}

func (b *ProblemDailyBuilder) UpdateTime(updateTime time.Time) *ProblemDailyBuilder {
	b.item.UpdateTime = updateTime
	return b
}

func (b *ProblemDailyBuilder) CreatorId(creatorId int) *ProblemDailyBuilder {
	b.item.CreatorId = creatorId
	return b
}

func (b *ProblemDailyBuilder) UpdaterId(updaterId int) *ProblemDailyBuilder {
	b.item.UpdaterId = updaterId
	return b
}

func (b *ProblemDailyBuilder) Build() *ProblemDaily {
	return b.item
}
