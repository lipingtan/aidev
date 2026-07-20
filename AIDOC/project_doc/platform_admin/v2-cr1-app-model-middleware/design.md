# 设计：CR-1 应用模型升级 + 中间件链重构

## 技术方案

### 1. DDL 变更（Model 层加字段）

#### admin_application 新增字段

```go
type Application struct {
    ID          int64           `gorm:"primaryKey" json:"id,string"`
    AppCode     string          `gorm:"size:64;uniqueIndex;not null" json:"app_code"`
    Name        string          `gorm:"size:128;not null" json:"name"`
    Description string          `gorm:"size:512" json:"description"`
    AppType     string          `gorm:"size:16;not null;default:BUILTIN;comment:BUILTIN/PLUGIN/EXTERNAL" json:"app_type"`
    RoutePrefix string          `gorm:"size:128;comment:API路由前缀" json:"route_prefix"`
    Platforms   datatypes.JSON  `gorm:"type:json;not null;default:'[]';comment:支持的平台" json:"platforms"`
    Modules     datatypes.JSON  `gorm:"type:json;comment:功能模块声明" json:"modules"`
    AppConfig   datatypes.JSON  `gorm:"type:json;comment:应用级配置" json:"app_config"`
    Icon        string          `gorm:"size:256" json:"icon"`
    SortOrder   int             `gorm:"not null;default:0" json:"sort_order"`
    Status      int             `gorm:"not null;default:1" json:"status"`
    Version     int             `gorm:"not null;default:0" json:"version"`
    DeletedAt   gorm.DeletedAt  `json:"-"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
}
```

#### admin_resource 新增字段

```go
// 新增:
Platform   string `gorm:"size:16;not null;default:admin;comment:归属前端平台admin/user" json:"platform"`
ModuleCode string `gorm:"size:64;comment:所属功能模块" json:"module_code"`
```

#### admin_api_permission 新增字段

```go
// 新增:
ModuleCode string `gorm:"size:64;comment:所属功能模块" json:"module_code"`
```

#### admin_tenant_app 新增字段

```go
// 新增:
EnabledModules datatypes.JSON `gorm:"type:json;comment:启用的模块列表,NULL=全部" json:"enabled_modules"`
UpdatedAt      time.Time      `json:"updated_at"`
```

#### admin_tenant 新增字段

```go
// 新增:
Timezone  string     `gorm:"size:64;not null;default:Asia/Shanghai;comment:租户时区" json:"timezone"`
Locale    string     `gorm:"size:16;not null;default:zh-CN;comment:默认语言" json:"locale"`
Currency  string     `gorm:"size:8;not null;default:CNY;comment:默认币种" json:"currency"`
ExpiredAt *time.Time `gorm:"comment:试用/订阅到期时间" json:"expired_at"`
// Status 字段 COMMENT 更新为: 1=正常 0=禁用 2=只读 3=注销中
```

### 2. 路由重构

#### 路由注册结构

```go
func RegisterRoutes(engine *gin.Engine, deps *Dependencies) {
    // 认证路由（公开，不走任何中间件）
    auth := engine.Group("/auth")
    deps.AuthHandler.RegisterRoutes(auth)

    // 公共路由（仅需认证，不走 AppResolve 和 Permission）
    common := engine.Group("/api/v1/common")
    common.Use(middleware.AuthMiddleware(deps.AuthService))
    {
        common.GET("/user-menu", deps.ResourceHandler.GetUserMenu)
        common.GET("/user-info", deps.UserHandler.GetCurrentUserInfo)
    }

    // 管理端应用路由（完整中间件链）
    admin := engine.Group("/api/v1/admin")
    admin.Use(middleware.AuthMiddleware(deps.AuthService))
    admin.Use(middleware.AppResolveMiddleware(deps.AppPrefixMap, publicPaths))
    admin.Use(middleware.DynamicPermissionMiddleware(deps.DB, deps.Cfg))
    {
        deps.TenantHandler.RegisterRoutes(admin)
        deps.UserHandler.RegisterRoutes(admin)
        deps.RoleHandler.RegisterRoutes(admin)
        deps.ResourceHandler.RegisterRoutes(admin)
        deps.ApiPermissionHandler.RegisterRoutes(admin)
        deps.ApplicationHandler.RegisterRoutes(admin)
        deps.DataScopeHandler.RegisterRoutes(admin)
        deps.ConfigHandler.RegisterRoutes(admin)
        deps.OperationLogHandler.RegisterRoutes(admin)
        deps.LoginLogHandler.RegisterRoutes(admin)
    }
}
```

### 3. AppResolveMiddleware 实现

```go
// AppPrefixMap 应用路由前缀映射（启动时从 admin_application 加载）
type AppPrefixMap struct {
    prefixes []prefixEntry // 按长度倒序排列（最长优先匹配）
}

