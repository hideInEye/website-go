package entity

import "time"

type Post struct {
	ID             string     `gorm:"primary_key;column:id" json:"id"`
	LegacyRecordID *string    `gorm:"column:legacyRecordId" json:"legacyRecordId"`
	Slug           string     `gorm:"column:slug" json:"slug"`
	Title          string     `gorm:"column:title" json:"title"`
	Excerpt        string     `gorm:"column:excerpt" json:"excerpt"`
	Content        string     `gorm:"column:content" json:"content"`
	CoverURL       *string    `gorm:"column:coverUrl" json:"coverUrl"`
	Status         string     `gorm:"column:status" json:"status"`
	SEOTitle       *string    `gorm:"column:seoTitle" json:"seoTitle"`
	SEODescription *string    `gorm:"column:seoDescription" json:"seoDescription"`
	AuthorID       string     `gorm:"column:authorId" json:"authorId"`
	CategoryID     string     `gorm:"column:categoryId" json:"categoryId"`
	PublishedAt    *time.Time `gorm:"column:publishedAt" json:"publishedAt"`
	CreatedAt      time.Time  `gorm:"column:createdAt" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"column:updatedAt" json:"updatedAt"`
	DeletedAt      *time.Time `gorm:"column:deletedAt" json:"-"`
}

func (Post) TableName() string {
	return "posts"
}

type PostTag struct {
	PostID string `gorm:"primary_key;column:postId" json:"postId"`
	TagID  string `gorm:"primary_key;column:tagId" json:"tagId"`
}

func (PostTag) TableName() string {
	return "post_tags"
}

type SlugHistory struct {
	ID        string    `gorm:"primary_key;column:id" json:"id"`
	PostID    string    `gorm:"column:postId" json:"postId"`
	Slug      string    `gorm:"column:slug" json:"slug"`
	CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
}

func (SlugHistory) TableName() string {
	return "post_slug_history"
}
