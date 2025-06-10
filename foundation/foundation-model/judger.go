package foundationmodel

import "time"

type Judger struct {
	Key        string    `json:"key" bson:"_id"`
	Name       string    `json:"name" bson:"name"`               // 评测器名称
	CpuUsage   float64   `json:"cpu_usage" bson:"cpu_usage"`     // CPU 使用率
	MemUsage   uint64    `json:"mem_usage" bson:"mem_usage"`     // 内存使用量
	MemTotal   uint64    `json:"mem_total" bson:"mem_total"`     // 内存总量
	AvgMessage string    `json:"avg_message" bson:"avg_message"` // 平均负载信息
	UpdateTime time.Time `json:"update_time" bson:"update_time"` // 更新时间
}

type JudgerBuilder struct {
	item *Judger
}

func NewJudgerBuilder() *JudgerBuilder {
	return &JudgerBuilder{
		item: &Judger{},
	}
}

func (b *JudgerBuilder) Key(key string) *JudgerBuilder {
	b.item.Key = key
	return b
}

func (b *JudgerBuilder) Name(name string) *JudgerBuilder {
	b.item.Name = name
	return b
}

func (b *JudgerBuilder) CpuUsage(cpuUsage float64) *JudgerBuilder {
	b.item.CpuUsage = cpuUsage
	return b
}

func (b *JudgerBuilder) MemUsage(memUsage uint64) *JudgerBuilder {
	b.item.MemUsage = memUsage
	return b
}

func (b *JudgerBuilder) MemTotal(memTotal uint64) *JudgerBuilder {
	b.item.MemTotal = memTotal
	return b
}

func (b *JudgerBuilder) AvgMessage(avgMessage string) *JudgerBuilder {
	b.item.AvgMessage = avgMessage
	return b
}

func (b *JudgerBuilder) UpdateTime(updateTime time.Time) *JudgerBuilder {
	b.item.UpdateTime = updateTime
	return b
}

func (b *JudgerBuilder) Build() *Judger {
	return b.item
}
