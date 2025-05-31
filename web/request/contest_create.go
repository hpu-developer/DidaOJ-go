package request

import foundationmodel "foundation/foundation-model"

type ContestCreate struct {
	Title        string                                `json:"title" validate:"required"` // 比赛标题
	Descriptions []*foundationmodel.ContestDescription `json:"descriptions"`
	StartTime    string                                `json:"start_time" validate:"required"` // 比赛开启时间
	EndTime      string                                `json:"end_time" validate:"required"`   // 比赛结束时间
}
