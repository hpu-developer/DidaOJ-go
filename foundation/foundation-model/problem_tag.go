package foundationmodel

type ProblemTag struct {
	Id    int   `gorm:"column:id;primaryKey"`
	TagId int   `gorm:"column:tag_id;primaryKey"`
	Index uint8 `gorm:"column:index"` // 避免用 Go 的关键字 index
}

func (p *ProblemTag) TableName() string {
	return "problem_tag"
}

type ProblemTagBuilder struct {
	item *ProblemTag
}

func NewProblemTagBuilder() *ProblemTagBuilder {
	return &ProblemTagBuilder{
		item: &ProblemTag{},
	}
}

func (b *ProblemTagBuilder) Id(id int) *ProblemTagBuilder {
	b.item.Id = id
	return b
}

func (b *ProblemTagBuilder) TagId(tagId int) *ProblemTagBuilder {
	b.item.TagId = tagId
	return b
}

func (b *ProblemTagBuilder) Index(index uint8) *ProblemTagBuilder {
	b.item.Index = index
	return b
}

func (b *ProblemTagBuilder) Build() *ProblemTag {
	return b.item
}
