package foundationmodel

type CollectionMember struct {
	Id     int `gorm:"column:id;primaryKey"`
	UserId int `gorm:"column:user_id;primaryKey"`
}

func (p *CollectionMember) TableName() string {
	return "collection_member"
}

type CollectionMemberBuilder struct {
	item *CollectionMember
}

func NewCollectionMemberBuilder() *CollectionMemberBuilder {
	return &CollectionMemberBuilder{
		item: &CollectionMember{},
	}
}

func (b *CollectionMemberBuilder) Id(id int) *CollectionMemberBuilder {
	b.item.Id = id
	return b
}

func (b *CollectionMemberBuilder) UserId(userId int) *CollectionMemberBuilder {
	b.item.UserId = userId
	return b
}

func (b *CollectionMemberBuilder) Build() *CollectionMember {
	return b.item
}
