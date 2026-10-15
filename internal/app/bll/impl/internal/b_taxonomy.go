package internal

import (
	"context"
	"time"

	"go-gin/internal/app/errors"
	"go-gin/internal/app/model"
	"go-gin/internal/app/schema"
)

type Taxonomy struct {
	TaxonomyModel model.ITaxonomy
	TransModel    model.ITrans
}

func NewTaxonomy(taxonomy model.ITaxonomy, transactions model.ITrans) *Taxonomy {
	return &Taxonomy{
		TaxonomyModel: taxonomy,
		TransModel:    transactions,
	}
}

func (a *Taxonomy) Categories(ctx context.Context) ([]schema.CategoryWithCount, error) {
	return a.TaxonomyModel.Categories(ctx)
}

func (a *Taxonomy) Tags(ctx context.Context) ([]schema.TagWithCount, error) {
	return a.TaxonomyModel.Tags(ctx)
}

func (a *Taxonomy) CreateCategory(ctx context.Context, actor schema.User, input schema.TaxonomyInput) (*schema.Category, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	if !input.Valid(true) {
		return nil, errors.NewHTTP(400, "分类名称、链接或描述不正确")
	}
	now := time.Now().UTC()
	category := schema.Category{
		ID:          contentID(),
		Name:        input.Name.Text(""),
		Slug:        input.Slug.Text(""),
		Description: input.Description.Value,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := a.TaxonomyModel.CreateCategory(ctx, category); err != nil {
		return nil, err
	}
	return a.TaxonomyModel.FindCategory(ctx, category.ID, false)
}

func (a *Taxonomy) CreateTag(ctx context.Context, actor schema.User, input schema.TaxonomyInput) (*schema.Tag, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	if !input.Valid(true) {
		return nil, errors.NewHTTP(400, "标签名称或链接不正确")
	}
	now := time.Now().UTC()
	tag := schema.Tag{
		ID:        contentID(),
		Name:      input.Name.Text(""),
		Slug:      input.Slug.Text(""),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := a.TaxonomyModel.CreateTag(ctx, tag); err != nil {
		return nil, err
	}
	return a.TaxonomyModel.FindTag(ctx, tag.ID, false)
}

func taxonomyValues(input schema.TaxonomyInput, category bool) map[string]interface{} {
	values := map[string]interface{}{"updatedAt": time.Now().UTC()}
	if input.Name.Present {
		values["name"] = input.Name.Text("")
	}
	if input.Slug.Present {
		values["slug"] = input.Slug.Text("")
	}
	if category && input.Description.Present {
		if input.Description.Value == nil {
			values["description"] = nil
		} else {
			values["description"] = *input.Description.Value
		}
	}
	return values
}

func (a *Taxonomy) UpdateCategory(ctx context.Context, actor schema.User, id string, input schema.TaxonomyInput) (*schema.Category, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	if !input.Valid(false) {
		return nil, errors.NewHTTP(400, "分类名称、链接或描述不正确")
	}
	err := ExecTrans(ctx, a.TransModel, func(txctx context.Context) error {
		category, err := a.TaxonomyModel.FindCategory(txctx, id, true)
		if err != nil {
			return err
		}
		if category == nil {
			return errors.NewHTTP(404, "记录不存在")
		}
		return a.TaxonomyModel.UpdateCategory(txctx, id, taxonomyValues(input, true))
	})
	if err != nil {
		return nil, err
	}
	return a.TaxonomyModel.FindCategory(ctx, id, false)
}

func (a *Taxonomy) UpdateTag(ctx context.Context, actor schema.User, id string, input schema.TaxonomyInput) (*schema.Tag, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	if !input.Valid(false) {
		return nil, errors.NewHTTP(400, "标签名称或链接不正确")
	}
	err := ExecTrans(ctx, a.TransModel, func(txctx context.Context) error {
		tag, err := a.TaxonomyModel.FindTag(txctx, id, true)
		if err != nil {
			return err
		}
		if tag == nil {
			return errors.NewHTTP(404, "记录不存在")
		}
		return a.TaxonomyModel.UpdateTag(txctx, id, taxonomyValues(input, false))
	})
	if err != nil {
		return nil, err
	}
	return a.TaxonomyModel.FindTag(ctx, id, false)
}

func (a *Taxonomy) DeleteCategory(ctx context.Context, actor schema.User, id string) (schema.Object, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	err := ExecTrans(ctx, a.TransModel, func(txctx context.Context) error {
		category, err := a.TaxonomyModel.FindCategory(txctx, id, true)
		if err != nil {
			return err
		}
		if category == nil {
			return errors.NewHTTP(404, "记录不存在")
		}
		count, err := a.TaxonomyModel.CategoryPostCount(txctx, id)
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.NewHTTP(409, "仍有文章使用此分类")
		}
		return a.TaxonomyModel.DeleteCategory(txctx, id)
	})
	if err != nil {
		return nil, err
	}
	return schema.Object{"deleted": true}, nil
}

func (a *Taxonomy) DeleteTag(ctx context.Context, actor schema.User, id string) (schema.Object, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	deleted, err := a.TaxonomyModel.DeleteTag(ctx, id)
	if err != nil {
		return nil, err
	}
	if !deleted {
		return nil, errors.NewHTTP(404, "记录不存在")
	}
	return schema.Object{"deleted": true}, nil
}
