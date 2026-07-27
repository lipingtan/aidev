package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"

	"go-admin/app/admin/apis"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerApprovalFlowRouter)
}

// registerApprovalFlowRouter 注册审批流定义路由
// 最终路径：/api/v1/admin/approval-flows
func registerApprovalFlowRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.ApprovalFlow{}
	r := v1.Group("/admin/approval-flows").
		Use(authMiddleware.MiddlewareFunc()).
		Use(middleware.AuthCheckRole())
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
		r.POST("", api.Insert)
		r.PUT("/:id", api.Update)
		r.DELETE("/:id", api.Delete)
	}
}
