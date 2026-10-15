package schema

import "time"

type Session struct {
	ID        string     `json:"id"`
	TokenHash string     `json:"-"`
	UserID    string     `json:"userId"`
	UserAgent *string    `json:"userAgent"`
	IPAddress *string    `json:"ipAddress"`
	ExpiresAt time.Time  `json:"expiresAt"`
	RevokedAt *time.Time `json:"revokedAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

type VerificationToken struct {
	ID        string
	TokenHash string
	UserID    string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type ResetToken VerificationToken
