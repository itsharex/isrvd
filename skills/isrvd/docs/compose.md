# Compose 部署 API

## Docker Compose（单机部署）

### 部署

```bash
# JSON 方式
isrvd_post "/compose/docker/deploy" "{\"projectName\":\"<PROJECT>\",\"content\":$(cat docker-compose.yml | jq -sR)}"

# multipart 方式（带初始化文件）
isrvd_upload "/compose/docker/deploy" "initFile" "./<INIT_FILE>" "projectName=<PROJECT>" "content=$(cat docker-compose.yml)"
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
isrvd_get "/compose/docker/<PROJECT>"
isrvd_get "/compose/docker/<PROJECT>" '.content'
```

### 重新部署

```bash
isrvd_post "/compose/docker/<PROJECT>/redeploy" '{"content":"<COMPOSE_YAML>"}'
```

---

## Swarm Stack（集群部署）

### 部署

```bash
isrvd_post "/compose/swarm/deploy" "{\"projectName\":\"<STACK>\",\"content\":$(cat docker-compose.yml | jq -sR)}"
```

### 获取 Stack 文件内容

```bash
isrvd_get "/compose/swarm/<STACK>"
isrvd_get "/compose/swarm/<STACK>" '.content'
```

### 重新部署

```bash
isrvd_post "/compose/swarm/<STACK>/redeploy" '{"content":"<COMPOSE_YAML>"}'
```
