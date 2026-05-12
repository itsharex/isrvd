// Package account 账号与认证业务模块
// 提供用户认证、登录、成员管理等功能
package account

import (
	"sync"
	"time"
)

// Service 账号业务服务
type Service struct {
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
	// 后台定期清理过期的 OIDC state 和 loginCode，避免在每次操作时全量扫描
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			s.cleanupOIDC()
		}
	}()
	return s
}
