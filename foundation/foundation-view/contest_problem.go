package foundationview

import foundationenum "foundation/foundation-enum"

type ContestProblemDetail struct {
	Id        int    `json:"id" gorm:"column:id;primaryKey"`
	ProblemId int    `json:"problem_id" gorm:"column:problem_id;primaryKey"`
	Index     uint8  `json:"index" gorm:"column:index;type:tinyint(1) unsigned;"`
	ViewId    *int   `json:"view_id" gorm:"column:view_id;"`
	Score     int    `json:"score" gorm:"column:score;"`
	Title     string `json:"title" gorm:"column:title"`

	Accept  int                                 `json:"accept,omitempty"`
	Attempt int                                 `json:"attempt,omitempty"`
	Status  foundationenum.ProblemAttemptStatus `json:"status,omitempty"`
}

type ContestProblemRank struct {
	ProblemId int   `json:"problem_id" gorm:"column:problem_id;primaryKey"`
	Index     uint8 `json:"index" gorm:"column:index;type:tinyint(1) unsigned;"`
	Score     int   `json:"score" gorm:"column:score;"`
}
