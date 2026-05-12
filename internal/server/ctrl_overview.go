package server

import (
	"github.com/gin-gonic/gin"

	svcAccount "isrvd/internal/service/account"
)

// defineOverviewRoutes 定义 Overview 模块路由
func (app *App) defineOverviewRoutes() []Route {
	return []Route{
		{Method: "GET", Path: "/overview/probe", Handler: app.overviewProbe, Module: "overview", Label: "探测服务可用性", Access: svcAccount.AccessAuth},
		{Method: "GET", Path: "/overview/status", Handler: app.overviewStat, Module: "overview", Label: "获取系统概览统计"},
	}
}

func (app *App) overviewStat(c *gin.Context) {
	respondSuccess(c, "ok", app.overviewSvc.Stat(c.Request.Context()))
}

func (app *App) overviewProbe(c *gin.Context) {
	respondSuccess(c, "ok", app.overviewSvc.Probe(c.Request.Context()))
}
