package foundationmodel

type CollectionProblem struct {
	Id string `json:"id,omitempty" bson:"id,omitempty"` // 实际的题目Id，添加的那一刻需要具有对应问题权限

	Title *string `json:"title,omitempty" bson:"title,omitempty"` // 题目标题

	Accept  int `json:"accept,omitempty" bson:"accept,omitempty"`   // 题目通过数（暂不存档，动态计算）
	Attempt int `json:"attempt,omitempty" bson:"attempt,omitempty"` // 题目尝试数（暂不存档，动态计算）
}

type CollectionProblemBuilder struct {
	item *CollectionProblem
}

func NewCollectionProblemBuilder() *CollectionProblemBuilder {
	return &CollectionProblemBuilder{item: &CollectionProblem{}}
}

func (b *CollectionProblemBuilder) Id(problemId string) *CollectionProblemBuilder {
	b.item.Id = problemId
	return b
}

func (b *CollectionProblemBuilder) Build() *CollectionProblem {
	return b.item
}
