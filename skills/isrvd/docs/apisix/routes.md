# APISIX 路由 API

## 列出路由

```bash
isrvd_get "/apisix/routes"
isrvd_get "/apisix/routes" '.[] | {id, name, uri, host, status}'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 路由 ID |
| name | string | 名称 |
| uri | string | 匹配路径 |
| uris | string[] | 匹配路径（多个） |
| host | string | 匹配域名 |
| hosts | string[] | 匹配域名（多个） |
| desc | string | 描述 |
| status | number | `1`=启用, `0`=禁用 |
| priority | number | 优先级（越大越优先） |
| enable_websocket | boolean | WebSocket 代理 |
| plugin_config_id | string | 插件配置 ID |
| upstream_id | string | 上游 ID |
| upstream | object | 内联上游 |
| plugins | object | 插件配置 |
| consumers | string[] | 白名单消费者（只读） |
| timeout | object | `{connect, send, read}`（秒） |

## 查看路由详情

```bash
isrvd_get "/apisix/route/ROUTE_ID"
```

## 创建路由

```bash
# 最小
isrvd_post "/apisix/route" '{"name":"my-api","uri":"/api/*","status":1,"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:3000":1}}}'

# 完整
isrvd_post "/apisix/route" '{
  "name": "web-app",
  "uri": "/*",
  "host": "app.example.com",
  "status": 1,
  "priority": 10,
  "enable_websocket": true,
  "upstream": {"type":"roundrobin","nodes":{"10.0.0.1:8080":1,"10.0.0.2:8080":1}},
  "plugins": {
    "proxy-rewrite": {"regex_uri": ["^/api/v1/(.*)", "/$1"]},
    "limit-req": {"rate":100,"burst":50,"key_type":"var","key":"remote_addr","rejected_code":429}
  }
}'
```

## 更新路由

```bash
isrvd_put "/apisix/route/ROUTE_ID" '{"name":"my-api","uri":"/api/*","status":1,"upstream":{"type":"roundrobin","nodes":{"10.0.0.1:3000":1,"10.0.0.2:3000":1}}}'
```

## 启用/禁用路由

```bash
isrvd_patch "/apisix/route/ROUTE_ID/status" '{"status":0}'  # 禁用
isrvd_patch "/apisix/route/ROUTE_ID/status" '{"status":1}'  # 启用
```

## 删除路由

```bash
isrvd_delete "/apisix/route/ROUTE_ID"
```
