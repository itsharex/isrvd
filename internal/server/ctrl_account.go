package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"isrvd/internal/helper"
	"isrvd/internal/service/account"
)

// defineAccountRoutes 定义 Account 模块路由
func (app *App) defineAccountRoutes() []Route {
	return []Route{
		{Method: "GET", Path: "/account/info", Handler: app.accountAuthInfo, Module: "account", Label: "获取认证信息", Access: account.AccessAnon},
		{Method: "POST", Path: "/account/login", Handler: app.accountLogin, Module: "account", Label: "登录账户", Access: account.AccessAnon},
		{Method: "GET", Path: "/account/routes", Handler: app.accountListRoutes, Module: "account", Label: "列出路由权限", Access: account.AccessAuth},
		{Method: "POST", Path: "/account/token", Handler: app.accountCreateApiToken, Module: "account", Label: "创建 API 令牌"},
		{Method: "PUT", Path: "/account/password", Handler: app.accountChangePassword, Module: "account", Label: "修改密码", Access: account.AccessAuth},
		{Method: "GET", Path: "/account/members", Handler: app.accountListMembers, Module: "account", Label: "列出成员"},
		{Method: "POST", Path: "/account/member", Handler: app.accountCreateMember, Module: "account", Label: "创建成员"},
		{Method: "PUT", Path: "/account/member/:username", Handler: app.accountUpdateMember, Module: "account", Label: "更新成员"},
		{Method: "DELETE", Path: "/account/member/:username", Handler: app.accountDeleteMember, Module: "account", Label: "删除成员"},
	}
}

// accountListRoutes 返回所有已注册路由及其权限元信息
func (app *App) accountListRoutes(c *gin.Context) {
	routes := make([]account.RouteInfo, 0, len(app.routePerms))
	for key, info := range app.routePerms {
		info.Key = key
		routes = append(routes, info)
	}
	helper.RespondSuccess(c, "ok", routes)
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

// accountCreateApiToken 创建长效 API Token
func (app *App) accountCreateApiToken(c *gin.Context) {
	var req account.CreateApiTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	username := c.GetString("username")
	resp, err := app.accountSvc.CreateApiToken(username, req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "令牌创建成功", resp)
}

// accountChangePassword 修改当前用户密码
func (app *App) accountChangePassword(c *gin.Context) {
	var req account.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	username := c.GetString("username")
	if err := app.accountSvc.ChangePassword(username, req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	helper.RespondSuccess(c, "密码修改成功", nil)
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
