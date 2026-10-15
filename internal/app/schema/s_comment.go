package schema

type CommentInput struct {
	PostID   OptionalString `json:"postId"`
	ParentID OptionalString `json:"parentId"`
	Content  OptionalString `json:"content"`
}

type CommentStatusInput struct {
	Status string `json:"status"`
}
