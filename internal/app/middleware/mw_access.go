package middleware

import (
	"net/http"
	"strings"

	"go-gin/internal/app/bll"
	"go-gin/internal/app/ginplus"

	"github.com/gin-gonic/gin"
)

func RequireAuth(service bll.IAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authenticateRequest(c, service) {
			c.Next()
		}
	}
}

func RequireAdmin(service bll.IAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authenticateRequest(c, service) {
			return
		}
		if ginplus.CurrentUser(c).Role != "ADMIN" {
			ginplus.Fail(c, http.StatusForbidden, "没有执行此操作的权限")
			return
		}
		c.Next()
	}
}

func authenticateRequest(c *gin.Context, service bll.IAuth) bool {
	if ginplus.CurrentUser(c) != nil {
		return true
	}
	token, err := c.Cookie("sf_access")
	if err != nil {
		fields := strings.Fields(c.GetHeader("Authorization"))
		if len(fields) == 2 && strings.EqualFold(fields[0], "Bearer") {
			token = fields[1]
		}
	}
	user, err := service.Authenticate(c.Request.Context(), token)
	if err != nil {
		ginplus.ResError(c, err)
		c.Abort()
		return false
	}
	ginplus.SetCurrentUser(c, user)
	return true
}
