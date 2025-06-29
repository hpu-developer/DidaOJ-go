package foundationmodel

import "time"

type Collection struct {
	Id          int        `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string     `json:"title" gorm:"type:varchar(30)"`
	Description *string    `json:"description,omitempty" gorm:"type:text"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Inserter    int        `json:"inserter"`
	InsertTime  time.Time  `json:"insert_time"`
	Modifier    int        `json:"modifier"`
	ModifyTime  time.Time  `json:"modify_time"`
	Private     bool       `json:"private" gorm:"type:tinyint(1)"`
	Password    *string    `json:"password" gorm:"type:varchar(30)"`
}

func (*Collection) TableName() string {
	return "collection"
}

type CollectionBuilder struct {
	item *Collection
}

func NewCollectionBuilder() *CollectionBuilder {
	return &CollectionBuilder{item: &Collection{}}
}

func (b *CollectionBuilder) Id(id int) *CollectionBuilder {
	b.item.Id = id
	return b
}

func (b *CollectionBuilder) Title(title string) *CollectionBuilder {
	b.item.Title = title
	return b
}

func (b *CollectionBuilder) Description(desc string) *CollectionBuilder {
	var descriptionPtr *string
	if desc != "" {
		descriptionPtr = &desc
	}
	b.item.Description = descriptionPtr
	return b
}

func (b *CollectionBuilder) StartTime(t *time.Time) *CollectionBuilder {
	b.item.StartTime = t
	return b
}

func (b *CollectionBuilder) EndTime(t *time.Time) *CollectionBuilder {
	b.item.EndTime = t
	return b
}

func (b *CollectionBuilder) Inserter(uid int) *CollectionBuilder {
	b.item.Inserter = uid
	return b
}

func (b *CollectionBuilder) InsertTime(t time.Time) *CollectionBuilder {
	b.item.InsertTime = t
	return b
}

func (b *CollectionBuilder) Modifier(uid int) *CollectionBuilder {
	b.item.Modifier = uid
	return b
}

func (b *CollectionBuilder) ModifyTime(t time.Time) *CollectionBuilder {
	b.item.ModifyTime = t
	return b
}

func (b *CollectionBuilder) Private(p bool) *CollectionBuilder {
	b.item.Private = p
	return b
}

func (b *CollectionBuilder) Password(pw *string) *CollectionBuilder {
	b.item.Password = pw
	return b
}

func (b *CollectionBuilder) Build() *Collection {
	return b.item
}
