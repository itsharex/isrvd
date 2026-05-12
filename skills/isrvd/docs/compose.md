# Compose API

Compose 接口用于单机 Docker Compose 与 Swarm Stack 的部署、读取配置、重部署，以及按服务更新镜像并重建。

## 字段说明

### ComposeDeploy

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `content` | string | 是 | 完整 compose yaml 文本 |
| `initURL` | string | 否 | 附加运行文件 zip 下载地址，仅 Docker Compose 部署生效 |
| `initFile` | file | 否 | 附加运行文件 zip，仅 Docker Compose 部署生效；与 `initURL` 互斥且文件优先 |

### ComposeRedeploy

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `content` | string | 二选一 | 完整 compose yaml 文本（全量重建） |
| `serviceName` | string | 二选一 | 要更新镜像的 compose 服务名（按服务更新） |
| `image` | string | 按需 | 新镜像名，`serviceName` 非空时必填 |

### ComposeDeployResult

| 字段 | 类型 | 说明 |
|---|---|---|
| `projectName` | string | 实际使用的项目名 |
| `items` | string[] | 创建或重建的容器/服务列表 |
| `installDir` | string | Docker Compose 落盘目录；Swarm 不返回 |

## Docker Compose

### 部署

Docker Compose 部署使用 multipart form，支持上传附加运行文件。

```bash
isrvd_upload "/compose/docker/deploy" "initFile" "./init.zip" "content=$(cat docker-compose.yml)"
```

仅提交 compose 内容：

```bash
isrvd_post "/compose/docker/deploy" "$(jq -n --arg content "$(cat docker-compose.yml)" '{content:$content}')"
```

使用远程附加文件：

```bash
isrvd_post "/compose/docker/deploy" '{"content":"<COMPOSE_YAML>","initURL":"<ZIP_URL>"}'
```

### 读取 compose 文件

```bash
isrvd_get "/compose/docker/<NAME>"
```

### 重部署

```bash
isrvd_post "/compose/docker/<NAME>/redeploy" "$(jq -n --arg content "$(cat docker-compose.yml)" '{content:$content}')"
```

### 按服务更新镜像并重建

```bash
isrvd_post "/compose/docker/<NAME>/redeploy" '{"serviceName":"<SERVICE_NAME>","image":"<NEW_IMAGE>"}'
```

## Swarm Compose

### 部署

```bash
isrvd_post "/compose/swarm/deploy" "$(jq -n --arg content "$(cat stack.yml)" '{content:$content}')"
```

### 读取 compose 文件

```bash
isrvd_get "/compose/swarm/<NAME>"
```

### 重部署

```bash
isrvd_post "/compose/swarm/<NAME>/redeploy" "$(jq -n --arg content "$(cat stack.yml)" '{content:$content}')"
```

### 按服务更新镜像并重建

```bash
isrvd_post "/compose/swarm/<NAME>/redeploy" '{"serviceName":"<SERVICE_NAME>","image":"<NEW_IMAGE>"}'
```
