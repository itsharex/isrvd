package server

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"isrvd/internal/service/account"
)

// defineAccountRoutes 定义 Account 模块路由
func (app *App) defineAccountRoutes() []Route {
	return []Route{
		{Method: "GET", Path: "/account/info", Handler: app.accountAuthInfo, Module: "account", Label: "获取认证信息", Access: account.AccessAnon},
		{Method: "POST", Path: "/account/login", Handler: app.accountLogin, Module: "account", Label: "登录账户", Access: account.AccessAnon},
		{Method: "GET", Path: "/account/oidc/login", Handler: app.accountOIDCLogin, Module: "account", Label: "OIDC 登录", Access: account.AccessAnon},
		{Method: "GET", Path: "/account/oidc/callback", Handler: app.accountOIDCCallback, Module: "account", Label: "OIDC 回调", Access: account.AccessAnon},
		{Method: "POST", Path: "/account/oidc/exchange", Handler: app.accountOIDCExchange, Module: "account", Label: "OIDC 登录码交换", Access: account.AccessAnon},
		{Method: "GET", Path: "/account/routes", Handler: app.accountRouteList, Module: "account", Label: "列出路由权限", Access: account.AccessAuth},
		{Method: "POST", Path: "/account/token", Handler: app.accountApiTokenCreate, Module: "account", Label: "创建 API 令牌"},
		{Method: "PUT", Path: "/account/password", Handler: app.accountPasswordChange, Module: "account", Label: "修改密码", Access: account.AccessAuth},
		{Method: "GET", Path: "/account/members", Handler: app.accountMemberList, Module: "account", Label: "列出成员"},
		{Method: "POST", Path: "/account/member", Handler: app.accountMemberCreate, Module: "account", Label: "创建成员"},
		{Method: "PUT", Path: "/account/member/:username", Handler: app.accountMemberUpdate, Module: "account", Label: "更新成员"},
		{Method: "DELETE", Path: "/account/member/:username", Handler: app.accountMemberDelete, Module: "account", Label: "删除成员"},
	}
}

// accountRouteList 返回所有已注册路由及其权限元信息
func (app *App) accountRouteList(c *gin.Context) {
	routes := make([]account.RouteInfo, 0, len(app.routePerms))
	for key, info := range app.routePerms {
		info.Key = key
		routes = append(routes, info)
	}
	respondSuccess(c, "ok", routes)
}

// accountAuthInfo 返回当前认证模式及已登录用户信息
func (app *App) accountAuthInfo(c *gin.Context) {
	username := c.GetString("username")
	respondSuccess(c, "ok", app.accountSvc.AuthInfo(username))
}

// accountLogin 校验用户名密码并签发 JWT Token
func (app *App) accountLogin(c *gin.Context) {
	var req account.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := app.accountSvc.Login(req)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err.Error())
		return
	}
	respondSuccess(c, "登录成功", resp)
}

// accountOIDCLogin 跳转到 OIDC Provider 登录页
func (app *App) accountOIDCLogin(c *gin.Context) {
	loginURL, err := app.accountSvc.OIDCLoginURL(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Redirect(http.StatusFound, loginURL)
}

// accountOIDCCallback 处理 OIDC Provider 回调
func (app *App) accountOIDCCallback(c *gin.Context) {
	code, err := app.accountSvc.OIDCCallback(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/?oidc_error="+url.QueryEscape(err.Error()))
		return
	}
	c.Redirect(http.StatusFound, "/?oidc_code="+url.QueryEscape(code))
}

// accountOIDCExchange 使用一次性登录码换取 JWT Token
func (app *App) accountOIDCExchange(c *gin.Context) {
	var req account.OIDCExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := app.accountSvc.OIDCExchange(req.Code)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err.Error())
		return
	}
	respondSuccess(c, "登录成功", resp)
}

// accountApiTokenCreate 创建长效 API Token
func (app *App) accountApiTokenCreate(c *gin.Context) {
	var req account.CreateApiTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	username := c.GetString("username")
	resp, err := app.accountSvc.ApiTokenCreate(username, req)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "令牌创建成功", resp)
}

// accountPasswordChange 修改当前用户密码
func (app *App) accountPasswordChange(c *gin.Context) {
	var req account.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	username := c.GetString("username")
	if err := app.accountSvc.PasswordChange(username, req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	respondSuccess(c, "密码修改成功", nil)
}

// accountMemberList 列出所有成员
func (app *App) accountMemberList(c *gin.Context) {
	respondSuccess(c, "ok", app.accountSvc.MemberList())
}

// accountMemberCreate 新建成员
func (app *App) accountMemberCreate(c *gin.Context) {
	var req account.MemberUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.accountSvc.MemberCreate(req); err != nil {
		switch {
		case errors.Is(err, account.ErrInvalidRequest), errors.Is(err, account.ErrMemberExists):
			respondError(c, http.StatusBadRequest, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondSuccess(c, "成员添加成功", nil)
}

// accountMemberUpdate 更新成员
func (app *App) accountMemberUpdate(c *gin.Context) {
	username := c.Param("username")
	var req account.MemberUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.accountSvc.MemberUpdate(username, req); err != nil {
		if errors.Is(err, account.ErrMemberNotFound) {
			respondError(c, http.StatusNotFound, err.Error())
		} else {
			respondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondSuccess(c, "成员更新成功", nil)
}

// accountMemberDelete 删除成员
func (app *App) accountMemberDelete(c *gin.Context) {
	username := c.Param("username")
	if err := app.accountSvc.MemberDelete(username); err != nil {
		switch {
		case errors.Is(err, account.ErrMemberNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondSuccess(c, "成员删除成功", nil)
}
