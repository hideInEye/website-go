package schema

import "time"

type Media struct {
	ID           string    `json:"id"`
	ObjectKey    string    `json:"objectKey"`
	OriginalName string    `json:"originalName"`
	MimeType     string    `json:"mimeType"`
	Size         int64     `json:"size"`
	Width        *int      `json:"width"`
	Height       *int      `json:"height"`
	AltText      *string   `json:"altText"`
	Status       string    `json:"status"`
	UploaderID   string    `json:"uploaderId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
