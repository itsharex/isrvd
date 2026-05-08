# 文件管理 API

> 均为 POST 方法

## ⚠️ 重要：filer 路径 ≠ 宿主机路径

filer 管理的是 **isrvd 容器内部**挂载的卷，这些路径只在 isrvd 容器内有效。

如果要把这些目录 volume mount 给其他容器，**不能直接用 filer 路径作为 hostPath**，必须先找出它们在宿主机上的真实路径：

```bash
# 查看 isrvd 容器自身的 volume 挂载，找到 filer 路径对应的宿主机路径
# 1. 从 isrvd_get "/docker/containers" 结果中找到 isrvd 容器的 id
# 2. 用 isrvd_get "/docker/container/<ISRVD_ID>" 查看 mounts 字段
```

**正确的静态文件更新流程**（不需要重建容器）：
1. 先查出 filer 可用目录：`isrvd_post "/filer/list" '{"path":"/"}'`
2. 用 filer 写文件：`isrvd_post "/filer/modify" '{"path":"<FILER_PATH>/<FILE>","content":"..."}'`
3. 确认 web 容器是否已经用正确的 hostPath 挂载了该目录（只在首次部署时配置一次）
4. 文件写入后立即生效，无需重启

**禁止用 base64**：不要用 `base64 -d` 方式在 Dockerfile 或 RUN 命令里写入文件内容。应使用 `isrvd_upload` 或直接 `isrvd_post "/filer/modify"` 写入。

## 列出文件

```bash
isrvd_post "/filer/list" '{"path":"<DIR>"}'
```

## 创建目录

```bash
isrvd_post "/filer/mkdir" '{"path":"<DIR>"}'
```

## 创建文件

```bash
isrvd_post "/filer/create" '{"path":"<FILE>"}'
```

## 读取文件

```bash
isrvd_post "/filer/read" '{"path":"<FILE>"}'
```

返回：`{"content": "文件内容..."}`

## 保存文件

```bash
isrvd_post "/filer/modify" '{"path":"<FILE>","content":"<CONTENT>"}'
```

## 重命名

```bash
isrvd_post "/filer/rename" '{"oldPath":"<OLD_PATH>","newPath":"<NEW_PATH>"}'
```

## 删除

```bash
isrvd_post "/filer/delete" '{"path":"<FILE>"}'
```

## 修改权限

```bash
isrvd_post "/filer/chmod" '{"path":"<FILE>","mode":"0644"}'
```

## 上传文件

```bash
isrvd_upload "/filer/upload" "file" "<LOCAL_FILE>" "path=<FILER_DIR>"
```

## 下载文件

```bash
isrvd_post "/filer/download" '{"path":"<FILE>"}'
```

返回文件二进制流。

## 压缩

```bash
isrvd_post "/filer/zip" '{"path":"<DIR>","dest":"<OUTPUT>.zip"}'
```

## 解压

```bash
isrvd_post "/filer/unzip" '{"path":"<FILE>.zip","dest":"<DEST_DIR>"}'
```
