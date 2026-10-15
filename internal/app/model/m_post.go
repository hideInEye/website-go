package model

import (
	"context"
	"go-gin/internal/app/schema"
)

type IPost interface {
	ListPublic(context.Context, schema.PostQuery) ([]schema.Post, int64, error)
	ListAdmin(context.Context, schema.PostQuery) ([]schema.Post, int64, error)
	FindPublicBySlug(context.Context, string) (*schema.Post, bool, error)
	FindByLegacyID(context.Context, string) (*schema.Post, error)
	FindAdminByID(context.Context, string, bool) (*schema.Post, error)
	SlugExists(context.Context, string, string) (bool, error)
	Create(context.Context, schema.Post, []string) error
	SaveSlugHistory(context.Context, string, string) error
	Update(context.Context, string, schema.PostChanges) error
	Remove(context.Context, string) (bool, error)
	PresentPosts(context.Context, []schema.Post, bool, bool) ([]schema.Object, error)
}
