# 系统配置与审计日志 API

## 配置存储

`CONFIG_PATH` 指定配置位置，未设置时读取 `./config.yml`。

```bash
CONFIG_PATH=/data/conf/isrvd.yml ./isrvd
CONFIG_PATH="etcd://user:pass@127.0.0.1:2379/isrvd/config?fallback=/data/conf/isrvd.yml" ./isrvd
```

说明：etcd value 使用同款 YAML；`fallback` 仅在 key 不存在时用于初始化；外部变更仅提示重启，不自动热更新。

## 获取配置

```bash
isrvd_get "/system/config"
```

| 字段 | 类型 | 说明 |
|------|------|------|
| server | object | `{debug, listenAddr, jwtExpiration, maxUploadSize, proxyHeaderName, rootDirectory, allowedOrigins}`（jwtSecret 不返回） |
| agent | object | `{model, baseUrl}`（apiKey 不返回） |
| oidc | object | `{enabled, issuerUrl, clientId, redirectUrl, usernameClaim, scopes}`（clientSecret 不返回） |
| apisix | object | `{adminUrl}`（adminKey 不返回） |
| docker | object | `{host, containerRoot, registries}` |
| marketplace | object | `{url}` |
| links | object[] | `{label, url, icon}` |

## 更新配置

> 先通过 `isrvd_get "/system/config"` 获取当前值，按需修改后提交，不要硬编码配置内容。

```bash
isrvd_put "/system/config" '<CURRENT_CONFIG_WITH_CHANGES>'
```

配置说明：

- `clientSecret`、`jwtSecret`、`apiKey`、`adminKey` 等敏感字段不会通过 GET 返回；PUT 时为空表示保留原值。
- `oidc.redirectUrl` 生产环境建议显式配置固定 HTTPS 地址；留空会按当前请求 Host 自动生成，适合本地开发。
- `oidc.usernameClaim` 默认 `sub`；如改用 `email`，需确保 IdP 已验证邮箱且本地 `members.username` 与邮箱完全一致。

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
