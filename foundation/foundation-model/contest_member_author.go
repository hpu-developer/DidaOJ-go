package foundationmodel

type ContestMemberAuthor struct {
	Id     int `gorm:"column:id;primaryKey"`
	UserId int `gorm:"column:user_id;primaryKey"`
}

func (p *ContestMemberAuthor) TableName() string {
	return "contest_member_author"
}

type ContestMemberAuthorBuilder struct {
	item *ContestMemberAuthor
}

func NewContestMemberAuthorBuilder() *ContestMemberAuthorBuilder {
	return &ContestMemberAuthorBuilder{
		item: &ContestMemberAuthor{},
	}
}

func (b *ContestMemberAuthorBuilder) Id(id int) *ContestMemberAuthorBuilder {
	b.item.Id = id
	return b
}

func (b *ContestMemberAuthorBuilder) UserId(userId int) *ContestMemberAuthorBuilder {
	b.item.UserId = userId
	return b
}

func (b *ContestMemberAuthorBuilder) Build() *ContestMemberAuthor {
	return b.item
}