type prefixEntry struct {
    Prefix  string
    AppCode string
}

// Match 最长前缀匹配
func (m *AppPrefixMap) Match(path string) string {
    for _, entry := range m.prefixes {
        if strings.HasPrefix(path, entry.Prefix) {
            return entry.AppCode
        }
    }
    return ""
}

// AppResolveMiddleware 中间件实现
func AppResolveMiddleware(prefixMap *AppPrefixMap, publicPaths map[string]bool) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 公共白名单放行
        if publicPaths[c.FullPath()] {
            c.Next()
            return
        }

        // 匹配 app_code
        appCode := prefixMap.Match(c.Request.URL.Path)
        if appCode == "" {
            c.AbortWithStatusJSON(403, gin.H{"code": 40303, "data": nil, "message": "无法识别请求所属应用"})
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
        if isSuperAdmin(c) {
            c.Next()
            return
        }

        // 校验租户订阅
        tenantApp := getTenantApp(authCtx.TenantID, appCode)
        if tenantApp == nil {
            c.AbortWithStatusJSON(403, gin.H{"code": 40302, "data": nil, "message": "租户未开通此应用"})
            return
        }

        // 校验模块启用（enabled_modules 非 NULL 时检查）
        if tenantApp.EnabledModules != nil {
            moduleCode := getModuleCodeForPath(c.FullPath(), c.Request.Method, appCode)
            if moduleCode != "" && !isModuleEnabled(tenantApp.EnabledModules, moduleCode) {
                c.AbortWithStatusJSON(403, gin.H{"code": 40304, "data": nil, "message": "功能模块未启用"})
                return
            }
        }

        c.Next()
    }
}

// getModuleCodeForPath 从 admin_api_permission 中查找该路由的 module_code
func getModuleCodeForPath(fullPath, method, appCode string) string {
    // 从缓存获取 method:path → module_code 映射
    key := method + ":" + fullPath
    if moduleCode, ok := moduleCodeMap.Load(key); ok {
        return moduleCode.(string)
    }
    return ""
}
```

### 4. AuthMiddleware 增强（租户四态 + 自动降级）

```go
func AuthMiddleware(authSvc *service.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... 现有 token 解析逻辑 ...

        // 检查租户状态（增强）
        tenant := authSvc.GetTenantByID(claims.TenantID)
        if tenant == nil {
            c.AbortWithStatusJSON(403, ...)
            return
        }

        // 自动降级：expired_at 已过期 + 当前状态为 ACTIVE → 自动切换为 READ_ONLY
        if tenant.Status == 1 && tenant.ExpiredAt != nil && tenant.ExpiredAt.Before(time.Now()) {
            // CAS 更新避免并发重复
            authSvc.AutoDegradeToReadOnly(tenant.ID)
            tenant.Status = 2
        }

        switch tenant.Status {
        case 1: // ACTIVE
            // 放行
        case 2: // READ_ONLY
            if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
                c.AbortWithStatusJSON(403, gin.H{"code": 40305, "data": nil, "message": "租户已降为只读模式"})
                return
            }
        case 0, 3: // DISABLED, CANCELLING
            c.AbortWithStatusJSON(403, gin.H{"code": 40105, "data": nil, "message": "租户已禁用"})
            return
        }

        // ... 注入 AuthContext ...
        c.Next()
    }
}

