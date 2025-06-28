package foundationmodel

type ContestLanguage struct {
	Id       int    `gorm:"column:id;primaryKey"`
	Language string `gorm:"column:language;primaryKey"`
}

func (p *ContestLanguage) TableName() string {
	return "contest_language"
}

type ContestLanguageBuilder struct {
	item *ContestLanguage
}

func NewContestLanguageBuilder() *ContestLanguageBuilder {
	return &ContestLanguageBuilder{
		item: &ContestLanguage{},
	}
}

func (b *ContestLanguageBuilder) Id(id int) *ContestLanguageBuilder {
	b.item.Id = id
	return b
}

func (b *ContestLanguageBuilder) Language(language string) *ContestLanguageBuilder {
	b.item.Language = language
	return b
}

func (b *ContestLanguageBuilder) Build() *ContestLanguage {
	return b.item
}
