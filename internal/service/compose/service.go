// Package compose 提供统一的 Compose 部署业务服务
package compose

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rehiy/libgo/archive"
	"github.com/rehiy/libgo/request"

	"isrvd/internal/registry"
	"isrvd/pkgs/compose"
	"isrvd/pkgs/docker"
	"isrvd/pkgs/swarm"
)

// Service Compose 部署业务服务
type Service struct {
	compose *compose.ComposeService
	docker  *docker.DockerService
	swarm   *swarm.SwarmService
}

// DeployRequest 部署请求
type DeployRequest struct {
	Content  string    `json:"content" binding:"required"`
	InitURL  string    `json:"initURL,omitempty"`
	InitFile io.Reader `json:"-"`
}

// DeployResult 部署结果
type DeployResult struct {
	ProjectName string   `json:"projectName"`
	Items       []string `json:"items"`
	InstallDir  string   `json:"installDir,omitempty"`
}

// RedeployRequest 重建请求
// - ServiceName + Image 非空：从现有内容读取后更新指定服务镜像重建
// - 否则：Content 必须非空，全量重建
type RedeployRequest struct {
	Content     string `json:"content,omitempty"`
	ServiceName string `json:"serviceName,omitempty"`
	Image       string `json:"image,omitempty"`
}

var safeName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

// ValidateName 校验 Compose 项目名，防止路径逃逸。
func ValidateName(name string) error {
	if name == "" || !safeName.MatchString(name) {
		return fmt.Errorf("非法的项目名称: %s", name)
	}
	return nil
}

// NewService 创建 Compose 业务服务
func NewService() (*Service, error) {
	d := registry.DockerService
	c := registry.ComposeService
	if d == nil {
		return nil, fmt.Errorf("docker 服务未初始化")
	}
	if c == nil {
		return nil, fmt.Errorf("compose 包服务未初始化")
	}
	return &Service{compose: c, docker: d, swarm: registry.SwarmService}, nil
}

// ==================== 内部工具 ====================

// updateServiceImage 将 compose 内容中指定服务的镜像替换为 image，返回更新后的 YAML 文本。
// 返回内容中的相对路径保持原样，调用方需通过 projectParse 以 installDir 展开后再创建容器。
func updateServiceImage(ctx context.Context, name, content, serviceName, image string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("compose 内容不能为空")
	}
	project, err := compose.LoadProjectFromContent(ctx, content, name)
	if err != nil {
		return "", err
	}
	if len(project.Services) == 0 {
		return "", fmt.Errorf("compose 文件中没有定义服务")
	}

	matched := false
	for key, svc := range project.Services {
		if svc.Name == serviceName {
			svc.Image = image
			project.Services[key] = svc
			matched = true
			break
		}
	}
	if !matched {
		return "", fmt.Errorf("compose 服务 %s 不存在", serviceName)
	}

	data, err := compose.ProjectToYAML(project)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func projectServiceFind(project *types.Project, serviceName string) (types.ServiceConfig, error) {
	if project == nil {
		return types.ServiceConfig{}, fmt.Errorf("compose 项目为空")
	}
	for _, svc := range project.Services {
		if svc.Name == serviceName {
			return svc, nil
		}
	}
	return types.ServiceConfig{}, fmt.Errorf("compose 服务 %s 不存在", serviceName)
}

func shortHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h[:4])
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// projectLoad 写入 compose.yml 并以 installDir 为 WorkingDir 加载，确保相对路径正确展开。
func (s *Service) projectLoad(ctx context.Context, name, content, installDir string) (*types.Project, error) {
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

// projectParse 解析 compose 内容（不写文件），相对路径基于 installDir 展开。
func (s *Service) projectParse(ctx context.Context, name, content, installDir string) (*types.Project, error) {
	if installDir == "" {
		return compose.LoadProjectFromContent(ctx, content, name)
	}
	return compose.LoadProjectFromContentInDir(ctx, content, installDir, name)
}

// contentSave 持久化 compose.yml；bak 非空时同时写 .bak。
func (s *Service) contentSave(installDir, content, bak string) {
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

// initFileHandle 处理附加运行文件（支持本地上传或 URL 下载），解压到 installDir。
func (s *Service) initFileHandle(installDir string, req DeployRequest) error {
	if req.InitFile == nil && req.InitURL == "" {
		return nil
	}
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}
	zipPath := filepath.Join(installDir, "init.zip")

	if req.InitFile != nil {
		if closer, ok := req.InitFile.(io.Closer); ok {
			defer closer.Close()
		}
		return writeAndUnzip(zipPath, req.InitFile)
	}

	if _, err := request.Download(req.InitURL, zipPath, false); err != nil {
		return fmt.Errorf("下载附加文件失败: %w", err)
	}
	if err := archive.NewZipper().Unzip(zipPath); err != nil {
		return fmt.Errorf("解压附加文件失败: %w", err)
	}
	_ = os.Remove(zipPath)
	return nil
}

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
