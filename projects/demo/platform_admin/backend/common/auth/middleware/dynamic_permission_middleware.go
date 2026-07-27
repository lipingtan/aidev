package middleware

import (
	"log"
	"net/http"
	"sync"

	"go-admin/common/auth/config"
	"go-admin/common/auth/engine"
	"go-admin/common/auth/service"
	"go-admin/common/auth/strategy"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// skipPaths 免检白名单（FullPath 匹配）
var skipPaths = map[string]bool{
	"/auth/login":                              true,
	"/auth/tenant/select":                      true,
	"/auth/refresh":                            true,
	"/auth/logout":                             true,
	"/setup":                                   true,
	"/setup/":                                  true,
	"/setup/status":                            true,
	"/setup/init":                              true,
	"/setup/test-db":                           true,
	"/api/v1/admin/dashboard/kpi":              true,
	"/api/v1/admin/dashboard/trend/workorder":  true,
	"/api/v1/admin/dashboard/trend/billing":    true,
	"/api/v1/admin/monitor/server":             true,
	"/api/v1/admin/sys-apis":                   true,
	"/api/v1/admin/app-catalog":                true,
	"/api/v1/admin/app-subscriptions":          true,
	// 审批流接口（通过 RegisterExtraAdminRoutes 注册，未在 admin_api_permission 自动发现）
	"/api/v1/admin/approval-flows":             true,
	"/api/v1/admin/approval-flows/:id":         true,
	"/api/v1/admin/approvals":                  true,
	"/api/v1/admin/approvals/:id":              true,
	"/api/v1/admin/approvals/:id/approve":      true,
	"/api/v1/admin/approvals/:id/reject":       true,
	"/api/v1/admin/approvals/:id/cancel":       true,
}

// DynamicPermissionMiddleware 动态权限检查中间件
// 按 FullPath + Method 从数据库查 permission_code，再校验用户是否拥有该权限
// 无 permission_code 的接口默认拒绝（除非在白名单中）
func DynamicPermissionMiddleware(db *gorm.DB, cfg *config.Config, configSvc ...*service.AdminConfigService) gin.HandlerFunc {
	// 启动时预加载 permission_code 映射（method:path → code）
	codeMap := loadPermissionCodeMap(db)
	var mu sync.RWMutex

	// 可选注入 AdminConfigService（功能开关用）
	var adminConfigSvc *service.AdminConfigService
	if len(configSvc) > 0 && configSvc[0] != nil {
		adminConfigSvc = configSvc[0]
	}

	return func(c *gin.Context) {
		fullPath := c.FullPath()

		// 白名单检查
		if skipPaths[fullPath] {
			c.Next()
			return
		}

		// 如果 codeMap 为空（可能是启动时 AutoDiscover 还未完成），尝试重新加载
		mu.RLock()
		empty := len(codeMap) == 0
		mu.RUnlock()
		if empty {
			mu.Lock()
			if len(codeMap) == 0 {
				codeMap = loadPermissionCodeMap(db)
			}
			mu.Unlock()
		}

		// 获取认证上下文
		authCtx := GetAuthContext(c)
		if authCtx == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 40101, "data": nil, "message": "未认证",
			})
			return
		}

		// C端用户（pool=user）跳过 RBAC 权限检查
		if authCtx.UserPool == strategy.UserPoolUser {
			c.Next()
			return
		}

		// SUPER_ADMIN 直接放行
		if isSuperAdminByDB(db, cfg, authCtx.Roles) {
			c.Next()
			return
		}

		// 功能开关检查：按 module_code 查 feature.{module_code}.enabled
		if adminConfigSvc != nil {
			moduleCode := getModuleCodeFromContext(c)
			if moduleCode != "" {
				featureKey := "feature." + moduleCode + ".enabled"
				if !adminConfigSvc.IsFeatureEnabled(authCtx.TenantID, featureKey) {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"code": 40302, "data": nil, "message": "该功能未开启",
					})
					return
				}
			}
		}

		// 查找 permission_code
		key := c.Request.Method + ":" + fullPath
		mu.RLock()
		code, exists := codeMap[key]
		mu.RUnlock()

		// 缓存未命中时尝试实时查询（覆盖 AutoDiscover 异步延迟场景）
		if !exists || code == "" {
			var freshCode string
			db.Table("admin_api_permission").
				Where("type = 'ENDPOINT' AND http_method = ? AND url_pattern = ? AND permission_code != ''",
					c.Request.Method, fullPath).
				Pluck("permission_code", &freshCode)
			if freshCode != "" {
				mu.Lock()
				codeMap[key] = freshCode
				mu.Unlock()
				code = freshCode
				exists = true
			} else {
				// 检查是否为免检接口（auth_required=0） → 放行
				var authReq int
				result := db.Table("admin_api_permission").
					Where("type = 'ENDPOINT' AND http_method = ? AND url_pattern = ?",
						c.Request.Method, fullPath).
					Pluck("auth_required", &authReq)
				if result.Error == nil && result.RowsAffected > 0 && authReq == 0 {
					c.Next()
					return
				}
			}
		}

		if !exists || code == "" {
			// 无注解且不在白名单且非免检 → 拒绝
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 40301, "data": nil, "message": "权限不足（接口未注册权限码）",
			})
			return
		}

		// 获取用户在当前租户下的所有 permission_code
		permCodes := getUserPermCodes(db, authCtx.UserID, authCtx.TenantID)
		pe := engine.NewPermissionEngine(permCodes)
		if pe.HasPermission(code) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code": 40301, "data": nil, "message": "权限不足",
		})
	}
}

