package foundationmodel

import "time"

type BotGame struct {
	Id          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Key         string    `json:"key" gorm:"type:varchar(20)"`
	Title       string    `json:"title" gorm:"type:varchar(30)"`
	Description string    `json:"description" gorm:"type:text"`
	JudgeCode   string    `json:"judge_code" gorm:"type:text"`
	Inserter    int       `json:"inserter" gorm:"type:int"`
	InsertTime  time.Time `json:"insert_time" gorm:"type:timestamp"`
	Modifier    int       `json:"modifier" gorm:"type:int"`
	ModifyTime  time.Time `json:"modify_time" gorm:"type:timestamp"`
}

func (*BotGame) TableName() string {
	return "bot_game"
}

type BotGameBuilder struct {
	item *BotGame
}

func NewBotGameBuilder() *BotGameBuilder {
	return &BotGameBuilder{item: &BotGame{}}
}

func (b *BotGameBuilder) Id(id int) *BotGameBuilder {
	b.item.Id = id
	return b
}

func (b *BotGameBuilder) Key(key string) *BotGameBuilder {
	b.item.Key = key
	return b
}

func (b *BotGameBuilder) Title(title string) *BotGameBuilder {
	b.item.Title = title
	return b
}

func (b *BotGameBuilder) Description(description string) *BotGameBuilder {
	b.item.Description = description
	return b
}

func (b *BotGameBuilder) JudgeCode(judgeCode string) *BotGameBuilder {
	b.item.JudgeCode = judgeCode
	return b
}

func (b *BotGameBuilder) Inserter(inserter int) *BotGameBuilder {
	b.item.Inserter = inserter
	return b
}

func (b *BotGameBuilder) InsertTime(insertTime time.Time) *BotGameBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *BotGameBuilder) Modifier(modifier int) *BotGameBuilder {
	b.item.Modifier = modifier
	return b
}

func (b *BotGameBuilder) ModifyTime(modifyTime time.Time) *BotGameBuilder {
	b.item.ModifyTime = modifyTime
	return b
}

func (b *BotGameBuilder) Build() *BotGame {
	return b.item
}
