package model

import (
	"context"

	"go-gin/internal/app/model/impl/gorm/internal/entity"
	"go-gin/internal/app/schema"
	"go-gin/pkg/gormplus"

	"github.com/jinzhu/gorm"
)

type Taxonomy struct {
	db *gormplus.DB
}

func NewTaxonomy(db *gormplus.DB) *Taxonomy {
	return &Taxonomy{db: db}
}

func (a *Taxonomy) database(ctx context.Context) *gorm.DB {
	return entity.GetDB(ctx, a.db.DB)
}

func (a *Taxonomy) Categories(ctx context.Context) ([]schema.CategoryWithCount, error) {
	rows := []struct {
		entity.Category
		PostCount int64 `gorm:"column:postCount"`
	}{}
	err := a.database(ctx).
		Table("categories").
		Select("categories.*, (SELECT COUNT(*) FROM posts WHERE posts.categoryId = categories.id AND posts.status = 'PUBLISHED' AND posts.deletedAt IS NULL) AS postCount").
		Order("categories.name ASC").
		Scan(&rows).Error
	result := make([]schema.CategoryWithCount, len(rows))
	for index, row := range rows {
		result[index] = schema.CategoryWithCount{
			Category:  schema.Category(row.Category),
			PostCount: row.PostCount,
		}
	}
	return result, err
}

func (a *Taxonomy) Tags(ctx context.Context) ([]schema.TagWithCount, error) {
	rows := []struct {
		entity.Tag
		PostCount int64 `gorm:"column:postCount"`
	}{}
	err := a.database(ctx).
		Table("tags").
		Select("tags.*, (SELECT COUNT(*) FROM post_tags pt INNER JOIN posts p ON p.id = pt.postId WHERE pt.tagId = tags.id AND p.status = 'PUBLISHED' AND p.deletedAt IS NULL) AS postCount").
		Order("tags.name ASC").
		Scan(&rows).Error
	result := make([]schema.TagWithCount, len(rows))
	for index, row := range rows {
		result[index] = schema.TagWithCount{
			Tag:       schema.Tag(row.Tag),
			PostCount: row.PostCount,
		}
	}
	return result, err
}

func (a *Taxonomy) FindCategory(ctx context.Context, id string, lock bool) (*schema.Category, error) {
	query := a.database(ctx).Where("id = ?", id)
	if lock {
		query = query.Set("gorm:query_option", "FOR UPDATE")
	}
	var row entity.Category
	if err := query.First(&row).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	result := schema.Category(row)
	return &result, nil
}

func (a *Taxonomy) FindTag(ctx context.Context, id string, lock bool) (*schema.Tag, error) {
	query := a.database(ctx).Where("id = ?", id)
	if lock {
		query = query.Set("gorm:query_option", "FOR UPDATE")
	}
	var row entity.Tag
	if err := query.First(&row).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	result := schema.Tag(row)
	return &result, nil
}

func (a *Taxonomy) CreateCategory(ctx context.Context, category schema.Category) error {
	row := entity.Category(category)
	return a.database(ctx).Create(&row).Error
}

func (a *Taxonomy) CreateTag(ctx context.Context, tag schema.Tag) error {
	row := entity.Tag(tag)
	return a.database(ctx).Create(&row).Error
}

func (a *Taxonomy) UpdateCategory(ctx context.Context, id string, values map[string]interface{}) error {
	return a.database(ctx).Model(&entity.Category{}).Where("id = ?", id).Updates(values).Error
}

func (a *Taxonomy) UpdateTag(ctx context.Context, id string, values map[string]interface{}) error {
	return a.database(ctx).Model(&entity.Tag{}).Where("id = ?", id).Updates(values).Error
}

func (a *Taxonomy) CategoryPostCount(ctx context.Context, id string) (int64, error) {
	var count int64
	err := a.database(ctx).Model(&entity.Post{}).Where("categoryId = ? AND deletedAt IS NULL", id).Count(&count).Error
	return count, err
}

func (a *Taxonomy) DeleteCategory(ctx context.Context, id string) error {
	return a.database(ctx).Where("id = ?", id).Delete(&entity.Category{}).Error
}

func (a *Taxonomy) DeleteTag(ctx context.Context, id string) (bool, error) {
	result := a.database(ctx).Where("id = ?", id).Delete(&entity.Tag{})
	return result.RowsAffected > 0, result.Error
}
