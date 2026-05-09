package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/rehiy/pango/logman"
)

// ConfigProvider 配置提供者接口
type ConfigProvider interface {
	Type() string
	Load() (*Config, error)
	Save(*Config) error
}

var provider ConfigProvider

func Init() error {
	if provider == nil {
		path := envOrDefault("CONFIG_PATH", "config.yml")
		switch {
		case strings.HasPrefix(strings.ToLower(path), "etcd://"):
			p, err := NewEtcdProvider(path)
			if err != nil {
				return err
			}
			provider = p
		case strings.HasPrefix(strings.ToLower(path), "file://"):
			provider = NewYamlProvider(strings.TrimPrefix(path, "file://"))
		case strings.Contains(path, "://"):
			return fmt.Errorf("不支持的配置路径: %s", path)
		default:
			provider = NewYamlProvider(path)
		}
	}

	logman.Info("load config", "provider", provider.Type())
	return Load()
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
