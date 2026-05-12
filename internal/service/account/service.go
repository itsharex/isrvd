// Package account 账号与认证业务模块
// 提供用户认证、登录、成员管理等功能
package account

import (
	"fmt"
	"isrvd/config"
	"slices"
	"sync"
	"time"
)

// Service 账号业务服务
type Service struct {
	// OIDC 临时状态存储（state/loginCode 均短期有效，内存存储即可）
	oidcMu         sync.Mutex
	oidcStates     map[string]oidcState
	oidcLoginCodes map[string]oidcLoginCode
	oidcProvider   oidcProviderCache
}

// NewService 创建账号业务服务
func NewService() *Service {
	s := &Service{
		oidcStates:     make(map[string]oidcState),
		oidcLoginCodes: make(map[string]oidcLoginCode),
	}
	// 后台定期清理过期的 OIDC 临时状态
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			s.cleanupOIDC()
		}
	}()
	return s
}

// ─── 路由权限校验 ───

// RouteAccess 路由访问级别（数值越大，要求越高）
type RouteAccess int

const (
	AccessPerm RouteAccess = iota // 0：需要具体权限
	AccessAuth                    // 1：登录即可访问
	AccessAnon                    // 2：匿名，无需认证
)

// RouteInfo 路由的权限元信息，供中间件做权限校验
type RouteInfo struct {
	Key    string      `json:"key"`    // "METHOD /api/path"
	Module string      `json:"module"` // 模块名，空字符串表示无需模块权限校验
	Label  string      `json:"label"`  // 显示名，用于错误提示
	Access RouteAccess `json:"access"` // 访问级别
}

// RoutePermCheck 在路由权限表中查找当前请求对应的路由，并校验用户权限。
// 返回 (found bool, err error)：found=false 表示路由未注册，err 非 nil 表示权限不足。
func (s *Service) RoutePermCheck(routePerms map[string]RouteInfo, method, path, username string) (found bool, err error) {
	route, exists := routePerms[method+" "+path]
	if !exists {
		route, exists = routePerms["ANY "+path]
	}
	if !exists {
		return false, nil
	}
	if route.Access == AccessAnon || route.Module == "" {
		return true, nil
	}
	if route.Access == AccessAuth {
		if username == "" {
			return true, fmt.Errorf("请先登录")
		}
		return true, nil
	}
	return true, s.PermCheck(username, route.Label, method, path)
}

// PermCheck 校验用户是否有权访问指定路由（"METHOD /api/path"）。
// label 用于错误提示。返回 nil 表示有权限，否则返回描述错误原因的 error。
func (s *Service) PermCheck(username, label, method, path string) error {
	member, exists := config.Members[username]
	if !exists {
		return fmt.Errorf("用户不存在")
	}
	if member.Founder {
		return nil
	}
	routeKey := method + " " + path
	if slices.Contains(member.Permissions, routeKey) {
		return nil
	}
	if label == "" {
		label = routeKey
	}
	return fmt.Errorf("无 %s 访问权限", label)
}
