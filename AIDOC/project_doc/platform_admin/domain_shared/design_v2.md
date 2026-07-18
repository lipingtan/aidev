# Platform Admin 架构设计 V2

## 1. 概述

### 1.1 平台定位

Platform Admin V2 是一套通用多租户 SaaS 管理框架，采用「一切扩展皆应用」的核心理念。平台本身只提供认证、权限、租户、配置等基础能力，所有业务功能通过「应用」形态接入——无论是内置模块、运行时插件还是外部系统集成。

### 1.2 部署架构

```
┌──────────────────────────────────────────────────┐
│              Go Backend（1 个服务）                │
│   统一 API 层 · 认证 · 权限 · 多租户 · 插件宿主    │
└────────────────────┬─────────────────────────────┘
                     │ HTTP API
       ┌─────────────┼──────────────┐
       │             │              │
┌──────┴──────┐ ┌────┴────────┐ ┌───┴──────────────┐
│dev-web-admin│ │dev-web-user │ │Plugin Frontends  │
│ 管理端       │ │ 用户端/C端  │ │ 各插件前端 bundle │
│ PC + H5     │ │ PC + H5     │ │ admin/user × PC/H5│
└─────────────┘ └─────────────┘ └──────────────────┘
```

### 1.3 技术栈

| 层 | 技术选型 |
|---|---|
| 后端语言 | Go 1.21+ |
| Web 框架 | Gin |
| ORM | GORM |
| 认证 | JWT（两阶段）+ 可插拔认证策略 |
| ID 生成 | 雪花算法（Snowflake） |
| 缓存 | 本地 / Redis / 两级缓存 |
| 数据库 | MySQL 8.0 |
| 插件通信 | hashicorp/go-plugin (gRPC) |
| 管理端前端 | Vue 3 + TypeScript + Element Plus |
| 用户端前端 | Vue 3 + TypeScript + Vant/Element Plus |
| 构建 | Vite |

### 1.4 模块目录结构

```
backend/common/auth/
├── model/              # GORM 模型
├── repository/         # 数据访问层
├── service/            # 业务逻辑层
├── handler/            # Gin HTTP Handler
├── middleware/         # 中间件链
│   ├── auth.go         # JWT 认证
│   ├── app_resolve.go  # 应用识别 + 租户订阅校验
│   ├── permission.go   # 动态权限检查
│   └── data_scope.go   # 数据权限注入
├── cache/              # 缓存实现
├── spi/                # SPI 接口定义
├── engine/             # 权限表达式引擎
├── discovery/          # API 自动发现
├── strategy/           # 认证策略
├── config/             # 配置结构体
├── errors/             # 异常体系
└── auth.go             # 模块入口

backend/common/plugin/
├── manager.go          # 插件生命周期管理
├── syncer.go           # PluginResourceSyncer（权限同步）
├── proxy.go            # HTTP 请求代理
├── installer.go        # 安装/卸载
├── registry.go         # 路由/事件注册表
├── event_bus.go        # 事件总线
└── types.go            # 类型定义
```

---

## 2. 核心概念模型

### 2.1 实体关系图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         全局层（无 tenant_id）                            │
│                                                                         │
│  ┌───────────────┐    1:N     ┌──────────────────┐                      │
│  │admin_application│◀──────────│admin_resource    │                      │
│  │(应用，含类型)   │           │(菜单/按钮，含     │                      │
│  │app_type:       │    1:N    │ platform+app_code)│                      │
│  │BUILTIN/PLUGIN/ │◀──────────│                  │                      │
│  │EXTERNAL        │           └──────────────────┘                      │
│  │                │    1:N     ┌──────────────────┐                      │
│  │                │◀──────────│admin_api_permission│                     │
│  └───────┬────────┘           │(接口权限树，含     │                      │
│          │                    │ app_code)         │                      │
│          │                    └──────────────────┘                      │
│  ┌───────┴────────┐                                                     │
│  │  sys_plugin    │  1:1 对应 PLUGIN 类型的 Application                  │
│  │(插件运行时管理) │                                                     │
│  └────────────────┘                                                     │
│                                                                         │
│  ┌────────────────┐                                                     │
│  │  admin_user    │  全局用户，无 tenant_id                              │
│  └────────┬───────┘                                                     │
│           │ M:N                                                          │
│  ┌────────┴───────┐                                                     │
│  │admin_user_tenant│ 用户-租户关联                                        │
│  └────────┬───────┘                                                     │
└───────────┼─────────────────────────────────────────────────────────────┘
            │
┌───────────┼─────────────────────────────────────────────────────────────┐
│           │         租户隔离层（含 tenant_id）                            │
│           ▼                                                              │
│  ┌────────────────┐         ┌──────────────────┐                        │
│  │  admin_tenant  │──M:N───▶│admin_tenant_app  │ 租户订阅应用            │
│  │                │         │(含 enabled_modules)│                       │
│  └────────┬───────┘         └──────────────────┘                        │
│           │ 1:N                                                          │
│  ┌────────┴───────┐                                                     │
│  │  admin_role    │  角色（tenant_id 隔离，parent_id 权限上界）           │
│  └────────┬───────┘                                                     │
│           │                                                              │
│     ┌─────┼──────────────┬─────────────────┐                            │
│     │ M:N │              │ M:N             │ 1:N                         │
│ ┌───┴─────────┐  ┌──────┴────────┐  ┌─────┴──────────┐                 │
│ │admin_role_  │  │admin_role_    │  │admin_data_scope│                  │
│ │resource     │  │api            │  │(数据权限绑定,   │                  │
│ │             │  │               │  │ 含 scope_type) │                  │
│ └─────────────┘  └───────────────┘  └────────────────┘                  │
│                                                                         │
│  ┌────────────────┐                                                     │
│  │admin_user_role │  用户-角色关联（含 tenant_id）                        │
│  └────────────────┘                                                     │
│                                                                         │
│  ┌────────────────┐                                                     │
│  │ admin_config   │  三级配置（SYSTEM/TENANT/USER scope）                │
│  └────────────────┘                                                     │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 关系总结

| 关系 | 说明 |
|------|------|
| 管理用户 ↔ 租户 | M:N，通过 admin_user_tenant |
| C端用户 → 租户 | N:1，biz_user.tenant_id 直接绑定 |
| 管理用户 → 角色 | M:N，通过 admin_user_role（含 tenant_id + 时间窗口） |
| 角色 → 角色 | parent_id 权限上界约束（子 ⊆ 父） |
| 角色 → 菜单 | M:N，通过 admin_role_resource |
| 角色 → API | M:N，通过 admin_role_api（仅 ENDPOINT 叶子） |
| 角色 → 数据权限 | 1:N，admin_data_scope（含 scope_type） |
| 租户 → 应用 | M:N，通过 admin_tenant_app（含 enabled_modules） |
| 应用 → 菜单 | 1:N，admin_resource.app_code |
| 应用 → API | 1:N，admin_api_permission.app_code |
| 应用 → 插件 | 1:1，PLUGIN 类型的 Application 对应 sys_plugin 记录 |

---

## 3. 认证体系

### 3.1 认证策略架构

```go
// AuthenticationStrategy 认证策略接口
type AuthenticationStrategy interface {
    // Type 返回策略类型标识
    Type() string
    // Authenticate 执行认证，返回认证结果
    Authenticate(ctx context.Context, credentials map[string]string) (*AuthResult, error)
    // Supports 判断是否支持给定的凭据类型
    Supports(credentials map[string]string) bool
}

// AuthResult 认证结果
type AuthResult struct {
    UserID   int64
    Username string
    Tenants  []TenantBrief // 可用租户列表
}
```

内置策略：
- `PasswordStrategy` — 用户名+密码（现有流程）
- `OAuth2Strategy` — OAuth2 授权码（预留）
- `LDAPStrategy` — LDAP/AD 认证（预留）

### 3.2 两阶段 JWT

| 阶段 | Token | 有效期 | 用途 |
|------|-------|--------|------|
| 第一阶段 | platform_token | 7 天 | 标识登录身份，含可用租户列表快照 |
| 第二阶段 | access_token | 2 小时 | 租户上下文确认后签发，含 userId + tenantId + roles + platform |

### 3.3 登录流程

```
1. POST /auth/login {grant_type: "password", username, password}
   → StrategyRouter 选择 PasswordStrategy
   → 验证凭据 → 查 admin_user_tenant 获取可用租户列表
   → 单租户: 直接签发 access_token
   → 多租户: 签发 platform_token + 返回租户列表

2. POST /auth/login {grant_type: "oauth2", code, redirect_uri}
   → StrategyRouter 选择 OAuth2Strategy
   → 用 code 换 token → 匹配/创建用户 → 同上流程

3. POST /auth/tenant/select {tenant_id, platform: "admin"|"user"}
   Header: Authorization: Bearer {platform_token}
   → 验证 platform_token
   → 验证用户有该租户访问权
   → SUPER_ADMIN 可选任意租户
   → 签发 access_token（claims 含 platform）
   → 返回 {access_token, permissions, menus}
```

### 3.4 Token Claims 结构

```go
type AccessTokenClaims struct {
    jwt.RegisteredClaims
    UserID   int64   `json:"uid"`
    TenantID int64   `json:"tid"`
    Roles    []int64 `json:"roles"`
    Platform string  `json:"platform"` // "admin" 或 "user"
    UserPool string  `json:"pool"`     // "admin" 或 "user"（标识用户池）
}
```

### 3.5 Token 黑名单

- 独立 `TokenBlacklistStore` SPI
- 条目在 Token 过期后自动清理
- 支持 local / redis 两种存储

### 3.6 双用户池架构

平台区分两类用户，使用独立的表和认证流程：

| 维度 | 管理端用户（admin_user） | C端用户（biz_user） |
|------|------------------------|-------------------|
| 表 | admin_user | biz_user |
| 典型角色 | 平台管理员、租户管理员、运营 | 终端消费者、商家员工 |
| 量级 | 百~千级 | 万~百万级 |
| 认证方式 | 账号密码为主 | 手机验证码/微信/OAuth2 为主 |
| 跨租户 | 支持（M:N via admin_user_tenant） | 不支持（单租户绑定 biz_user.tenant_id） |
| 权限模型 | RBAC（角色+菜单+API+数据权限） | 简化权限（角色标签/会员等级） |
| 前端 | dev-web-admin | dev-web-user |

**认证流程区分**：
- `/auth/login` → grant_type=password/oauth2 → 查 admin_user → 签发含 `pool:"admin"` 的 token
- `/auth/user/login` → grant_type=sms/wechat → 查 biz_user → 签发含 `pool:"user"` 的 token

**中间件层用 UserPool 路由**（Token Claims 结构见 3.4）：
- `pool=admin` → 走完整 RBAC 权限检查链
- `pool=user` → 走简化权限检查（基于角色标签/会员等级，由业务应用自行扩展）

