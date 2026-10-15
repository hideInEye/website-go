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

type Comment struct {
	db *gormplus.DB
}

var _ contract.IComment = (*Comment)(nil)

func NewComment(db *gormplus.DB) *Comment {
	return &Comment{db: db}
}

func (a *Comment) database(ctx context.Context) *gorm.DB {
	return entity.GetDB(ctx, a.db.DB)
}

func commentSchemas(rows []entity.Comment) []schema.Comment {
	comments := make([]schema.Comment, len(rows))
	for index, row := range rows {
		comments[index] = schema.Comment(row)
	}
	return comments
}

func (a *Comment) PublicRoots(ctx context.Context, postID string) ([]schema.Comment, error) {
	rows := []entity.Comment{}
	err := a.database(ctx).Where("postId = ? AND status = ? AND deletedAt IS NULL AND parentId IS NULL", postID, "APPROVED").
		Order("createdAt DESC").Find(&rows).Error
	return commentSchemas(rows), err
}

func (a *Comment) PublicReplies(ctx context.Context, parentIDs []string) ([]schema.Comment, error) {
	rows := []entity.Comment{}
	if len(parentIDs) == 0 {
		return []schema.Comment{}, nil
	}
	err := a.database(ctx).Where("parentId IN (?) AND status = ? AND deletedAt IS NULL", parentIDs, "APPROVED").
		Order("createdAt ASC").Find(&rows).Error
	return commentSchemas(rows), err
}

func (a *Comment) ByUser(ctx context.Context, userID, status string, limit int) ([]schema.Comment, error) {
	query := a.database(ctx).Where("userId = ? AND deletedAt IS NULL", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	rows := []entity.Comment{}
	err := query.Order("createdAt DESC").Find(&rows).Error
	return commentSchemas(rows), err
}

func (a *Comment) Relations(ctx context.Context, comments []schema.Comment) (map[string]schema.User, map[string]schema.Post, error) {
	users := make(map[string]schema.User)
	posts := make(map[string]schema.Post)
	if len(comments) == 0 {
		return users, posts, nil
	}

	userIDs := make([]string, 0, len(comments))
	postIDs := make([]string, 0, len(comments))
	for _, comment := range comments {
		userIDs = append(userIDs, comment.UserID)
		postIDs = append(postIDs, comment.PostID)
	}

	userRows := []entity.User{}
	if err := a.database(ctx).Unscoped().Where("id IN (?)", userIDs).Find(&userRows).Error; err != nil {
		return nil, nil, err
	}
	for _, row := range userRows {
		users[row.ID] = schema.User(row)
	}

	postRows := []entity.Post{}
	if err := a.database(ctx).Unscoped().Where("id IN (?)", postIDs).Find(&postRows).Error; err != nil {
		return nil, nil, err
	}
	for _, row := range postRows {
		posts[row.ID] = schema.Post(row)
	}
	return users, posts, nil
}

func (a *Comment) PublishedPostExists(ctx context.Context, postID string) (bool, error) {
	var count int64
	err := a.database(ctx).Model(&entity.Post{}).Where("id = ? AND status = ? AND deletedAt IS NULL", postID, "PUBLISHED").Count(&count).Error
	return count > 0, err
}

func (a *Comment) FindReplyParent(ctx context.Context, id, postID string) (*schema.Comment, error) {
	var row entity.Comment
	err := a.database(ctx).Where("id = ? AND postId = ? AND parentId IS NULL AND status = ? AND deletedAt IS NULL", id, postID, "APPROVED").First(&row).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	comment := schema.Comment(row)
	return &comment, nil
}

func (a *Comment) CountApproved(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := a.database(ctx).Unscoped().Model(&entity.Comment{}).Where("userId = ? AND status = ?", userID, "APPROVED").Count(&count).Error
	return count, err
}

func (a *Comment) Create(ctx context.Context, comment schema.Comment) error {
	row := entity.Comment(comment)
	return a.database(ctx).Create(&row).Error
}

func (a *Comment) FindIncludingDeleted(ctx context.Context, id string) (*schema.Comment, error) {
	var row entity.Comment
	err := a.database(ctx).Unscoped().Where("id = ?", id).First(&row).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	comment := schema.Comment(row)
	return &comment, nil
}

func (a *Comment) SoftDelete(ctx context.Context, id string, now time.Time) error {
	return a.database(ctx).Unscoped().Model(&entity.Comment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    "DELETED",
		"deletedAt": now,
		"updatedAt": now,
		"content":   "[评论已删除]",
	}).Error
}

func (a *Comment) Pending(ctx context.Context) ([]schema.Comment, error) {
	rows := []entity.Comment{}
	err := a.database(ctx).Where("status = ? AND deletedAt IS NULL", "PENDING").Order("createdAt ASC").Find(&rows).Error
	return commentSchemas(rows), err
}

func (a *Comment) SetStatus(ctx context.Context, id, status string) (*schema.Comment, error) {
	var row entity.Comment
	err := ExecTrans(ctx, a.db, func(txctx context.Context) error {
		db := a.database(txctx)
		if err := db.Set("gorm:query_option", "FOR UPDATE").Where("id = ? AND deletedAt IS NULL", id).First(&row).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := db.Model(&entity.Comment{}).Where("id = ?", id).Updates(map[string]interface{}{
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
	comment := schema.Comment(row)
	return &comment, nil
}
