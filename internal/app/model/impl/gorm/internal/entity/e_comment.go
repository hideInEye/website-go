package entity

import "time"

type Comment struct {
	ID        string     `gorm:"primary_key;column:id" json:"id"`
	Content   string     `gorm:"column:content" json:"content"`
	Status    string     `gorm:"column:status" json:"status"`
	PostID    string     `gorm:"column:postId" json:"postId"`
	UserID    string     `gorm:"column:userId" json:"userId"`
	ParentID  *string    `gorm:"column:parentId" json:"parentId"`
	CreatedAt time.Time  `gorm:"column:createdAt" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updatedAt" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deletedAt" json:"-"`
}

func (Comment) TableName() string {
	return "comments"
}
