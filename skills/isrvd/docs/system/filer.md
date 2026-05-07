# 文件管理 API

> 均为 POST 方法

## 列出文件

```bash
isrvd_post "/filer/list" '{"path":"/data"}'
```

## 创建目录

```bash
isrvd_post "/filer/mkdir" '{"path":"/data/newdir"}'
```

## 创建文件

```bash
isrvd_post "/filer/create" '{"path":"/data/file.txt"}'
```

## 读取文件

```bash
isrvd_post "/filer/read" '{"path":"/data/file.txt"}'
isrvd_post "/filer/read" '{"path":"/data/config.yml"}' '.content'
```

返回：`{"content": "文件内容..."}`

## 保存文件

```bash
isrvd_post "/filer/modify" '{"path":"/data/file.txt","content":"新内容"}'
```

## 重命名

```bash
isrvd_post "/filer/rename" '{"oldPath":"/data/old.txt","newPath":"/data/new.txt"}'
```

## 删除

```bash
isrvd_post "/filer/delete" '{"path":"/data/file.txt"}'
```

## 修改权限

```bash
isrvd_post "/filer/chmod" '{"path":"/data/file.txt","mode":"0644"}'
```

## 上传文件

```bash
isrvd_upload "/filer/upload" "file" "./local-file.tar.gz" "path=/data"
```

## 下载文件

```bash
isrvd_post "/filer/download" '{"path":"/data/file.tar.gz"}'
```

返回文件二进制流。

## 压缩

```bash
isrvd_post "/filer/zip" '{"path":"/data/mydir","dest":"/data/mydir.zip"}'
```

## 解压

```bash
isrvd_post "/filer/unzip" '{"path":"/data/mydir.zip","dest":"/data/extracted"}'
```
