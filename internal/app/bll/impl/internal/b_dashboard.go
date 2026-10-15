package internal

import (
	"context"

	"go-gin/internal/app/model"
	"go-gin/internal/app/schema"
)

type Dashboard struct {
	DashboardModel model.IDashboard
}

func NewDashboard(dashboard model.IDashboard) *Dashboard {
	return &Dashboard{DashboardModel: dashboard}
}

func (a *Dashboard) Dashboard(ctx context.Context, actor schema.User) (schema.Object, error) {
	if err := requireContentAdmin(actor); err != nil {
		return nil, err
	}
	return a.DashboardModel.Counts(ctx)
}
