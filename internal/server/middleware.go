package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"isrvd/internal/helper"
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
			if username := svc.MixAuth(c); username != "" {
				c.Set("username", username)
			}
			c.Next()
			return
		}

		username, errMsg := svc.Auth(c)
		if username == "" {
			helper.RespondError(c, http.StatusUnauthorized, errMsg)
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
			helper.RespondError(c, http.StatusForbidden, "未授权的访问路径")
			c.Abort()
			return
		}

		found, err := svc.CheckRoutePerm(routePerms, c.Request.Method, path, c.GetString("username"))
		if !found || err != nil {
			msg := "未授权的访问路径"
			if err != nil {
				msg = err.Error()
			}
			helper.RespondError(c, http.StatusForbidden, msg)
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
			body = svc.ReadRequestBody(c)
		}
		c.Next()
		svc.RecordRequest(c, startTime, body)
	}
}
