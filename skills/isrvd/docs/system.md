# 系统管理 API

> 各模块接口前缀不同，权限基于路由细粒度控制

---

## §1 系统概览

### §1.1 服务可用性探测

```
GET /api/overview/probe
```

返回各服务模块的可用性状态（Docker / Swarm / APISIX）。

响应示例：

```json
{
  "success": true,
  "payload": {
    "docker": true,
    "swarm": false,
    "apisix": true
  }
}
```

### §1.2 系统资源统计

```
GET /api/overview/status
```

返回 CPU、内存、磁盘、GPU 等系统资源使用率。

---

## §2 系统配置

> 配置接口前缀: `/api/system`

### §2.1 获取配置

```
GET /api/system/config
```

> ⚠️ 敏感字段（密码/密钥）不返回明文，只返回 `xxxSet: true/false` 表示是否已设置。

可选查询参数：`?reload=true` 从磁盘重新加载配置。

### §2.2 更新配置

```
PUT /api/system/config
```

请求体 `UpdateAllConfigRequest`（部分更新，`null` 的字段跳过）：

```json
{
  "server": {
    "debug": false,
    "listenAddr": ":8080",
    "jwtSecret": "",
    "proxyHeaderName": "",
    "rootDirectory": "/data"
  },
  "agent": {
    "model": "gpt-4",
    "baseUrl": "https://api.openai.com",
    "apiKey": ""
  },
  "apisix": {
    "adminUrl": "http://apisix:9180",
    "adminKey": ""
  },
  "docker": {
    "host": "unix:///var/run/docker.sock",
    "containerRoot": "/opt/containers"
  },
  "marketplace": {
    "url": "https://marketplace.example.com"
  },
  "links": [
    { "label": "监控", "url": "https://grafana.example.com", "icon": "chart-bar" }
  ]
}
```

> ⚠️ 敏感字段（`jwtSecret` / `apiKey` / `adminKey`）留空表示不修改。

---

## §3 审计日志

### §3.1 查询审计日志

```
GET /api/system/audit/logs?username=admin&limit=100
```

| 参数 | 说明 |
|------|------|
| username | 按用户名过滤（可选） |
| limit | 返回条数（默认 100） |

---

## §4 账号管理

> 账号接口前缀: `/api/account`

### §4.1 获取认证信息

```
GET /api/account/info
```

返回当前认证模式及已登录用户信息（匿名也可访问）。

### §4.2 登录

```
POST /api/account/login
Body: { "username": "admin", "password": "secret" }
```

响应：

```json
{
  "success": true,
  "payload": {
    "token": "jwt-token",
    "username": "admin"
  }
}
```

### §4.3 列出路由权限

```
GET /api/account/routes
```

返回所有已注册路由及其权限元信息（用于配置成员权限时参考）。

### §4.4 创建 API Token

```
POST /api/account/token
Body: { "name": "token-name", "expiresIn": "720h" }
```

创建长效 API Token（替代 JWT Token）。

### §4.5 修改密码

```
PUT /api/account/password
Body: { "oldPassword": "旧密码", "newPassword": "新密码" }
```

### §4.6 列出成员

```
GET /api/account/members
```

### §4.7 创建成员

```
POST /api/account/member
```

### §4.8 更新成员

```
PUT /api/account/member/:username
```

### §4.9 删除成员

```
DELETE /api/account/member/:username
```

> ⚠️ 首个系统账号禁止删除（前后端双重保护）。

请求体 `MemberUpsertRequest`：

```json
{
  "username": "user1",
  "password": "newpass",
  "homeDirectory": "/data/user1",
  "permissions": [
    "POST /api/filer/list",
    "POST /api/filer/upload",
    "POST /api/filer/modify",
    "POST /api/docker/container",
    "POST /api/swarm/service",
    "POST /api/apisix/route",
    "PUT /api/system/config",
    "POST /api/account/member",
    "ANY /api/agent/*path"
  ]
}
```

> `permissions` 为字符串数组，每个元素是一个路由权限点。
> 更新时密码留空表示不修改。

---

## §5 文件管理

> 文件接口前缀: `/api/filer`
> 所有文件操作基于用户的 home 目录，路径为相对路径。系统自动防止目录遍历。

### §5.1 列出文件和目录

```
POST /api/filer/list
Body: { "path": "/" }
```

### §5.2 创建目录

```
POST /api/filer/mkdir
Body: { "path": "/newdir" }
```

### §5.3 创建文件

```
POST /api/filer/create
Body: { "path": "/new.txt", "content": "内容" }
```

### §5.4 读取文件内容

```
POST /api/filer/read
Body: { "path": "/file.txt" }
```

### §5.5 保存文件内容

```
POST /api/filer/modify
Body: { "path": "/file.txt", "content": "新内容" }
```

### §5.6 重命名文件/目录

```
POST /api/filer/rename
Body: { "path": "/old.txt", "target": "new.txt" }
```

### §5.7 删除文件/目录

```
POST /api/filer/delete
Body: { "path": "/file.txt" }
```

### §5.8 修改文件权限

```
POST /api/filer/chmod
Body: { "path": "/file.txt", "mode": "755" }
```

### §5.9 上传文件

```
POST /api/filer/upload
Content-Type: multipart/form-data
字段: file(文件), path(目标目录,可选)
```

### §5.10 下载文件

```
POST /api/filer/download
Body: { "path": "/file.txt" }
```

返回文件流，触发浏览器下载。

### §5.11 压缩文件/目录

```
POST /api/filer/zip
Body: { "path": "/dir" }
```

### §5.12 解压文件

```
POST /api/filer/unzip
Body: { "path": "/file.zip" }
```

---

## §6 Web 终端

### §6.1 系统终端（WebSocket）

```
GET /api/shell?shell=bash
```

WebSocket 连接，需在查询参数中指定 shell（默认 `bash`）。

系统根据用户 homeDirectory 设置工作目录，并自动降级：
- 优先使用 PTY 模式（提供更完整的终端体验）
- PTY 不可用时自动降级到 Pipe 模式

### §6.2 容器终端（WebSocket）

```
GET /api/docker/container/:id/exec?shell=/bin/sh
```

WebSocket 连接，进入指定容器的终端。

---

## §7 LLM 代理

### §7.1 代理 LLM 请求

```
ANY /api/agent/*path
```

代理请求到配置的 LLM API（自动重写 model 字段）。

请求会被转发到 `config.Agent.BaseURL + path`，并自动附加 `config.Agent.APIKey`。

---

## 常见工作流

### 检查系统健康

```bash
# 系统资源
isrvd_get "/overview/status"

# 服务可用性
isrvd_get "/overview/probe"
```

### 创建新用户

```bash
isrvd_post "/account/member" '{
  "username": "developer",
  "password": "secure-password",
  "homeDirectory": "/data/developer",
  "permissions": [
    "POST /api/filer/list",
    "POST /api/filer/upload",
    "POST /api/docker/container",
    "POST /api/swarm/service",
    "POST /api/apisix/route",
    "PUT /api/system/config",
    "POST /api/account/member",
    "ANY /api/agent/*path"
  ]
}'
```

### 文件操作

```bash
# 列出文件
isrvd_post "/filer/list" '{"path": "/"}'

# 读取文件
isrvd_post "/filer/read" '{"path": "/test.txt"}'

# 保存文件
isrvd_post "/filer/modify" '{"path": "/test.txt", "content": "新内容"}'

# 上传文件
isrvd_upload "/filer/upload" "file" "./local-file.txt" "path=/uploads"
```
