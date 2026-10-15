package middleware

import (
	"time"

	"go-gin/internal/app/ginplus"
	"go-gin/pkg/logger"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware(skipper ...SkipperFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(skipper) > 0 && skipper[0](c) {
			c.Next()
			return
		}

		p := c.Request.URL.Path
		method := c.Request.Method
		span := logger.StartSpan(ginplus.NewContext(c), logger.SetSpanTitle("访问日志"), logger.SetSpanFuncName(JoinRouter(method, p)))
		start := time.Now()

		fields := make(map[string]interface{})
		fields["ip"] = c.ClientIP()
		fields["method"] = method
		// Query strings, headers and bodies can contain credentials or session tokens.
		fields["url"] = p
		fields["proto"] = c.Request.Proto
		fields["content_length"] = c.Request.ContentLength
		c.Next()

		timeConsuming := time.Since(start).Nanoseconds() / 1e6
		fields["res_status"] = c.Writer.Status()
		fields["res_length"] = c.Writer.Size()
		fields[logger.UserIDKey] = ginplus.GetUserID(c)
		span.WithFields(fields).Infof("[http] %s-%s-%s-%d(%dms)",
			p, c.Request.Method, c.ClientIP(), c.Writer.Status(), timeConsuming)
	}
}
