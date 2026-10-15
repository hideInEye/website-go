package ctl

import "go.uber.org/dig"

// Inject registers the HTTP controllers used by RegisterRouter.
func Inject(container *dig.Container) error {
	if err := container.Provide(NewAuth); err != nil {
		return err
	}
	if err := container.Provide(NewUser); err != nil {
		return err
	}
	if err := container.Provide(NewPost); err != nil {
		return err
	}
	if err := container.Provide(NewTaxonomy); err != nil {
		return err
	}
	if err := container.Provide(NewComment); err != nil {
		return err
	}
	if err := container.Provide(NewBookmark); err != nil {
		return err
	}
	if err := container.Provide(NewDashboard); err != nil {
		return err
	}
	if err := container.Provide(NewMedia); err != nil {
		return err
	}
	if err := container.Provide(NewFeed); err != nil {
		return err
	}
	if err := container.Provide(NewHealth); err != nil {
		return err
	}
	return nil
}
