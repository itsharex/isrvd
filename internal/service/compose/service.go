// Package compose 提供统一的 Compose 部署业务服务
package compose

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"sync"

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
	Content     string    `json:"content" binding:"required"`
	ProjectName string    `json:"projectName" binding:"required"`
	InitURL     string    `json:"initURL,omitempty"`
	InitFile    io.Reader `json:"-"`
}

// RedeployRequest 重建请求
type RedeployRequest struct {
	Content string `json:"content" binding:"required"`
}

// DeployResult 部署结果
type DeployResult struct {
	Target     Target   `json:"target"`
	Items      []string `json:"items"`
	InstallDir string   `json:"installDir,omitempty"`
}

// ==================== 服务定义 ====================

// Service Compose 部署业务服务
type Service struct {
	compose *compose.ComposeService
	docker  *docker.DockerService
	swarm   *swarm.SwarmService
}

var (
	instance *Service
	once     sync.Once
	safeName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)
)

// NewService 创建服务（单例，供 server 层调用）
func NewService() (*Service, error) {
	var err error
	once.Do(func() {
		d := registry.DockerService
		c := registry.ComposeService
		s := registry.SwarmService

		if d == nil {
			err = fmt.Errorf("docker 服务未初始化")
			return
		}
		if c == nil {
			err = fmt.Errorf("compose 包服务未初始化")
			return
		}

		instance = &Service{
			compose: c,
			docker:  d,
			swarm:   s,
		}
	})

	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, fmt.Errorf("Compose 部署服务未初始化")
	}
	return instance, nil
}

// ==================== 公共入口 ====================

// Deploy 统一部署入口
func (s *Service) Deploy(ctx context.Context, target Target, req DeployRequest) (*DeployResult, error) {
	if !safeName.MatchString(req.ProjectName) {
		return nil, fmt.Errorf("非法的实例名")
	}
	switch target {
	case TargetDocker:
		return s.deployDocker(ctx, req)
	case TargetSwarm:
		return s.deploySwarm(ctx, req)
	default:
		return nil, fmt.Errorf("不支持的目标: %s", target)
	}
}

// GetContent 获取 compose 文件内容
func (s *Service) GetContent(ctx context.Context, target Target, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("名称不能为空")
	}
	if !safeName.MatchString(name) {
		return "", fmt.Errorf("非法的实例名")
	}
	switch target {
	case TargetDocker:
		return s.getDockerContent(ctx, name)
	case TargetSwarm:
		return s.getSwarmContent(ctx, name)
	default:
		return "", fmt.Errorf("不支持的目标: %s", target)
	}
}

// Redeploy 重建
func (s *Service) Redeploy(ctx context.Context, target Target, name string, req RedeployRequest) (*DeployResult, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("compose 内容不能为空")
	}
	if !safeName.MatchString(name) {
		return nil, fmt.Errorf("非法的实例名")
	}
	switch target {
	case TargetDocker:
		return s.redeployDocker(ctx, name, req.Content)
	case TargetSwarm:
		return s.redeploySwarm(ctx, name, req.Content)
	default:
		return nil, fmt.Errorf("不支持的目标: %s", target)
	}
}
