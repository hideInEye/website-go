package middleware

import (
	"net/http"

	"go-gin/internal/app/config"
	"go-gin/internal/app/ginplus"

	"github.com/gin-gonic/gin"
)

// SecurityMiddleware keeps the existing browser origin and request-size rules.
func SecurityMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		origin := c.GetHeader("Origin")
		if origin == cfg.Site.Origin {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		if c.Request.Method == http.MethodOptions {
			if origin != "" && origin != cfg.Site.Origin {
				ginplus.Fail(c, http.StatusForbidden, "请求来源未被允许")
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			if origin != "" && origin != cfg.Site.Origin {
				ginplus.Fail(c, http.StatusForbidden, "请求来源未被允许")
				return
			}
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 12*1024*1024)
		c.Next()
	}
}