---

## 4. 权限模型

### 4.1 中间件链

```
Request
  │
  ├── AuthMiddleware         → 解析 JWT，注入 AuthContext (userId, tenantId, roles, platform)
  │
  ├── AppResolveMiddleware   → URL 前缀 → app_code；校验租户订阅
  │
  ├── PermissionMiddleware   → permission_code 匹配（限定 app_code 范围）
  │
  ├── DataScopeMiddleware    → GORM Callback 注入数据过滤条件
  │
  └── Handler
```

### 4.2 AppResolveMiddleware（新增）

```go
// 启动时从 admin_application 加载 route_prefix → app_code 映射
// 请求时通过最长前缀匹配确定 app_code
func AppResolveMiddleware(prefixMap *AppPrefixMap, publicPaths map[string]bool) gin.HandlerFunc {
    return func(c *gin.Context) {
        path := c.Request.URL.Path
        
        // 公共路径白名单（如 /auth/*, /api/v1/common/*）直接放行
        if publicPaths[c.FullPath()] {
            c.Next()
            return
        }
        
        appCode := prefixMap.Match(path)
        if appCode == "" {
            // 非白名单且无法识别应用 → 拒绝（防止遗漏注册的接口被静默放行）
            c.AbortWithStatusJSON(403, gin.H{"code": 40303, "message": "无法识别请求所属应用"})
            return
        }
        c.Set("app_code", appCode)

        // 校验租户订阅
        authCtx := GetAuthContext(c)
        if authCtx == nil {
            c.Next() // AuthMiddleware 会处理
            return
        }
        if !isSubscribed(authCtx.TenantID, appCode) {
            c.AbortWithStatusJSON(403, gin.H{"code": 40302, "message": "租户未开通此应用"})
            return
        }
        c.Next()
    }
}
```

**公共路径白名单**（不走应用级权限，仅走 AuthMiddleware）：
- `/auth/*` — 认证相关
- `/api/v1/common/user-menu` — 用户菜单（管理端/用户端共用）
- `/api/v1/common/user-info` — 当前用户信息

### 4.3 菜单权限（归属应用 + 平台）

`admin_resource` 字段设计：
- `app_code` — 归属应用
- `platform` — 归属前端平台（admin/user）
- `type` — MENU / BUTTON / PAGE
- 树形结构：`parent_id` 邻接表

**GetUserMenu 逻辑**：
1. 查用户在当前租户的角色 → roleIDs
2. SUPER_ADMIN → 返回对应 platform 的所有菜单
3. 普通角色：
   - 查 admin_tenant_app WHERE tenant_id → 租户已订阅的 appCodes
   - 查 admin_role_resource WHERE role_id IN roleIDs → resourceIDs
   - 查 admin_resource WHERE id IN resourceIDs AND app_code IN appCodes AND platform = ?
   - 构建树返回

### 4.4 接口权限（归属应用）

`admin_api_permission` 关键字段：
- `app_code` — 归属应用
- `permission_code` — 权限码，格式 `{module}:{entity}:{action}`
- `url_pattern` — 路由模板（如 `/api/v1/billing/invoices/:id`）
- `http_method` — HTTP 方法
- `auth_required` — 是否需要权限校验（0=免检）
- `visible` — 配置界面可见性

**DynamicPermissionMiddleware 逻辑（V2 优化）**：
1. 白名单检查 → 放行
2. SUPER_ADMIN → 放行
3. 从缓存 codeMap 获取 `method:path → permission_code`
4. 从缓存获取用户权限集（限定当前 app_code）
5. PermissionEngine 匹配 → 放行或 403

### 4.5 角色继承 — 权限上界模型

```
角色 A (顶级，parent_id = NULL)
  权限范围 = 租户订阅应用的全部权限
  │
  ├── 角色 B (parent_id = A)
  │     可配置权限 ⊆ 角色 A 的已分配权限
  │     │
  │     └── 角色 C (parent_id = B)
  │           可配置权限 ⊆ 角色 B 的已分配权限
  │
  └── 角色 D (parent_id = A)
        可配置权限 ⊆ 角色 A 的已分配权限
```

**校验逻辑（写时校验）**：

```go
func (s *RoleService) AssignResources(roleID int64, resourceIDs []int64) error {
    role := s.getRole(roleID)
    
    if role.ParentID != nil {
        // 查父角色已分配的资源集
        parentResourceIDs := s.getRoleResourceIDs(*role.ParentID)
        // 校验子集关系
        if !isSubset(resourceIDs, parentResourceIDs) {
            return ErrExceedsParentPermission
        }
    }
    
    // 全量替换
    s.replaceRoleResources(roleID, resourceIDs)
    
    // 级联裁剪子角色超出部分
    s.cascadeTrimChildren(roleID, resourceIDs)
    return nil
}

func (s *RoleService) cascadeTrimChildren(parentRoleID int64, parentResourceIDs []int64) {
    children := s.getChildRoles(parentRoleID)
    for _, child := range children {
        childResIDs := s.getRoleResourceIDs(child.ID)
        trimmed := intersect(childResIDs, parentResourceIDs)
        if len(trimmed) != len(childResIDs) {
            s.replaceRoleResources(child.ID, trimmed)
            // 递归处理孙角色
            s.cascadeTrimChildren(child.ID, trimmed)
        }
    }
}
```

### 4.6 数据权限（维度注册 + scope_type）

scope_type 枚举：

| scope_type | 行为 | WHERE 条件 |
|-----------|------|-----------|
| ALL | 不过滤 | 无 |
| SELF | 仅本人创建 | `create_by = ?` (currentUserID) |
| DEPT | 本部门 | `dept_id IN ?` (用户所属部门 ID) |
| DEPT_TREE | 本部门及下级 | `dept_id IN ?` (部门 + 所有子部门 ID) |
| CUSTOM | 自定义值列表 | `{column} IN ?` (dimension_values) |

**GORM Callback 实现**：

```go
func DataScopeCallback(db *gorm.DB) {
    // 跳过系统操作
    if isSystemOp(db.Statement.Context) { return }
    
    scopes := getUserDataScopes(db.Statement.Context)
    for _, scope := range scopes {
        switch scope.ScopeType {
        case "ALL":
            continue
        case "SELF":
            db.Where("create_by = ?", getCurrentUserID(db.Statement.Context))
        case "DEPT":
            deptIDs := orgProvider.GetOrgIds(userID, tenantID)
            db.Where(scope.TableColumn+" IN ?", deptIDs)
        case "DEPT_TREE":
            deptIDs := orgProvider.GetSubOrgIds(userID, tenantID)
            db.Where(scope.TableColumn+" IN ?", deptIDs)
        case "CUSTOM":
            db.Where(scope.TableColumn+" IN ?", scope.DimensionValues)
        }
    }
}
```

### 4.7 权限表达式引擎

```go
// 通配符规则:
//   *  = 匹配恰好一层（不含冒号）
//   ** = 匹配一层或多层

// 示例:
//   "billing:*:list"  匹配 "billing:invoice:list"
//   "billing:**"      匹配 "billing:invoice:list"、"billing:payment:create"
```

### 4.8 字段级权限（Field-Level Security）

不同角色看到同一页面/接口但不同字段（如销售看不到成本价、客服看不到合同金额）。

**数据模型**：
```sql
CREATE TABLE admin_field_permission (
    id          BIGINT PRIMARY KEY,
    role_id     BIGINT NOT NULL,
    object_code VARCHAR(64) NOT NULL COMMENT '业务对象标识（如 invoice、customer）',
    field_name  VARCHAR(64) NOT NULL COMMENT '字段名（对应 JSON key）',
    access      VARCHAR(16) NOT NULL DEFAULT 'VISIBLE' COMMENT 'VISIBLE/EDITABLE/HIDDEN',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_object_field (role_id, object_code, field_name)
);
```

**access 枚举**：

| 值 | 含义 | API 响应行为 | 前端行为 |
|---|------|------------|---------|
| VISIBLE | 可见只读 | 正常返回该字段 | 显示但 disabled |
| EDITABLE | 可见可编辑 | 正常返回该字段 | 正常可编辑 |
| HIDDEN | 隐藏 | 响应 JSON 中移除该字段 | 不渲染 |

**实现机制**：
```go
// FieldFilter 在 Handler 返回响应前过滤字段
type FieldFilter struct {
    fieldPermRepo FieldPermissionRepository
}

func (f *FieldFilter) Filter(ctx context.Context, objectCode string, data interface{}) interface{} {
    roleIDs := GetAuthContext(ctx).Roles
    // 查询该角色对该对象的字段权限配置
    perms := f.fieldPermRepo.GetByRolesAndObject(roleIDs, objectCode)
    // 未配置的字段默认 EDITABLE（不限制）
    // 配置为 HIDDEN 的字段从响应中移除
    return removeHiddenFields(data, perms)
}
```

**默认策略**：未配置字段权限的对象/字段 → 默认 EDITABLE（不限制）。仅对显式配置了规则的对象生效。

### 4.9 记录级共享规则（Record Sharing）

在数据权限（scope_type）之外，支持将特定记录额外共享给指定用户/角色/部门。

**场景**：跨部门协作项目、共享客户、临时授权查看某条记录。

**数据模型**：
```sql
CREATE TABLE admin_record_share (
    id            BIGINT PRIMARY KEY,
    tenant_id     BIGINT NOT NULL,
    object_code   VARCHAR(64) NOT NULL COMMENT '业务对象',
    record_id     BIGINT NOT NULL COMMENT '被共享的记录 ID',
    share_to_type VARCHAR(16) NOT NULL COMMENT '共享目标类型: USER/ROLE/DEPT',
    share_to_id   BIGINT NOT NULL COMMENT '共享目标 ID',
    access_level  VARCHAR(16) NOT NULL DEFAULT 'READ' COMMENT 'READ/EDIT',
    expire_at     DATETIME NULL COMMENT '共享过期时间（NULL=永久）',
    created_by    BIGINT NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_record (tenant_id, object_code, record_id),
    INDEX idx_share_to (share_to_type, share_to_id, tenant_id)
);
```

**查询时 WHERE 条件扩展**：
```sql
-- 原有数据权限条件 OR 共享规则命中
WHERE (原有 scope_type 生成的条件)
   OR record_id IN (
       SELECT record_id FROM admin_record_share
       WHERE tenant_id = ? AND object_code = ?
         AND (
           (share_to_type = 'USER' AND share_to_id = ?) OR
           (share_to_type = 'ROLE' AND share_to_id IN (?...)) OR
           (share_to_type = 'DEPT' AND share_to_id IN (?...))
         )
         AND (expire_at IS NULL OR expire_at > NOW())
   )
```

**注意**：记录共享是数据权限的**补充**（OR 关系），不会缩小原有范围。

### 4.10 权限集（Permission Set）

