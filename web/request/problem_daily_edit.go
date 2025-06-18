package request

type ProblemDailyEdit struct {
	Id        string `json:"id,omitempty" binding:"omitempty"`
	ProblemId string `json:"problem_id" binding:"required"`
	Solution  string `json:"solution" binding:"required"`
	Code      string `json:"code" binding:"required"`
}
