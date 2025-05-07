package foundationmodel

type ProblemView struct {
	Id          string `json:"id" bson:"_id"`
	ProblemId   string `json:"problem_id" bson:"problem_id"`   // 题目Id
	Title       string `json:"title" bson:"title"`             // 题目标题
	Description string `json:"description" bson:"description"` // 题目描述
	Author      int    `json:"author" bson:"author"`           // 上传者
}

type ProblemViewBuilder struct {
	item *ProblemView
}

func NewProblemViewBuilder() *ProblemViewBuilder {
	return &ProblemViewBuilder{item: &ProblemView{}}
}

func (b *ProblemViewBuilder) Id(id string) *ProblemViewBuilder {
	b.item.Id = id
	return b
}

func (b *ProblemViewBuilder) ProblemId(problemId string) *ProblemViewBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *ProblemViewBuilder) Title(title string) *ProblemViewBuilder {
	b.item.Title = title
	return b
}

func (b *ProblemViewBuilder) Description(description string) *ProblemViewBuilder {
	b.item.Description = description
	return b
}

func (b *ProblemViewBuilder) Author(author int) *ProblemViewBuilder {
	b.item.Author = author
	return b
}

func (b *ProblemViewBuilder) Build() *ProblemView {
	return b.item
}
