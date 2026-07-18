package middleware

import (
	"fmt"
	"net/http"

	"go-admin/common/auth/cache"
	"go-admin/common/auth/config"
	"go-admin/common/auth/engine"

	"github.com/gin-gonic/gin"
)

// SuperAdminChecker 判断角色 ID 是否为超级管理员的回调
type SuperAdminChecker func(roleIDs []int64) bool

// PermissionMiddleware 授权中间件
// 从缓存获取用户权限集，使用 PermissionEngine 匹配权限码
type PermissionMiddleware struct {
	permCache         *cache.PermissionCache
	cfg               *config.Config
	superAdminChecker SuperAdminChecker
}

// NewPermissionMiddleware 构造授权中间件
// superAdminChecker 可为 nil，为 nil 时不做 SUPER_ADMIN 快捷放行（仅依赖权限集中的 ** 通配）
func NewPermissionMiddleware(permCache *cache.PermissionCache, cfg *config.Config, checker SuperAdminChecker) *PermissionMiddleware {
	return &PermissionMiddleware{
		permCache:         permCache,
		cfg:               cfg,
		superAdminChecker: checker,
	}
}

// RequirePermission 声明式权限检查，返回 gin.HandlerFunc
// 用法: router.GET("/xxx", pm.RequirePermission("user:list"), handler)
func (pm *PermissionMiddleware) RequirePermission(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取认证上下文
		authCtx := GetAuthContext(c)
		if authCtx == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"data":    nil,
				"message": "未认证",
			})
			return
		}

		// SUPER_ADMIN 角色直接放行
		if pm.isSuperAdmin(authCtx.Roles) {
			c.Next()
			return
		}

		// 从缓存获取用户权限码列表
		permCodes := pm.getUserPermissionCodes(authCtx)

		// 构建 PermissionEngine 进行匹配
		pe := engine.NewPermissionEngine(permCodes)
		if pe.HasPermission(code) {
			c.Next()
			return
		}

		// 无权限，返回 403
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":    40301,
			"data":    nil,
			"message": "权限不足",
		})
	}
}

// isSuperAdmin 检查角色列表是否包含超级管理员
func (pm *PermissionMiddleware) isSuperAdmin(roleIDs []int64) bool {
	if pm.superAdminChecker != nil {
		return pm.superAdminChecker(roleIDs)
	}
	return false
}

// getUserPermissionCodes 从缓存获取用户权限码列表
func (pm *PermissionMiddleware) getUserPermissionCodes(authCtx *AuthContext) []string {
	userID := fmt.Sprintf("%d", authCtx.UserID)
	tenantID := fmt.Sprintf("%d", authCtx.TenantID)
	appCode := pm.cfg.App.DefaultAppCode
	if appCode == "" {
		appCode = "default"
	}

	roleIDs := make([]string, 0, len(authCtx.Roles))
	for _, rid := range authCtx.Roles {
		roleIDs = append(roleIDs, fmt.Sprintf("%d", rid))
	}

	val, ok := pm.permCache.GetUserPermissions(userID, tenantID, appCode, roleIDs)
	if !ok {
		return nil
	}

	return toStringSlice(val)
}

// toStringSlice 将缓存值转为字符串切片
func toStringSlice(val interface{}) []string {
	switch v := val.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	default:
		return nil
	}
}
