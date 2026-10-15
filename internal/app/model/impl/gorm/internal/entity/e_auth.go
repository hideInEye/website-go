package entity

import "time"

type Session struct {
	ID        string     `gorm:"primary_key;column:id" json:"id"`
	TokenHash string     `gorm:"column:tokenHash" json:"-"`
	UserID    string     `gorm:"column:userId" json:"userId"`
	UserAgent *string    `gorm:"column:userAgent" json:"userAgent"`
	IPAddress *string    `gorm:"column:ipAddress" json:"ipAddress"`
	ExpiresAt time.Time  `gorm:"column:expiresAt" json:"expiresAt"`
	RevokedAt *time.Time `gorm:"column:revokedAt" json:"revokedAt"`
	CreatedAt time.Time  `gorm:"column:createdAt" json:"createdAt"`
}

func (Session) TableName() string {
	return "sessions"
}

type VerificationToken struct {
	ID        string     `gorm:"primary_key;column:id"`
	TokenHash string     `gorm:"column:tokenHash"`
	UserID    string     `gorm:"column:userId"`
	ExpiresAt time.Time  `gorm:"column:expiresAt"`
	UsedAt    *time.Time `gorm:"column:usedAt"`
	CreatedAt time.Time  `gorm:"column:createdAt"`
}

func (VerificationToken) TableName() string {
	return "email_verification_tokens"
}

type ResetToken VerificationToken

func (ResetToken) TableName() string {
	return "password_reset_tokens"
}
