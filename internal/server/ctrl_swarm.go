package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"isrvd/internal/helper"
	pkgswarm "isrvd/pkgs/swarm"
)

// defineSwarmRoutes 定义 Swarm 模块路由
func (app *App) defineSwarmRoutes() []Route {
	return []Route{
		// Swarm 信息
		{Method: "GET", Path: "/swarm/info", Handler: app.swarmInfo, Module: "swarm", Label: "获取 Swarm 信息"},
		// 节点管理
		{Method: "GET", Path: "/swarm/nodes", Handler: app.swarmListNodes, Module: "swarm", Label: "列出 Swarm 节点"},
		{Method: "GET", Path: "/swarm/node/:id", Handler: app.swarmInspectNode, Module: "swarm", Label: "查看 Swarm 节点"},
		{Method: "POST", Path: "/swarm/node/:id/action", Handler: app.NodeDTOAction, Module: "swarm", Label: "操作 Swarm 节点"},
		{Method: "GET", Path: "/swarm/tokens", Handler: app.swarmGetJoinTokens, Module: "swarm", Label: "获取 Swarm 加入令牌"},
		// 服务管理
		{Method: "GET", Path: "/swarm/services", Handler: app.swarmListServices, Module: "swarm", Label: "列出 Swarm 服务"},
		{Method: "GET", Path: "/swarm/service/:id", Handler: app.swarmInspectService, Module: "swarm", Label: "查看 Swarm 服务"},
		{Method: "POST", Path: "/swarm/service", Handler: app.swarmCreateService, Module: "swarm", Label: "创建 Swarm 服务"},
		{Method: "POST", Path: "/swarm/service/:id/action", Handler: app.swarmServiceAction, Module: "swarm", Label: "操作 Swarm 服务"},
		{Method: "POST", Path: "/swarm/service/:id/force-update", Handler: app.swarmForceUpdateService, Module: "swarm", Label: "强制更新 Swarm 服务"},
		{Method: "GET", Path: "/swarm/service/:id/logs", Handler: app.swarmServiceLogs, Module: "swarm", Label: "查看 Swarm 服务日志"},
		// 任务
		{Method: "GET", Path: "/swarm/tasks", Handler: app.swarmListTasks, Module: "swarm", Label: "列出 Swarm 任务"},
	}
}

func (app *App) swarmInfo(c *gin.Context) {
	result, err := app.swarmSvc.SwarmInfo(c.Request.Context())
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Swarm info retrieved", result)
}

func (app *App) swarmListNodes(c *gin.Context) {
	result, err := app.swarmSvc.ListNodes(c.Request.Context())
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Nodes listed", result)
}

func (app *App) swarmInspectNode(c *gin.Context) {
	id := c.Param("id")
	result, err := app.swarmSvc.InspectNode(c.Request.Context(), id)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Node inspected", result)
}

func (app *App) NodeDTOAction(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.swarmSvc.NodeAction(c.Request.Context(), c.Param("id"), req.Action); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Node updated", nil)
}

func (app *App) swarmListServices(c *gin.Context) {
	result, err := app.swarmSvc.ListServices(c.Request.Context())
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Services listed", result)
}

func (app *App) swarmInspectService(c *gin.Context) {
	id := c.Param("id")
	result, err := app.swarmSvc.InspectService(c.Request.Context(), id)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Service inspected", result)
}

func (app *App) swarmCreateService(c *gin.Context) {
	var req pkgswarm.ServiceSpec
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	id, err := app.swarmSvc.CreateService(c.Request.Context(), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Service created", gin.H{"id": id})
}

func (app *App) swarmServiceAction(c *gin.Context) {
	var req struct {
		Action   string  `json:"action"`
		Replicas *uint64 `json:"replicas,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.swarmSvc.ServiceAction(c.Request.Context(), c.Param("id"), req.Action, req.Replicas); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Service "+req.Action+" successfully", nil)
}

func (app *App) swarmForceUpdateService(c *gin.Context) {
	if err := app.swarmSvc.ForceUpdateService(c.Request.Context(), c.Param("id")); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Service force updated", nil)
}

func (app *App) swarmServiceLogs(c *gin.Context) {
	serviceID := c.Param("id")
	tail := c.DefaultQuery("tail", "100")
	logs, err := app.swarmSvc.GetServiceLogs(c.Request.Context(), serviceID, tail)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Logs retrieved", gin.H{"logs": logs})
}

func (app *App) swarmListTasks(c *gin.Context) {
	serviceID := c.Query("serviceID")
	result, err := app.swarmSvc.ListTasks(c.Request.Context(), serviceID)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Tasks listed", result)
}

func (app *App) swarmGetJoinTokens(c *gin.Context) {
	result, err := app.swarmSvc.GetJoinTokens(c.Request.Context())
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Join tokens retrieved", result)
}
