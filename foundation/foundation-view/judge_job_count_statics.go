package foundationview

import (
	"time"
)

type JudgeJobCountStatics struct {
	Accept  int       `json:"accept" bson:"accept"`   // 接受的数量
	Attempt int       `json:"attempt" bson:"attempt"` // 尝试的数量
	Date    time.Time `json:"date" bson:"date"`
}
