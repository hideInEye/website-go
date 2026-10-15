package model

import (
	"context"
	"time"

	"go-gin/internal/app/schema"
)

type IComment interface {
	PublicRoots(context.Context, string) ([]schema.Comment, error)
	PublicReplies(context.Context, []string) ([]schema.Comment, error)
	ByUser(context.Context, string, string, int) ([]schema.Comment, error)
	Relations(context.Context, []schema.Comment) (map[string]schema.User, map[string]schema.Post, error)
	PublishedPostExists(context.Context, string) (bool, error)
	FindReplyParent(context.Context, string, string) (*schema.Comment, error)
	CountApproved(context.Context, string) (int64, error)
	Create(context.Context, schema.Comment) error
	FindIncludingDeleted(context.Context, string) (*schema.Comment, error)
	SoftDelete(context.Context, string, time.Time) error
	Pending(context.Context) ([]schema.Comment, error)
	SetStatus(context.Context, string, string) (*schema.Comment, error)
}
