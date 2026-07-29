# 设计：V2-CR8 缓存事件驱动 + OAuth2 预留 + 扩展能力收尾

## 技术方案

### Phase 1: Redis L2 缓存 + 事件驱动失效

#### 1.1 PermCodeCache 结构（新增）

新增 `common/auth/cache/perm_code_cache.go`，封装 `DynamicPermissionMiddleware` 专用的用户权限码 Redis 缓存逻辑：

```go
// PermCodeCache 用户权限码 Redis 缓存（DynamicPermissionMiddleware 专用）
// 与现有 PermissionCache（角色级 L1/L2 内存缓存）完全独立
type PermCodeCache struct {
    adapter storage.AdapterCache  // sdk.Runtime.GetCacheAdapter() 返回的接口（go-admin-core/storage 包定义）
    ttl     time.Duration         // 默认 10 分钟
}
// 注：storage.AdapterCache 为 go-admin-core 框架定义的缓存适配器接口，
// 支持 Get/Set/Del 操作。具体类型在执行阶段对齐框架实际导出类型。

func NewPermCodeCache(adapter storage.AdapterCache, ttl time.Duration) *PermCodeCache

// Get 获取用户权限码列表，未命中返回 nil, false
func (c *PermCodeCache) Get(tenantID, userID int64) ([]string, bool)

// Set 写入用户权限码列表
func (c *PermCodeCache) Set(tenantID, userID int64, codes []string)

// Invalidate 删除指定用户的缓存
func (c *PermCodeCache) Invalidate(tenantID, userID int64)

// key 格式: perm:{tenantID}:{userID}
func (c *PermCodeCache) key(tenantID, userID int64) string
```

**降级逻辑：** 如果 `adapter` 为 nil（安装前或 Redis 不可用），所有方法 no-op（Get 返回 false，Set/Invalidate 不操作）。

#### 1.2 DynamicPermissionMiddleware 改造

改造 `getUserPermCodes` 调用链：

```
getUserPermCodes(db, userID, tenantID)
    │
    ├── permCodeCache.Get(tenantID, userID)
    │       ├── 命中 → 返回 []string
    │       └── 未命中 ↓
    │
    ├── DB 3表JOIN 查询 → codes
    │
    ├── permCodeCache.Set(tenantID, userID, codes)
    │
    └── 返回 codes
```

中间件构造函数新增可选参数 `permCodeCache *cache.PermCodeCache`（nil 则不启用缓存，行为不变）。

#### 1.3 事件定义与发布

在 `common/event/bus.go` 新增事件 payload：

```go
// PermissionChangedEvent 权限变更事件
type PermissionChangedEvent struct {
    AffectedUsers []AffectedUser  // 受影响的用户列表
    Source        string          // 变更来源标识（"assign_resources"/"assign_apis"/"replace_roles" 等）
}

// AffectedUser 受影响的用户
type AffectedUser struct {
    UserID   int64
    TenantID int64
}
```

**事件名常量：** `EventPermissionChanged = "permission.changed"`

#### 1.4 事件发布点

| Service 方法 | 触发时机 | payload 构造 |
|-------------|---------|-------------|
| `RoleService.AssignResources()` | 事务提交成功后 | 查 role_id 关联的 userIDs + tenantID |
| `RoleService.AssignApis()` | 事务提交成功后 | 同上 |
| `RoleService.DeleteRole()` | PERMISSION_SET 删除后 | affectedUserIDs + role.TenantID |
| `UserRoleService.AssignRoles()` | 方法返回 nil 后 | [{userID, req.TenantID}] |
| `UserRoleService.ReplaceRoles()` | 事务回调返回 nil 后 | [{userID, req.TenantID}] |

**关键约束：** 发布必须在事务 Commit 成功后，不得在事务内部发布。

#### 1.5 事件订阅与缓存清除

在 auth-rbac 模块初始化时注册订阅：