解决"90% 权限一样但个别人多一个功能"的常见场景，无需为每种组合创建新角色。

**设计**：
- `admin_role.role_type` 增加 `PERMISSION_SET` 类型
- Permission Set 不受 parent_id 权限上界约束（它是叠加增量）
- 一个用户可关联多个 Permission Set
- 用户最终权限 = 角色权限 ∪ 所有 Permission Set 权限

**角色类型完整枚举**：

| role_type | 用途 | 权限上界约束 |
|-----------|------|------------|
| SUPER_ADMIN | 平台超管 | 无（全部权限） |
| TENANT_ADMIN | 租户管理员 | 受父角色约束 |
| NORMAL | 普通角色 | 受父角色约束 |
| PERMISSION_SET | 权限集（叠加） | **不受**父角色约束 |

**权限计算伪代码**：
```go
func getUserPermissions(userID, tenantID int64) []string {
    roles := getUserRoles(userID, tenantID)
    
    var allPerms []string
    for _, role := range roles {
        if role.RoleType == "PERMISSION_SET" {
            // 权限集：直接叠加
            allPerms = append(allPerms, getRolePermCodes(role.ID)...)
        } else {
            // 普通角色：取角色自身权限
            allPerms = append(allPerms, getRolePermCodes(role.ID)...)
        }
    }
    return deduplicate(allPerms)
}
```

**前端交互**：角色管理页增加「权限集」Tab，创建/分配方式与普通角色一致，但 UI 标注为"叠加权限"。

---

## 5. 多租户架构

### 5.1 设计原则

| 原则 | 实现 |
|------|------|
| 全局用户 | admin_user 无 tenant_id，一人可管理多租户 |
| C端用户隔离 | biz_user 含 tenant_id，单租户绑定 |
| M:N 关联 | admin_user_tenant 记录管理端用户-租户关系 |
| 租户订阅应用 | admin_tenant_app 控制租户可用哪些应用及功能模块 |
| tenant_id 从 token 取 | 禁止使用请求参数中的 tenant_id |

### 5.2 管理员分级与操作边界

| 级别 | role_type | 操作范围 | 限制 |
|------|-----------|---------|------|
| 平台超管 | SUPER_ADMIN | 管理所有租户、应用、插件、全局配置、平台用户 | 不能删除最后一个 SUPER_ADMIN |
| 租户管理员 | TENANT_ADMIN | 管理本租户的用户、角色（受权限上界约束）、租户级配置 | 不能操作其他租户、不能修改应用订阅（由平台管控） |
| 普通管理角色 | NORMAL | 按角色已分配权限使用管理端功能 | 完全受 RBAC 控制 |

**SUPER_ADMIN 保护规则**：
- 禁止删除最后一个 SUPER_ADMIN 用户（Service 层校验）
- 禁止 SUPER_ADMIN 降级自身角色（防止失去管理权）
- 敏感操作（删除租户、卸载插件、修改认证策略）记入操作日志并标记为「高风险」

**TENANT_ADMIN 权限边界**：
- 创建/管理本租户的普通角色（子角色，权限 ⊆ TENANT_ADMIN 自身权限）
- 管理本租户关联的用户（增删改 admin_user_tenant、分配角色）
- 修改租户级配置（admin_config scope=TENANT）
- **不可**：修改应用订阅、管理插件、操作其他租户数据

### 5.3 租户生命周期

```
创建(ACTIVE) → 正常运营 → 欠费降级(READ_ONLY) → 续费恢复(ACTIVE)
                         → 手动禁用(DISABLED) → 重新启用(ACTIVE)
                         → 注销申请(CANCELLING) → 数据保留期(30天) → 清理(DELETED)
```

| status | 值 | 含义 | 读操作 | 写操作 |
|--------|---|------|--------|--------|
| ACTIVE | 1 | 正常 | ✅ | ✅ |
| DISABLED | 0 | 禁用（管理员手动） | ❌ | ❌ |
| READ_ONLY | 2 | 只读降级（欠费/试用到期） | ✅ | ❌ |
| CANCELLING | 3 | 注销中（数据保留期） | ❌ | ❌ |

**AuthMiddleware 行为**：
- status=ACTIVE → 放行
- status=READ_ONLY → 仅允许 GET/HEAD 方法，其余返回 403 "租户已降为只读模式"
- status=DISABLED/CANCELLING → 返回 403 "租户已禁用"

### 5.4 租户配置 — 三级配置链

配置优先级：`USER > TENANT > SYSTEM`

查询逻辑：
```sql
-- 优先取用户级，其次租户级，最后系统默认
-- USER scope: scope_id=userId, tenant_id=tenantId（同用户不同租户偏好独立）
-- TENANT scope: scope_id=tenantId, tenant_id=tenantId
-- SYSTEM scope: scope_id=0, tenant_id=0
SELECT config_value FROM admin_config 
WHERE config_key = ? 
  AND (
    (scope = 'USER' AND scope_id = ? AND tenant_id = ?) OR
    (scope = 'TENANT' AND scope_id = ? AND tenant_id = ?) OR
    (scope = 'SYSTEM' AND scope_id = 0 AND tenant_id = 0)
  )
ORDER BY FIELD(scope, 'USER', 'TENANT', 'SYSTEM')
LIMIT 1;
```

### 5.5 租户功能开关

`admin_tenant_app.enabled_modules` 控制租户对某应用的模块级启用：

```json
// 租户 A 订阅了 billing 应用，但只启用了发票和支付模块
{
  "tenant_id": 1001,
  "app_code": "billing",
  "enabled_modules": ["invoice", "payment"]  
  // billing 应用还有 "refund" 模块，但该租户未启用
}
```

角色权限配置时，仅展示租户已启用模块对应的菜单/API 节点。

**enabled_modules 语义规则**：
- `NULL` = 启用该应用当前声明的全部模块（适合"包月全功能"订阅模式）
- 显式数组如 `["invoice","payment"]` = 仅启用列出的模块
- 空数组 `[]` = 订阅了应用但未启用任何模块（占位状态）
- **新模块上线策略**：应用新增模块后，已订阅的 `NULL` 租户自动获得；显式数组租户需手动添加。如需"新模块默认关闭"的商业模式，则新租户订阅时应显式写入当前模块列表而非 NULL。

**运行时模块过滤**：
- 租户管理员关闭某模块 → 发布 `EventTenantAppChanged` 事件 → 清除相关缓存
- DynamicPermissionMiddleware 检查权限时，如果请求 API 所属 `module_code` 不在该租户的 `enabled_modules` 中 → 403
- 这确保即使角色已分配了某模块权限，租户关闭后也即时失效

### 5.6 C 端用户权限策略

C 端用户（biz_user）与管理端用户（admin_user）采用不同的权限模型：

**菜单可见性**：
- C 端用户无角色-菜单绑定机制
- GetUserMenu(platform=user) 对 C 端用户返回：该租户已订阅应用中 `platform=user` 的所有菜单
- 即：C 端可见范围 = 租户订阅 × 已启用模块 × platform=user
- 按钮级细粒度控制由业务应用通过 `biz_user.member_level` / `tags` 自行处理

**API 权限**：
- C 端 API（route_prefix 匹配到的应用接口）仅校验：
  1. Token 有效（pool=user）
  2. 租户已订阅该应用
  3. 该模块已启用
- **不走** RBAC 的 permission_code 匹配（C 端量级大，全量权限检查性能不可接受）
- 业务级细粒度权限（如会员才能查看某内容）由业务应用 Handler 内自行校验 `biz_user.member_level`

**中间件链（pool=user 时）**：
```
Request (pool=user)
  │
  ├── AuthMiddleware       → 解析 JWT，识别 pool=user
  │
  ├── AppResolveMiddleware → URL 前缀 → app_code；校验租户订阅 + 模块启用
  │
  ├── (跳过 PermissionMiddleware)
  │
  └── Handler（业务应用自行校验 member_level/tags）
```

### 5.7 C 端用户注册与创建

biz_user 的创建方式：

| 方式 | 场景 | 流程 |
|------|------|------|
| 管理端创建 | 租户管理员导入员工 | POST /api/v1/admin/biz-users（管理端 API） |
| 自助注册 | C端用户扫码/下载后注册 | POST /auth/user/register（公开接口） |
| 外部同步 | 微信关注公众号、企微成员同步 | 通过 AuthenticationStrategy(wechat) 自动创建 |
| 邀请注册 | 管理员发送邀请链接 | 邀请码绑定 tenant_id，注册时自动关联 |

**自助注册流程**：
```
POST /auth/user/register {phone, sms_code, tenant_code}
  → 验证短信验证码
  → 通过 tenant_code 或邀请码确定 tenant_id
  → 创建 biz_user 记录（tenant_id 绑定）
  → 自动签发 access_token (pool=user)
```

**租户归属确定规则**：
- 有邀请码 → 从邀请码解析 tenant_id
- 有 tenant_code → 从 admin_tenant 查询
- 扫商家二维码 → 二维码含 tenant_code
- 无任何标识 → 拒绝注册（C端用户必须归属租户）

---

## 6. 应用管理

### 6.1 统一应用模型

```
admin_application
├── app_type: BUILTIN    → 内置应用（编译期集成）
├── app_type: PLUGIN     → 插件应用（运行时安装）
└── app_type: EXTERNAL   → 外部应用（OAuth2/Webhook 集成）
```

### 6.2 应用与路由前缀

| app_code | app_type | route_prefix | platforms |
|----------|----------|-------------|-----------|
| platform_admin | BUILTIN | /api/v1/admin | ["admin:pc","admin:h5"] |
| billing | PLUGIN | /api/v1/billing | ["admin:pc","user:pc","user:h5"] |
| crm | PLUGIN | /api/v1/crm | ["admin:pc","admin:h5"] |
| wechat_work | EXTERNAL | — | ["admin:pc"] |

### 6.3 插件生命周期与应用的关系

```
安装插件 → 创建 sys_plugin 记录 + 创建 admin_application (app_type=PLUGIN)
    │
启动插件 → Plugin.Register() → PluginResourceSyncer 同步菜单/API 到 RBAC 体系
    │
租户订阅 → admin_tenant_app 关联 → 该租户角色可配置该应用权限
    │
停止插件 → 标记 sys_plugin.status = stopped（资源保留，仅插件进程停止）
    │
卸载插件 → 清理 admin_resource/admin_api_permission + admin_application + sys_plugin
         → 级联清理角色绑定
```

**插件状态 × 订阅状态 — 用户感知矩阵**：

| 插件状态 | 租户订阅状态 | 管理端用户感知 | C端用户感知 |
|---------|------------|-------------|-----------|
| 运行中 | 已订阅 | 功能正常可用 | 功能正常可用 |
| 已停止 | 已订阅 | 菜单灰显 + 提示"功能维护中" | 相关入口隐藏或显示维护提示 |
| 运行中 | 未订阅 | 应用目录中可见（可申请开通） | 完全不可见 |
| 已卸载 | — | 完全不可见 | 完全不可见 |

