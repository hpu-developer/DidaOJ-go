package foundationmodel

import (
	metautf "meta/meta-utf"
	"time"
)

type DiscussComment struct {
	Id             int       `json:"id" bson:"_id"`                                              // 数据库索引时的Id
	DiscussId      int       `json:"discuss_id,omitempty" bson:"discuss_id,omitempty"`           // 讨论的Id
	Content        string    `json:"content" bson:"content"`                                     // 讨论内容
	AuthorId       int       `json:"author_id" bson:"author_id"`                                 // 讨论作者Id
	AuthorUsername *string   `json:"author_username,omitempty" bson:"author_username,omitempty"` // 讨论作者用户名
	AuthorNickname *string   `json:"author_nickname,omitempty" bson:"author_nickname,omitempty"` // 讨论作者昵称
	InsertTime     time.Time `json:"insert_time" bson:"insert_time"`                             // 创建时间
	UpdateTime     time.Time `json:"update_time" bson:"update_time"`                             // 更新时间

	MigrateEojBlogId          int
	MigrateDidaOJId           int
	MigrateEojClarificationId int
}

type DiscussCommentViewEdit struct {
	Id        int    `json:"id" bson:"_id"`                                    // 数据库索引时的Id
	DiscussId int    `json:"discuss_id,omitempty" bson:"discuss_id,omitempty"` // 讨论的Id
	Content   string `json:"content" bson:"content"`                           // 讨论内容
	AuthorId  int    `json:"author_id" bson:"author_id"`                       // 讨论作者Id
}

type DiscussCommentBuilder struct {
	item *DiscussComment
}

func NewDiscussCommentBuilder() *DiscussCommentBuilder {
	return &DiscussCommentBuilder{item: &DiscussComment{}}
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
	b.item.Content = metautf.SanitizeText(content)
	return b
}

func (b *DiscussCommentBuilder) AuthorId(authorId int) *DiscussCommentBuilder {
	b.item.AuthorId = authorId
	return b
}

func (b *DiscussCommentBuilder) InsertTime(insertTime time.Time) *DiscussCommentBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *DiscussCommentBuilder) UpdateTime(updateTime time.Time) *DiscussCommentBuilder {
	b.item.UpdateTime = updateTime
	return b
}

func (b *DiscussCommentBuilder) Build() *DiscussComment {
	return b.item
}
