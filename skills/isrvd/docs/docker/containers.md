# Docker 容器 API

## 列出容器

```bash
isrvd_get "/docker/containers?all=true"
isrvd_get "/docker/containers" '.[].{name,state,image}'
```

返回 `ContainerInfo[]`：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 容器短 ID（12位） |
| name | string | 容器名 |
| image | string | 镜像名 |
| state | string | `running` / `exited` / `paused` / `created` / `restarting` / `dead` |
| status | string | 状态描述（如 "Up 2 hours"） |
| ports | string[] | 端口映射（如 `"0.0.0.0:8080->80/tcp"`） |
| networks | string[] | 所属网络名 |
| created | number | 创建时间戳（Unix 秒） |
| isSwarm | boolean | 是否为 Swarm 管理的容器 |
| labels | object | 标签键值对 |

## 创建容器

```bash
isrvd_post "/docker/container" '{
  "image": "nginx:latest",
  "name": "my-nginx",
  "ports": {"8080": "80"},
  "env": ["NGINX_HOST=example.com"],
  "volumes": [{"hostPath": "/data/html", "containerPath": "/usr/share/nginx/html", "readOnly": true}],
  "network": "my-network",
  "restart": "unless-stopped",
  "memory": 256,
  "cpus": 0.5
}'
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| image | string | ✅ | 镜像名 |
| name | string | ✅ | 容器名 |
| cmd | string[] | | 启动命令 |
| env | string[] | | 环境变量（`KEY=VALUE`） |
| ports | object | | `{"宿主端口": "容器端口"}` |
| volumes | object[] | | `{hostPath, containerPath, readOnly}` |
| network | string | | 网络名 |
| restart | string | | `no` / `always` / `on-failure` / `unless-stopped` |
| memory | number | | 内存限制（MB） |
| cpus | number | | CPU 限制（核数） |
| workdir | string | | 工作目录 |
| user | string | | 运行用户 |
| hostname | string | | 主机名 |
| privileged | boolean | | 特权模式 |
| capAdd | string[] | | 添加的 Capabilities |
| capDrop | string[] | | 移除的 Capabilities |

## 容器操作

```bash
isrvd_post "/docker/container/CONTAINER_ID/action" '{"action":"start"}'
isrvd_post "/docker/container/CONTAINER_ID/action" '{"action":"stop"}'
isrvd_post "/docker/container/CONTAINER_ID/action" '{"action":"restart"}'
isrvd_post "/docker/container/CONTAINER_ID/action" '{"action":"remove"}'
isrvd_post "/docker/container/CONTAINER_ID/action" '{"action":"pause"}'
isrvd_post "/docker/container/CONTAINER_ID/action" '{"action":"unpause"}'
```

## 容器日志

```bash
isrvd_get "/docker/container/CONTAINER_ID/logs?tail=100"
isrvd_get "/docker/container/CONTAINER_ID/logs?tail=50" '.logs[-10:]'
```

返回：`{id, logs: string[]}`

## 容器资源统计

```bash
isrvd_get "/docker/container/CONTAINER_ID/stats"
isrvd_get "/docker/container/CONTAINER_ID/stats" '{cpu: .cpuPercent, mem: .memoryPercent}'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| cpuPercent | number | CPU 使用率 (%) |
| memoryUsage | number | 内存使用（字节） |
| memoryLimit | number | 内存限制（字节） |
| memoryPercent | number | 内存使用率 (%) |
| networkRx | number | 网络接收（字节） |
| networkTx | number | 网络发送（字节） |
| blockRead | number | 磁盘读取（字节） |
| blockWrite | number | 磁盘写入（字节） |
| pids | number | 进程数 |

## 容器终端

WebSocket 连接，不通过 harness 调用：`GET /api/docker/container/:id/exec?shell=/bin/sh`
