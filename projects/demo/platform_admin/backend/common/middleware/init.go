package middleware

import (
	"github.com/gin-gonic/gin"
)

const (
	JwtTokenCheck   string = "JwtToken"
	RoleCheck       string = "AuthCheckRole"
	PermissionCheck string = "PermissionAction"
)

func InitMiddleware(r *gin.Engine) {
	r.Use(DemoEvn())
	// 数据库链接
	r.Use(WithContextDb)
	// 日志处理
	r.Use(LoggerToFile())
	// 自定义错误处理
	r.Use(CustomError)
	// NoCache is a middleware function that appends headers
	r.Use(NoCache)
	// 跨域处理
	r.Use(Options)
	// Secure is a middleware function that appends security
	r.Use(Secure)
	// 旧 JWT/Casbin 中间件已废弃，由 auth-rbac 模块 AuthMiddleware 替代
	// sdk.Runtime.SetMiddleware(JwtTokenCheck, (*jwt.GinJWTMiddleware).MiddlewareFunc)
	// sdk.Runtime.SetMiddleware(RoleCheck, AuthCheckRole())
	// sdk.Runtime.SetMiddleware(PermissionCheck, actions.PermissionAction())
}
