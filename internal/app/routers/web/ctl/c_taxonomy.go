package ctl

import (
	"go-gin/internal/app/bll"
	"go-gin/internal/app/ginplus"
	"go-gin/internal/app/schema"

	"github.com/gin-gonic/gin"
)

type Taxonomy struct {
	TaxonomyBll bll.ITaxonomy
}

func NewTaxonomy(taxonomy bll.ITaxonomy) *Taxonomy {
	return &Taxonomy{TaxonomyBll: taxonomy}
}

func (a *Taxonomy) Categories(c *gin.Context) {
	result, err := a.TaxonomyBll.Categories(c.Request.Context())
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Taxonomy) Tags(c *gin.Context) {
	result, err := a.TaxonomyBll.Tags(c.Request.Context())
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Taxonomy) CreateCategory(c *gin.Context) {
	var input schema.TaxonomyInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.TaxonomyBll.CreateCategory(c.Request.Context(), contentActor(c), input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 201, result)
}

func (a *Taxonomy) CreateTag(c *gin.Context) {
	var input schema.TaxonomyInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.TaxonomyBll.CreateTag(c.Request.Context(), contentActor(c), input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 201, result)
}

func (a *Taxonomy) UpdateCategory(c *gin.Context) {
	var input schema.TaxonomyInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.TaxonomyBll.UpdateCategory(c.Request.Context(), contentActor(c), c.Param("id"), input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Taxonomy) UpdateTag(c *gin.Context) {
	var input schema.TaxonomyInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.TaxonomyBll.UpdateTag(c.Request.Context(), contentActor(c), c.Param("id"), input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Taxonomy) DeleteCategory(c *gin.Context) {
	result, err := a.TaxonomyBll.DeleteCategory(c.Request.Context(), contentActor(c), c.Param("id"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Taxonomy) DeleteTag(c *gin.Context) {
	result, err := a.TaxonomyBll.DeleteTag(c.Request.Context(), contentActor(c), c.Param("id"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}
