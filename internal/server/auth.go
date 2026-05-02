package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"isrvd/config"
	"isrvd/internal/helper"
	svcAccount "isrvd/internal/service/account"
)

// 认证中间件工厂
func AuthMiddleware(svc *svcAccount.Service) gin.HandlerFunc {
	if config.ProxyHeaderName == "" {
		return JwtAuthMiddleware(svc)
	}
	return HeaderAuthMiddleware(svc)
}

// MixAuthMiddleware 可选认证中间件
// 认证成功时写入 username，失败时直接放行（不中断请求）
// 认证模式在工厂函数调用时确定，避免每次请求重复判断静态配置
func MixAuthMiddleware(svc *svcAccount.Service) gin.HandlerFunc {
	if config.ProxyHeaderName != "" {
		// Header 认证模式
		return func(c *gin.Context) {
			if username := svc.ExtractHeaderUsername(c); username != "" {
				c.Set("username", username)
			}
			c.Next()
		}
	}
	// JWT 认证模式
	return func(c *gin.Context) {
		if username := svc.ExtractJwtUsername(c); username != "" {
			c.Set("username", username)
		}
		c.Next()
	}
}

// JWT 认证中间件
func JwtAuthMiddleware(svc *svcAccount.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := svc.ExtractJwtUsername(c)
		if username == "" {
			// 区分"未提供 token"与"token 无效"两种情况给出不同提示
			authHeader := c.GetHeader("Authorization")
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == "" && c.GetHeader("Upgrade") != "websocket" {
				helper.RespondError(c, http.StatusUnauthorized, "未提供认证令牌")
			} else {
				helper.RespondError(c, http.StatusUnauthorized, "认证令牌无效")
			}
			c.Abort()
			return
		}
		c.Set("username", username)
		c.Next()
	}
}

// 内网代理 Header 认证中间件
// 启用条件：config.ProxyHeaderName 非空
// Header 缺失或用户不存在时返回 403，不回退到 JWT
func HeaderAuthMiddleware(svc *svcAccount.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := svc.ExtractHeaderUsername(c)
		if username == "" {
			if c.GetHeader(config.ProxyHeaderName) == "" {
				helper.RespondError(c, http.StatusForbidden, "代理 Header 缺失")
			} else {
				helper.RespondError(c, http.StatusForbidden, "用户不存在")
			}
			c.Abort()
			return
		}
		c.Set("username", username)
		c.Next()
	}
}

// RoutePermMiddleware 基于 METHOD+PATH 的集中式权限验证中间件
// 根据 Gin 已匹配的完整路由模板一次定位当前请求所需的权限
func RoutePermMiddleware(routePerms map[string]Route) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.FullPath()

		// Gin 未匹配到路由时 FullPath 为空，直接拒绝（防御性检查，正常不会触发）
		if path == "" {
			helper.RespondError(c, http.StatusForbidden, "未授权的访问路径")
			c.Abort()
			return
		}

		route, exists := routePerms[method+" "+path]
		if !exists {
			route, exists = routePerms["ANY "+path]
		}
		if !exists {
			helper.RespondError(c, http.StatusForbidden, "未授权的访问路径")
			c.Abort()
			return
		}

		// 匹配成功，进行权限验证
		if route.Module == "" {
			c.Next()
			return
		}

		// 验证模块权限
		username := c.GetString("username")
		member, exists := config.Members[username]
		if !exists {
			helper.RespondError(c, http.StatusForbidden, "用户不存在")
			c.Abort()
			return
		}

		perm := member.Permissions[route.Module]
		label := route.Label
		if label == "" {
			label = route.Module
		}

		// 直接比较：route.Perm 为空或 "r" 时只需有读权限；"rw" 时必须有 rw
		if route.Perm == "rw" {
			if perm != "rw" {
				helper.RespondError(c, http.StatusForbidden, "无 "+label+" 模块写权限")
				c.Abort()
				return
			}
		} else {
			if perm != "r" && perm != "rw" {
				helper.RespondError(c, http.StatusForbidden, "无 "+label+" 模块访问权限")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
