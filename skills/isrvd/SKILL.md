---
name: isrvd-ops
description: 通过 isrvd API 进行容器部署、服务管理、镜像操作、路由配置、文件管理等运维操作。当用户要求"部署服务"、"管理容器"、"拉取/推送镜像"、"配置路由"、"管理 Swarm"、"Compose 部署"、"文件管理"、"Web 终端"等运维任务时使用此 Skill。
---

# isrvd 运维 Skill

```bash
source ./scripts/api.sh
# 首次认证（自动保存到 ~/.config/isrvd/profile.json，以后 source 自动加载）
isrvd_login "http://10.0.0.1:8080" "admin" "password"
# 或直接用 token: isrvd_token "http://10.0.0.1:8080" "eyJhbG..."
```

调用：`isrvd_get "/path" '[jq]'` / `isrvd_post "/path" 'body' '[jq]'`，输出紧凑 JSON。

---

## API 文档索引（按需读取，勿全部加载）

### Docker

| 文档 | 覆盖内容 |
|------|----------|
| [docs/docker/containers.md](docs/docker/containers.md) | 容器列表、创建、操作、日志、stats、终端 |
| [docs/docker/images.md](docs/docker/images.md) | 镜像列表、详情、搜索、构建、标签、拉取、推送、删除 |
| [docs/docker/networks.md](docs/docker/networks.md) | 网络列表、详情、创建、删除 |
| [docs/docker/volumes.md](docs/docker/volumes.md) | 数据卷列表、详情、创建、删除 |
| [docs/docker/registries.md](docs/docker/registries.md) | 镜像仓库 CRUD |

### Swarm

| 文档 | 覆盖内容 |
|------|----------|
| [docs/swarm/info.md](docs/swarm/info.md) | 集群信息、节点列表/详情/操作、加入令牌 |
| [docs/swarm/services.md](docs/swarm/services.md) | 服务列表、创建、扩缩容、强制更新、日志 |
| [docs/swarm/tasks.md](docs/swarm/tasks.md) | 任务列表 |

### Compose

| 文档 | 覆盖内容 |
|------|----------|
| [docs/compose.md](docs/compose.md) | Docker Compose 部署/重部署、Swarm Stack 部署/重部署 |

### APISIX

| 文档 | 覆盖内容 |
|------|----------|
| [docs/apisix/routes.md](docs/apisix/routes.md) | 路由 CRUD、启用/禁用 |
| [docs/apisix/upstreams.md](docs/apisix/upstreams.md) | 上游 CRUD、负载均衡配置 |
| [docs/apisix/consumers.md](docs/apisix/consumers.md) | Consumer CRUD、白名单管理 |
| [docs/apisix/ssl.md](docs/apisix/ssl.md) | SSL 证书、PluginConfig、插件列表 |

### 系统

| 文档 | 覆盖内容 |
|------|----------|
| [docs/system/overview.md](docs/system/overview.md) | 服务探测、系统资源统计 |
| [docs/system/config.md](docs/system/config.md) | 系统配置、审计日志 |
| [docs/system/account.md](docs/system/account.md) | 登录、成员管理、权限、API Token |
| [docs/system/filer.md](docs/system/filer.md) | 文件 CRUD、上传下载、压缩解压 |

---

## 决策树

