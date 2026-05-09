package config

import "path/filepath"

const (
	defaultListenAddr    = ":8080"
	defaultJWTExpiration = 86400
	defaultMaxUploadSize = 100 << 20
	defaultRootDirectory = "."
)

var (
	// Server 服务器配置
	Server = &ServerConfig{
		ListenAddr:    defaultListenAddr,
		JWTExpiration: defaultJWTExpiration,
		MaxUploadSize: defaultMaxUploadSize,
		RootDirectory: defaultRootDirectory,
	}
	// Agent LLM 配置
	Agent = &AgentConfig{}
	// Apisix 配置
	Apisix = &ApisixConfig{}
	// Docker 配置
	Docker = &DockerConfig{}
	// 应用市场配置
	Marketplace = &MarketplaceConfig{}
	// 工具栏链接配置
	Links []*LinkConfig
	// 成员配置
	Members = map[string]*MemberConfig{}
	// 版本信息（编译时通过脚本注入）
	Version = "v0.0.0"
)

// Load 从配置提供者加载配置
func Load() error {
	conf, err := provider.Load()
	if err != nil {
		return err
	}

	Apply(conf)
	return nil
}

// Save 将当前全局配置保存到配置文件
func Save() error {
	members := make([]*MemberConfig, 0, len(Members))
	for _, m := range Members {
		members = append(members, m)
	}

	conf := &Config{
		Server:      Server,
		Agent:       Agent,
		Apisix:      Apisix,
		Docker:      Docker,
		Marketplace: Marketplace,
		Links:       Links,
		Members:     members,
	}

	return provider.Save(conf)
}

// Apply 应用配置到全局变量（不存储）
func Apply(conf *Config) {
	if conf == nil {
		return
	}

	Server = normalizeServerConfig(conf.Server)

	if conf.Agent != nil {
		Agent = conf.Agent
	}

	if conf.Apisix != nil {
		Apisix = conf.Apisix
	}

	if conf.Docker != nil {
		Docker = conf.Docker
		if !filepath.IsAbs(Docker.ContainerRoot) {
			Docker.ContainerRoot = filepath.Join(Server.RootDirectory, Docker.ContainerRoot)
		}
	}

	if conf.Marketplace != nil {
		Marketplace = conf.Marketplace
	}

	if conf.Links != nil {
		Links = conf.Links
	}

	Members = make(map[string]*MemberConfig, len(conf.Members))
	for _, m := range conf.Members {
		if m.HomeDirectory == "" {
			m.HomeDirectory = filepath.Join(Server.RootDirectory, m.Username)
		} else if !filepath.IsAbs(m.HomeDirectory) {
			m.HomeDirectory = filepath.Join(Server.RootDirectory, m.HomeDirectory)
		}
		Members[m.Username] = m
	}
}

func normalizeServerConfig(server *ServerConfig) *ServerConfig {
	if server == nil {
		server = &ServerConfig{}
	}
	if server.ListenAddr == "" {
		server.ListenAddr = defaultListenAddr
	}
	if server.JWTExpiration == 0 {
		server.JWTExpiration = defaultJWTExpiration
	}
	if server.MaxUploadSize == 0 {
		server.MaxUploadSize = defaultMaxUploadSize
	}
	if server.RootDirectory == "" {
		server.RootDirectory = defaultRootDirectory
	}
	if !filepath.IsAbs(server.RootDirectory) {
		if abs, err := filepath.Abs(server.RootDirectory); err == nil {
			server.RootDirectory = abs
		}
	}
	return server
}
