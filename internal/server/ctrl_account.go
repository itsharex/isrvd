package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"isrvd/internal/helper"
	"isrvd/internal/service/account"
)

// defineAccountRoutes 定义 Account 模块受保护路由（成员管理）
// 公开路由（/auth/info、/auth/login）由 initRoutes 直接注册
func (app *App) defineAccountRoutes() []Route {
	return []Route{
		{Method: "GET", Path: "/account/members", Handler: app.accountListMembers, Module: "account", Label: "成员管理", Perm: "r"},
		{Method: "POST", Path: "/account/members", Handler: app.accountCreateMember, Module: "account", Label: "成员管理", Perm: "rw"},
		{Method: "PUT", Path: "/account/members/:username", Handler: app.accountUpdateMember, Module: "account", Label: "成员管理", Perm: "rw"},
		{Method: "DELETE", Path: "/account/members/:username", Handler: app.accountDeleteMember, Module: "account", Label: "成员管理", Perm: "rw"},
	}
}

// accountAuthInfo 返回当前认证模式及已登录用户信息
func (app *App) accountAuthInfo(c *gin.Context) {
	username := c.GetString("username")
	helper.RespondSuccess(c, "ok", app.accountSvc.GetAuthInfo(username))
}

// accountLogin 校验用户名密码并签发 JWT Token
func (app *App) accountLogin(c *gin.Context) {
	var req account.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := app.accountSvc.Login(req)
	if err != nil {
		helper.RespondError(c, http.StatusUnauthorized, err.Error())
		return
	}
	helper.RespondSuccess(c, "登录成功", resp)
}

// accountListMembers 列出所有成员
func (app *App) accountListMembers(c *gin.Context) {
	helper.RespondSuccess(c, "ok", app.accountSvc.ListMembers())
}

// accountCreateMember 新建成员
func (app *App) accountCreateMember(c *gin.Context) {
	var req account.MemberUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.accountSvc.CreateMember(req); err != nil {
		switch {
		case errors.Is(err, account.ErrInvalidRequest), errors.Is(err, account.ErrMemberExists):
			helper.RespondError(c, http.StatusBadRequest, err.Error())
		default:
			helper.RespondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	helper.RespondSuccess(c, "成员添加成功", nil)
}

// accountUpdateMember 更新成员
func (app *App) accountUpdateMember(c *gin.Context) {
	username := c.Param("username")
	var req account.MemberUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.accountSvc.UpdateMember(username, req); err != nil {
		if errors.Is(err, account.ErrMemberNotFound) {
			helper.RespondError(c, http.StatusNotFound, err.Error())
		} else {
			helper.RespondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	helper.RespondSuccess(c, "成员更新成功", nil)
}

// accountDeleteMember 删除成员
func (app *App) accountDeleteMember(c *gin.Context) {
	username := c.Param("username")
	if err := app.accountSvc.DeleteMember(username); err != nil {
		switch {
		case errors.Is(err, account.ErrMemberNotFound):
			helper.RespondError(c, http.StatusNotFound, err.Error())
		default:
			helper.RespondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	helper.RespondSuccess(c, "成员删除成功", nil)
}
