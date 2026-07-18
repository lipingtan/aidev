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
// 扫描 engine.Routes()，新 endpoint 插入 admin_api_permission (tenant_id=0, status=UNASSIGNED)
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
}

// endpointDisplayNames method:path → 中文描述
var endpointDisplayNames = map[string]string{
	"GET:/api/v1/users":                          "获取用户列表",
	"POST:/api/v1/users":                         "创建用户",
	"PUT:/api/v1/users/:id":                      "更新用户",
	"DELETE:/api/v1/users/:id":                   "删除用户",
	"POST:/api/v1/users/:id/tenants":             "关联用户到租户",
	"DELETE:/api/v1/users/:id/tenants/:tenantId": "解除用户租户关联",
	"GET:/api/v1/users/:id/tenants":              "获取用户租户列表",
	"POST:/api/v1/users/:id/roles":               "为用户分配角色",
	"PUT:/api/v1/users/:id/roles":                "替换用户角色",
	"POST:/api/v1/users/:id/force-offline":       "强制下线用户",
	"GET:/api/v1/roles":                          "获取角色列表",
	"POST:/api/v1/roles":                         "创建角色",
	"PUT:/api/v1/roles/:id":                      "更新角色",
	"DELETE:/api/v1/roles/:id":                   "删除角色",
	"GET:/api/v1/roles/:id/resources":            "获取角色菜单权限",
	"PUT:/api/v1/roles/:id/resources":            "分配角色菜单权限",
	"GET:/api/v1/roles/:id/apis":                 "获取角色接口权限",
	"PUT:/api/v1/roles/:id/apis":                 "分配角色接口权限",
	"GET:/api/v1/roles/:id/apps":                 "获取角色应用绑定",
	"PUT:/api/v1/roles/:id/apps":                 "绑定角色应用",
	"GET:/api/v1/roles/:id/permission-summary":   "获取角色权限汇总",
	"GET:/api/v1/roles/:id/data-scopes":          "获取角色数据权限",
	"PUT:/api/v1/roles/:id/data-scopes":          "配置角色数据权限",
	"GET:/api/v1/tenants":                        "获取租户列表",
	"POST:/api/v1/tenants":                       "创建租户",
	"PUT:/api/v1/tenants/:id":                    "更新租户",
	"PUT:/api/v1/tenants/:id/status":             "切换租户状态",
	"DELETE:/api/v1/tenants/:id":                 "删除租户",
	"PUT:/api/v1/tenants/:id/apps":               "设置租户订阅应用",
	"GET:/api/v1/tenants/:id/apps":               "获取租户应用列表",
	"GET:/api/v1/resources/tree":                 "获取菜单树",
	"POST:/api/v1/resources":                     "创建菜单",
	"PUT:/api/v1/resources/:id":                  "更新菜单",
	"DELETE:/api/v1/resources/:id":               "删除菜单",
	"PUT:/api/v1/resources/sort":                 "菜单排序",
	"GET:/api/v1/resources/user-menu":            "获取用户菜单",
	"GET:/api/v1/api-permissions/tree":           "获取接口权限树",
	"POST:/api/v1/api-permissions":               "创建接口权限节点",
	"PUT:/api/v1/api-permissions/:id":            "更新接口权限节点",
	"DELETE:/api/v1/api-permissions/:id":         "删除接口权限节点",
	"PUT:/api/v1/api-permissions/:id/move":       "移动接口权限节点",
	"PUT:/api/v1/api-permissions/:id/visible":    "切换接口权限可见性",
	"GET:/api/v1/api-permissions/unassigned":     "获取待分组接口",
	"GET:/api/v1/applications":                   "获取应用列表",
	"POST:/api/v1/applications":                  "创建应用",
	"PUT:/api/v1/applications/:id":               "更新应用",
	"DELETE:/api/v1/applications/:id":            "删除应用",
	"GET:/api/v1/configs":                        "获取配置列表",
	"GET:/api/v1/configs/:id":                    "获取配置详情",
	"GET:/api/v1/configs/key/:key":               "按Key查询配置",
	"POST:/api/v1/configs":                       "创建配置",
	"PUT:/api/v1/configs/:id":                    "更新配置",
	"DELETE:/api/v1/configs/:id":                 "删除配置",
	"GET:/api/v1/login-logs":                     "获取登录日志",
	"DELETE:/api/v1/login-logs/:id":              "删除登录日志",
	"GET:/api/v1/operation-logs":                 "获取操作日志",
	"GET:/api/v1/data-scope-configs":             "获取数据权限维度",
	"POST:/api/v1/data-scope-configs":            "注册数据权限维度",
	"PUT:/api/v1/data-scope-configs/:id":         "更新数据权限维度",
	"DELETE:/api/v1/data-scope-configs/:id":      "删除数据权限维度",
	"GET:/api/v1/plugins":                        "获取插件列表",
	"POST:/api/v1/plugins/install":               "安装插件",
	"POST:/api/v1/plugins/:name/start":           "启动插件",
	"POST:/api/v1/plugins/:name/stop":            "停止插件",
	"DELETE:/api/v1/plugins/:name":               "卸载插件",
	"GET:/api/v1/plugins/:name/health":           "插件健康检查",
	"POST:/auth/login":                           "用户登录",
	"POST:/auth/tenant/select":                   "选择租户",
	"POST:/auth/refresh":                         "刷新Token",
	"POST:/auth/logout":                          "用户登出",
}