**插件停止时的技术行为**：
- 菜单资源保留在 admin_resource 中（不删除），但 API 请求由 PluginProxy 返回 503
- 前端收到 503 时展示"功能维护中"提示
- 租户订阅关系保留，重新启动后即刻恢复

### 6.4 应用发现与订阅流程

**平台侧（SUPER_ADMIN）**：
1. 插件管理页：上传插件包 → 安装 → 启动 → 应用自动出现在应用列表
2. 应用管理页：查看所有应用（含 BUILTIN/PLUGIN/EXTERNAL），管理元信息
3. 未来扩展：对接外部应用市场（Plugin Marketplace），浏览/一键安装

**租户侧（TENANT_ADMIN）**：
1. 「可用应用」页面：展示平台已上架的应用列表（状态=运行中 + 对外可见）
2. 点击「申请开通」→ 创建订阅工单（或直接订阅，取决于商业模式配置）
3. 已订阅应用管理：查看已开通应用、模块开关配置
4. 退订：取消订阅 → 级联清理权限（二次确认）

**自助订阅 vs 审批订阅**：
通过系统配置 `app.subscription_mode` 控制：
- `self_service` — 租户管理员自助订阅/退订
- `approval_required` — 需平台管理员审批后生效

### 6.5 PluginResourceSyncer

```go
// PluginResourceSyncer 负责将插件声明的资源同步到 RBAC 体系
type PluginResourceSyncer struct {
    db *gorm.DB
}

// SyncOnStart 插件启动时全量同步
func (s *PluginResourceSyncer) SyncOnStart(info *proto.PluginInfo) error {
    appCode := info.Name
    
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. 清理该 app_code 的旧资源（幂等）
        tx.Where("app_code = ?", appCode).Delete(&model.Resource{})
        tx.Where("app_code = ?", appCode).Delete(&model.ApiPermission{})
        
        // 2. 写入新菜单
        for _, menu := range info.Menus {
            resource := menuToResource(menu, appCode)
            tx.Create(resource)
        }
        
        // 3. 写入新 API 权限
        for _, api := range info.Apis {
            perm := apiToPermission(api, appCode)
            tx.Create(perm)
        }
        
        // 4. 清理失效的角色绑定（引用了已删除的 resource/api）
        s.cleanOrphanBindings(tx, appCode)
        
        return nil
    })
}

// SyncOnUninstall 插件卸载时清理
func (s *PluginResourceSyncer) SyncOnUninstall(appCode string) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 查出要删除的资源 ID
        var resIDs []int64
        tx.Model(&model.Resource{}).Where("app_code = ?", appCode).Pluck("id", &resIDs)
        var apiIDs []int64
        tx.Model(&model.ApiPermission{}).Where("app_code = ?", appCode).Pluck("id", &apiIDs)
        
        // 级联清理角色绑定
        if len(resIDs) > 0 {
            tx.Where("resource_id IN ?", resIDs).Delete(&model.RoleResource{})
        }
        if len(apiIDs) > 0 {
            tx.Where("api_permission_id IN ?", apiIDs).Delete(&model.RoleApi{})
        }
        
        // 删除资源和 API 权限
        tx.Where("app_code = ?", appCode).Delete(&model.Resource{})
        tx.Where("app_code = ?", appCode).Delete(&model.ApiPermission{})
        
        // 清理 admin_role_app
        tx.Where("app_code = ?", appCode).Delete(&model.RoleApp{})
        
        return nil
    })
}
```

### 6.5 外部应用集成（EXTERNAL）

EXTERNAL 类型应用在 `admin_application.app_config` 中存储集成配置：

```json
{
  "oauth2": {
    "client_id": "xxx",
    "client_secret": "***",
    "authorize_url": "https://open.work.weixin.qq.com/wwopen/sso/qrConnect",
    "token_url": "https://qyapi.weixin.qq.com/cgi-bin/gettoken",
    "userinfo_url": "https://qyapi.weixin.qq.com/cgi-bin/user/getuserinfo",
    "scopes": ["snsapi_base"]
  },
  "webhook": {
    "callback_url": "https://example.com/callback",
    "secret": "***"
  }
}
```

EXTERNAL 应用不注册 API 路由前缀，仅注册菜单（跳转链接）和权限码（控制可见性）。

---

## 7. 插件系统

### 7.1 插件 Manifest（V2）

```json
{
  "name": "billing",
  "version": "1.2.0",
  "displayName": "计费管理",
  "description": "发票、支付、退款管理",
  "routePrefix": "/api/v1/billing",
  "platforms": ["admin:pc", "user:pc", "user:h5"],
  "modules": [
    {"code": "invoice", "name": "发票管理"},
    {"code": "payment", "name": "支付管理"},
    {"code": "refund", "name": "退款管理"}
  ],
  "frontends": {
    "admin": {"pc": "dist/admin-pc/"},
    "user": {"pc": "dist/user-pc/", "h5": "dist/user-h5/"}
  },
  "exposedActions": [
    {
      "name": "getInvoice",
      "version": "1.0",
      "inputSchema": {"invoiceId": "int64"},
      "outputSchema": {"invoice": "InvoiceDTO"}
    }
  ],
  "subscribedEvents": ["order.created", "payment.completed"]
}
```

### 7.2 插件通信契约

```go
// ActionDescriptor 插件暴露的可调用 Action
type ActionDescriptor struct {
    Name         string `json:"name"`
    Version      string `json:"version"`
    InputSchema  string `json:"inputSchema"`  // JSON Schema
    OutputSchema string `json:"outputSchema"` // JSON Schema
}

// CallPluginRequest 增加版本字段
type CallPluginRequest struct {
    TargetPlugin  string            `json:"targetPlugin"`
    Action        string            `json:"action"`
    ActionVersion string            `json:"actionVersion"` // 新增
    Payload       map[string][]byte `json:"payload"`
}
```

**调用时校验流程**：
1. 查找目标插件是否运行中
2. 查找目标插件是否暴露该 Action
3. 比较 ActionVersion 兼容性（major 版本匹配）
4. 不兼容 → 返回 `ErrActionVersionIncompatible`

### 7.3 事件总线

- 插件通过 Manifest `subscribedEvents` 声明订阅
- 宿主 EventBus 启动时自动注册订阅关系
- 事件发布为异步广播，失败不阻塞
- 事件类型采用 `{domain}.{action}` 命名约定

---

## 8. 缓存策略（V2 优化）

### 8.1 缓存层级

```
L1 (本地内存):
  - role:{roleId}:resources → []int64        TTL: 5min
  - role:{roleId}:apis → []int64             TTL: 5min
  - app_prefix_map → map[prefix]appCode      TTL: 10min

L2 (Redis/本地):
  - user:{userId}:tenant:{tenantId}:perms → []string   TTL: 30min
    # 存该用户在当前租户下的全量 permCodes（跨应用）
    # app_code 过滤在权限码命名约定中通过前缀实现（如 billing:invoice:list）
    # 中间件匹配时用 PermissionEngine 做精确/通配匹配即可
  - tenant:{tenantId}:apps → []string                   TTL: 10min
  - tenant:{tenantId}:app:{appCode}:modules → []string  TTL: 10min  # 已启用模块
  - config:{scope}:{scopeId}:{tenantId}:{key} → string  TTL: 5min
  - api_code_map:v{version} → map[method:path]code     TTL: 永不过期（靠版本号刷新）
```

### 8.2 事件驱动失效

```go
// 权限相关变更 → 发布事件 → 缓存层订阅并失效

// 事件类型:
const (
    EventRolePermChanged   = "role.perm.changed"    // 角色权限变更
    EventUserRoleChanged   = "user.role.changed"    // 用户角色变更
    EventTenantAppChanged  = "tenant.app.changed"   // 租户订阅变更
    EventApiPermRegistered = "api.perm.registered"  // API 权限注册/变更
    EventTenantDisabled    = "tenant.disabled"      // 租户禁用
)

// 缓存失效规则:
EventRolePermChanged   → 清 L1(role:{roleId}:*) + 分批清 L2(user:*:tenant:*:perms 涉及该角色的)
EventUserRoleChanged   → 清 L2(user:{userId}:tenant:{tenantId}:perms)
EventTenantAppChanged  → 清 L2(tenant:{tenantId}:apps)
EventApiPermRegistered → bump api_code_map 版本号 → 下次请求自动加载新版
EventTenantDisabled    → 清该 tenant 下所有 L2（prefix 删除）
```

### 8.3 api_code_map 版本化

```go
// 启动时加载 api_code_map:v1
// API 权限变更时:
//   1. 加载新数据构建 map
//   2. 写入 api_code_map:v2
//   3. 原子替换 currentVersion = 2
//   4. 旧版本延迟删除（等活跃请求完成）
```

---

## 9. 数据库设计

### 9.1 完整 DDL

