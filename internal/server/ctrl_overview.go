package server

import (
	"github.com/gin-gonic/gin"

	"isrvd/internal/helper"
	svcAccount "isrvd/internal/service/account"
)

// defineOverviewRoutes 定义 Overview 模块路由（系统统计 + 服务探测，无需权限）
func (app *App) defineOverviewRoutes() []Route {
	return []Route{
		{Method: "GET", Path: "/overview/status", Handler: app.overviewStat, Module: "overview", Label: "获取系统概览状态"},
		{Method: "GET", Path: "/overview/probe", Handler: app.overviewProbe, Module: "overview", Label: "探测服务可用性", Access: svcAccount.AccessAuth},
	}
}

func (app *App) overviewStat(c *gin.Context) {
	helper.RespondSuccess(c, "ok", app.overviewSvc.Stat(c.Request.Context()))
}

func (app *App) overviewProbe(c *gin.Context) {
	helper.RespondSuccess(c, "ok", app.overviewSvc.Probe(c.Request.Context()))
}
