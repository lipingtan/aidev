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
// 路由组结构: /auth（公开）+ /api/v1/common（仅认证）+ /api/v1/admin（完整中间件链）
func RegisterRoutes(rg *gin.RouterGroup, deps *Dependencies) {
	// 认证路由（公开，不走任何中间件）
	deps.AuthHandler.RegisterRoutes(rg)

	// 公共路由（仅需认证，不走权限中间件）
	common := rg.Group("/api/v1/common")
	common.Use(middleware.AuthMiddleware(deps.AuthService))
	{
		common.GET("/user-menu", deps.ResourceHandler.GetUserMenu)
	}

	// 管理端业务路由（认证 + 动态权限检查）
	admin := rg.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware(deps.AuthService))
	admin.Use(middleware.DynamicPermissionMiddleware(deps.DB, deps.Cfg))
	{
		deps.TenantHandler.RegisterRoutes(admin)
		deps.UserHandler.RegisterRoutes(admin)
		deps.RoleHandler.RegisterRoutes(admin)

		if deps.ApiPermissionHandler != nil {
			deps.ApiPermissionHandler.RegisterRoutes(admin)
		}
		if deps.ResourceHandler != nil {
			deps.ResourceHandler.RegisterRoutes(admin)
		}
		if deps.ApplicationHandler != nil {
			deps.ApplicationHandler.RegisterRoutes(admin)
		}
		if deps.DataScopeHandler != nil {
			deps.DataScopeHandler.RegisterRoutes(admin)
		}
		if deps.OperationLogHandler != nil {
			deps.OperationLogHandler.RegisterRoutes(admin)
		}
		if deps.ConfigHandler != nil {
			deps.ConfigHandler.RegisterRoutes(admin)
		}
		if deps.LoginLogHandler != nil {
			deps.LoginLogHandler.RegisterRoutes(admin)
		}

		// Stub 路由
		registerDashboardStub(admin)
		registerSysApiStub(admin)
		registerMonitorStub(admin)
	}
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
