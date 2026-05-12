// Package compose 提供统一的 Compose 部署业务服务
package compose

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"regexp"

	"github.com/compose-spec/compose-go/v2/types"

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

// updateServiceImage 将 compose 内容中指定服务的镜像替换为 image，
// 返回更新后的 YAML 文本和已修改的 project（避免调用方二次加载）
func updateServiceImage(ctx context.Context, name, content, serviceName, image string) (string, *types.Project, error) {
	if content == "" {
		return "", nil, fmt.Errorf("compose 内容不能为空")
	}
	project, err := compose.LoadProjectFromContent(ctx, content, name)
	if err != nil {
		return "", nil, err
	}
	if len(project.Services) == 0 {
		return "", nil, fmt.Errorf("compose 文件中没有定义服务")
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
		return "", nil, fmt.Errorf("compose 服务 %s 不存在", serviceName)
	}

	data, err := compose.ProjectToYAML(project)
	if err != nil {
		return "", nil, err
	}
	return string(data), project, nil
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

// shortHash 返回内容的短 hash 字符串
func shortHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h[:4])
}

// shortID 返回 ID 的前 12 字符
func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
