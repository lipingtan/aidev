package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"

	"go-admin/app/game/apis"
	"go-admin/common/middleware"
)

func withTenant(authMiddleware *jwt.GinJWTMiddleware) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		authMiddleware.MiddlewareFunc(),
		middleware.WithTenantId(),
		middleware.AuthCheckRole(),
	}
}

func registerDlcRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.Dlc{}
	r := v1.Group("/game-dlc").Use(withTenant(authMiddleware)...)
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
		r.POST("", api.Insert)
		r.PUT("/:id", api.Update)
		r.DELETE("/:id", api.Delete)
		r.POST("/:id/upload", api.UploadPck)
	}
}

func registerPlayerRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.Player{}
	r := v1.Group("/game-player").Use(withTenant(authMiddleware)...)
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
		r.PUT("/:id/ban", api.Ban)
	}
}

func registerOrderRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.Order{}
	r := v1.Group("/game-order").Use(withTenant(authMiddleware)...)
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
		r.POST("/:id/refund", api.Refund)
	}
}

func registerPaymentConfigRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.PaymentConfig{}
	r := v1.Group("/game-payment-config").Use(withTenant(authMiddleware)...)
	{
		r.GET("", api.GetPage)
		r.POST("", api.Save)
	}
}

func registerH5PageRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.H5Page{}
	r := v1.Group("/game-h5").Use(withTenant(authMiddleware)...)
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
		r.POST("", api.Insert)
		r.PUT("/:id", api.Update)
		r.DELETE("/:id", api.Delete)
	}
}
