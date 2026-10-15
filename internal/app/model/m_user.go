package model

import (
	"context"

	"go-gin/internal/app/schema"
)

// IUser is the persistence boundary for profiles and account status.
type IUser interface {
	FindPublic(context.Context, string) (*schema.User, error)
	UpdateProfile(context.Context, string, map[string]interface{}) (*schema.User, error)
	List(context.Context) ([]schema.User, error)
	SetStatus(context.Context, string, string) (*schema.User, error)
}
