package request

type DiscussCreate struct {
	Title   string `json:"title" validate:"required"` // 标题
	Content string `json:"content"`
}
