package config

import (
	"path/filepath"

	"github.com/rehiy/pango/logman"

	"isrvd/internal/helper"
)

var (
	// 模式
	Debug = false
	// 监听地址
	ListenAddr = ":8080"
	// JWT 密钥
	JWTSecret = "jwt-secret-key"
	// 内网代理用户名 Header 名（为空则不启用）
	ProxyHeaderName = ""
	// 础目录
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

	applyConfig(conf)

	if err := migratePlaintextPasswords(); err != nil {
		logman.Warn("密码迁移失败", "error", err)
	}

	return nil
}

// Save 将当前全局配置保存到配置文件
func Save() error {
	members := make([]*MemberConfig, 0, len(Members))
	for _, m := range Members {
		members = append(members, m)
	}

	conf := &Config{
		Server: &Server{
			Debug:           Debug,
			ListenAddr:      ListenAddr,
			JWTSecret:       JWTSecret,
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

// applyConfig 应用配置到全局变量
func applyConfig(conf *Config) {
	if conf.Server != nil {
		Debug = conf.Server.Debug
		ListenAddr = conf.Server.ListenAddr
		JWTSecret = conf.Server.JWTSecret
		ProxyHeaderName = conf.Server.ProxyHeaderName
		RootDirectory = conf.Server.RootDirectory
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

// migratePlaintextPasswords 自动迁移明文密码为加密格式
func migratePlaintextPasswords() error {
	needSave := false

	for _, m := range Members {
		if m.Password == "" || helper.HashedBcrypt(m.Password) {
			continue
		}

		hashedPassword, err := helper.HashPassword(m.Password)
		if err != nil {
			logman.Warn("密码加密失败", "username", m.Username, "error", err)
			continue
		}

		logman.Info("密码已自动迁移为加密格式", "username", m.Username)
		m.Password = hashedPassword
		needSave = true
	}

	if needSave {
		if err := Save(); err != nil {
			return err
		}
		logman.Info("配置文件已自动更新（密码迁移）")
	}

	return nil
}
