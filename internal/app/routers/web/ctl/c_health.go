package ctl

import (
	"go-gin/internal/app/bll"
	"go-gin/internal/app/ginplus"

	"github.com/gin-gonic/gin"
)

func NewHealth(service bll.IHealth) *Health {
	return &Health{service: service}
}

type Health struct {
	service bll.IHealth
}

func (a *Health) Live(c *gin.Context) {
	result, err := a.service.Live(c.Request.Context())
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *Health) Ready(c *gin.Context) {
	result, err := a.service.Ready(c.Request.Context())
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}
