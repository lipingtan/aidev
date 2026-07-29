package middleware

import "github.com/gin-gonic/gin"

// IsSuperAdmin 判断当前请求用户是否为 SUPER_ADMIN
// 通过 gin.Context 中缓存的 "is_super_admin" 标志判断（由 DynamicPermissionMiddleware 设置）
// 如无缓存标志，回退检查 AuthContext.Roles 是否已在 isSuperAdminByDB 判定中通过
func IsSuperAdmin(c *gin.Context) bool {
	if v, exists := c.Get("is_super_admin"); exists {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// IsTenantAdmin 判断当前请求用户是否为租户管理员
// 通过 gin.Context 中缓存的 "is_tenant_admin" 标志判断
// 需在中间件中显式设置（如检测到 role_type 含 TENANT_ADMIN 或非 SUPER_ADMIN 的管理端用户）
func IsTenantAdmin(c *gin.Context) bool {
	if v, exists := c.Get("is_tenant_admin"); exists {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	// 管理端用户（非 SUPER_ADMIN）默认视为 TENANT_ADMIN 级别
	// 因为管理端路由受 DynamicPermissionMiddleware 保护，能到达 handler 的必有有效角色
	if !IsSuperAdmin(c) {
		authCtx := GetAuthContext(c)
		if authCtx != nil && authCtx.UserPool != "user" && len(authCtx.Roles) > 0 {
			return true
		}
	}
	return false
}

// SetSuperAdminFlag 设置 SUPER_ADMIN 标志到 gin.Context
func SetSuperAdminFlag(c *gin.Context, flag bool) {
	c.Set("is_super_admin", flag)
}
