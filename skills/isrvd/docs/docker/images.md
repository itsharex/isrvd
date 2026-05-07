# Docker 镜像 API

## 列出镜像

```bash
isrvd_get "/docker/images"
isrvd_get "/docker/images" '.[].{tags: .repoTags, size}'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 镜像 ID |
| shortId | string | 短 ID |
| repoTags | string[] | 标签列表 |
| repoDigests | string[] | Digest 列表 |
| size | number | 大小（字节） |
| created | number | 创建时间戳 |

## 查看镜像详情

```bash
isrvd_get "/docker/image/<IMAGE_ID>"
isrvd_get "/docker/image/<IMAGE_ID>" '{tags: .repoTags, arch: .architecture, cmd, env}'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| repoTags | string[] | 标签 |
| size | number | 大小 |
| created | string | 创建时间 |
| architecture | string | 架构 |
| os | string | 操作系统 |
| cmd | string[] | 默认命令 |
| entrypoint | string[] | 入口命令 |
| env | string[] | 环境变量 |
| exposedPorts | string[] | 暴露端口 |
| labels | object | 标签 |
| layers | number | 层数 |
| layerDetails | object[] | `{digest, createdBy, created, size, empty}` |

## 搜索镜像

```bash
isrvd_get "/docker/images/search?name=<KEYWORD>"
```

## 构建镜像

```bash
isrvd_post "/docker/image/build" '{"dockerfile":"<DOCKERFILE_CONTENT>","tag":"<IMAGE>:<TAG>"}'
```

## 镜像打标签

```bash
isrvd_post "/docker/image/<IMAGE_ID>/tag" '{"repoTag":"<REPO>/<NAME>:<TAG>"}'
```

## 镜像删除

```bash
isrvd_post "/docker/image/<IMAGE_ID>/action" '{"action":"remove"}'
```

## 拉取镜像

```bash
isrvd_post "/docker/image/pull" '{"image":"<IMAGE>"}'
isrvd_post "/docker/image/pull" '{"image":"<IMAGE>","registryUrl":"<REGISTRY_URL>","namespace":"<NS>"}'
```

> `registryUrl` 为空则从 Docker Hub 拉取

## 推送镜像

```bash
isrvd_post "/docker/image/push" '{"image":"<IMAGE>","registryUrl":"<REGISTRY_URL>","namespace":"<NS>"}'
```
