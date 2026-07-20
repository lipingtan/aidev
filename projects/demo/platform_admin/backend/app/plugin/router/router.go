package router

import (
	"github.com/gin-gonic/gin"
	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk"
	"github.com/go-admin-team/go-admin-core/sdk/config"

	"go-admin/app/plugin/apis"
	pluginModels "go-admin/app/plugin/models"
	"go-admin/app/plugin/service"
	authMiddleware "go-admin/common/auth/middleware"
	authService "go-admin/common/auth/service"
	common "go-admin/common/middleware"
	pluginPkg "go-admin/common/plugin"
)

// InitPluginRouter 插件管理路由初始化（新 auth-rbac 体系）
func InitPluginRouter() {
	var r *gin.Engine
	h := sdk.Runtime.GetEngine()
	if h == nil {
		log.Fatal("not found engine...")
		return
	}
	switch h.(type) {
	case *gin.Engine:
		r = h.(*gin.Engine)
	default:
		log.Fatal("not support other engine")
		return
	}

	// 初始化插件服务（PluginManager + Installer）
	for _, db := range sdk.Runtime.GetDb() {
		if db != nil {
			_ = db.AutoMigrate(&pluginModels.SysPlugin{})
			service.Init(db)
			break
		}
	}

	// 获取新 auth-rbac 的 AuthService（从全局获取）
	authSvc := getAuthService()
	if authSvc == nil {
		log.Warn("[plugin-router] auth service 不可用，插件路由未注册认证中间件")
	}

	api := apis.Plugin{}

	// 插件管理 API（使用新 auth-rbac 认证中间件）
	v1 := r.Group("/api/v1/admin")
	pluginGroup := v1.Group("/plugins")
	if authSvc != nil {
		pluginGroup.Use(authMiddleware.AuthMiddleware(authSvc))
	}
	{
		pluginGroup.GET("", api.List)
		pluginGroup.POST("/install", api.Install)
		pluginGroup.POST("/:name/start", api.Start)
		pluginGroup.POST("/:name/stop", api.Stop)
		pluginGroup.DELETE("/:name", api.Uninstall)
		pluginGroup.GET("/:name/health", api.Health)
	}

	// 插件代理路由（需要认证 + 租户中间件）
	dbDsn := ""
	if config.DatabaseConfig != nil {
		dbDsn = config.DatabaseConfig.Source
	}
	proxy := pluginPkg.NewPluginProxy(service.Manager, dbDsn)
	proxyGroup := v1.Group("")
	if authSvc != nil {
		proxyGroup.Use(authMiddleware.AuthMiddleware(authSvc))
	}
	proxyGroup.Use(common.WithTenantId())
	proxy.RegisterRoutes(proxyGroup)
}

// getAuthService 从全局获取 AuthService 实例
// auth.Init 会在 InitPluginRouter 之前调用，所以实例已可用
var globalAuthService *authService.AuthService

// SetAuthService 由 auth.Init 调用，注入 AuthService 供插件路由使用
func SetAuthService(svc *authService.AuthService) {
	globalAuthService = svc
}

func getAuthService() *authService.AuthService {
	return globalAuthService
}
