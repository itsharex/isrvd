package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"isrvd/config"
	"isrvd/internal/helper"
	svcCompose "isrvd/internal/service/compose"
)

// defineComposeRoutes 定义 Compose 模块路由
func (app *App) defineComposeRoutes() []Route {
	return []Route{
		// Docker Compose
		{Method: "GET", Path: "/compose/docker/:name", Handler: app.composeContentInspect, Module: "compose", Label: "读取 Docker Compose 文件"},
		{Method: "POST", Path: "/compose/docker/deploy", Handler: app.composeDockerDeploy, Module: "compose", Label: "部署 Docker Compose"},
		{Method: "POST", Path: "/compose/docker/:name/redeploy", Handler: app.composeRedeploy, Module: "compose", Label: "重部署 Docker Compose"},
		// Swarm Compose
		{Method: "GET", Path: "/compose/swarm/:name", Handler: app.composeContentInspect, Module: "compose", Label: "读取 Swarm Compose 文件"},
		{Method: "POST", Path: "/compose/swarm/deploy", Handler: app.composeSwarmDeploy, Module: "compose", Label: "部署 Swarm Compose"},
		{Method: "POST", Path: "/compose/swarm/:name/redeploy", Handler: app.composeRedeploy, Module: "compose", Label: "重部署 Swarm Compose"},
	}
}

// composeContentInspect 获取 compose 文件内容（Docker/Swarm 通用）
func (app *App) composeContentInspect(c *gin.Context) {
	target := parseComposeTarget(c)
	name := c.Param("name")

	content, err := app.composeSvc.ContentInspect(c.Request.Context(), target, name)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "获取 compose 文件成功", gin.H{"content": content})
}

// composeDockerDeploy Docker 部署（multipart form，支持文件上传）
func (app *App) composeDockerDeploy(c *gin.Context) {
	if c.Request.ContentLength > config.Server.MaxUploadSize {
		helper.RespondError(c, http.StatusBadRequest, "文件大小超过限制")
		return
	}

	req := svcCompose.DeployRequest{
		Content:     c.PostForm("content"),
		ProjectName: c.PostForm("projectName"),
		InitURL:     c.PostForm("initURL"),
	}
	if req.Content == "" || req.ProjectName == "" {
		helper.RespondError(c, http.StatusBadRequest, "content 和 projectName 不能为空")
		return
	}

	// 读取上传的 zip 文件（可选）
	if fh, err := c.FormFile("initFile"); err == nil {
		if fh.Size > config.Server.MaxUploadSize {
			helper.RespondError(c, http.StatusBadRequest, "文件大小超过限制")
			return
		}
		f, err := fh.Open()
		if err != nil {
			helper.RespondError(c, http.StatusBadRequest, "读取上传文件失败: "+err.Error())
			return
		}
		req.InitFile = f
		defer f.Close()
	}

	result, err := app.composeSvc.Deploy(c.Request.Context(), svcCompose.TargetDocker, req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "部署成功", result)
}

// composeSwarmDeploy Swarm 部署（JSON body）
func (app *App) composeSwarmDeploy(c *gin.Context) {
	var req svcCompose.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := app.composeSvc.Deploy(c.Request.Context(), svcCompose.TargetSwarm, req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "部署成功", result)
}

// composeRedeploy 重建（Docker/Swarm 通用）
func (app *App) composeRedeploy(c *gin.Context) {
	target := parseComposeTarget(c)
	name := c.Param("name")

	var req svcCompose.RedeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := app.composeSvc.Redeploy(c.Request.Context(), target, name, req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "重建成功", result)
}

// parseComposeTarget 从路由路径解析部署目标
func parseComposeTarget(c *gin.Context) svcCompose.Target {
	if strings.HasPrefix(c.FullPath(), "/api/compose/swarm") {
		return svcCompose.TargetSwarm
	}
	return svcCompose.TargetDocker
}
