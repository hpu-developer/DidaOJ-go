package foundationmodel

type ProblemTag struct {
	Id   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
}

type ProblemTagBuilder struct {
	item *ProblemTag
}

func NewProblemTagBuilder() *ProblemTagBuilder {
	return &ProblemTagBuilder{item: &ProblemTag{}}
}

func (b *ProblemTagBuilder) Id(id string) *ProblemTagBuilder {
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
