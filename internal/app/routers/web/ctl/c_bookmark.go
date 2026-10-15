package ctl

import (
	"go-gin/internal/app/bll"
	"go-gin/internal/app/ginplus"

	"github.com/gin-gonic/gin"
)

type Bookmark struct {
	BookmarkBll bll.IBookmark
}

func NewBookmark(bookmarks bll.IBookmark) *Bookmark {
	return &Bookmark{BookmarkBll: bookmarks}
}

func (a *Bookmark) HasBookmark(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	result, err := a.BookmarkBll.HasBookmark(c.Request.Context(), *user, c.Param("postId"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Bookmark) ListBookmarks(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	result, err := a.BookmarkBll.ListBookmarks(c.Request.Context(), *user)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Bookmark) AddBookmark(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	result, err := a.BookmarkBll.AddBookmark(c.Request.Context(), *user, c.Param("postId"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Bookmark) RemoveBookmark(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	result, err := a.BookmarkBll.RemoveBookmark(c.Request.Context(), *user, c.Param("postId"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}
