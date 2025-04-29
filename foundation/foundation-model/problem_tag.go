package foundationmodel

import "time"

type ProblemTag struct {
	Id         int        `json:"id" bson:"_id"`
	Name       string     `json:"name" bson:"name"`
	UpdateTime *time.Time `json:"update_time,omitempty" bson:"update_time,omitempty"` // 更新时间，定义为本身修改或者题目修改时更新
}

type ProblemTagBuilder struct {
	item *ProblemTag
}

func NewProblemTagBuilder() *ProblemTagBuilder {
	return &ProblemTagBuilder{item: &ProblemTag{}}
}

func (b *ProblemTagBuilder) Id(id int) *ProblemTagBuilder {
	b.item.Id = id
	return b
}

func (b *ProblemTagBuilder) Name(name string) *ProblemTagBuilder {
	b.item.Name = name
	return b
}

func (b *ProblemTagBuilder) Build() *ProblemTag {
	return b.item
}
