package user_auth

import (
	"go-admin/app/user_auth/handler"
	userStrategy "go-admin/app/user_auth/strategy"
	"go-admin/common/auth/strategy"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册 user_auth 域的所有路由
// - C端公开路由：send-code, login（不需要 auth 中间件）
// - C端私有路由：logout, menu（需 auth 中间件）
// - 管理端路由：biz-users CRUD（需 auth + permission 中间件）
func RegisterRoutes(
	publicGroup *gin.RouterGroup,
	userGroup *gin.RouterGroup,
	adminGroup *gin.RouterGroup,
	authHandler *handler.UserAuthHandler,
	bizUserHandler *handler.BizUserHandler,
) {
	// C端公开接口（无需认证）
	publicGroup.POST("/send-code", authHandler.SendCode)
	publicGroup.POST("/login", authHandler.Login)

	// C端认证接口（需 auth 中间件）
	userGroup.POST("/logout", authHandler.Logout)
	userGroup.GET("/menu", authHandler.GetMenu)

	// 管理端路由（需 auth + permission 中间件）
	bizUserHandler.RegisterRoutes(adminGroup)
}

// RegisterSmsStrategy 将 SmsStrategy 注册到 StrategyRouter
func RegisterSmsStrategy(router *strategy.StrategyRouter, smsStrategy *userStrategy.SmsStrategy) {
	router.Register(smsStrategy)
}
