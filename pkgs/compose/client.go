// Package compose 提供基于 compose-spec 官方规范的 docker-compose 加载、
// 变量插值、校验与部署能力，完全拥抱 compose-go/v2 官方实现。
//
// 本包不负责底层容器/网络/卷的 CRUD，这些能力来自 pkgs/docker；
// compose 包只负责 compose 文件的"解析 → 转换 → 编排部署"。
package compose

import (
	"context"
	"fmt"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rehiy/libgo/logman"

	pkgdocker "isrvd/pkgs/docker"
)

// ComposeService compose 部署服务，封装对外能力入口
type ComposeService struct {
	docker *pkgdocker.DockerService
}

// NewComposeService 创建 compose 服务
// 依赖注入 docker 服务，用于复用容器/网络创建能力
func NewComposeService(d *pkgdocker.DockerService) (*ComposeService, error) {
	if d == nil {
		return nil, fmt.Errorf("docker service 未提供")
	}
	return &ComposeService{docker: d}, nil
}

// DeployContent 解析 compose yaml 并部署所有服务
func (s *ComposeService) DeployContent(ctx context.Context, content string) ([]string, error) {
	project, err := LoadProjectFromContent(ctx, content, "")
	if err != nil {
		return nil, err
	}
	return s.DeployProject(ctx, project)
}

// DeployProject 部署 compose project，失败时回滚已创建的容器
func (s *ComposeService) DeployProject(ctx context.Context, project *types.Project) ([]string, error) {
	if project == nil || len(project.Services) == 0 {
		return nil, fmt.Errorf("compose 项目为空或未定义服务")
	}

	// 先确保所有用到的外部网络存在
	for _, name := range collectNetworks(project) {
		if err := s.ensureNetwork(ctx, name); err != nil {
			return nil, fmt.Errorf("网络 %s 不存在，创建失败: %w", name, err)
		}
	}

	var createdIDs []string
	var containers []string

	rollback := func() {
		for _, id := range createdIDs {
			if err := s.docker.ContainerAction(ctx, id, "remove"); err != nil {
				logman.Warn("Rollback remove container failed", "id", pkgdocker.ShortID(id), "error", err)
			}
		}
	}

	for _, svc := range project.Services {
		id, name, err := s.ServiceContainerCreate(ctx, project, svc)
		if err != nil {
			rollback()
			return nil, err
		}

		createdIDs = append(createdIDs, id)
		containers = append(containers, fmt.Sprintf("%s (%s)", name, pkgdocker.ShortID(id)))

		logman.Info("Compose container deployed",
			"project", project.Name,
			"service", svc.Name,
			"container", name,
			"id", pkgdocker.ShortID(id),
		)
	}

	return containers, nil
}

// ServiceContainerCreate 根据 compose service 创建对应 Docker 容器
func (s *ComposeService) ServiceContainerCreate(ctx context.Context, project *types.Project, svc types.ServiceConfig) (string, string, error) {
	req, err := ServiceToDockerRequest(project, svc)
	if err != nil {
		return "", "", err
	}
	if err := s.docker.ImageEnsure(ctx, svc.Image); err != nil {
		return "", "", fmt.Errorf("镜像 %s 不存在，拉取失败: %w", svc.Image, err)
	}
	id, err := s.docker.ContainerCreate(ctx, req)
	if err != nil {
		return "", "", fmt.Errorf("创建容器 %s 失败: %w", req.Name, err)
	}
	return id, req.Name, nil
}

// ensureNetwork 确保 bridge 网络存在，不存在则创建
func (s *ComposeService) ensureNetwork(ctx context.Context, name string) error {
	if _, err := s.docker.NetworkInspect(ctx, name); err == nil {
		return nil
	}
	_, err := s.docker.NetworkCreate(ctx, name, "bridge")
	return err
}

// collectNetworks 收集 project 中所有需要确保存在的网络名（去重）
// 包含：自定义 network_mode + 各 service.networks 解析后的真实名称
func collectNetworks(project *types.Project) []string {
	set := map[string]struct{}{}
	for _, svc := range project.Services {
		if svc.NetworkMode != "" && !isBuiltinNetworkMode(svc.NetworkMode) {
			set[svc.NetworkMode] = struct{}{}
		}
		for k := range svc.Networks {
			set[resolveNetworkName(project, k)] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	return result
}
