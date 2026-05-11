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
		{Method: "GET", Path: "/compose/:target/:name", Handler: app.composeContentInspect, Module: "compose", Label: "读取 Compose 文件"},
		{Method: "POST", Path: "/compose/:target/deploy", Handler: app.composeDeploy, Module: "compose", Label: "部署 Compose"},
		{Method: "POST", Path: "/compose/:target/:name/redeploy", Handler: app.composeRedeploy, Module: "compose", Label: "重部署 Compose"},
	}
}

// parseComposeTarget 从路径参数解析部署目标
func parseComposeTarget(c *gin.Context) (svcCompose.Target, bool) {
	switch c.Param("target") {
	case "docker":
		return svcCompose.TargetDocker, true
	case "swarm":
		return svcCompose.TargetSwarm, true
	default:
		respondError(c, http.StatusBadRequest, "不支持的部署目标: "+c.Param("target"))
		return "", false
	}
}

// composeContentInspect 获取 compose 文件内容（Docker/Swarm 通用）
func (app *App) composeContentInspect(c *gin.Context) {
	target, ok := parseComposeTarget(c)
	if !ok {
		return
	}
	content, err := app.composeSvc.ContentInspect(c.Request.Context(), target, c.Param("name"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "获取 compose 文件成功", gin.H{"content": content})
}

// composeDeploy 部署（Docker: multipart form 支持文件上传；Swarm: JSON body）
func (app *App) composeDeploy(c *gin.Context) {
	target, ok := parseComposeTarget(c)
	if !ok {
		return
	}

	var req svcCompose.DeployRequest

	if target == svcCompose.TargetDocker {
		if c.Request.ContentLength > config.Server.MaxUploadSize {
			respondError(c, http.StatusBadRequest, "文件大小超过限制")
			return
		}
		req.Content = c.PostForm("content")
		req.InitURL = c.PostForm("initURL")
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
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			respondError(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	result, err := app.composeSvc.Deploy(c.Request.Context(), target, req)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "部署成功", result)
}

// composeRedeploy 重建（Docker/Swarm 通用）
// body: { content } 全量重建，或 { serviceName, image } 按服务更新镜像重建
func (app *App) composeRedeploy(c *gin.Context) {
	target, ok := parseComposeTarget(c)
	if !ok {
		return
	}

	var req svcCompose.RedeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := app.composeSvc.Redeploy(c.Request.Context(), target, c.Param("name"), req)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, "重建成功", result)
}