// AutoDegradeToReadOnly CAS 条件更新
func (s *AuthService) AutoDegradeToReadOnly(tenantID int64) {
    s.db.Model(&model.Tenant{}).
        Where("id = ? AND status = 1 AND expired_at < ?", tenantID, time.Now()).
        Update("status", 2)
}
```

### 5. GetUserMenu 增强

```go
func (s *ResourceService) GetUserMenu(tenantID, userID int64, platform string) ([]*ResourceNode, error) {
    // 查用户角色
    var roleIDs []int64
    s.db.Model(&model.UserRole{}).
        Where("user_id = ? AND tenant_id = ?", userID, tenantID).
        Pluck("role_id", &roleIDs)

    if len(roleIDs) == 0 {
        return []*ResourceNode{}, nil
    }

    // SUPER_ADMIN → 返回对应 platform 的所有菜单
    if s.isSuperAdmin(roleIDs) {
        var resources []model.Resource
        s.db.Where("platform = ?", platform).
            Order("sort_order ASC, created_at ASC").
            Find(&resources)
        return buildTree(resources), nil
    }

    // 普通角色：
    // 1. 查租户已订阅的应用及其 enabled_modules
    tenantApps := s.getTenantApps(tenantID) // [{appCode, enabledModules}]
    if len(tenantApps) == 0 {
        return []*ResourceNode{}, nil
    }

    // 2. 查角色已分配的资源 ID
    var resourceIDs []int64
    s.db.Model(&model.RoleResource{}).
        Where("role_id IN ?", roleIDs).
        Pluck("resource_id", &resourceIDs)
    if len(resourceIDs) == 0 {
        return []*ResourceNode{}, nil
    }

    // 3. 构建过滤条件
    appCodes := extractAppCodes(tenantApps)
    query := s.db.Where("id IN ? AND app_code IN ? AND platform = ?", resourceIDs, appCodes, platform)

    // 4. enabled_modules 过滤
    // 对每个应用，如果 enabled_modules 非 NULL，排除不在列表中的 module_code
    excludeConditions := buildModuleExcludeConditions(tenantApps)
    if len(excludeConditions) > 0 {
        for _, cond := range excludeConditions {
            query = query.Where("NOT (app_code = ? AND module_code NOT IN ?)", cond.AppCode, cond.AllowedModules)
        }
    }

    var resources []model.Resource
    query.Order("sort_order ASC, created_at ASC").Find(&resources)

    return buildTree(resources), nil
}
```

### 6. SUPER_ADMIN 保护规则

```go
// DeleteUser 中增加保护
func (s *UserService) DeleteUser(id int64) error {
    // 检查该用户是否是最后一个 SUPER_ADMIN
    if s.isLastSuperAdmin(id) {
        return errors.NewAuthError(errors.ErrProtectedEntity, "无法删除最后一个超级管理员")
    }
    // ... 正常删除
}

// UpdateUserRoles 中增加保护
func (s *UserRoleService) UpdateUserRoles(userID, tenantID int64, roleIDs []int64) error {
    // 如果是 SUPER_ADMIN 用户且新角色列表不包含 SUPER_ADMIN 角色 → 拒绝自我降级
    if s.isSelfDegradingSuperAdmin(userID, tenantID, roleIDs) {
        return errors.NewAuthError(errors.ErrProtectedEntity, "超级管理员不能移除自身的超级管理员角色")
    }
    // ... 正常更新
}

// DeleteRole 中增加保护
func (s *RoleService) DeleteRole(id int64) error {
    role := s.getRole(id)
    if role.RoleType == "SUPER_ADMIN" && s.isLastSuperAdminRole(role.TenantID) {
        return errors.NewAuthError(errors.ErrProtectedEntity, "无法删除最后一个超级管理员角色")
    }
    // ... 正常删除
}

// isLastSuperAdmin 检查
func (s *UserService) isLastSuperAdmin(userID int64) bool {
    // 查询该用户是否有 SUPER_ADMIN 角色
    var count int64
    s.db.Table("admin_user_role ur").
        Joins("JOIN admin_role r ON r.id = ur.role_id").
        Where("ur.user_id = ? AND r.role_type = 'SUPER_ADMIN'", userID).
        Count(&count)
    if count == 0 {
        return false // 该用户不是 SUPER_ADMIN，不需要保护
    }

    // 检查是否还有其他 SUPER_ADMIN 用户
    var otherCount int64
    s.db.Table("admin_user_role ur").
        Joins("JOIN admin_role r ON r.id = ur.role_id").
        Where("ur.user_id != ? AND r.role_type = 'SUPER_ADMIN'", userID).
        Count(&otherCount)
    return otherCount == 0
}
```

### 7. Seed 数据更新

```go
func Seed(db *gorm.DB, cfg *config.Config) error {
    // 1. 创建超级管理员
    admin := &model.User{Username: "admin", Password: hash("admin123"), Status: 1}
    db.Create(admin)

    // 2. 创建默认租户
    tenant := &model.Tenant{
        TenantCode: "default", Name: "默认租户", Status: 1,
        Timezone: "Asia/Shanghai", Locale: "zh-CN", Currency: "CNY",
    }
    db.Create(tenant)

    // 3. 关联 admin → default 租户
    db.Create(&model.UserTenant{UserID: admin.ID, TenantID: tenant.ID})

    // 4. 创建 SUPER_ADMIN 角色
    role := &model.Role{
        TenantID: tenant.ID, RoleCode: "SUPER_ADMIN",
        RoleName: "超级管理员", RoleType: "SUPER_ADMIN", Status: 1,
    }
    db.Create(role)

    // 5. 分配角色
    db.Create(&model.UserRole{UserID: admin.ID, RoleID: role.ID, TenantID: tenant.ID})

    // 6. 创建内置应用 platform_admin
    app := &model.Application{
        AppCode: "platform_admin", Name: "平台管理", AppType: "BUILTIN",
        RoutePrefix: "/api/v1/admin",
        Platforms: []byte(`["admin:pc","admin:h5"]`),
        Status: 1,
    }
    db.Create(app)

    // 7. 默认租户订阅 platform_admin
    db.Create(&model.TenantApp{TenantID: tenant.ID, AppCode: "platform_admin"})

    // 8. 注册管理端菜单（全部 platform=admin, app_code=platform_admin）
    seedMenus(db)

    return nil
}

