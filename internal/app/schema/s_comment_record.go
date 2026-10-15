package schema

import "time"

type Comment struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Status    string     `json:"status"`
	PostID    string     `json:"postId"`
	UserID    string     `json:"userId"`
	ParentID  *string    `json:"parentId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"-"`
}
