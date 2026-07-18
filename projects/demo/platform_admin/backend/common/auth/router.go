package auth

import (
	"go-admin/common/auth/config"
	"go-admin/common/auth/handler"
	"go-admin/common/auth/middleware"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Dependencies 封装 auth 模块所有依赖，便于测试注入
type Dependencies struct {
	DB  *gorm.DB
	Cfg *config.Config

	AuthService    *service.AuthService
	TenantService  *service.TenantService
	UserService    *service.UserService
	RoleService    *service.RoleService
	UserRoleService *service.UserRoleService

	AuthHandler   *handler.AuthHandler
	TenantHandler *handler.TenantHandler
	UserHandler   *handler.UserHandler
	RoleHandler   *handler.RoleHandler

	ApiPermissionService *service.ApiPermissionService
	ApiPermissionHandler *handler.ApiPermissionHandler

	ResourceService *service.ResourceService
	ResourceHandler *handler.ResourceHandler

	ApplicationService *service.ApplicationService
	ApplicationHandler *handler.ApplicationHandler

	DataScopeService *service.DataScopeService
	DataScopeHandler *handler.DataScopeHandler

	OperationLogQueryService *service.OperationLogQueryService
	OperationLogHandler      *handler.OperationLogHandler

	ConfigService  *service.ConfigService
	ConfigHandler  *handler.ConfigHandler

	LoginLogService *service.LoginLogService
	LoginLogHandler *handler.LoginLogHandler
}

// RegisterRoutes 注册 auth 模块所有路由到指定路由组
// 可独立挂载到任意 gin.Engine
func RegisterRoutes(rg *gin.RouterGroup, deps *Dependencies) {
	// 认证路由（无需认证中间件）
	deps.AuthHandler.RegisterRoutes(rg)

	// 需要认证的 API 路由
	api := rg.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(deps.AuthService))

	// 动态权限检查（SUPER_ADMIN 放行 + 白名单放行 + permission_code 匹配）
	api.Use(middleware.DynamicPermissionMiddleware(deps.DB, deps.Cfg))

	// 注册各业务 handler 路由
	deps.TenantHandler.RegisterRoutes(api)
	deps.UserHandler.RegisterRoutes(api)
	deps.RoleHandler.RegisterRoutes(api)

	if deps.ApiPermissionHandler != nil {
		deps.ApiPermissionHandler.RegisterRoutes(api)
	}
	if deps.ResourceHandler != nil {
		deps.ResourceHandler.RegisterRoutes(api)
	}
	if deps.ApplicationHandler != nil {
		deps.ApplicationHandler.RegisterRoutes(api)
	}
	if deps.DataScopeHandler != nil {
		deps.DataScopeHandler.RegisterRoutes(api)
	}
	if deps.OperationLogHandler != nil {
		deps.OperationLogHandler.RegisterRoutes(api)
	}
	if deps.ConfigHandler != nil {
		deps.ConfigHandler.RegisterRoutes(api)
	}
	if deps.LoginLogHandler != nil {
		deps.LoginLogHandler.RegisterRoutes(api)
	}

	// Stub 路由
	registerDashboardStub(api)
	registerSysApiStub(api)
	registerMonitorStub(api)
}

// registerDashboardStub 注册 dashboard 占位路由，避免前端 404
func registerDashboardStub(api *gin.RouterGroup) {
	dash := api.Group("/dashboard")
	{
		dash.GET("/kpi", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"code":    0,
				"data":    gin.H{"status": "EMPTY", "message": "Dashboard 模块尚未实现", "data": nil},
				"message": "ok",
			})
		})
		dash.GET("/trend/workorder", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"code":    0,
				"data":    gin.H{"status": "EMPTY", "message": "工单趋势模块尚未实现", "data": nil},
				"message": "ok",
			})
		})
		dash.GET("/trend/billing", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"code":    0,
				"data":    gin.H{"status": "EMPTY", "message": "收费趋势模块尚未实现", "data": nil},
				"message": "ok",
			})
		})
	}
}

// registerSysApiStub 注册接口管理占位路由
func registerSysApiStub(api *gin.RouterGroup) {
	api.GET("/sys-apis", func(c *gin.Context) {
		handler.Success(c, gin.H{"list": []interface{}{}, "total": 0})
	})
}

// registerMonitorStub 注册监控占位路由
func registerMonitorStub(api *gin.RouterGroup) {
	api.GET("/monitor/server", func(c *gin.Context) {
		handler.Success(c, gin.H{"status": "EMPTY", "message": "服务监控模块尚未实现"})
	})
}
