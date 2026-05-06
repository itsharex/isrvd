---
name: isrvd-ops
description: 通过 isrvd API 进行容器部署、服务管理、镜像操作、路由配置、文件管理等运维操作。当用户要求"部署服务"、"管理容器"、"拉取/推送镜像"、"配置路由"、"管理 Swarm"、"Compose 部署"、"文件管理"、"Web 终端"等运维任务时使用此 Skill。
---

# isrvd 运维操作 Skill

isrvd 是一个集成了 Docker、Swarm、APISIX、Compose 和文件管理的轻量级运维平台。

> **渐进式披露**：本文件只包含概述和决策树。各模块的详细 API 文档在 `docs/` 子目录中，按需读取。
> **Script Harness**：`scripts/` 目录提供可直接执行的 shell 脚本，封装了认证、调用和常见工作流。

---

## 快速开始

### 1. 环境准备

```bash
# 检测 isrvd 服务状态并获取认证 token
source ./scripts/api.sh
isrvd_init "http://localhost:8080" "admin" "password"
```

### 2. 常用快捷脚本

| 脚本 | 用途 | 示例 |
|------|------|------|
| `scripts/api.sh` | 通用 API 封装（认证、GET/POST/PUT/DELETE） | `source scripts/api.sh && isrvd_get "/docker/containers"` |
| `scripts/deploy.sh` | 一键部署（Compose/容器/Swarm 服务） | `./scripts/deploy.sh compose my-app ./docker-compose.yml` |
| `scripts/health-check.sh` | 健康检查与状态报告 | `./scripts/health-check.sh` |

---

## 决策树：用户想做什么？

```
用户需求
├── 部署/创建
│   ├── 单个容器        → ./scripts/deploy.sh container <name> <image> [json]
│   ├── 多容器应用(单机) → ./scripts/deploy.sh compose <name> <file> [--init-url=]
│   ├── 集群服务(Swarm)  → ./scripts/deploy.sh swarm <name> <file>
│   ├── 集群服务(单服务) → ./scripts/deploy.sh service <name> <image> [json]
│   └── 配置路由         → 读 docs/apisix.md §1-§3
│
├── 更新/变更
│   ├── 更新容器镜像 → docs/docker.md §3.7 拉取 + 重建（或读 §2.3 操作容器）
│   ├── 扩缩容 Swarm 服务 → docs/swarm.md §4.1 (action=scale)
│   ├── 重新部署 Swarm 服务 → docs/swarm.md §4.2 (action=redeploy)
│   ├── 修改路由规则 → docs/apisix.md §4
│   └── 修改系统配置 → docs/system.md §2 (PUT /api/system/config)
│
├── 查询/监控
│   ├── 查看容器/镜像/网络/卷 → docs/docker.md §1-§6（或用 scripts/api.sh）
│   ├── 查看集群/服务/任务 → docs/swarm.md §1-§2
│   ├── 查看路由/上游/插件 → docs/apisix.md §1
│   ├── 系统状态/健康检查 → ./scripts/health-check.sh
│   ├── 查看日志 → docs/docker.md §2.4 或 docs/swarm.md §4.3
│   └── 文件管理 → docs/system.md §5
│
├── 删除/清理
│   ├── 删除容器/镜像/网络/卷 → docs/docker.md 各模块 action=remove
│   ├── 删除 Swarm 服务 → docs/swarm.md §4.1 (action=remove)
│   └── 删除路由/消费者 → docs/apisix.md §5-§6
│
└── 管理
    ├── 镜像仓库管理 → docs/docker.md §7
    ├── 成员/权限管理 → docs/system.md §4
    ├── 文件管理 → docs/system.md §5
    └── Web 终端 → GET /api/shell (WebSocket)
```

---

## 通用约定（所有模块共享）

### 认证

- **JWT**：`Authorization: Bearer <token>`（通过 `POST /api/account/login` 获取）
- **代理 Header**：反向代理注入用户名（配置项 `proxyHeaderName`）
- **API Token**：通过 `POST /api/account/token` 创建长效令牌

### 统一响应格式

```json
{ "success": true|false, "message": "描述", "payload": <数据或null> }
```

### 权限模块

权限基于路由进行细粒度控制，每个用户可以独立授予各模块下具体 API 路由的访问权限。

**权限格式**：`METHOD /api/<模块>/<路由>`（具体 HTTP 方法与路径）

**访问级别**：
- `anon`：无需认证（如登录接口）
- `auth`：需登录但无需额外权限（如获取用户信息）
- `perm`：需要对应路由权限（默认）

**常用模块与路由示例**：

| 模块 | 路由权限点示例 | 说明 |
|------|---------------|------|
| `overview` | `GET /api/overview/probe` | 服务探测 |
| `overview` | `GET /api/overview/status` | 系统概览统计 |
| `account` | `POST /api/account/login` | 登录（匿名） |
| `account` | `GET /api/account/members` | 成员管理 |
| `system` | `GET /api/system/config` | 系统配置（需认证） |
| `system` | `PUT /api/system/config` | 更新配置 |
| `system` | `GET /api/system/audit/logs` | 审计日志 |
| `filer` | `POST /api/filer/list` | 文件管理（列出） |
| `filer` | `POST /api/filer/upload` | 文件管理（上传） |
| `filer` | `POST /api/filer/modify` | 文件管理（修改） |
| `shell` | `GET /api/shell` | Web 终端（WebSocket） |
| `agent` | `ANY /api/agent/*path` | LLM 代理 |
| `apisix` | `GET /api/apisix/routes` | APISIX 管理 |
| `docker` | `GET /api/docker/containers` | Docker 管理 |
| `swarm` | `GET /api/swarm/services` | Swarm 管理 |
| `compose` | `POST /api/compose/docker/deploy` | Compose 管理 |

