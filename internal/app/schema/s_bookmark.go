package schema

import "time"

type Bookmark struct {
	UserID    string    `json:"userId"`
	PostID    string    `json:"postId"`
	CreatedAt time.Time `json:"createdAt"`
}
