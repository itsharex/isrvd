// Package system 系统配置查询与修改
package system

import (
	"encoding/json"
	"fmt"

	"isrvd/config"
)

// AllConfigResponse 全部配置聚合响应
type AllConfigResponse struct {
	Server      *config.Server            `json:"server"`
	Agent       *config.AgentConfig       `json:"agent"`
	Apisix      *config.ApisixConfig      `json:"apisix"`
	Docker      *config.DockerConfig      `json:"docker"`
	Marketplace *config.MarketplaceConfig `json:"marketplace"`
	Links       []*config.LinkConfig      `json:"links"`
}

// UpdateAllConfigRequest 全量更新请求
type UpdateAllConfigRequest struct {
	Server      *config.Server            `json:"server"`
	Agent       *config.AgentConfig       `json:"agent"`
	Apisix      *config.ApisixConfig      `json:"apisix"`
	Docker      *config.DockerConfig      `json:"docker"`
	Marketplace *config.MarketplaceConfig `json:"marketplace"`
	Links       []*config.LinkConfig      `json:"links"`
}

// ConfigService 系统配置业务服务
type ConfigService struct{}

// NewConfigService 创建系统配置业务服务
func NewConfigService() *ConfigService {
	return &ConfigService{}
}

// pickSecret 新值为空时保留原值，否则用新值
func pickSecret(newVal, oldVal string) string {
	if newVal == "" {
		return oldVal
	}
	return newVal
}

// GetAll 获取全部配置
func (s *ConfigService) GetAll() *AllConfigResponse {
	// 构造响应结构（敏感字段 json:"-" 会自动排除）
	resp := &AllConfigResponse{
		Server: &config.Server{
			Debug:           config.Debug,
			ListenAddr:      config.ListenAddr,
			JWTSecret:       config.JWTSecret,
			ProxyHeaderName: config.ProxyHeaderName,
			RootDirectory:   config.RootDirectory,
		},
		Agent:       config.Agent,
		Apisix:      config.Apisix,
		Docker:      config.Docker,
		Marketplace: config.Marketplace,
		Links:       config.Links,
	}

	// JSON 深拷贝（自动处理敏感字段过滤）
	data, err := json.Marshal(resp)
	if err != nil {
		return resp
	}

	var result AllConfigResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return resp
	}

	return &result
}

// UpdateAll 一次性更新全部配置（任何 nil 分区将跳过）
func (s *ConfigService) UpdateAll(req UpdateAllConfigRequest) error {
	if req.Server != nil {
		config.Debug = req.Server.Debug
		config.ListenAddr = req.Server.ListenAddr
		config.JWTSecret = pickSecret(req.Server.JWTSecret, config.JWTSecret)
		config.ProxyHeaderName = req.Server.ProxyHeaderName
		config.RootDirectory = req.Server.RootDirectory
	}
	if req.Agent != nil {
		config.Agent.Model = req.Agent.Model
		config.Agent.BaseURL = req.Agent.BaseURL
		config.Agent.APIKey = pickSecret(req.Agent.APIKey, config.Agent.APIKey)
	}
	if req.Apisix != nil {
		config.Apisix.AdminURL = req.Apisix.AdminURL
		config.Apisix.AdminKey = pickSecret(req.Apisix.AdminKey, config.Apisix.AdminKey)
	}
	if req.Docker != nil {
		config.Docker.Host = req.Docker.Host
		config.Docker.ContainerRoot = req.Docker.ContainerRoot
	}
	if req.Marketplace != nil {
		config.Marketplace.URL = req.Marketplace.URL
	}
	if req.Links != nil {
		config.Links = req.Links
	}
	if err := config.Save(); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	return nil
}