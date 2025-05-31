package request

type CollectionCreate struct {
	Title       string `json:"title" validate:"required"` // 比赛标题
	Description string `json:"description"`
	StartTime   string `json:"start_time" validate:"required"` // 比赛开启时间
	EndTime     string `json:"end_time" validate:"required"`   // 比赛结束时间
}
