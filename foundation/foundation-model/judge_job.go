package foundationmodel

import (
	foundationjudge "foundation/foundation-judge"
	"time"
)

type JudgeJob struct {
	Id              int                           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ProblemId       int                           `json:"problem_id" gorm:"column:problem_id;not null"`
	ContestId       *int                          `json:"contest_id,omitempty" gorm:"column:contest_id"`
	Language        foundationjudge.JudgeLanguage `json:"language" gorm:"column:language;not null"`
	Code            string                        `json:"code" gorm:"column:code;type:text;not null"`
	CodeLength      int                           `json:"code_length" gorm:"column:code_length;not null"`
	Status          foundationjudge.JudgeStatus   `json:"status" gorm:"column:status;not null"`
	Judger          *string                       `json:"judger,omitempty" gorm:"column:judger"`
	JudgeTime       *time.Time                    `json:"judge_time,omitempty" gorm:"column:judge_time"`
	TaskCurrent     *int                          `json:"task_current,omitempty" gorm:"column:task_current"`
	TaskTotal       *int                          `json:"task_total,omitempty" gorm:"column:task_total"`
	Score           int                           `json:"score,omitempty" gorm:"column:score"`
	Time            int                           `json:"time,omitempty" gorm:"column:time"`
	Memory          int                           `json:"memory,omitempty" gorm:"column:memory"`
	Private         bool                          `json:"private" gorm:"column:private;not null"`
	RemoteJudgeId   *string                       `json:"remote_judge_id,omitempty" gorm:"column:remote_judge_id"`
	RemoteAccountId *string                       `json:"remote_account_id,omitempty" gorm:"column:remote_account_id"`
	Inserter        int                           `json:"inserter" gorm:"column:inserter;not null"`
	InsertTime      time.Time                     `json:"insert_time" gorm:"column:insert_time;not null"`
}

// TableName 重写表名
func (JudgeJob) TableName() string {
	return "judge_job"
}

type JudgeJobBuilder struct {
	item *JudgeJob
}

func NewJudgeJobBuilder() *JudgeJobBuilder {
	return &JudgeJobBuilder{
		item: &JudgeJob{},
	}
}

func (b *JudgeJobBuilder) Id(id int) *JudgeJobBuilder {
	b.item.Id = id
	return b
}

func (b *JudgeJobBuilder) ProblemId(problemId int) *JudgeJobBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *JudgeJobBuilder) ContestId(contestId int) *JudgeJobBuilder {
	if contestId <= 0 {
		b.item.ContestId = nil
	} else {
		b.item.ContestId = &contestId
	}
	return b
}

func (b *JudgeJobBuilder) Language(language foundationjudge.JudgeLanguage) *JudgeJobBuilder {
	b.item.Language = language
	return b
}

func (b *JudgeJobBuilder) Code(code string) *JudgeJobBuilder {
	b.item.Code = code
	return b
}

func (b *JudgeJobBuilder) CodeLength(length int) *JudgeJobBuilder {
	b.item.CodeLength = length
	return b
}

func (b *JudgeJobBuilder) Status(status foundationjudge.JudgeStatus) *JudgeJobBuilder {
	b.item.Status = status
	return b
}

func (b *JudgeJobBuilder) Judger(judger *string) *JudgeJobBuilder {
	b.item.Judger = judger
	return b
}

func (b *JudgeJobBuilder) JudgeTime(t *time.Time) *JudgeJobBuilder {
	b.item.JudgeTime = t
	return b
}

func (b *JudgeJobBuilder) TaskCurrent(current *int) *JudgeJobBuilder {
	b.item.TaskCurrent = current
	return b
}

func (b *JudgeJobBuilder) TaskTotal(total *int) *JudgeJobBuilder {
	b.item.TaskTotal = total
	return b
}

func (b *JudgeJobBuilder) Score(score int) *JudgeJobBuilder {
	b.item.Score = score
	return b
}

func (b *JudgeJobBuilder) Time(timeVal int) *JudgeJobBuilder {
	b.item.Time = timeVal
	return b
}

func (b *JudgeJobBuilder) Memory(memory int) *JudgeJobBuilder {
	b.item.Memory = memory
	return b
}

func (b *JudgeJobBuilder) Private(private bool) *JudgeJobBuilder {
	b.item.Private = private
	return b
}

func (b *JudgeJobBuilder) RemoteJudgeId(id *string) *JudgeJobBuilder {
	b.item.RemoteJudgeId = id
	return b
}

func (b *JudgeJobBuilder) RemoteAccountId(id *string) *JudgeJobBuilder {
	b.item.RemoteAccountId = id
	return b
}

func (b *JudgeJobBuilder) Inserter(inserter int) *JudgeJobBuilder {
	b.item.Inserter = inserter
	return b
}

func (b *JudgeJobBuilder) InsertTime(t time.Time) *JudgeJobBuilder {
	b.item.InsertTime = t
	return b
}

func (b *JudgeJobBuilder) Build() *JudgeJob {
	return b.item
}
