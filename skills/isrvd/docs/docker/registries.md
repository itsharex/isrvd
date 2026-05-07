# Docker 镜像仓库 API

## 列出仓库

```bash
isrvd_get "/docker/registries"
isrvd_get "/docker/registries" '.[].{name,url,username}'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 仓库名 |
| url | string | 仓库地址 |
| username | string | 用户名 |
| description | string | 描述 |

## 添加仓库

```bash
isrvd_post "/docker/registry" '{"name":"私有仓库","url":"https://registry.example.com","username":"user","password":"pass","description":"描述"}'
```

## 更新仓库

```bash
isrvd_put "/docker/registry?url=https://registry.example.com" '{"name":"新名称","url":"https://registry.example.com","username":"user","password":"newpass","description":"描述"}'
```

> 密码为空时保留原密码。

## 删除仓库

```bash
isrvd_delete "/docker/registry?url=https://registry.example.com"
```
