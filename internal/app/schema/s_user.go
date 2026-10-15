package schema

import "time"

type User struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	Username        string     `json:"username"`
	DisplayName     string     `json:"displayName"`
	AvatarURL       *string    `json:"avatarUrl"`
	Bio             *string    `json:"bio"`
	PasswordHash    string     `json:"-"`
	Role            string     `json:"role"`
	Status          string     `json:"status"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"-"`
}
