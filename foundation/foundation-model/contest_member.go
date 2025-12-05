package foundationmodel

type ContestMember struct {
	Id          int    `gorm:"column:id;primaryKey"`
	UserId      int    `gorm:"column:user_id;primaryKey"`
	ContestName string `gorm:"column:contest_name"`
}

func (p *ContestMember) TableName() string {
	return "contest_member"
}

type ContestMemberBuilder struct {
	item *ContestMember
}

func NewContestMemberBuilder() *ContestMemberBuilder {
	return &ContestMemberBuilder{
		item: &ContestMember{},
	}
}

func (b *ContestMemberBuilder) Id(id int) *ContestMemberBuilder {
	b.item.Id = id
	return b
}

func (b *ContestMemberBuilder) UserId(userId int) *ContestMemberBuilder {
	b.item.UserId = userId
	return b
}

func (b *ContestMemberBuilder) ContestName(contestName string) *ContestMemberBuilder {
	b.item.ContestName = contestName
	return b
}

func (b *ContestMemberBuilder) Build() *ContestMember {
	return b.item
}
