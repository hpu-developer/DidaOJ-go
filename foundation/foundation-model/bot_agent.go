package foundationmodel

import (
	foundationjudge "foundation/foundation-judge"
	"time"
)

type BotAgent struct {
	Id         int                           `json:"id" gorm:"primaryKey;autoIncrement"`
	Language   foundationjudge.JudgeLanguage `json:"language" gorm:"column:language;not null"`
	Code       string                        `json:"code" gorm:"type:text"`
	GameId     int                           `json:"game_id" gorm:"type:int"`
	Version    int                           `json:"version" gorm:"type:int"`
	Inserter   int                           `json:"inserter" gorm:"type:int"`
	InsertTime time.Time                     `json:"insert_time" gorm:"type:timestamp"`
	ModifyTime time.Time                     `json:"modify_time" gorm:"type:timestamp"`
}

func (*BotAgent) TableName() string {
	return "bot_agent"
}
