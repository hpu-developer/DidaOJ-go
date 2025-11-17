package foundationview

import (
	"time"
)

type ContestRemoteList struct {
	Title     string    `json:"title" gorm:"type:varchar(30)"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"` // 比赛状态
	Type      string    `json:"type"`   // 比赛类型
	Source    string    `json:"source"` // 比赛来源
	Link      string    `json:"link"`   // 比赛链接
}
