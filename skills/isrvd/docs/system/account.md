# 账户与成员管理 API

## 获取认证信息

```bash
isrvd_get "/account/info"
```

返回：`{mode, username, member}`

## 登录

```bash
isrvd_post "/account/login" '{"username":"<USER>","password":"<PASS>"}'
```

返回：`{"token": "eyJ..."}`

> 通常使用 `isrvd_login` 命令而非直接调用此接口。

## 列出路由权限

```bash
isrvd_get "/account/routes"
isrvd_get "/account/routes" '.[] | {key, module, label, access}'
```

> access: `0`=需权限, `1`=需登录, `2`=匿名

## 创建 API Token

```bash
isrvd_post "/account/token" '{"name":"<TOKEN_NAME>","expiresIn":"720h"}'
```

返回：`{"token": "长效token..."}`

## 修改密码

```bash
isrvd_put "/account/password" '{"oldPassword":"<OLD>","newPassword":"<NEW>"}'
```

## 列出成员

```bash
isrvd_get "/account/members"
isrvd_get "/account/members" '.[] | {username, founder, permissions: (.permissions | length)}'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| username | string | 用户名 |
| homeDirectory | string | 主目录 |
| founder | boolean | 是否为创建者 |
| description | string | 描述 |
| permissions | string[] | 权限列表 |

## 创建成员

```bash
isrvd_post "/account/member" '{"username":"<USER>","password":"<PASS>","homeDirectory":"<HOME_DIR>","description":"<DESC>","permissions":["GET /api/docker/containers","GET /api/docker/images"]}'
```

## 更新成员

```bash
isrvd_put "/account/member/<USER>" '{"description":"<DESC>","permissions":["GET /api/docker/containers","GET /api/docker/images","GET /api/swarm/services"]}'
```

> password 为空则不修改。

## 删除成员

```bash
isrvd_delete "/account/member/<USER>"
```
