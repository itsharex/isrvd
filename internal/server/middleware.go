package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	
	svcAccount "isrvd/internal/service/account"
	svcSystem "isrvd/internal/service/system"
)

// AuthMiddleware 认证中间件
// - AccessAnon 路由：可选认证，失败时放行
// - 其他路由：强制认证，失败时返回 401
func AuthMiddleware(routePerms map[string]svcAccount.RouteInfo, svc *svcAccount.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Request.Method + " " + c.FullPath()
		if info, ok := routePerms[key]; ok && info.Access == svcAccount.AccessAnon {
			if username := svc.AuthMix(c); username != "" {
				c.Set("username", username)
			}
			c.Next()
			return
		}

		username, errMsg := svc.Auth(c)
		if username == "" {
			respondError(c, http.StatusUnauthorized, errMsg)
			c.Abort()
			return
		}
		c.Set("username", username)
		c.Next()
	}
}

// PermMiddleware 权限验证中间件
// 基于 METHOD+PATH 进行集中式权限校验
func PermMiddleware(routePerms map[string]svcAccount.RouteInfo, svc *svcAccount.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			respondError(c, http.StatusForbidden, "未授权的访问路径")
			c.Abort()
			return
		}

		found, err := svc.RoutePermCheck(routePerms, c.Request.Method, path, c.GetString("username"))
		if !found || err != nil {
			msg := "未授权的访问路径"
			if err != nil {
				msg = err.Error()
			}
			respondError(c, http.StatusForbidden, msg)
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuditMiddleware 操作审计中间件
// 记录所有非 GET 请求及 WebSocket 连接
func AuditMiddleware(svc *svcSystem.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		isWS := strings.EqualFold(c.GetHeader("Upgrade"), "websocket")
		if !isWS && c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		startTime := time.Now()
		var body string
		if !isWS {
			body = svc.RequestBodyRead(c)
		}

		c.Next()
		svc.RequestRecord(c, startTime, body)
	}
}

// securityHeadersMiddleware 安全响应头中间件
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")
		// 防止 MIME 类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		// XSS 保护
		c.Header("X-XSS-Protection", "1; mode=block")
		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// CSP：限制资源加载和脚本执行，降低 XSS 风险
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'")

		c.Next()
	}
}
