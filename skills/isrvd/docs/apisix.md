# APISIX 管理 API

> 所有接口前缀: `/api/apisix`
> 权限：依赖路由级细粒度控制，需对应路由权限

---

## §1 路由管理

### §1.1 列出路由

```
GET /api/apisix/routes
```

返回 `Route[]`：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 路由 ID |
| name | string | 路由名称 |
| uri | string | 匹配路径（单个） |
| uris | string[] | 匹配路径（多个） |
| host | string | 匹配域名（单个） |
| hosts | string[] | 匹配域名（多个） |
| desc | string | 描述 |
| status | number | `1`=启用, `0`=禁用 |
| priority | number | 优先级（数字越大越优先） |
| enable_websocket | boolean | 是否启用 WebSocket 代理 |
| plugin_config_id | string | 引用的插件配置 ID |
| upstream_id | string | 引用的上游 ID |
| upstream | object | 内联上游配置 |
| plugins | object | 插件配置 |
| consumers | string[] | 白名单消费者列表（只读） |
| timeout | RouteTimeout | 超时配置（connect/send/read，秒） |
| create_time | number | 创建时间（Unix 时间戳，只读） |
| update_time | number | 更新时间（Unix 时间戳，只读） |

`RouteTimeout`：

```json
{ "connect": 5, "send": 10, "read": 10 }
```

### §1.2 查看路由详情

```
GET /api/apisix/route/:id
```

### §1.3 创建路由

```
POST /api/apisix/route
```

**最小示例**：

```json
{
  "name": "my-api",
  "uri": "/api/*",
  "status": 1,
  "upstream": {
    "type": "roundrobin",
    "nodes": { "127.0.0.1:3000": 1 }
  }
}
```

**完整示例**（含域名、插件、WebSocket）：

```json
{
  "name": "web-app",
  "uri": "/*",
  "host": "app.example.com",
  "status": 1,
  "priority": 10,
  "enable_websocket": true,
  "upstream": {
    "type": "roundrobin",
    "nodes": { "10.0.0.1:8080": 1, "10.0.0.2:8080": 1 }
  },
  "plugins": {
    "proxy-rewrite": {
      "regex_uri": ["^/api/v1/(.*)", "/$1"]
    },
    "limit-req": {
      "rate": 100,
      "burst": 50,
      "key_type": "var",
      "key": "remote_addr",
      "rejected_code": 429
    }
  }
}
```

### §1.4 更新路由

```
PUT /api/apisix/route/:id
Body: <同创建路由的 Route 结构>
```

### §1.5 启用/禁用路由

```
PATCH /api/apisix/route/:id/status
Body: { "status": 1 }   // 1=启用, 0=禁用
```

### §1.6 删除路由

```
DELETE /api/apisix/route/:id
```

---

## §2 上游（Upstream）管理

`upstream` 和 `upstream_id` 二选一：

- **内联 upstream**：适合简单场景，直接在路由中定义
- **upstream_id**：引用已有上游，适合多路由共享

内联 upstream 结构：

```json
{
  "type": "roundrobin",
  "nodes": {
    "127.0.0.1:8080": 1,
    "127.0.0.1:8081": 2
  }
}
```

> `type` 可选: `roundrobin`（加权轮询）、`chash`（一致性哈希）、`ewma`（最小延迟）
> `nodes` 的值为权重

**Upstream 完整字段**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | ✅ | 上游名称 |
| type | string | ✅ | 负载算法: `roundrobin` / `chash` / `ewma` |
| nodes | object | ✅ | 节点地址与权重映射 |
| desc | string | | 描述 |
| hash_on | string | | 一致性哈希依据（chash 模式，如 `consumer` / `ip` / `header` / `cookie` / `vars`） |
| key | string | | hash_on 对应的 key（如 `$http_x_api_key`） |
| scheme | string | | 协议: `http`(默认) / `https` / `grpc` / `grpcs` |
| pass_host | string | | 传递主机头: `pass`(默认) / `node` / `rewrite` |
| upstream_host | string | | pass_host=rewrite 时使用的主机名 |
| retries | number | | 重试次数 |
| retry_timeout | number | | 重试超时（秒） |
| timeout | object | | 超时配置 `{ "connect": 5, "send": 10, "read": 10 }` |
| id | string | | 上游 ID（只读） |
| create_time | number | | 创建时间（Unix 时间戳，只读） |
| update_time | number | | 更新时间（Unix 时间戳，只读） |

