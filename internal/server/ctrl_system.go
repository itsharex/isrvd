package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"isrvd/config"
	"isrvd/internal/helper"
	svcSystem "isrvd/internal/service/system"
)

// defineSystemRoutes 定义 System 模块路由（系统配置 + 审计日志）
func (app *App) defineSystemRoutes() []Route {
	return []Route{
		// 系统配置
		{Method: "GET", Path: "/system/config", Handler: app.systemGetConfig, Module: "system", Label: "系统配置", Perm: "r"},
		{Method: "POST", Path: "/system/config", Handler: app.systemUpdateConfig, Module: "system", Label: "系统配置", Perm: "rw"},
		// 审计日志
		{Method: "GET", Path: "/system/audit/logs", Handler: app.systemGetAuditLogs, Module: "system", Label: "操作审计", Perm: "r"},
	}
}

func (app *App) systemGetConfig(c *gin.Context) {
	if c.Query("reload") == "true" {
		if err := config.Load(); err != nil {
			helper.RespondError(c, http.StatusInternalServerError, err.Error())
			return
		}
	}
	helper.RespondSuccess(c, "ok", app.configSvc.GetAll())
}

func (app *App) systemUpdateConfig(c *gin.Context) {
	var req svcSystem.UpdateAllConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.configSvc.UpdateAll(req); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "全部配置已保存，部分项需重启生效", nil)
}

// systemGetAuditLogs 获取操作审计日志
func (app *App) systemGetAuditLogs(c *gin.Context) {
	username := c.Query("username")
	limitStr := c.DefaultQuery("limit", "100")
	limit := 100

	// 解析 limit 参数
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	logs := app.auditSvc.GetLogs(username, limit)
	helper.RespondSuccess(c, "ok", logs)
}
