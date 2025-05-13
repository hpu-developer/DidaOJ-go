package foundationmodel

type ContestDescription struct {
	Title   string `json:"title" bson:"title"`     // 描述标题
	Content string `json:"content" bson:"content"` // 描述内容
	Sort    int    `json:"sort" bson:"sort"`       // 描述排序
}

type ContestDescriptionBuilder struct {
	item *ContestDescription
}

func NewContestDescriptionBuilder() *ContestDescriptionBuilder {
	return &ContestDescriptionBuilder{item: &ContestDescription{}}
}

func (b *ContestDescriptionBuilder) Title(title string) *ContestDescriptionBuilder {
	b.item.Title = title
	return b
}

func (b *ContestDescriptionBuilder) Content(content string) *ContestDescriptionBuilder {
	b.item.Content = content
	return b
}

func (b *ContestDescriptionBuilder) Sort(sort int) *ContestDescriptionBuilder {
	b.item.Sort = sort
	return b
}

func (b *ContestDescriptionBuilder) Build() *ContestDescription {
	return b.item
}
