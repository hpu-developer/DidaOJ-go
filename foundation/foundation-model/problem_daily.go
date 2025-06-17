package foundationmodel

type ProblemDaily struct {
	Id        string `json:"id" bson:"_id"`
	ProblemId string `json:"problem_id" bson:"problem_id"`
	Solution  string `json:"solution" bson:"solution"`
	Code      string `json:"code" bson:"code"`

	Title   *string `json:"title,omitempty" bson:"title,omitempty"`
	Tags    []int   `json:"tags,omitempty" bson:"tags,omitempty"`
	Accept  int     `json:"accept,omitempty" bson:"accept,omitempty"`
	Attempt int     `json:"attempt,omitempty" bson:"attempt,omitempty"`
}

type ProblemDailyBuilder struct {
	item *ProblemDaily
}

func NewProblemDailyBuilder() *ProblemDailyBuilder {
	return &ProblemDailyBuilder{item: &ProblemDaily{}}
}

func (b *ProblemDailyBuilder) Id(id string) *ProblemDailyBuilder {
	b.item.Id = id
	return b
}

func (b *ProblemDailyBuilder) ProblemId(problemId string) *ProblemDailyBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *ProblemDailyBuilder) Solution(solution string) *ProblemDailyBuilder {
	b.item.Solution = solution
	return b
}

func (b *ProblemDailyBuilder) Code(code string) *ProblemDailyBuilder {
	b.item.Code = code
	return b
}

func (b *ProblemDailyBuilder) Build() *ProblemDaily {
	return b.item
}
