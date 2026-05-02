package account

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rehiy/pango/logman"

	"isrvd/config"
	"isrvd/internal/helper"
)

// 哨兵错误，供 handler 层进行错误类型判断
var (
	ErrMemberNotFound = errors.New("成员不存在")
	ErrMemberExists   = errors.New("用户名已存在")
	ErrInvalidRequest = errors.New("用户名不能为空")
)

// GetMember 获取单个成员信息
func (s *Service) GetMember(username string) *MemberInfo {
	m, exists := config.Members[username]
	if !exists {
		return nil
	}
	return s.buildMemberInfo(m)
}

// MemberInfo 成员信息（不包含密码明文）
type MemberInfo struct {
	Username      string            `json:"username"`
	HomeDirectory string            `json:"homeDirectory"`
	PasswordSet   bool              `json:"passwordSet"`
	Permissions   map[string]string `json:"permissions"`
}

// ListMembers 列出所有成员
func (s *Service) ListMembers() []*MemberInfo {
	list := make([]*MemberInfo, 0, len(config.Members))
	for _, m := range config.Members {
		list = append(list, s.buildMemberInfo(m))
	}
	return list
}

// buildMemberInfo 从配置构建成员信息（确保权限不为 nil）
func (s *Service) buildMemberInfo(m *config.MemberConfig) *MemberInfo {
	perms := m.Permissions
	if perms == nil {
		perms = map[string]string{}
	}
	return &MemberInfo{
		Username:      m.Username,
		HomeDirectory: m.HomeDirectory,
		PasswordSet:   m.Password != "",
		Permissions:   perms,
	}
}

// ensureHomeDir 生成并创建成员 home 目录（空值时使用基础目录 + 用户名）
func (s *Service) ensureHomeDir(home, username string) (string, error) {
	if home == "" {
		home = username
	}
	if !filepath.IsAbs(home) {
		home = filepath.Join(config.RootDirectory, home)
	}
	if err := os.MkdirAll(home, 0755); err != nil {
		return "", err
	}
	return home, nil
}

// MemberUpsertRequest 成员新建/更新请求
type MemberUpsertRequest struct {
	Username      string            `json:"username"`
	Password      string            `json:"password"`
	HomeDirectory string            `json:"homeDirectory"`
	Permissions   map[string]string `json:"permissions"`
}

// CreateMember 新建成员
func (s *Service) CreateMember(req MemberUpsertRequest) error {
	if req.Username == "" {
		return ErrInvalidRequest
	}
	if _, exists := config.Members[req.Username]; exists {
		return ErrMemberExists
	}

	home, err := s.ensureHomeDir(req.HomeDirectory, req.Username)
	if err != nil {
		return fmt.Errorf("创建 home 目录失败: %w", err)
	}

	// 对密码进行 bcrypt 加密
	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	config.Members[req.Username] = &config.MemberConfig{
		Username:      req.Username,
		Password:      hashedPassword,
		HomeDirectory: home,
		Permissions:   req.Permissions,
	}
	if err := config.Save(); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	logman.Info("Member created", "username", req.Username)
	return nil
}

// UpdateMember 更新成员
func (s *Service) UpdateMember(username string, req MemberUpsertRequest) error {
	member, exists := config.Members[username]
	if !exists {
		return ErrMemberNotFound
	}

	home, err := s.ensureHomeDir(req.HomeDirectory, username)
	if err != nil {
		return fmt.Errorf("创建 home 目录失败: %w", err)
	}

	// 密码为空时 HashPassword 返回空，保持原密码不变
	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	if hashedPassword != "" {
		member.Password = hashedPassword
	}

	member.HomeDirectory = home
	member.Permissions = req.Permissions

	if err := config.Save(); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	logman.Info("Member updated", "username", username)
	return nil
}

// DeleteMember 删除成员
func (s *Service) DeleteMember(username string) error {
	if _, exists := config.Members[username]; !exists {
		return ErrMemberNotFound
	}
	delete(config.Members, username)
	if err := config.Save(); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	logman.Info("Member deleted", "username", username)
	return nil
}
