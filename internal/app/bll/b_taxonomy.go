package bll

import (
	"context"
	"go-gin/internal/app/schema"
)

type ITaxonomy interface {
	Categories(context.Context) ([]schema.CategoryWithCount, error)
	Tags(context.Context) ([]schema.TagWithCount, error)
	CreateCategory(context.Context, schema.User, schema.TaxonomyInput) (*schema.Category, error)
	CreateTag(context.Context, schema.User, schema.TaxonomyInput) (*schema.Tag, error)
	UpdateCategory(context.Context, schema.User, string, schema.TaxonomyInput) (*schema.Category, error)
	UpdateTag(context.Context, schema.User, string, schema.TaxonomyInput) (*schema.Tag, error)
	DeleteCategory(context.Context, schema.User, string) (schema.Object, error)
	DeleteTag(context.Context, schema.User, string) (schema.Object, error)
}
