package server

import (
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rehiy/pango/logman"

	"isrvd/config"
	"isrvd/internal/helper"
)

// defineFilerRoutes 定义 Filer 模块路由（文件管理）
func (app *App) defineFilerRoutes() []Route {
	return []Route{
		{Method: "POST", Path: "/filer/list", Handler: app.filerFileList, Module: "filer", Label: "列出文件"},
		{Method: "POST", Path: "/filer/mkdir", Handler: app.filerFileMkdir, Module: "filer", Label: "创建目录"},
		{Method: "POST", Path: "/filer/create", Handler: app.filerFileCreate, Module: "filer", Label: "创建文件"},
		{Method: "POST", Path: "/filer/read", Handler: app.filerFileRead, Module: "filer", Label: "读取文件"},
		{Method: "POST", Path: "/filer/modify", Handler: app.filerFileModify, Module: "filer", Label: "保存文件"},
		{Method: "POST", Path: "/filer/rename", Handler: app.filerFileRename, Module: "filer", Label: "重命名文件"},
		{Method: "POST", Path: "/filer/delete", Handler: app.filerFileDelete, Module: "filer", Label: "删除文件"},
		{Method: "POST", Path: "/filer/chmod", Handler: app.filerFileChmod, Module: "filer", Label: "修改文件权限"},
		{Method: "POST", Path: "/filer/upload", Handler: app.filerFileUpload, Module: "filer", Label: "上传文件"},
		{Method: "POST", Path: "/filer/download", Handler: app.filerFileDownload, Module: "filer", Label: "下载文件"},
		{Method: "POST", Path: "/filer/zip", Handler: app.filerFileZip, Module: "filer", Label: "压缩文件"},
		{Method: "POST", Path: "/filer/unzip", Handler: app.filerFileUnzip, Module: "filer", Label: "解压文件"},
	}
}

// ─── 请求结构 ───

type filerPathReq struct {
	Path string `json:"path" binding:"required"`
}

type filerContentReq struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type filerChmodReq struct {
	Path string `json:"path" binding:"required"`
	Mode string `json:"mode" binding:"required"`
}

type filerRenameReq struct {
	Path   string `json:"path" binding:"required"`
	Target string `json:"target" binding:"required"`
}

// ─── Handler 方法 ───

func (app *App) filerFileList(c *gin.Context) {
	var req filerPathReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	username := c.GetString("username")
	absPath := app.filerSvc.AbsPath(username, req.Path)
	files, err := app.filerSvc.FileList(absPath, req.Path)
	if err != nil {
		logman.Error("List files failed", "path", absPath, "error", err)
		helper.RespondError(c, http.StatusNotFound, "Directory not found")
		return
	}
	helper.RespondSuccess(c, "Files listed successfully", gin.H{"path": req.Path, "files": files})
}

func (app *App) filerFileDelete(c *gin.Context) {
	var req filerPathReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	absPath := app.filerSvc.AbsPath(c.GetString("username"), req.Path)
	if err := app.filerSvc.FileDelete(absPath); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, "Cannot delete file")
		return
	}
	helper.RespondSuccess(c, "File deleted successfully", nil)
}

func (app *App) filerFileMkdir(c *gin.Context) {
	var req filerPathReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	absPath := app.filerSvc.AbsPath(c.GetString("username"), req.Path)
	if err := app.filerSvc.FileMkdir(absPath); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, "Cannot create directory")
		return
	}
	helper.RespondSuccess(c, "Directory created successfully", nil)
}

func (app *App) filerFileCreate(c *gin.Context) {
	var req filerContentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	absPath := app.filerSvc.AbsPath(c.GetString("username"), req.Path)
	if err := app.filerSvc.FileCreate(absPath, []byte(req.Content)); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, "Cannot create file")
		return
	}
	helper.RespondSuccess(c, "File created successfully", nil)
}

