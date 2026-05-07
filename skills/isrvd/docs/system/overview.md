# 系统概览 API

## 服务可用性探测

```bash
isrvd_get "/overview/probe"
```

| 字段 | 类型 | 说明 |
|------|------|------|
| agent | object | `{available: bool}` |
| apisix | object | `{available: bool}` |
| docker | object | `{available: bool}` |
| swarm | object | `{available: bool}` |
| compose | object | `{available: bool}` |

## 系统资源统计

```bash
isrvd_get "/overview/status"
isrvd_get "/overview/status" '{cpu: .system.cpuPercent, mem: .system.memPercent}'
isrvd_get "/overview/status" '.gpu'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| system | object | CPU、内存、磁盘、网络、uptime |
| diskIO | object[] | `{Name, ReadBytes, WriteBytes}` |
| gpu | object[] | `{name, vendor, memoryUsed, memoryTotal, utilization, temperature}` |
| go | object | `{version, numCPU, numGoroutine}` |
| version | string | isrvd 版本 |