```go
event.DefaultBus.Subscribe(event.EventPermissionChanged, func(payload interface{}) {
    ev, ok := payload.(*event.PermissionChangedEvent)
    if !ok { return }
    for _, u := range ev.AffectedUsers {
        permCodeCache.Invalidate(u.TenantID, u.UserID)
    }
})
```

### Phase 2: 操作日志风险分级

#### 2.1 Model 变更

`admin_operation_log` 新增 `risk_level` 字段：

```go
RiskLevel  string `gorm:"type:varchar(16);default:LOW;index:idx_log_risk" json:"risk_level"` // LOW/MEDIUM/HIGH
```

#### 2.2 高风险操作白名单

在 `common/auth/middleware/` 下新增 `risk_level.go`：

```go
// highRiskPaths 高风险操作白名单（method:path 格式）
var highRiskPaths = map[string]bool{
    "DELETE:/api/v1/admin/tenants/:id":      true, // 删除租户
    "PUT:/api/v1/admin/users/:id/status":    true, // 禁用/启用用户
    "PUT:/api/v1/admin/users/:id/reset-pwd": true, // 重置密码
    "DELETE:/api/v1/admin/roles/:id":        true, // 删除角色
    "PUT:/api/v1/admin/roles/:id/resources": true, // 修改角色权限（含级联裁剪）
    "PUT:/api/v1/admin/roles/:id/apis":      true, // 修改角色API权限
    "DELETE:/api/v1/admin/applications/:id": true, // 删除应用
}

func GetRiskLevel(method, fullPath string) string {
    key := method + ":" + fullPath
    if highRiskPaths[key] { return "HIGH" }
    if method != "GET" { return "LOW" }
    return ""  // GET 请求不记录操作日志
}
```

#### 2.3 操作日志记录改造

现有 `OperationLogger` 接口的实现（`common/auth/handler/` 中的 operation_log 记录逻辑）在写入日志时调用 `GetRiskLevel()` 填充 `risk_level` 字段。

#### 2.4 日志查询接口权限控制

操作日志查询 Handler 中加入角色判断：

| 角色 | 查询范围 |
|------|---------|
| SUPER_ADMIN | 全部日志（可选按 tenant_id 过滤） |
| TENANT_ADMIN | WHERE tenant_id = 当前租户 |
| 普通用户 | WHERE tenant_id = 当前租户 AND user_id = 当前用户 |

### Phase 3: OAuth2/LDAP 骨架

#### 3.1 OAuth2Strategy

新增 `common/auth/strategy/oauth2_strategy.go`：

```go
// ErrNotImplemented 新增错误码（501xx 系列）
// 在 common/auth/errors/errors.go 中补充: ErrNotImplemented = 50101

type OAuth2Strategy struct{}

func (s *OAuth2Strategy) GrantType() string { return "oauth2" }

func (s *OAuth2Strategy) Authenticate(c *gin.Context) (*AuthResult, error) {
    return nil, &errors.AuthError{
        Code:    errors.ErrNotImplemented,  // 50101
        Message: "OAuth2 策略暂未对接，请联系管理员",
    }
}
```

#### 3.2 LDAPStrategy

新增 `common/auth/strategy/ldap_strategy.go`：

```go
type LDAPStrategy struct{}

func (s *LDAPStrategy) GrantType() string { return "ldap" }

func (s *LDAPStrategy) Authenticate(c *gin.Context) (*AuthResult, error) {
    return nil, &errors.AuthError{
        Code:    errors.ErrNotImplemented,  // 50101
        Message: "LDAP 策略暂未对接，请联系管理员",
    }
}
```

#### 3.3 注册到 StrategyRouter

在 auth-rbac 初始化流程中增加注册：

```go
router.Register(&strategy.OAuth2Strategy{})
router.Register(&strategy.LDAPStrategy{})
```

### Phase 4: 扩展能力收尾

#### 4.1 ext_fields 字段（4 张表）

