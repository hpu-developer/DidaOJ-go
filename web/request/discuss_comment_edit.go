package request

type DiscussCommentEdit struct {
	Id        int    `json:"id,omitempty"`
	DiscussId int    `json:"discuss_id,omitempty"`
	CommentId int    `json:"comment_id,omitempty"`
	Content   string `json:"content" validate:"required"`
}
