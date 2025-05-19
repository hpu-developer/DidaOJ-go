package foundationstatus

import "time"

type WeberConfig struct {
	Key  string `yaml:"key"`
	Name string `yaml:"name"`
}

type WeberStatus struct {
	Name       string    `json:"name"`
	CpuUsage   float64   `json:"cpu_usage"` // CPU 使用率
	MemUsage   uint64    `json:"mem_usage"`
	MemTotal   uint64    `json:"mem_total"`
	AvgMessage string    `json:"avg_message"` // 平均负载信息
	UpdateTime time.Time `json:"update_time"`
}

type WeberStatusBuilder struct {
	item *WeberStatus
}

func NewWeberStatusBuilder() *WeberStatusBuilder {
	return &WeberStatusBuilder{
		item: &WeberStatus{},
	}
}

func (b *WeberStatusBuilder) Name(name string) *WeberStatusBuilder {
	b.item.Name = name
	return b
}

func (b *WeberStatusBuilder) UpdateTime(updateTime time.Time) *WeberStatusBuilder {
	b.item.UpdateTime = updateTime
	return b
}

func (b *WeberStatusBuilder) CpuUsage(cpuUsage float64) *WeberStatusBuilder {
	b.item.CpuUsage = cpuUsage
	return b
}

func (b *WeberStatusBuilder) MemUsage(memUsage uint64) *WeberStatusBuilder {
	b.item.MemUsage = memUsage
	return b
}

func (b *WeberStatusBuilder) MemTotal(memTotal uint64) *WeberStatusBuilder {
	b.item.MemTotal = memTotal
	return b
}

func (b *WeberStatusBuilder) AvgMessage(avgMessage string) *WeberStatusBuilder {
	b.item.AvgMessage = avgMessage
	return b
}

func (b *WeberStatusBuilder) Build() *WeberStatus {
	return b.item
}
