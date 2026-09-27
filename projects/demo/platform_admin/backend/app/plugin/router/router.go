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

	authSvc := getAuthService()
	if authSvc == nil {
		log.Warn("[plugin-router] auth service 不可用，插件路由未注册认证中间件")
	}

	dbDsn := ""
	if config.DatabaseConfig != nil {
		dbDsn = config.DatabaseConfig.Source
	}
	lazyProxy := &lazyPluginProxy{dbDsn: dbDsn}

	// ── 1. admin 端代理：/api/v1/admin/plugin/:name/*action
	//    auth + 权限中间件，需要 admin token
	adminGroup := r.Group("/api/v1/admin")
	if authSvc != nil {
		adminGroup.Use(authMiddleware.AuthMiddleware(authSvc))
	}
	adminGroup.Use(common.WithTenantId())
	lazyProxy.RegisterAdminRoutes(adminGroup)

	// ── 2. user 公开端代理：/api/v1/user/public/plugin/:name/*action
	//    无需任何 token，匿名可访问
	publicGroup := r.Group("/api/v1/user/public")
	lazyProxy.RegisterPublicRoutes(publicGroup)

	// ── 3. user 认证端代理：/api/v1/user/auth/plugin/:name/*action
	//    需要 C端用户 token（由 auth.Init 注入 bizUserSvc 后才有效）
	bizChecker := getBizUserChecker()
	userGroup := r.Group("/api/v1/user/auth")
	if authSvc != nil && bizChecker != nil {
		userGroup.Use(authMiddleware.AuthMiddleware(authSvc, bizChecker))
	} else if authSvc != nil {
		userGroup.Use(authMiddleware.AuthMiddleware(authSvc))
	}
	lazyProxy.RegisterUserRoutes(userGroup)
}

// lazyPluginProxy 延迟绑定的插件代理，每次请求时从 globalPluginManager 取最新实例
type lazyPluginProxy struct {
	dbDsn string
}

func (lp *lazyPluginProxy) handler() func(c *gin.Context) {
	return func(c *gin.Context) {
		mgr := globalPluginManager
		if mgr == nil {
			mgr = service.Manager
		}
		proxy := pluginPkg.NewPluginProxy(mgr, lp.dbDsn)
		proxy.Handler()(c)
	}
}

// RegisterAdminRoutes 注册 admin 端代理路由
func (lp *lazyPluginProxy) RegisterAdminRoutes(r *gin.RouterGroup) {
	r.Any("/plugin/:name/*action", lp.handler())
}

// RegisterPublicRoutes 注册 user 端公开代理路由（无需登录）
// 前端请求路径：/api/v1/user/public/plugin/{name}/{action}
func (lp *lazyPluginProxy) RegisterPublicRoutes(r *gin.RouterGroup) {
	r.Any("/plugin/:name/*action", lp.handler())
}

// RegisterUserRoutes 注册 user 端认证代理路由（需要 C端用户 token）
// 前端请求路径：/api/v1/user/auth/plugin/{name}/{action}
func (lp *lazyPluginProxy) RegisterUserRoutes(r *gin.RouterGroup) {
	r.Any("/plugin/:name/*action", lp.handler())
}
var globalAuthService *authService.AuthService

// globalPluginManager 从 auth.Init 注入的 PluginManager（与 PluginHandler 共享同一实例）
var globalPluginManager *pluginPkg.PluginManager

// globalBizUserChecker C端用户检查器（由 auth.Init 注入，供 user 端认证代理使用）
var globalBizUserChecker authMiddleware.BizUserChecker

// SetAuthService 由 auth.Init 调用，注入 AuthService 供插件路由使用
func SetAuthService(svc *authService.AuthService) {
	globalAuthService = svc
}

// SetPluginManager 由 auth.Init 调用，注入共享 PluginManager 供代理路由使用
func SetPluginManager(mgr *pluginPkg.PluginManager) {
	globalPluginManager = mgr
}

// SetBizUserChecker 由 auth.Init 调用，注入 C端用户检查器供 user 端认证代理使用
func SetBizUserChecker(checker authMiddleware.BizUserChecker) {
	globalBizUserChecker = checker
}

func getAuthService() *authService.AuthService {
	return globalAuthService
}

func getBizUserChecker() authMiddleware.BizUserChecker {
	return globalBizUserChecker
}
