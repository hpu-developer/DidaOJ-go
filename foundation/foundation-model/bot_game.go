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
	ModifyTime  time.Time `json:"modify_time" gorm:"type:timestamp"`
}

func (*BotGame) TableName() string {
	return "bot_game"
}
