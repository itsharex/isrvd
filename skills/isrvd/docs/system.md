# 系统管理 API

> 系统接口前缀: `/api/system`，账号接口前缀: `/api/auth` 和 `/api/members`，文件接口前缀: `/api/filer`  
> 权限基于路由细粒度控制，如 `POST /api/system/config`、
> `POST /api/members`、`POST /api/filer/list` 等

---

## §1 系统状态

### §1.1 系统资源统计

```
GET /api/system/stats
```

返回 CPU、内存、磁盘使用率等系统资源信息。

### §1.2 服务可用性探测

```
GET /api/system/probe
```

返回各服务模块的可用性状态（Docker / Swarm / APISIX）。

---

## §2 系统配置

### §2.1 获取配置

```
GET /api/system/config
```

> ⚠️ 敏感字段（密码/密钥）不返回明文，只返回 `xxxSet: true/false` 表示是否已设置。

### §2.2 更新配置

```
PUT /api/system/config
```

请求体 `UpdateAllRequest`（部分更新，`null` 的分区跳过）：

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

## §3 成员管理

> 成员管理接口前缀: `/api/members`，需要对应路由权限（如 `POST /api/members`）

### §3.1 列出成员

```
GET /api/members
```

### §3.2 创建成员

```
POST /api/member
```

### §3.3 更新成员

```
PUT /api/member/:username
```

### §3.4 删除成员

```
DELETE /api/member/:username
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
    "POST /api/docker/containers",
    "POST /api/swarm/services",
    "POST /api/apisix/routes",
    "POST /api/system/config",
    "POST /api/member",
    "POST /api/agent/chat"
  ]
}
```

> `permissions` 为字符串数组，每个元素是一个路由权限点。
> 更新时密码留空表示不修改。

---

## §4 文件管理

所有文件操作基于用户的 home 目录，路径为相对路径。系统自动防止目录遍历。

### §4.1 查看操作（需要对应路由权限）

```
POST /api/filer/list       Body: { "path": "/" }              # 列出目录（需要 `POST /api/filer/list`）
POST /api/filer/read       Body: { "path": "/file.txt" }      # 读取文件内容（需要 `POST /api/filer/read`）
POST /api/filer/download   Body: { "path": "/file.txt" }      # 下载文件（需要 `POST /api/filer/download`）
```
### §4.2 修改操作（需要对应路由权限）

```
POST /api/filer/create     Body: { "path": "/new.txt", "content": "内容" }        # 需要 `POST /api/filer/create`
POST /api/filer/modify     Body: { "path": "/file.txt", "content": "新内容" }    # 需要 `POST /api/filer/modify`
POST /api/filer/mkdir      Body: { "path": "/newdir" }                           # 需要 `POST /api/filer/mkdir`
POST /api/filer/upload                                                                # 需要 `POST /api/filer/upload`
POST /api/filer/delete     Body: { "path": "/file.txt" }                            # 需要 `POST /api/filer/delete`
POST /api/filer/rename     Body: { "path": "/old.txt", "target": "new.txt" }        # 需要 `POST /api/filer/rename`
POST /api/filer/chmod      Body: { "path": "/file.txt", "mode": "755" }            # 需要 `POST /api/filer/chmod`
POST /api/filer/zip        Body: { "path": "/dir" }                                  # 需要 `POST /api/filer/zip`
POST /api/filer/unzip      Body: { "path": "/file.zip" }                            # 需要 `POST /api/filer/unzip`
```

### §4.3 文件上传

```
POST /api/filer/upload
Content-Type: multipart/form-data
字段: file(文件), path(目标目录,可选)
```

---

## §5 认证

### §5.1 登录

```
POST /api/auth/login
Body: { "username": "admin", "password": "secret" }
Response: { "success": true, "payload": { "token": "jwt-token", "username": "admin" } }
```

### §5.2 获取当前用户信息

```
GET /api/auth/info
```

需要认证（JWT 或代理 Header）。

### §5.3 登出

```
POST /api/auth/logout
```

---

## §6 WebSocket 终端

### §6.1 系统终端

```
WS /api/shell?token=<jwt-token>&shell=bash
```

### §6.2 容器终端

```
WS /api/docker/containers/<容器ID>/exec?token=<jwt-token>&shell=/bin/sh
```
---

## 常见工作流

### 检查系统健康

```bash
# 系统资源
isrvd_get "/system/stats"

# 服务可用性
isrvd_get "/system/probe"
```

### 创建新用户

```bash
isrvd_post "/members" '{
  "username": "developer",
  "password": "secure-password",
  "homeDirectory": "/data/developer",
  "permissions": [
    "POST /api/filer/list",
    "POST /api/filer/upload",
    "POST /api/docker/containers",
    "POST /api/swarm/services",
    "POST /api/apisix/routes",
    "POST /api/system/config",
    "POST /api/member",
    "POST /api/agent/chat"
  ]
}'
```
