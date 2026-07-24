# 设计：V2-CR5 双用户池 + C端认证 + 认证策略接口 + TenantIsolationCallback

## 技术方案

### 整体架构

```
┌─ dev-web-admin ─┐     ┌─ dev-web-user ─┐
│  管理端前端      │     │  C端前端        │
└────────┬────────┘     └───────┬────────┘
         │                      │
    /api/v1/admin/...      /api/v1/user/...
         │                      │
┌────────┴──────────────────────┴────────────────────┐
│                   Go 后端（单体）                    │
│                                                     │
│  ┌─ common/auth/strategy/ ─┐  ← 认证基础设施       │
│  │ strategy.go (接口)       │                       │
│  │ claims.go (Claims定义)   │                       │
│  │ token.go (JWT签发/解析)  │                       │
│  │ password.go (管理端策略) │                       │
│  │ router.go (策略路由)     │                       │
│  └──────────────────────────┘                       │
│                                                     │
│  ┌─ app/user_auth/ ─────────┐  ← C端认证域         │
│  │ strategy/sms_strategy.go  │                       │
│  │ handler/ (C端+管理端API) │                       │
│  │ service/ (业务逻辑)       │                       │
│  │ repository/ (数据访问)    │                       │
│  │ model/ (biz_user)        │                       │
│  │ spi/ (SmsSender)         │                       │
│  │ router.go                │                       │
│  └───────────────────────────┘                       │
│                                                     │
│  ┌─ common/auth/middleware/ ─┐                      │
│  │ tenant_isolation_callback │  ← 新增              │
│  │ auth_middleware (改造)     │                       │
│  │ dynamic_permission (改造)  │                       │
│  └───────────────────────────┘                       │
└─────────────────────────────────────────────────────┘
```

### API 设计

#### C端认证接口（`/api/v1/user/auth/`）

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | /api/v1/user/auth/send-code | 发送短信验证码 | 无 |

**send-code 请求/响应示例：**
```json
// 请求
POST /api/v1/user/auth/send-code
{ "phone": "13800138000", "tenant_code": "abc" }

// 成功响应
{ "code": 200, "data": { "expires_in": 300 }, "message": "验证码已发送" }

// 限频响应 (429)
{ "code": 42901, "data": { "retry_after": 45 }, "message": "发送过于频繁，请45秒后重试" }
```
| POST | /api/v1/user/auth/login | C端短信登录（自动注册） | 无 |
| POST | /api/v1/user/auth/logout | C端登出 | JWT(user) |
| GET | /api/v1/user/menu | C端获取菜单 | JWT(user) |

#### 管理端 biz_user 管理接口（`/api/v1/admin/biz-users/`）

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/biz-users | biz_user 列表（分页） | JWT(admin) + 权限 |
| GET | /api/v1/admin/biz-users/:id | biz_user 详情 | JWT(admin) + 权限 |
| POST | /api/v1/admin/biz-users | 创建 biz_user | JWT(admin) + 权限 |
| PUT | /api/v1/admin/biz-users/:id | 更新 biz_user | JWT(admin) + 权限 |
| DELETE | /api/v1/admin/biz-users/:id | 删除 biz_user | JWT(admin) + 权限 |
| POST | /api/v1/admin/biz-users/:id/reset-password | 重置密码 | JWT(admin) + 权限 |
| POST | /api/v1/admin/biz-users/:id/force-logout | 强制登出 | JWT(admin) + 权限 |
| POST | /api/v1/admin/biz-users/:id/toggle-status | 启用/禁用 | JWT(admin) + 权限 |

#### 管理端登录接口改造

| 方法 | 路径 | 描述 | 变更 |
|------|------|------|------|
| POST | /auth/login | 统一登录入口 | 新增 grant_type 字段路由 |

### 数据库设计

```sql
-- biz_user 表（C端用户池）
CREATE TABLE biz_user (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 0,
    phone VARCHAR(20) NOT NULL,
    password VARCHAR(128) DEFAULT '',
    nickname VARCHAR(64) DEFAULT '',
    avatar VARCHAR(256) DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    token_version INT NOT NULL DEFAULT 1,
    last_login_at TIMESTAMP NULL,
    last_login_ip VARCHAR(45) DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    create_by BIGINT DEFAULT 0,
    update_by BIGINT DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    UNIQUE KEY uk_tenant_phone (tenant_id, phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_biz_user_tenant ON biz_user(tenant_id);
CREATE INDEX idx_biz_user_phone ON biz_user(phone);
CREATE INDEX idx_biz_user_deleted ON biz_user(deleted_at);
```

### 核心逻辑

#### 1. 认证策略接口（common/auth/strategy/strategy.go）