```sql
-- ==================== 租户 ====================
CREATE TABLE admin_tenant (
    id              BIGINT PRIMARY KEY,
    tenant_code     VARCHAR(64) NOT NULL UNIQUE,
    name            VARCHAR(128) NOT NULL,
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '1=正常 0=禁用 2=只读 3=注销中',
    config          JSON COMMENT '租户级别配置（预留）',
    timezone        VARCHAR(64) NOT NULL DEFAULT 'Asia/Shanghai' COMMENT '租户时区',
    locale          VARCHAR(16) NOT NULL DEFAULT 'zh-CN' COMMENT '默认语言',
    currency        VARCHAR(8) NOT NULL DEFAULT 'CNY' COMMENT '默认币种',
    expired_at      DATETIME NULL COMMENT '试用/订阅到期时间',
    version         INT NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- ==================== 用户（全局） ====================
CREATE TABLE admin_user (
    id              BIGINT PRIMARY KEY,
    username        VARCHAR(64) NOT NULL UNIQUE,
    password        VARCHAR(128) NOT NULL,
    email           VARCHAR(128) NULL,
    phone           VARCHAR(32) NULL,
    nickname        VARCHAR(64) NULL,
    avatar          VARCHAR(256) NULL,
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '1=启用 0=禁用',
    auth_type       VARCHAR(32) NOT NULL DEFAULT 'password' COMMENT '认证方式: password/oauth2/ldap',
    external_id     VARCHAR(256) NULL COMMENT '外部系统用户标识（OAuth2/LDAP 场景）',
    version         INT NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_external (auth_type, external_id)
);

-- ==================== 用户-租户关联 ====================
CREATE TABLE admin_user_tenant (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    tenant_id       BIGINT NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_tenant (user_id, tenant_id)
);

-- ==================== C端用户（租户隔离，独立用户池） ====================
CREATE TABLE biz_user (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL COMMENT '所属租户（单租户绑定）',
    phone           VARCHAR(32) NULL COMMENT '手机号（验证码登录）',
    email           VARCHAR(128) NULL,
    nickname        VARCHAR(64) NULL,
    avatar          VARCHAR(256) NULL,
    auth_type       VARCHAR(32) NOT NULL DEFAULT 'sms' COMMENT 'sms/wechat/oauth2',
    external_id     VARCHAR(256) NULL COMMENT '外部标识（微信 openid 等）',
    union_id        VARCHAR(256) NULL COMMENT '跨平台标识（微信 unionid 等）',
    password        VARCHAR(128) NULL COMMENT '可选密码（部分场景需要）',
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '1=正常 0=禁用 2=注销中',
    member_level    VARCHAR(32) NULL COMMENT '会员等级（业务扩展）',
    tags            JSON NULL COMMENT '用户标签（业务扩展）',
    last_login_at   DATETIME NULL,
    version         INT NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant (tenant_id),
    INDEX idx_phone (phone),
    INDEX idx_external (auth_type, external_id),
    UNIQUE KEY uk_tenant_phone (tenant_id, phone)
);

-- ==================== 应用（全局，统一模型） ====================
CREATE TABLE admin_application (
    id              BIGINT PRIMARY KEY,
    app_code        VARCHAR(64) NOT NULL UNIQUE,
    name            VARCHAR(128) NOT NULL,
    description     VARCHAR(512) NULL,
    app_type        VARCHAR(16) NOT NULL DEFAULT 'BUILTIN' COMMENT 'BUILTIN/PLUGIN/EXTERNAL',
    route_prefix    VARCHAR(128) NULL COMMENT 'API 路由前缀，如 /api/v1/billing',
    platforms       JSON NOT NULL DEFAULT '[]' COMMENT '支持的平台 ["admin:pc","user:h5"]',
    modules         JSON NULL COMMENT '应用声明的功能模块 [{"code":"invoice","name":"发票管理"}]',
    app_config      JSON NULL COMMENT '应用级配置（EXTERNAL 存 OAuth2 配置等）',
    icon            VARCHAR(256) NULL,
    sort_order      INT NOT NULL DEFAULT 0,
    status          TINYINT NOT NULL DEFAULT 1,
    version         INT NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- ==================== 租户-应用订阅 ====================
CREATE TABLE admin_tenant_app (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    app_code        VARCHAR(64) NOT NULL,
    enabled_modules JSON NULL COMMENT '启用的模块列表 ["invoice","payment"]，NULL=全部启用',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tenant_app (tenant_id, app_code)
);

-- ==================== 角色（租户隔离，权限上界继承） ====================
CREATE TABLE admin_role (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL DEFAULT 0 COMMENT '0=全局角色',
    role_code       VARCHAR(64) NOT NULL,
    role_name       VARCHAR(128) NOT NULL,
    role_type       VARCHAR(32) NOT NULL DEFAULT 'NORMAL' COMMENT 'NORMAL/SUPER_ADMIN/TENANT_ADMIN/PERMISSION_SET',
    parent_id       BIGINT NULL COMMENT '父角色 ID（权限上界约束）',
    description     VARCHAR(512) NULL,
    sort_order      INT NOT NULL DEFAULT 0,
    status          TINYINT NOT NULL DEFAULT 1,
    version         INT NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tenant_role_code (tenant_id, role_code)
);

-- ==================== 用户-角色关联 ====================
CREATE TABLE admin_user_role (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    role_id         BIGINT NOT NULL,
    tenant_id       BIGINT NOT NULL,
    effective_start DATETIME NULL COMMENT '临时授权开始',
    effective_end   DATETIME NULL COMMENT '临时授权结束',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_role_tenant (user_id, role_id, tenant_id)
);

-- ==================== 菜单/按钮资源树（全局，按应用+平台） ====================
CREATE TABLE admin_resource (
    id              BIGINT PRIMARY KEY,
    parent_id       BIGINT NULL,
    type            VARCHAR(16) NOT NULL COMMENT 'MENU/BUTTON/PAGE',
    name            VARCHAR(128) NOT NULL,
    permission_code VARCHAR(128) NULL,
    path            VARCHAR(256) NULL,
    component       VARCHAR(256) NULL,
    icon            VARCHAR(128) NULL,
    app_code        VARCHAR(64) NOT NULL COMMENT '所属应用编码',
    platform        VARCHAR(16) NOT NULL DEFAULT 'admin' COMMENT '归属前端平台: admin/user',
    module_code     VARCHAR(64) NULL COMMENT '所属功能模块（对应应用的 modules.code）',
    sort_order      INT NOT NULL DEFAULT 0,
    status          TINYINT NOT NULL DEFAULT 1,
    version         INT NOT NULL DEFAULT 1,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_platform (app_code, platform)
);

-- ==================== 角色-资源关联 ====================
CREATE TABLE admin_role_resource (
    id              BIGINT PRIMARY KEY,
    role_id         BIGINT NOT NULL,
    resource_id     BIGINT NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_resource (role_id, resource_id)
);

-- ==================== 接口权限树（全局，按应用） ====================
CREATE TABLE admin_api_permission (
    id              BIGINT PRIMARY KEY,
    parent_id       BIGINT NULL,
    type            VARCHAR(16) NOT NULL COMMENT 'GROUP/ENDPOINT',
    name            VARCHAR(128) NOT NULL,
    display_name    VARCHAR(128) NULL COMMENT '中文显示名称',
    permission_code VARCHAR(128) NULL COMMENT '权限码，格式 module:entity:action',
    url_pattern     VARCHAR(256) NULL COMMENT '仅 ENDPOINT，路由模板',
    http_method     VARCHAR(16) NULL COMMENT '仅 ENDPOINT: GET/POST/PUT/DELETE',
    app_code        VARCHAR(64) NOT NULL COMMENT '所属应用编码',
    module_code     VARCHAR(64) NULL COMMENT '所属功能模块',
    visible         TINYINT NOT NULL DEFAULT 1 COMMENT '1=显示 0=隐藏',
    auth_required   TINYINT NOT NULL DEFAULT 1 COMMENT '1=需权限校验 0=免检',
    status          VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' COMMENT 'ACTIVE/DEPRECATED',
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_code (app_code),
    INDEX idx_method_url (http_method, url_pattern)
);

-- ==================== 角色-接口权限关联 ====================
CREATE TABLE admin_role_api (
    id                  BIGINT PRIMARY KEY,
    role_id             BIGINT NOT NULL,
    api_permission_id   BIGINT NOT NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_api (role_id, api_permission_id)
);

-- ==================== 角色-应用关联（保存权限时自动维护） ====================
CREATE TABLE admin_role_app (
    id              BIGINT PRIMARY KEY,
    role_id         BIGINT NOT NULL,
    app_code        VARCHAR(64) NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_app (role_id, app_code)
);

-- ==================== 数据权限维度注册（全局） ====================
CREATE TABLE admin_data_scope_config (
    id              BIGINT PRIMARY KEY,
    dimension_name  VARCHAR(64) NOT NULL UNIQUE,
    display_name    VARCHAR(128) NOT NULL,
    table_column    VARCHAR(128) NOT NULL,
    supported_scope_types JSON NOT NULL DEFAULT '["ALL","SELF","CUSTOM"]' COMMENT '该维度支持的 scope_type',
    value_source    VARCHAR(32) NOT NULL COMMENT 'user_attr/role_config/custom/org',
    handler_name    VARCHAR(256) NULL COMMENT 'SPI handler 名称',
    status          TINYINT NOT NULL DEFAULT 1,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ==================== 数据权限绑定（挂角色） ====================
CREATE TABLE admin_data_scope (
    id              BIGINT PRIMARY KEY,
    role_id         BIGINT NOT NULL,
    dimension_name  VARCHAR(64) NOT NULL,
    scope_type      VARCHAR(16) NOT NULL DEFAULT 'CUSTOM' COMMENT 'ALL/SELF/DEPT/DEPT_TREE/CUSTOM',
    target_entity   VARCHAR(64) NULL COMMENT '绑定业务实体',
    dimension_values JSON NULL COMMENT 'scope_type=CUSTOM 时的值列表',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_role_dimension (role_id, dimension_name)
);

-- ==================== 三级配置表 ====================
CREATE TABLE admin_config (
    id              BIGINT PRIMARY KEY,
    config_key      VARCHAR(128) NOT NULL,
    config_value    TEXT NULL,
    config_type     VARCHAR(32) NOT NULL DEFAULT 'string' COMMENT 'string/number/boolean/json',
    scope           VARCHAR(16) NOT NULL DEFAULT 'SYSTEM' COMMENT 'SYSTEM/TENANT/USER',
    scope_id        BIGINT NOT NULL DEFAULT 0 COMMENT 'SYSTEM=0, TENANT=tenantId, USER=userId',
    tenant_id       BIGINT NOT NULL DEFAULT 0 COMMENT '所属租户(SYSTEM=0,TENANT=tenantId,USER=tenantId)',
    display_name    VARCHAR(128) NULL,
    description     VARCHAR(512) NULL,
    is_feature_flag TINYINT NOT NULL DEFAULT 0 COMMENT '1=功能开关',
    status          TINYINT NOT NULL DEFAULT 1,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_scope_key (scope, scope_id, tenant_id, config_key)
);

-- ==================== 插件管理表 ====================
CREATE TABLE sys_plugin (
    id              BIGINT PRIMARY KEY,
    name            VARCHAR(64) NOT NULL UNIQUE COMMENT '插件名（= app_code）',
    version         VARCHAR(32) NOT NULL,
    display_name    VARCHAR(128) NULL,
    description     VARCHAR(512) NULL,
    status          TINYINT NOT NULL DEFAULT 0 COMMENT '0=已安装 1=运行中 2=已停止 3=异常',
    binary_path     VARCHAR(512) NULL,
    frontend_path   VARCHAR(512) NULL,
    manifest        JSON NULL COMMENT '完整 manifest 快照',
    installed_at    DATETIME NULL,
    started_at      DATETIME NULL,
    stopped_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- ==================== 字段级权限 ====================
CREATE TABLE admin_field_permission (
    id          BIGINT PRIMARY KEY,
    role_id     BIGINT NOT NULL,
    object_code VARCHAR(64) NOT NULL COMMENT '业务对象标识（如 invoice、customer）',
    field_name  VARCHAR(64) NOT NULL COMMENT '字段名（对应 JSON key）',
    access      VARCHAR(16) NOT NULL DEFAULT 'VISIBLE' COMMENT 'VISIBLE/EDITABLE/HIDDEN',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_object_field (role_id, object_code, field_name)
);

-- ==================== 记录共享规则 ====================
CREATE TABLE admin_record_share (
    id            BIGINT PRIMARY KEY,
    tenant_id     BIGINT NOT NULL,
    object_code   VARCHAR(64) NOT NULL COMMENT '业务对象',
    record_id     BIGINT NOT NULL COMMENT '被共享的记录 ID',
    share_to_type VARCHAR(16) NOT NULL COMMENT 'USER/ROLE/DEPT',
    share_to_id   BIGINT NOT NULL COMMENT '共享目标 ID',
    access_level  VARCHAR(16) NOT NULL DEFAULT 'READ' COMMENT 'READ/EDIT',
    expire_at     DATETIME NULL COMMENT '共享过期时间（NULL=永久）',
    created_by    BIGINT NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_record (tenant_id, object_code, record_id),
    INDEX idx_share_to (share_to_type, share_to_id, tenant_id)
);

-- ==================== 审批流程定义 ====================
CREATE TABLE admin_approval_flow (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL DEFAULT 0 COMMENT '0=全局默认流程, >0=租户自定义',
    flow_code       VARCHAR(64) NOT NULL COMMENT 'APP_SUBSCRIBE/ROLE_REQUEST/CUSTOM',
    name            VARCHAR(128) NOT NULL,
    description     VARCHAR(512) NULL,
    flow_config     JSON NOT NULL COMMENT '流程配置（节点定义）',
    status          TINYINT NOT NULL DEFAULT 1,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tenant_flow (tenant_id, flow_code)
);

-- ==================== 审批实例 ====================
CREATE TABLE admin_approval (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    flow_code       VARCHAR(64) NOT NULL,
    title           VARCHAR(256) NOT NULL,
    applicant_id    BIGINT NOT NULL,
    payload         JSON NOT NULL COMMENT '申请内容',
    status          VARCHAR(16) NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING/APPROVED/REJECTED/CANCELLED',
    current_node    INT NOT NULL DEFAULT 0,
    result_remark   VARCHAR(512) NULL,
    completed_at    DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_status (tenant_id, status),
    INDEX idx_applicant (applicant_id)
);

-- ==================== 审批节点记录 ====================
CREATE TABLE admin_approval_node (
    id              BIGINT PRIMARY KEY,
    approval_id     BIGINT NOT NULL,
    node_seq        INT NOT NULL COMMENT '节点序号',
    node_type       VARCHAR(16) NOT NULL COMMENT 'SINGLE/AND_SIGN/OR_SIGN',
    approver_type   VARCHAR(16) NOT NULL COMMENT 'USER/ROLE/DEPT_HEAD/APPLICANT_HEAD',
    approver_ids    JSON NULL,
    status          VARCHAR(16) NOT NULL DEFAULT 'WAITING' COMMENT 'WAITING/PROCESSING/APPROVED/REJECTED',
    handled_by      BIGINT NULL,
    remark          VARCHAR(512) NULL,
    handled_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_approval (approval_id, node_seq)
);

-- ==================== 操作日志 ====================
CREATE TABLE admin_operation_log (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    tenant_id       BIGINT NULL,
    module          VARCHAR(64) NOT NULL,
    action          VARCHAR(32) NOT NULL,
    target_type     VARCHAR(64) NOT NULL,
    target_id       BIGINT NULL,
    summary         VARCHAR(512) NOT NULL,
    old_value       JSON NULL,
    new_value       JSON NULL,
    client_ip       VARCHAR(64) NULL,
    user_agent      VARCHAR(256) NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_tenant_module (tenant_id, module),
    INDEX idx_created_at (created_at)
);

-- ==================== 登录日志 ====================
CREATE TABLE admin_login_log (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NULL,
    username        VARCHAR(128) NOT NULL,
    ip              VARCHAR(64) NULL,
    location        VARCHAR(256) NULL,
    browser         VARCHAR(256) NULL,
    os              VARCHAR(128) NULL,
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '1=成功 0=失败',
    message         VARCHAR(512) NULL,
    login_time      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_login_time (login_time)
);
```