```
用户需求
├── 部署/创建
│   ├── 单个容器        → docs/docker/containers.md
│   ├── 多容器应用(单机) → docs/compose.md §1
│   ├── 集群服务(Stack)  → docs/compose.md §2
│   ├── 集群服务(单服务) → docs/swarm/services.md
│   └── 配置路由         → docs/apisix/routes.md
│
├── 更新/变更
│   ├── 更新容器镜像     → docs/docker/images.md (拉取) + docs/docker/containers.md (重建)
│   ├── 扩缩容           → docs/swarm/services.md
│   ├── 重新部署         → docs/swarm/services.md (force-update)
│   ├── 修改路由/上游    → docs/apisix/routes.md 或 docs/apisix/upstreams.md
│   └── 修改系统配置     → docs/system/config.md
│
├── 查询/监控
│   ├── 容器/镜像/网络/卷 → docs/docker/ 下对应文件
│   ├── 集群/服务/任务    → docs/swarm/ 下对应文件
│   ├── 路由/上游/插件    → docs/apisix/ 下对应文件
│   ├── 系统状态          → docs/system/overview.md
│   ├── 日志             → docs/docker/containers.md 或 docs/swarm/services.md
│   └── 文件管理         → docs/system/filer.md
│
├── 删除/清理
│   ├── 容器/镜像/网络/卷 → docs/docker/ 下对应文件（action=remove）
│   ├── Swarm 服务        → docs/swarm/services.md（action=remove）
│   └── 路由/消费者       → docs/apisix/routes.md 或 docs/apisix/consumers.md
│
└── 管理
    ├── 镜像仓库         → docs/docker/registries.md
    ├── 成员/权限/Token  → docs/system/account.md
    ├── 文件管理         → docs/system/filer.md
    └── Web 终端         → GET /api/shell (WebSocket)
```

---

## 常见工作流

### 健康检查

```bash
isrvd_get "/overview/probe"
isrvd_get "/overview/status" '{cpu: .system.cpuPercent, mem: .system.memPercent}'
isrvd_get "/docker/containers" '.[].{name,state}'
isrvd_get "/swarm/services" '.[] | {name, running: .runningTasks, total: .replicas}'
```

### 拉取镜像并创建容器

```bash
isrvd_post "/docker/image/pull" '{"image":"nginx:latest"}'
isrvd_post "/docker/container" '{"image":"nginx:latest","name":"web","ports":{"80":"80"},"restart":"unless-stopped"}'
isrvd_get "/docker/containers" '.[].{name,state,image}'
```

### 更新容器镜像

```bash
isrvd_post "/docker/image/pull" '{"image":"nginx:1.25"}'
isrvd_post "/docker/container/OLD_ID/action" '{"action":"stop"}'
isrvd_post "/docker/container/OLD_ID/action" '{"action":"remove"}'
isrvd_post "/docker/container" '{"image":"nginx:1.25","name":"web","ports":{"80":"80"},"restart":"unless-stopped"}'
```

### Compose 部署

```bash
isrvd_post "/compose/docker/deploy" "{\"projectName\":\"my-app\",\"content\":$(cat docker-compose.yml | jq -sR)}"
```

### 部署 Swarm 服务并验证

```bash
isrvd_post "/swarm/service" '{"name":"api","image":"myapp:v1","replicas":2,"ports":[{"targetPort":8080,"publishedPort":8080}]}'
isrvd_get "/swarm/services" '.[] | select(.name=="api") | {runningTasks, replicas}'
isrvd_get "/swarm/tasks?serviceID=SVC_ID" '.[] | {nodeName, state}'
```

### 扩缩容

```bash
isrvd_post "/swarm/service/SVC_ID/action" '{"action":"scale","replicas":5}'
```

### 滚动更新

```bash
isrvd_post "/swarm/service/SVC_ID/force-update"
```

### 为新服务配置路由

```bash
isrvd_get "/apisix/routes" '.[] | {name, uri, host}'
isrvd_post "/apisix/route" '{"name":"new-svc","uri":"/svc/*","host":"api.example.com","status":1,"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:3000":1}}}'
```

### 创建共享上游 + 路由引用

```bash
isrvd_post "/apisix/upstream" '{"name":"backend","type":"roundrobin","nodes":{"10.0.0.1:8080":1}}'
isrvd_post "/apisix/route" '{"name":"route-a","uri":"/a/*","status":1,"upstream_id":"UPSTREAM_ID"}'
```

### 临时禁用/启用路由

```bash
isrvd_patch "/apisix/route/ROUTE_ID/status" '{"status":0}'
isrvd_patch "/apisix/route/ROUTE_ID/status" '{"status":1}'
```

### 文件操作

```bash
isrvd_post "/filer/list" '{"path":"/data"}'
isrvd_post "/filer/read" '{"path":"/data/config.yml"}' '.content'
isrvd_post "/filer/modify" '{"path":"/data/config.yml","content":"new content"}'
isrvd_upload "/filer/upload" "file" "./backup.tar.gz" "path=/data/backups"
```
