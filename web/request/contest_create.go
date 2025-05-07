package request

type ContestCreate struct {
	Title       string   `json:"title" validate:"required"` // 比赛标题
	Description string   `json:"description"`
	OpenTime    []string `json:"open_time" validate:"required"` // 比赛开启时间
}
