package foundationmodel

type Tag struct {
	Id   int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"` // 题目Id
	Name string `json:"name,omitempty" bson:"name,omitempty" gorm:"column:name;size:20"`
}

func (p *Tag) TableName() string {
	return "tag"
}

type TagBuilder struct {
	item *Tag
}

func NewTagBuilder() *TagBuilder {
	return &TagBuilder{
		item: &Tag{},
	}
}

func (b *TagBuilder) Id(id int) *TagBuilder {
	b.item.Id = id
	return b
}

func (b *TagBuilder) Name(name string) *TagBuilder {
	b.item.Name = name
	return b
}

func (b *TagBuilder) Build() *Tag {
	return b.item
}
