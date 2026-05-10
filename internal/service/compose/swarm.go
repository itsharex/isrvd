package compose

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rehiy/libgo/logman"

	"isrvd/pkgs/compose"
	pkgdocker "isrvd/pkgs/docker"
)

// ==================== 部署 ====================

func (s *Service) swarmDeploy(ctx context.Context, req DeployRequest) (*DeployResult, error) {
	if s.swarm == nil {
		return nil, fmt.Errorf("swarm 服务未初始化")
	}

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

	return &DeployResult{Target: TargetSwarm, ProjectName: projectName, Items: items}, nil
}

// ==================== 获取内容 ====================

func (s *Service) swarmContentGet(ctx context.Context, name string) (string, error) {
	if s.swarm == nil {
		return "", fmt.Errorf("swarm 服务未初始化")
	}

	// 优先读持久化文件
	if content := s.swarmContentLoad(name); content != "" {
		return content, nil
	}

	// 文件不存在，从运行态反推（仅能反推单服务）
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

// ==================== 重建 ====================

func (s *Service) swarmRedeploy(ctx context.Context, name, content string) (*DeployResult, error) {
	if s.swarm == nil {
		return nil, fmt.Errorf("swarm 服务未初始化")
	}

	oldContent, _ := s.swarmContentGet(ctx, name)

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
	return &DeployResult{Target: TargetSwarm, Items: items}, nil
}

// ==================== 辅助函数 ====================

// swarmProjectDeploy 部署 compose project 中的所有服务，失败时回滚已创建的服务
func (s *Service) swarmProjectDeploy(ctx context.Context, project *types.Project) ([]string, error) {
	// 确保所有非 external 网络存在（swarm 服务需要 overlay 网络预先存在）
	if err := s.swarmEnsureNetworks(ctx, project); err != nil {
		return nil, err
	}

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
		req, err := compose.ServiceToSwarmRequest(project, svc)
		if err != nil {
			rollback()
			return nil, err
		}

		id, err := s.swarm.ServiceCreate(ctx, req)
		if err != nil {
			rollback()
			return nil, fmt.Errorf("创建服务 %s 失败: %w", req.Name, err)
		}

		createdIDs = append(createdIDs, id)
		items = append(items, fmt.Sprintf("%s (%s)", req.Name, pkgdocker.ShortID(id)))
		logman.Info("Swarm service deployed", "service", svc.Name, "id", pkgdocker.ShortID(id))
	}
	return items, nil
}

// swarmServicesRemove 删除实例的所有服务
func (s *Service) swarmServicesRemove(ctx context.Context, name, content string) {
	if s.swarm == nil {
		return
	}
	if content == "" {
		content, _ = s.swarmContentGet(ctx, name)
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
