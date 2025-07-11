package request

type ProblemDailyEdit struct {
	Id         string `json:"id,omitempty" binding:"omitempty"`
	ProblemKey string `json:"problem_key" binding:"required"`
	Solution   string `json:"solution" binding:"required"`
	Code       string `json:"code" binding:"required"`
}
