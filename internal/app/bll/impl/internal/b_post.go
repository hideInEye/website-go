package internal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"time"

	"go-gin/internal/app/errors"
	"go-gin/internal/app/model"
	"go-gin/internal/app/schema"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

type Post struct {
	PostModel  model.IPost
	TransModel model.ITrans
}

func NewPost(posts model.IPost, transactions model.ITrans) *Post {
	return &Post{
		PostModel:  posts,
		TransModel: transactions,
	}
}

func requireContentAdmin(actor schema.User) error {
	if actor.ID == "" {
		return errors.NewHTTP(401, "请先登录")
	}
	if actor.Role != "ADMIN" {
		return errors.NewHTTP(403, "没有执行此操作的权限")
	}
	return nil
}

func contentID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

func (a *Post) ListPublic(ctx context.Context, query schema.PostQuery) (schema.Object, error) {
	posts, total, err := a.PostModel.ListPublic(ctx, query)
	if err != nil {
		return nil, err
	}
	items, err := a.PostModel.PresentPosts(ctx, posts, false, false)
	if err != nil {
		return nil, err
	}
	pages := (total + int64(query.PageSize) - 1) / int64(query.PageSize)
	if pages < 1 {
		pages = 1
	}
	return schema.Object{
		"items":      items,
		"page":       query.Page,
		"pageSize":   query.PageSize,
		"total":      total,
		"totalPages": pages,
	}, nil
}

func (a *Post) FindPublicBySlug(ctx context.Context, slug string) (schema.Object, error) {
	post, canonical, err := a.PostModel.FindPublicBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.NewHTTP(404, "文章不存在")
	}
	item, err := a.present(ctx, *post)
	if err != nil {
		return nil, err
	}
	if canonical {
		item["canonicalSlug"] = post.Slug
	}
	return item, nil
}

func (a *Post) FindByLegacyID(ctx context.Context, id string) (schema.Object, error) {
	post, err := a.PostModel.FindByLegacyID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.NewHTTP(404, "旧文章记录不存在")
	}
	return schema.Object{"slug": post.Slug}, nil
}

func (a *Post) ListAdmin(ctx context.Context, actor schema.User, query schema.PostQuery) (schema.Object, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	posts, total, err := a.PostModel.ListAdmin(ctx, query)
	if err != nil {
		return nil, err
	}
	items, err := a.PostModel.PresentPosts(ctx, posts, true, false)
	if err != nil {
		return nil, err
	}
	return schema.Object{
		"items": items,
		"total": total,
	}, nil
}

func (a *Post) FindAdminByID(ctx context.Context, actor schema.User, id string) (schema.Object, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	post, err := a.PostModel.FindAdminByID(ctx, id, false)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.NewHTTP(404, "文章不存在")
	}
	return a.present(ctx, *post)
}

func (a *Post) present(ctx context.Context, post schema.Post) (schema.Object, error) {
	items, err := a.PostModel.PresentPosts(ctx, []schema.Post{post}, true, false)
	if err != nil {
		return nil, err
	}
	return items[0], nil
}

var postSlugSymbols = regexp.MustCompile(`[^a-z0-9\x{4e00}-\x{9fff}]+`)

func normalizePostSlug(input string) string {
	value := strings.Trim(
		postSlugSymbols.ReplaceAllString(norm.NFKD.String(strings.ToLower(strings.TrimSpace(input))), "-"),
		"-",
	)
	runes := []rune(value)
	if len(runes) > 170 {
		value = string(runes[:170])
	}
	return value
}

