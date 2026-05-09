package config

import (
	"context"
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

type ConfigWatcher interface {
	Watch(context.Context) (<-chan struct{}, <-chan error)
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
	if err := Load(); err != nil {
		return err
	}

	watchConfigChanges()
	return nil
}

func watchConfigChanges() {
	watcher, ok := provider.(ConfigWatcher)
	if !ok {
		return
	}

	changes, errs := watcher.Watch(context.Background())
	go func() {
		for changes != nil || errs != nil {
			select {
			case _, ok := <-changes:
				if !ok {
					changes = nil
					continue
				}
				logman.Warn("Config changed", "provider", provider.Type(), "msg", "检测到配置变更，当前不会自动热更新，请重启服务使配置完整生效")
			case err, ok := <-errs:
				if !ok {
					errs = nil
					continue
				}
				logman.Warn("Config watch error", "provider", provider.Type(), "error", err)
			}
		}
	}()
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
