package foundationmodel

import "time"

type DiscussComment struct {
	Id         int       `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	DiscussId  int       `json:"discuss_id" gorm:"column:discuss_id;not null"`
	Content    string    `json:"content" gorm:"column:content;type:text"` // 允许null，使用指针
	Banned     bool      `json:"banned,omitempty" gorm:"column:banned;type:tinyint(1)"`
	Inserter   int       `json:"inserter" gorm:"column:inserter;not null"`
	InsertTime time.Time `json:"insert_time,omitempty" gorm:"column:insert_time"`
	Modifier   int       `json:"modifier" gorm:"column:modifier;not null"`
	ModifyTime time.Time `json:"modify_time,omitempty" gorm:"column:modify_time"`
}

// TableName 重写表名
func (DiscussComment) TableName() string {
	return "discuss_comment"
}

type DiscussCommentBuilder struct {
	item *DiscussComment
}

func NewDiscussCommentBuilder() *DiscussCommentBuilder {
	return &DiscussCommentBuilder{
		item: &DiscussComment{},
	}
}

func (b *DiscussCommentBuilder) Id(id int) *DiscussCommentBuilder {
	b.item.Id = id
	return b
}

func (b *DiscussCommentBuilder) DiscussId(discussId int) *DiscussCommentBuilder {
	b.item.DiscussId = discussId
	return b
}

func (b *DiscussCommentBuilder) Content(content string) *DiscussCommentBuilder {
	b.item.Content = content
	return b
}

func (b *DiscussCommentBuilder) Banned(banned bool) *DiscussCommentBuilder {
	b.item.Banned = banned
	return b
}

func (b *DiscussCommentBuilder) Inserter(inserter int) *DiscussCommentBuilder {
	b.item.Inserter = inserter
	return b
}

func (b *DiscussCommentBuilder) InsertTime(insertTime time.Time) *DiscussCommentBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *DiscussCommentBuilder) Modifier(modifier int) *DiscussCommentBuilder {
	b.item.Modifier = modifier
	return b
}

func (b *DiscussCommentBuilder) ModifyTime(modifyTime time.Time) *DiscussCommentBuilder {
	b.item.ModifyTime = modifyTime
	return b
}

func (b *DiscussCommentBuilder) Build() *DiscussComment {
	return b.item
}
