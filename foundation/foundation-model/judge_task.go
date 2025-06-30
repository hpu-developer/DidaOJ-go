package foundationmodel

import (
	foundationjudge "foundation/foundation-judge"
)

type JudgeTask struct {
	Id      int                         `json:"id" gorm:"column:id;primaryKey;not null"`           // 任务Id
	TaskId  string                      `json:"task_id" gorm:"column:task_id;primaryKey;not null"` // 任务标识
	Status  foundationjudge.JudgeStatus `json:"status" gorm:"column:status"`                       // 评测状态
	Time    int                         `json:"time,omitempty" gorm:"column:time"`                 // 所用的时间
	Memory  int                         `json:"memory,omitempty" gorm:"column:memory"`             // 所用的内存
	Score   int                         `json:"score,omitempty" gorm:"column:score"`               // 所得分数
	Content string                      `json:"content,omitempty" gorm:"column:content"`           // 输出内容
	Hint    string                      `json:"hint,omitempty" gorm:"column:hint"`                 // 提示
}

// TableName 重写表名
func (JudgeTask) TableName() string {
	return "judge_task"
}

type JudgeTaskBuilder struct {
	item *JudgeTask
}

func NewJudgeTaskBuilder() *JudgeTaskBuilder {
	return &JudgeTaskBuilder{item: &JudgeTask{}}
}

func (b *JudgeTaskBuilder) Id(id int) *JudgeTaskBuilder {
	b.item.Id = id
	return b
}

func (b *JudgeTaskBuilder) TaskId(taskId string) *JudgeTaskBuilder {
	b.item.TaskId = taskId
	return b
}

func (b *JudgeTaskBuilder) Status(status foundationjudge.JudgeStatus) *JudgeTaskBuilder {
	b.item.Status = status
	return b
}

func (b *JudgeTaskBuilder) Time(time int) *JudgeTaskBuilder {
	b.item.Time = time
	return b
}

func (b *JudgeTaskBuilder) Memory(memory int) *JudgeTaskBuilder {
	b.item.Memory = memory
	return b
}

func (b *JudgeTaskBuilder) Score(score int) *JudgeTaskBuilder {
	b.item.Score = score
	return b
}

func (b *JudgeTaskBuilder) Content(content string) *JudgeTaskBuilder {
	b.item.Content = content
	return b
}

func (b *JudgeTaskBuilder) Hint(hint string) *JudgeTaskBuilder {
	b.item.Hint = hint
	return b
}

func (b *JudgeTaskBuilder) Build() *JudgeTask {
	return b.item
}
