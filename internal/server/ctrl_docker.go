package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"isrvd/internal/helper"
	svcDocker "isrvd/internal/service/docker"
	pkgdocker "isrvd/pkgs/docker"
)

// defineDockerRoutes 定义 Docker 模块路由
func (app *App) defineDockerRoutes() []Route {
	return []Route{
		// Docker 信息
		{Method: "GET", Path: "/docker/info", Handler: app.dockerInfo, Module: "docker", Label: "Docker", Perm: "r"},
		// 容器管理
		{Method: "GET", Path: "/docker/containers", Handler: app.dockerListContainers, Module: "docker", Label: "容器", Perm: "r"},
		{Method: "POST", Path: "/docker/containers", Handler: app.dockerCreateContainer, Module: "docker", Label: "容器", Perm: "rw"},
		{Method: "GET", Path: "/docker/containers/:id/stats", Handler: app.dockerContainerStats, Module: "docker", Label: "容器", Perm: "r"},
		{Method: "POST", Path: "/docker/containers/:id/action", Handler: app.dockerContainerAction, Module: "docker", Label: "容器", Perm: "rw"},
		{Method: "GET", Path: "/docker/containers/:id/logs", Handler: app.dockerContainerLogs, Module: "docker", Label: "容器", Perm: "r"},
		{Method: "GET", Path: "/docker/containers/:id/exec", Handler: app.dockerContainerExec, Module: "docker", Label: "容器", Perm: "rw"},
		// 镜像管理
		{Method: "GET", Path: "/docker/images", Handler: app.dockerListImages, Module: "docker", Label: "镜像", Perm: "r"},
		{Method: "POST", Path: "/docker/images/:id/action", Handler: app.dockerImageAction, Module: "docker", Label: "镜像", Perm: "rw"},
		{Method: "POST", Path: "/docker/images/tag", Handler: app.dockerTagImage, Module: "docker", Label: "镜像", Perm: "rw"},
		{Method: "GET", Path: "/docker/images/search/:term", Handler: app.dockerSearchImages, Module: "docker", Label: "镜像", Perm: "r"},
		{Method: "POST", Path: "/docker/images/build", Handler: app.dockerBuildImage, Module: "docker", Label: "镜像", Perm: "rw"},
		{Method: "GET", Path: "/docker/images/:id", Handler: app.dockerInspectImage, Module: "docker", Label: "镜像", Perm: "r"},
		// 网络管理
		{Method: "GET", Path: "/docker/networks", Handler: app.dockerListNetworks, Module: "docker", Label: "网络", Perm: "r"},
		{Method: "POST", Path: "/docker/networks/:id/action", Handler: app.dockerNetworkAction, Module: "docker", Label: "网络", Perm: "rw"},
		{Method: "POST", Path: "/docker/networks", Handler: app.dockerCreateNetwork, Module: "docker", Label: "网络", Perm: "rw"},
		{Method: "GET", Path: "/docker/networks/:id", Handler: app.dockerNetworkInspect, Module: "docker", Label: "网络", Perm: "r"},
		// 卷管理
		{Method: "GET", Path: "/docker/volumes", Handler: app.dockerListVolumes, Module: "docker", Label: "数据卷", Perm: "r"},
		{Method: "POST", Path: "/docker/volumes/:name/action", Handler: app.dockerVolumeAction, Module: "docker", Label: "数据卷", Perm: "rw"},
		{Method: "POST", Path: "/docker/volumes", Handler: app.dockerCreateVolume, Module: "docker", Label: "数据卷", Perm: "rw"},
		{Method: "GET", Path: "/docker/volumes/:name", Handler: app.dockerVolumeInspect, Module: "docker", Label: "数据卷", Perm: "r"},
		// 镜像仓库
		{Method: "GET", Path: "/docker/registries", Handler: app.dockerListRegistries, Module: "docker", Label: "镜像仓库", Perm: "r"},
		{Method: "POST", Path: "/docker/registries", Handler: app.dockerCreateRegistry, Module: "docker", Label: "镜像仓库", Perm: "rw"},
		{Method: "PUT", Path: "/docker/registries", Handler: app.dockerUpdateRegistry, Module: "docker", Label: "镜像仓库", Perm: "rw"},
		{Method: "DELETE", Path: "/docker/registries", Handler: app.dockerDeleteRegistry, Module: "docker", Label: "镜像仓库", Perm: "rw"},
		{Method: "POST", Path: "/docker/images/push", Handler: app.dockerPushImage, Module: "docker", Label: "镜像仓库", Perm: "rw"},
		{Method: "POST", Path: "/docker/images/pull", Handler: app.dockerPullFromRegistry, Module: "docker", Label: "镜像仓库", Perm: "rw"},
	}
}

// svcDockerRegistryUpsertRequest 是 service/docker 中 RegistryUpsertRequest 的本地别名
type svcDockerRegistryUpsertRequest = svcDocker.RegistryUpsertRequest

func (app *App) dockerInfo(c *gin.Context) {
	result, err := app.dockerSvc.Info(c.Request.Context())
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Docker info retrieved", result)
}

func (app *App) dockerListContainers(c *gin.Context) {
	all := c.DefaultQuery("all", "false") == "true"
	result, err := app.dockerSvc.ListContainers(c.Request.Context(), all)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Containers listed successfully", result)
}

func (app *App) dockerCreateContainer(c *gin.Context) {
	var req pkgdocker.ContainerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.dockerSvc.CreateContainer(c.Request.Context(), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "容器创建成功", result)
}

func (app *App) dockerContainerStats(c *gin.Context) {
	id := c.Param("id")
	result, err := app.dockerSvc.ContainerStats(c.Request.Context(), id)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Container stats retrieved", result)
}

