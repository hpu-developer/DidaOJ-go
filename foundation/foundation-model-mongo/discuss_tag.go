package foundationmodelmongo

import "time"

type DiscussTag struct {
	Id         int        `json:"id" bson:"_id"`
	Name       string     `json:"name" bson:"name"`
	UpdateTime *time.Time `json:"update_time,omitempty" bson:"update_time,omitempty"` // 更新时间，定义为本身修改或者题目修改时更新
}

type DiscussTagBuilder struct {
	item *DiscussTag
}

func NewDiscussTagBuilder() *DiscussTagBuilder {
	return &DiscussTagBuilder{item: &DiscussTag{}}
}

func (b *DiscussTagBuilder) Id(id int) *DiscussTagBuilder {
	b.item.Id = id
	return b
}

func (b *DiscussTagBuilder) Name(name string) *DiscussTagBuilder {
	b.item.Name = name
	return b
}

func (b *DiscussTagBuilder) Build() *DiscussTag {
	return b.item
}
