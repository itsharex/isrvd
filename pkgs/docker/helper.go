package docker

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
)

// ActionRequest 资源操作请求（容器/镜像/网络/卷通用）。
// ID 字段对所有资源使用，卷场景下 ID 即卷名。
type ActionRequest struct {
	ID     string `json:"id" binding:"required"`
	Action string `json:"action" binding:"required"`
}

// ShortID 返回 ID 的前 12 字符，不足 12 则返回原值
func ShortID(id string) string {
	id = strings.TrimPrefix(id, "sha256:")
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// ParseDockerLogs 解析 Docker multiplexed stream 格式的日志数据
// 移除每帧前 8 字节的头部，返回纯文本行列表
func ParseDockerLogs(data []byte) []string {
	var logs []string
	for i := 0; i < len(data); {
		if i+8 > len(data) {
			break
		}
		size := int(data[i+4])<<24 | int(data[i+5])<<16 | int(data[i+6])<<8 | int(data[i+7])
		i += 8
		if i+size > len(data) || size <= 0 {
			break
		}
		logs = append(logs, string(data[i:i+size]))
		i += size
	}
	return logs
}

// formatPorts 格式化端口列表：IPv4 优先、去重、通配地址省略 IP
//
// 算法：单次遍历，用 seen map 记录已输出的 key；
// IPv6 条目若已有对应 IPv4 条目则跳过，否则正常输出。
func formatPorts(ports []container.Port) []string {
	seen := make(map[string]bool, len(ports))
	result := make([]string, 0, len(ports))

	for _, p := range ports {
		var entry, key string
		isIPv6 := strings.Contains(p.IP, ":")

		if p.PublicPort > 0 {
			key = fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, p.Type)
			// IPv6 且已有 IPv4 同 key，跳过
			if isIPv6 && seen[key] {
				continue
			}
			if p.IP == "" || p.IP == "0.0.0.0" || p.IP == "::" {
				entry = fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, p.Type)
			} else {
				entry = fmt.Sprintf("%s:%d:%d/%s", p.IP, p.PublicPort, p.PrivatePort, p.Type)
			}
		} else {
			key = fmt.Sprintf("%d/%s", p.PrivatePort, p.Type)
			entry = key
		}

		if !seen[key] {
			seen[key] = true
			result = append(result, entry)
		}
	}
	return result
}

// buildDockerfileTar 构建 Dockerfile 的 tar 包
func buildDockerfileTar(dockerfile string) (*bytes.Buffer, error) {
	tarBuf := new(bytes.Buffer)
	tw := tar.NewWriter(tarBuf)
	hdr := &tar.Header{
		Name: "Dockerfile",
		Mode: 0644,
		Size: int64(len(dockerfile)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(dockerfile)); err != nil {
		return nil, err
	}
	tw.Close()
	return tarBuf, nil
}

// registryHost 从仓库 URL 中提取 host 部分（去掉协议前缀和路径），用于拼接镜像引用
// 例如：https://csighub.tencentyun.com -> csighub.tencentyun.com
func registryHost(registryURL string) string {
	host := strings.TrimPrefix(registryURL, "https://")
	host = strings.TrimPrefix(host, "http://")
	if idx := strings.Index(host, "/"); idx >= 0 {
		host = host[:idx]
	}
	return host
}

// consumeImageStream 消费 Docker 镜像操作的 JSON 流，返回最后一条 status 消息。
// 遇到流中 error 字段时立即返回错误。
func consumeImageStream(dec *json.Decoder) (string, error) {
	var lastMessage string
	for {
		var msg struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		}
		if err := dec.Decode(&msg); err != nil {
			break
		}
		if msg.Error != "" {
			return "", errors.New(msg.Error)
		}
		if msg.Status != "" {
			lastMessage = msg.Status
		}
	}
	return lastMessage, nil
}