func (app *App) dockerContainerAction(c *gin.Context) {
	req := pkgdocker.ContainerActionRequest{
		ID: c.Param("id"),
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.dockerSvc.ContainerAction(c.Request.Context(), req); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Container "+req.Action+" successfully", nil)
}

func (app *App) dockerContainerLogs(c *gin.Context) {
	req := pkgdocker.ContainerLogsRequest{
		ID:     c.Param("id"),
		Tail:   c.DefaultQuery("tail", "100"),
		Follow: c.DefaultQuery("follow", "false") == "true",
	}
	if req.ID == "" {
		helper.RespondError(c, http.StatusBadRequest, "container id is required")
		return
	}
	result, err := app.dockerSvc.ContainerLogs(c.Request.Context(), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Container logs retrieved", result)
}

func (app *App) dockerContainerExec(c *gin.Context) {
	conn, err := helper.WsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, "WebSocket 升级失败")
		return
	}
	defer conn.Close()

	containerID := c.Query("id")
	shell := c.DefaultQuery("shell", "/bin/sh")
	if containerID == "" {
		conn.WriteMessage(1, []byte("[错误: 缺少容器ID]\r\n"))
		return
	}
	app.dockerSvc.GetDockerService().ContainerExec(conn, containerID, shell)
}

// ─── 镜像 ───

func (app *App) dockerListImages(c *gin.Context) {
	all := c.DefaultQuery("all", "false") == "true"
	result, err := app.dockerSvc.ListImages(c.Request.Context(), all)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Images listed successfully", result)
}

func (app *App) dockerImageAction(c *gin.Context) {
	req := pkgdocker.ImageActionRequest{
		ID: c.Param("id"),
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.dockerSvc.ImageAction(c.Request.Context(), req); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Image "+req.Action+" successfully", nil)
}

func (app *App) dockerTagImage(c *gin.Context) {
	var req pkgdocker.ImageTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.dockerSvc.TagImage(c.Request.Context(), req); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "镜像打标签成功", nil)
}

func (app *App) dockerSearchImages(c *gin.Context) {
	term := c.Param("term")
	result, err := app.dockerSvc.SearchImages(c.Request.Context(), term)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Images searched successfully", result)
}

func (app *App) dockerBuildImage(c *gin.Context) {
	var req pkgdocker.ImageBuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.dockerSvc.BuildImage(c.Request.Context(), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "镜像构建成功", result)
}

func (app *App) dockerInspectImage(c *gin.Context) {
	id := c.Param("id")
	result, err := app.dockerSvc.InspectImage(c.Request.Context(), id)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Image inspected successfully", result)
}

// ─── 网络 ───

func (app *App) dockerListNetworks(c *gin.Context) {
	result, err := app.dockerSvc.ListNetworks(c.Request.Context())
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Networks listed successfully", result)
}

func (app *App) dockerNetworkAction(c *gin.Context) {
	req := pkgdocker.NetworkActionRequest{
		ID: c.Param("id"),
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.dockerSvc.NetworkAction(c.Request.Context(), req); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Network "+req.Action+" successfully", nil)
}

func (app *App) dockerCreateNetwork(c *gin.Context) {
	var req pkgdocker.NetworkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.dockerSvc.CreateNetwork(c.Request.Context(), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "网络创建成功", result)
}

func (app *App) dockerNetworkInspect(c *gin.Context) {
	id := c.Param("id")
	result, err := app.dockerSvc.NetworkInspect(c.Request.Context(), id)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Network inspected successfully", result)
}

// ─── 卷 ───

func (app *App) dockerListVolumes(c *gin.Context) {
	result, err := app.dockerSvc.ListVolumes(c.Request.Context())
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Volumes listed successfully", result)
}

func (app *App) dockerVolumeAction(c *gin.Context) {
	req := pkgdocker.VolumeActionRequest{
		Name: c.Param("name"),
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.dockerSvc.VolumeAction(c.Request.Context(), req); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Volume "+req.Action+" successfully", nil)
}

func (app *App) dockerCreateVolume(c *gin.Context) {
	var req pkgdocker.VolumeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.dockerSvc.CreateVolume(c.Request.Context(), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "卷创建成功", result)
}

func (app *App) dockerVolumeInspect(c *gin.Context) {
	name := c.Param("name")
	result, err := app.dockerSvc.VolumeInspect(c.Request.Context(), name)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "Volume inspected successfully", result)
}

// ─── 镜像仓库 ───

func (app *App) dockerListRegistries(c *gin.Context) {
	helper.RespondSuccess(c, "Registries listed successfully", app.dockerSvc.ListRegistries())
}

func (app *App) dockerCreateRegistry(c *gin.Context) {
	var req svcDockerRegistryUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.dockerSvc.CreateRegistry(req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	helper.RespondSuccess(c, "仓库添加成功", nil)
}

func (app *App) dockerUpdateRegistry(c *gin.Context) {
	originalURL := c.Query("url")
	var req svcDockerRegistryUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := app.dockerSvc.UpdateRegistry(originalURL, req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	helper.RespondSuccess(c, "仓库更新成功", nil)
}

func (app *App) dockerDeleteRegistry(c *gin.Context) {
	url := c.Query("url")
	if err := app.dockerSvc.DeleteRegistry(url); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	helper.RespondSuccess(c, "仓库删除成功", nil)
}

func (app *App) dockerPushImage(c *gin.Context) {
	var req pkgdocker.ImagePushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.dockerSvc.PushImage(c.Request.Context(), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "镜像推送成功", result)
}

func (app *App) dockerPullFromRegistry(c *gin.Context) {
	var req pkgdocker.ImagePullFromRegistryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := app.dockerSvc.PullFromRegistry(c.Request.Context(), req)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	helper.RespondSuccess(c, "镜像拉取成功", result)
}
