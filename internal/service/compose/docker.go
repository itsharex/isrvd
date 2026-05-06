package compose

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rehiy/pango/logman"
	"github.com/rehiy/pango/request"

	"isrvd/pkgs/archive"
	"isrvd/pkgs/compose"
)

// ==================== Docker 部署 ====================

func (s *Service) deployDocker(ctx context.Context, req DeployRequest) (*DeployResult, error) {
	root := s.docker.ContainerRoot()
	if root == "" {
		return nil, fmt.Errorf("未配置容器数据根目录")
	}

	installDir := filepath.Join(root, req.ProjectName)
	if _, err := os.Stat(installDir); err == nil {
		return nil, fmt.Errorf("目录已存在：%s，请先移除或使用其它实例名", installDir)
	}
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return nil, fmt.Errorf("创建安装目录失败: %w", err)
	}

	// 异常清理
	ok := false
	defer func() {
		if !ok {
			_ = os.RemoveAll(installDir)
		}
	}()

	// 处理附加运行文件
	if err := s.handleInitFile(installDir, req); err != nil {
		return nil, err
	}

	// 写入 compose 文件
	composeFile := filepath.Join(installDir, "compose.yml")
	if err := os.WriteFile(composeFile, []byte(req.Content), 0644); err != nil {
		return nil, fmt.Errorf("写入 compose 文件失败: %w", err)
	}

	// 加载并部署
	project, err := compose.LoadProject(ctx, compose.LoadOptions{
		WorkingDir:  installDir,
		ProjectName: req.ProjectName,
	})
	if err != nil {
		return nil, err
	}

	items, err := s.compose.DeployProject(ctx, project)
	if err != nil {
		return nil, err
	}

	ok = true
	logman.Info("Compose deployed", "name", req.ProjectName, "dir", installDir)
	return &DeployResult{Target: TargetDocker, Items: items, InstallDir: installDir}, nil
}

func (s *Service) handleInitFile(installDir string, req DeployRequest) error {
	zipPath := filepath.Join(installDir, "init.zip")

	if req.InitFile != nil {
		if err := writeAndUnzip(zipPath, req.InitFile); err != nil {
			return err
		}
		return nil
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

// ==================== Docker 获取内容 ====================

func (s *Service) getDockerContent(ctx context.Context, name string) (string, error) {
	root := s.docker.ContainerRoot()
	if root == "" {
		return "", fmt.Errorf("未配置容器数据根目录")
	}

	path := filepath.Join(root, name, "compose.yml")
	data, err := os.ReadFile(path)
	if err == nil {
		return string(data), nil
	}

	// 文件不存在，从运行态反推
	info, err := s.docker.ContainerInspect(ctx, name)
	if err != nil {
		return "", fmt.Errorf("compose 文件不存在且读取运行态失败: %w", err)
	}

	imageConfig, _ := s.docker.ImageConfig(ctx, info.Config.Image)
	project, err := compose.ProjectFromInspect(info, imageConfig)
	if err != nil {
		return "", err
	}

	data, err = compose.ProjectToYAML(project)
	if err != nil {
		return "", err
	}

	// 写入快照
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("写入 compose 文件失败: %w", err)
	}

	return string(data), nil
}

// ==================== Docker 重建 ====================

func (s *Service) redeployDocker(ctx context.Context, name, content string) (*DeployResult, error) {
	root := s.docker.ContainerRoot()
	installDir := ""
	if root != "" {
		installDir = filepath.Join(root, name)
	}

	// 停止并删除旧容器
	_ = s.docker.ContainerAction(ctx, name, "stop")
	_ = s.docker.ContainerAction(ctx, name, "remove")

	// 更新 compose 文件
	if installDir != "" {
		if err := os.MkdirAll(installDir, 0755); err != nil {
			return nil, fmt.Errorf("创建安装目录失败: %w", err)
		}
		composeFile := filepath.Join(installDir, "compose.yml")
		if err := os.WriteFile(composeFile, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("写入 compose 文件失败: %w", err)
		}
	}

	// 重新部署
	project, err := compose.LoadProjectFromContent(ctx, content, name)
	if err != nil {
		return nil, err
	}

	items, err := s.compose.DeployProject(ctx, project)
	if err != nil {
		return nil, err
	}

	logman.Info("Compose redeployed", "name", name)
	return &DeployResult{Target: TargetDocker, Items: items, InstallDir: installDir}, nil
}

// ==================== 工具函数 ====================

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
