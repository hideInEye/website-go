package ctl

import (
	"go-gin/internal/app/bll"
	"go-gin/internal/app/ginplus"
	"go-gin/internal/app/schema"

	"github.com/gin-gonic/gin"
)

type Post struct {
	PostBll bll.IPost
}

func NewPost(posts bll.IPost) *Post {
	return &Post{PostBll: posts}
}

func contentActor(c *gin.Context) schema.User {
	if user := ginplus.CurrentUser(c); user != nil {
		return *user
	}
	return schema.User{}
}

func (a *Post) ListPublic(c *gin.Context) {
	query, err := schema.ParsePostQuery(c.Request.URL.Query())
	if err != nil {
		ginplus.Fail(c, 400, err.Error())
		return
	}
	result, err := a.PostBll.ListPublic(c.Request.Context(), query)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Post) FindPublicBySlug(c *gin.Context) {
	result, err := a.PostBll.FindPublicBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Post) FindByLegacyID(c *gin.Context) {
	result, err := a.PostBll.FindByLegacyID(c.Request.Context(), c.Param("recordId"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Post) ListAdmin(c *gin.Context) {
	query, err := schema.ParsePostQuery(c.Request.URL.Query())
	if err != nil {
		ginplus.Fail(c, 400, err.Error())
		return
	}
	result, err := a.PostBll.ListAdmin(c.Request.Context(), contentActor(c), query)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Post) FindAdminByID(c *gin.Context) {
	result, err := a.PostBll.FindAdminByID(c.Request.Context(), contentActor(c), c.Param("id"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Post) Create(c *gin.Context) {
	var input schema.PostInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.PostBll.Create(c.Request.Context(), contentActor(c), input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 201, result)
}

func (a *Post) Update(c *gin.Context) {
	var input schema.PostInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.PostBll.Update(c.Request.Context(), contentActor(c), c.Param("id"), input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Post) Remove(c *gin.Context) {
	result, err := a.PostBll.Remove(c.Request.Context(), contentActor(c), c.Param("id"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}
