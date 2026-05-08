package account

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rehiy/pango/logman"

	"isrvd/config"
	"isrvd/internal/helper"
)

// RouteAccess 路由访问级别（数值越大，要求越高）
type RouteAccess int

const (
	AccessPerm RouteAccess = iota // 0：权限控制，需要登录
	AccessAuth                    // 1：已认证，需要登录
	AccessAnon                    // 2：匿名，无需认证
)

// RouteInfo 路由的权限元信息，供中间件做权限校验
type RouteInfo struct {
	Key    string      `json:"key"`    // "METHOD /api/path"
	Module string      `json:"module"` // 模块名，空字符串表示无需模块权限校验
	Label  string      `json:"label"`  // 显示名，用于错误提示
	Access RouteAccess `json:"access"` // 访问级别：0=perm（默认）/ 1=auth / 2=anon
}

// Auth 根据配置选择认证方式，返回用户名和错误原因。
// 供中间件统一调用，避免在 server 层判断认证模式。
func (s *Service) Auth(c *gin.Context) (username, errMsg string) {
	if config.ProxyHeaderName != "" {
		return s.HeaderTokenCheck(c)
	}
	return s.JwtTokenCheck(c)
}

// AuthMix 可选认证：成功返回用户名，失败返回空字符串（不中断请求）。
func (s *Service) AuthMix(c *gin.Context) string {
	username, _ := s.Auth(c)
	return username
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
	// 匿名路由或无模块限制的路由，无需权限校验
	if route.Access == AccessAnon || route.Module == "" {
		return true, nil
	}
	// 已认证路由：只需登录，无需特定权限
	if route.Access == AccessAuth {
		if username == "" {
			return true, fmt.Errorf("请先登录")
		}
		return true, nil
	}
	return true, s.PermCheck(username, route.Label, method, path)
}

// AuthInfo 返回当前认证模式及已登录用户信息
func (s *Service) AuthInfo(username string) *AuthInfoResponse {
	mode := "jwt"
	if config.ProxyHeaderName != "" {
		mode = "header"
	}
	return &AuthInfoResponse{
		Mode:     mode,
		Username: username,
		Member:   s.MemberInspect(username),
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

	// 密码 hash 后 8 位作为校验，修改密码后 token 自动失效
	// 注意：bcrypt hash 前 7 位是固定格式（如 $2a$10$），后 8 位会随每次密码重置而变化
	pwd := ""
	if len(member.Password) >= 8 {
		pwd = member.Password[len(member.Password)-8:]
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": req.Username,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"pwd": pwd,
	})
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("token 生成失败: %w", err)
	}

	logman.Info("User logged in", "username", req.Username)
	return &LoginResponse{Token: tokenString, Username: req.Username}, nil
}

// CreateApiTokenRequest 创建 API Token 请求
type CreateApiTokenRequest struct {
	Name      string `json:"name"`      // 令牌名称（用于标识）
	ExpiresIn int64  `json:"expiresIn"` // 过期时间（秒），0 表示永不过期
}

// CreateApiTokenResponse 创建 API Token 响应
type CreateApiTokenResponse struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

// ApiTokenCreate 为已认证用户创建长效 API Token
func (s *Service) ApiTokenCreate(username string, req CreateApiTokenRequest) (*CreateApiTokenResponse, error) {
	member, exists := config.Members[username]
	if !exists {
		return nil, fmt.Errorf("用户不存在")
	}

	// 密码 hash 后 8 位作为校验，修改密码后 token 自动失效
	// 注意：bcrypt hash 前 7 位是固定格式（如 $2a$10$），后 8 位会随每次密码重置而变化
	pwd := ""
	if len(member.Password) >= 8 {
		pwd = member.Password[len(member.Password)-8:]
	}

	claims := jwt.MapClaims{
		"sub":  username,
		"iat":  time.Now().Unix(),
		"type": "api", // 标记为 API token
		"name": req.Name,
		"pwd":  pwd,
	}
	if req.ExpiresIn > 0 {
		claims["exp"] = time.Now().Add(time.Duration(req.ExpiresIn) * time.Second).Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("token 生成失败: %w", err)
	}

	logman.Info("API token created", "username", username, "name", req.Name)
	return &CreateApiTokenResponse{Token: tokenString, Name: req.Name}, nil
}

// JwtUsernameExtract 从 Authorization Header（或 WebSocket query）中解析 JWT，
// 返回有效且存在于成员列表中的用户名；否则返回空字符串
func (s *Service) JwtUsernameExtract(c *gin.Context) string {
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
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
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
	member, exists := config.Members[sub]
	if !exists {
		return ""
	}

	// 校验密码 hash 后 8 位（修改密码后自动失效）
	pwd, _ := claims["pwd"].(string)
	if pwd != "" && len(member.Password) >= 8 {
		if pwd != member.Password[len(member.Password)-8:] {
			return ""
		}
	}

	return sub
}

// HeaderUsernameExtract 从代理 Header 中读取用户名，
// 返回存在于成员列表中的用户名；否则返回空字符串
func (s *Service) HeaderUsernameExtract(c *gin.Context) string {
	username := c.GetHeader(config.ProxyHeaderName)
	if username == "" {
		return ""
	}
	if _, exists := config.Members[username]; !exists {
		return ""
	}
	return username
}

// JwtTokenCheck 解析 JWT 并返回用户名；失败时返回空用户名和具体错误原因。
// errMsg 区分"未提供令牌"与"令牌无效"两种情况。
func (s *Service) JwtTokenCheck(c *gin.Context) (username, errMsg string) {
	tokenStr := s.extractJwtToken(c)
	if tokenStr == "" {
		return "", "未提供认证令牌"
	}
	username = s.JwtUsernameExtract(c)
	if username == "" {
		return "", "认证令牌无效"
	}
	return username, ""
}

// extractJwtToken 从 Authorization Header 或 WebSocket query 中提取原始 token 字符串。
func (s *Service) extractJwtToken(c *gin.Context) string {
	tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if tokenStr == "" && c.GetHeader("Upgrade") == "websocket" {
		tokenStr = c.Query("token")
	}
	return tokenStr
}

// HeaderTokenCheck 从代理 Header 读取用户名；失败时返回空用户名和具体错误原因。
// errMsg 区分"Header 缺失"与"用户不存在"两种情况。
func (s *Service) HeaderTokenCheck(c *gin.Context) (username, errMsg string) {
	raw := c.GetHeader(config.ProxyHeaderName)
	if raw == "" {
		return "", "代理 Header 缺失"
	}
	username = s.HeaderUsernameExtract(c)
	if username == "" {
		return "", "用户不存在"
	}
	return username, ""
}

// PermCheck 校验用户是否有权访问指定路由（"METHOD /api/path"）。
// label 用于错误提示。返回 nil 表示有权限，否则返回描述错误原因的 error。
func (s *Service) PermCheck(username, label, method, path string) error {
	member, exists := config.Members[username]
	if !exists {
		return fmt.Errorf("用户不存在")
	}
	// 创始人拥有所有权限
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