| 表 | Model 文件 | 新增字段 |
|----|-----------|---------|
| admin_user | `common/auth/model/user.go` | `ExtFields datatypes.JSON \`gorm:"type:json" json:"ext_fields"\`` |
| admin_tenant | `common/auth/model/tenant.go` | `ExtFields datatypes.JSON \`gorm:"type:json" json:"ext_fields"\`` |
| biz_user | `app/user_auth/model/biz_user.go` | `ExtFields datatypes.JSON \`gorm:"type:json" json:"ext_fields"\`` |
| admin_application | `common/auth/model/application.go` | `ExtFields datatypes.JSON \`gorm:"type:json" json:"ext_fields"\`` |

所有字段 DEFAULT NULL，不影响现有写入。

#### 4.2 admin_custom_field 表

新增 `common/auth/model/custom_field.go`：

```go
type CustomField struct {
    ID         int64          `gorm:"primaryKey" json:"id,string"`
    TenantID   int64          `gorm:"not null;index:idx_cf_tenant_object,priority:1" json:"tenant_id,string"`
    ObjectCode string         `gorm:"type:varchar(64);not null;index:idx_cf_tenant_object,priority:2" json:"object_code"`
    FieldName  string         `gorm:"type:varchar(64);not null" json:"field_name"`
    FieldType  string         `gorm:"type:varchar(32);not null;default:string" json:"field_type"` // string/number/boolean/date/enum
    FieldLabel string         `gorm:"type:varchar(128)" json:"field_label"`
    SortOrder  int            `gorm:"default:0" json:"sort_order"`
    IsRequired bool           `gorm:"default:false" json:"is_required"`
    Options    datatypes.JSON `gorm:"type:json" json:"options"`  // enum 类型的选项列表
    CreatedAt  *time.Time     `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt  *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

func (CustomField) TableName() string { return "admin_custom_field" }
```

仅执行 AutoMigrate 建表，不实现 CRUD Handler。

#### 4.3 安装向导 Redis 可选配置

**SetupRequest 扩展：**

```go
type SetupRequest struct {
    // ... 现有字段 ...

    // Redis 缓存配置（可选）
    RedisAddr     string `json:"redisAddr"`     // Redis 地址，如 127.0.0.1:6379
    RedisPassword string `json:"redisPassword"` // Redis 密码
    RedisDB       int    `json:"redisDB"`       // Redis DB 编号
}
```

**writeConfig 改造：**

```go
func writeConfig(req SetupRequest, _ string) error {
    cfg := map[string]interface{}{
        "settings": map[string]interface{}{
            // ... 现有配置 ...
        },
    }

    // Redis 配置：仅当 RedisAddr 非空时写入
    if req.RedisAddr != "" {
        settings := cfg["settings"].(map[string]interface{})
        settings["cache"] = map[string]interface{}{
            "driver":   "redis",
            "addr":     req.RedisAddr,
            "password": req.RedisPassword,
            "db":       req.RedisDB,
        }
    }
    // 未填写则不写 cache 节，框架默认 memory
    // ...
}
```

### 数据库设计

#### DDL 变更汇总

```sql
-- 1. admin_operation_log 新增 risk_level
ALTER TABLE admin_operation_log ADD COLUMN risk_level VARCHAR(16) DEFAULT 'LOW';
CREATE INDEX idx_log_risk ON admin_operation_log(risk_level);

-- 2. ext_fields 预留（4张表）
ALTER TABLE admin_user ADD COLUMN ext_fields JSON DEFAULT NULL;
ALTER TABLE admin_tenant ADD COLUMN ext_fields JSON DEFAULT NULL;
ALTER TABLE admin_application ADD COLUMN ext_fields JSON DEFAULT NULL;
-- biz_user 在 app/user_auth 域
ALTER TABLE biz_user ADD COLUMN ext_fields JSON DEFAULT NULL;