### §2.1 列出上游

```
GET /api/apisix/upstreams
```

### §2.2 查看上游详情

```
GET /api/apisix/upstream/:id
```

### §2.3 创建上游

```
POST /api/apisix/upstream
```

### §2.4 更新上游

```
PUT /api/apisix/upstream/:id
```

### §2.5 删除上游

```
DELETE /api/apisix/upstream/:id
```

---

## §3 SSL 证书管理

### §3.1 列出证书

```
GET /api/apisix/ssls
```

### §3.2 查看证书详情

```
GET /api/apisix/ssl/:id
```

### §3.3 创建证书

```
POST /api/apisix/ssl
```

请求体 `SSL`：

```json
{
  "cert": "-----BEGIN CERTIFICATE-----\n...",
  "key": "-----BEGIN PRIVATE KEY-----\n...",
  "snis": ["example.com", "www.example.com"]
}
```

> 响应中还包含以下只读字段：
> - `id` — 证书 ID
> - `status` — 证书状态（`1`=启用, `0`=禁用）
> - `create_time` — 创建时间（Unix 时间戳）
> - `update_time` — 更新时间（Unix 时间戳）

### §3.4 更新证书

```
PUT /api/apisix/ssl/:id
```

### §3.5 删除证书

```
DELETE /api/apisix/ssl/:id
```

---

## §4 Consumer 管理

### §4.1 列出消费者

```
GET /api/apisix/consumers
```

### §4.2 创建消费者

```
POST /api/apisix/consumer
Body: { "username": "consumer1", "desc": "描述", "plugins": {...} }
```

### §4.3 更新消费者

```
PUT /api/apisix/consumer/:username
Body: { "desc": "新描述", "plugins": {...} }
```

### §4.4 删除消费者

```
DELETE /api/apisix/consumer/:username
```

---

## §5 白名单管理

### §5.1 获取白名单

```
GET /api/apisix/whitelist
```

返回路由的消费者白名单配置。

### §5.2 撤销白名单

```
POST /api/apisix/whitelist/revoke
Body: { "route_id": "路由ID", "consumer_name": "消费者名" }
```

---

## §6 PluginConfig 管理

### §6.1 列出插件配置

```
GET /api/apisix/plugin-configs
```

### §6.2 查看插件配置详情

```
GET /api/apisix/plugin-config/:id
```

### §6.3 创建插件配置

```
POST /api/apisix/plugin-config
```

请求体 `PluginConfig`：

```json
{
  "id": "my-plugin-config",
  "plugins": {
    "limit-req": {
      "rate": 100,
      "burst": 50,
      "key": "remote_addr",
      "rejected_code": 429
    }
  }
}
```

### §6.4 更新插件配置

```
PUT /api/apisix/plugin-config/:id
```

### §6.5 删除插件配置

```
DELETE /api/apisix/plugin-config/:id
```

---

## §7 插件列表

### §7.1 列出所有可用插件

```
GET /api/apisix/plugins
```

返回 APISIX 已加载的插件列表及其配置 schema。

---

## 常见工作流

### 为新服务配置路由

```bash
# 1. 查看现有路由，避免冲突
isrvd_get "/apisix/routes"

# 2. 查看可用上游
isrvd_get "/apisix/upstreams"

# 3. 创建路由
isrvd_post "/apisix/route" '{
  "name": "new-service",
  "uri": "/new-service/*",
  "host": "api.example.com",
  "status": 1,
  "upstream": {
    "type": "roundrobin",
    "nodes": {"127.0.0.1:3000": 1}
  }
}'

# 4. 验证
isrvd_get "/apisix/routes"
```

### 临时禁用路由

```bash
isrvd_patch "/apisix/route/ROUTE_ID/status" '{"status": 0}'
```

### 修改路由上游节点

```bash
# 获取当前路由配置
isrvd_get "/apisix/route/ROUTE_ID"

# 更新（需要传完整 Route 结构）
isrvd_put "/apisix/route/ROUTE_ID" '{
  "name": "my-api",
  "uri": "/api/*",
  "status": 1,
  "upstream": {
    "type": "roundrobin",
    "nodes": {"10.0.0.1:3000": 1, "10.0.0.2:3000": 1}
  }
}'
```
