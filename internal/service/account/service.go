// Package account 账号与认证业务模块
// 提供用户认证、登录、成员管理等功能
package account

// Service 账号业务服务
type Service struct{}

// NewService 创建账号业务服务
func NewService() *Service {
	return &Service{}
}
