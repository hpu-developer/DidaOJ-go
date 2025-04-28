package foundationmodel

import (
	foundationjudge "foundation/foundation-judge"
	"time"
)

type JudgeJob struct {
	Id             int                           `json:"id" bson:"_id"`
	ProblemId      string                        `json:"problem_id" bson:"problem_id"`           // 题目ID
	Author         string                        `json:"author" bson:"author"`                   // 提交者UserId
	ApproveTime    time.Time                     `json:"approve_time" bson:"approve_time"`       //申请时间
	Language       foundationjudge.JudgeLanguage `json:"language" bson:"language"`               // 代码语言
	Code           string                        `json:"code" bson:"code"`                       // 所评测代码
	CodeLength     int                           `json:"code_length" bson:"code_length"`         // 代码长度
	Status         foundationjudge.JudgeStatus   `json:"status" bson:"status"`                   // 评测状态
	JudgeTime      time.Time                     `json:"judge_time" bson:"judge_time"`           // 评测的开始时间
	TaskCurrent    int                           `json:"task_current" bson:"task_current"`       // 评测完成子任务数量
	TaskTotal      int                           `json:"task_total" bson:"task_total"`           // 评测子任务总数量
	Score          int                           `json:"score" bson:"score"`                     // 所得分数
	Judger         string                        `json:"judger" bson:"judger"`                   // 评测机
	Time           int                           `json:"time" bson:"time"`                       // 所用的时间
	Memory         int                           `json:"memory" bson:"memory"`                   // 所用的内存
	CompileMessage string                        `json:"compile_message" bson:"compile_message"` // 编译信息
}

type JudgeJobBuilder struct {
	item *JudgeJob
}

func NewJudgeJobBuilder() *JudgeJobBuilder {
	return &JudgeJobBuilder{item: &JudgeJob{}}
}

func (b *JudgeJobBuilder) Id(id int) *JudgeJobBuilder {
	b.item.Id = id
	return b
}

func (b *JudgeJobBuilder) ProblemId(problemId string) *JudgeJobBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *JudgeJobBuilder) Author(author string) *JudgeJobBuilder {
	b.item.Author = author
	return b
}

func (b *JudgeJobBuilder) ApproveTime(approveTime time.Time) *JudgeJobBuilder {
	b.item.ApproveTime = approveTime
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

func (b *JudgeJobBuilder) CodeLength(codeLength int) *JudgeJobBuilder {
	b.item.CodeLength = codeLength
	return b
}

func (b *JudgeJobBuilder) Status(status foundationjudge.JudgeStatus) *JudgeJobBuilder {
	b.item.Status = status
	return b
}

func (b *JudgeJobBuilder) JudgeTime(judgeTime time.Time) *JudgeJobBuilder {
	b.item.JudgeTime = judgeTime
	return b
}

func (b *JudgeJobBuilder) JudgeTaskComplete(judgeTaskComplete int) *JudgeJobBuilder {
	b.item.TaskCurrent = judgeTaskComplete
	return b
}

func (b *JudgeJobBuilder) JudgeTaskTotal(judgeTaskTotal int) *JudgeJobBuilder {
	b.item.TaskTotal = judgeTaskTotal
	return b
}

func (b *JudgeJobBuilder) Score(score int) *JudgeJobBuilder {
	b.item.Score = score
	return b
}

func (b *JudgeJobBuilder) Judger(judger string) *JudgeJobBuilder {
	b.item.Judger = judger
	return b
}

func (b *JudgeJobBuilder) Time(time int) *JudgeJobBuilder {
	b.item.Time = time
	return b
}

func (b *JudgeJobBuilder) Memory(memory int) *JudgeJobBuilder {
	b.item.Memory = memory
	return b
}

func (b *JudgeJobBuilder) Build() *JudgeJob {
	return b.item
}
