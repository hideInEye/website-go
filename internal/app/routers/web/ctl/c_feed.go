package ctl

import (
	"net/http"

	"go-gin/internal/app/bll"
	"go-gin/internal/app/ginplus"

	"github.com/gin-gonic/gin"
)

func NewFeed(service bll.IFeed) *Feed {
	return &Feed{service: service}
}

type Feed struct {
	service bll.IFeed
}

func (a *Feed) RSS(c *gin.Context) {
	content, err := a.service.RSS(c.Request.Context())
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	c.Header("Cache-Control", "public, max-age=0, s-maxage=3600")
	c.Data(http.StatusOK, "application/rss+xml; charset=utf-8", content)
}

func (a *Feed) Sitemap(c *gin.Context) {
	content, err := a.service.Sitemap(c.Request.Context())
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	c.Header("Cache-Control", "public, max-age=0, s-maxage=3600")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", content)
}

func (a *Feed) Robots(c *gin.Context) {
	content, err := a.service.Robots(c.Request.Context())
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", content)
}

func (a *Feed) LegacyRedirect(c *gin.Context) {
	location, err := a.service.LegacyRedirect(c.Request.Context(), c.Query("id"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	c.Redirect(http.StatusPermanentRedirect, location)
}
