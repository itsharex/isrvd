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
		{Method: "GET", Path: "/apisix/routes", Handler: app.apisixRouteList, Module: "apisix", Label: "列出 APISIX 路由"},
		{Method: "GET", Path: "/apisix/route/:id", Handler: app.apisixRouteInspect, Module: "apisix", Label: "查看 APISIX 路由"},
		{Method: "POST", Path: "/apisix/route", Handler: app.apisixRouteCreate, Module: "apisix", Label: "创建 APISIX 路由"},
		{Method: "PUT", Path: "/apisix/route/:id", Handler: app.apisixRouteUpdate, Module: "apisix", Label: "更新 APISIX 路由"},
		{Method: "PATCH", Path: "/apisix/route/:id/status", Handler: app.apisixRouteStatusPatch, Module: "apisix", Label: "切换 APISIX 路由状态"},
		{Method: "DELETE", Path: "/apisix/route/:id", Handler: app.apisixRouteDelete, Module: "apisix", Label: "删除 APISIX 路由"},
		// Consumer 管理
		{Method: "GET", Path: "/apisix/consumers", Handler: app.apisixConsumerList, Module: "apisix", Label: "列出 APISIX 消费者"},
		{Method: "POST", Path: "/apisix/consumer", Handler: app.apisixConsumerCreate, Module: "apisix", Label: "创建 APISIX 消费者"},
		{Method: "PUT", Path: "/apisix/consumer/:username", Handler: app.apisixConsumerUpdate, Module: "apisix", Label: "更新 APISIX 消费者"},
		{Method: "DELETE", Path: "/apisix/consumer/:username", Handler: app.apisixConsumerDelete, Module: "apisix", Label: "删除 APISIX 消费者"},
		// 白名单
		{Method: "GET", Path: "/apisix/whitelist", Handler: app.apisixWhitelistList, Module: "apisix", Label: "查看 APISIX 白名单"},
		{Method: "POST", Path: "/apisix/whitelist/revoke", Handler: app.apisixWhitelistRevoke, Module: "apisix", Label: "撤销 APISIX 白名单"},
		// PluginConfig 管理
		{Method: "GET", Path: "/apisix/plugin-configs", Handler: app.apisixPluginConfigList, Module: "apisix", Label: "列出 APISIX 插件配置"},
		{Method: "GET", Path: "/apisix/plugin-config/:id", Handler: app.apisixPluginConfigInspect, Module: "apisix", Label: "查看 APISIX 插件配置"},
		{Method: "POST", Path: "/apisix/plugin-config", Handler: app.apisixPluginConfigCreate, Module: "apisix", Label: "创建 APISIX 插件配置"},
		{Method: "PUT", Path: "/apisix/plugin-config/:id", Handler: app.apisixPluginConfigUpdate, Module: "apisix", Label: "更新 APISIX 插件配置"},
		{Method: "DELETE", Path: "/apisix/plugin-config/:id", Handler: app.apisixPluginConfigDelete, Module: "apisix", Label: "删除 APISIX 插件配置"},
		// Upstream 管理
		{Method: "GET", Path: "/apisix/upstreams", Handler: app.apisixUpstreamList, Module: "apisix", Label: "列出 APISIX 上游"},
		{Method: "GET", Path: "/apisix/upstream/:id", Handler: app.apisixUpstreamInspect, Module: "apisix", Label: "查看 APISIX 上游"},
		{Method: "POST", Path: "/apisix/upstream", Handler: app.apisixUpstreamCreate, Module: "apisix", Label: "创建 APISIX 上游"},
		{Method: "PUT", Path: "/apisix/upstream/:id", Handler: app.apisixUpstreamUpdate, Module: "apisix", Label: "更新 APISIX 上游"},
		{Method: "DELETE", Path: "/apisix/upstream/:id", Handler: app.apisixUpstreamDelete, Module: "apisix", Label: "删除 APISIX 上游"},
		// SSL 管理
		{Method: "GET", Path: "/apisix/ssls", Handler: app.apisixSSLList, Module: "apisix", Label: "列出 APISIX 证书"},
		{Method: "GET", Path: "/apisix/ssl/:id", Handler: app.apisixSSLInspect, Module: "apisix", Label: "查看 APISIX 证书"},
		{Method: "POST", Path: "/apisix/ssl", Handler: app.apisixSSLCreate, Module: "apisix", Label: "创建 APISIX 证书"},
		{Method: "PUT", Path: "/apisix/ssl/:id", Handler: app.apisixSSLUpdate, Module: "apisix", Label: "更新 APISIX 证书"},
		{Method: "DELETE", Path: "/apisix/ssl/:id", Handler: app.apisixSSLDelete, Module: "apisix", Label: "删除 APISIX 证书"},
		// 插件列表
		{Method: "GET", Path: "/apisix/plugins", Handler: app.apisixPluginList, Module: "apisix", Label: "列出 APISIX 插件"},
	}
}