-- 3. admin_custom_field 新表（由 AutoMigrate 创建）
CREATE TABLE admin_custom_field (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    object_code VARCHAR(64) NOT NULL,
    field_name VARCHAR(64) NOT NULL,
    field_type VARCHAR(32) NOT NULL DEFAULT 'string',
    field_label VARCHAR(128),
    sort_order INT DEFAULT 0,
    is_required TINYINT(1) DEFAULT 0,
    options JSON,
    created_at DATETIME,
    updated_at DATETIME,
    INDEX idx_cf_tenant_object (tenant_id, object_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

所有变更通过 GORM AutoMigrate 自动执行，无需手动 SQL。

### 核心逻辑

#### 缓存读写时序（DynamicPermissionMiddleware）

```
Request → AuthMiddleware(解析JWT) → DynamicPermissionMiddleware
    │
    ├── skipPaths 白名单? → Next()
    ├── C端用户(pool=user)? → Next()
    ├── SUPER_ADMIN? → Next()
    │
    ├── 查 codeMap → permission_code
    │
    ├── permCodeCache.Get(tenantID, userID)
    │   ├── 命中(codes) → PermissionEngine.HasPermission(code)
    │   └── 未命中 ↓
    │
    ├── getUserPermCodes(db, userID, tenantID) → DB 查询
    ├── permCodeCache.Set(tenantID, userID, codes)
    ├── PermissionEngine.HasPermission(code)
    │
    └── 有权限? → Next() / 403
```

#### 事件驱动失效时序

```
RoleService.AssignResources(roleID, req)
    │
    ├── tx.Transaction(替换+级联裁剪)
    │   └── tx.Commit() 成功
    │
    ├── 收集 affectedRoleIDs（当前角色 + 被裁剪子角色）
    ├── 查询 affectedRoleIDs → userIDs（admin_user_role 表）
    │
    └── event.DefaultBus.Publish("permission.changed", &PermissionChangedEvent{
            AffectedUsers: [{userID, tenantID}, ...],
            Source:        "assign_resources",
        })
            │
            └── 订阅方 goroutine:
                for _, u := range ev.AffectedUsers {
                    permCodeCache.Invalidate(u.TenantID, u.UserID)
                }
```

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 现有登录流程（password 策略）不受影响 | /auth/login 正常返回 token |
| RG-2 | 租户隔离不被破坏 | 不同租户数据不互通 |
| RG-3 | DynamicPermissionMiddleware 在 Redis 不可用时降级为直接查 DB | 断开 Redis 后接口仍可访问 |
| RG-4 | SUPER_ADMIN 直接放行逻辑不变 | SUPER_ADMIN 角色用户访问任意接口不被拦截 |
| RG-5 | 现有 PermissionCache（角色级内存缓存）功能不变 | AssignResources 后角色级缓存正常失效 |
| RG-6 | sys_opera_log 原生日志记录不受影响 | 原生中间件继续写入 sys_opera_log |
| RG-7 | 安装向导不填 Redis 时系统正常启动运行 | 不配置 Redis 的 settings.yml 正常引导 |
| RG-8 | 现有 StrategyRouter 的 password/sms 策略不受影响 | grant_type=password/sms 正常工作 |

## 正确性属性

- 缓存失效必须在事务 Commit 成功后触发，失败/回滚时不得失效缓存
- Redis key 包含 tenantID 防止跨租户缓存命中
- PermCodeCache.Get() 对 Redis 连接失败返回 false（降级），不 panic
- OAuth2/LDAP 骨架返回 HTTP 501（Not Implemented），不返回 500
- ext_fields 列不影响现有 INSERT（DEFAULT NULL）
- 高风险操作白名单仅匹配 FullPath（Gin 模板路径），不匹配实际 URL

## API 设计

本 CR 不新增 API 端点。变更点为内部实现（缓存层/事件/字段），对外接口行为不变，仅新增以下查询参数支持：

| 方法 | 路径 | 变更 |
|------|------|------|
| GET | /api/v1/admin/operation-logs | 新增 `risk_level` 查询参数过滤 |

OAuth2/LDAP 骨架注册后，`/auth/login` 的 `grant_type=oauth2` 和 `grant_type=ldap` 将返回 501 而非原来的"不支持的认证类型"错误。
