package router

import (
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk"

	"go-admin/app/plugin/apis"
	"go-admin/app/plugin/service"
	common "go-admin/common/middleware"
)

// InitPluginRouter 插件管理路由初始化
func InitPluginRouter() {
	var r *gin.Engine
	h := sdk.Runtime.GetEngine()
	if h == nil {
		log.Fatal("not found engine...")
		os.Exit(-1)
	}
	switch h.(type) {
	case *gin.Engine:
		r = h.(*gin.Engine)
	default:
		log.Fatal("not support other engine")
		os.Exit(-1)
	}

	authMiddleware, err := common.AuthInit()
	if err != nil {
		log.Fatalf("JWT Init Error, %s", err.Error())
	}

	// 初始化插件服务（PluginManager + Installer）
	// 从 Runtime 获取第一个可用的数据库连接
	for _, db := range sdk.Runtime.GetDb() {
		if db != nil {
			service.Init(db)
			break
		}
	}

	api := apis.Plugin{}

	// 插件管理 API（需要 JWT + 管理员权限）
	v1 := r.Group("/api/v1")
	pluginGroup := v1.Group("/plugins").
		Use(authMiddleware.MiddlewareFunc()).
		Use(common.AuthCheckRole())
	{
		pluginGroup.GET("", api.List)
		pluginGroup.POST("/install", api.Install)
		pluginGroup.POST("/:name/start", api.Start)
		pluginGroup.POST("/:name/stop", api.Stop)
		pluginGroup.DELETE("/:name", api.Uninstall)
		pluginGroup.GET("/:name/health", api.Health)
	}
}
