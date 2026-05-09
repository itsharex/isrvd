package config

import (
	"path/filepath"
)

var (
	// 模式
	Debug = false
	// 监听地址
	ListenAddr = ":8080"
	// JWT 密钥（由启动脚本配置）
	JWTSecret = ""
	// JWT 过期时间（秒），0 表示使用默认值 24小时
	JWTExpiration int64 = 0
	// 文件上传最大大小（字节），0 表示使用默认值 100M
	MaxUploadSize int64 = 0
	// 内网代理用户名 Header 名（为空则不启用）
	ProxyHeaderName = ""
	// 基础目录
	RootDirectory = "."
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
		Server: &ServerConfig{
			Debug:           Debug,
			ListenAddr:      ListenAddr,
			JWTSecret:       JWTSecret,
			JWTExpiration:   JWTExpiration,
			MaxUploadSize:   MaxUploadSize,
			ProxyHeaderName: ProxyHeaderName,
			RootDirectory:   RootDirectory,
		},
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
	if conf.Server != nil {
		Debug = conf.Server.Debug
		ListenAddr = conf.Server.ListenAddr
		JWTSecret = conf.Server.JWTSecret
		JWTExpiration = conf.Server.JWTExpiration
		MaxUploadSize = conf.Server.MaxUploadSize
		ProxyHeaderName = conf.Server.ProxyHeaderName
		RootDirectory = conf.Server.RootDirectory
		// JWT 过期时间默认值
		if JWTExpiration == 0 {
			JWTExpiration = 86400
		}
		// 文件上传大小默认值（100MB）
		if MaxUploadSize == 0 {
			MaxUploadSize = 100 << 20
		}
		// 将 RootDirectory 转换为绝对路径
		if !filepath.IsAbs(RootDirectory) {
			abs, err := filepath.Abs(RootDirectory)
			if err == nil {
				RootDirectory = abs
			}
		}
	}

	if conf.Agent != nil {
		Agent = conf.Agent
	}

	if conf.Apisix != nil {
		Apisix = conf.Apisix
	}

	if conf.Docker != nil {
		Docker = conf.Docker
		if !filepath.IsAbs(Docker.ContainerRoot) {
			Docker.ContainerRoot = filepath.Join(RootDirectory, Docker.ContainerRoot)
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
			m.HomeDirectory = filepath.Join(RootDirectory, m.Username)
		} else if !filepath.IsAbs(m.HomeDirectory) {
			m.HomeDirectory = filepath.Join(RootDirectory, m.HomeDirectory)
		}
		Members[m.Username] = m
	}
}
