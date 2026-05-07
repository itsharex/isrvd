# Docker 网络 API

## 列出网络

```bash
isrvd_get "/docker/networks"
isrvd_get "/docker/networks" '.[].{name,driver,subnet}'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 网络 ID |
| name | string | 网络名 |
| driver | string | 驱动（bridge/overlay/host/...） |
| subnet | string | 子网 |
| scope | string | 作用域（local/swarm） |

## 查看网络详情

```bash
isrvd_get "/docker/network/<NETWORK_ID>"
isrvd_get "/docker/network/<NETWORK_ID>" '{name, subnet, gateway, containers}'
```

额外字段：`gateway, internal, enableIPv6, containers[]{id, name, ipv4, ipv6, macAddress}`

## 创建网络

```bash
isrvd_post "/docker/network" '{"name":"<NAME>","driver":"<DRIVER>","subnet":"<CIDR>"}'
```

## 删除网络

```bash
isrvd_post "/docker/network/<NETWORK_ID>/action" '{"action":"remove"}'
```
