# 系统配置与审计日志 API

## 获取配置

```bash
isrvd_get "/system/config"
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

> 先通过 `isrvd_get "/system/config"` 获取当前值，按需修改后提交，不要硬编码配置内容。

```bash
isrvd_put "/system/config" '<CURRENT_CONFIG_WITH_CHANGES>'
```

---

## 审计日志

```bash
isrvd_get "/system/audit/logs?limit=20"
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
