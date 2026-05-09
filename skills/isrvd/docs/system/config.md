# 系统配置与审计日志 API

## 配置存储后端

`CONFIG_PATH` 指定配置位置，未设置时读取 `./config.yml`；etcd 中的 value 仍是 `config.yml` 同款 YAML。

```bash
# 本地 YAML
CONFIG_PATH=/data/conf/isrvd.yml ./isrvd

# etcd：先导入，再启动
etcdctl put /isrvd/config "$(cat /data/conf/isrvd.yml)"
CONFIG_PATH="etcd://user:pass@127.0.0.1:2379/isrvd/config?scheme=http&timeout=5s" ./isrvd

# etcd key 不存在时，用 fallback YAML 初始化
CONFIG_PATH="etcd://127.0.0.1:2379/isrvd/config?fallback=/data/conf/isrvd.yml" ./isrvd
```

格式：`etcd://user:pass@host1:2379,host2:2379/key?scheme=http&timeout=5s&fallback=/path/config.yml`。

- `user:pass@` 可省略，认证也可通过 `ETCD_USERNAME` / `ETCD_PASSWORD` 提供。
- `scheme` 默认 `http`，`timeout` 默认 `5s`。
- `fallback` 仅在 etcd key 不存在时触发；连接失败、权限错误、超时、已有值解析失败不会 fallback。
- YAML provider 会保留历史兼容逻辑：成员明文密码自动迁移为 bcrypt 并写回 YAML。

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
