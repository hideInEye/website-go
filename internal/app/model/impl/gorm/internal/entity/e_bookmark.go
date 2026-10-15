package entity

import "time"

type Bookmark struct {
	UserID    string    `gorm:"primary_key;column:userId" json:"userId"`
	PostID    string    `gorm:"primary_key;column:postId" json:"postId"`
	CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
}

func (Bookmark) TableName() string {
	return "bookmarks"
}
