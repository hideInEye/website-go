package model

import (
	"context"
	"time"

	"go-gin/internal/app/schema"
)

type IMedia interface {
	Create(context.Context, *schema.Media) error
	FindPending(context.Context, string, string) (*schema.Media, error)
	Complete(context.Context, string, string, schema.MediaCompleteInput, time.Time) (*schema.Media, error)
}