---

## 10. API 设计

所有业务接口统一 `/api/v1/` 前缀，所有 ID 使用 string 类型传输（雪花 ID 精度保护）。

### 10.1 认证

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | /auth/login | 管理端登录（grant_type 区分策略） | 无 |
| POST | /auth/user/login | 用户端登录（sms/wechat） | 无 |
| POST | /auth/tenant/select | 选择/切换租户（含 platform 参数） | platform_token |
| POST | /auth/refresh | 刷新 access_token | platform_token |
| POST | /auth/logout | 登出 | access_token |

### 10.2 租户管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/tenants | 租户列表 | access_token + SUPER_ADMIN |
| POST | /api/v1/admin/tenants | 创建租户 | access_token + SUPER_ADMIN |
| PUT | /api/v1/admin/tenants/:id | 更新租户 | access_token + SUPER_ADMIN |
| PUT | /api/v1/admin/tenants/:id/status | 启用/禁用 | access_token + SUPER_ADMIN |
| DELETE | /api/v1/admin/tenants/:id | 删除租户 | access_token + SUPER_ADMIN |
| PUT | /api/v1/admin/tenants/:id/apps | 设置租户订阅应用 | access_token + SUPER_ADMIN |
| GET | /api/v1/admin/tenants/:id/apps | 租户已订阅应用列表 | access_token |
| PUT | /api/v1/admin/tenants/:id/apps/:appCode/modules | 设置租户应用模块开关 | access_token + SUPER_ADMIN |

### 10.3 用户管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/users | 用户列表（租户上下文过滤） | access_token |
| POST | /api/v1/admin/users | 创建用户 | access_token |
| PUT | /api/v1/admin/users/:id | 更新用户 | access_token |
| DELETE | /api/v1/admin/users/:id | 删除用户 | access_token |
| POST | /api/v1/admin/users/:id/tenants | 关联租户 | access_token |
| DELETE | /api/v1/admin/users/:id/tenants/:tenantId | 解除租户关联 | access_token |
| POST | /api/v1/admin/users/:id/roles | 分配角色 | access_token |
| PUT | /api/v1/admin/users/:id/roles | 全量替换角色 | access_token |
| POST | /api/v1/admin/users/:id/force-offline | 强制下线 | access_token |

### 10.4 角色管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/roles | 角色列表/树（当前租户） | access_token |
| POST | /api/v1/admin/roles | 创建角色 | access_token |
| PUT | /api/v1/admin/roles/:id | 更新角色 | access_token |
| DELETE | /api/v1/admin/roles/:id | 删除角色 | access_token |
| GET | /api/v1/admin/roles/:id/resources | 角色已分配资源 ID | access_token |
| PUT | /api/v1/admin/roles/:id/resources | 分配菜单权限（权限上界校验） | access_token |
| GET | /api/v1/admin/roles/:id/apis | 角色已分配接口 ID | access_token |
| PUT | /api/v1/admin/roles/:id/apis | 分配接口权限（权限上界校验） | access_token |
| GET | /api/v1/admin/roles/:id/assignable-resources | 可分配资源树（受父角色约束） | access_token |
| GET | /api/v1/admin/roles/:id/assignable-apis | 可分配接口树（受父角色约束） | access_token |
| GET | /api/v1/admin/roles/:id/permission-summary | 各应用权限统计摘要 | access_token |
| PUT | /api/v1/admin/roles/:id/data-scopes | 配置数据权限 | access_token |
| GET | /api/v1/admin/roles/:id/data-scopes | 查询数据权限 | access_token |

### 10.5 菜单/资源管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/resources/tree | 资源树（?app_code&platform） | access_token |
| POST | /api/v1/admin/resources | 创建资源 | access_token |
| PUT | /api/v1/admin/resources/:id | 更新资源 | access_token |
| DELETE | /api/v1/admin/resources/:id | 删除资源 | access_token |
| PUT | /api/v1/admin/resources/sort | 拖拽排序 | access_token |
| GET | /api/v1/common/user-menu | 当前用户菜单树（?platform=admin\|user） | access_token（免权限码检查） |

### 10.6 接口权限管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/api-permissions/tree | 接口权限树（?app_code） | access_token |
| POST | /api/v1/admin/api-permissions | 创建节点 | access_token |
| PUT | /api/v1/admin/api-permissions/:id | 更新节点 | access_token |
| DELETE | /api/v1/admin/api-permissions/:id | 删除节点 | access_token |
| PUT | /api/v1/admin/api-permissions/:id/move | 移动节点 | access_token |

### 10.7 应用管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/applications | 应用列表 | access_token |
| POST | /api/v1/admin/applications | 创建应用 | access_token + SUPER_ADMIN |
| PUT | /api/v1/admin/applications/:id | 更新应用 | access_token + SUPER_ADMIN |
| DELETE | /api/v1/admin/applications/:id | 删除应用 | access_token + SUPER_ADMIN |

### 10.8 插件管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/plugins | 插件列表 | access_token + SUPER_ADMIN |
| POST | /api/v1/admin/plugins/install | 上传安装插件 | access_token + SUPER_ADMIN |
| POST | /api/v1/admin/plugins/:name/start | 启动插件 | access_token + SUPER_ADMIN |
| POST | /api/v1/admin/plugins/:name/stop | 停止插件 | access_token + SUPER_ADMIN |
| DELETE | /api/v1/admin/plugins/:name | 卸载插件 | access_token + SUPER_ADMIN |
| GET | /api/v1/admin/plugins/:name/health | 健康检查 | access_token + SUPER_ADMIN |

### 10.9 数据权限配置

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/data-scope-configs | 维度列表 | access_token |
| POST | /api/v1/admin/data-scope-configs | 注册维度 | access_token |
| PUT | /api/v1/admin/data-scope-configs/:id | 更新维度 | access_token |
| DELETE | /api/v1/admin/data-scope-configs/:id | 删除维度 | access_token |

### 10.10 配置管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/configs | 配置列表（?scope&scope_id） | access_token |
| GET | /api/v1/admin/configs/resolve/:key | 解析配置值（三级合并） | access_token |
| POST | /api/v1/admin/configs | 创建/覆盖配置 | access_token |
| PUT | /api/v1/admin/configs/:id | 更新配置 | access_token |
| DELETE | /api/v1/admin/configs/:id | 删除配置 | access_token |
| GET | /api/v1/admin/configs/feature-flags | 功能开关列表（当前租户） | access_token |
| PUT | /api/v1/admin/configs/feature-flags/:key | 设置功能开关 | access_token |

### 10.11 日志管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/operation-logs | 操作日志列表 | access_token |
| GET | /api/v1/admin/login-logs | 登录日志列表 | access_token |

### 10.12 插件业务接口（由插件自行注册）

```
/api/v1/{plugin_app_code}/...  → 由 PluginProxy 转发到对应插件进程
```

---

## 11. 前端架构

### 11.1 工程结构

