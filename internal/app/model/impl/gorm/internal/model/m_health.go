package model

import (
	"context"

	contract "go-gin/internal/app/model"
	"go-gin/pkg/gormplus"
)

func NewHealth(db *gormplus.DB) *Health {
	return &Health{db: db}
}

type Health struct {
	db *gormplus.DB
}

var _ contract.IHealth = (*Health)(nil)

func (a *Health) Ping(ctx context.Context) error {
	return a.db.GetDB().DB().PingContext(ctx)
}
