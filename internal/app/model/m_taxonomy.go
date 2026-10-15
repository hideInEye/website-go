package model

import (
	"context"
	"go-gin/internal/app/schema"
)

type ITaxonomy interface {
	Categories(context.Context) ([]schema.CategoryWithCount, error)
	Tags(context.Context) ([]schema.TagWithCount, error)
	FindCategory(context.Context, string, bool) (*schema.Category, error)
	FindTag(context.Context, string, bool) (*schema.Tag, error)
	CreateCategory(context.Context, schema.Category) error
	CreateTag(context.Context, schema.Tag) error
	UpdateCategory(context.Context, string, map[string]interface{}) error
	UpdateTag(context.Context, string, map[string]interface{}) error
	CategoryPostCount(context.Context, string) (int64, error)
	DeleteCategory(context.Context, string) error
	DeleteTag(context.Context, string) (bool, error)
}
