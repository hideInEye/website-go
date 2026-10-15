package web

import (
	"testing"

	"go-gin/internal/app/bll"
	"go-gin/internal/app/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// These stubs supply interfaces only for router construction. No business
// method is invoked, so route registration never needs MySQL, MinIO or SMTP.
type routeAuthStub struct {
	bll.IAuth
}

type routeUserStub struct {
	bll.IUser
}

type routePostStub struct {
	bll.IPost
}

type routeTaxonomyStub struct {
	bll.ITaxonomy
}

type routeCommentStub struct {
	bll.IComment
}

type routeBookmarkStub struct {
	bll.IBookmark
}

type routeDashboardStub struct {
	bll.IDashboard
}

type routeMediaStub struct {
	bll.IMedia
}

type routeFeedStub struct {
	bll.IFeed
}

type routeHealthStub struct {
	bll.IHealth
}

func TestRoutesPreserveExistingContract(t *testing.T) {
	container := dig.New()
	providers := []interface{}{
		func() *config.Config { return &config.Config{} },
		func() bll.IAuth { return &routeAuthStub{} },
		func() bll.IUser { return &routeUserStub{} },
		func() bll.IPost { return &routePostStub{} },
		func() bll.ITaxonomy { return &routeTaxonomyStub{} },
		func() bll.IComment { return &routeCommentStub{} },
		func() bll.IBookmark { return &routeBookmarkStub{} },
		func() bll.IDashboard { return &routeDashboardStub{} },
		func() bll.IMedia { return &routeMediaStub{} },
		func() bll.IFeed { return &routeFeedStub{} },
		func() bll.IHealth { return &routeHealthStub{} },
	}
	for _, provider := range providers {
		if err := container.Provide(provider); err != nil {
			t.Fatal(err)
		}
	}
	router := gin.New()
	if err := RegisterRouter(router, container); err != nil {
		t.Fatal(err)
	}

	// The 44 original Nest API routes, plus the four former frontend outputs.
	// Gin shares :slug between article and comment routes; the latter still
	// passes the original post ID to its controller, not an article slug.
	want := []string{
		"GET /rss.xml",
		"GET /sitemap.xml",
		"GET /robots.txt",
		"GET /article/detail",
		"GET /api/v1/health/live",
		"GET /api/v1/health/ready",
		"POST /api/v1/auth/register",
		"POST /api/v1/auth/verify-email",
		"POST /api/v1/auth/login",
		"POST /api/v1/auth/refresh",
		"POST /api/v1/auth/logout",
		"POST /api/v1/auth/forgot-password",
		"POST /api/v1/auth/reset-password",
		"GET /api/v1/auth/me",
		"GET /api/v1/posts",
		"GET /api/v1/posts/legacy/:recordId",
		"GET /api/v1/posts/:slug",
		"GET /api/v1/posts/:slug/comments",
		"GET /api/v1/categories",
		"GET /api/v1/tags",
		"GET /api/v1/users/:username",
		"GET /api/v1/users/me/profile",
		"PATCH /api/v1/users/me/profile",
		"GET /api/v1/comments/mine",
		"POST /api/v1/comments",
		"DELETE /api/v1/comments/:id",
		"GET /api/v1/bookmarks",
		"GET /api/v1/bookmarks/:postId",
		"PUT /api/v1/bookmarks/:postId",
		"DELETE /api/v1/bookmarks/:postId",
		"POST /api/v1/media/presign",
		"POST /api/v1/media/:id/complete",
		"GET /api/v1/admin/dashboard",
		"GET /api/v1/admin/posts",
		"GET /api/v1/admin/posts/:id",
		"POST /api/v1/admin/posts",
		"PATCH /api/v1/admin/posts/:id",
		"DELETE /api/v1/admin/posts/:id",
		"POST /api/v1/admin/categories",
		"PATCH /api/v1/admin/categories/:id",
		"DELETE /api/v1/admin/categories/:id",
		"POST /api/v1/admin/tags",
		"PATCH /api/v1/admin/tags/:id",
		"DELETE /api/v1/admin/tags/:id",
		"GET /api/v1/admin/comments",
		"PATCH /api/v1/admin/comments/:id",
		"GET /api/v1/admin/users",
		"PATCH /api/v1/admin/users/:id/status",
	}
	routes := make(map[string]bool)
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range want {
		if !routes[route] {
			t.Errorf("missing route %s", route)
		}
	}
	if len(routes) != len(want) {
		t.Errorf("unexpected route count: got %d, want %d", len(routes), len(want))
	}
}
