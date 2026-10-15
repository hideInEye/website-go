package model

import (
	"context"

	"go-gin/internal/app/schema"
)

type IBookmark interface {
	Exists(context.Context, string, string) (bool, error)
	ListPosts(context.Context, string) ([]schema.Post, error)
	PublishedPostExists(context.Context, string) (bool, error)
	Add(context.Context, string, string) error
	Remove(context.Context, string, string) error
}
