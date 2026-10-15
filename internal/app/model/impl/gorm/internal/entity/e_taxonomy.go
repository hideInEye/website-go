package entity

import "time"

type Category struct {
	ID          string    `gorm:"primary_key;column:id" json:"id"`
	Name        string    `gorm:"column:name" json:"name"`
	Slug        string    `gorm:"column:slug" json:"slug"`
	Description *string   `gorm:"column:description" json:"description"`
	CreatedAt   time.Time `gorm:"column:createdAt" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Category) TableName() string {
	return "categories"
}

type Tag struct {
	ID        string    `gorm:"primary_key;column:id" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	Slug      string    `gorm:"column:slug" json:"slug"`
	CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Tag) TableName() string {
	return "tags"
}
