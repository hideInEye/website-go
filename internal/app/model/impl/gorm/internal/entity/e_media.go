package entity

import "time"

type Media struct {
	ID           string    `gorm:"primary_key;column:id" json:"id"`
	ObjectKey    string    `gorm:"column:objectKey" json:"objectKey"`
	OriginalName string    `gorm:"column:originalName" json:"originalName"`
	MimeType     string    `gorm:"column:mimeType" json:"mimeType"`
	Size         int64     `gorm:"column:size" json:"size"`
	Width        *int      `gorm:"column:width" json:"width"`
	Height       *int      `gorm:"column:height" json:"height"`
	AltText      *string   `gorm:"column:altText" json:"altText"`
	Status       string    `gorm:"column:status" json:"status"`
	UploaderID   string    `gorm:"column:uploaderId" json:"uploaderId"`
	CreatedAt    time.Time `gorm:"column:createdAt" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Media) TableName() string {
	return "media"
}
