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

func NewAuth(db *gormplus.DB) *Auth {
	return &Auth{db: db.DB}
}

type Auth struct {
	db *gorm.DB
}

var _ contract.IAuth = (*Auth)(nil)

func (a *Auth) query(ctx context.Context) *gorm.DB {
	db := entity.GetDB(ctx, a.db).New()
	if err := ctx.Err(); err != nil {
		db.AddError(err)
	}
	return db
}

func authFind(db *gorm.DB, value interface{}) (bool, error) {
	err := db.First(value).Error
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	}
	return err == nil, err
}

func authChanged(result *gorm.DB) (bool, error) {
	return result.RowsAffected > 0, result.Error
}

func (a *Auth) FindUserByIdentity(ctx context.Context, email, username string) (*schema.User, error) {
	var item entity.User
	ok, err := authFind(a.query(ctx).Unscoped().Where("email = ? OR username = ?", email, username), &item)
	if !ok {
		return nil, err
	}
	result := schema.User(item)
	return &result, nil
}

func (a *Auth) FindUserByEmail(ctx context.Context, email string) (*schema.User, error) {
	var item entity.User
	ok, err := authFind(a.query(ctx).Where("email = ? AND deletedAt IS NULL", email), &item)
	if !ok {
		return nil, err
	}
	result := schema.User(item)
	return &result, nil
}

func (a *Auth) FindActiveUser(ctx context.Context, id string) (*schema.User, error) {
	var item entity.User
	ok, err := authFind(a.query(ctx).Where("id = ? AND status = ? AND deletedAt IS NULL", id, "ACTIVE"), &item)
	if !ok {
		return nil, err
	}
	result := schema.User(item)
	return &result, nil
}

func (a *Auth) CreateUser(ctx context.Context, user *schema.User) error {
	item := entity.User(*user)
	return a.query(ctx).Create(&item).Error
}

func (a *Auth) ActivateUser(ctx context.Context, id string, now time.Time) (bool, error) {
	result := a.query(ctx).Model(&entity.User{}).Where("id = ? AND deletedAt IS NULL", id).
		Updates(map[string]interface{}{
			"status":          "ACTIVE",
			"emailVerifiedAt": now,
			"updatedAt":       now,
		})
	return authChanged(result)
}

func (a *Auth) UpdatePassword(ctx context.Context, id, hash string, now time.Time) (bool, error) {
	result := a.query(ctx).Model(&entity.User{}).Where("id = ? AND deletedAt IS NULL", id).
		Updates(map[string]interface{}{
			"passwordHash": hash,
			"updatedAt":    now,
		})
	return authChanged(result)
}

func (a *Auth) CreateVerification(ctx context.Context, token *schema.VerificationToken) error {
	item := entity.VerificationToken(*token)
	return a.query(ctx).Create(&item).Error
}

func (a *Auth) FindVerification(ctx context.Context, hash string, lock bool) (*schema.VerificationToken, error) {
	query := a.query(ctx).Where("tokenHash = ?", hash)
	if lock {
		query = query.Set("gorm:query_option", "FOR UPDATE")
	}
	var item entity.VerificationToken
	ok, err := authFind(query, &item)
	if !ok {
		return nil, err
	}
	result := schema.VerificationToken(item)
	return &result, nil
}

func (a *Auth) ConsumeVerification(ctx context.Context, id string, now time.Time) (bool, error) {
	result := a.query(ctx).Model(&entity.VerificationToken{}).
		Where("id = ? AND usedAt IS NULL AND expiresAt > ?", id, now).UpdateColumn("usedAt", now)
	return authChanged(result)
}

func (a *Auth) CreateResetToken(ctx context.Context, token *schema.ResetToken) error {
	item := entity.ResetToken(*token)
	return a.query(ctx).Create(&item).Error
}

func (a *Auth) FindResetToken(ctx context.Context, hash string) (*schema.ResetToken, error) {
	var item entity.ResetToken
	ok, err := authFind(a.query(ctx).Where("tokenHash = ?", hash), &item)
	if !ok {
		return nil, err
	}
	result := schema.ResetToken(item)
	return &result, nil
}

func (a *Auth) ConsumeResetToken(ctx context.Context, id string, now time.Time) (bool, error) {
	result := a.query(ctx).Model(&entity.ResetToken{}).
		Where("id = ? AND usedAt IS NULL AND expiresAt > ?", id, now).UpdateColumn("usedAt", now)
	return authChanged(result)
}

func (a *Auth) CreateSession(ctx context.Context, session *schema.Session) error {
	item := entity.Session(*session)
	return a.query(ctx).Create(&item).Error
}

func (a *Auth) FindSession(ctx context.Context, hash string, lock bool) (*schema.Session, error) {
	query := a.query(ctx).Where("tokenHash = ?", hash)
	if lock {
		query = query.Set("gorm:query_option", "FOR UPDATE")
	}
	var item entity.Session
	ok, err := authFind(query, &item)
	if !ok {
		return nil, err
	}
	result := schema.Session(item)
	return &result, nil
}

func (a *Auth) RevokeSession(ctx context.Context, id string, now time.Time) error {
	return a.query(ctx).Model(&entity.Session{}).Where("id = ?", id).UpdateColumn("revokedAt", now).Error
}

func (a *Auth) RevokeSessionToken(ctx context.Context, hash string, now time.Time) error {
	return a.query(ctx).Model(&entity.Session{}).Where("tokenHash = ? AND revokedAt IS NULL", hash).
		UpdateColumn("revokedAt", now).Error
}

func (a *Auth) RevokeUserSessions(ctx context.Context, userID string, now time.Time) error {
	return a.query(ctx).Model(&entity.Session{}).Where("userId = ? AND revokedAt IS NULL", userID).
		UpdateColumn("revokedAt", now).Error
}
