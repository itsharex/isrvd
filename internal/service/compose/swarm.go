package compose

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rehiy/libgo/logman"

	"isrvd/pkgs/compose"
)

// ==================== 部署 ====================

func (s *Service) SwarmDeploy(ctx context.Context, req DeployRequest) (*DeployResult, error) {
	// 从 compose 内容加载项目，获取项目名
	project, err := compose.LoadProjectFromContent(ctx, req.Content, "")
	if err != nil {
		return nil, err
	}
	projectName := project.Name
	if projectName == "" || projectName == "." {
		projectName = shortHash(req.Content)
	}
	if !safeName.MatchString(projectName) {
		return nil, fmt.Errorf("非法的项目名称: %s", projectName)
	}

	if len(project.Services) == 0 {
		return nil, fmt.Errorf("compose 文件中没有定义服务")
	}

	for _, svc := range project.Services {
		if _, err := s.swarm.ServiceInspect(ctx, svc.Name); err == nil {
			return nil, fmt.Errorf("服务 %s 已存在，请先移除", svc.Name)
		}
	}

	items, err := s.swarmProjectDeploy(ctx, project)
	if err != nil {
		return nil, err
	}

	s.swarmContentSave(projectName, req.Content, "")

	return &DeployResult{ProjectName: projectName, Items: items}, nil
}

// ==================== 获取内容 ====================

func (s *Service) SwarmContentGet(ctx context.Context, name string) (string, error) {
	// 优先读持久化文件
	if content := s.swarmContentLoad(name); content != "" {
		return content, nil
	}

	// 文件不存在，从运行态反推（仅能反推单服务）
	raw, err := s.swarm.ServiceInspectRaw(ctx, name)
	if err != nil {
		return "", err
	}

	project, err := compose.ProjectFromSwarmInspect(raw)
	if err != nil {
		return "", err
	}

	data, err := compose.ProjectToYAML(project)
	return string(data), err
}

// ==================== 重建 ====================

func (s *Service) SwarmRedeploy(ctx context.Context, name, content string) (*DeployResult, error) {
	oldContent, _ := s.SwarmContentGet(ctx, name)

	s.swarmServicesRemove(ctx, name, oldContent)

	rollback := func() {
		s.swarmRollback(ctx, name, oldContent)
		s.swarmContentSave(name, oldContent, "")
	}

	project, err := compose.LoadProjectFromContent(ctx, content, name)
	if err != nil {
		rollback()
		return nil, err
	}
	if len(project.Services) == 0 {
		rollback()
		return nil, fmt.Errorf("compose 文件中没有定义服务")
	}

	items, err := s.swarmProjectDeploy(ctx, project)
	if err != nil {
		rollback()
		return nil, err
	}

	s.swarmContentSave(name, content, oldContent)

	logman.Info("Swarm compose redeployed", "name", name)
	return &DeployResult{Items: items}, nil
}

func (s *Service) SwarmImageRedeploy(ctx context.Context, name, serviceName, image string) (*DeployResult, error) {
	oldContent, err := s.SwarmContentGet(ctx, name)
	if err != nil {
		return nil, err
	}
	newContent, newProject, err := updateServiceImage(ctx, name, oldContent, serviceName, image)
	if err != nil {
		return nil, err
	}

	oldProject, err := compose.LoadProjectFromContent(ctx, oldContent, name)
	if err != nil {
		return nil, err
	}
	oldSvc, err := projectServiceFind(oldProject, serviceName)
	if err != nil {
		return nil, err
	}
	newSvc, err := projectServiceFind(newProject, serviceName)
	if err != nil {
		return nil, err
	}

	if err := s.swarm.ServiceAction(ctx, oldSvc.Name, "remove", nil); err != nil {
		return nil, fmt.Errorf("删除旧服务 %s 失败: %w", oldSvc.Name, err)
	}

	id, err := s.swarmServiceCreate(ctx, newProject, newSvc)
	if err != nil {
		if _, rbErr := s.swarmServiceCreate(ctx, oldProject, oldSvc); rbErr != nil {
			logman.Warn("Swarm service rollback failed", "name", name, "service", serviceName, "error", rbErr)
		}
		s.swarmContentSave(name, oldContent, "")
		return nil, err
	}

	s.swarmContentSave(name, newContent, oldContent)

	item := fmt.Sprintf("%s (%s)", newSvc.Name, shortID(id))
	logman.Info("Swarm compose service image redeployed", "name", name, "service", serviceName, "image", image)
	return &DeployResult{ProjectName: name, Items: []string{item}}, nil
}

