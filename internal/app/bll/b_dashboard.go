package bll

import (
	"context"
	"go-gin/internal/app/schema"
)

type IDashboard interface {
	Dashboard(context.Context, schema.User) (schema.Object, error)
}
