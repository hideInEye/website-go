package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-gin/internal/app/ginplus"
)

// TraceMiddleware 跟踪ID中间件
func TraceMiddleware(skipper ...SkipperFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(skipper) > 0 && skipper[0](c) {
			c.Next()
			return
		}

		// 优先从请求头中获取请求ID，如果没有则使用UUID
		traceID := c.GetHeader("X-Request-Id")
		if traceID == "" || len(traceID) > 128 {
			traceID = uuid.NewString()
		}
		c.Set(ginplus.TraceIDKey, traceID)
		c.Header("X-Request-ID", traceID)
		c.Next()
	}
}
