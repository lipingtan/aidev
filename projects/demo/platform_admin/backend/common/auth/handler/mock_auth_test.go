package handler

import (
	"strconv"

	"go-admin/common/auth/middleware"

	"github.com/gin-gonic/gin"
)

// mockAuthMiddleware 测试用模拟认证中间件
// 从自定义 header X-Test-Tenant-ID / X-Test-User-ID 获取租户和用户 ID
// 默认 tenantID=1, userID=100
func mockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantIDStr := c.GetHeader("X-Test-Tenant-ID")
		userIDStr := c.GetHeader("X-Test-User-ID")

		tenantID := int64(1)
		userID := int64(100)

		if tenantIDStr != "" {
			if id, err := strconv.ParseInt(tenantIDStr, 10, 64); err == nil {
				tenantID = id
			}
		}
		if userIDStr != "" {
			if id, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
				userID = id
			}
		}

		middleware.SetAuthContext(c, &middleware.AuthContext{
			UserID:   userID,
			TenantID: tenantID,
			Roles:    []int64{1},
		})
		c.Next()
	}
}