// loadPermissionCodeMap 从数据库加载 method:path → permission_code 映射
func loadPermissionCodeMap(db *gorm.DB) map[string]string {
	var perms []struct {
		HTTPMethod     string `gorm:"column:http_method"`
		URLPattern     string `gorm:"column:url_pattern"`
		PermissionCode string `gorm:"column:permission_code"`
	}
	if err := db.Table("admin_api_permission").
		Where("type = ? AND permission_code != ''", "ENDPOINT").
		Select("http_method, url_pattern, permission_code").
		Find(&perms).Error; err != nil {
		log.Printf("[dynamic-perm] 加载 permission_code 映射失败: %v", err)
		return make(map[string]string)
	}

	result := make(map[string]string, len(perms))
	for _, p := range perms {
		result[p.HTTPMethod+":"+p.URLPattern] = p.PermissionCode
	}
	log.Printf("[dynamic-perm] 加载 %d 条 permission_code 映射", len(result))
	return result
}

// getUserPermCodes 获取用户在当前租户下拥有的所有 permission_code
func getUserPermCodes(db *gorm.DB, userID, tenantID int64) []string {
	var codes []string
	db.Table("admin_api_permission ap").
		Joins("JOIN admin_role_api ra ON ra.api_permission_id = ap.id").
		Joins("JOIN admin_user_role ur ON ur.role_id = ra.role_id").
		Where("ur.user_id = ? AND ur.tenant_id = ? AND ap.permission_code != ''", userID, tenantID).
		Pluck("ap.permission_code", &codes)
	return codes
}

// isSuperAdminByDB 通过数据库查询判断角色列表中是否包含 SUPER_ADMIN
func isSuperAdminByDB(db *gorm.DB, cfg *config.Config, roleIDs []int64) bool {
	if len(roleIDs) == 0 {
		return false
	}
	superAdminCode := cfg.Permission.SuperAdminRole
	if superAdminCode == "" {
		superAdminCode = "SUPER_ADMIN"
	}
	var count int64
	db.Table("admin_role").
		Where("id IN ? AND role_code = ?", roleIDs, superAdminCode).
		Count(&count)
	return count > 0
}

// getModuleCodeFromContext 从 gin.Context 获取当前请求的 module_code
func getModuleCodeFromContext(c *gin.Context) string {
	if mc, exists := c.Get("module_code"); exists {
		if s, ok := mc.(string); ok {
			return s
		}
	}
	return ""
}
