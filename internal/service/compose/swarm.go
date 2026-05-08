package compose

import (
	"context"
	"fmt"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rehiy/pango/logman"

	pkgdocker "isrvd/pkgs/docker"
	"isrvd/pkgs/compose"
)

// ==================== Swarm 部署 ====================

func (s *Service) deploySwarm(ctx context.Context, req DeployRequest) (*DeployResult, error) {
	if s.swarm == nil {
		return nil, fmt.Errorf("swarm 服务未初始化")
	}

	// 检查服务是否已存在
	if _, err := s.swarm.ServiceInspect(ctx, req.ProjectName); err == nil {
		return nil, fmt.Errorf("实例 %s 已存在，请使用重建功能", req.ProjectName)
	}

	project, err := compose.LoadProjectFromContent(ctx, req.Content, req.ProjectName)
	if err != nil {
		return nil, err
	}
	if len(project.Services) == 0 {
		return nil, fmt.Errorf("compose 文件中没有定义服务")
	}

	items, err := s.deploySwarmProject(ctx, project)
	if err != nil {
		return nil, err
	}

	return &DeployResult{Target: TargetSwarm, Items: items}, nil
}

// ==================== Swarm 获取内容 ====================

func (s *Service) getSwarmContent(ctx context.Context, name string) (string, error) {
	if s.swarm == nil {
		return "", fmt.Errorf("swarm 服务未初始化")
	}

	info, err := s.swarm.ServiceInspect(ctx, name)
	if err != nil {
		return "", err
	}

	project, err := compose.ProjectFromSwarmInspect(info)
	if err != nil {
		return "", err
	}

	data, err := compose.ProjectToYAML(project)
	return string(data), err
}

// ==================== Swarm 重建 ====================

func (s *Service) redeploySwarm(ctx context.Context, name, content string) (*DeployResult, error) {
	if s.swarm == nil {
		return nil, fmt.Errorf("swarm 服务未初始化")
	}

	// 删除所有旧服务
	s.removeSwarmServices(ctx, name)

	project, err := compose.LoadProjectFromContent(ctx, content, name)
	if err != nil {
		return nil, err
	}
	if len(project.Services) == 0 {
		return nil, fmt.Errorf("compose 文件中没有定义服务")
	}

	items, err := s.deploySwarmProject(ctx, project)
	if err != nil {
		return nil, err
	}

	logman.Info("Swarm compose redeployed", "name", name)
	return &DeployResult{Target: TargetSwarm, Items: items}, nil
}

// removeSwarmServices 删除指定实例的所有服务
func (s *Service) removeSwarmServices(ctx context.Context, name string) {
	if s.swarm == nil {
		return
	}
	oldContent, err := s.getSwarmContent(ctx, name)
	if err != nil {
		return
	}
	oldProject, err := compose.LoadProjectFromContent(ctx, oldContent, name)
	if err != nil {
		return
	}
	for _, svc := range oldProject.Services {
		_ = s.swarm.ServiceAction(ctx, svc.Name, "remove", nil)
	}
}

// ==================== Swarm 辅助 ====================

func (s *Service) deploySwarmProject(ctx context.Context, project *types.Project) ([]string, error) {
	var items []string
	for _, svc := range project.Services {
		req, err := compose.ServiceToSwarmRequest(project, svc)
		if err != nil {
			return items, err
		}

		id, err := s.swarm.ServiceCreate(ctx, req)
		if err != nil {
			return items, fmt.Errorf("创建服务 %s 失败: %w", req.Name, err)
		}

		items = append(items, fmt.Sprintf("%s (%s)", req.Name, pkgdocker.ShortID(id)))
		logman.Info("Swarm service deployed", "service", svc.Name, "id", pkgdocker.ShortID(id))
	}
	return items, nil
}
