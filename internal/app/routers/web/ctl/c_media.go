package ctl

import (
	"go-gin/internal/app/bll"
	"go-gin/internal/app/ginplus"
	"go-gin/internal/app/schema"

	"github.com/gin-gonic/gin"
)

func NewMedia(service bll.IMedia) *Media {
	return &Media{service: service}
}

type Media struct {
	service bll.IMedia
}

func (a *Media) Presign(c *gin.Context) {
	var input schema.MediaPresignInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.service.Presign(c.Request.Context(), ginplus.CurrentUser(c).ID, input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 201, result)
}

func (a *Media) Complete(c *gin.Context) {
	var input schema.MediaCompleteInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.service.Complete(c.Request.Context(), ginplus.CurrentUser(c).ID, c.Param("id"), input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 201, result)
}
