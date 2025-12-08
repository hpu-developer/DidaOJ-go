package foundationmodel

import (
	foundationbot "foundation/foundation-bot"
	metapostgresql "meta/meta-postgresql"
	"time"
)

type BotReplay struct {
	Id         int                         `json:"id" gorm:"primaryKey;autoIncrement"`
	GameId     int                         `json:"game_id" gorm:"type:int"`
	Info       string                      `json:"info,omitempty" gorm:"type:text"`
	Param      string                      `json:"param,omitempty" gorm:"type:text"`
	Message    string                      `json:"message,omitempty" gorm:"type:text"`
	Bots       metapostgresql.IntArray     `json:"bots,omitempty" gorm:"type:jsonb"`
	Status     foundationbot.BotGameStatus `json:"status" gorm:"type:smallint"`
	Inserter   int                         `json:"inserter" gorm:"type:int"` // 发起人，0代表系统匹配
	InsertTime time.Time                   `json:"insert_time" gorm:"type:timestamp"`
	Judger     string                      `json:"judger,omitempty" gorm:"column:judger;not null;type:varchar(10)"`
	JudgeTime  time.Time                   `json:"judge_time,omitempty" gorm:"type:timestamp"`
}

func (*BotReplay) TableName() string {
	return "bot_replay"
}
