package foundationmodelmongo

import (
	metautf "meta/meta-utf"
	"time"
)

type Discuss struct {
	Id             int       `json:"id" bson:"_id"`                                              // 数据库索引时的Id
	Title          string    `json:"title" bson:"title"`                                         // 讨论标题
	Content        string    `json:"content,omitempty" bson:"content"`                           // 讨论内容
	AuthorId       int       `json:"author_id" bson:"author_id"`                                 // 作者
	AuthorUsername *string   `json:"author_username,omitempty" bson:"author_username,omitempty"` // 作者用户名，不存档
	AuthorNickname *string   `json:"author_nickname,omitempty" bson:"author_nickname,omitempty"` // 作者昵称，不存档
	InsertTime     time.Time `json:"insert_time" bson:"insert_time"`                             // 创建时间
	ModifyTime     time.Time `json:"modify_time" bson:"modify_time"`                             // 修改时间
	UpdateTime     time.Time `json:"update_time" bson:"update_time"`                             // 更新时间，有回复时会更新
	ViewCount      int       `json:"view_count" bson:"view_count"`                               // 浏览次数

	Banned bool `json:"banned,omitempty" bson:"banned,omitempty"` // 是否被封禁，封禁后无法查看和回复

	// 用于标识帖子的关联
	ProblemId           *string `json:"problem_id,omitempty" bson:"problem_id,omitempty"`                       // 问题Id
	ProblemTitle        *string `json:"problem_title,omitempty" bson:"problem_title,omitempty"`                 // 问题标题，不存档
	ContestId           int     `json:"contest_id,omitempty" bson:"contest_id,omitempty"`                       // 比赛Id
	ContestTitle        *string `json:"contest_title,omitempty" bson:"contest_title,omitempty"`                 // 比赛标题，不存档
	ContestProblemIndex int     `json:"contest_problem_index,omitempty" bson:"contest_problem_index,omitempty"` // 比赛题目序号，不会存档，用于标识题目并且隐藏真实题目
	JudgeId             int     `json:"judge_id,omitempty" bson:"judge_id,omitempty"`                           // 评测Id

	Tags []int `json:"tags,omitempty" bson:"tags,omitempty"` // 讨论标签，用于给帖子分类，用户自定义的标签

	MigrateDidaOJId           int
	MigrateEojBlogId          int
	MigrateEojClarificationId int
}

type DiscussBuilder struct {
	item *Discuss
}

func NewDiscussBuilder() *DiscussBuilder {
	return &DiscussBuilder{item: &Discuss{}}
}

func (b *DiscussBuilder) Id(id int) *DiscussBuilder {
	b.item.Id = id
	return b
}

func (b *DiscussBuilder) Title(title string) *DiscussBuilder {
	b.item.Title = metautf.SanitizeText(title)
	return b
}

func (b *DiscussBuilder) Content(content string) *DiscussBuilder {
	b.item.Content = metautf.SanitizeText(content)
	return b
}

func (b *DiscussBuilder) AuthorId(authorId int) *DiscussBuilder {
	b.item.AuthorId = authorId
	return b
}

func (b *DiscussBuilder) AuthorUsername(authorUsername string) *DiscussBuilder {
	b.item.AuthorUsername = &authorUsername
	return b
}

func (b *DiscussBuilder) AuthorNickname(authorNickname string) *DiscussBuilder {
	b.item.AuthorNickname = &authorNickname
	return b
}

func (b *DiscussBuilder) InsertTime(insertTime time.Time) *DiscussBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *DiscussBuilder) ModifyTime(modifyTime time.Time) *DiscussBuilder {
	b.item.ModifyTime = modifyTime
	return b
}

func (b *DiscussBuilder) UpdateTime(updateTime time.Time) *DiscussBuilder {
	b.item.UpdateTime = updateTime
	return b
}

func (b *DiscussBuilder) ViewCount(viewCount int) *DiscussBuilder {
	b.item.ViewCount = viewCount
	return b
}

func (b *DiscussBuilder) ProblemId(problemId *string) *DiscussBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *DiscussBuilder) ContestId(contestId int) *DiscussBuilder {
	b.item.ContestId = contestId
	return b
}

func (b *DiscussBuilder) Tags(tags []int) *DiscussBuilder {
	b.item.Tags = tags
	return b
}

func (b *DiscussBuilder) Build() *Discuss {
	return b.item
}
