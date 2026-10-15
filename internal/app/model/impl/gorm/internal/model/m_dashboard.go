package model

import (
	"context"

	"go-gin/internal/app/model/impl/gorm/internal/entity"
	"go-gin/internal/app/schema"
	"go-gin/pkg/gormplus"
)

type Dashboard struct {
	db *gormplus.DB
}

func NewDashboard(db *gormplus.DB) *Dashboard {
	return &Dashboard{db: db}
}

func (a *Dashboard) Counts(ctx context.Context) (schema.Object, error) {
	var posts, published, drafts, pending, users, bookmarks int64
	queries := []struct {
		table string
		where string
		args  []interface{}
		count *int64
	}{
		{
			table: "posts",
			where: "deletedAt IS NULL",
			count: &posts,
		},
		{
			table: "posts",
			where: "deletedAt IS NULL AND status = ?",
			args:  []interface{}{"PUBLISHED"},
			count: &published,
		},
		{
			table: "posts",
			where: "deletedAt IS NULL AND status = ?",
			args:  []interface{}{"DRAFT"},
			count: &drafts,
		},
		{
			table: "comments",
			where: "status = ?",
			args:  []interface{}{"PENDING"},
			count: &pending,
		},
		{
			table: "users",
			where: "deletedAt IS NULL",
			count: &users,
		},
		{
			table: "bookmarks",
			where: "1 = 1",
			count: &bookmarks,
		},
	}
	for _, query := range queries {
		if err := entity.GetDB(ctx, a.db.DB).
			Table(query.table).
			Where(query.where, query.args...).
			Count(query.count).Error; err != nil {
			return nil, err
		}
	}
	return schema.Object{
		"posts": schema.Object{
			"total":     posts,
			"published": published,
			"drafts":    drafts,
		},
		"pendingComments": pending,
		"users":           users,
		"bookmarks":       bookmarks,
	}, nil
}
