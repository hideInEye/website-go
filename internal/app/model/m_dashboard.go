package model

import (
	"context"
	"go-gin/internal/app/schema"
)

type IDashboard interface {
	Counts(context.Context) (schema.Object, error)
}
