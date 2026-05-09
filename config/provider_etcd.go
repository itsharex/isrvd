package config

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	defaultEtcdScheme    = "http"
	defaultEtcdConfigKey = "/isrvd/config"
	defaultEtcdTimeout   = 5 * time.Second
)

// EtcdProvider etcd 配置提供者，value 使用 config.yml 同款 YAML 文本。
type EtcdProvider struct {
	client   *clientv3.Client
	key      string
	fallback string
	timeout  time.Duration
	mu       sync.Mutex
}

// NewEtcdProvider 创建 etcd 配置提供者。
// CONFIG_PATH 示例：etcd://user:pass@host1:2379,host2:2379/isrvd/config?scheme=http&timeout=5s
func NewEtcdProvider(path string) (*EtcdProvider, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	if u.Host == "" {
		return nil, fmt.Errorf("etcd 配置路径缺少 endpoints")
	}

	q := u.Query()
	timeout := defaultEtcdTimeout
	if raw := q.Get("timeout"); raw != "" {
		if timeout, err = time.ParseDuration(raw); err != nil {
			return nil, fmt.Errorf("etcd timeout 无效: %w", err)
		}
	}

	username := envOrDefault("ETCD_USERNAME", u.User.Username())
	password, _ := u.User.Password()
	password = envOrDefault("ETCD_PASSWORD", password)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints(u.Host, q.Get("scheme")),
		Username:    username,
		Password:    password,
		DialTimeout: timeout,
	})
	if err != nil {
		return nil, err
	}

	key := u.Path
	if key == "" || key == "/" {
		key = defaultEtcdConfigKey
	}
	return &EtcdProvider{client: cli, key: key, fallback: q.Get("fallback"), timeout: timeout}, nil
}

func (e *EtcdProvider) Type() string {
	return "etcd"
}

func (e *EtcdProvider) Load() (*Config, error) {
	ctx, cancel := e.context()
	defer cancel()

	resp, err := e.client.Get(ctx, e.key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return e.loadFallback()
	}

	conf := &Config{}
	if err := yaml.Unmarshal(resp.Kvs[0].Value, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func (e *EtcdProvider) loadFallback() (*Config, error) {
	if e.fallback == "" {
		return nil, fmt.Errorf("etcd 配置不存在: %s", e.key)
	}

	conf, err := NewYamlProvider(e.fallback).Load()
	if err != nil {
		return nil, fmt.Errorf("读取 fallback 配置失败: %w", err)
	}
	if err := e.Save(conf); err != nil {
		return nil, fmt.Errorf("写入 etcd fallback 配置失败: %w", err)
	}
	return conf, nil
}

func (e *EtcdProvider) Watch(ctx context.Context) (<-chan struct{}, <-chan error) {
	changes := make(chan struct{}, 1)
	errs := make(chan error, 1)

	go func() {
		defer close(changes)
		defer close(errs)

		for resp := range e.client.Watch(ctx, e.key) {
			if err := resp.Err(); err != nil {
				select {
				case errs <- err:
				default:
				}
				continue
			}
			for _, event := range resp.Events {
				switch event.Type {
				case clientv3.EventTypePut:
					var conf Config
					if err := yaml.Unmarshal(event.Kv.Value, &conf); err != nil {
						select {
						case errs <- fmt.Errorf("etcd 配置解析失败: %w", err):
						default:
						}
						continue
					}
					select {
					case changes <- struct{}{}:
					default:
					}
				case clientv3.EventTypeDelete:
					select {
					case errs <- fmt.Errorf("etcd 配置已删除: %s", e.key):
					default:
					}
				}
			}
		}
	}()

	return changes, errs
}

func (e *EtcdProvider) Save(conf *Config) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	ctx, cancel := e.context()
	defer cancel()
	_, err = e.client.Put(ctx, e.key, string(data))
	return err
}

func (e *EtcdProvider) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), e.timeout)
}

func etcdEndpoints(hosts, scheme string) []string {
	if scheme == "" {
		scheme = defaultEtcdScheme
	}
	var endpoints []string
	for _, host := range strings.Split(hosts, ",") {
		if host = strings.TrimSpace(host); host != "" {
			endpoints = append(endpoints, scheme+"://"+host)
		}
	}
	return endpoints
}
