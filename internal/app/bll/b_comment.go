package bll

import (
	"context"

	"go-gin/internal/app/schema"
)

type IComment interface {
	PublicComments(context.Context, string) ([]schema.Object, error)
	MyComments(context.Context, schema.User) ([]schema.Object, error)
	CreateComment(context.Context, schema.User, schema.CommentInput) (schema.Object, error)
	DeleteComment(context.Context, schema.User, string) (schema.Object, error)
	PendingComments(context.Context, schema.User) ([]schema.Object, error)
	ModerateComment(context.Context, schema.User, string, schema.CommentStatusInput) (schema.Object, error)
}
