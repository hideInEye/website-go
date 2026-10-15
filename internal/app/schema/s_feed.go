package schema

import "time"

type FeedPost struct {
	Slug        string
	Title       string
	Excerpt     string
	PublishedAt *time.Time
	CreatedAt   time.Time
}
