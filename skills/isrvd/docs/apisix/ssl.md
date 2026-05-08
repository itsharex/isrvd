# APISIX SSL、PluginConfig、插件 API

## SSL 证书

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 证书 ID |
| snis | string[] | 域名列表 |
| cert | string | 证书（PEM） |
| key | string | 私钥（PEM） |
| status | number | `1`=启用, `0`=禁用 |

```bash
isrvd_get "/apisix/ssls"
isrvd_get "/apisix/ssl/<SSL_ID>"
isrvd_post "/apisix/ssl" '{"cert":"<CERT_PEM>","key":"<KEY_PEM>","snis":["<DOMAIN>"]}'
isrvd_put "/apisix/ssl/<SSL_ID>" '{"cert":"<CERT_PEM>","key":"<KEY_PEM>","snis":["<DOMAIN>"]}'
isrvd_delete "/apisix/ssl/<SSL_ID>"
```

---

## PluginConfig

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 配置 ID |
| desc | string | 描述 |
| plugins | object | 插件配置集合 |

```bash
isrvd_get "/apisix/plugin-configs"
isrvd_get "/apisix/plugin-config/<CONFIG_ID>"
isrvd_post "/apisix/plugin-config" '{"id":"<ID>","desc":"<DESC>","plugins":{...}}'
isrvd_put "/apisix/plugin-config/<CONFIG_ID>" '{"desc":"<DESC>","plugins":{...}}'
isrvd_delete "/apisix/plugin-config/<CONFIG_ID>"
```

---

## 插件列表

```bash
isrvd_get "/apisix/plugins"
```

返回 APISIX 已加载的插件名称及配置 schema。