func postSlugSuffix(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func (a *Post) uniqueSlug(ctx context.Context, input, currentID string) (string, error) {
	base := normalizePostSlug(input)
	if base == "" {
		suffix, err := postSlugSuffix(4)
		if err != nil {
			return "", err
		}
		base = "post-" + suffix
	}
	exists, err := a.PostModel.SlugExists(ctx, base, currentID)
	if err != nil || !exists {
		return base, err
	}
	if input == base {
		return "", errors.NewHTTP(409, "Slug 已被使用")
	}
	suffix, err := postSlugSuffix(3)
	if err != nil {
		return "", err
	}
	return base + "-" + suffix, nil
}

func (a *Post) Create(ctx context.Context, actor schema.User, input schema.PostInput) (schema.Object, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	if !input.Valid(true) {
		return nil, errors.NewHTTP(400, "文章参数不正确")
	}
	id := contentID()
	err := ExecTrans(ctx, a.TransModel, func(txctx context.Context) error {
		slugInput := input.Slug.Text("")
		if slugInput == "" {
			slugInput = input.Title.Text("")
		}
		slug, err := a.uniqueSlug(txctx, slugInput, "")
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		post := schema.Post{
			ID:             id,
			Slug:           slug,
			Title:          input.Title.Text(""),
			Excerpt:        input.Excerpt.Text(""),
			Content:        input.Content.Text(""),
			CoverURL:       input.CoverURL.Value,
			Status:         input.Status.Text("DRAFT"),
			SEOTitle:       input.SEOTitle.Value,
			SEODescription: input.SEODescription.Value,
			AuthorID:       actor.ID,
			CategoryID:     input.CategoryID.Text(""),
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if post.Status == "PUBLISHED" {
			post.PublishedAt = &now
		}
		return a.PostModel.Create(txctx, post, input.TagIDs.Value)
	})
	if err != nil {
		return nil, err
	}
	return a.FindAdminByID(ctx, actor, id)
}

func (a *Post) Update(ctx context.Context, actor schema.User, id string, input schema.PostInput) (schema.Object, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	if !input.Valid(false) {
		return nil, errors.NewHTTP(400, "文章参数不正确")
	}
	err := ExecTrans(ctx, a.TransModel, func(txctx context.Context) error {
		current, err := a.PostModel.FindAdminByID(txctx, id, true)
		if err != nil {
			return err
		}
		if current == nil {
			return errors.NewHTTP(404, "文章不存在")
		}
		nextSlug := current.Slug
		if value := input.Slug.Text(""); value != "" {
			nextSlug, err = a.uniqueSlug(txctx, value, id)
			if err != nil {
				return err
			}
		}
		if nextSlug != current.Slug {
			if err := a.PostModel.SaveSlugHistory(txctx, id, current.Slug); err != nil {
				return err
			}
		}
		values := map[string]interface{}{
			"slug":      nextSlug,
			"updatedAt": time.Now().UTC(),
		}
		for name, field := range map[string]schema.OptionalString{
			"title":          input.Title,
			"excerpt":        input.Excerpt,
			"content":        input.Content,
			"coverUrl":       input.CoverURL,
			"status":         input.Status,
			"seoTitle":       input.SEOTitle,
			"seoDescription": input.SEODescription,
			"categoryId":     input.CategoryID,
		} {
			if field.Present {
				if field.Value == nil {
					values[name] = nil
				} else {
					values[name] = *field.Value
				}
			}
		}
		if input.Status.Text("") == "PUBLISHED" && current.Status != "PUBLISHED" {
			values["publishedAt"] = time.Now().UTC()
		}
		return a.PostModel.Update(txctx, id, schema.PostChanges{
			Values:      values,
			TagIDs:      input.TagIDs.Value,
			ReplaceTags: input.TagIDs.Present && input.TagIDs.Value != nil,
		})
	})
	if err != nil {
		return nil, err
	}
	return a.FindAdminByID(ctx, actor, id)
}

func (a *Post) Remove(ctx context.Context, actor schema.User, id string) (schema.Object, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	deleted, err := a.PostModel.Remove(ctx, id)
	if err != nil {
		return nil, err
	}
	if !deleted {
		return nil, errors.NewHTTP(404, "文章不存在")
	}
	return schema.Object{"deleted": true}, nil
}
