package web

import (
	"time"

	"go-gin/internal/app/bll"
	"go-gin/internal/app/middleware"
	"go-gin/internal/app/routers/web/ctl"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// RegisterRouter keeps all HTTP registration in the template's router entry.
func RegisterRouter(app *gin.Engine, container *dig.Container) error {
	if err := ctl.Inject(container); err != nil {
		return err
	}

	return container.Invoke(func(
		auth bll.IAuth,
		cAuth *ctl.Auth,
		cUser *ctl.User,
		cPost *ctl.Post,
		cTaxonomy *ctl.Taxonomy,
		cComment *ctl.Comment,
		cBookmark *ctl.Bookmark,
		cDashboard *ctl.Dashboard,
		cMedia *ctl.Media,
		cFeed *ctl.Feed,
		cHealth *ctl.Health,
	) {
		app.GET("/rss.xml", cFeed.RSS)
		app.GET("/sitemap.xml", cFeed.Sitemap)
		app.GET("/robots.txt", cFeed.Robots)
		app.GET("/article/detail", cFeed.LegacyRedirect)

		api := app.Group("/api/v1")
		api.GET("/health/live", cHealth.Live)
		api.GET("/health/ready", cHealth.Ready)
		api.Use(middleware.APIRateLimit(120, time.Minute))

		account := api.Group("/auth")
		account.POST("/register", middleware.APIRateLimit(4, time.Minute), cAuth.Register)
		account.POST("/verify-email", middleware.APIRateLimit(8, time.Minute), cAuth.VerifyEmail)
		account.POST("/login", middleware.APIRateLimit(8, time.Minute), cAuth.Login)
		account.POST("/refresh", cAuth.Refresh)
		account.POST("/logout", cAuth.Logout)
		account.POST("/forgot-password", middleware.APIRateLimit(4, time.Minute), cAuth.ForgotPassword)
		account.POST("/reset-password", middleware.APIRateLimit(6, time.Minute), cAuth.ResetPassword)
		account.GET("/me", middleware.RequireAuth(auth), cAuth.Me)

		api.GET("/posts", cPost.ListPublic)
		api.GET("/posts/legacy/:recordId", cPost.FindByLegacyID)
		api.GET("/posts/:slug", cPost.FindPublicBySlug)
		// Gin requires the same wildcard name; comments receive a post ID here.
		api.GET("/posts/:slug/comments", cComment.PublicComments)
		api.GET("/categories", cTaxonomy.Categories)
		api.GET("/tags", cTaxonomy.Tags)
		api.GET("/users/:username", cUser.PublicProfile)

		member := api.Group("")
		member.Use(middleware.RequireAuth(auth))
		member.GET("/users/me/profile", cUser.MyProfile)
		member.PATCH("/users/me/profile", cUser.UpdateProfile)
		member.GET("/comments/mine", cComment.MyComments)
		member.POST("/comments", middleware.APIRateLimit(10, time.Minute), cComment.CreateComment)
		member.DELETE("/comments/:id", cComment.DeleteComment)
		member.GET("/bookmarks", cBookmark.ListBookmarks)
		member.GET("/bookmarks/:postId", cBookmark.HasBookmark)
		member.PUT("/bookmarks/:postId", cBookmark.AddBookmark)
		member.DELETE("/bookmarks/:postId", cBookmark.RemoveBookmark)
		member.POST("/media/presign", cMedia.Presign)
		member.POST("/media/:id/complete", cMedia.Complete)

		admin := api.Group("/admin")
		admin.Use(middleware.RequireAdmin(auth))
		admin.GET("/dashboard", cDashboard.Dashboard)
		admin.GET("/posts", cPost.ListAdmin)
		admin.GET("/posts/:id", cPost.FindAdminByID)
		admin.POST("/posts", cPost.Create)
		admin.PATCH("/posts/:id", cPost.Update)
		admin.DELETE("/posts/:id", cPost.Remove)
		admin.POST("/categories", cTaxonomy.CreateCategory)
		admin.PATCH("/categories/:id", cTaxonomy.UpdateCategory)
		admin.DELETE("/categories/:id", cTaxonomy.DeleteCategory)
		admin.POST("/tags", cTaxonomy.CreateTag)
		admin.PATCH("/tags/:id", cTaxonomy.UpdateTag)
		admin.DELETE("/tags/:id", cTaxonomy.DeleteTag)
		admin.GET("/comments", cComment.PendingComments)
		admin.PATCH("/comments/:id", cComment.ModerateComment)
		admin.GET("/users", cUser.ListUsers)
		admin.PATCH("/users/:id/status", cUser.SetUserStatus)
	})
}
