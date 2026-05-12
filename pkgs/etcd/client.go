// Package etcd 提供基于 etcd gRPC-gateway HTTP v3 API 的轻量客户端，
// 无 gRPC/protobuf 依赖，仅使用 Go 标准库。
package etcd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client etcd HTTP v3 客户端
type Client struct {
	endpoints  []string
	username   string
	password   string
	httpClient *http.Client
}

// Config 客户端配置
type Config struct {
	Endpoints   []string
	Username    string
	Password    string
	DialTimeout time.Duration
}

// New 创建 etcd 客户端
func New(cfg Config) *Client {
	timeout := cfg.DialTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Client{
		endpoints:  cfg.Endpoints,
		username:   cfg.Username,
		password:   cfg.Password,
		httpClient: &http.Client{}, // 不设置全局 Timeout，避免 watch 长连接被定时切断；超时由 ctx 控制
	}
}

// Get 读取 key 的值，key 不存在时返回空字符串
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	body, _ := json.Marshal(map[string]string{"key": b64(key)})
	raw, err := c.do(ctx, "/v3/kv/range", body)
	if err != nil {
		return "", err
	}

	var result struct {
		Kvs []struct {
			Value string `json:"value"`
		} `json:"kvs"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if len(result.Kvs) == 0 {
		return "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(result.Kvs[0].Value)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// Put 写入 key/value
func (c *Client) Put(ctx context.Context, key, value string) error {
	body, _ := json.Marshal(map[string]string{
		"key":   b64(key),
		"value": base64.StdEncoding.EncodeToString([]byte(value)),
	})
	_, err := c.do(ctx, "/v3/kv/put", body)
	return err
}

// WatchEvent watch 收到的事件
type WatchEvent struct {
	Type  string // "PUT" 或 "DELETE"
	Value string // PUT 时的新值（已解码）
}

// Watch 监听 key 变化，通过 events/errs channel 通知，ctx 取消时停止
func (c *Client) Watch(ctx context.Context, key string) (<-chan WatchEvent, <-chan error) {
	events := make(chan WatchEvent, 4)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		body, _ := json.Marshal(map[string]any{
			"key":             b64(key),
			"progress_notify": false,
		})

		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			c.endpoint()+"/v3/watch", bytes.NewReader(body))
		if err != nil {
			errs <- err
			return
		}
		c.setAuth(req)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			errs <- err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			raw, _ := io.ReadAll(resp.Body)
			errs <- fmt.Errorf("etcd watch 失败 %d: %s", resp.StatusCode, raw)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			line := scanner.Text()
			if line == "" {
				continue
			}
			var msg struct {
				Result struct {
					Events []struct {
						Type string `json:"type"`
						Kv   struct {
							Value string `json:"value"`
						} `json:"kv"`
					} `json:"events"`
				} `json:"result"`
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				continue
			}
			if msg.Error.Message != "" {
				select {
				case errs <- fmt.Errorf("etcd watch 错误: %s", msg.Error.Message):
				default:
				}
				continue
			}
			for _, ev := range msg.Result.Events {
				val, err := base64.StdEncoding.DecodeString(ev.Kv.Value)
				if err != nil && ev.Type == "PUT" {
					select {
					case errs <- fmt.Errorf("etcd watch value 解码失败: %w", err):
					default:
					}
					continue
				}
				select {
				case events <- WatchEvent{Type: ev.Type, Value: string(val)}:
				default:
				}
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			select {
			case errs <- err:
			default:
			}
		}
	}()

	return events, errs
}

func (c *Client) do(ctx context.Context, path string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.endpoint()+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("etcd %s 失败 %d: %s", path, resp.StatusCode, raw)
	}
	return raw, nil
}

func (c *Client) endpoint() string {
	if len(c.endpoints) > 0 {
		return c.endpoints[0]
	}
	return "http://localhost:2379"
}

func (c *Client) setAuth(req *http.Request) {
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}
}

func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
