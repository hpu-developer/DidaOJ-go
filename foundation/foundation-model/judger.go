package foundationmodel

type Judger struct {
	Key  string `json:"key" bson:"_id"`
	Name string `json:"name" bson:"name"` // 评测器名称
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

func (b *JudgerBuilder) Build() *Judger {
	return b.item
}
