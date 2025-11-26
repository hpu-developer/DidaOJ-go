package foundationmodel

type BotGame struct {
	Id          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Key         string `json:"key" gorm:"type:varchar(20)"`
	Title       string `json:"title" gorm:"type:varchar(30)"`
	Description string `json:"description" gorm:"type:text"`
	JudgeCode   string `json:"judge_code" gorm:"type:text"`
}

func (*BotGame) TableName() string {
	return "bot_game"
}