```
dev-web-admin/          # 管理端前端
├── src/
│   ├── views/          # 管理端页面
│   ├── plugin-sdk/     # 共享 Plugin SDK
│   ├── plugin-loader/  # 插件加载器
│   └── ...
├── pc/                 # PC 入口
└── h5/                 # H5 入口（共享 src，不同布局）

dev-web-user/           # 用户端前端
├── src/
│   ├── views/          # 用户端页面
│   ├── plugin-sdk/     # 共享 Plugin SDK（同一个包）
│   ├── plugin-loader/  # 插件加载器
│   └── ...
├── pc/                 # PC 入口
└── h5/                 # H5 入口
```

### 11.2 Plugin SDK V2

```typescript
export interface PluginConfig {
  manifest: PluginManifest
  routes?: PluginRoute[]
  menus?: PluginMenu[]
  permissions?: string[]
  extensions?: Record<string, any>
  setup?: (ctx: PluginContext) => void | Promise<void>
  teardown?: (ctx: PluginContext) => void | Promise<void>  // V2 新增
}

export interface PluginContext {
  currentUser: Readonly<UserInfo>
  currentTenant: Readonly<TenantInfo>
  permissions: Readonly<string[]>
  platform: 'admin' | 'user'        // V2 新增：当前运行平台
  device: 'pc' | 'h5'               // V2 新增：当前终端形态
  eventBus: EventBus
  registerExtension: (point: string, component: any, sort?: number) => UnregisterFn  // V2: 返回清理函数
  i18n: { mergeLocale: (locale: Record<string, any>) => void }  // V2 新增
}

type UnregisterFn = () => void
```

### 11.3 插件前端加载

```typescript
// 框架侧 plugin-loader
async function loadPlugin(pluginName: string, platform: 'admin' | 'user', device: 'pc' | 'h5') {
  const bundlePath = `/static/plugins/${pluginName}/${platform}-${device}/index.js`
  const module = await import(bundlePath)
  const config: PluginConfig = module.default
  
  const ctx = createPluginContext(pluginName, platform, device)
  
  if (config.setup) {
    await config.setup(ctx)
  }
  
  // 注册路由、菜单、扩展点
  registerPluginRoutes(config.routes)
  
  // 记录清理函数
  pluginTeardowns.set(pluginName, async () => {
    if (config.teardown) await config.teardown(ctx)
    unregisterPluginRoutes(pluginName)
  })
}

async function unloadPlugin(pluginName: string) {
  const teardown = pluginTeardowns.get(pluginName)
  if (teardown) await teardown()
  pluginTeardowns.delete(pluginName)
}
```

### 11.4 管理端与用户端的菜单加载

```typescript
// dev-web-admin 中:
const menus = await api.get('/api/v1/common/user-menu?platform=admin')

// dev-web-user 中:
const menus = await api.get('/api/v1/common/user-menu?platform=user')
```

后端根据 platform 参数过滤返回对应平台的菜单树。

> **注意**：user-menu 接口放在 `/api/v1/common/` 公共路径下，不走 AppResolveMiddleware 的应用订阅校验（因为管理端和用户端都需要调用，且用户端用户未必订阅了 platform_admin 应用）。该接口仅需 access_token 认证，无需权限码检查。

---

## 12. 初始化与种子数据

### 12.1 启动流程

1. GORM AutoMigrate 创建所有表
2. 幂等 Seed：admin_user 表无记录时执行
3. 内置应用注册：确保 platform_admin 应用存在
4. API 自动发现：扫描路由注册到 admin_api_permission
5. 插件恢复：读取 sys_plugin 中 status=1 的插件自动启动

### 12.2 Seed 数据

| 步骤 | 数据 |
|------|------|
| 1 | 创建超级管理员 admin / admin123 |
| 2 | 创建默认租户 default |
| 3 | 关联 admin → default 租户 |
| 4 | 创建 SUPER_ADMIN 角色（tenant_id=default） |
| 5 | 分配 admin → SUPER_ADMIN 角色 |
| 6 | 创建内置应用 platform_admin |
| 7 | default 租户订阅 platform_admin |
| 8 | 注册管理端菜单（platform=admin） |
| 9 | 注册系统默认配置 |

### 12.3 管理端菜单种子

```
├── 首页 (/home)                           [platform=admin]
├── 系统管理 (/system)                      [platform=admin]
│   ├── 用户管理 (/system/users)
│   ├── 角色管理 (/system/roles)
│   ├── 租户管理 (/system/tenants)
│   ├── 应用管理 (/system/applications)
│   ├── 数据权限配置 (/system/data-scope)
│   ├── 系统配置 (/system/config)
│   ├── 接口管理 (/system/api)
│   └── 插件管理 (/system/plugins)
├── 日志管理 (/log)                         [platform=admin]
│   ├── 登录日志 (/log/login-logs)
│   └── 操作日志 (/log/operation-logs)
└── 监控 (/monitor)                        [platform=admin]
    └── 服务监控 (/monitor/server)
```

---

## 13. SPI 接口清单

```go
// ===================== 认证 =====================
type AuthenticationStrategy interface {
    Type() string
    Authenticate(ctx context.Context, credentials map[string]string) (*AuthResult, error)
    Supports(credentials map[string]string) bool
}

// ===================== 用户 =====================
type UserProvider interface {
    LoadByUsername(username string) (*AuthUser, error)
    LoadByExternalID(authType, externalID string) (*AuthUser, error)
}

// ===================== 组织架构 =====================
type OrganizationProvider interface {
    GetUserDeptIds(userID, tenantID int64) ([]int64, error)
    GetDeptTree(deptID, tenantID int64) ([]int64, error)  // 含自身及所有下级
    GetSubDeptIds(deptID, tenantID int64) ([]int64, error)
}

// ===================== 缓存 =====================
type CacheAdapter interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration)
    Delete(keys ...string)
    DeleteByPrefix(prefix string)
}

// ===================== Token 黑名单 =====================
type TokenBlacklistStore interface {
    Add(tokenJTI string, ttl time.Duration) error
    Contains(tokenJTI string) bool
}

// ===================== 数据权限 =====================
type DataScopeHandler interface {
    Handle(ctx context.Context, dimension string, scopeType string, userID int64) ([]string, error)
}

// ===================== 事件 =====================
type EventPublisher interface {
    Publish(event string, payload interface{}) error
    Subscribe(event string, handler func(payload interface{})) error
}

// ===================== 操作日志 =====================
type OperationLogger interface {
    Log(ctx context.Context, entry OperationLogEntry) error
}
```

---

## 14. 业务应用数据隔离规范

### 14.1 业务表强制规范

所有接入平台的业务应用（BUILTIN/PLUGIN）的数据表必须遵循：

| 规范 | 说明 |
|------|------|
| tenant_id 必备 | 所有业务表必须包含 `tenant_id BIGINT NOT NULL` 字段 |
| 自动注入 | GORM 全局 Callback 对带 tenant_id 的表自动注入 `WHERE tenant_id = ?` |
| 创建时自动填充 | Create 操作自动从 AuthContext 填充 tenant_id |
| 跳过机制 | 系统操作使用 `db.WithContext(SystemOpContext())` 跳过注入 |

### 14.2 租户隔离 GORM Callback

```go
// TenantIsolationCallback 自动注入租户过滤
func TenantIsolationCallback(db *gorm.DB) {
    if isSystemOp(db.Statement.Context) { return }
    
    tenantID := getTenantIDFromContext(db.Statement.Context)
    if tenantID == 0 { return }
    
    // 检查目标表是否有 tenant_id 字段
    if hasTenantColumn(db.Statement.Schema) {
        db.Where("tenant_id = ?", tenantID)
    }
}
```

### 14.3 插件数据表约定

- 插件业务表命名格式：`{plugin_name}_{entity}`（如 `billing_invoice`）
- 必须包含 `tenant_id` + `created_at` + `updated_at`
- 插件通过 `HttpRequest.TenantId` 获得当前租户 ID
- 卸载时可选清理插件表（通过 `Installer.Uninstall(cleanData=true)`）

---

## 15. 配额管理

### 15.1 配额定义

通过 `admin_config`（scope=TENANT）存储租户配额：

| config_key | 含义 | 示例值 |
|-----------|------|--------|
| quota.max_admin_users | 最大管理端用户数 | 50 |
| quota.max_biz_users | 最大 C 端用户数 | 10000 |
| quota.max_roles | 最大角色数 | 20 |
| quota.max_apps | 最大可订阅应用数 | 5 |
| quota.storage_mb | 存储配额（MB） | 1024 |

### 15.2 配额校验

Service 层在关键操作前查询配额：

```go
func (s *UserService) CreateUser(tenantID int64, req *CreateUserRequest) error {
    // 查询当前租户用户数
    currentCount := s.countUsersByTenant(tenantID)
    // 查询配额
    maxUsers := s.configService.ResolveInt(tenantID, "quota.max_admin_users", 999999)
    if currentCount >= maxUsers {
        return errors.NewAuthError(errors.ErrQuotaExceeded, "管理用户数已达上限")
    }
    // ... 正常创建
}
```

### 15.3 配额无配置时的默认行为

- 未配置配额 → 系统默认值（通过 SYSTEM scope 的 admin_config 设定）
- 默认值建议设为较大数值（如 999999），表示「不限制」

---

## 16. 操作日志与审计

### 16.1 日志可见性权限

| 角色 | 可查看范围 |
|------|-----------|
| SUPER_ADMIN | 全部操作日志（跨租户） |
| TENANT_ADMIN | 本租户操作日志（WHERE tenant_id = 当前租户） |
| 普通角色 | 仅自己的操作记录（WHERE user_id = 当前用户）或无权限 |

### 16.2 高风险操作标记

以下操作在日志中标记为「高风险」（`risk_level = 'HIGH'`），并支持独立查询/告警：
- 删除租户
- 卸载插件
- 修改认证策略配置
- 删除 SUPER_ADMIN 用户
- 批量删除用户/角色
- 修改数据权限维度

### 16.3 日志保留策略

- 操作日志默认保留 180 天
- 登录日志默认保留 90 天
- 保留时间可通过 admin_config（scope=SYSTEM）配置
- 过期数据通过定时任务清理（或归档到冷存储）

---

---

## 17. 审批流引擎

### 18.1 设计定位

轻量级审批引擎，支持从简单的「单人审批」到「多级会签/或签」的扩展路径。当前实现单级审批，数据模型预留多级扩展。

### 18.2 数据模型

