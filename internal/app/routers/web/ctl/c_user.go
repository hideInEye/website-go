package ctl

import (
	"go-gin/internal/app/bll"
	"go-gin/internal/app/ginplus"
	"go-gin/internal/app/schema"

	"github.com/gin-gonic/gin"
)

type User struct {
	UserBll bll.IUser
}

func NewUser(users bll.IUser) *User {
	return &User{UserBll: users}
}

func communityUser(c *gin.Context) *schema.User {
	user := ginplus.CurrentUser(c)
	if user == nil {
		ginplus.Fail(c, 401, "请先登录")
	}
	return user
}

func (a *User) PublicProfile(c *gin.Context) {
	result, err := a.UserBll.PublicProfile(c.Request.Context(), c.Param("username"))
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *User) MyProfile(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	result, err := a.UserBll.MyProfile(c.Request.Context(), *user)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *User) UpdateProfile(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	var input schema.ProfileInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.UserBll.UpdateProfile(c.Request.Context(), *user, input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *User) ListUsers(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	result, err := a.UserBll.ListUsers(c.Request.Context(), *user)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}

func (a *User) SetUserStatus(c *gin.Context) {
	user := communityUser(c)
	if user == nil {
		return
	}
	var input schema.UserStatusInput
	if !ginplus.Bind(c, &input) {
		return
	}
	result, err := a.UserBll.SetUserStatus(c.Request.Context(), *user, c.Param("id"), input)
	if err != nil {
		ginplus.ResError(c, err)
		return
	}
	ginplus.ResData(c, 200, result)
}
