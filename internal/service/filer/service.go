// Package filer 文件管理业务服务
package filer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rehiy/pango/filer"

	"isrvd/config"
	"isrvd/pkgs/archive"
)

// Service 文件管理业务服务
type Service struct{}

// NewService 创建文件管理业务服务
func NewService() *Service {
	return &Service{}
}

// FileInfo 文件信息
type FileInfo struct {
	Path    string    `json:"path"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"isDir"`
	Mode    string    `json:"mode"`
	ModeO   string    `json:"modeO"`
	ModTime time.Time `json:"modTime"`
}

// AbsPath 解析用户相对路径为绝对路径，并防止目录遍历
func (s *Service) AbsPath(username, path string) string {
	home := filepath.Clean(filepath.Join(config.RootDirectory, "share"))
	if username != "" {
		if member, ok := config.Members[username]; ok {
			home = filepath.Clean(member.HomeDirectory)
		}
	}
	abs := filepath.Clean(filepath.Join(home, path))
	rel, err := filepath.Rel(home, abs)
	if err == nil && rel != ".." && !filepath.IsAbs(rel) && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return abs
	}
	return home
}

// FileList 列出目录下的文件
func (s *Service) FileList(absPath, relPath string) ([]*FileInfo, error) {
	list, err := filer.List(absPath)
	if err != nil {
		return nil, err
	}
	var result []*FileInfo
	for _, f := range list {
		p := filepath.ToSlash(filepath.Join(relPath, f.Name))
		result = append(result, &FileInfo{
			Path:    p,
			Name:    f.Name,
			Size:    f.Size,
			IsDir:   f.IsDir,
			Mode:    f.Mode.String(),
			ModeO:   strconv.FormatInt(int64(f.Mode), 8),
			ModTime: time.Unix(f.ModTime, 0),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir && !result[j].IsDir {
			return true
		}
		if !result[i].IsDir && result[j].IsDir {
			return false
		}
		return result[i].Name < result[j].Name
	})
	return result, nil
}

// FileRead 读取文件内容
func (s *Service) FileRead(absPath string) ([]byte, error) {
	return os.ReadFile(absPath)
}

// FileWrite 写入文件内容（覆盖）
func (s *Service) FileWrite(absPath string, content []byte) error {
	return os.WriteFile(absPath, content, 0644)
}

// FileCreate 创建文件（使用 pango/filer）
func (s *Service) FileCreate(absPath string, content []byte) error {
	return filer.Write(absPath, content)
}

// FileMkdir 创建目录
func (s *Service) FileMkdir(absPath string) error {
	return os.Mkdir(absPath, 0755)
}

// FileDelete 删除文件或目录
func (s *Service) FileDelete(absPath string) error {
	return os.RemoveAll(absPath)
}

// FileRename 重命名文件
func (s *Service) FileRename(absPath, targetPath string) error {
	return os.Rename(absPath, targetPath)
}

// FileChmod 修改文件权限
func (s *Service) FileChmod(absPath string, modeStr string) error {
	mode, err := strconv.ParseUint(modeStr, 8, 32)
	if err != nil {
		return fmt.Errorf("无效的权限值: %w", err)
	}
	return os.Chmod(absPath, os.FileMode(mode))
}

// FileZip 压缩文件或目录
func (s *Service) FileZip(absPath string) error {
	return archive.NewZipper().Zip(absPath)
}

// FileUnzip 解压 zip 文件
func (s *Service) FileUnzip(absPath string) error {
	return archive.NewZipper().Unzip(absPath)
}
