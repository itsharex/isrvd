package account

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rehiy/pango/logman"

	"isrvd/config"
	"isrvd/internal/helper"
)

// GetAuthInfo 返回当前认证模式及已登录用户信息
func (s *Service) GetAuthInfo(username string) *AuthInfoResponse {
	mode := "jwt"
	if config.ProxyHeaderName != "" {
		mode = "header"
	}
	return &AuthInfoResponse{
		Mode:     mode,
		Username: username,
		Member:   s.GetMember(username),
	}
}

// AuthInfoResponse 认证模式及当前用户信息
type AuthInfoResponse struct {
	Mode     string      `json:"mode"`
	Username string      `json:"username,omitempty"`
	Member   *MemberInfo `json:"member,omitempty"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// Login 校验用户名密码并签发 JWT Token
func (s *Service) Login(req LoginRequest) (*LoginResponse, error) {
	member, exists := config.Members[req.Username]
	if !exists || !helper.VerifyPassword(req.Password, member.Password) {
		logman.Warn("Login failed", "username", req.Username)
		return nil, fmt.Errorf("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": req.Username,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("token 生成失败: %w", err)
	}

	logman.Info("User logged in", "username", req.Username)
	return &LoginResponse{Token: tokenString, Username: req.Username}, nil
}
