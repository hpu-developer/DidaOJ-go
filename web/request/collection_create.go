package request

type CollectionEdit struct {
	Id          int      `json:"id"`                        // 题集Id
	Title       string   `json:"title" validate:"required"` // 题集标题
	Description string   `json:"description"`
	StartTime   string   `json:"start_time" validate:"required"` // 题集开启时间
	EndTime     string   `json:"end_time" validate:"required"`   // 题集结束时间
	Users       []int    `json:"users"`                          // 题集成员列表，逗号分隔的用户Id列表
	Problems    []string `json:"problems"`                       // 题目列表，逗号分隔的题目Id列表
}