```go
package strategy

import "github.com/gin-gonic/gin"

// AuthResult 认证结果
type AuthResult struct {
    UserID      int64
    TenantID    int64
    UserPool    string // "admin" | "user"
    Roles       []int64
}

// AuthenticationStrategy 认证策略接口
type AuthenticationStrategy interface {
    // GrantType 返回该策略对应的 grant_type 值
    GrantType() string
    // Authenticate 执行认证，返回认证结果
    Authenticate(c *gin.Context) (*AuthResult, error)
}
```

#### 2. Claims 扩展（common/auth/strategy/claims.go）

```go
// AccessClaims 新增 UserPool 字段
type AccessClaims struct {
    jwt.RegisteredClaims
    UserID       int64   `json:"user_id"`
    TenantID     int64   `json:"tenant_id"`
    Roles        []int64 `json:"roles"`
    UserPool     string  `json:"user_pool"`      // "admin" | "user"
    TokenVersion int     `json:"token_version"`  // 强制登出用
}

const (
    UserPoolAdmin = "admin"
    UserPoolUser  = "user"
)
```

#### 3. JWT 工具函数（common/auth/strategy/token.go）

从 `auth_service.go` 抽出为独立函数，使用 Options struct 模式：

```go
// AccessTokenOptions 签发 access_token 的参数
type AccessTokenOptions struct {
    UserID       int64
    TenantID     int64
    Roles        []int64
    UserPool     string // "admin" | "user"
    TokenVersion int
}

func GeneratePlatformToken(cfg *config.Config, userID int64, tenants []TenantInfo) (string, error)
func GenerateAccessToken(cfg *config.Config, opts *AccessTokenOptions) (string, error)
func ParsePlatformToken(cfg *config.Config, tokenStr string) (*PlatformClaims, error)
func ParseAccessToken(cfg *config.Config, tokenStr string) (*AccessClaims, error)
```

现有 `AuthService` 的 Login/SelectTenant/Refresh 改为调用这些公共函数。

#### 4. PasswordStrategy（common/auth/strategy/password.go）

封装现有管理端登录逻辑：
- 查 admin_user 表
- 验证 bcrypt 密码
- 查关联租户
- 单租户直接签发 access_token，多租户返回 platform_token
- UserPool = "admin"

行为与现有 `AuthService.Login` 完全一致。

#### 5. SmsStrategy（app/user_auth/strategy/sms_strategy.go）

```
请求体: { "phone": "13800138000", "code": "1234", "tenant_code": "abc" }

流程:
1. 验证验证码（从内存 store 校验，4位/5分钟有效）
2. 通过 tenant_code 查 tenant_id
3. 按 (tenant_id, phone) 查 biz_user
   - 存在 → 直接签发 access_token
   - 不存在 → 创建 biz_user → 签发 access_token
4. 签发时 UserPool = "user"，TTL 使用 cfg.JWT.UserAccessTokenTTL
5. 不走 platform_token 阶段（C端单租户绑定）
```

**验证码存储演进说明**：本 CR 使用内存 store（sync.Map），仅适用于单实例部署。生产环境多实例部署时需替换为 Redis store（共享验证码状态），替换点为 `app/user_auth/service/sms_service.go` 中的 `CodeStore` 接口实现。

#### 6. 强制登出设计（token_version 方案）

**原理**：biz_user 表有 `token_version` 字段。签发 token 时将当前 version 写入 Claims。AuthMiddleware 解析 token 后，对 UserPool="user" 的请求额外查询 biz_user.token_version，不匹配则拒绝。

**管理端 vs C端登出机制区分**：
- UserPool="admin" → 沿用现有黑名单机制（`AuthService.Logout` 加 JTI 到黑名单）
- UserPool="user" → token_version 机制（无黑名单存储开销）

**强制登出操作**：`UPDATE biz_user SET token_version = token_version + 1 WHERE id = ?`

**缓存策略**：
- token_version 按 user_id 缓存（TwoLevelCache，TTL 5 分钟）
- 强制登出时**主动失效**对应 user_id 的缓存条目
- 缓存 miss 时**必须 fallback 查库**，不得返回默认"通过"
- 查库失败（DB 不可达）时拒绝请求（fail-closed）

#### 7. TenantIsolationCallback（common/auth/middleware/tenant_isolation_callback.go）

```go
// 注册四个 Callback：Create/Query/Update/Delete
// 顺序保证：tenant_isolation → data_scope → gorm:query
func RegisterTenantIsolationCallback(db *gorm.DB) {
    db.Callback().Create().Before("gorm:create").Register("auth:tenant_create", tenantCreateCallback)
    db.Callback().Query().Before("auth:data_scope").Register("auth:tenant_query", tenantQueryCallback)
    db.Callback().Update().Before("gorm:update").Register("auth:tenant_update", tenantUpdateCallback)
    db.Callback().Delete().Before("gorm:delete").Register("auth:tenant_delete", tenantDeleteCallback)
}

// DataScopeCallback 注册为 Before("gorm:query")
// 最终执行链：auth:tenant_query → auth:data_scope → gorm:query
```

