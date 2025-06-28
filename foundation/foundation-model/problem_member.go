package foundationmodel

type ProblemMember struct {
	Id     int `json:"id" gorm:"column:id;primaryKey;autoIncrement"`          // 题目Id
	UserId int `json:"user_id" bson:"user_id" gorm:"column:user_id;not null"` // 用户Id
}

func (p *ProblemMember) TableName() string {
	return "problem_member"
}

type ProblemMemberBuilder struct {
	item *ProblemMember
}

func NewProblemMemberBuilder() *ProblemMemberBuilder {
	return &ProblemMemberBuilder{
		item: &ProblemMember{},
	}
}

func (b *ProblemMemberBuilder) Id(id int) *ProblemMemberBuilder {
	b.item.Id = id
	return b
}

func (b *ProblemMemberBuilder) UserId(userId int) *ProblemMemberBuilder {
	b.item.UserId = userId
	return b
}

func (b *ProblemMemberBuilder) Build() *ProblemMember {
	return b.item
}
