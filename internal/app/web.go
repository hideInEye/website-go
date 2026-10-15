package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go-gin/internal/app/config"
	"go-gin/internal/app/middleware"
	"go-gin/internal/app/routers/web"
	"go-gin/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// InitWeb 初始化web引擎
func InitWeb(container *dig.Container) *gin.Engine {
	cfg := config.GetGlobalConfig()
	gin.SetMode(cfg.RunMode)
	gin.EnableJsonDecoderDisallowUnknownFields()

	app := gin.New()
	handleError(app.SetTrustedProxies(cfg.HTTP.TrustedProxies))
	app.HandleMethodNotAllowed = true
	app.NoMethod(middleware.NoMethodHandler())
	app.NoRoute(middleware.NoRouteHandler())

	// 跟踪ID
	app.Use(middleware.TraceMiddleware())

	// 访问日志
	app.Use(middleware.LoggerMiddleware(middleware.AllowPathPrefixNoSkipper("/api/v1/")))

	// 崩溃恢复
	app.Use(middleware.RecoveryMiddleware())

	// 原站点 Cookie 登录所需的来源校验与跨域配置。
	app.Use(middleware.SecurityMiddleware(cfg))

	// 在模板路由入口注册原站点业务 API。
	err := web.RegisterRouter(app, container)
	handleError(err)

	// swagger文档
	if dir := cfg.Swagger; dir != "" {
		app.Static("/swagger", dir)
	}
	// 静态站点
	if dir := cfg.WWW; dir != "" {
		app.Use(middleware.WWWMiddleware(dir))
	}

	return app
}

// InitHTTPServer 初始化http服务
func InitHTTPServer(ctx context.Context, container *dig.Container) func() {
	cfg := config.GetGlobalConfig().HTTP
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           InitWeb(container),
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      300 * time.Second,
		IdleTimeout:       15 * time.Second,
	}

	go func() {
		logger.Printf(ctx, "HTTP服务开始启动，地址监听在：[%s]", addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Errorf(ctx, "%v", err)
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(cfg.ShutdownTimeout))
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Errorf(ctx, "%v", err)
		}
	}
}
