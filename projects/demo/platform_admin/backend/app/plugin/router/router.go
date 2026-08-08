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

	// 插件代理路由（需要认证 + 租户中间件）：延迟获取 Manager，避免时序问题
	dbDsn := ""
	if config.DatabaseConfig != nil {
		dbDsn = config.DatabaseConfig.Source
	}
	lazyProxy := &lazyPluginProxy{dbDsn: dbDsn}
	proxyGroup := r.Group("/api/v1/admin")
	if authSvc != nil {
		proxyGroup.Use(authMiddleware.AuthMiddleware(authSvc))
	}
	proxyGroup.Use(common.WithTenantId())
	lazyProxy.RegisterRoutes(proxyGroup)
}

// lazyPluginProxy 延迟绑定的插件代理，每次请求时从 globalPluginManager 取最新实例
type lazyPluginProxy struct {
	dbDsn string
}

func (lp *lazyPluginProxy) RegisterRoutes(r *gin.RouterGroup) {
	r.Any("/plugin/:name/*action", func(c *gin.Context) {
		// 延迟获取 Manager：auth.Init 已在 AppRouters 之后执行完毕，globalPluginManager 已设置
		mgr := globalPluginManager
		if mgr == nil {
			mgr = service.Manager
		}
		proxy := pluginPkg.NewPluginProxy(mgr, lp.dbDsn)
		proxy.Handler()(c)
	})
}
var globalAuthService *authService.AuthService

// globalPluginManager 从 auth.Init 注入的 PluginManager（与 PluginHandler 共享同一实例）
var globalPluginManager *pluginPkg.PluginManager

// SetAuthService 由 auth.Init 调用，注入 AuthService 供插件路由使用
func SetAuthService(svc *authService.AuthService) {
	globalAuthService = svc
}

// SetPluginManager 由 auth.Init 调用，注入共享 PluginManager 供代理路由使用
func SetPluginManager(mgr *pluginPkg.PluginManager) {
	globalPluginManager = mgr
}

func getAuthService() *authService.AuthService {
	return globalAuthService
}
