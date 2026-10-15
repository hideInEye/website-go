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

type User struct {
	db *gormplus.DB
}

var _ contract.IUser = (*User)(nil)

func NewUser(db *gormplus.DB) *User {
	return &User{db: db}
}

func (a *User) database(ctx context.Context) *gorm.DB {
	return entity.GetDB(ctx, a.db.DB)
}

func (a *User) FindPublic(ctx context.Context, username string) (*schema.User, error) {
	var row entity.User
	err := a.database(ctx).Where("username = ? AND status = ? AND deletedAt IS NULL", username, "ACTIVE").First(&row).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	user := schema.User(row)
	return &user, nil
}

func (a *User) UpdateProfile(ctx context.Context, id string, updates map[string]interface{}) (*schema.User, error) {
	var row entity.User
	err := ExecTrans(ctx, a.db, func(txctx context.Context) error {
		db := a.database(txctx)
		if err := db.Model(&entity.User{}).Where("id = ? AND deletedAt IS NULL", id).Updates(updates).Error; err != nil {
			return err
		}
		return db.Where("id = ? AND deletedAt IS NULL", id).First(&row).Error
	})
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	user := schema.User(row)
	return &user, nil
}

func (a *User) List(ctx context.Context) ([]schema.User, error) {
	rows := []entity.User{}
	if err := a.database(ctx).Unscoped().Order("createdAt DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	users := make([]schema.User, len(rows))
	for index, row := range rows {
		users[index] = schema.User(row)
	}
	return users, nil
}

func (a *User) SetStatus(ctx context.Context, id, status string) (*schema.User, error) {
	var row entity.User
	err := ExecTrans(ctx, a.db, func(txctx context.Context) error {
		db := a.database(txctx)
		if err := db.Unscoped().Set("gorm:query_option", "FOR UPDATE").Where("id = ?", id).First(&row).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := db.Unscoped().Model(&entity.User{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":    status,
			"updatedAt": now,
		}).Error; err != nil {
			return err
		}
		row.Status = status
		row.UpdatedAt = now
		return nil
	})
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	user := schema.User(row)
	return &user, nil
}