// endpointPermissionCodes method:path → permission_code 映射
var endpointPermissionCodes = map[string]string{
	"GET:/api/v1/users":                          "user:list",
	"POST:/api/v1/users":                         "user:create",
	"PUT:/api/v1/users/:id":                      "user:update",
	"DELETE:/api/v1/users/:id":                   "user:delete",
	"POST:/api/v1/users/:id/tenants":             "user:tenant:assign",
	"DELETE:/api/v1/users/:id/tenants/:tenantId": "user:tenant:remove",
	"GET:/api/v1/users/:id/tenants":              "user:tenant:list",
	"POST:/api/v1/users/:id/roles":               "user:role:assign",
	"PUT:/api/v1/users/:id/roles":                "user:role:replace",
	"POST:/api/v1/users/:id/force-offline":       "user:force-offline",
	"GET:/api/v1/roles":                          "role:list",
	"POST:/api/v1/roles":                         "role:create",
	"PUT:/api/v1/roles/:id":                      "role:update",
	"DELETE:/api/v1/roles/:id":                   "role:delete",
	"GET:/api/v1/roles/:id/resources":            "role:resource:list",
	"PUT:/api/v1/roles/:id/resources":            "role:resource:assign",
	"GET:/api/v1/roles/:id/apis":                 "role:api:list",
	"PUT:/api/v1/roles/:id/apis":                 "role:api:assign",
	"GET:/api/v1/roles/:id/apps":                 "role:app:list",
	"PUT:/api/v1/roles/:id/apps":                 "role:app:assign",
	"GET:/api/v1/roles/:id/permission-summary":   "role:permission:summary",
	"GET:/api/v1/roles/:id/data-scopes":          "role:data-scope:list",
	"PUT:/api/v1/roles/:id/data-scopes":          "role:data-scope:assign",
	"GET:/api/v1/tenants":                        "tenant:list",
	"POST:/api/v1/tenants":                       "tenant:create",
	"PUT:/api/v1/tenants/:id":                    "tenant:update",
	"PUT:/api/v1/tenants/:id/status":             "tenant:status",
	"DELETE:/api/v1/tenants/:id":                 "tenant:delete",
	"PUT:/api/v1/tenants/:id/apps":               "tenant:app:assign",
	"GET:/api/v1/tenants/:id/apps":               "tenant:app:list",
	"GET:/api/v1/resources/tree":                 "resource:tree",
	"POST:/api/v1/resources":                     "resource:create",
	"PUT:/api/v1/resources/:id":                  "resource:update",
	"DELETE:/api/v1/resources/:id":               "resource:delete",
	"PUT:/api/v1/resources/sort":                 "resource:sort",
	"GET:/api/v1/api-permissions/tree":           "api-perm:tree",
	"POST:/api/v1/api-permissions":               "api-perm:create",
	"PUT:/api/v1/api-permissions/:id":            "api-perm:update",
	"DELETE:/api/v1/api-permissions/:id":         "api-perm:delete",
	"PUT:/api/v1/api-permissions/:id/move":       "api-perm:move",
	"PUT:/api/v1/api-permissions/:id/visible":    "api-perm:visible",
	"GET:/api/v1/api-permissions/unassigned":     "api-perm:unassigned",
	"GET:/api/v1/applications":                   "app:list",
	"POST:/api/v1/applications":                  "app:create",
	"PUT:/api/v1/applications/:id":               "app:update",
	"DELETE:/api/v1/applications/:id":            "app:delete",
	"GET:/api/v1/configs":                        "config:list",
	"GET:/api/v1/configs/:id":                    "config:detail",
	"GET:/api/v1/configs/key/:key":               "config:detail",
	"POST:/api/v1/configs":                       "config:create",
	"PUT:/api/v1/configs/:id":                    "config:update",
	"DELETE:/api/v1/configs/:id":                 "config:delete",
	"GET:/api/v1/login-logs":                     "log:login:list",
	"DELETE:/api/v1/login-logs/:id":              "log:login:delete",
	"GET:/api/v1/operation-logs":                 "log:operation:list",
	"GET:/api/v1/data-scope-configs":             "data-scope:list",
	"POST:/api/v1/data-scope-configs":            "data-scope:create",
	"PUT:/api/v1/data-scope-configs/:id":         "data-scope:update",
	"DELETE:/api/v1/data-scope-configs/:id":      "data-scope:delete",
	"GET:/api/v1/plugins":                        "plugin:list",
	"POST:/api/v1/plugins/install":               "plugin:install",
	"POST:/api/v1/plugins/:name/start":           "plugin:start",
	"POST:/api/v1/plugins/:name/stop":            "plugin:stop",
	"DELETE:/api/v1/plugins/:name":               "plugin:uninstall",
	"GET:/api/v1/plugins/:name/health":           "plugin:health",
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
}

