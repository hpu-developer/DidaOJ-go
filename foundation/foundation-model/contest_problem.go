package foundationmodel

type ContestProblem struct {
	ProblemId string  `json:"problem_id,omitempty" bson:"problem_id,omitempty"` // 实际的题目Id，添加的那一刻需要具有对应问题权限
	ViewId    *string `json:"view_id,omitempty" bson:"view_id,omitempty"`       // 题目描述Id，如果不存在则使用默认描述
	Score     int     `json:"score" bson:"score"`                               // 搭配ScoreType使用，定义题目分数，不填写则为0分
	Index     int     `json:"index" bson:"index"`                               // 问题顺序，用于在展示时标识问题，为了方便标识，从1开始

	Title *string `json:"title,omitempty" bson:"title,omitempty"` // 题目标题

	Accept  int `json:"accept,omitempty" bson:"accept,omitempty"`   // 题目通过数（暂不存档，动态计算）
	Attempt int `json:"attempt,omitempty" bson:"attempt,omitempty"` // 题目尝试数（暂不存档，动态计算）
}

type ContestProblemBuilder struct {
	item *ContestProblem
}

func NewContestProblemBuilder() *ContestProblemBuilder {
	return &ContestProblemBuilder{item: &ContestProblem{}}
}

func (b *ContestProblemBuilder) ProblemId(problemId string) *ContestProblemBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *ContestProblemBuilder) ViewId(viewId *string) *ContestProblemBuilder {
	b.item.ViewId = viewId
	return b
}

func (b *ContestProblemBuilder) Score(score int) *ContestProblemBuilder {
	b.item.Score = score
	return b
}

func (b *ContestProblemBuilder) Index(index int) *ContestProblemBuilder {
	b.item.Index = index
	return b
}

func (b *ContestProblemBuilder) Build() *ContestProblem {
	return b.item
}
