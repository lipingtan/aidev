package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"

	"go-admin/app/game/apis"
	"go-admin/common/middleware"
)

// routerCheckRole 路由注册函数列表
var routerCheckRole = make([]func(*gin.RouterGroup, *jwt.GinJWTMiddleware), 0)

func init() {
	routerCheckRole = append(routerCheckRole,
		registerGameRouter,
		registerDlcRouter,
		registerPlayerRouter,
		registerOrderRouter,
		registerPaymentConfigRouter,
		registerH5PageRouter,
	)
}

// registerGameRouter 游戏管理路由
func registerGameRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.Game{}
	r := v1.Group("/game").
		Use(authMiddleware.MiddlewareFunc()).
		Use(middleware.WithTenantId()).
		Use(middleware.AuthCheckRole())
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
		r.POST("", api.Insert)
		r.PUT("/:id", api.Update)
		r.DELETE("/:id", api.Delete)
		r.POST("/:id/secret", api.RegenerateSecret)
	}
}