// doDiscover 执行实际的路由发现和注册逻辑
// 自动按 URL 资源路径创建 GROUP（对象），ENDPOINT 归入对应 GROUP
func doDiscover(routes gin.RoutesInfo, db *gorm.DB) {
	// 查询所有已有记录（含已分组和未分组），用于去重
	var existing []model.ApiPermission
	if err := db.Where("app_code = ?", "admin").Find(&existing).Error; err != nil {
		log.Printf("[discovery] 查询已有记录失败: %v", err)
		return
	}

	// 已有 ENDPOINT 去重 map
	existingEPMap := make(map[string]struct{})
	for _, ep := range existing {
		if ep.Type == "ENDPOINT" {
			existingEPMap[ep.HTTPMethod+":"+ep.URLPattern] = struct{}{}
		}
	}

	// 补充已存在但 permission_code 为空的记录
	for _, ep := range existing {
		if ep.Type != "ENDPOINT" || ep.PermissionCode != "" {
			continue
		}
		key := ep.HTTPMethod + ":" + ep.URLPattern
		if code, ok := endpointPermissionCodes[key]; ok && code != "" {
			db.Model(&model.ApiPermission{}).Where("id = ?", ep.ID).Update("permission_code", code)
		}
	}

	// 已有 GROUP 名称 map
	existingGroupMap := make(map[string]int64)
	for _, g := range existing {
		if g.Type == "GROUP" {
			existingGroupMap[g.Name] = g.ID
		}
	}

	// 按资源路径提取分组名
	// /api/v1/users → "用户(users)"
	// /api/v1/roles/:id/resources → "角色(roles)"
	// /auth/login → "认证(auth)"
	groupEndpoints := make(map[string][]model.ApiPermission) // groupName → endpoints

	for _, route := range routes {
		key := route.Method + ":" + route.Path
		if _, exists := existingEPMap[key]; exists {
			continue
		}

		groupName := extractGroupName(route.Path)
		displayName := endpointDisplayNames[key]
		permCode := endpointPermissionCodes[key]
		groupEndpoints[groupName] = append(groupEndpoints[groupName], model.ApiPermission{
			Type:           "ENDPOINT",
			Name:           route.Path,
			DisplayName:    displayName,
			PermissionCode: permCode,
			URLPattern:     route.Path,
			HTTPMethod:     route.Method,
			AppCode:        "admin",
			Status:         "ACTIVE",
		})
	}

	if len(groupEndpoints) == 0 {
		return
	}

	// 逐组创建 GROUP + ENDPOINT
	for groupName, endpoints := range groupEndpoints {
		var groupID int64
		if id, ok := existingGroupMap[groupName]; ok {
			groupID = id
		} else {
			// 创建 GROUP
			group := &model.ApiPermission{
				Type:         "GROUP",
				Name:         groupName,
				DisplayName:  groupDisplayNames[groupName],
				AppCode:      "admin",
				Status:       "ACTIVE",
				Visible:      intPtr(1),
				AuthRequired: intPtr(1),
			}
			if hiddenGroups[groupName] {
				group.Visible = intPtr(0)
				group.AuthRequired = intPtr(0)
			}
			if err := db.Create(group).Error; err != nil {
				log.Printf("[discovery] 创建 GROUP [%s] 失败: %v", groupName, err)
				continue
			}
			groupID = group.ID
			existingGroupMap[groupName] = groupID
		}

		// 设置 parent_id 并批量插入
		for i := range endpoints {
			endpoints[i].ParentID = &groupID
			if hiddenGroups[groupName] {
				endpoints[i].Visible = intPtr(0)
				endpoints[i].AuthRequired = intPtr(0)
			} else {
				endpoints[i].Visible = intPtr(1)
				endpoints[i].AuthRequired = intPtr(1)
			}
		}
		if err := db.Create(&endpoints).Error; err != nil {
			log.Printf("[discovery] 批量插入 GROUP [%s] 的 endpoint 失败: %v", groupName, err)
		}
	}
}

// extractGroupName 从 URL 路径提取对象分组名
// /api/v1/users → "users"
// /api/v1/users/:id → "users"
// /api/v1/roles/:id/resources → "roles"
// /auth/login → "auth"
// /setup/status → "setup"
// /dashboard/kpi → "dashboard"
func extractGroupName(path string) string {
	// 按 / 分割，找第一个非空、非版本号、非参数的路径段
	parts := splitPath(path)

	// 跳过 api/v1 等前缀
	skipPrefixes := map[string]bool{"api": true, "v1": true, "v2": true}
	for _, p := range parts {
		if skipPrefixes[p] {
			continue
		}
		if len(p) > 0 && p[0] == ':' {
			continue // 跳过路径参数
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
