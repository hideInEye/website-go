package internal

import (
	"context"
	"time"

	"go-gin/internal/app/bll"
	"go-gin/internal/app/errors"
	"go-gin/internal/app/model"
	"go-gin/internal/app/schema"
)

func NewHealth(store model.IHealth) *Health {
	return &Health{store: store}
}

type Health struct {
	store model.IHealth
}

var _ bll.IHealth = (*Health)(nil)

func (a *Health) Live(ctx context.Context) (*schema.HealthLive, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &schema.HealthLive{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	}, nil
}

func (a *Health) Ready(ctx context.Context) (*schema.HealthReady, error) {
	if err := a.store.Ping(ctx); err != nil {
		return nil, errors.NewHTTP(503, "数据库暂时不可用")
	}
	checks := map[string]schema.HealthStatus{"database": {Status: "up"}}
	return &schema.HealthReady{
		Status:  "ok",
		Info:    checks,
		Error:   map[string]schema.HealthStatus{},
		Details: checks,
	}, nil
}
