package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"isrvd/internal/helper"
	pkgapisix "isrvd/pkgs/apisix"
)

// defineApisixRoutes 定义 Apisix 模块路由
func (app *App) defineApisixRoutes() []Route {
	return []Route{
		// Route 管理
		{Method: "GET", Path: "/apisix/routes", Handler: app.apisixListRoutes, Module: "apisix", Label: "列出 APISIX 路由"},
		{Method: "GET", Path: "/apisix/routes/:id", Handler: app.apisixGetRoute, Module: "apisix", Label: "查看 APISIX 路由"},
		{Method: "POST", Path: "/apisix/routes", Handler: app.apisixCreateRoute, Module: "apisix", Label: "创建 APISIX 路由"},
		{Method: "PUT", Path: "/apisix/routes/:id", Handler: app.apisixUpdateRoute, Module: "apisix", Label: "更新 APISIX 路由"},
		{Method: "PATCH", Path: "/apisix/routes/:id/status", Handler: app.apisixPatchRouteStatus, Module: "apisix", Label: "切换 APISIX 路由状态"},
		{Method: "DELETE", Path: "/apisix/routes/:id", Handler: app.apisixDeleteRoute, Module: "apisix", Label: "删除 APISIX 路由"},
		// Consumer 管理
		{Method: "GET", Path: "/apisix/consumers", Handler: app.apisixListConsumers, Module: "apisix", Label: "列出 APISIX 消费者"},
		{Method: "POST", Path: "/apisix/consumers", Handler: app.apisixCreateConsumer, Module: "apisix", Label: "创建 APISIX 消费者"},
		{Method: "PUT", Path: "/apisix/consumers/:username", Handler: app.apisixUpdateConsumer, Module: "apisix", Label: "更新 APISIX 消费者"},
		{Method: "DELETE", Path: "/apisix/consumers/:username", Handler: app.apisixDeleteConsumer, Module: "apisix", Label: "删除 APISIX 消费者"},
		// 白名单
		{Method: "GET", Path: "/apisix/whitelist", Handler: app.apisixGetWhitelist, Module: "apisix", Label: "查看 APISIX 白名单"},
		{Method: "POST", Path: "/apisix/whitelist/revoke", Handler: app.apisixRevokeWhitelist, Module: "apisix", Label: "撤销 APISIX 白名单"},
		// PluginConfig 管理
		{Method: "GET", Path: "/apisix/plugin-configs", Handler: app.apisixListPluginConfigs, Module: "apisix", Label: "列出 APISIX 插件配置"},
		{Method: "GET", Path: "/apisix/plugin-configs/:id", Handler: app.apisixGetPluginConfig, Module: "apisix", Label: "查看 APISIX 插件配置"},
		{Method: "POST", Path: "/apisix/plugin-configs", Handler: app.apisixCreatePluginConfig, Module: "apisix", Label: "创建 APISIX 插件配置"},
		{Method: "PUT", Path: "/apisix/plugin-configs/:id", Handler: app.apisixUpdatePluginConfig, Module: "apisix", Label: "更新 APISIX 插件配置"},
		{Method: "DELETE", Path: "/apisix/plugin-configs/:id", Handler: app.apisixDeletePluginConfig, Module: "apisix", Label: "删除 APISIX 插件配置"},
		// Upstream 管理
		{Method: "GET", Path: "/apisix/upstreams", Handler: app.apisixListUpstreams, Module: "apisix", Label: "列出 APISIX 上游"},
		{Method: "GET", Path: "/apisix/upstreams/:id", Handler: app.apisixGetUpstream, Module: "apisix", Label: "查看 APISIX 上游"},
		{Method: "POST", Path: "/apisix/upstreams", Handler: app.apisixCreateUpstream, Module: "apisix", Label: "创建 APISIX 上游"},
		{Method: "PUT", Path: "/apisix/upstreams/:id", Handler: app.apisixUpdateUpstream, Module: "apisix", Label: "更新 APISIX 上游"},
		{Method: "DELETE", Path: "/apisix/upstreams/:id", Handler: app.apisixDeleteUpstream, Module: "apisix", Label: "删除 APISIX 上游"},
		// SSL 管理
		{Method: "GET", Path: "/apisix/ssls", Handler: app.apisixListSSLs, Module: "apisix", Label: "列出 APISIX 证书"},
		{Method: "GET", Path: "/apisix/ssls/:id", Handler: app.apisixGetSSL, Module: "apisix", Label: "查看 APISIX 证书"},
		{Method: "POST", Path: "/apisix/ssls", Handler: app.apisixCreateSSL, Module: "apisix", Label: "创建 APISIX 证书"},
		{Method: "PUT", Path: "/apisix/ssls/:id", Handler: app.apisixUpdateSSL, Module: "apisix", Label: "更新 APISIX 证书"},
		{Method: "DELETE", Path: "/apisix/ssls/:id", Handler: app.apisixDeleteSSL, Module: "apisix", Label: "删除 APISIX 证书"},
		// 插件列表
		{Method: "GET", Path: "/apisix/plugins", Handler: app.apisixListPlugins, Module: "apisix", Label: "列出 APISIX 插件"},
	}
}

