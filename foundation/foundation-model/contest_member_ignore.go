package foundationmodel

type ContestMemberIgnore struct {
	Id     int `gorm:"column:id;primaryKey"`
	UserId int `gorm:"column:user_id;primaryKey"`
}

func (p *ContestMemberIgnore) TableName() string {
	return "contest_member_ignore"
}

type ContestMemberIgnoreBuilder struct {
	item *ContestMemberIgnore
}

func NewContestMemberIgnoreBuilder() *ContestMemberIgnoreBuilder {
	return &ContestMemberIgnoreBuilder{
		item: &ContestMemberIgnore{},
	}
}

func (b *ContestMemberIgnoreBuilder) Id(id int) *ContestMemberIgnoreBuilder {
	b.item.Id = id
	return b
}

func (b *ContestMemberIgnoreBuilder) UserId(userId int) *ContestMemberIgnoreBuilder {
	b.item.UserId = userId
	return b
}

func (b *ContestMemberIgnoreBuilder) Build() *ContestMemberIgnore {
	return b.item
}
