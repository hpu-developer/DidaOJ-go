package foundationmodelmongo

import (
	foundationjudge "foundation/foundation-judge"
)

type JudgeTask struct {
	TaskId  string                      `json:"task_id" bson:"task_id"`           // 代码长度
	Status  foundationjudge.JudgeStatus `json:"status" bson:"status"`             // 评测状态
	Time    int                         `json:"time" bson:"time,omitempty"`       // 所用的时间
	Memory  int                         `json:"memory" bson:"memory,omitempty"`   // 所用的内存
	Score   int                         `json:"score" bson:"score,omitempty"`     // 所得分数
	Content string                      `json:"content" bson:"content,omitempty"` // 输出内容
	WaHint  string                      `json:"wa_hint" bson:"wa_hint,omitempty"` // 错误提示
}

type JudgeTaskBuilder struct {
	item *JudgeTask
}

func NewJudgeTaskBuilder() *JudgeTaskBuilder {
	return &JudgeTaskBuilder{item: &JudgeTask{}}
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

func (b *JudgeTaskBuilder) WaHint(waHint string) *JudgeTaskBuilder {
	b.item.WaHint = waHint
	return b
}

func (b *JudgeTaskBuilder) Build() *JudgeTask {
	return b.item
}
