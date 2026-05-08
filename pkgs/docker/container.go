package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/rehiy/pango/logman"
)

// VolumeMapping 挂载映射
type VolumeMapping struct {
	Type          string `json:"type,omitempty"`
	Source        string `json:"source,omitempty"`
	HostPath      string `json:"hostPath,omitempty"`
	ContainerPath string `json:"containerPath"`
	ReadOnly      bool   `json:"readOnly"`
}

// ContainerInfo Docker 容器信息
type ContainerInfo struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Image    string            `json:"image"`
	State    string            `json:"state"`
	Status   string            `json:"status"`
	Ports    []string          `json:"ports"`
	Networks []string          `json:"networks,omitempty"`
	Created  int64             `json:"created"`
	IsSwarm  bool              `json:"isSwarm,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
}

// ContainerList 获取容器列表
func (s *DockerService) ContainerList(ctx context.Context, all bool) ([]*ContainerInfo, error) {
	containers, err := s.client.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		logman.Error("List containers failed", "error", err)
		return nil, err
	}

	var result []*ContainerInfo
	for _, ct := range containers {
		name := ""
		if len(ct.Names) > 0 {
			name = strings.TrimPrefix(ct.Names[0], "/")
		}
		var networks []string
		if ct.NetworkSettings != nil {
			for netName := range ct.NetworkSettings.Networks {
				networks = append(networks, netName)
			}
		}
		result = append(result, &ContainerInfo{
			ID:       ShortID(ct.ID),
			Name:     name,
			Image:    ct.Image,
			State:    ct.State,
			Status:   ct.Status,
			Ports:    formatPorts(ct.Ports),
			Networks: networks,
			Created:  ct.Created,
			IsSwarm:  ct.Labels["com.docker.swarm.service.id"] != "",
			Labels:   ct.Labels,
		})
	}

	return result, nil
}

// ContainerInspect 获取容器详细配置（运行态快照依赖此接口）
func (s *DockerService) ContainerInspect(ctx context.Context, id string) (container.InspectResponse, error) {
	info, err := s.client.ContainerInspect(ctx, id)
	if err != nil {
		logman.Error("Inspect container failed", "id", id, "error", err)
		return container.InspectResponse{}, err
	}
	return info, nil
}

// ContainerActionRequest 容器操作请求
type ContainerActionRequest struct {
	ID     string `json:"id" binding:"required"`
	Action string `json:"action" binding:"required"`
}

// ContainerAction 容器操作（start/stop/restart/remove/pause/unpause）
func (s *DockerService) ContainerAction(ctx context.Context, id, action string) error {
	var err error
	switch action {
	case "start":
		err = s.client.ContainerStart(ctx, id, container.StartOptions{})
	case "stop":
		timeout := 10
		err = s.client.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout})
	case "restart":
		timeout := 10
		err = s.client.ContainerRestart(ctx, id, container.StopOptions{Timeout: &timeout})
	case "remove":
		err = s.client.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
	case "pause":
		err = s.client.ContainerPause(ctx, id)
	case "unpause":
		err = s.client.ContainerUnpause(ctx, id)
	default:
		return fmt.Errorf("不支持的操作: %s", action)
	}

	if err != nil {
		logman.Error("Container action failed", "action", action, "id", id, "error", err)
		return err
	}

	logman.Info("Container action performed", "action", action, "id", id)
	return nil
}

// ContainerLogs 获取容器日志
func (s *DockerService) ContainerLogs(ctx context.Context, id, tail string) ([]string, error) {
	if tail == "" {
		tail = "100"
	}

	options := container.LogsOptions{
		ShowStdout: true, ShowStderr: true,
		Tail: tail, Follow: false, Timestamps: true,
	}

	reader, err := s.client.ContainerLogs(ctx, id, options)
	if err != nil {
		logman.Error("Get container logs failed", "id", id, "error", err)
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		logman.Error("Read container logs failed", "id", id, "error", err)
		return nil, err
	}

	return ParseDockerLogs(data), nil
}

// ContainerLogsRequest 日志请求
type ContainerLogsRequest struct {
	ID     string `json:"id" binding:"required"`
	Tail   string `json:"tail"`
	Follow bool   `json:"follow"`
}

// ContainerCreateRequest 创建容器请求
type ContainerCreateRequest struct {
	Image      string            `json:"image" binding:"required"`
	Name       string            `json:"name" binding:"required"`
	Cmd        []string          `json:"cmd"`
	Env        []string          `json:"env"`
	Ports      map[string]string `json:"ports"`
	Volumes    []VolumeMapping   `json:"volumes"`
	Network    string            `json:"network"`
	Restart    string            `json:"restart"`
	Memory     int64             `json:"memory"`   // 内存限制，单位 MB
	Cpus       float64           `json:"cpus"`     // CPU 核数限制，如 0.5、2.0
	Workdir    string            `json:"workdir"`
	User       string            `json:"user"`
	Hostname   string            `json:"hostname"`
	Privileged bool              `json:"privileged"`
	CapAdd     []string          `json:"capAdd"`
	CapDrop    []string          `json:"capDrop"`
}

// ContainerCreate 创建容器
func (s *DockerService) ContainerCreate(ctx context.Context, req ContainerCreateRequest) (string, error) {
	containerConfig := &container.Config{
		Image:      req.Image,
		Cmd:        req.Cmd,
		Env:        req.Env,
		WorkingDir: req.Workdir,
		User:       req.User,
		Hostname:   req.Hostname,
	}

	hostConfig := &container.HostConfig{}

	// 处理重启策略
	switch req.Restart {
	case "always":
		hostConfig.RestartPolicy = container.RestartPolicy{Name: "always"}
	case "on-failure":
		hostConfig.RestartPolicy = container.RestartPolicy{Name: "on-failure"}
	case "unless-stopped":
		hostConfig.RestartPolicy = container.RestartPolicy{Name: "unless-stopped"}
	default:
		hostConfig.RestartPolicy = container.RestartPolicy{Name: "no"}
	}

	// 处理网络模式
	if req.Network != "" {
		hostConfig.NetworkMode = container.NetworkMode(req.Network)
	}

	// 处理资源限制
	if req.Memory > 0 {
		hostConfig.Memory = req.Memory * 1024 * 1024
	}
	if req.Cpus > 0 {
		hostConfig.NanoCPUs = int64(req.Cpus * 1e9)
	}

	// 处理端口映射，key 格式：hostPort 或 hostPort/proto（如 "8080" 或 "53/udp"）
	if len(req.Ports) > 0 {
		portBindings := make(nat.PortMap)
		exposedPorts := make(nat.PortSet)
		for hostPortSpec, containerPort := range req.Ports {
			hostPort := hostPortSpec
			proto := "tcp"
			if idx := strings.LastIndex(hostPortSpec, "/"); idx >= 0 {
				hostPort = hostPortSpec[:idx]
				proto = hostPortSpec[idx+1:]
			}
			// 防御性剥离 containerPort 中可能携带的协议后缀（如 "80/tcp"）
			if idx := strings.Index(containerPort, "/"); idx >= 0 {
				containerPort = containerPort[:idx]
			}
			port := nat.Port(containerPort + "/" + proto)
			portBindings[port] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: hostPort}}
			exposedPorts[port] = struct{}{}
		}
		hostConfig.PortBindings = portBindings
		containerConfig.ExposedPorts = exposedPorts
	}

	// 处理挂载映射
	if len(req.Volumes) > 0 {
		hostConfig.Mounts = make([]mount.Mount, 0, len(req.Volumes))
		for _, vol := range req.Volumes {
			m, err := s.buildMount(req.Name, vol)
			if err != nil {
				return "", err
			}
			hostConfig.Mounts = append(hostConfig.Mounts, m)
		}
	}

	// 处理安全配置
	if req.Privileged {
		hostConfig.Privileged = true
	}
	if len(req.CapAdd) > 0 {
		hostConfig.CapAdd = req.CapAdd
	}
	if len(req.CapDrop) > 0 {
		hostConfig.CapDrop = req.CapDrop
	}

	resp, err := s.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, req.Name)
	if err != nil {
		logman.Error("Create container failed", "name", req.Name, "error", err)
		return "", err
	}

	// 启动容器
	if err := s.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		logman.Error("Start container failed", "id", ShortID(resp.ID), "name", req.Name, "error", err)
		// 启动失败时删除已创建的容器，避免同名容器残留导致后续调用失败
		if rmErr := s.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}); rmErr != nil {
			logman.Warn("Remove container after start failure", "id", ShortID(resp.ID), "error", rmErr)
		}
		return "", fmt.Errorf("启动容器失败: %w", err)
	}

	logman.Info("Container created", "id", ShortID(resp.ID), "name", req.Name)

	return resp.ID, nil
}

// ContainerUpdateRequest 容器配置更新请求
type ContainerUpdateRequest struct {
	Name       string            `json:"name" binding:"required"`
	Image      string            `json:"image" binding:"required"`
	Cmd        []string          `json:"cmd"`
	Env        []string          `json:"env"`
	Ports      map[string]string `json:"ports"`
	Volumes    []VolumeMapping   `json:"volumes"`
	Network    string            `json:"network"`
	Restart    string            `json:"restart"`
	Memory     int64             `json:"memory"`   // 内存限制，单位 MB
	Cpus       float64           `json:"cpus"`     // CPU 核数限制，如 0.5、2.0
	Workdir    string            `json:"workdir"`
	User       string            `json:"user"`
	Hostname   string            `json:"hostname"`
	Privileged bool              `json:"privileged"`
	CapAdd     []string          `json:"capAdd"`
	CapDrop    []string          `json:"capDrop"`
}

// ToCreateRequest 将更新请求转换为创建请求，复用创建逻辑（供 UpdateContainer 和快照服务等共用）
func (req ContainerUpdateRequest) ToCreateRequest() ContainerCreateRequest {
	return ContainerCreateRequest{
		Image:      req.Image,
		Name:       req.Name,
		Cmd:        req.Cmd,
		Env:        req.Env,
		Ports:      req.Ports,
		Volumes:    req.Volumes,
		Network:    req.Network,
		Restart:    req.Restart,
		Memory:     req.Memory,
		Cpus:       req.Cpus,
		Workdir:    req.Workdir,
		User:       req.User,
		Hostname:   req.Hostname,
		Privileged: req.Privileged,
		CapAdd:     req.CapAdd,
		CapDrop:    req.CapDrop,
	}
}

// ContainerUpdate 更新容器配置并重建
func (s *DockerService) ContainerUpdate(ctx context.Context, req ContainerUpdateRequest) (string, error) {
	// 查找并停止旧容器
	containers, err := s.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return "", err
	}

	var oldContainerID string
	for _, ct := range containers {
		ctName := ""
		if len(ct.Names) > 0 {
			ctName = strings.TrimPrefix(ct.Names[0], "/")
		}
		if ctName == req.Name {
			oldContainerID = ct.ID
			break
		}
	}

	// 停止并删除旧容器
	if oldContainerID != "" {
		timeout := 10
		if err := s.client.ContainerStop(ctx, oldContainerID, container.StopOptions{Timeout: &timeout}); err != nil {
			logman.Warn("Stop old container failed", "id", ShortID(oldContainerID), "error", err)
		}
		if err := s.client.ContainerRemove(ctx, oldContainerID, container.RemoveOptions{Force: true}); err != nil {
			logman.Warn("Remove old container failed", "id", ShortID(oldContainerID), "error", err)
		}
	}

	return s.ContainerCreate(ctx, req.ToCreateRequest())
}

func (s *DockerService) buildMount(containerName string, vol VolumeMapping) (mount.Mount, error) {
	mountType := strings.ToLower(strings.TrimSpace(vol.Type))
	source := firstNonEmpty(vol.Source, vol.HostPath)
	if source == "" {
		return mount.Mount{}, fmt.Errorf("挂载源不能为空")
	}
	if vol.ContainerPath == "" {
		return mount.Mount{}, fmt.Errorf("挂载目标不能为空")
	}
	if mountType == "" {
		mountType = inferMountType(source)
	}

	switch mountType {
	case string(mount.TypeVolume):
		// volume 名不能含路径分隔符
		if strings.ContainsRune(source, '/') {
			return mount.Mount{}, fmt.Errorf("volume 名称不能包含路径分隔符，请使用 bind 类型或改用合法的 volume 名: %s", source)
		}
		return mount.Mount{
			Type:     mount.TypeVolume,
			Source:   source,
			Target:   vol.ContainerPath,
			ReadOnly: vol.ReadOnly,
		}, nil
	case string(mount.TypeBind):
		bindSource, err := s.resolveBindSource(containerName, source)
		if err != nil {
			return mount.Mount{}, err
		}
		return mount.Mount{
			Type:        mount.TypeBind,
			Source:      bindSource,
			Target:      vol.ContainerPath,
			ReadOnly:    vol.ReadOnly,
			BindOptions: &mount.BindOptions{CreateMountpoint: true},
		}, nil
	default:
		return mount.Mount{}, fmt.Errorf("不支持的挂载类型: %s", mountType)
	}
}

func (s *DockerService) resolveBindSource(containerName string, source string) (string, error) {
	bindSource := source
	if s.config.ContainerRoot != "" && !filepath.IsAbs(bindSource) {
		bindSource = filepath.Join(s.config.ContainerRoot, containerName, bindSource)
	}

	if _, err := os.Stat(bindSource); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("检查挂载源失败: %w", err)
	}
	return bindSource, nil
}

func inferMountType(source string) string {
	if filepath.IsAbs(source) || strings.HasPrefix(source, ".") {
		return string(mount.TypeBind)
	}
	return string(mount.TypeVolume)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
