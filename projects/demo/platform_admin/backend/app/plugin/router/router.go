package router

import (
	"github.com/gin-gonic/gin"
	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk"
	"github.com/go-admin-team/go-admin-core/sdk/config"

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

	// 插件管理 API 已迁移至 common/auth/handler/PluginHandler（CR4）
	// 此处不再注册 /api/v1/admin/plugins 路由组，避免重复注册 panic
	// 仅保留插件代理路由

	// 插件代理路由（需要认证 + 租户中间件）
	dbDsn := ""
	if config.DatabaseConfig != nil {
		dbDsn = config.DatabaseConfig.Source
	}
	proxy := pluginPkg.NewPluginProxy(service.Manager, dbDsn)
	proxyGroup := r.Group("/api/v1/admin")
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
