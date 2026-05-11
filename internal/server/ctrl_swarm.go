package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	
	pkgswarm "isrvd/pkgs/swarm"
)

// defineSwarmRoutes 定义 Swarm 模块路由
func (app *App) defineSwarmRoutes() []Route {
	return []Route{
		// Swarm 信息
		{Method: "GET", Path: "/swarm/info", Handler: app.swarmInfo, Module: "swarm", Label: "获取 Swarm 信息"},
		// 节点管理
		{Method: "GET", Path: "/swarm/nodes", Handler: app.swarmNodeList, Module: "swarm", Label: "列出 Swarm 节点"},
		{Method: "GET", Path: "/swarm/node/:id", Handler: app.swarmNodeInspect, Module: "swarm", Label: "查看 Swarm 节点"},
		{Method: "POST", Path: "/swarm/node/:id/action", Handler: app.swarmNodeAction, Module: "swarm", Label: "操作 Swarm 节点"},
		{Method: "GET", Path: "/swarm/token", Handler: app.swarmJoinToken, Module: "swarm", Label: "获取 Swarm 加入令牌"},
		// 服务管理
		{Method: "GET", Path: "/swarm/services", Handler: app.swarmServiceList, Module: "swarm", Label: "列出 Swarm 服务"},
		{Method: "GET", Path: "/swarm/service/:id", Handler: app.swarmServiceInspect, Module: "swarm", Label: "查看 Swarm 服务"},
		{Method: "POST", Path: "/swarm/service", Handler: app.swarmServiceCreate, Module: "swarm", Label: "创建 Swarm 服务"},
		{Method: "POST", Path: "/swarm/service/:id/action", Handler: app.swarmServiceAction, Module: "swarm", Label: "操作 Swarm 服务"},
		{Method: "POST", Path: "/swarm/service/:id/force-update", Handler: app.swarmServiceForceUpdate, Module: "swarm", Label: "强制更新 Swarm 服务"},
		{Method: "GET", Path: "/swarm/service/:id/logs", Handler: app.swarmServiceLogs, Module: "swarm", Label: "查看 Swarm 服务日志"},
		// 任务
		{Method: "GET", Path: "/swarm/tasks", Handler: app.swarmTaskList, Module: "swarm", Label: "列出 Swarm 任务"},
	}
}

func (app *App) swarmInfo(c *gin.Context) {
	result, err := app.swarmSvc.Info(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Swarm info retrieved", result)
}

func (app *App) swarmNodeList(c *gin.Context) {
	result, err := app.swarmSvc.NodeList(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Nodes listed", result)
}

func (app *App) swarmNodeInspect(c *gin.Context) {
	id := c.Param("id")
	result, err := app.swarmSvc.NodeInspect(c.Request.Context(), id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Node detail retrieved", result)
}

func (app *App) swarmNodeAction(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.swarmSvc.NodeAction(c.Request.Context(), c.Param("id"), req.Action); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Node updated", nil)
}

func (app *App) swarmServiceList(c *gin.Context) {
	result, err := app.swarmSvc.ServiceList(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Services listed", result)
}

func (app *App) swarmServiceInspect(c *gin.Context) {
	id := c.Param("id")
	result, err := app.swarmSvc.ServiceInspect(c.Request.Context(), id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Service detail retrieved", result)
}

func (app *App) swarmServiceCreate(c *gin.Context) {
	var req pkgswarm.ServiceSpec
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	id, err := app.swarmSvc.ServiceCreate(c.Request.Context(), req)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Service created", gin.H{"id": id})
}

func (app *App) swarmServiceAction(c *gin.Context) {
	var req struct {
		Action   string  `json:"action"`
		Replicas *uint64 `json:"replicas,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.swarmSvc.ServiceAction(c.Request.Context(), c.Param("id"), req.Action, req.Replicas); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Service "+req.Action+" successfully", nil)
}

func (app *App) swarmServiceForceUpdate(c *gin.Context) {
	if err := app.swarmSvc.ServiceForceUpdate(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Service force updated", nil)
}

func (app *App) swarmServiceLogs(c *gin.Context) {
	serviceID := c.Param("id")
	tail := c.DefaultQuery("tail", "100")
	logs, err := app.swarmSvc.ServiceLogs(c.Request.Context(), serviceID, tail)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Logs retrieved", gin.H{"logs": logs})
}

func (app *App) swarmTaskList(c *gin.Context) {
	serviceID := c.Query("serviceID")
	result, err := app.swarmSvc.TaskList(c.Request.Context(), serviceID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Tasks listed", result)
}

func (app *App) swarmJoinToken(c *gin.Context) {
	result, err := app.swarmSvc.JoinToken(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "Join tokens retrieved", result)
}