**检测逻辑**（每个 Callback 共用）：
```go
func hasTenantField(db *gorm.DB) bool {
    if db.Statement.Schema == nil {
        return false
    }
    return db.Statement.Schema.LookUpField("TenantID") != nil
}
```

**注入规则**：
- 从 context 获取 tenant_id（`GetAuthContext(ctx).TenantID`）
- tenant_id == 0 → 跳过（超管不过滤）
- Create → `db.Statement.SetColumn("tenant_id", tenantID)`
- Query/Update/Delete → `db.Where("tenant_id = ?", tenantID)`
- SystemOp 标记 → 跳过（与 DataScopeCallback 共用机制）

**注册顺序保证**：Query Callback 命名为 `auth:tenant_query`，DataScopeCallback 为 `auth:data_scope`，通过 `Before("auth:data_scope")` 确保 TenantIsolation 先执行。

#### 8. DynamicPermissionMiddleware 改造

```go
// 在白名单检查后、权限检查前新增：
authCtx := GetAuthContext(c)
if authCtx != nil && authCtx.UserPool == UserPoolUser {
    c.Next()
    return
}
```

AuthContext 结构扩展：
```go
type AuthContext struct {
    UserID       int64
    TenantID     int64
    Roles        []int64
    UserPool     string  // 新增
    TokenVersion int     // 新增
}
```

#### 9. C端 GetUserMenu 缓存设计

- 缓存 key：`user_menu:{tenant_id}`
- 缓存存储：TwoLevelCache（本地 + 可选 Redis）
- TTL：10 分钟
- 失效时机（均按 tenant_id 精确失效，仅影响对应租户的缓存 key）：
  - 租户订阅变更（app_subscription 变更时）→ `InvalidateUserMenuCache(tenantID)`
  - 资源树变更（admin_resource 变更时）→ `InvalidateUserMenuCache(tenantID)`
  - 管理员手动触发

#### 10. 配置扩展（common/auth/config/config.go）

```go
type JWTConfig struct {
    Secret          string        `yaml:"secret"`
    Issuer          string        `yaml:"issuer"`
    AccessTokenTTL  time.Duration `yaml:"accessTokenTTL"`   // 管理端，默认 2h
    PlatformTokenTTL time.Duration `yaml:"platformTokenTTL"` // 默认 10min
    UserAccessTokenTTL time.Duration `yaml:"userAccessTokenTTL"` // C端，默认 7d
}
```

### 模块依赖图

```
app/user_auth/ ──→ common/auth/strategy/  (接口/Claims/JWT工具)
app/user_auth/ ──→ common/auth/config/    (配置)
app/user_auth/ ──→ common/auth/cache/     (TwoLevelCache)
app/user_auth/ ✗── common/auth/service/   (不依赖)
app/user_auth/ ✗── common/auth/handler/   (不依赖)
app/user_auth/ ✗── common/auth/model/     (不依赖，有自己的 model)

common/auth/strategy/ ──→ common/auth/config/
common/auth/strategy/ ✗── common/auth/service/  (反向不依赖)

common/auth/service/ ──→ common/auth/strategy/  (调用 token.go 公共函数)
```

**拆分单元**：`app/user_auth/` + `common/auth/strategy/` + `common/auth/config/` + `common/auth/cache/`

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 管理端 /auth/login 不传 grant_type 时行为完全不变 | 现有登录流程正常返回 token |
| RG-2 | 管理端 /auth/login 传 grant_type=password 行为与不传一致 | 对比响应结构和 token 内容 |
| RG-3 | SelectTenant 流程不受影响 | platform_token 选租户正常签发 access_token |
| RG-4 | Refresh 流程不受影响 | platform_token 刷新正常工作 |
| RG-5 | Logout 黑名单机制不受影响 | 管理端 logout 后 token 失效 |
| RG-6 | DynamicPermissionMiddleware 对管理端用户行为不变 | 管理端权限检查正常 |
| RG-7 | DataScopeCallback 数据权限不受影响 | 数据范围过滤正常 |
| RG-8 | 现有业务表查询结果不变 | TenantIsolationCallback 不影响已有 tenant_id 过滤逻辑 |
| RG-9 | admin_user 等全局表不被 TenantIsolationCallback 影响 | 无 tenant_id 字段的表查询无变化 |
| RG-10 | 现有单元测试编译通过且行为不变 | `go test ./common/auth/...` 全部 PASS |

## 正确性属性

- C端 token 中 UserPool 必须为 "user"，管理端必须为 "admin" 或 ""
- biz_user 表所有记录的 tenant_id 必须 > 0（C端不存在无租户用户）
- TenantIsolationCallback 的 Create 操作仅在 tenant_id > 0 时填充
- 同一租户内 biz_user.phone 唯一
- token_version 只递增不递减
- SmsStrategy 验证码 4 位数字、5 分钟过期、60 秒限频
- PasswordStrategy 执行路径与重构前 AuthService.Login 输出完全一致（输入相同 → 输出相同）
