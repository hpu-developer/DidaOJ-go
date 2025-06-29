package foundationview

import (
	foundationmodel "foundation/foundation-model"
	"time"
)

type CollectionList struct {
	Id               int        `json:"id" gorm:"primaryKey;autoIncrement"`
	Title            string     `json:"title" gorm:"type:varchar(30)"`
	StartTime        *time.Time `json:"start_time"`
	EndTime          *time.Time `json:"end_time"`
	Inserter         int        `json:"inserter"`
	InserterUsername string     `json:"inserter_username"`
	InserterNickname string     `json:"inserter_nickname"`
}

type CollectionDetail struct {
	foundationmodel.Collection

	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	ModifierUsername string `json:"modifier_username"`
	ModifierNickname string `json:"modifier_nickname"`
}
