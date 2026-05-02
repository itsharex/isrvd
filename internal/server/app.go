package server

import (
	"github.com/gin-gonic/gin"
	"github.com/rehiy/pango/httpd"
	"github.com/rehiy/pango/logman"

	"isrvd/config"
	svcAccount "isrvd/internal/service/account"
	svcApisix "isrvd/internal/service/apisix"
	svcCompose "isrvd/internal/service/compose"
	svcDocker "isrvd/internal/service/docker"
	svcOverview "isrvd/internal/service/overview"
	svcSwarm "isrvd/internal/service/swarm"
	svcSystem "isrvd/internal/service/system"

	"isrvd/public"
)

const APINamespace = "/api"

// App 应用实例，持有各业务服务
type App struct {
	*gin.Engine
	overviewSvc *svcOverview.Service
	settingsSvc *svcSystem.SettingsService
	auditSvc    *svcSystem.AuditService
	accountSvc  *svcAccount.Service
	apisixSvc   *svcApisix.Service
	dockerSvc   *svcDocker.Service
	swarmSvc    *svcSwarm.Service
	composeSvc  *svcCompose.DeployService
	routePerms  map[string]svcAccount.RouteInfo // METHOD+完整路径 → 路由权限索引
}

func StartApp() {
	app := &App{Engine: httpd.Engine(config.Debug), routePerms: make(map[string]svcAccount.RouteInfo)}

	// 初始化各业务服务
	app.overviewSvc = svcOverview.NewService()
	app.settingsSvc = svcSystem.NewSettingsService()
	app.auditSvc = svcSystem.NewAuditService()
	app.accountSvc = svcAccount.NewService()

	if apisixSvc, err := svcApisix.NewService(); err != nil {
		logman.Warn("Apisix service unavailable", "error", err)
	} else {
		app.apisixSvc = apisixSvc
	}

	if dockerSvc, err := svcDocker.NewService(); err != nil {
		logman.Warn("Docker service unavailable", "error", err)
	} else {
		app.dockerSvc = dockerSvc
		app.swarmSvc = svcSwarm.NewService()
	}

	if composeSvc, err := svcCompose.NewService(); err != nil {
		logman.Warn("Compose service unavailable", "error", err)
	} else {
		app.composeSvc = composeSvc
	}

	// 统一注册路由
	app.initRoutes()
	httpd.StaticEmbed(public.Efs, "", "")

	// 启动 HTTP 服务
	httpd.Server(config.ListenAddr)
}

// Route 定义单个路由的完整信息（同时用于注册和权限验证）
type Route struct {
	Method  string          // HTTP 方法：GET/POST/PUT/PATCH/DELETE/ANY
	Path    string          // 路由路径（Gin 格式，支持 :param 和 *）
	Handler gin.HandlerFunc // 处理函数
	Module  string          // 模块名，空字符串表示无需模块权限（如 /auth/*）
	Label   string          // 模块显示名，用于错误提示
	Perm    string          // 所需权限：空或"r"=只读，"rw"=读写
}

// initRoutes 初始化路由表并注册所有路由
// 按模块注册路由，每个模块自己管理路由定义和注册
func (app *App) initRoutes() {
	r := app.Group(APINamespace)

	// 公开路由组：MixAuthMiddleware（认证失败时放行）
	publicGroup := r.Group("")
	publicGroup.Use(MixAuthMiddleware(app.accountSvc))

	// 注册 account 模块的公开路由（无需认证）
	publicGroup.GET("/auth/info", app.accountAuthInfo)
	publicGroup.POST("/auth/login", app.accountLogin)

	// 受保护路由组：AuthMiddleware + RoutePermMiddleware + AuditMiddleware
	protectedGroup := r.Group("")
	protectedGroup.Use(AuthMiddleware(app.accountSvc))
	protectedGroup.Use(RoutePermMiddleware(app.routePerms, app.accountSvc))
	protectedGroup.Use(AuditMiddleware(app.auditSvc))

	// 加载所有模块的受保护路由定义（含 account 模块）
	allRoutes := app.collectRoutes()

	// 按条件注册路由（处理服务可用性依赖）
	for _, route := range allRoutes {
		if !app.isRouteAvailable(route) {
			continue
		}
		app.registerRoute(protectedGroup, route)
	}
}

// collectRoutes 收集所有模块的路由定义
// 每个模块通过 defineXxxRoutes() 方法返回自己的路由列表
func (app *App) collectRoutes() []Route {
	var routes []Route

	// 概览（系统统计 + 服务探测，无需权限）
	routes = append(routes, app.defineOverviewRoutes()...)
	// 系统设置
	routes = append(routes, app.defineSystemRoutes()...)
	// Account 模块：受保护路由（公开路由 /auth/info、/auth/login 由 initRoutes 直接注册）
	routes = append(routes, app.defineAccountRoutes()...)
	// Web 终端
	routes = append(routes, app.defineShellRoutes()...)
	// 文件管理
	routes = append(routes, app.defineFilerRoutes()...)
	// LLM 代理
	routes = append(routes, app.defineAgentRoutes()...)
	// APISIX 管理
	routes = append(routes, app.defineApisixRoutes()...)
	// Docker 管理
	routes = append(routes, app.defineDockerRoutes()...)
	// Swarm 管理
	routes = append(routes, app.defineSwarmRoutes()...)
	// Compose 部署
	routes = append(routes, app.defineComposeRoutes()...)

	return routes
}

// registerRoute 注册单个路由，并同步建立 METHOD+完整路由模板的权限索引
func (app *App) registerRoute(group *gin.RouterGroup, route Route) {
	app.routePerms[route.Method+" "+APINamespace+route.Path] = svcAccount.RouteInfo{
		Module: route.Module,
		Label:  route.Label,
		Perm:   route.Perm,
	}

	switch route.Method {
	case "GET":
		group.GET(route.Path, route.Handler)
	case "POST":
		group.POST(route.Path, route.Handler)
	case "PUT":
		group.PUT(route.Path, route.Handler)
	case "PATCH":
		group.PATCH(route.Path, route.Handler)
	case "DELETE":
		group.DELETE(route.Path, route.Handler)
	case "ANY":
		group.Any(route.Path, route.Handler)
	}
}

// isRouteAvailable 检查路由是否满足服务可用性条件（基于 Module 字段判断）
func (app *App) isRouteAvailable(route Route) bool {
	switch route.Module {
	case "apisix":
		return app.apisixSvc != nil
	case "docker", "shell":
		return app.dockerSvc != nil
	case "swarm":
		return app.dockerSvc != nil && app.swarmSvc != nil
	case "compose":
		return app.dockerSvc != nil && app.composeSvc != nil
	default:
		return true
	}
}
