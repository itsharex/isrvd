package compose

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rehiy/libgo/archive"
	"github.com/rehiy/libgo/logman"
	"github.com/rehiy/libgo/request"

	"isrvd/pkgs/compose"
)

// ==================== 部署 ====================

func (s *Service) DockerDeploy(ctx context.Context, req DeployRequest) (*DeployResult, error) {
	root := s.docker.ContainerRoot()
	if root == "" {
		return nil, fmt.Errorf("未配置容器数据根目录")
	}

	// 先从 compose 内容加载项目，获取项目名
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

	installDir := filepath.Join(root, projectName)
	composeFile := filepath.Join(installDir, "compose.yml")
	if _, err := os.Stat(composeFile); err == nil {
		return nil, fmt.Errorf("目录 %s 已包含 compose 配置，请先移除", installDir)
	}

	_, err = os.Stat(installDir)
	installDirExists := err == nil
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return nil, fmt.Errorf("创建安装目录失败: %w", err)
	}

	deployed := false
	defer func() {
		if !deployed && !installDirExists {
			_ = os.RemoveAll(installDir)
		}
	}()

	if err := s.dockerInitFileHandle(installDir, req); err != nil {
		return nil, err
	}

	if err := os.WriteFile(composeFile, []byte(req.Content), 0644); err != nil {
		return nil, fmt.Errorf("写入 compose 文件失败: %w", err)
	}

	// 重新加载项目，使用正确的 WorkingDir 以解析相对路径
	project, err = compose.LoadProject(ctx, compose.LoadOptions{
		WorkingDir: installDir,
	})
	if err != nil {
		return nil, err
	}
	if len(project.Services) == 0 {
		return nil, fmt.Errorf("compose 文件中没有定义服务")
	}

	for _, svc := range project.Services {
		cname := dockerContainerNameOf(svc)
		if _, err := s.docker.ContainerInspect(ctx, cname); err == nil {
			return nil, fmt.Errorf("容器 %s 已存在，请先移除", cname)
		}
	}

	items, err := s.compose.DeployProject(ctx, project)
	if err != nil {
		return nil, err
	}

	deployed = true
	logman.Info("Compose deployed", "name", projectName, "dir", installDir)
	return &DeployResult{ProjectName: projectName, Items: items, InstallDir: installDir}, nil
}

// dockerInitFileHandle 处理附加运行文件（支持 URL 下载或直接上传）
func (s *Service) dockerInitFileHandle(installDir string, req DeployRequest) error {
	zipPath := filepath.Join(installDir, "init.zip")

	if req.InitFile != nil {
		return writeAndUnzip(zipPath, req.InitFile)
	}

	if req.InitURL != "" {
		if _, err := request.Download(req.InitURL, zipPath, false); err != nil {
			return fmt.Errorf("下载附加文件失败: %w", err)
		}
		if err := archive.NewZipper().Unzip(zipPath); err != nil {
			return fmt.Errorf("解压附加文件失败: %w", err)
		}
		_ = os.Remove(zipPath)
	}
	return nil
}

// ==================== 获取内容 ====================

