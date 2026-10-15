package impl

import (
	"go-gin/internal/app/bll"
	"go-gin/internal/app/bll/impl/internal"
	"go.uber.org/dig"
)

// Inject registers business services in the template's dependency container.
func Inject(container *dig.Container) error {
	if err := container.Provide(internal.NewTrans, dig.As(new(bll.ITrans))); err != nil {
		return err
	}
	if err := container.Provide(internal.NewAuth, dig.As(new(bll.IAuth))); err != nil {
		return err
	}
	if err := container.Provide(internal.NewUser, dig.As(new(bll.IUser))); err != nil {
		return err
	}
	if err := container.Provide(internal.NewPost, dig.As(new(bll.IPost))); err != nil {
		return err
	}
	if err := container.Provide(internal.NewTaxonomy, dig.As(new(bll.ITaxonomy))); err != nil {
		return err
	}
	if err := container.Provide(internal.NewComment, dig.As(new(bll.IComment))); err != nil {
		return err
	}
	if err := container.Provide(internal.NewBookmark, dig.As(new(bll.IBookmark))); err != nil {
		return err
	}
	if err := container.Provide(internal.NewDashboard, dig.As(new(bll.IDashboard))); err != nil {
		return err
	}
	if err := container.Provide(internal.NewMedia, dig.As(new(bll.IMedia))); err != nil {
		return err
	}
	if err := container.Provide(internal.NewFeed, dig.As(new(bll.IFeed))); err != nil {
		return err
	}
	if err := container.Provide(internal.NewHealth, dig.As(new(bll.IHealth))); err != nil {
		return err
	}
	if err := container.Provide(internal.NewMail, dig.As(new(bll.IMail))); err != nil {
		return err
	}
	return nil
}