// ==================== 辅助函数 ====================

// swarmProjectDeploy 部署 compose project 中的所有服务，失败时回滚已创建的服务
func (s *Service) swarmProjectDeploy(ctx context.Context, project *types.Project) ([]string, error) {
	var createdIDs []string
	var items []string

	rollback := func() {
		for _, id := range createdIDs {
			if err := s.swarm.ServiceAction(ctx, id, "remove", nil); err != nil {
				logman.Warn("Rollback remove service failed", "id", id, "error", err)
			}
		}
	}

	for _, svc := range project.Services {
		id, err := s.swarmServiceCreate(ctx, project, svc)
		if err != nil {
			rollback()
			return nil, err
		}

		createdIDs = append(createdIDs, id)
		items = append(items, fmt.Sprintf("%s (%s)", svc.Name, shortID(id)))
		logman.Info("Swarm service deployed", "service", svc.Name, "id", shortID(id))
	}
	return items, nil
}

func (s *Service) swarmServiceCreate(ctx context.Context, project *types.Project, svc types.ServiceConfig) (string, error) {
	if err := s.swarmEnsureNetworks(ctx, project); err != nil {
		return "", err
	}
	req, err := compose.ServiceToSwarmRequest(project, svc)
	if err != nil {
		return "", err
	}
	id, err := s.swarm.ServiceCreate(ctx, req)
	if err != nil {
		return "", fmt.Errorf("创建服务 %s 失败: %w", req.Name, err)
	}
	return id, nil
}

// swarmServicesRemove 删除实例的所有服务
func (s *Service) swarmServicesRemove(ctx context.Context, name, content string) {
	if content == "" {
		content, _ = s.SwarmContentGet(ctx, name)
	}
	if content == "" {
		return
	}
	project, err := compose.LoadProjectFromContent(ctx, content, name)
	if err != nil {
		return
	}
	for _, svc := range project.Services {
		_ = s.swarm.ServiceAction(ctx, svc.Name, "remove", nil)
	}
}

// swarmRollback 用指定配置内容重建服务（回滚用）
func (s *Service) swarmRollback(ctx context.Context, name, content string) {
	if content == "" {
		return
	}
	project, err := compose.LoadProjectFromContent(ctx, content, name)
	if err != nil {
		logman.Warn("Rollback load project failed", "name", name, "error", err)
		return
	}
	if _, err := s.swarmProjectDeploy(ctx, project); err != nil {
		logman.Warn("Rollback deploy failed", "name", name, "error", err)
	}
}

// swarmContentSave 持久化 swarm compose 文件，bak 非空时同时写 .bak
func (s *Service) swarmContentSave(name, content, bak string) {
	root := s.docker.ContainerRoot()
	if root == "" {
		return
	}
	dir := filepath.Join(root, ".swarm", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	if content != "" {
		_ = os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(content), 0644)
	}
	if bak != "" {
		_ = os.WriteFile(filepath.Join(dir, "compose.yml.bak"), []byte(bak), 0644)
	}
}

// swarmContentLoad 读取持久化的 swarm compose 文件
func (s *Service) swarmContentLoad(name string) string {
	root := s.docker.ContainerRoot()
	if root == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(root, ".swarm", name, "compose.yml"))
	if err != nil {
		return ""
	}
	return string(data)
}

// swarmEnsureNetworks 确保 project 中所有非 external 的网络以 overlay driver 存在
// 网络已存在时跳过，创建失败时返回错误
func (s *Service) swarmEnsureNetworks(ctx context.Context, project *types.Project) error {
	for key, netCfg := range project.Networks {
		// external 网络由用户自己管理，跳过
		if bool(netCfg.External) {
			continue
		}
		netName := netCfg.Name
		if netName == "" {
			netName = key
		}
		// 已存在则跳过
		if _, err := s.docker.NetworkInspect(ctx, netName); err == nil {
			continue
		}
		driver := netCfg.Driver
		if driver == "" {
			driver = "overlay"
		}
		if _, err := s.docker.NetworkCreate(ctx, netName, driver); err != nil {
			return fmt.Errorf("创建网络 %s 失败: %w", netName, err)
		}
		logman.Info("Swarm network created", "network", netName, "driver", driver)
	}
	return nil
}