func (app *App) apisixListRoutes(c *gin.Context) {
	result, err := app.apisixSvc.ListRoutes()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixGetRoute(c *gin.Context) {
	result, err := app.apisixSvc.GetRoute(c.Param("id"))
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixCreateRoute(c *gin.Context) {
	var req pkgapisix.Route
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.CreateRoute(req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Route created successfully", result)
}

func (app *App) apisixUpdateRoute(c *gin.Context) {
	var req pkgapisix.Route
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.UpdateRoute(c.Param("id"), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Route updated successfully", result)
}

func (app *App) apisixPatchRouteStatus(c *gin.Context) {
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.apisixSvc.PatchRouteStatus(c.Param("id"), req.Status); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Route status updated successfully", nil)
}

func (app *App) apisixDeleteRoute(c *gin.Context) {
	if err := app.apisixSvc.DeleteRoute(c.Param("id")); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Route deleted successfully", nil)
}

func (app *App) apisixListConsumers(c *gin.Context) {
	result, err := app.apisixSvc.ListConsumers()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixCreateConsumer(c *gin.Context) {
	var req struct {
		Username string         `json:"username" binding:"required"`
		Desc     string         `json:"desc"`
		Plugins  map[string]any `json:"plugins"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.CreateConsumer(req.Username, req.Desc, req.Plugins)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Consumer created successfully", result)
}

func (app *App) apisixUpdateConsumer(c *gin.Context) {
	username := c.Param("username")
	var req struct {
		Desc    string         `json:"desc"`
		Plugins map[string]any `json:"plugins"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.apisixSvc.UpdateConsumer(username, req.Desc, req.Plugins); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Consumer updated successfully", gin.H{"username": username, "desc": req.Desc})
}

func (app *App) apisixDeleteConsumer(c *gin.Context) {
	if err := app.apisixSvc.DeleteConsumer(c.Param("username")); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Consumer deleted successfully", nil)
}

func (app *App) apisixGetWhitelist(c *gin.Context) {
	result, err := app.apisixSvc.GetWhitelist()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixRevokeWhitelist(c *gin.Context) {
	var req struct {
		RouteID      string `json:"route_id" binding:"required"`
		ConsumerName string `json:"consumer_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.apisixSvc.RevokeWhitelist(req.RouteID, req.ConsumerName); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Whitelist revoked successfully", nil)
}

func (app *App) apisixListPluginConfigs(c *gin.Context) {
	result, err := app.apisixSvc.ListPluginConfigs()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixGetPluginConfig(c *gin.Context) {
	result, err := app.apisixSvc.GetPluginConfig(c.Param("id"))
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixCreatePluginConfig(c *gin.Context) {
	var req pkgapisix.PluginConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.CreatePluginConfig(req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Plugin Config created successfully", result)
}

func (app *App) apisixUpdatePluginConfig(c *gin.Context) {
	var req pkgapisix.PluginConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.UpdatePluginConfig(c.Param("id"), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Plugin Config updated successfully", result)
}

func (app *App) apisixDeletePluginConfig(c *gin.Context) {
	if err := app.apisixSvc.DeletePluginConfig(c.Param("id")); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Plugin Config deleted successfully", nil)
}

func (app *App) apisixListUpstreams(c *gin.Context) {
	result, err := app.apisixSvc.ListUpstreams()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixGetUpstream(c *gin.Context) {
	result, err := app.apisixSvc.GetUpstream(c.Param("id"))
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixCreateUpstream(c *gin.Context) {
	var req pkgapisix.Upstream
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.CreateUpstream(req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Upstream created successfully", result)
}

func (app *App) apisixUpdateUpstream(c *gin.Context) {
	var req pkgapisix.Upstream
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.UpdateUpstream(c.Param("id"), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Upstream updated successfully", result)
}

func (app *App) apisixDeleteUpstream(c *gin.Context) {
	if err := app.apisixSvc.DeleteUpstream(c.Param("id")); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Upstream deleted successfully", nil)
}

func (app *App) apisixListSSLs(c *gin.Context) {
	result, err := app.apisixSvc.ListSSLs()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixGetSSL(c *gin.Context) {
	result, err := app.apisixSvc.GetSSL(c.Param("id"))
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixCreateSSL(c *gin.Context) {
	var req pkgapisix.SSL
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.CreateSSL(req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "SSL created successfully", result)
}

func (app *App) apisixUpdateSSL(c *gin.Context) {
	var req pkgapisix.SSL
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.UpdateSSL(c.Param("id"), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "SSL updated successfully", result)
}

func (app *App) apisixDeleteSSL(c *gin.Context) {
	if err := app.apisixSvc.DeleteSSL(c.Param("id")); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "SSL deleted successfully", nil)
}

func (app *App) apisixListPlugins(c *gin.Context) {
	result, err := app.apisixSvc.ListPlugins()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}
