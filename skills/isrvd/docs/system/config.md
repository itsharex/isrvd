# 系统配置与审计日志 API

## 获取配置

```bash
isrvd_get "/system/config"
isrvd_get "/system/config" '{server, agent, apisix}'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| server | object | `{debug, listenAddr, proxyHeaderName, rootDirectory}` |
| agent | object | `{model, baseUrl}`（apiKey 不返回） |
| apisix | object | `{adminUrl}`（adminKey 不返回） |
| docker | object | `{host, containerRoot, registries}` |
| marketplace | object | `{url}` |
| links | object[] | `{label, url, icon}` |

## 更新配置

```bash
isrvd_put "/system/config" '{
  "server": {"debug": false, "listenAddr": ":8080", "proxyHeaderName": "", "rootDirectory": "/data"},
  "agent": {"model": "gpt-4", "baseUrl": "https://api.openai.com/v1", "apiKey": "sk-..."},
  "apisix": {"adminUrl": "http://127.0.0.1:9180", "adminKey": "edd1c9f..."},
  "docker": {"host": "unix:///var/run/docker.sock", "containerRoot": "/data/containers"},
  "marketplace": {"url": ""},
  "links": [{"label": "Grafana", "url": "https://grafana.example.com", "icon": "chart"}]
}'
```

---

## 审计日志

```bash
isrvd_get "/system/audit/logs?limit=20"
isrvd_get "/system/audit/logs?username=admin&limit=10" '.[] | {timestamp, method, uri, success}'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| timestamp | string | 时间戳 |
| username | string | 操作用户 |
| method | string | HTTP 方法 |
| uri | string | 请求路径 |
| body | string | 请求体 |
| ip | string | 来源 IP |
| statusCode | number | 响应状态码 |
| success | boolean | 是否成功 |
| duration | number | 耗时（ms） |
