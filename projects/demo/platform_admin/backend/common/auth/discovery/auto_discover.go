// Package discovery API 自动发现
package discovery

import (
	"log"
	"strings"
	"time"

	"go-admin/common/auth/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AutoDiscover 自动发现 gin 路由并注册为全局模板
// 扫描 engine.Routes()，新 endpoint 插入 admin_api_permission
// 幂等：按 http_method + url_pattern 去重，已存在不重复插入
// asyncThreshold: 路由数超过此阈值时异步执行（goroutine + 5s 延迟）
func AutoDiscover(engine *gin.Engine, db *gorm.DB, asyncThreshold int) {
	routes := engine.Routes()

	if len(routes) >= asyncThreshold {
		go func() {
			time.Sleep(5 * time.Second)
			doDiscover(routes, db)
		}()
		return
	}

	doDiscover(routes, db)
}

// groupDisplayNames GROUP 路径段 → 中文显示名
var groupDisplayNames = map[string]string{
	"users":              "用户管理(users)",
	"roles":              "角色管理(roles)",
	"tenants":            "租户管理(tenants)",
	"resources":          "菜单管理(resources)",
	"api-permissions":    "接口权限(api-permissions)",
	"applications":      "应用管理(applications)",
	"configs":            "系统配置(configs)",
	"login-logs":         "登录日志(login-logs)",
	"operation-logs":     "操作日志(operation-logs)",
	"data-scope-configs": "数据权限(data-scope-configs)",
	"plugins":            "插件管理(plugins)",
	"dashboard":          "仪表盘(dashboard)",
	"monitor":            "系统监控(monitor)",
	"auth":               "认证(auth)",
	"captcha":            "验证码(captcha)",
	"setup":              "系统安装(setup)",
	"sys-apis":           "系统接口(sys-apis)",
	"user-menu":          "用户菜单(user-menu)",
}

// endpointDisplayNames method:path → 中文描述
var endpointDisplayNames = map[string]string{
	"GET:/api/v1/admin/users":                          "获取用户列表",
	"POST:/api/v1/admin/users":                         "创建用户",
	"PUT:/api/v1/admin/users/:id":                      "更新用户",
	"DELETE:/api/v1/admin/users/:id":                   "删除用户",
	"POST:/api/v1/admin/users/:id/tenants":             "关联用户到租户",
	"DELETE:/api/v1/admin/users/:id/tenants/:tenantId": "解除用户租户关联",
	"GET:/api/v1/admin/users/:id/tenants":              "获取用户租户列表",
	"POST:/api/v1/admin/users/:id/roles":               "为用户分配角色",
	"PUT:/api/v1/admin/users/:id/roles":                "替换用户角色",
	"POST:/api/v1/admin/users/:id/force-offline":       "强制下线用户",
	"GET:/api/v1/admin/roles":                          "获取角色列表",
	"POST:/api/v1/admin/roles":                         "创建角色",
	"PUT:/api/v1/admin/roles/:id":                      "更新角色",
	"DELETE:/api/v1/admin/roles/:id":                   "删除角色",
	"GET:/api/v1/admin/roles/:id/resources":            "获取角色菜单权限",
	"PUT:/api/v1/admin/roles/:id/resources":            "分配角色菜单权限",
	"GET:/api/v1/admin/roles/:id/apis":                 "获取角色接口权限",
	"PUT:/api/v1/admin/roles/:id/apis":                 "分配角色接口权限",
	"GET:/api/v1/admin/roles/:id/apps":                 "获取角色应用绑定",
	"PUT:/api/v1/admin/roles/:id/apps":                 "绑定角色应用",
	"GET:/api/v1/admin/roles/:id/permission-summary":   "获取角色权限汇总",
	"GET:/api/v1/admin/roles/:id/data-scopes":          "获取角色数据权限",
	"PUT:/api/v1/admin/roles/:id/data-scopes":          "配置角色数据权限",
	"GET:/api/v1/admin/tenants":                        "获取租户列表",
	"POST:/api/v1/admin/tenants":                       "创建租户",
	"PUT:/api/v1/admin/tenants/:id":                    "更新租户",
	"PUT:/api/v1/admin/tenants/:id/status":             "切换租户状态",
	"DELETE:/api/v1/admin/tenants/:id":                 "删除租户",
	"PUT:/api/v1/admin/tenants/:id/apps":               "设置租户订阅应用",
	"GET:/api/v1/admin/tenants/:id/apps":               "获取租户应用列表",
	"GET:/api/v1/admin/resources/tree":                 "获取菜单树",
	"POST:/api/v1/admin/resources":                     "创建菜单",
	"PUT:/api/v1/admin/resources/:id":                  "更新菜单",
	"DELETE:/api/v1/admin/resources/:id":               "删除菜单",
	"PUT:/api/v1/admin/resources/sort":                 "菜单排序",
	"GET:/api/v1/common/user-menu":                     "获取用户菜单",
	"GET:/api/v1/admin/api-permissions/tree":           "获取接口权限树",
	"POST:/api/v1/admin/api-permissions":               "创建接口权限节点",
	"PUT:/api/v1/admin/api-permissions/:id":            "更新接口权限节点",
	"DELETE:/api/v1/admin/api-permissions/:id":         "删除接口权限节点",
	"PUT:/api/v1/admin/api-permissions/:id/move":       "移动接口权限节点",
	"PUT:/api/v1/admin/api-permissions/:id/visible":    "切换接口权限可见性",
	"GET:/api/v1/admin/api-permissions/unassigned":     "获取待分组接口",
	"GET:/api/v1/admin/applications":                   "获取应用列表",
	"POST:/api/v1/admin/applications":                  "创建应用",
	"PUT:/api/v1/admin/applications/:id":               "更新应用",
	"DELETE:/api/v1/admin/applications/:id":            "删除应用",
	"GET:/api/v1/admin/configs":                        "获取配置列表",
	"GET:/api/v1/admin/configs/:id":                    "获取配置详情",
	"GET:/api/v1/admin/configs/key/:key":               "按Key查询配置",
	"POST:/api/v1/admin/configs":                       "创建配置",
	"PUT:/api/v1/admin/configs/:id":                    "更新配置",
	"DELETE:/api/v1/admin/configs/:id":                 "删除配置",
	"GET:/api/v1/admin/login-logs":                     "获取登录日志",
	"DELETE:/api/v1/admin/login-logs/:id":              "删除登录日志",
	"GET:/api/v1/admin/operation-logs":                 "获取操作日志",
	"GET:/api/v1/admin/data-scope-configs":             "获取数据权限维度",
	"POST:/api/v1/admin/data-scope-configs":            "注册数据权限维度",
	"PUT:/api/v1/admin/data-scope-configs/:id":         "更新数据权限维度",
	"DELETE:/api/v1/admin/data-scope-configs/:id":      "删除数据权限维度",
	"GET:/api/v1/admin/plugins":                        "获取插件列表",
	"POST:/api/v1/admin/plugins/install":               "安装插件",
	"POST:/api/v1/admin/plugins/:name/start":           "启动插件",
	"POST:/api/v1/admin/plugins/:name/stop":            "停止插件",
	"DELETE:/api/v1/admin/plugins/:name":               "卸载插件",
	"GET:/api/v1/admin/plugins/:name/health":           "插件健康检查",
	"POST:/auth/login":                                 "用户登录",
	"POST:/auth/tenant/select":                         "选择租户",
	"POST:/auth/refresh":                               "刷新Token",
	"POST:/auth/logout":                                "用户登出",
}

// endpointPermissionCodes method:path → permission_code 映射
var endpointPermissionCodes = map[string]string{
	"GET:/api/v1/admin/users":                          "user:list",
	"POST:/api/v1/admin/users":                         "user:create",
	"PUT:/api/v1/admin/users/:id":                      "user:update",
	"DELETE:/api/v1/admin/users/:id":                   "user:delete",
	"POST:/api/v1/admin/users/:id/tenants":             "user:tenant:assign",
	"DELETE:/api/v1/admin/users/:id/tenants/:tenantId": "user:tenant:remove",
	"GET:/api/v1/admin/users/:id/tenants":              "user:tenant:list",
	"POST:/api/v1/admin/users/:id/roles":               "user:role:assign",
	"PUT:/api/v1/admin/users/:id/roles":                "user:role:replace",
	"POST:/api/v1/admin/users/:id/force-offline":       "user:force-offline",
	"GET:/api/v1/admin/roles":                          "role:list",
	"POST:/api/v1/admin/roles":                         "role:create",
	"PUT:/api/v1/admin/roles/:id":                      "role:update",
	"DELETE:/api/v1/admin/roles/:id":                   "role:delete",
	"GET:/api/v1/admin/roles/:id/resources":            "role:resource:list",
	"PUT:/api/v1/admin/roles/:id/resources":            "role:resource:assign",
	"GET:/api/v1/admin/roles/:id/apis":                 "role:api:list",
	"PUT:/api/v1/admin/roles/:id/apis":                 "role:api:assign",
	"GET:/api/v1/admin/roles/:id/apps":                 "role:app:list",
	"PUT:/api/v1/admin/roles/:id/apps":                 "role:app:assign",
	"GET:/api/v1/admin/roles/:id/permission-summary":   "role:permission:summary",
	"GET:/api/v1/admin/roles/:id/data-scopes":          "role:data-scope:list",
	"PUT:/api/v1/admin/roles/:id/data-scopes":          "role:data-scope:assign",
	"GET:/api/v1/admin/tenants":                        "tenant:list",
	"POST:/api/v1/admin/tenants":                       "tenant:create",
	"PUT:/api/v1/admin/tenants/:id":                    "tenant:update",
	"PUT:/api/v1/admin/tenants/:id/status":             "tenant:status",
	"DELETE:/api/v1/admin/tenants/:id":                 "tenant:delete",
	"PUT:/api/v1/admin/tenants/:id/apps":               "tenant:app:assign",
	"GET:/api/v1/admin/tenants/:id/apps":               "tenant:app:list",
	"GET:/api/v1/admin/resources/tree":                 "resource:tree",
	"POST:/api/v1/admin/resources":                     "resource:create",
	"PUT:/api/v1/admin/resources/:id":                  "resource:update",
	"DELETE:/api/v1/admin/resources/:id":               "resource:delete",
	"PUT:/api/v1/admin/resources/sort":                 "resource:sort",
	"GET:/api/v1/admin/api-permissions/tree":           "api-perm:tree",
	"POST:/api/v1/admin/api-permissions":               "api-perm:create",
	"PUT:/api/v1/admin/api-permissions/:id":            "api-perm:update",
	"DELETE:/api/v1/admin/api-permissions/:id":         "api-perm:delete",
	"PUT:/api/v1/admin/api-permissions/:id/move":       "api-perm:move",
	"PUT:/api/v1/admin/api-permissions/:id/visible":    "api-perm:visible",
	"GET:/api/v1/admin/api-permissions/unassigned":     "api-perm:unassigned",
	"GET:/api/v1/admin/applications":                   "app:list",
	"POST:/api/v1/admin/applications":                  "app:create",
	"PUT:/api/v1/admin/applications/:id":               "app:update",
	"DELETE:/api/v1/admin/applications/:id":            "app:delete",
	"GET:/api/v1/admin/configs":                        "config:list",
	"GET:/api/v1/admin/configs/:id":                    "config:detail",
	"GET:/api/v1/admin/configs/key/:key":               "config:detail",
	"POST:/api/v1/admin/configs":                       "config:create",
	"PUT:/api/v1/admin/configs/:id":                    "config:update",
	"DELETE:/api/v1/admin/configs/:id":                 "config:delete",
	"GET:/api/v1/admin/login-logs":                     "log:login:list",
	"DELETE:/api/v1/admin/login-logs/:id":              "log:login:delete",
	"GET:/api/v1/admin/operation-logs":                 "log:operation:list",
	"GET:/api/v1/admin/data-scope-configs":             "data-scope:list",
	"POST:/api/v1/admin/data-scope-configs":            "data-scope:create",
	"PUT:/api/v1/admin/data-scope-configs/:id":         "data-scope:update",
	"DELETE:/api/v1/admin/data-scope-configs/:id":      "data-scope:delete",
	"GET:/api/v1/admin/plugins":                        "plugin:list",
	"POST:/api/v1/admin/plugins/install":               "plugin:install",
	"POST:/api/v1/admin/plugins/:name/start":           "plugin:start",
	"POST:/api/v1/admin/plugins/:name/stop":            "plugin:stop",
	"DELETE:/api/v1/admin/plugins/:name":               "plugin:uninstall",
	"GET:/api/v1/admin/plugins/:name/health":           "plugin:health",
}

// moduleCodeMap 资源路径段 → module_code 映射
var moduleCodeMap = map[string]string{
	"users":              "user-mgmt",
	"roles":              "role-mgmt",
	"tenants":            "tenant-mgmt",
	"resources":          "resource-mgmt",
	"api-permissions":    "api-perm-mgmt",
	"applications":      "app-mgmt",
	"configs":            "config-mgmt",
	"login-logs":         "log-mgmt",
	"operation-logs":     "log-mgmt",
	"data-scope-configs": "data-scope-mgmt",
	"plugins":            "plugin-mgmt",
	"dashboard":          "dashboard",
	"monitor":            "monitor",
}

// hiddenGroups 默认隐藏的分组列表
var hiddenGroups = map[string]bool{
	"auth":      true,
	"captcha":   true,
	"dashboard": true,
	"plugins":   true,
	"monitor":   true,
	"sys-apis":  true,
	"setup":     true,
	"user-menu": true,
}

// doDiscover 执行实际的路由发现和注册逻辑
// 自动按 URL 资源路径创建 GROUP（对象），ENDPOINT 归入对应 GROUP
func doDiscover(routes gin.RoutesInfo, db *gorm.DB) {
	// 查询所有已有 ENDPOINT 记录用于去重
	var existing []model.ApiPermission
	if err := db.Where("type = 'ENDPOINT'").Find(&existing).Error; err != nil {
		log.Printf("[discovery] 查询已有记录失败: %v", err)
		return
	}

	// 已有 ENDPOINT 去重 map
	existingEPMap := make(map[string]struct{})
	for _, ep := range existing {
		existingEPMap[ep.HTTPMethod+":"+ep.URLPattern] = struct{}{}
	}

	// 补充已存在但 permission_code 为空的记录
	for _, ep := range existing {
		if ep.PermissionCode != "" {
			continue
		}
		key := ep.HTTPMethod + ":" + ep.URLPattern
		if code, ok := endpointPermissionCodes[key]; ok && code != "" {
			db.Model(&model.ApiPermission{}).Where("id = ?", ep.ID).Update("permission_code", code)
		}
	}

	// 补充已存在但 module_code 为空的记录
	for _, ep := range existing {
		if ep.ModuleCode != "" {
			continue
		}
		mc := inferModuleCode(ep.URLPattern)
		if mc != "" {
			db.Model(&model.ApiPermission{}).Where("id = ?", ep.ID).Update("module_code", mc)
		}
	}

	// 查已有 GROUP（按 app_code 分）
	var existingGroups []model.ApiPermission
	db.Where("type = 'GROUP'").Find(&existingGroups)
	existingGroupMap := make(map[string]int64) // "app_code:groupName" → ID
	for _, g := range existingGroups {
		existingGroupMap[g.AppCode+":"+g.Name] = g.ID
	}

	// 按资源路径提取分组名并分类
	type endpointEntry struct {
		perm      model.ApiPermission
		groupName string
	}
	var newEndpoints []endpointEntry

	for _, route := range routes {
		// 只扫描 /api/v1/ 前缀的路由
		if !strings.HasPrefix(route.Path, "/api/v1/") {
			continue
		}

		key := route.Method + ":" + route.Path
		if _, exists := existingEPMap[key]; exists {
			continue
		}

		appCode := resolveAppCode(route.Path)
		moduleCode := inferModuleCode(route.Path)
		groupName := extractGroupName(route.Path)
		displayName := endpointDisplayNames[key]
		permCode := endpointPermissionCodes[key]

		newEndpoints = append(newEndpoints, endpointEntry{
			groupName: groupName,
			perm: model.ApiPermission{
				Type:           "ENDPOINT",
				Name:           route.Path,
				DisplayName:    displayName,
				PermissionCode: permCode,
				URLPattern:     route.Path,
				HTTPMethod:     route.Method,
				AppCode:        appCode,
				ModuleCode:     moduleCode,
				Status:         "ACTIVE",
			},
		})
	}

	if len(newEndpoints) == 0 {
		return
	}

	// 按 groupName + appCode 分组
	type groupKey struct {
		appCode   string
		groupName string
	}
	grouped := make(map[groupKey][]model.ApiPermission)
	for _, e := range newEndpoints {
		k := groupKey{appCode: e.perm.AppCode, groupName: e.groupName}
		grouped[k] = append(grouped[k], e.perm)
	}

	// 逐组创建 GROUP + ENDPOINT
	for gk, endpoints := range grouped {
		mapKey := gk.appCode + ":" + gk.groupName
		var groupID int64
		if id, ok := existingGroupMap[mapKey]; ok {
			groupID = id
		} else {
			// 创建 GROUP
			group := &model.ApiPermission{
				Type:        "GROUP",
				Name:        gk.groupName,
				DisplayName: groupDisplayNames[gk.groupName],
				AppCode:     gk.appCode,
				ModuleCode:  moduleCodeMap[gk.groupName],
				Status:      "ACTIVE",
				Visible:     intPtr(1),
				AuthRequired: intPtr(1),
			}
			if hiddenGroups[gk.groupName] {
				group.Visible = intPtr(0)
				group.AuthRequired = intPtr(0)
			}
			if err := db.Create(group).Error; err != nil {
				log.Printf("[discovery] 创建 GROUP [%s] 失败: %v", gk.groupName, err)
				continue
			}
			groupID = group.ID
			existingGroupMap[mapKey] = groupID
		}

		// 设置 parent_id 并批量插入
		for i := range endpoints {
			endpoints[i].ParentID = &groupID
			if hiddenGroups[gk.groupName] {
				endpoints[i].Visible = intPtr(0)
				endpoints[i].AuthRequired = intPtr(0)
			} else {
				endpoints[i].Visible = intPtr(1)
				endpoints[i].AuthRequired = intPtr(1)
			}
		}
		if err := db.Create(&endpoints).Error; err != nil {
			log.Printf("[discovery] 批量插入 GROUP [%s] 的 endpoint 失败: %v", gk.groupName, err)
		}
	}

	log.Printf("[discovery] 发现并注册 %d 条新 endpoint", len(newEndpoints))
}

// resolveAppCode 从路由路径推导 app_code
// /api/v1/admin/xxx → "platform_admin"
// /api/v1/common/xxx → "platform_admin"（公共接口也归属管理端应用）
func resolveAppCode(path string) string {
	if strings.HasPrefix(path, "/api/v1/admin/") {
		return "platform_admin"
	}
	if strings.HasPrefix(path, "/api/v1/common/") {
		return "platform_admin"
	}
	// 插件路由等其他前缀，暂时归属 platform_admin
	return "platform_admin"
}

// inferModuleCode 从路由路径推导 module_code
// /api/v1/admin/tenants → "tenant-mgmt"
// /api/v1/admin/users/:id/roles → "user-mgmt"
// /api/v1/common/user-menu → "resource-mgmt"
func inferModuleCode(path string) string {
	parts := splitPath(path)
	// 跳过前缀段：api, v1, admin/common
	skipSet := map[string]bool{"api": true, "v1": true, "admin": true, "common": true}
	for _, p := range parts {
		if skipSet[p] {
			continue
		}
		if len(p) > 0 && p[0] == ':' {
			continue
		}
		if mc, ok := moduleCodeMap[p]; ok {
			return mc
		}
		return ""
	}
	return ""
}

// extractGroupName 从 URL 路径提取对象分组名
// /api/v1/admin/users → "users"
// /api/v1/admin/users/:id → "users"
// /api/v1/admin/roles/:id/resources → "roles"
// /api/v1/common/user-menu → "user-menu"
func extractGroupName(path string) string {
	parts := splitPath(path)
	// 跳过前缀段
	skipSet := map[string]bool{"api": true, "v1": true, "v2": true, "admin": true, "common": true}
	for _, p := range parts {
		if skipSet[p] {
			continue
		}
		if len(p) > 0 && p[0] == ':' {
			continue
		}
		return p
	}
	return "other"
}

// splitPath 将路径按 / 分割为非空段
func splitPath(path string) []string {
	var result []string
	for _, seg := range strings.Split(path, "/") {
		if seg != "" {
			result = append(result, seg)
		}
	}
	return result
}

// intPtr 返回 int 指针（用于 GORM 写入零值）
func intPtr(v int) *int {
	return &v
}