func (app *App) filerFileRead(c *gin.Context) {
	var req filerPathReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	absPath := app.filerSvc.AbsPath(c.GetString("username"), req.Path)
	content, err := app.filerSvc.FileRead(absPath)
	if err != nil {
		helper.RespondError(c, http.StatusNotFound, "File not found")
		return
	}
	helper.RespondSuccess(c, "File content retrieved", gin.H{"path": req.Path, "content": string(content)})
}

func (app *App) filerFileModify(c *gin.Context) {
	var req filerContentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	absPath := app.filerSvc.AbsPath(c.GetString("username"), req.Path)
	if err := app.filerSvc.FileWrite(absPath, []byte(req.Content)); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, "Cannot save file")
		return
	}
	helper.RespondSuccess(c, "File saved successfully", nil)
}

func (app *App) filerFileRename(c *gin.Context) {
	var req filerRenameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	username := c.GetString("username")
	absPath := app.filerSvc.AbsPath(username, req.Path)
	targetPath := app.filerSvc.AbsPath(username, filepath.Join(filepath.Dir(req.Path), req.Target))
	if err := app.filerSvc.FileRename(absPath, targetPath); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, "Cannot rename file")
		return
	}
	helper.RespondSuccess(c, "File renamed successfully", nil)
}

func (app *App) filerFileChmod(c *gin.Context) {
	var req filerChmodReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	absPath := app.filerSvc.AbsPath(c.GetString("username"), req.Path)
	if err := app.filerSvc.FileChmod(absPath, req.Mode); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	helper.RespondSuccess(c, "Permissions changed successfully", nil)
}

func (app *App) filerFileUpload(c *gin.Context) {
	if c.Request.ContentLength > config.Server.MaxUploadSize {
		helper.RespondError(c, http.StatusBadRequest, "文件大小超过限制")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.Server.MaxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		helper.RespondError(c, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	username := c.GetString("username")
	path := c.PostForm("path")
	var absPath string
	if path == "" {
		absPath = app.filerSvc.AbsPath(username, header.Filename)
	} else {
		absPath = app.filerSvc.AbsPath(username, filepath.Join(path, header.Filename))
	}

	data, err := io.ReadAll(file)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, "Cannot read uploaded file")
		return
	}
	if err := app.filerSvc.FileWrite(absPath, data); err != nil {
		helper.RespondError(c, http.StatusInternalServerError, "Cannot write file")
		return
	}
	helper.RespondSuccess(c, "File uploaded successfully", nil)
}

func (app *App) filerFileDownload(c *gin.Context) {
	var req filerPathReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	absPath := app.filerSvc.AbsPath(c.GetString("username"), req.Path)
	content, err := app.filerSvc.FileRead(absPath)
	if err != nil {
		helper.RespondError(c, http.StatusNotFound, "File not found")
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(req.Path))
	if _, err := c.Writer.Write(content); err != nil {
		logman.Error("Download write failed", "path", absPath, "error", err)
	}
}

func (app *App) filerFileZip(c *gin.Context) {
	var req filerPathReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	absPath := app.filerSvc.AbsPath(c.GetString("username"), req.Path)
	if err := app.filerSvc.FileZip(absPath); err != nil {
		logman.Error("Create zip failed", "path", absPath, "error", err)
		helper.RespondError(c, http.StatusInternalServerError, "无法创建压缩文件")
		return
	}
	helper.RespondSuccess(c, "Archive created successfully", nil)
}

func (app *App) filerFileUnzip(c *gin.Context) {
	var req filerPathReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	absPath := app.filerSvc.AbsPath(c.GetString("username"), req.Path)
	if err := app.filerSvc.FileUnzip(absPath); err != nil {
		logman.Error("Unzip failed", "path", absPath, "error", err)
		helper.RespondError(c, http.StatusInternalServerError, "无法解压文件")
		return
	}
	helper.RespondSuccess(c, "Archive extracted successfully", nil)
}
