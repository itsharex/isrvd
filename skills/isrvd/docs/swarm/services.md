# Swarm 服务 API

## 列出服务

```bash
isrvd_get "/swarm/services"
isrvd_get "/swarm/services" '.[] | {name, image, replicas, runningTasks}'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 服务 ID |
| name | string | 服务名 |
| image | string | 镜像 |
| mode | string | `replicated` / `global` |
| replicas | number | 副本数 |
| runningTasks | number | 运行中的任务数 |
| env | string[] | 环境变量 |
| args | string[] | 命令参数 |
| networks | string[] | 网络名 |
| ports | object[] | `{protocol, targetPort, publishedPort, publishMode}` |
| mounts | object[] | `{type, source, target, readOnly}` |
| labels | object | 标签 |
| constraints | string[] | 放置约束 |
| createdAt | string | 创建时间 |
| updatedAt | string | 更新时间 |

## 查看服务详情

```bash
isrvd_get "/swarm/service/SVC_ID"
isrvd_get "/swarm/service/SVC_ID" '{name, image, replicas, runningTasks, ports, env}'
```

## 创建服务

```bash
isrvd_post "/swarm/service" '{
  "name": "web",
  "image": "nginx:latest",
  "replicas": 3,
  "ports": [{"targetPort": 80, "publishedPort": 80, "protocol": "tcp"}],
  "mounts": [{"type": "bind", "source": "/data", "target": "/app/data"}]
}'
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | ✅ | 服务名 |
| image | string | ✅ | 镜像 |
| mode | string | | `replicated`（默认）/ `global` |
| replicas | number | | 副本数（默认 1） |
| env | string[] | | 环境变量 |
| args | string[] | | 命令参数 |
| networks | string[] | | 网络名 |
| ports | object[] | | `{targetPort, publishedPort, protocol, publishMode}` |
| mounts | object[] | | `{type, source, target, readOnly}` |
| labels | object | | 标签 |
| constraints | string[] | | 放置约束 |

## 扩缩容

```bash
isrvd_post "/swarm/service/SVC_ID/action" '{"action":"scale","replicas":5}'
```

## 删除服务

```bash
isrvd_post "/swarm/service/SVC_ID/action" '{"action":"remove"}'
```

## 强制更新（重新部署）

```bash
isrvd_post "/swarm/service/SVC_ID/force-update"
```

## 服务日志

```bash
isrvd_get "/swarm/service/SVC_ID/logs?tail=100"
isrvd_get "/swarm/service/SVC_ID/logs?tail=20" '.logs'
```

返回：`{"logs": ["timestamped log line", ...]}`
