package bll

import (
	"context"

	"go-gin/internal/app/schema"
)

type IBookmark interface {
	HasBookmark(context.Context, schema.User, string) (schema.Object, error)
	ListBookmarks(context.Context, schema.User) ([]schema.Object, error)
	AddBookmark(context.Context, schema.User, string) (schema.Object, error)
	RemoveBookmark(context.Context, schema.User, string) (schema.Object, error)
}
