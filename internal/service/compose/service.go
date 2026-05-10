// Package compose 提供统一的 Compose 部署业务服务
package compose

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"regexp"
	"sync"

	"github.com/compose-spec/compose-go/v2/types"

	"isrvd/internal/registry"
	"isrvd/pkgs/compose"
	"isrvd/pkgs/docker"
	"isrvd/pkgs/swarm"
)

// ==================== 类型定义 ====================

// Target 部署目标
type Target string

const (
	TargetDocker Target = "docker"
	TargetSwarm  Target = "swarm"
)

// DeployRequest 部署请求
type DeployRequest struct {
	Content  string    `json:"content" binding:"required"`
	InitURL  string    `json:"initURL,omitempty"`
	InitFile io.Reader `json:"-"`
}

// RedeployRequest 重建请求
type RedeployRequest struct {
	Content string `json:"content" binding:"required"`
}

// ImageRedeployRequest 按服务更新镜像并重建请求
type ImageRedeployRequest struct {
	ServiceName string `json:"serviceName" binding:"required"`
	Image       string `json:"image" binding:"required"`
}

// DeployResult 部署结果
type DeployResult struct {
	Target      Target   `json:"target"`
	ProjectName string   `json:"projectName"`
	Items       []string `json:"items"`
	InstallDir  string   `json:"installDir,omitempty"`
}

// ==================== 服务定义 ====================

// Service Compose 部署业务服务
type Service struct {
	compose *compose.ComposeService
	docker  *docker.DockerService
	swarm   *swarm.SwarmService
}

var (
	instance   *Service
	instanceMu sync.Mutex
	safeName   = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)
)

// NewService 创建服务（单例，供 server 层调用）
// 依赖未就绪时返回错误，下次调用可重试，直到依赖初始化完成
func NewService() (*Service, error) {
	instanceMu.Lock()
	defer instanceMu.Unlock()

	if instance != nil {
		return instance, nil
	}

	d := registry.DockerService
	c := registry.ComposeService
	s := registry.SwarmService

	if d == nil {
		return nil, fmt.Errorf("docker 服务未初始化")
	}
	if c == nil {
		return nil, fmt.Errorf("compose 包服务未初始化")
	}

	instance = &Service{
		compose: c,
		docker:  d,
		swarm:   s,
	}
	return instance, nil
}

// ==================== 公共入口 ====================

// Deploy 统一部署入口
func (s *Service) Deploy(ctx context.Context, target Target, req DeployRequest) (*DeployResult, error) {
	switch target {
	case TargetDocker:
		return s.dockerDeploy(ctx, req)
	case TargetSwarm:
		return s.swarmDeploy(ctx, req)
	default:
		return nil, fmt.Errorf("不支持的目标: %s", target)
	}
}

// ContentInspect 获取 compose 文件内容
func (s *Service) ContentInspect(ctx context.Context, target Target, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("名称不能为空")
	}
	switch target {
	case TargetDocker:
		return s.dockerContentGet(ctx, name)
	case TargetSwarm:
		return s.swarmContentGet(ctx, name)
	default:
		return "", fmt.Errorf("不支持的目标: %s", target)
	}
}

// Redeploy 重建
func (s *Service) Redeploy(ctx context.Context, target Target, name string, req RedeployRequest) (*DeployResult, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("compose 内容不能为空")
	}
	switch target {
	case TargetDocker:
		return s.dockerRedeploy(ctx, name, req.Content)
	case TargetSwarm:
		return s.swarmRedeploy(ctx, name, req.Content)
	default:
		return nil, fmt.Errorf("不支持的目标: %s", target)
	}
}

// ImageRedeploy 按服务更新镜像并重建
func (s *Service) ImageRedeploy(ctx context.Context, target Target, name string, req ImageRedeployRequest) (*DeployResult, error) {
	if req.ServiceName == "" {
		return nil, fmt.Errorf("serviceName 不能为空")
	}
	if req.Image == "" {
		return nil, fmt.Errorf("image 不能为空")
	}
	switch target {
	case TargetDocker:
		return s.dockerImageRedeploy(ctx, name, req.ServiceName, req.Image)
	case TargetSwarm:
		return s.swarmImageRedeploy(ctx, name, req.ServiceName, req.Image)
	default:
		return nil, fmt.Errorf("不支持的目标: %s", target)
	}
}

func updateServiceImageContent(ctx context.Context, name, content, serviceName, image string) (string, error) {
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

// shortHash 返回内容的短 hash 字符串
func shortHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h[:4])
}
