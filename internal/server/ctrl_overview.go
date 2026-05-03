package server

import (
	"github.com/gin-gonic/gin"

	"isrvd/internal/helper"
)

// defineOverviewRoutes 定义 Overview 模块路由（系统统计 + 服务探测，无需权限）
func (app *App) defineOverviewRoutes() []Route {
	return []Route{
		{Method: "GET", Path: "/overview/status", Handler: app.overviewStat, Module: "overview", Label: "系统概览", Perm: "r"},
		{Method: "GET", Path: "/overview/probe", Handler: app.overviewProbe, Module: "overview", Label: "服务探测", Perm: "r"},
	}
}

func (app *App) overviewStat(c *gin.Context) {
	helper.RespondSuccess(c, "ok", app.overviewSvc.Stat(c.Request.Context()))
}

func (app *App) overviewProbe(c *gin.Context) {
	helper.RespondSuccess(c, "ok", app.overviewSvc.Probe(c.Request.Context()))
}
