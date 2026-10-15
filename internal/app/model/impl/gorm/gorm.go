package gorm

import (
	"go-gin/internal/app/model"
	"go-gin/internal/app/model/impl/gorm/internal/entity"
	imodel "go-gin/internal/app/model/impl/gorm/internal/model"
	"go-gin/pkg/gormplus"
	"go.uber.org/dig"
)

func SetTablePrefix(prefix string) {
	entity.SetTablePrefix(prefix)
}

// AutoMigrate is retained for the template plugin; deployments use Migrate.
func AutoMigrate(db *gormplus.DB) error {
	return db.AutoMigrate(
		new(entity.User),
		new(entity.Session),
		new(entity.VerificationToken),
		new(entity.ResetToken),
		new(entity.Category),
		new(entity.Tag),
		new(entity.Post),
		new(entity.PostTag),
		new(entity.SlugHistory),
		new(entity.Comment),
		new(entity.Bookmark),
		new(entity.Media),
	).Error
}

// Inject registers business storage in the template's dependency container.
func Inject(container *dig.Container) error {
	if err := container.Provide(imodel.NewTrans, dig.As(new(model.ITrans))); err != nil {
		return err
	}
	if err := container.Provide(imodel.NewAuth, dig.As(new(model.IAuth))); err != nil {
		return err
	}
	if err := container.Provide(imodel.NewUser, dig.As(new(model.IUser))); err != nil {
		return err
	}
	if err := container.Provide(imodel.NewPost, dig.As(new(model.IPost))); err != nil {
		return err
	}
	if err := container.Provide(imodel.NewTaxonomy, dig.As(new(model.ITaxonomy))); err != nil {
		return err
	}
	if err := container.Provide(imodel.NewComment, dig.As(new(model.IComment))); err != nil {
		return err
	}
	if err := container.Provide(imodel.NewBookmark, dig.As(new(model.IBookmark))); err != nil {
		return err
	}
	if err := container.Provide(imodel.NewDashboard, dig.As(new(model.IDashboard))); err != nil {
		return err
	}
	if err := container.Provide(imodel.NewMedia, dig.As(new(model.IMedia))); err != nil {
		return err
	}
	if err := container.Provide(imodel.NewFeed, dig.As(new(model.IFeed))); err != nil {
		return err
	}
	if err := container.Provide(imodel.NewHealth, dig.As(new(model.IHealth))); err != nil {
		return err
	}
	return nil
}
