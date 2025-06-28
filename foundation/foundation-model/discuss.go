package foundationmodel

import "time"

type Discuss struct {
	Id         int       `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Title      string    `json:"title,omitempty" gorm:"column:title;size:100;not null"`
	Content    string    `json:"content,omitempty" gorm:"column:content;type:mediumtext;not null"`
	ViewCount  int       `json:"view_count,omitempty" gorm:"column:view_count"`
	Banned     bool      `json:"banned,omitempty" gorm:"column:banned"`
	ProblemId  *int      `json:"problem_id,omitempty" gorm:"column:problem_id"`
	ContestId  *int      `json:"contest_id,omitempty" gorm:"column:contest_id"`
	JudgeId    *int      `json:"judge_id,omitempty" gorm:"column:judge_id"`
	Inserter   int       `json:"inserter" gorm:"column:inserter;not null"`
	InsertTime time.Time `json:"insert_time,omitempty" gorm:"column:insert_time"`
	Modifier   int       `json:"modifier" gorm:"column:modifier;not null"`
	ModifyTime time.Time `json:"modify_time,omitempty" gorm:"column:modify_time"`
	Updater    int       `json:"updater" gorm:"column:updater;not null"`
	UpdateTime time.Time `json:"update_time,omitempty" gorm:"column:update_time"`
}

// TableName 重写表名
func (Discuss) TableName() string {
	return "discuss"
}

type DiscussBuilder struct {
	item *Discuss
}

func NewDiscussBuilder() *DiscussBuilder {
	return &DiscussBuilder{
		item: &Discuss{},
	}
}

func (b *DiscussBuilder) Id(id int) *DiscussBuilder {
	b.item.Id = id
	return b
}

func (b *DiscussBuilder) Title(title string) *DiscussBuilder {
	b.item.Title = title
	return b
}

func (b *DiscussBuilder) Content(content string) *DiscussBuilder {
	b.item.Content = content
	return b
}

func (b *DiscussBuilder) ViewCount(count int) *DiscussBuilder {
	b.item.ViewCount = count
	return b
}

func (b *DiscussBuilder) Banned(banned bool) *DiscussBuilder {
	b.item.Banned = banned
	return b
}

func (b *DiscussBuilder) ProblemId(problemId *int) *DiscussBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *DiscussBuilder) ContestId(contestId *int) *DiscussBuilder {
	b.item.ContestId = contestId
	return b
}

func (b *DiscussBuilder) JudgeId(judgeId *int) *DiscussBuilder {
	b.item.JudgeId = judgeId
	return b
}

func (b *DiscussBuilder) Inserter(inserter int) *DiscussBuilder {
	b.item.Inserter = inserter
	return b
}

func (b *DiscussBuilder) InsertTime(t time.Time) *DiscussBuilder {
	b.item.InsertTime = t
	return b
}

func (b *DiscussBuilder) Modifier(modifier int) *DiscussBuilder {
	b.item.Modifier = modifier
	return b
}

func (b *DiscussBuilder) ModifyTime(t time.Time) *DiscussBuilder {
	b.item.ModifyTime = t
	return b
}

func (b *DiscussBuilder) Updater(updater int) *DiscussBuilder {
	b.item.Updater = updater
	return b
}

func (b *DiscussBuilder) UpdateTime(t time.Time) *DiscussBuilder {
	b.item.UpdateTime = t
	return b
}

func (b *DiscussBuilder) Build() *Discuss {
	return b.item
}
