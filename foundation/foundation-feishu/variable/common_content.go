package variable

type CommonContent struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
}

type CommonContentBuilder struct {
	v CommonContent
}

func NewCommonContentBuilder() *CommonContentBuilder {
	return &CommonContentBuilder{
		v: CommonContent{},
	}
}

func (b *CommonContentBuilder) Title(title string) *CommonContentBuilder {
	b.v.Title = &title
	return b
}

func (b *CommonContentBuilder) Content(content string) *CommonContentBuilder {
	b.v.Content = &content
	return b
}

func (b *CommonContentBuilder) Build() *CommonContent {
	return &b.v
}
