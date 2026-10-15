package schema

import "time"

type Post struct {
	ID             string     `json:"id"`
	LegacyRecordID *string    `json:"legacyRecordId"`
	Slug           string     `json:"slug"`
	Title          string     `json:"title"`
	Excerpt        string     `json:"excerpt"`
	Content        string     `json:"content"`
	CoverURL       *string    `json:"coverUrl"`
	Status         string     `json:"status"`
	SEOTitle       *string    `json:"seoTitle"`
	SEODescription *string    `json:"seoDescription"`
	AuthorID       string     `json:"authorId"`
	CategoryID     string     `json:"categoryId"`
	PublishedAt    *time.Time `json:"publishedAt"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"-"`
}

type PostTag struct {
	PostID string `json:"postId"`
	TagID  string `json:"tagId"`
}

type SlugHistory struct {
	ID        string    `json:"id"`
	PostID    string    `json:"postId"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"createdAt"`
}
