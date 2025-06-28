package foundationmodel

type CollectionProblem struct {
	Id        int `gorm:"column:id;primaryKey"`
	ProblemId int `gorm:"column:problem_id;primaryKey"`
	Index     int `gorm:"column:index"`
}

func (p *CollectionProblem) TableName() string {
	return "collection_problem"
}

type CollectionProblemBuilder struct {
	item *CollectionProblem
}

func NewCollectionProblemBuilder() *CollectionProblemBuilder {
	return &CollectionProblemBuilder{
		item: &CollectionProblem{},
	}
}

func (b *CollectionProblemBuilder) Id(id int) *CollectionProblemBuilder {
	b.item.Id = id
	return b
}

func (b *CollectionProblemBuilder) ProblemId(problemId int) *CollectionProblemBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *CollectionProblemBuilder) Index(index int) *CollectionProblemBuilder {
	b.item.Index = index
	return b
}

func (b *CollectionProblemBuilder) Build() *CollectionProblem {
	return b.item
}
