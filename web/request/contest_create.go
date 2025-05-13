package request

import foundationmodel "foundation/foundation-model"

type ContestCreate struct {
	Title        string                                `json:"title" validate:"required"` // 比赛标题
	Descriptions []*foundationmodel.ContestDescription `json:"descriptions"`
	OpenTime     []string                              `json:"open_time" validate:"required"` // 比赛开启时间
}
