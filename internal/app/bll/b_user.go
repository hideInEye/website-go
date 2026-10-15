package bll

import (
	"context"

	"go-gin/internal/app/schema"
)

// IUser owns profile visibility, profile validation and administrator status changes.
type IUser interface {
	PublicProfile(context.Context, string) (schema.Object, error)
	MyProfile(context.Context, schema.User) (schema.Object, error)
	UpdateProfile(context.Context, schema.User, schema.ProfileInput) (schema.Object, error)
	ListUsers(context.Context, schema.User) ([]schema.Object, error)
	SetUserStatus(context.Context, schema.User, string, schema.UserStatusInput) (schema.Object, error)
}
