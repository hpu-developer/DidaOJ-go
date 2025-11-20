package foundationmodel

import foundationjudge "foundation/foundation-judge"

type ProblemLocal struct {
	Id        int                            `json:"id" gorm:"column:id;primaryKey;autoIncrement"`                          // Local题目Id
	ProblemId int                            `json:"problem_id" bson:"problem_id" gorm:"column:problem_id;unique;not null"` // 题目Id
	JudgeMd5  *string                        `json:"judge_md5,omitempty" bson:"judge_md5,omitempty" gorm:"column:judge_md5;size:32"`
	JudgeJob  foundationjudge.JudgeJobConfig `json:"judge_job,omitempty" bson:"judge_job,omitempty" gorm:"column:judge_job"`
}

func (p *ProblemLocal) TableName() string {
	return "problem_local"
}

type ProblemLocalBuilder struct {
	item *ProblemLocal
}

func NewProblemLocalBuilder() *ProblemLocalBuilder {
	return &ProblemLocalBuilder{
		item: &ProblemLocal{},
	}
}

func (b *ProblemLocalBuilder) Id(id int) *ProblemLocalBuilder {
	b.item.Id = id
	return b
}

func (b *ProblemLocalBuilder) ProblemId(problemId int) *ProblemLocalBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *ProblemLocalBuilder) JudgeMd5(judgeMd5 *string) *ProblemLocalBuilder {
	b.item.JudgeMd5 = judgeMd5
	return b
}

func (b *ProblemLocalBuilder) JudgeJob(judgeJob foundationjudge.JudgeJobConfig) *ProblemLocalBuilder {
	b.item.JudgeJob = judgeJob
	return b
}

func (b *ProblemLocalBuilder) Build() *ProblemLocal {
	return b.item
}