> 留空 = 无权限；具体可用路由见各模块 API 文档

---

## 详细文档索引

| 文档 | 覆盖模块 | 何时读取 |
|------|----------|----------|
| [docs/docker.md](docs/docker.md) | 容器、镜像、网络、数据卷、镜像仓库 | 需要管理 Docker 资源时 |
| [docs/swarm.md](docs/swarm.md) | Swarm 集群、节点、服务、任务 | 需要管理集群服务时 |
| [docs/compose.md](docs/compose.md) | Docker Compose、Swarm Stack 部署 | 需要通过 compose 文件部署时 |
| [docs/apisix.md](docs/apisix.md) | 路由、Consumer、上游、SSL、插件配置、白名单 | 需要配置 API 网关时 |
| [docs/system.md](docs/system.md) | 系统状态、配置、成员、文件管理、审计日志 | 需要系统管理操作时 |

## 脚本索引

| 脚本 | 用途 | 何时使用 |
|------|------|----------|
| [scripts/api.sh](scripts/api.sh) | 通用 API 调用封装 | 所有 API 调用的基础，其他脚本依赖此文件 |
| [scripts/deploy.sh](scripts/deploy.sh) | 部署快捷脚本 | 需要快速部署容器/Compose/Swarm 服务时 |
| [scripts/health-check.sh](scripts/health-check.sh) | 健康检查与状态报告 | 需要检查系统和服务状态时 |

---

## 实际 API 路由速查表

### Overview 模块
- `GET /api/overview/probe` - 探测服务可用性（Docker、Swarm、APISIX）
- `GET /api/overview/status` - 获取系统资源统计（CPU、内存、磁盘、GPU）

### Account 模块
- `GET /api/account/info` - 获取认证信息（匿名可访问）
- `POST /api/account/login` - 登录获取 JWT Token（匿名可访问）
- `GET /api/account/routes` - 列出所有路由及其权限元信息
- `POST /api/account/token` - 创建长效 API Token
- `PUT /api/account/password` - 修改当前用户密码
- `GET /api/account/members` - 列出所有成员
- `POST /api/account/member` - 创建成员
- `PUT /api/account/member/:username` - 更新成员
- `DELETE /api/account/member/:username` - 删除成员

### Filer 模块（文件管理）
- `POST /api/filer/list` - 列出文件和目录
- `POST /api/filer/mkdir` - 创建目录
- `POST /api/filer/create` - 创建文件
- `POST /api/filer/read` - 读取文件内容
- `POST /api/filer/modify` - 保存文件内容
- `POST /api/filer/rename` - 重命名文件/目录
- `POST /api/filer/delete` - 删除文件/目录
- `POST /api/filer/chmod` - 修改文件权限
- `POST /api/filer/upload` - 上传文件（multipart）
- `POST /api/filer/download` - 下载文件
- `POST /api/filer/zip` - 压缩文件/目录
- `POST /api/filer/unzip` - 解压文件

### Shell 模块（Web 终端）
- `GET /api/shell` - 打开 Web 终端（WebSocket 连接）

### Agent 模块（LLM 代理）
- `ANY /api/agent/*path` - 代理 LLM API 请求（自动重写 model）

### Docker 模块
- `GET /api/docker/info` - 获取 Docker 引擎信息
- **容器**：`GET /api/docker/containers`、`POST /api/docker/container`、`:id/stats`、`:id/action`、`logs`、`exec`
- **镜像**：`GET /api/docker/images`、`search`、`:id/action`、`tag`、`inspect`、`build`、`pull`、`push`
- **网络**：`GET /api/docker/networks`、`:id/action`、`POST /api/docker/network`、`:id`
- **卷**：`GET /api/docker/volumes`、`:name/action`、`POST /api/docker/volume`、`:name`
- **仓库**：`GET /api/docker/registries`、`POST /api/docker/registry`、`PUT`、`DELETE`

### Swarm 模块
- `GET /api/swarm/info` - 获取 Swarm 集群信息
- **节点**：`GET /api/swarm/nodes`、`:id`、`token`、`POST :id/action`
- **服务**：`GET /api/swarm/services`、`:id`、`POST /api/swarm/service`、`:id/action`、`force-update`、`logs`
- **任务**：`GET /api/swarm/tasks?serviceID=`

### Compose 模块
- **Docker Compose**：`GET /api/compose/docker/:name`、`POST /api/compose/docker/deploy`、`redeploy`
- **Swarm Stack**：`GET /api/compose/swarm/:name`、`POST /api/compose/swarm/deploy`、`redeploy`

### APISIX 模块
- **路由**：`GET/POST /api/apisix/routes`、`PUT/PATCH/DELETE /api/apisix/route/:id`
- **Consumer**：`GET/POST /api/apisix/consumers`、`PUT/DELETE /api/apisix/consumer/:username`
- **白名单**：`GET /api/apisix/whitelist`、`POST /api/apisix/whitelist/revoke`
- **PluginConfig**：`GET/POST /api/apisix/plugin-configs`、`PUT/DELETE /api/apisix/plugin-config/:id`
- **Upstream**：`GET/POST /api/apisix/upstreams`、`PUT/DELETE /api/apisix/upstream/:id`
- **SSL**：`GET/POST /api/apisix/ssls`、`PUT/DELETE /api/apisix/ssl/:id`
- **插件**：`GET /api/apisix/plugins`