func (s *Service) DockerContentGet(ctx context.Context, name string) (string, error) {
	root := s.docker.ContainerRoot()
	if root == "" {
		return "", fmt.Errorf("未配置容器数据根目录")
	}

	path := filepath.Join(root, name, "compose.yml")
	if data, err := os.ReadFile(path); err == nil {
		return string(data), nil
	}

	// 文件不存在，从运行态反推
	info, err := s.docker.ContainerInspect(ctx, name)
	if err != nil {
		return "", fmt.Errorf("compose 文件不存在且读取运行态失败: %w", err)
	}

	imageConfig, _ := s.docker.ImageConfig(ctx, info.Config.Image)
	project, err := compose.ProjectFromDockerInspect(info, imageConfig)
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

func (s *Service) DockerRedeploy(ctx context.Context, name, content string) (*DeployResult, error) {
	root := s.docker.ContainerRoot()
	installDir := ""
	if root != "" {
		installDir = filepath.Join(root, name)
	}

	oldContent, _ := s.DockerContentGet(ctx, name)

	s.dockerContainersRemove(ctx, name, oldContent)

	rollback := func() {
		s.dockerRollback(ctx, name, oldContent)
		s.dockerContentSave(installDir, oldContent, "")
	}

	project, err := s.dockerProjectLoad(ctx, name, content, installDir)
	if err != nil {
		rollback()
		return nil, err
	}
	if len(project.Services) == 0 {
		rollback()
		return nil, fmt.Errorf("compose 文件中没有定义服务")
	}

	items, err := s.compose.DeployProject(ctx, project)
	if err != nil {
		rollback()
		return nil, err
	}

	s.dockerContentSave(installDir, content, oldContent)

	logman.Info("Compose redeployed", "name", name)
	return &DeployResult{Items: items, InstallDir: installDir}, nil
}

func (s *Service) DockerImageRedeploy(ctx context.Context, name, serviceName, image string) (*DeployResult, error) {
	root := s.docker.ContainerRoot()
	installDir := ""
	if root != "" {
		installDir = filepath.Join(root, name)
	}

	oldContent, err := s.DockerContentGet(ctx, name)
	if err != nil {
		return nil, err
	}
	newContent, newProject, err := updateServiceImage(ctx, name, oldContent, serviceName, image)
	if err != nil {
		return nil, err
	}

	oldProject, err := s.dockerProjectLoad(ctx, name, oldContent, installDir)
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

	oldContainerName := dockerContainerNameOf(oldSvc)
	_ = s.docker.ContainerAction(ctx, oldContainerName, "stop")
	if err := s.docker.ContainerAction(ctx, oldContainerName, "remove"); err != nil {
		s.dockerContentSave(installDir, oldContent, "")
		return nil, fmt.Errorf("删除旧容器 %s 失败: %w", oldContainerName, err)
	}

	id, _, err := s.compose.ServiceContainerCreate(ctx, newProject, newSvc)
	if err != nil {
		if _, _, rbErr := s.compose.ServiceContainerCreate(ctx, oldProject, oldSvc); rbErr != nil {
			logman.Warn("Docker service rollback failed", "name", name, "service", serviceName, "error", rbErr)
		}
		s.dockerContentSave(installDir, oldContent, "")
		return nil, err
	}

	s.dockerContentSave(installDir, newContent, oldContent)

	item := fmt.Sprintf("%s (%s)", dockerContainerNameOf(newSvc), shortID(id))
	logman.Info("Compose service image redeployed", "name", name, "service", serviceName, "image", image)
	return &DeployResult{ProjectName: name, Items: []string{item}, InstallDir: installDir}, nil
}

// ==================== 辅助函数 ====================

// dockerProjectLoad 写入 compose.yml 后用 LoadProject 加载，确保相对路径基于 installDir 展开
func (s *Service) dockerProjectLoad(ctx context.Context, name, content, installDir string) (*types.Project, error) {
	if installDir == "" {
		return compose.LoadProjectFromContent(ctx, content, name)
	}
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return nil, fmt.Errorf("创建安装目录失败: %w", err)
	}
	if err := os.WriteFile(filepath.Join(installDir, "compose.yml"), []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("写入 compose 文件失败: %w", err)
	}
	return compose.LoadProject(ctx, compose.LoadOptions{
		WorkingDir:  installDir,
		ProjectName: name,
	})
}

// dockerContainersRemove 停止并删除实例的所有容器
func (s *Service) dockerContainersRemove(ctx context.Context, name, content string) {
	if content == "" {
		content, _ = s.DockerContentGet(ctx, name)
	}
	if content == "" {
		return
	}
	project, err := compose.LoadProjectFromContent(ctx, content, name)
	if err != nil {
		return
	}
	for _, svc := range project.Services {
		cname := dockerContainerNameOf(svc)
		_ = s.docker.ContainerAction(ctx, cname, "stop")
		_ = s.docker.ContainerAction(ctx, cname, "remove")
	}
}

// dockerRollback 用指定配置内容重建容器（回滚用）
func (s *Service) dockerRollback(ctx context.Context, name, content string) {
	if content == "" {
		return
	}
	project, err := compose.LoadProjectFromContent(ctx, content, name)
	if err != nil {
		logman.Warn("Rollback load project failed", "name", name, "error", err)
		return
	}
	if _, err := s.compose.DeployProject(ctx, project); err != nil {
		logman.Warn("Rollback deploy failed", "name", name, "error", err)
	}
}

// dockerContentSave 持久化 compose.yml，bak 非空时同时写 .bak
func (s *Service) dockerContentSave(installDir, content, bak string) {
	if installDir == "" {
		return
	}
	_ = os.MkdirAll(installDir, 0755)
	if content != "" {
		_ = os.WriteFile(filepath.Join(installDir, "compose.yml"), []byte(content), 0644)
	}
	if bak != "" {
		_ = os.WriteFile(filepath.Join(installDir, "compose.yml.bak"), []byte(bak), 0644)
	}
}

// dockerContainerNameOf 返回 compose service 对应的容器名
func dockerContainerNameOf(svc types.ServiceConfig) string {
	if svc.ContainerName != "" {
		return svc.ContainerName
	}
	return svc.Name
}

// ==================== 工具函数 ====================

// writeAndUnzip 将 reader 内容写入 zip 文件并解压
func writeAndUnzip(zipPath string, r io.Reader) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("创建附加文件失败: %w", err)
	}
	defer f.Close()

	if _, err = io.Copy(f, r); err != nil {
		return fmt.Errorf("写入附加文件失败: %w", err)
	}

	if err := archive.NewZipper().Unzip(zipPath); err != nil {
		return fmt.Errorf("解压附加文件失败: %w", err)
	}
	_ = os.Remove(zipPath)
	return nil
}