```sql
-- 审批流程定义（全局，按租户可覆盖）
CREATE TABLE admin_approval_flow (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL DEFAULT 0 COMMENT '0=全局默认流程, >0=租户自定义',
    flow_code       VARCHAR(64) NOT NULL COMMENT '流程标识: APP_SUBSCRIBE/ROLE_REQUEST/CUSTOM',
    name            VARCHAR(128) NOT NULL,
    description     VARCHAR(512) NULL,
    flow_config     JSON NOT NULL COMMENT '流程配置（节点定义）',
    status          TINYINT NOT NULL DEFAULT 1,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tenant_flow (tenant_id, flow_code)
);

-- 审批实例（每次发起审批生成一条）
CREATE TABLE admin_approval (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    flow_code       VARCHAR(64) NOT NULL COMMENT '关联流程定义',
    title           VARCHAR(256) NOT NULL COMMENT '审批标题',
    applicant_id    BIGINT NOT NULL COMMENT '申请人 ID',
    payload         JSON NOT NULL COMMENT '申请内容（业务数据快照）',
    status          VARCHAR(16) NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING/APPROVED/REJECTED/CANCELLED',
    current_node    INT NOT NULL DEFAULT 0 COMMENT '当前所在节点序号',
    result_remark   VARCHAR(512) NULL COMMENT '最终审批意见',
    completed_at    DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_status (tenant_id, status),
    INDEX idx_applicant (applicant_id)
);

-- 审批节点记录（每个审批节点一条，支持多级）
CREATE TABLE admin_approval_node (
    id              BIGINT PRIMARY KEY,
    approval_id     BIGINT NOT NULL COMMENT '关联审批实例',
    node_seq        INT NOT NULL COMMENT '节点序号（从 0 开始）',
    node_type       VARCHAR(16) NOT NULL COMMENT 'SINGLE/AND_SIGN/OR_SIGN',
    approver_type   VARCHAR(16) NOT NULL COMMENT 'USER/ROLE/DEPT_HEAD/APPLICANT_HEAD',
    approver_ids    JSON NULL COMMENT '指定审批人 ID 列表（USER 类型时）',
    status          VARCHAR(16) NOT NULL DEFAULT 'WAITING' COMMENT 'WAITING/PROCESSING/APPROVED/REJECTED',
    handled_by      BIGINT NULL COMMENT '实际处理人',
    remark          VARCHAR(512) NULL,
    handled_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_approval (approval_id, node_seq)
);
```

### 18.3 flow_config 结构

```json
{
  "nodes": [
    {
      "seq": 0,
      "name": "直属上级审批",
      "type": "SINGLE",
      "approver_type": "APPLICANT_HEAD",
      "approver_ids": null
    },
    {
      "seq": 1,
      "name": "部门负责人会签",
      "type": "AND_SIGN",
      "approver_type": "ROLE",
      "approver_ids": ["dept_manager"]
    }
  ]
}
```

### 18.4 审批类型

| flow_code | 场景 | 默认流程 |
|-----------|------|---------|
| APP_SUBSCRIBE | 租户申请订阅应用 | SUPER_ADMIN 单人审批 |
| ROLE_REQUEST | 用户申请角色权限 | TENANT_ADMIN 单人审批 |
| DATA_EXPORT | 敏感数据导出 | TENANT_ADMIN 单人审批 |
| CUSTOM | 业务自定义（由插件注册） | 按 flow_config 走 |

### 18.5 节点类型

| node_type | 含义 | 通过条件 |
|-----------|------|---------|
| SINGLE | 单人审批 | 任一审批人通过即通过 |
| AND_SIGN | 会签 | 所有审批人都通过才通过 |
| OR_SIGN | 或签 | 任一审批人通过即通过（其余自动跳过） |

### 18.6 审批人确定方式

| approver_type | 含义 | 审批人来源 |
|---------------|------|-----------|
| USER | 指定用户 | approver_ids 列表 |
| ROLE | 指定角色 | 该租户下拥有该角色的所有用户 |
| DEPT_HEAD | 申请人部门主管 | 通过 OrganizationProvider 获取 |
| APPLICANT_HEAD | 申请人直属上级 | 通过 OrganizationProvider 获取 |

### 18.7 扩展预留

- 流程定义支持租户覆盖（tenant_id=0 为默认，租户可自定义同 flow_code 的流程）
- 支持条件分支（后续在 flow_config 中增加 `condition` 字段）
- 支持抄送人（node_type=CC，不需审批仅通知）
- 支持超时自动通过/驳回（node 增加 timeout_hours + timeout_action）
- 审批完成后通过 EventPublisher 发布事件（如 `approval.completed`），业务侧订阅执行后续动作

### 18.8 API

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | /api/v1/admin/approvals | 发起审批 | access_token |
| GET | /api/v1/admin/approvals | 审批列表（我发起的/待我审批的） | access_token |
| GET | /api/v1/admin/approvals/:id | 审批详情 | access_token |
| POST | /api/v1/admin/approvals/:id/approve | 通过 | access_token |
| POST | /api/v1/admin/approvals/:id/reject | 驳回 | access_token |
| POST | /api/v1/admin/approvals/:id/cancel | 撤销（仅申请人） | access_token |
| GET | /api/v1/admin/approval-flows | 流程定义列表 | access_token |
| POST | /api/v1/admin/approval-flows | 创建/覆盖流程定义 | access_token |

---

## 18. 扩展能力

### 19.1 多时区/多语言/多币种

`admin_tenant` 增加租户级国际化配置：

```sql
-- 在 admin_tenant 表增加字段
ALTER TABLE admin_tenant ADD COLUMN timezone VARCHAR(64) DEFAULT 'Asia/Shanghai' COMMENT '租户时区';
ALTER TABLE admin_tenant ADD COLUMN locale VARCHAR(16) DEFAULT 'zh-CN' COMMENT '默认语言';
ALTER TABLE admin_tenant ADD COLUMN currency VARCHAR(8) DEFAULT 'CNY' COMMENT '默认币种';
```

**行为**：
- 后端时间存储统一 UTC，返回给前端时根据租户 timezone 转换
- 前端通过 tenant info 获取 locale 设置 i18n
- 币种影响金额显示格式（由业务应用自行处理）

### 19.2 通用扩展字段

业务实体表通过 `ext_fields` JSON 列支持灵活扩展，无需 DDL 变更：

```sql
-- 所有业务表建议预留
ext_fields JSON NULL COMMENT '扩展字段（动态，无需改表结构）'
```

**biz_user 已有 `tags` JSON 列，同理。**

**使用约定**：
- `ext_fields` 中的字段可被前端动态渲染（基于配置）
- 支持按 ext_fields 中的 key 做筛选查询（MySQL JSON 函数）
- 如某个扩展字段被频繁查询，考虑升级为独立列

### 19.3 扩展字段元数据（预留）

后续如需 Low-Code 级的自定义字段能力，可引入：

```sql
CREATE TABLE admin_custom_field (
    id          BIGINT PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    object_code VARCHAR(64) NOT NULL COMMENT '所属业务对象',
    field_key   VARCHAR(64) NOT NULL COMMENT '字段标识',
    field_label VARCHAR(128) NOT NULL COMMENT '显示名',
    field_type  VARCHAR(32) NOT NULL COMMENT 'STRING/NUMBER/DATE/ENUM/RELATION',
    config      JSON NULL COMMENT '字段配置（枚举选项、校验规则等）',
    required    TINYINT NOT NULL DEFAULT 0,
    sort_order  INT NOT NULL DEFAULT 0,
    status      TINYINT NOT NULL DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tenant_object_field (tenant_id, object_code, field_key)
);
```

当前版本**不实现**，仅预留设计方向。

---

## 19. 通知与告警（待后续 CR 详细设计）

### 19.1 预留的通知场景

| 事件 | 通知对象 | 渠道（待定） |
|------|---------|------------|
| 租户即将到期（提前 7 天） | TENANT_ADMIN | 站内信 / 邮件 / 短信 |
| 配额即将用满（>80%） | TENANT_ADMIN | 站内信 |
| 插件异常/崩溃 | SUPER_ADMIN | 站内信 / 邮件 |
| 高风险操作执行 | SUPER_ADMIN | 站内信 |
| 新插件/应用上架 | TENANT_ADMIN | 站内信 |
| 审批待处理 | 审批人 | 站内信 / 推送 |

### 19.2 技术预留

- EventBus 已支持事件广播，通知服务可作为内部订阅者
- 通知表（admin_notification）和推送渠道适配器在后续 CR 中实现
- 当前版本仅记录操作日志，不发送主动通知

---

## 20. V1 → V2 变更对照

| 维度 | V1 | V2 |
|------|----|----|
| 应用模型 | 仅 app_code + name | 增加 app_type / route_prefix / platforms / modules |
| 插件权限 | 独立 Registry + sys_menu | 统一到 admin_resource + admin_api_permission |
| 角色继承 | parent_id 未使用 | parent_id = 权限上界，写时校验 + 级联裁剪 |
| 权限集 | 无 | PERMISSION_SET 类型角色，叠加增量权限 |
| 字段级权限 | 无 | admin_field_permission 表，按对象+字段+角色控制 VISIBLE/EDITABLE/HIDDEN |
| 记录共享 | 无 | admin_record_share 表，支持将记录共享给用户/角色/部门 |
| 数据权限 | 仅 CUSTOM 固定值 | scope_type: ALL/SELF/DEPT/DEPT_TREE/CUSTOM + 记录共享 OR 补充 |
| 配置 | sys_config 单层 | admin_config 三级（SYSTEM/TENANT/USER）+ 功能开关 + 配额 |
| 缓存失效 | 无主动刷新 | 事件驱动 + 版本化 codeMap |
| 认证 | 硬编码 password | 可插拔 AuthenticationStrategy |
| 用户模型 | 单表 admin_user | 双用户池：admin_user（管理端）+ biz_user（C端） |
| C端权限 | 无 | 简化模型：订阅+模块校验，不走 RBAC permission_code |
| 前端 | 单一管理端 | dev-web-admin + dev-web-user，各支持 PC/H5 |
| 资源 platform | 无 | admin_resource.platform 区分管理端/用户端菜单 |
| 插件通信 | 无契约 CallPlugin | 声明式 Action + 版本兼容校验 |
| Plugin SDK | 无 teardown | setup + teardown + unregister + platform/device 感知 |
| 租户功能控制 | 仅订阅/取消 | enabled_modules 模块级开关 |
| 租户状态 | 0/1 二态 | 四态：正常/禁用/只读/注销中 |
| 管理员分级 | 隐含 | 显式定义 SUPER_ADMIN / TENANT_ADMIN / NORMAL 操作边界 |
| 数据隔离 | 仅平台表 | 补充业务表强制 tenant_id + GORM Callback 自动注入 |
| 配额 | 无 | admin_config 存储配额 + Service 层校验 |
| 审批流 | 无 | 轻量级审批引擎（单级/会签/或签，预留多级扩展） |
| 扩展字段 | 无 | ext_fields JSON 列 + 自定义字段元数据预留 |
| 多时区/多语言 | 无 | admin_tenant 增加 timezone/locale/currency |
| 操作日志 | 无权限区分 | 按角色分级查看 + 高风险标记 |
| 通知 | 无 | EventBus 预留 + 通知场景定义（待后续实现） |
| 应用发现 | 手动上传 | 应用目录 + 自助/审批订阅流程 |
| C端注册 | 无 | 多渠道注册（短信/微信/邀请码）+ 租户绑定 |
