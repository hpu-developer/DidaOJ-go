package foundationmodel

type DiscussTag struct {
	Id    int   `gorm:"column:id;primaryKey"`
	TagId int   `gorm:"column:tag_id;primaryKey"`
	Index uint8 `gorm:"column:index"` // 避免用 Go 的关键字 index
}

func (p *DiscussTag) TableName() string {
	return "discuss_tag"
}

type DiscussTagBuilder struct {
	item *DiscussTag
}

func NewDiscussTagBuilder() *DiscussTagBuilder {
	return &DiscussTagBuilder{
		item: &DiscussTag{},
	}
}

func (b *DiscussTagBuilder) Id(id int) *DiscussTagBuilder {
	b.item.Id = id
	return b
}

func (b *DiscussTagBuilder) TagId(tagId int) *DiscussTagBuilder {
	b.item.TagId = tagId
	return b
}

func (b *DiscussTagBuilder) Index(index uint8) *DiscussTagBuilder {
	b.item.Index = index
	return b
}

func (b *DiscussTagBuilder) Build() *DiscussTag {
	return b.item
}
