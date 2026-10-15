package model

import (
	"context"

	"go-gin/internal/app/schema"
)

type IFeed interface {
	ListPosts(context.Context, int) ([]schema.FeedPost, error)
	ListCategorySlugs(context.Context) ([]string, error)
	ListTagSlugs(context.Context) ([]string, error)
	FindLegacySlug(context.Context, string) (string, error)
}
