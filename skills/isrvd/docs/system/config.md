# 系统配置与审计日志 API

## 配置存储后端

默认从 YAML 文件读取配置：

```bash
CONFIG_PATH=/data/conf/isrvd.yml ./isrvd
```

也可使用 etcd 存储同样的 YAML 内容，便于多实例共享配置：

```bash
CONFIG_PATH="etcd://127.0.0.1:2379/isrvd/config?scheme=http&timeout=5s&fallback=/data/conf/isrvd.yml" ./isrvd
```

| 配置 | 默认值 | 说明 |
|------|------|------|
| CONFIG_PATH | config.yml | 普通路径使用 YAML；`etcd://...` 使用 etcd |
| scheme | http | etcd endpoint 协议，放在 `CONFIG_PATH` query 中 |
| timeout | 5s | etcd 连接超时，放在 `CONFIG_PATH` query 中 |
| user:pass@ | 空 | etcd 用户名密码，放在 `CONFIG_PATH` authority 中；特殊字符需 URL encode |
| fallback | 空 | etcd key 不存在时读取的 YAML 文件路径，读取成功后会写入 etcd |
| ETCD_USERNAME / ETCD_PASSWORD | 空 | 可补充或覆盖 URI 中的认证信息 |

> etcd 中的值仍为完整 `config.yml` YAML 文本。首次使用前可通过 `etcdctl put /isrvd/config "$(cat config.yml)"` 导入。

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
