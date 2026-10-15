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

type Bookmark struct {
	db *gormplus.DB
}

var _ contract.IBookmark = (*Bookmark)(nil)

func NewBookmark(db *gormplus.DB) *Bookmark {
	return &Bookmark{db: db}
}

func (a *Bookmark) database(ctx context.Context) *gorm.DB {
	return entity.GetDB(ctx, a.db.DB)
}

func (a *Bookmark) Exists(ctx context.Context, userID, postID string) (bool, error) {
	var count int64
	err := a.database(ctx).Model(&entity.Bookmark{}).Where("userId = ? AND postId = ?", userID, postID).Count(&count).Error
	return count > 0, err
}

func (a *Bookmark) ListPosts(ctx context.Context, userID string) ([]schema.Post, error) {
	rows := []entity.Post{}
	err := a.database(ctx).Model(&entity.Post{}).
		Joins("INNER JOIN bookmarks ON bookmarks.postId = posts.id").
		Where("bookmarks.userId = ? AND posts.status = ? AND posts.deletedAt IS NULL", userID, "PUBLISHED").
		Order("bookmarks.createdAt DESC").
		Select("posts.*").
		Find(&rows).Error
	posts := make([]schema.Post, len(rows))
	for index, row := range rows {
		posts[index] = schema.Post(row)
	}
	return posts, err
}

func (a *Bookmark) PublishedPostExists(ctx context.Context, postID string) (bool, error) {
	var count int64
	err := a.database(ctx).Model(&entity.Post{}).Where("id = ? AND status = ? AND deletedAt IS NULL", postID, "PUBLISHED").Count(&count).Error
	return count > 0, err
}

func (a *Bookmark) Add(ctx context.Context, userID, postID string) error {
	return a.database(ctx).Exec(
		"INSERT INTO bookmarks (userId, postId, createdAt) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE postId = postId",
		userID, postID, time.Now().UTC(),
	).Error
}

func (a *Bookmark) Remove(ctx context.Context, userID, postID string) error {
	return a.database(ctx).Where("userId = ? AND postId = ?", userID, postID).Delete(&entity.Bookmark{}).Error
}
