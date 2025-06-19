package request

type DiscussEdit struct {
	Id        int     `json:"id,omitempty"`              // Discuss ID, optional for creation
	Title     string  `json:"title" validate:"required"` // 标题
	Content   string  `json:"content"`
	ProblemId *string `json:"problem_id,omitempty"` // 关联的ProblemId
}
