package foundationview

import "time"

type WebStatus struct {
	Name       string    `json:"name"`
	CpuUsage   float64   `json:"cpu_usage"` // CPU 使用率
	MemUsage   uint64    `json:"mem_usage"`
	MemTotal   uint64    `json:"mem_total"`
	AvgMessage string    `json:"avg_message"` // 平均负载信息
	UpdateTime time.Time `json:"update_time"`
}

type WebStatusBuilder struct {
	item *WebStatus
}

func NewWebStatusBuilder() *WebStatusBuilder {
	return &WebStatusBuilder{
		item: &WebStatus{},
	}
}

func (b *WebStatusBuilder) Name(name string) *WebStatusBuilder {
	b.item.Name = name
	return b
}

func (b *WebStatusBuilder) UpdateTime(updateTime time.Time) *WebStatusBuilder {
	b.item.UpdateTime = updateTime
	return b
}

func (b *WebStatusBuilder) CpuUsage(cpuUsage float64) *WebStatusBuilder {
	b.item.CpuUsage = cpuUsage
	return b
}

func (b *WebStatusBuilder) MemUsage(memUsage uint64) *WebStatusBuilder {
	b.item.MemUsage = memUsage
	return b
}

func (b *WebStatusBuilder) MemTotal(memTotal uint64) *WebStatusBuilder {
	b.item.MemTotal = memTotal
	return b
}

func (b *WebStatusBuilder) AvgMessage(avgMessage string) *WebStatusBuilder {
	b.item.AvgMessage = avgMessage
	return b
}

func (b *WebStatusBuilder) Build() *WebStatus {
	return b.item
}
