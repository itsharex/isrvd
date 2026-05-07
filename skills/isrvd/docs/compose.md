# Compose 部署 API

## Docker Compose（单机部署）

### 部署

```bash
# JSON 方式
isrvd_post "/compose/docker/deploy" "{\"projectName\":\"my-app\",\"content\":$(cat docker-compose.yml | jq -sR)}"

# multipart 方式（带初始化文件）
isrvd_upload "/compose/docker/deploy" "initFile" "./init.tar.gz" "projectName=my-app" "content=$(cat docker-compose.yml)"
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| projectName | string | ✅ | 项目名 |
| content | string | ✅ | docker-compose.yml 内容 |
| initURL | string | | 初始化数据 URL |
| initFile | file | | 初始化数据文件 |

返回 `DeployResult`：`{target, items: string[], installDir}`

### 获取 Compose 文件内容

```bash
isrvd_get "/compose/docker/my-app"
isrvd_get "/compose/docker/my-app" '.content'
```

### 重新部署

```bash
isrvd_post "/compose/docker/my-app/redeploy" '{"content":"version: \"3\"\nservices:\n  web:\n    image: nginx:latest"}'
```

---

## Swarm Stack（集群部署）

### 部署

```bash
isrvd_post "/compose/swarm/deploy" "{\"projectName\":\"my-stack\",\"content\":$(cat docker-compose.yml | jq -sR)}"
```

### 获取 Stack 文件内容

```bash
isrvd_get "/compose/swarm/my-stack"
isrvd_get "/compose/swarm/my-stack" '.content'
```

### 重新部署

```bash
isrvd_post "/compose/swarm/my-stack/redeploy" '{"content":"新的 compose 内容"}'
```