func seedMenus(db *gorm.DB) {
    menus := []model.Resource{
        {Name: "首页", Path: "/home", Type: "MENU", AppCode: "platform_admin", Platform: "admin", SortOrder: 0},
        {Name: "系统管理", Path: "/system", Type: "MENU", AppCode: "platform_admin", Platform: "admin", ModuleCode: "system", SortOrder: 10},
        // ... 子菜单省略，均含 AppCode + Platform + ModuleCode
    }
    for _, m := range menus {
        db.Create(&m)
    }
}
```

### 8. AutoDiscover 适配

```go
func AutoDiscover(engine *gin.Engine, db *gorm.DB, asyncThreshold int) {
    routes := engine.Routes()

    for _, route := range routes {
        // 只扫描 /api/v1/ 前缀的路由
        if !strings.HasPrefix(route.Path, "/api/v1/") {
            continue
        }

        // 确定 app_code：通过 route_prefix 匹配
        appCode := resolveAppCodeFromPath(route.Path) // "platform_admin" 或其他

        // 确定 module_code：从路径第四段推断
        moduleCode := inferModuleCode(route.Path, appCode)

        // 创建或更新 admin_api_permission (ENDPOINT)
        perm := &model.ApiPermission{
            Type:        "ENDPOINT",
            Name:        route.Handler,
            URLPattern:  route.Path,
            HTTPMethod:  route.Method,
            AppCode:     appCode,
            ModuleCode:  moduleCode,
            Visible:     1,
            AuthRequired: 1,
        }
        // 幂等写入（按 method + url_pattern 去重）
        upsertApiPermission(db, perm)
    }
}
```

### 9. 前端路径迁移

**axios 基础配置**（无需改，前缀通过各 API 文件控制）：
```typescript
// src/utils/request.ts — baseURL 保持不变
const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 10000,
})
```

**API 文件批量替换规则**：
```
/api/v1/tenants       → /api/v1/admin/tenants
/api/v1/users         → /api/v1/admin/users
/api/v1/roles         → /api/v1/admin/roles
/api/v1/resources     → /api/v1/admin/resources
/api/v1/api-permissions → /api/v1/admin/api-permissions
/api/v1/applications  → /api/v1/admin/applications
/api/v1/data-scope-configs → /api/v1/admin/data-scope-configs
/api/v1/operation-logs → /api/v1/admin/operation-logs
/api/v1/login-logs    → /api/v1/admin/login-logs
/api/v1/configs       → /api/v1/admin/configs

特殊处理:
/api/v1/resources/user-menu → /api/v1/common/user-menu?platform=admin
```

---

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 登录流程正常（用户名密码 → token） | POST /auth/login 正常返回 |
| RG-2 | 租户选择/切换正常 | POST /auth/tenant/select 签发 access_token |
| RG-3 | 角色 CRUD 正常 | 新路径 /api/v1/admin/roles CRUD 无报错 |
| RG-4 | 用户 CRUD + 角色分配正常 | 创建用户+关联租户+分配角色完整 |
| RG-5 | 菜单权限分配和获取正常 | 角色分配菜单后 GetUserMenu 返回正确 |
| RG-6 | API 权限分配和动态检查正常 | 有权限可访问，无权限 403 |
| RG-7 | SUPER_ADMIN 可访问所有接口 | 不被权限中间件拦截 |
| RG-8 | 操作日志正常记录 | 关键操作后有日志 |
