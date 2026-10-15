package entity

import "time"

type User struct {
	ID              string     `gorm:"primary_key;column:id" json:"id"`
	Email           string     `gorm:"column:email" json:"email"`
	Username        string     `gorm:"column:username" json:"username"`
	DisplayName     string     `gorm:"column:displayName" json:"displayName"`
	AvatarURL       *string    `gorm:"column:avatarUrl" json:"avatarUrl"`
	Bio             *string    `gorm:"column:bio" json:"bio"`
	PasswordHash    string     `gorm:"column:passwordHash" json:"-"`
	Role            string     `gorm:"column:role" json:"role"`
	Status          string     `gorm:"column:status" json:"status"`
	EmailVerifiedAt *time.Time `gorm:"column:emailVerifiedAt" json:"emailVerifiedAt"`
	CreatedAt       time.Time  `gorm:"column:createdAt" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updatedAt" json:"updatedAt"`
	DeletedAt       *time.Time `gorm:"column:deletedAt" json:"-"`
}

func (User) TableName() string {
	return "users"
}
