package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"isrvd/config"
	"isrvd/internal/helper"
	svcSystem "isrvd/internal/service/system"
)

// defineSystemRoutes 定义 System 模块路由（系统设置 + 成员管理）
func (app *App) defineSystemRoutes() []Route {
	return []Route{
		// 系统信息
		{Method: "GET", Path: "/system/stat", Handler: app.systemStat, Module: "system", Label: "系统", Perm: "r"},
		{Method: "GET", Path: "/system/probe", Handler: app.systemProbe, Module: "system", Label: "系统", Perm: "r"},
		// 系统设置
		{Method: "GET", Path: "/system/settings", Handler: app.systemGetSettings, Module: "system", Label: "系统设置", Perm: "r"},
		{Method: "POST", Path: "/system/settings", Handler: app.systemUpdateSettings, Module: "system", Label: "系统设置", Perm: "rw"},
		// 审计日志
		{Method: "GET", Path: "/system/audit-logs", Handler: app.systemGetAuditLogs, Module: "system", Label: "系统", Perm: "r"},
	}
}

func (app *App) systemStat(c *gin.Context) {
	helper.RespondSuccess(c, "ok", app.systemSvc.Stat(c.Request.Context()))
}

func (app *App) systemProbe(c *gin.Context) {
	helper.RespondSuccess(c, "ok", app.systemSvc.Probe(c.Request.Context()))
}

func (app *App) systemGetSettings(c *gin.Context) {
	if c.Query("reload") == "true" {
		if err := config.Load(); err != nil {
			helper.RespondError(c, http.StatusInternalServerError, err.Error())
			return
		}
	}
	helper.RespondSuccess(c, "ok", app.settingsSvc.GetAll())
}

func (app *App) systemUpdateSettings(c *gin.Context) {
	var req svcSystem.UpdateAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.settingsSvc.UpdateAll(req); err != nil {
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

	logs := svcSystem.GetAuditLogs(username, limit)
	helper.RespondSuccess(c, "ok", logs)
}
