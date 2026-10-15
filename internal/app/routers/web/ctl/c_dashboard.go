package ctl

import (
	"go-gin/internal/app/bll"
	"go-gin/internal/app/ginplus"

	"github.com/gin-gonic/gin"
)

type Dashboard struct {
	DashboardBll bll.IDashboard
}

func NewDashboard(dashboard bll.IDashboard) *Dashboard {
	return &Dashboard{DashboardBll: dashboard}
}

func (a *Dashboard) Dashboard(c *gin.Context) {
	result, err := a.DashboardBll.Dashboard(c.Request.Context(), contentActor(c))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}
