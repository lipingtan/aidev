package auth

import (
	authCache "go-admin/common/auth/cache"
	"go-admin/common/auth/config"
	"go-admin/common/auth/handler"
	"go-admin/common/auth/middleware"
	"go-admin/common/auth/repository"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Dependencies 封装 auth 模块所有依赖，便于测试注入
type Dependencies struct {
	DB  *gorm.DB
	Cfg *config.Config

	AuthService     *service.AuthService
	TenantService   *service.TenantService
	UserService     *service.UserService
	RoleService     *service.RoleService
	UserRoleService *service.UserRoleService
	FieldRegistry   *service.FieldRegistry

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

	LoginLogService *service.LoginLogService
	LoginLogHandler *handler.LoginLogHandler

	FieldPermissionService *service.FieldPermissionService
	FieldPermissionHandler *handler.FieldPermissionHandler

	RecordShareService *service.RecordShareService
	RecordShareHandler *handler.RecordShareHandler

	OrgUnitService      *service.OrgUnitService
	OrgUnitHandler      *handler.OrgUnitHandler

	AdminConfigService *service.AdminConfigService
	AdminConfigHandler *handler.AdminConfigHandler

	// 应用目录与订阅管理
	AppCatalogHandler *handler.AppCatalogHandler

	// 域名-租户映射
	TenantDomainService *service.TenantDomainService
	TenantDomainHandler *handler.TenantDomainHandler

	// 插件管理（可选，Task 11 负责注入）
	PluginHandler *handler.PluginHandler

	// 中间件缓存
	AppPrefixMap         *middleware.AppPrefixMap
	ModuleCodeCache      *middleware.ModuleCodeCache
	FieldObjectRegistry  *middleware.FieldObjectRegistry
	FieldPermissionRepo  repository.FieldPermissionRepository

	// CR-8: 权限码 Redis 缓存
	PermCodeCache *authCache.PermCodeCache
}

// ExtraAdminRoutesFn 允许外部模块（如 app/admin/apis）注册额外的 /api/v1/admin/ 路由
// 在 auth.Init() 调用前通过 RegisterExtraAdminRoutes() 注册，在 RegisterRoutes 中执行
var extraAdminRoutesFns []func(admin *gin.RouterGroup)

// RegisterExtraAdminRoutes 注册额外的 admin 路由函数（必须在 auth.Init 之前调用）
func RegisterExtraAdminRoutes(fn func(admin *gin.RouterGroup)) {
	extraAdminRoutesFns = append(extraAdminRoutesFns, fn)
}
// 路由组结构: /auth（公开）+ /api/v1/public（公开业务）+ /api/v1/common（仅认证）+ /api/v1/admin（完整中间件链）
func RegisterRoutes(rg *gin.RouterGroup, deps *Dependencies) {
	// 认证路由（公开，不走任何中间件）
	deps.AuthHandler.RegisterRoutes(rg)

	// 公开业务路由（无需认证）
	public := rg.Group("/api/v1/public")
	{
		if deps.TenantDomainHandler != nil {
			public.GET("/tenant-domain", deps.TenantDomainHandler.QueryByDomain)
		}
	}

	// 公共路由（仅需认证，不走权限中间件）
	common := rg.Group("/api/v1/common")
	common.Use(middleware.AuthMiddleware(deps.AuthService))
	{
		common.GET("/user-menu", deps.ResourceHandler.GetUserMenu)
	}

	// 管理端业务路由（认证 + 应用解析 + 动态权限检查）
	admin := rg.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware(deps.AuthService))
	if deps.AppPrefixMap != nil && deps.ModuleCodeCache != nil {
		admin.Use(middleware.AppResolveMiddleware(deps.AppPrefixMap, deps.ModuleCodeCache, deps.DB))
	}
	admin.Use(middleware.DynamicPermissionMiddleware(deps.DB, deps.Cfg, deps.AdminConfigService))
	if deps.FieldObjectRegistry != nil && deps.FieldPermissionRepo != nil {
		admin.Use(middleware.FieldFilterMiddleware(deps.FieldObjectRegistry, deps.FieldPermissionRepo, deps.DB))
	}
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
		if deps.LoginLogHandler != nil {
			deps.LoginLogHandler.RegisterRoutes(admin)
		}
		if deps.FieldPermissionHandler != nil {
			deps.FieldPermissionHandler.RegisterRoutes(admin)
		}
		if deps.RecordShareHandler != nil {
			deps.RecordShareHandler.RegisterRoutes(admin)
		}
		if deps.OrgUnitHandler != nil {
			orgUnits := admin.Group("/org-units")
			{
				orgUnits.GET("/tree", deps.OrgUnitHandler.GetTree)
				orgUnits.POST("", deps.OrgUnitHandler.Create)
				orgUnits.PUT("/:id", deps.OrgUnitHandler.Update)
				orgUnits.DELETE("/:id", deps.OrgUnitHandler.Delete)
				orgUnits.GET("/:id/users", deps.OrgUnitHandler.GetNodeUsers)
				orgUnits.PUT("/:id/users", deps.OrgUnitHandler.SetNodeUsers)
			}
		}
		if deps.AdminConfigHandler != nil {
			cfgs := admin.Group("/configs")
			{
				cfgs.GET("", deps.AdminConfigHandler.List)
				cfgs.POST("", deps.AdminConfigHandler.Create)
				cfgs.PUT("/:id", deps.AdminConfigHandler.Update)
				cfgs.DELETE("/:id", deps.AdminConfigHandler.Delete)
				cfgs.GET("/resolve/:key", deps.AdminConfigHandler.Resolve)
				cfgs.GET("/feature-flags", deps.AdminConfigHandler.GetFeatureFlags)
			}
		}

		// 插件管理路由（可选，依赖 Task 11 注入 PluginHandler）
		if deps.PluginHandler != nil {
			deps.PluginHandler.RegisterRoutes(admin)
		}

		// 应用目录与订阅管理路由
		if deps.AppCatalogHandler != nil {
			deps.AppCatalogHandler.RegisterRoutes(admin)
		}

		// 域名-租户映射管理路由
		if deps.TenantDomainHandler != nil {
			deps.TenantDomainHandler.RegisterRoutes(admin)
		}

		// 外部模块扩展路由（如审批流）
		for _, fn := range extraAdminRoutesFns {
			fn(admin)
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
