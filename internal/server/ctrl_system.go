package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"isrvd/config"
	"isrvd/internal/helper"
	svcAccount "isrvd/internal/service/account"
	svcSystem "isrvd/internal/service/system"
)

// defineSystemRoutes 定义 System 模块路由（系统配置 + 审计日志）
func (app *App) defineSystemRoutes() []Route {
	return []Route{
		// 系统配置
		{Method: "GET", Path: "/system/config", Handler: app.systemConfigInspect, Module: "system", Label: "获取系统配置", Access: svcAccount.AccessAuth},
		{Method: "PUT", Path: "/system/config", Handler: app.systemConfigUpdate, Module: "system", Label: "更新系统配置"},
		// 审计日志
		{Method: "GET", Path: "/system/audit/logs", Handler: app.systemAuditLogList, Module: "system", Label: "查询审计日志"},
	}
}

func (app *App) systemConfigInspect(c *gin.Context) {
	if c.Query("reload") == "true" {
		if err := config.Load(); err != nil {
			helper.RespondError(c, http.StatusInternalServerError, err.Error())
			return
		}
	}
	helper.RespondSuccess(c, "ok", app.configSvc.ConfigAll())
}

func (app *App) systemConfigUpdate(c *gin.Context) {
	var req svcSystem.UpdateAllConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.configSvc.ConfigUpdateAll(req); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "全部配置已保存，部分项需重启生效", nil)
}

// systemAuditLogList 获取操作审计日志
func (app *App) systemAuditLogList(c *gin.Context) {
	username := c.Query("username")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 {
		limit = l
	}
	logs := app.auditSvc.LogList(username, limit)
	helper.RespondSuccess(c, "ok", logs)
}
