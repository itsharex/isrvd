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

// AuthMiddleware 认证中间件：认证失败时中断请求。
// anon 路由走可选认证（认证失败时放行），其余路由强制认证。
func AuthMiddleware(routePerms map[string]svcAccount.RouteInfo, svc *svcAccount.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// anon 路由：可选认证，不强制
		if info, ok := routePerms[c.Request.Method+" "+c.FullPath()]; ok && info.Access == svcAccount.AccessAnon {
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

// MixAuthMiddleware 可选认证中间件：认证成功时写入 username，失败时直接放行
func MixAuthMiddleware(svc *svcAccount.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if username := svc.MixAuth(c); username != "" {
			c.Set("username", username)
		}
		c.Next()
	}
}

// RoutePermMiddleware 基于 METHOD+PATH 的集中式权限验证中间件
func RoutePermMiddleware(routePerms map[string]svcAccount.RouteInfo, svc *svcAccount.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			helper.RespondError(c, http.StatusForbidden, "未授权的访问路径")
			c.Abort()
			return
		}

		found, err := svc.CheckRoutePerm(routePerms, c.Request.Method, path, c.GetString("username"))
		if !found {
			helper.RespondError(c, http.StatusForbidden, "未授权的访问路径")
			c.Abort()
			return
		}
		if err != nil {
			helper.RespondError(c, http.StatusForbidden, err.Error())
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuditMiddleware 返回操作审计中间件。
// 记录所有非 GET 请求及 WebSocket 连接的方法、URI、请求体、状态码和耗时。
func AuditMiddleware(svc *svcSystem.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		isWS := strings.EqualFold(c.GetHeader("Upgrade"), "websocket")

		// GET 请求（非 WebSocket）不记录审计
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
