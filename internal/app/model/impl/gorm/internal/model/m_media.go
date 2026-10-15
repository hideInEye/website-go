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

func NewMedia(db *gormplus.DB) *Media {
	return &Media{db: db}
}

type Media struct {
	db *gormplus.DB
}

var _ contract.IMedia = (*Media)(nil)

func (a *Media) Create(ctx context.Context, value *schema.Media) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	row := entity.Media(*value)
	return entity.GetDB(ctx, a.db.DB).Create(&row).Error
}

func (a *Media) FindPending(ctx context.Context, id, uploaderID string) (*schema.Media, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var row entity.Media
	err := entity.GetDB(ctx, a.db.DB).Where("id = ? AND uploaderId = ? AND status = ?", id, uploaderID, "PENDING").First(&row).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	result := schema.Media(row)
	return &result, nil
}

func (a *Media) Complete(ctx context.Context, id, uploaderID string, input schema.MediaCompleteInput, now time.Time) (*schema.Media, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	changes := map[string]interface{}{
		"status":    "READY",
		"updatedAt": now,
	}
	if input.Width != nil {
		changes["width"] = *input.Width
	}
	if input.Height != nil {
		changes["height"] = *input.Height
	}
	if input.AltText != nil {
		changes["altText"] = *input.AltText
	}
	db := entity.GetDB(ctx, a.db.DB)
	updated := db.Model(&entity.Media{}).Where("id = ? AND uploaderId = ? AND status = ?", id, uploaderID, "PENDING").Updates(changes)
	if updated.Error != nil {
		return nil, updated.Error
	}
	if updated.RowsAffected != 1 {
		return nil, nil
	}
	var row entity.Media
	if err := db.Where("id = ? AND uploaderId = ?", id, uploaderID).First(&row).Error; err != nil {
		return nil, err
	}
	result := schema.Media(row)
	return &result, nil
}
