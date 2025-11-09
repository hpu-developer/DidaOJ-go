package request

type DiscussEdit struct {
	Id         int    `json:"id,omitempty"`              // Discuss ID, optional for creation
	Title      string `json:"title" validate:"required"` // 标题
	Content    string `json:"content" validate:"required"`
	ContestId  int    `json:"contest_id,omitempty"`  // 关联的ContestId
	ProblemKey string `json:"problem_key,omitempty"` // 关联的ProblemId
}
