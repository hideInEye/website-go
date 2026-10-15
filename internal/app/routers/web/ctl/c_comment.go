package ctl

import (
	"go-gin/internal/app/bll"
	"go-gin/internal/app/ginplus"
	"go-gin/internal/app/schema"

	"github.com/gin-gonic/gin"
)

type Comment struct {
	CommentBll bll.IComment
}

func NewComment(comments bll.IComment) *Comment {
	return &Comment{CommentBll: comments}
}

func (a *Comment) PublicComments(c *gin.Context) {
	// The existing /posts/:slug/comments route carries a post ID in :slug.
	result, err := a.CommentBll.PublicComments(c.Request.Context(), c.Param("slug"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Comment) MyComments(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	result, err := a.CommentBll.MyComments(c.Request.Context(), *user)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Comment) CreateComment(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	var input schema.CommentInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.CommentBll.CreateComment(c.Request.Context(), *user, input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 201, result)
}

func (a *Comment) DeleteComment(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	result, err := a.CommentBll.DeleteComment(c.Request.Context(), *user, c.Param("id"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Comment) PendingComments(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	result, err := a.CommentBll.PendingComments(c.Request.Context(), *user)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Comment) ModerateComment(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	var input schema.CommentStatusInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.CommentBll.ModerateComment(c.Request.Context(), *user, c.Param("id"), input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}
