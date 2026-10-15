package bll

import (
	"context"

	"go-gin/internal/app/schema"
)

type IHealth interface {
	Live(context.Context) (*schema.HealthLive, error)
	Ready(context.Context) (*schema.HealthReady, error)
}
