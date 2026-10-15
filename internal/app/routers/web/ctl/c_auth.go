package ctl

import (
	"net/http"
	"time"

	"go-gin/internal/app/bll"
	"go-gin/internal/app/config"
	"go-gin/internal/app/ginplus"
	"go-gin/internal/app/schema"

	"github.com/gin-gonic/gin"
)

func NewAuth(service bll.IAuth, cfg *config.Config) *Auth {
	return &Auth{
		service: service,
		config:  cfg,
	}
}

type Auth struct {
	service bll.IAuth
	config  *config.Config
}

func (a *Auth) Register(c *gin.Context) {
	var input schema.AuthRegisterInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.service.Register(c.Request.Context(), input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, http.StatusCreated, result)
}

func (a *Auth) VerifyEmail(c *gin.Context) {
	var input schema.AuthTokenInput
	if !ginplus.Bind(c, &input) {
		return
	}
	if err := a.service.VerifyEmail(c.Request.Context(), input.Token); err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, http.StatusCreated, gin.H{"verified": true})
}

func (a *Auth) Login(c *gin.Context) {
	var input schema.AuthLoginInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.service.Login(c.Request.Context(), input, authRequestMeta(c))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	a.setCookies(c, result)
	ginplus.ResData(c, http.StatusCreated, result)
}

func (a *Auth) Refresh(c *gin.Context) {
	token, _ := c.Cookie("sf_refresh")
	result, err := a.service.Refresh(c.Request.Context(), token, authRequestMeta(c))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	a.setCookies(c, result)
	ginplus.ResData(c, http.StatusCreated, result)
}

func (a *Auth) Logout(c *gin.Context) {
	token, _ := c.Cookie("sf_refresh")
	if err := a.service.Logout(c.Request.Context(), token); err != nil {
		ginplus.ResError(c, err)
		return
	}
	cookiePaths := map[string]string{
		"sf_access":  "/",
		"sf_refresh": "/api/v1/auth",
	}
	for name, path := range cookiePaths {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     path,
			MaxAge:   -1,
			Expires:  time.Unix(1, 0),
			HttpOnly: true,
			Secure:   a.config.JWTAuth.CookieSecure,
			SameSite: http.SameSiteLaxMode,
		})
	}
	ginplus.ResData(c, http.StatusCreated, gin.H{"loggedOut": true})
}

func (a *Auth) ForgotPassword(c *gin.Context) {
	var input schema.AuthForgotPasswordInput
	if !ginplus.Bind(c, &input) {
		return
	}
	if err := a.service.ForgotPassword(c.Request.Context(), input); err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, http.StatusCreated, gin.H{"accepted": true})
}

func (a *Auth) ResetPassword(c *gin.Context) {
	var input schema.AuthResetPasswordInput
	if !ginplus.Bind(c, &input) {
		return
	}
	if err := a.service.ResetPassword(c.Request.Context(), input); err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, http.StatusCreated, gin.H{"reset": true})
}

func (a *Auth) Me(c *gin.Context) {
	user := ginplus.CurrentUser(c)
	if user == nil {
		ginplus.Fail(c, http.StatusUnauthorized, "请先登录")
		return
	}
	ginplus.ResData(c, http.StatusOK, schema.PublicAuthUser(user))
}

func authRequestMeta(c *gin.Context) schema.AuthSessionMeta {
	return schema.AuthSessionMeta{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}

func (a *Auth) setCookies(c *gin.Context, session *schema.AuthSessionResult) {
	for _, cookie := range []http.Cookie{
		{
			Name:   "sf_access",
			Value:  session.AccessToken,
			Path:   "/",
			MaxAge: int(a.config.JWTAuth.Expired),
		},
		{
			Name:   "sf_refresh",
			Value:  session.RefreshToken,
			Path:   "/api/v1/auth",
			MaxAge: a.config.JWTAuth.RefreshTTLDays * 86400,
		},
	} {
		cookie.HttpOnly = true
		cookie.Secure = a.config.JWTAuth.CookieSecure
		cookie.SameSite = http.SameSiteLaxMode
		http.SetCookie(c.Writer, &cookie)
	}
}
