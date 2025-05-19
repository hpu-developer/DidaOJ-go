package foundationstatus

import "time"

type JudgerConfig struct {
	Key  string `yaml:"key"`  // 评测器标识
	Name string `yaml:"name"` // 评测器名称
}

type JudgerStatus struct {
	Name       string    `json:"name"`
	CpuUsage   float64   `json:"cpu_usage"` // CPU 使用率
	MemUsage   uint64    `json:"mem_usage"`
	MemTotal   uint64    `json:"mem_total"`
	AvgMessage string    `json:"avg_message"` // 平均负载信息
	UpdateTime time.Time `json:"update_time"`
}

type JudgerStatusBuilder struct {
	item *JudgerStatus
}

func NewJudgerStatusBuilder() *JudgerStatusBuilder {
	return &JudgerStatusBuilder{
		item: &JudgerStatus{},
	}
}

func (b *JudgerStatusBuilder) Name(name string) *JudgerStatusBuilder {
	b.item.Name = name
	return b
}

func (b *JudgerStatusBuilder) UpdateTime(updateTime time.Time) *JudgerStatusBuilder {
	b.item.UpdateTime = updateTime
	return b
}

func (b *JudgerStatusBuilder) CpuUsage(cpuUsage float64) *JudgerStatusBuilder {
	b.item.CpuUsage = cpuUsage
	return b
}

func (b *JudgerStatusBuilder) MemUsage(memUsage uint64) *JudgerStatusBuilder {
	b.item.MemUsage = memUsage
	return b
}

func (b *JudgerStatusBuilder) MemTotal(memTotal uint64) *JudgerStatusBuilder {
	b.item.MemTotal = memTotal
	return b
}

func (b *JudgerStatusBuilder) AvgMessage(avgMessage string) *JudgerStatusBuilder {
	b.item.AvgMessage = avgMessage
	return b
}

func (b *JudgerStatusBuilder) Build() *JudgerStatus {
	return b.item
}
