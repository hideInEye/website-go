package schema

import "unicode/utf8"

type MediaPresignInput struct {
	FileName *string `json:"fileName"`
	MimeType string  `json:"mimeType"`
	Size     int64   `json:"size"`
}

func (input MediaPresignInput) Valid() bool {
	if input.FileName == nil || utf8.RuneCountInString(*input.FileName) > 255 || input.Size < 1 || input.Size > 10*1024*1024 {
		return false
	}
	switch input.MimeType {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return true
	default:
		return false
	}
}

type MediaCompleteInput struct {
	Width   *int    `json:"width"`
	Height  *int    `json:"height"`
	AltText *string `json:"altText"`
}

func (input MediaCompleteInput) Valid() bool {
	return (input.Width == nil || *input.Width >= 1) &&
		(input.Height == nil || *input.Height >= 1) &&
		(input.AltText == nil || utf8.RuneCountInString(*input.AltText) <= 500)
}

type MediaPresignResult struct {
	ID        string            `json:"id"`
	UploadURL string            `json:"uploadUrl"`
	ObjectKey string            `json:"objectKey"`
	ExpiresIn int               `json:"expiresIn"`
	Headers   map[string]string `json:"headers"`
}

type MediaCompleteResult struct {
	Media
	URL string `json:"url"`
}