func (app *App) apisixRouteList(c *gin.Context) {
	result, err := app.apisixSvc.RouteList()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixRouteInspect(c *gin.Context) {
	result, err := app.apisixSvc.RouteInspect(c.Param("id"))
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixRouteCreate(c *gin.Context) {
	var req pkgapisix.Route
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.RouteCreate(req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Route created successfully", result)
}

func (app *App) apisixRouteUpdate(c *gin.Context) {
	var req pkgapisix.Route
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.RouteUpdate(c.Param("id"), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Route updated successfully", result)
}

func (app *App) apisixRouteStatusPatch(c *gin.Context) {
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.apisixSvc.RouteStatusPatch(c.Param("id"), req.Status); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Route status updated successfully", nil)
}

func (app *App) apisixRouteDelete(c *gin.Context) {
	if err := app.apisixSvc.RouteDelete(c.Param("id")); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Route deleted successfully", nil)
}

func (app *App) apisixConsumerList(c *gin.Context) {
	result, err := app.apisixSvc.ConsumerList()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixConsumerCreate(c *gin.Context) {
	var req struct {
		Username string         `json:"username" binding:"required"`
		Desc     string         `json:"desc"`
		Plugins  map[string]any `json:"plugins"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.ConsumerCreate(req.Username, req.Desc, req.Plugins)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Consumer created successfully", result)
}

func (app *App) apisixConsumerUpdate(c *gin.Context) {
	username := c.Param("username")
	var req struct {
		Desc    string         `json:"desc"`
		Plugins map[string]any `json:"plugins"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.apisixSvc.ConsumerUpdate(username, req.Desc, req.Plugins); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Consumer updated successfully", gin.H{"username": username, "desc": req.Desc})
}

func (app *App) apisixConsumerDelete(c *gin.Context) {
	if err := app.apisixSvc.ConsumerDelete(c.Param("username")); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Consumer deleted successfully", nil)
}

func (app *App) apisixWhitelistList(c *gin.Context) {
	result, err := app.apisixSvc.WhitelistList()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixWhitelistRevoke(c *gin.Context) {
	var req struct {
		RouteID      string `json:"route_id" binding:"required"`
		ConsumerName string `json:"consumer_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.apisixSvc.WhitelistRevoke(req.RouteID, req.ConsumerName); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Whitelist revoked successfully", nil)
}

func (app *App) apisixPluginConfigList(c *gin.Context) {
	result, err := app.apisixSvc.PluginConfigList()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixPluginConfigInspect(c *gin.Context) {
	result, err := app.apisixSvc.PluginConfigInspect(c.Param("id"))
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixPluginConfigCreate(c *gin.Context) {
	var req pkgapisix.PluginConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.PluginConfigCreate(req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Plugin Config created successfully", result)
}

func (app *App) apisixPluginConfigUpdate(c *gin.Context) {
	var req pkgapisix.PluginConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.PluginConfigUpdate(c.Param("id"), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Plugin Config updated successfully", result)
}

func (app *App) apisixPluginConfigDelete(c *gin.Context) {
	if err := app.apisixSvc.PluginConfigDelete(c.Param("id")); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Plugin Config deleted successfully", nil)
}

func (app *App) apisixUpstreamList(c *gin.Context) {
	result, err := app.apisixSvc.UpstreamList()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixUpstreamInspect(c *gin.Context) {
	result, err := app.apisixSvc.UpstreamInspect(c.Param("id"))
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixUpstreamCreate(c *gin.Context) {
	var req pkgapisix.Upstream
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.UpstreamCreate(req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Upstream created successfully", result)
}

func (app *App) apisixUpstreamUpdate(c *gin.Context) {
	var req pkgapisix.Upstream
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.UpstreamUpdate(c.Param("id"), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Upstream updated successfully", result)
}

func (app *App) apisixUpstreamDelete(c *gin.Context) {
	if err := app.apisixSvc.UpstreamDelete(c.Param("id")); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Upstream deleted successfully", nil)
}

func (app *App) apisixSSLList(c *gin.Context) {
	result, err := app.apisixSvc.SSLList()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixSSLInspect(c *gin.Context) {
	result, err := app.apisixSvc.SSLInspect(c.Param("id"))
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}

func (app *App) apisixSSLCreate(c *gin.Context) {
	var req pkgapisix.SSL
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.SSLCreate(req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "SSL created successfully", result)
}

func (app *App) apisixSSLUpdate(c *gin.Context) {
	var req pkgapisix.SSL
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.apisixSvc.SSLUpdate(c.Param("id"), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "SSL updated successfully", result)
}

func (app *App) apisixSSLDelete(c *gin.Context) {
	if err := app.apisixSvc.SSLDelete(c.Param("id")); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "SSL deleted successfully", nil)
}

func (app *App) apisixPluginList(c *gin.Context) {
	result, err := app.apisixSvc.PluginList()
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "", result)
}
