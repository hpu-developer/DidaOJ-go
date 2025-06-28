package foundationmodel

type ContestMemberAuth struct {
	Id     int `gorm:"column:id;primaryKey"`
	UserId int `gorm:"column:user_id;primaryKey"`
}

func (p *ContestMemberAuth) TableName() string {
	return "contest_member_auth"
}

type ContestMemberAuthBuilder struct {
	item *ContestMemberAuth
}

func NewContestMemberAuthBuilder() *ContestMemberAuthBuilder {
	return &ContestMemberAuthBuilder{
		item: &ContestMemberAuth{},
	}
}

func (b *ContestMemberAuthBuilder) Id(id int) *ContestMemberAuthBuilder {
	b.item.Id = id
	return b
}

func (b *ContestMemberAuthBuilder) UserId(userId int) *ContestMemberAuthBuilder {
	b.item.UserId = userId
	return b
}

func (b *ContestMemberAuthBuilder) Build() *ContestMemberAuth {
	return b.item
}
