package middleware

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

const TenantIdKey = "tenantId"

// WithTenantId 从 JWT claims 提取 tenant_id 注入 gin context
// 必须在 JWT 验证中间件之后使用
func WithTenantId() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		if tenantId, ok := claims[TenantIdKey]; ok {
			switch v := tenantId.(type) {
			case float64:
				c.Set(TenantIdKey, int(v))
			case int:
				c.Set(TenantIdKey, v)
			default:
				c.Set(TenantIdKey, 0)
			}
		} else {
			c.Set(TenantIdKey, 0)
		}
		c.Next()
	}
}

// GetTenantId 从 gin context 获取当前租户 ID
func GetTenantId(c *gin.Context) int {
	if v, exists := c.Get(TenantIdKey); exists {
		if id, ok := v.(int); ok {
			return id
		}
	}
	return 0
}
