package bll

import "context"

type IFeed interface {
	RSS(context.Context) ([]byte, error)
	Sitemap(context.Context) ([]byte, error)
	Robots(context.Context) ([]byte, error)
	LegacyRedirect(context.Context, string) (string, error)
}
