package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"

	"go-admin/app/admin/apis"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerApprovalRouter)
}

// registerApprovalRouter 注册审批实例路由
// 最终路径：/api/v1/admin/approvals
func registerApprovalRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.Approval{}
	r := v1.Group("/admin/approvals").
		Use(authMiddleware.MiddlewareFunc()).
		Use(middleware.AuthCheckRole())
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
		r.POST("", api.Insert)
		r.POST("/:id/approve", api.Approve)
		r.POST("/:id/reject", api.Reject)
		r.POST("/:id/cancel", api.Cancel)
	}
}
