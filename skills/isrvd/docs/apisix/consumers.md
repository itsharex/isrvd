# APISIX Consumer 与白名单 API

## Consumer 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| username | string | 消费者名（唯一标识） |
| desc | string | 描述 |
| plugins | object | 认证插件（如 key-auth, jwt-auth） |

## 列出消费者

```bash
isrvd_get "/apisix/consumers"
isrvd_get "/apisix/consumers" '.[] | {username, desc}'
```

## 创建消费者

```bash
isrvd_post "/apisix/consumer" '{"username":"app-client","desc":"移动端","plugins":{"key-auth":{"key":"my-secret-key"}}}'
```

## 更新消费者

```bash
isrvd_put "/apisix/consumer/app-client" '{"desc":"新描述","plugins":{"key-auth":{"key":"new-key"}}}'
```

## 删除消费者

```bash
isrvd_delete "/apisix/consumer/app-client"
```

---

## 白名单

### 获取白名单

```bash
isrvd_get "/apisix/whitelist"
isrvd_get "/apisix/whitelist" '.[] | {id, name, consumers}'
```

返回含 key-auth + consumer-restriction 的路由列表，`consumers` 字段标识白名单消费者。

### 撤销白名单

```bash
isrvd_post "/apisix/whitelist/revoke" '{"route_id":"ROUTE_ID","consumer_name":"app-client"}'
```
