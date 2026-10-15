package model

import (
	"context"
	"time"

	contract "go-gin/internal/app/model"
	"go-gin/internal/app/model/impl/gorm/internal/entity"
	"go-gin/internal/app/schema"
	"go-gin/pkg/gormplus"

	"github.com/jinzhu/gorm"
)

func NewFeed(db *gormplus.DB) *Feed {
	return &Feed{db: db}
}

type Feed struct {
	db *gormplus.DB
}

var _ contract.IFeed = (*Feed)(nil)

func (a *Feed) ListPosts(ctx context.Context, limit int) ([]schema.FeedPost, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	query := entity.GetDB(ctx, a.db.DB).Select("slug, title, excerpt, publishedAt, createdAt").
		Where("status = ? AND publishedAt <= ? AND deletedAt IS NULL", "PUBLISHED", time.Now().UTC()).Order("publishedAt DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	rows := []entity.Post{}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	posts := make([]schema.FeedPost, 0, len(rows))
	for _, row := range rows {
		posts = append(posts, schema.FeedPost{
			Slug:        row.Slug,
			Title:       row.Title,
			Excerpt:     row.Excerpt,
			PublishedAt: row.PublishedAt,
			CreatedAt:   row.CreatedAt,
		})
	}
	return posts, nil
}

func (a *Feed) ListCategorySlugs(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	slugs := []string{}
	err := entity.GetDB(ctx, a.db.DB).Model(&entity.Category{}).Pluck("slug", &slugs).Error
	return slugs, err
}

func (a *Feed) ListTagSlugs(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	slugs := []string{}
	err := entity.GetDB(ctx, a.db.DB).Model(&entity.Tag{}).Pluck("slug", &slugs).Error
	return slugs, err
}

func (a *Feed) FindLegacySlug(ctx context.Context, id string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var row entity.Post
	err := entity.GetDB(ctx, a.db.DB).Select("slug").
		Where("legacyRecordId = ? AND status = ? AND deletedAt IS NULL", id, "PUBLISHED").First(&row).Error
	if gorm.IsRecordNotFoundError(err) {
		return "", nil
	}
	return row.Slug, err
}
