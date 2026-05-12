package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"isrvd/config"
	svcCompose "isrvd/internal/service/compose"
)

// defineComposeRoutes 定义 Compose 模块路由
func (app *App) defineComposeRoutes() []Route {
	return []Route{
		// Docker Compose
		{Method: "GET", Path: "/compose/docker/:name", Handler: app.composeDockerInspect, Module: "compose", Label: "读取 Docker Compose 配置"},
		{Method: "GET", Path: "/compose/swarm/:name", Handler: app.composeSwarmInspect, Module: "compose", Label: "读取 Swarm Stack 配置"},
		{Method: "POST", Path: "/compose/docker/deploy", Handler: app.composeDockerDeploy, Module: "compose", Label: "部署 Docker Compose 应用"},
		// Swarm Compose
		{Method: "POST", Path: "/compose/swarm/deploy", Handler: app.composeSwarmDeploy, Module: "compose", Label: "部署 Swarm Stack 应用"},
		{Method: "POST", Path: "/compose/docker/:name/redeploy", Handler: app.composeDockerRedeploy, Module: "compose", Label: "重新部署 Docker Compose 应用"},
		{Method: "POST", Path: "/compose/swarm/:name/redeploy", Handler: app.composeSwarmRedeploy, Module: "compose", Label: "重新部署 Swarm Stack 应用"},
	}
}

func (app *App) composeDockerInspect(c *gin.Context) {
	content, err := app.composeSvc.DockerContentGet(c.Request.Context(), c.Param("name"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "获取 compose 文件成功", gin.H{"content": content})
}

func (app *App) composeSwarmInspect(c *gin.Context) {
	content, err := app.composeSvc.SwarmContentGet(c.Request.Context(), c.Param("name"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "获取 compose 文件成功", gin.H{"content": content})
}

func (app *App) composeDockerDeploy(c *gin.Context) {
	if c.Request.ContentLength > config.Server.MaxUploadSize {
		respondError(c, http.StatusBadRequest, "文件大小超过限制")
		return
	}
	req := svcCompose.DeployRequest{
		Content: c.PostForm("content"),
		InitURL: c.PostForm("initURL"),
	}
	if req.Content == "" {
		respondError(c, http.StatusBadRequest, "content 不能为空")
		return
	}
	if fh, err := c.FormFile("initFile"); err == nil {
		if fh.Size > config.Server.MaxUploadSize {
			respondError(c, http.StatusBadRequest, "文件大小超过限制")
			return
		}
		f, err := fh.Open()
		if err != nil {
			respondError(c, http.StatusBadRequest, "读取上传文件失败: "+err.Error())
			return
		}
		req.InitFile = f
		defer f.Close()
	}
	result, err := app.composeSvc.DockerDeploy(c.Request.Context(), req)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "部署成功", result)
}

func (app *App) composeSwarmDeploy(c *gin.Context) {
	var req svcCompose.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.composeSvc.SwarmDeploy(c.Request.Context(), req)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "部署成功", result)
}

func (app *App) composeDockerRedeploy(c *gin.Context) {
	var req svcCompose.RedeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	name := c.Param("name")
	var (
		result *svcCompose.DeployResult
		err    error
	)
	if req.ServiceName != "" {
		if req.Image == "" {
			respondError(c, http.StatusBadRequest, "image 不能为空")
			return
		}
		result, err = app.composeSvc.DockerImageRedeploy(c.Request.Context(), name, req.ServiceName, req.Image)
	} else {
		if req.Content == "" {
			respondError(c, http.StatusBadRequest, "content 不能为空")
			return
		}
		result, err = app.composeSvc.DockerRedeploy(c.Request.Context(), name, req.Content)
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "重建成功", result)
}

func (app *App) composeSwarmRedeploy(c *gin.Context) {
	var req svcCompose.RedeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	name := c.Param("name")
	var (
		result *svcCompose.DeployResult
		err    error
	)
	if req.ServiceName != "" {
		if req.Image == "" {
			respondError(c, http.StatusBadRequest, "image 不能为空")
			return
		}
		result, err = app.composeSvc.SwarmImageRedeploy(c.Request.Context(), name, req.ServiceName, req.Image)
	} else {
		if req.Content == "" {
			respondError(c, http.StatusBadRequest, "content 不能为空")
			return
		}
		result, err = app.composeSvc.SwarmRedeploy(c.Request.Context(), name, req.Content)
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "重建成功", result)
}
