package foundationmodel

import "time"

type Judger struct {
	Key        string    `json:"key" gorm:"primaryKey;column:key"` // 主键
	Name       string    `json:"name" gorm:"column:name"`
	MaxJob     int       `json:"max_job" gorm:"column:max_job"`
	CpuUsage   float64   `json:"cpu_usage" gorm:"column:cpu_usage"`
	MemUsage   uint64    `json:"mem_usage" gorm:"column:mem_usage"`
	MemTotal   uint64    `json:"mem_total" gorm:"column:mem_total"`
	AvgMessage string    `json:"avg_message" gorm:"column:avg_message"`
	InsertTime time.Time `json:"insert_time" gorm:"column:insert_time;autoCreateTime"` // 创建时间
	ModifyTime time.Time `json:"modify_time" gorm:"column:modify_time"`
}

func (Judger) TableName() string {
	return "judger"
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

func (b *JudgerBuilder) MaxJob(maxJob int) *JudgerBuilder {
	b.item.MaxJob = maxJob
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

func (b *JudgerBuilder) ModifyTime(modifyTime time.Time) *JudgerBuilder {
	b.item.ModifyTime = modifyTime
	return b
}

func (b *JudgerBuilder) Build() *Judger {
	return b.item
}
