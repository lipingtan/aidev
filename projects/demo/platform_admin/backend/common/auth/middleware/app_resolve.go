package middleware

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"

	"go-admin/common/auth/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AppPrefixMap 应用路由前缀映射（启动时从 admin_application 加载）
type AppPrefixMap struct {
	mu       sync.RWMutex
	prefixes []prefixEntry // 按长度倒序排列（最长优先匹配）
}

type prefixEntry struct {
	Prefix  string
	AppCode string
}

// NewAppPrefixMap 创建空的前缀映射
func NewAppPrefixMap() *AppPrefixMap {
	return &AppPrefixMap{}
}

// Load 从数据库加载 route_prefix → app_code 映射
func (m *AppPrefixMap) Load(db *gorm.DB) {
	var apps []model.Application
	db.Where("route_prefix != '' AND status = 1").Find(&apps)

	entries := make([]prefixEntry, 0, len(apps))
	for _, app := range apps {
		if app.RoutePrefix != "" {
			entries = append(entries, prefixEntry{Prefix: app.RoutePrefix, AppCode: app.AppCode})
		}
	}
	// 按前缀长度倒序（最长优先匹配）
	sort.Slice(entries, func(i, j int) bool {
		return len(entries[i].Prefix) > len(entries[j].Prefix)
	})

	m.mu.Lock()
	m.prefixes = entries
	m.mu.Unlock()

	log.Printf("[app-resolve] 加载 %d 条应用前缀映射", len(entries))
}

// Match 最长前缀匹配，返回 app_code
func (m *AppPrefixMap) Match(path string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, entry := range m.prefixes {
		if strings.HasPrefix(path, entry.Prefix) {
			return entry.AppCode
		}
	}
	return ""
}

// ModuleCodeCache method:path → module_code 缓存
type ModuleCodeCache struct {
	mu    sync.RWMutex
	cache map[string]string // "GET:/api/v1/admin/tenants" → "tenant-mgmt"
}

// NewModuleCodeCache 创建空缓存
func NewModuleCodeCache() *ModuleCodeCache {
	return &ModuleCodeCache{cache: make(map[string]string)}
}

// Load 从 admin_api_permission 加载 method:url_pattern → module_code
func (c *ModuleCodeCache) Load(db *gorm.DB) {
	var perms []struct {
		HTTPMethod string `gorm:"column:http_method"`
		URLPattern string `gorm:"column:url_pattern"`
		ModuleCode string `gorm:"column:module_code"`
	}
	db.Table("admin_api_permission").
		Where("type = 'ENDPOINT' AND module_code != ''").
		Select("http_method, url_pattern, module_code").
		Find(&perms)

	c.mu.Lock()
	c.cache = make(map[string]string, len(perms))
	for _, p := range perms {
		c.cache[p.HTTPMethod+":"+p.URLPattern] = p.ModuleCode
	}
	c.mu.Unlock()

	log.Printf("[app-resolve] 加载 %d 条 module_code 映射", len(perms))
}

// Get 获取 module_code
func (c *ModuleCodeCache) Get(method, path string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache[method+":"+path]
}

// AppResolveMiddleware 应用解析中间件
// 职责：从请求路径匹配 app_code → 校验租户订阅 → 校验模块启用
func AppResolveMiddleware(prefixMap *AppPrefixMap, moduleCache *ModuleCodeCache, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 匹配 app_code
		appCode := prefixMap.Match(path)
		if appCode == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 40303, "data": nil, "message": "无法识别请求所属应用",
			})
			return
		}
		c.Set("app_code", appCode)

		// 获取认证上下文
		authCtx := GetAuthContext(c)
		if authCtx == nil {
			c.Next()
			return
		}

		// SUPER_ADMIN 跳过订阅校验
		if len(authCtx.Roles) > 0 {
			var superCount int64
			db.Table("admin_role").
				Where("id IN ? AND role_code = 'SUPER_ADMIN'", authCtx.Roles).
				Count(&superCount)
			if superCount > 0 {
				c.Set("is_super_admin", true)
				c.Next()
				return
			}
		}

		// 校验租户订阅
		var tenantApp model.TenantApp
		result := db.Where("tenant_id = ? AND app_code = ?", authCtx.TenantID, appCode).First(&tenantApp)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 40302, "data": nil, "message": "租户未开通此应用",
			})
			return
		}

		// 校验模块启用（enabled_modules 非 NULL 时检查）
		if tenantApp.EnabledModules != nil && len(tenantApp.EnabledModules) > 2 {
			moduleCode := moduleCache.Get(c.Request.Method, c.FullPath())
			if moduleCode != "" {
				enabledStr := string(tenantApp.EnabledModules)
				if !strings.Contains(enabledStr, `"`+moduleCode+`"`) {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"code": 40304, "data": nil, "message": "功能模块未启用",
					})
					return
				}
			}
		}

		c.Next()
	}
}
