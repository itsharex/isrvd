package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"isrvd/internal/helper"
	svcCompose "isrvd/internal/service/compose"
)

// defineComposeRoutes 定义 Compose 模块路由
func (app *App) defineComposeRoutes() []Route {
	return []Route{
		// Docker Compose
		{Method: "GET", Path: "/compose/docker/:name", Handler: app.composeGetDockerContent, Module: "compose", Label: "读取 Docker Compose 文件"},
		{Method: "POST", Path: "/compose/docker/deploy", Handler: app.composeDeployDocker, Module: "compose", Label: "部署 Docker Compose"},
		{Method: "POST", Path: "/compose/docker/:name/redeploy", Handler: app.composeRedeployDocker, Module: "compose", Label: "重部署 Docker Compose"},
		// Swarm Compose
		{Method: "GET", Path: "/compose/swarm/:name", Handler: app.composeGetSwarmContent, Module: "compose", Label: "读取 Swarm Compose 文件"},
		{Method: "POST", Path: "/compose/swarm/deploy", Handler: app.composeDeploySwarm, Module: "compose", Label: "部署 Swarm Compose"},
		{Method: "POST", Path: "/compose/swarm/:name/redeploy", Handler: app.composeRedeploySwarm, Module: "compose", Label: "重部署 Swarm Compose"},
	}
}

func (app *App) composeGetDockerContent(c *gin.Context) {
	name := c.Param("name")
	content, err := app.composeSvc.GetContent(c.Request.Context(), svcCompose.TargetDocker, name)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "获取 compose 文件成功", gin.H{"content": content})
}

func (app *App) composeGetSwarmContent(c *gin.Context) {
	name := c.Param("name")
	content, err := app.composeSvc.GetContent(c.Request.Context(), svcCompose.TargetSwarm, name)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "获取 compose 文件成功", gin.H{"content": content})
}

func (app *App) composeDeployDocker(c *gin.Context) {
	req := svcCompose.DeployDockerRequest{
		Content:     c.PostForm("content"),
		ProjectName: c.PostForm("projectName"),
		InitURL:     c.PostForm("initURL"),
	}
	if req.Content == "" || req.ProjectName == "" {
		helper.RespondError(c, http.StatusBadRequest, "content 和 projectName 不能为空")
		return
	}

	// 读取上传的 zip 文件（可选），流式传入 service 避免内存双份拷贝
	if fh, err := c.FormFile("initFile"); err == nil {
		f, err := fh.Open()
		if err != nil {
			helper.RespondError(c, http.StatusBadRequest, "读取上传文件失败: "+err.Error())
			return
		}
		req.InitFile = f
		defer f.Close()
	}

	result, err := app.composeSvc.DeployDocker(c.Request.Context(), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "部署成功", result)
}

func (app *App) composeDeploySwarm(c *gin.Context) {
	var req svcCompose.DeploySwarmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.composeSvc.DeploySwarm(c.Request.Context(), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "部署成功", result)
}

func (app *App) composeRedeployDocker(c *gin.Context) {
	name := c.Param("name")
	var req svcCompose.RedeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.composeSvc.RedeployDocker(c.Request.Context(), name, req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "重建成功", result)
}

func (app *App) composeRedeploySwarm(c *gin.Context) {
	name := c.Param("name")
	var req svcCompose.RedeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.composeSvc.RedeploySwarm(c.Request.Context(), name, req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "重建成功", result)
}
