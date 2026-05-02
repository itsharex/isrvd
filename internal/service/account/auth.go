package account

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

// ExtractJwtUsername 从 Authorization Header（或 WebSocket query）中解析 JWT，
// 返回有效且存在于成员列表中的用户名；否则返回空字符串
func (s *Service) ExtractJwtUsername(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	// WebSocket 连接时允许从 query 参数获取 token
	if tokenStr == "" && c.GetHeader("Upgrade") == "websocket" {
		tokenStr = c.Query("token")
	}
	if tokenStr == "" {
		return ""
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	sub, _ := claims["sub"].(string)
	if _, exists := config.Members[sub]; !exists {
		return ""
	}
	return sub
}

// ExtractHeaderUsername 从代理 Header 中读取用户名，
// 返回存在于成员列表中的用户名；否则返回空字符串
func (s *Service) ExtractHeaderUsername(c *gin.Context) string {
	username := c.GetHeader(config.ProxyHeaderName)
	if username == "" {
		return ""
	}
	if _, exists := config.Members[username]; !exists {
		return ""
	}
	return username
}
