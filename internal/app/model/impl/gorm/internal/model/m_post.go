package model

import (
	"context"
	"strings"
	"time"
	"unicode/utf16"

	"go-gin/internal/app/model/impl/gorm/internal/entity"
	"go-gin/internal/app/schema"
	"go-gin/pkg/gormplus"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Post struct {
	db *gormplus.DB
}

func NewPost(db *gormplus.DB) *Post {
	return &Post{db: db}
}

func (a *Post) database(ctx context.Context) *gorm.DB {
	return entity.GetDB(ctx, a.db.DB)
}

func postSchemas(rows []entity.Post) []schema.Post {
	posts := make([]schema.Post, len(rows))
	for index, row := range rows {
		posts[index] = schema.Post(row)
	}
	return posts
}

func (a *Post) ListPublic(ctx context.Context, query schema.PostQuery) ([]schema.Post, int64, error) {
	base := a.database(ctx).
		Model(&entity.Post{}).
		Where("status = ? AND deletedAt IS NULL AND publishedAt <= ?", "PUBLISHED", time.Now().UTC())
	if query.Category != "" {
		base = base.Where("categoryId IN (SELECT id FROM categories WHERE slug = ?)", query.Category)
	}
	if query.Tag != "" {
		base = base.Where(
			"id IN (SELECT pt.postId FROM post_tags pt INNER JOIN tags t ON t.id = pt.tagId WHERE t.slug = ?)",
			query.Tag,
		)
	}
	if query.Search != "" {
		base = base.Where("MATCH(title, excerpt, content) AGAINST (? IN NATURAL LANGUAGE MODE)", query.Search)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := []entity.Post{}
	err := base.Order("publishedAt DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&rows).Error
	return postSchemas(rows), total, err
}

func (a *Post) ListAdmin(ctx context.Context, query schema.PostQuery) ([]schema.Post, int64, error) {
	base := a.database(ctx).Model(&entity.Post{}).Where("deletedAt IS NULL")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := []entity.Post{}
	err := base.Order("updatedAt DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&rows).Error
	return postSchemas(rows), total, err
}

func (a *Post) FindPublicBySlug(ctx context.Context, slug string) (*schema.Post, bool, error) {
	var post entity.Post
	err := a.database(ctx).Where("slug = ? AND status = ? AND deletedAt IS NULL", slug, "PUBLISHED").First(&post).Error
	canonical := false
	if gorm.IsRecordNotFoundError(err) {
		var history entity.SlugHistory
		err = a.database(ctx).Where("slug = ?", slug).First(&history).Error
		if err == nil {
			err = a.database(ctx).
				Where("id = ? AND status = ? AND deletedAt IS NULL", history.PostID, "PUBLISHED").
				First(&post).Error
			canonical = true
		}
	}
	if gorm.IsRecordNotFoundError(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	result := schema.Post(post)
	return &result, canonical, nil
}

func (a *Post) FindByLegacyID(ctx context.Context, id string) (*schema.Post, error) {
	var post entity.Post
	err := a.database(ctx).
		Where("legacyRecordId = ? AND status = ? AND deletedAt IS NULL", id, "PUBLISHED").
		First(&post).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	result := schema.Post(post)
	return &result, nil
}

func (a *Post) FindAdminByID(ctx context.Context, id string, lock bool) (*schema.Post, error) {
	query := a.database(ctx).Where("id = ? AND deletedAt IS NULL", id)
	if lock {
		query = query.Set("gorm:query_option", "FOR UPDATE")
	}
	var post entity.Post
	if err := query.First(&post).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	result := schema.Post(post)
	return &result, nil
}

func (a *Post) SlugExists(ctx context.Context, slug, currentID string) (bool, error) {
	query := a.database(ctx).Unscoped().Model(&entity.Post{}).Where("slug = ?", slug)
	if currentID != "" {
		query = query.Where("id <> ?", currentID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func (a *Post) Create(ctx context.Context, post schema.Post, tags []string) error {
	row := entity.Post(post)
	if err := a.database(ctx).Create(&row).Error; err != nil {
		return err
	}
	for _, tagID := range tags {
		if err := a.database(ctx).Create(&entity.PostTag{
			PostID: post.ID,
			TagID:  tagID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (a *Post) SaveSlugHistory(ctx context.Context, postID, slug string) error {
	return a.database(ctx).Exec(
		"INSERT INTO post_slug_history (id, postId, slug, createdAt) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE slug = slug",
		strings.ReplaceAll(uuid.NewString(), "-", ""),
		postID,
		slug,
		time.Now().UTC(),
	).Error
}

func (a *Post) Update(ctx context.Context, id string, changes schema.PostChanges) error {
	if changes.ReplaceTags {
		if err := a.database(ctx).Where("postId = ?", id).Delete(&entity.PostTag{}).Error; err != nil {
			return err
		}
		for _, tagID := range changes.TagIDs {
			if err := a.database(ctx).Create(&entity.PostTag{
				PostID: id,
				TagID:  tagID,
			}).Error; err != nil {
				return err
			}
		}
	}
	return a.database(ctx).Model(&entity.Post{}).Where("id = ? AND deletedAt IS NULL", id).Updates(changes.Values).Error
}

func (a *Post) Remove(ctx context.Context, id string) (bool, error) {
	now := time.Now().UTC()
	result := a.database(ctx).
		Model(&entity.Post{}).
		Where("id = ? AND deletedAt IS NULL", id).
		Updates(map[string]interface{}{
			"deletedAt": now,
			"updatedAt": now,
		})
	return result.RowsAffected > 0, result.Error
}

type postCountRow struct {
	PostID string `gorm:"column:postId"`
	Total  int64  `gorm:"column:total"`
}

type postTagRow struct {
	PostID    string    `gorm:"column:postId"`
	ID        string    `gorm:"column:id"`
	Name      string    `gorm:"column:name"`
	Slug      string    `gorm:"column:slug"`
	CreatedAt time.Time `gorm:"column:createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt"`
}

// PresentPosts batches the original Prisma include fields for lists and bookmarks.
func (a *Post) PresentPosts(ctx context.Context, posts []schema.Post, includeContent, bookmarkShape bool) ([]schema.Object, error) {
	items := make([]schema.Object, 0, len(posts))
	if len(posts) == 0 {
		return items, nil
	}
	ids, authorIDs, categoryIDs := []string{}, []string{}, []string{}
	for _, post := range posts {
		ids = append(ids, post.ID)
		authorIDs = append(authorIDs, post.AuthorID)
		categoryIDs = append(categoryIDs, post.CategoryID)
	}
	users, categories := []entity.User{}, []entity.Category{}
	if err := a.database(ctx).Unscoped().Where("id IN (?)", authorIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	if err := a.database(ctx).Where("id IN (?)", categoryIDs).Find(&categories).Error; err != nil {
		return nil, err
	}
	userMap, categoryMap := make(map[string]entity.User), make(map[string]entity.Category)
	for _, user := range users {
		userMap[user.ID] = user
	}
	for _, category := range categories {
		categoryMap[category.ID] = category
	}
	tagRows := []postTagRow{}
	if err := a.database(ctx).
		Table("post_tags pt").
		Select("pt.postId, t.id, t.name, t.slug, t.createdAt, t.updatedAt").
		Joins("INNER JOIN tags t ON t.id = pt.tagId").
		Where("pt.postId IN (?)", ids).
		Scan(&tagRows).Error; err != nil {
		return nil, err
	}
	tagMap := make(map[string][]schema.Tag)
	for _, row := range tagRows {
		tagMap[row.PostID] = append(tagMap[row.PostID], schema.Tag{
			ID:        row.ID,
			Name:      row.Name,
			Slug:      row.Slug,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	bookmarks, comments := []postCountRow{}, []postCountRow{}
	if err := a.database(ctx).
		Table("bookmarks").
		Select("postId, COUNT(*) AS total").
		Where("postId IN (?)", ids).
		Group("postId").
		Scan(&bookmarks).Error; err != nil {
		return nil, err
	}
	if err := a.database(ctx).
		Table("comments").
		Select("postId, COUNT(*) AS total").
		Where("postId IN (?) AND status = ?", ids, "APPROVED").
		Group("postId").
		Scan(&comments).Error; err != nil {
		return nil, err
	}
	bookmarkCount, commentCount := make(map[string]int64), make(map[string]int64)
	for _, row := range bookmarks {
		bookmarkCount[row.PostID] = row.Total
	}
	for _, row := range comments {
		commentCount[row.PostID] = row.Total
	}
	for _, post := range posts {
		category, user := categoryMap[post.CategoryID], userMap[post.AuthorID]
		tags := tagMap[post.ID]
		if tags == nil {
			tags = []schema.Tag{}
		}
		minutes := (len(utf16.Encode([]rune(post.Content))) + 499) / 500
		if minutes < 1 {
			minutes = 1
		}
		item := schema.Object{
			"id":             post.ID,
			"legacyRecordId": post.LegacyRecordID,
			"slug":           post.Slug,
			"title":          post.Title,
			"excerpt":        post.Excerpt,
			"coverUrl":       post.CoverURL,
			"seoTitle":       post.SEOTitle,
			"seoDescription": post.SEODescription,
			"status":         post.Status,
			"publishedAt":    post.PublishedAt,
			"readingMinutes": minutes,
			"author": schema.Object{
				"username":    user.Username,
				"displayName": user.DisplayName,
				"avatarUrl":   user.AvatarURL,
				"bio":         user.Bio,
			},
			"category": schema.Object{
				"id":          category.ID,
				"name":        category.Name,
				"slug":        category.Slug,
				"description": category.Description,
			},
			"tags":          tags,
			"bookmarkCount": bookmarkCount[post.ID],
			"commentCount":  commentCount[post.ID],
		}
		if includeContent {
			item["content"] = post.Content
			item["updatedAt"] = post.UpdatedAt
		}
		if bookmarkShape {
			item["content"] = post.Content
			item["createdAt"] = post.CreatedAt
			item["updatedAt"] = post.UpdatedAt
			item["deletedAt"] = post.DeletedAt
			item["authorId"] = post.AuthorID
			item["categoryId"] = post.CategoryID
			item["category"] = schema.Category(category)
			item["_count"] = schema.Object{
				"bookmarks": bookmarkCount[post.ID],
				"comments":  commentCount[post.ID],
			}
			delete(item, "readingMinutes")
		}
		items = append(items, item)
	}
	return items, nil
}
