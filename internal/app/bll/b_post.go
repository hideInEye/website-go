package bll

import (
	"context"
	"go-gin/internal/app/schema"
)

type IPost interface {
	ListPublic(context.Context, schema.PostQuery) (schema.Object, error)
	FindPublicBySlug(context.Context, string) (schema.Object, error)
	FindByLegacyID(context.Context, string) (schema.Object, error)
	ListAdmin(context.Context, schema.User, schema.PostQuery) (schema.Object, error)
	FindAdminByID(context.Context, schema.User, string) (schema.Object, error)
	Create(context.Context, schema.User, schema.PostInput) (schema.Object, error)
	Update(context.Context, schema.User, string, schema.PostInput) (schema.Object, error)
	Remove(context.Context, schema.User, string) (schema.Object, error)
}
