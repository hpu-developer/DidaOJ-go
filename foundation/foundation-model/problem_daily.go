package foundationmodel

import "time"

type ProblemDaily struct {
	Key        string    `json:"key" gorm:"column:key;type:char(10);primaryKey" `
	ProblemId  int       `json:"problem_id" gorm:"column:problem_id;unique;not null" `
	Solution   string    `json:"solution" gorm:"column:solution;type:text;not null"`
	Code       string    `json:"code" gorm:"column:code;type:text;not null" `
	Inserter   int       `json:"inserter" gorm:"column:inserter;not null" `
	Modifier   int       `json:"modifier" gorm:"column:modifier;not null" `
	InsertTime time.Time `json:"insert_time" gorm:"column:insert_time;not null"`
	ModifyTime time.Time `json:"modify_time" gorm:"column:modify_time;not null"`
}

func (*ProblemDaily) TableName() string {
	return "problem_daily"
}

type ProblemDailyBuilder struct {
	item *ProblemDaily
}

func NewProblemDailyBuilder() *ProblemDailyBuilder {
	return &ProblemDailyBuilder{item: &ProblemDaily{}}
}

func (b *ProblemDailyBuilder) Key(key string) *ProblemDailyBuilder {
	b.item.Key = key
	return b
}

func (b *ProblemDailyBuilder) ProblemId(problemId int) *ProblemDailyBuilder {
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

func (b *ProblemDailyBuilder) Inserter(inserter int) *ProblemDailyBuilder {
	b.item.Inserter = inserter
	return b
}

func (b *ProblemDailyBuilder) Modifier(modifier int) *ProblemDailyBuilder {
	b.item.Modifier = modifier
	return b
}

func (b *ProblemDailyBuilder) InsertTime(insertTime time.Time) *ProblemDailyBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *ProblemDailyBuilder) ModifyTime(modifyTime time.Time) *ProblemDailyBuilder {
	b.item.ModifyTime = modifyTime
	return b
}

func (b *ProblemDailyBuilder) Build() *ProblemDaily {
	return b.item
}
