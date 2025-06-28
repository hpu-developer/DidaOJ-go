package foundationmodel

type ProblemMemberAuth struct {
	Id     int `json:"id" gorm:"column:id;primaryKey;autoIncrement"`          // 题目Id
	UserId int `json:"user_id" bson:"user_id" gorm:"column:user_id;not null"` // 用户Id
}

func (p *ProblemMemberAuth) TableName() string {
	return "problem_member_auth"
}

type ProblemMemberAuthBuilder struct {
	item *ProblemMemberAuth
}

func NewProblemMemberAuthBuilder() *ProblemMemberAuthBuilder {
	return &ProblemMemberAuthBuilder{
		item: &ProblemMemberAuth{},
	}
}

func (b *ProblemMemberAuthBuilder) Id(id int) *ProblemMemberAuthBuilder {
	b.item.Id = id
	return b
}

func (b *ProblemMemberAuthBuilder) UserId(userId int) *ProblemMemberAuthBuilder {
	b.item.UserId = userId
	return b
}

func (b *ProblemMemberAuthBuilder) Build() *ProblemMemberAuth {
	return b.item
}
