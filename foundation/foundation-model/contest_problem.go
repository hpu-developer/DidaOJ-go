package foundationmodel

type ContestProblem struct {
	Id        int   `gorm:"column:id;primaryKey"`
	ProblemId int   `gorm:"column:problem_id;primaryKey"`
	Index     uint8 `gorm:"column:index;type:tinyint(1) unsigned;"`
	ViewId    *int  `gorm:"column:view_id;"`
	Score     int   `gorm:"column:score;"`
}

func (p *ContestProblem) TableName() string {
	return "contest_problem"
}

type ContestProblemBuilder struct {
	item *ContestProblem
}

func NewContestProblemBuilder() *ContestProblemBuilder {
	return &ContestProblemBuilder{
		item: &ContestProblem{},
	}
}

func (b *ContestProblemBuilder) Id(id int) *ContestProblemBuilder {
	b.item.Id = id
	return b
}

func (b *ContestProblemBuilder) ProblemId(problemId int) *ContestProblemBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *ContestProblemBuilder) Index(index uint8) *ContestProblemBuilder {
	b.item.Index = index
	return b
}

func (b *ContestProblemBuilder) ViewId(viewId *int) *ContestProblemBuilder {
	b.item.ViewId = viewId
	return b
}

func (b *ContestProblemBuilder) Score(score int) *ContestProblemBuilder {
	b.item.Score = score
	return b
}

func (b *ContestProblemBuilder) Build() *ContestProblem {
	return b.item
}
