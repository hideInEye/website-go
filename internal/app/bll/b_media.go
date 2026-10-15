package bll

import (
	"context"

	"go-gin/internal/app/schema"
)

type IMedia interface {
	Presign(context.Context, string, schema.MediaPresignInput) (*schema.MediaPresignResult, error)
	Complete(context.Context, string, string, schema.MediaCompleteInput) (*schema.MediaCompleteResult, error)
}
