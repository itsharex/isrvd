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

// SwarmDeploy 部署新的 Swarm Compose 项目。
func (s *Service) SwarmDeploy(ctx context.Context, req DeployRequest) (*DeployResult, error) {
	root := s.docker.ContainerRoot()
	if root == "" {
		return nil, fmt.Errorf("未配置容器数据根目录")
	}

	project, err := compose.LoadProjectFromContent(ctx, req.Content, "")
	if err != nil {
		return nil, err
	}
	projectName := project.Name
	if projectName == "" || projectName == "." {
		projectName = shortHash(req.Content)
	}
	if err := ValidateName(projectName); err != nil {
		return nil, err
	}

	installDir := filepath.Join(root, projectName)
	composeFile := filepath.Join(installDir, "compose.yml")
	if _, err := os.Stat(composeFile); err == nil {
		return nil, fmt.Errorf("目录 %s 已包含 compose 配置，请先移除", installDir)
	}

	_, err = os.Stat(installDir)
	installDirExists := err == nil

	deployed := false
	defer func() {
		if !deployed && !installDirExists {
			_ = os.RemoveAll(installDir)
		}
	}()

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return nil, fmt.Errorf("创建安装目录失败: %w", err)
	}
	if err := s.initFileHandle(installDir, req); err != nil {
		return nil, err
	}

	project, err = s.projectLoad(ctx, projectName, req.Content, installDir)
	if err != nil {
		return nil, err
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

	deployed = true
	logman.Info("Swarm compose deployed", "name", projectName, "dir", installDir)
	return &DeployResult{ProjectName: projectName, Items: items, InstallDir: installDir}, nil
}

// ==================== 获取内容 ====================

// SwarmContentGet 读取项目的 compose.yml；文件不存在时从运行态反推。
func (s *Service) SwarmContentGet(ctx context.Context, name string) (string, error) {
	if err := ValidateName(name); err != nil {
		return "", err
	}

	root := s.docker.ContainerRoot()
	if root == "" {
		return "", fmt.Errorf("未配置容器数据根目录")
	}

	path := filepath.Join(root, name, "compose.yml")
	if data, err := os.ReadFile(path); err == nil {
		return string(data), nil
	}

	raw, err := s.swarm.ServiceInspectRaw(ctx, name)
	if err != nil {
		return "", fmt.Errorf("compose 文件不存在且读取运行态失败: %w", err)
	}

	project, err := compose.ProjectFromSwarmInspect(raw, filepath.Join(root, name))
	if err != nil {
		return "", err
	}

	data, err := compose.ProjectToYAML(project)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ==================== 重建 ====================

// SwarmRedeploy 用新 compose 内容全量重建项目。
func (s *Service) SwarmRedeploy(ctx context.Context, name, content string) (*DeployResult, error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}

	root := s.docker.ContainerRoot()
	installDir := ""
	if root != "" {
		installDir = filepath.Join(root, name)
	}

	oldContent, _ := s.SwarmContentGet(ctx, name)

	s.swarmServicesRemove(ctx, name, oldContent)

	rollback := func() {
		s.swarmRollback(ctx, name, oldContent, installDir)
		s.contentSave(installDir, oldContent, "")
	}

	project, err := s.projectLoad(ctx, name, content, installDir)
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

	s.contentSave(installDir, content, oldContent)

	logman.Info("Swarm compose redeployed", "name", name)
	return &DeployResult{ProjectName: name, Items: items, InstallDir: installDir}, nil
}

// SwarmImageRedeploy 更新项目中指定服务的镜像并重建该服务。
func (s *Service) SwarmImageRedeploy(ctx context.Context, name, serviceName, image string) (*DeployResult, error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}

	root := s.docker.ContainerRoot()
	installDir := ""
	if root != "" {
		installDir = filepath.Join(root, name)
	}

	oldContent, err := s.SwarmContentGet(ctx, name)
	if err != nil {
		return nil, err
	}

	newContent, err := updateServiceImage(ctx, name, oldContent, serviceName, image)
	if err != nil {
		return nil, err
	}

	oldProject, err := s.projectParse(ctx, name, oldContent, installDir)
	if err != nil {
		return nil, err
	}
	newProject, err := s.projectParse(ctx, name, newContent, installDir)
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
		s.contentSave(installDir, oldContent, "")
		return nil, fmt.Errorf("删除旧服务 %s 失败: %w", oldSvc.Name, err)
	}

	id, err := s.swarmServiceCreate(ctx, newProject, newSvc)
	if err != nil {
		if _, rbErr := s.swarmServiceCreate(ctx, oldProject, oldSvc); rbErr != nil {
			logman.Warn("Swarm service rollback failed", "name", name, "service", serviceName, "error", rbErr)
		}
		s.contentSave(installDir, oldContent, "")
		return nil, err
	}

	s.contentSave(installDir, newContent, oldContent)

	item := fmt.Sprintf("%s (%s)", newSvc.Name, shortID(id))
	logman.Info("Swarm compose service image redeployed", "name", name, "service", serviceName, "image", image)
	return &DeployResult{ProjectName: name, Items: []string{item}, InstallDir: installDir}, nil
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

func (s *Service) swarmRollback(ctx context.Context, name, content, installDir string) {
	if content == "" {
		return
	}
	project, err := s.projectParse(ctx, name, content, installDir)
	if err != nil {
		logman.Warn("Rollback load project failed", "name", name, "error", err)
		return
	}
	if _, err := s.swarmProjectDeploy(ctx, project); err != nil {
		logman.Warn("Rollback deploy failed", "name", name, "error", err)
	}
}

// swarmEnsureNetworks 确保 project 中所有非 external 的网络以 overlay driver 存在
func (s *Service) swarmEnsureNetworks(ctx context.Context, project *types.Project) error {
	for key, netCfg := range project.Networks {
		if bool(netCfg.External) {
			continue
		}
		netName := netCfg.Name
		if netName == "" {
			netName = key
		}
		if _, err := s.docker.NetworkInspect(ctx, netName); err == nil {
			continue
		}
		driver := netCfg.Driver
		if driver == "" {
			driver = "overlay"
		}
		if _, err := s.docker.NetworkCreate(ctx, netName, driver, ""); err != nil {
			return fmt.Errorf("创建网络 %s 失败: %w", netName, err)
		}
		logman.Info("Swarm network created", "network", netName, "driver", driver)
	}
	return nil
}
