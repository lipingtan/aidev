# 设计：V2-CR6 域名-租户映射管理模块

## 技术方案

### 整体架构

```
┌─ dev-web-user 登录页 ─────────────────────────────────────────┐
│  onMounted → GET /api/v1/public/tenant-domain?domain=hostname │
│  成功 → 存入 tenantCode 状态；失败 → fallback VITE_TENANT_CODE │
└───────────────────────────────────────────────────────────────┘
                          │
           ┌──────────────┴──────────────────────────────┐
           │ 公开路由（无 AuthMiddleware）                  │
           │ /api/v1/public/                              │
           └──────────────┬──────────────────────────────┘
                          │
              TenantDomainHandler.QueryByDomain
                          │
              ┌───────────▼──────────────┐
              │   LocalCache 查询         │
              │   key: tenant_domain:xxx  │
              │   TTL: 5min              │
              └───────────┬──────────────┘
                    hit   │   miss
               ┌──────────┴────────────────┐
               │                           │
           返回缓存值          TenantDomainService.QueryByDomain
                                    │
                           精确匹配 tenant_domain 表
                                    │
                           未找到 → 查 admin_tenant WHERE tenant_code='default'
                                    │
                           结果写入缓存
                                    │
                           返回 { tenant_code, tenant_name }

┌─ dev-web-admin 域名管理页 ─────────────────────────────────────┐
│  CRUD → /api/v1/admin/tenant-domains                          │
│  认证 + DynamicPermissionMiddleware                           │
│  permission_code: system:tenant-domain:list/create/update/delete │
└───────────────────────────────────────────────────────────────┘
                          │
              TenantDomainHandler（CRUD）
                          │
              TenantDomainService
                  写操作完成后 → cache.Delete("tenant_domain:" + domain)
                          │
              tenant_domain 表（MySQL）
```

---

### API 设计

#### 公开接口

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/public/tenant-domain | 按域名查询租户信息 | 无 |

**查询参数**：`domain` (string, required) — 当前访问的 hostname，例如 `abc.example.com`

**成功响应**（无论是否找到都返回 200）：
```json
{
  "code": 0,
  "data": {
    "tenant_code": "abc",
    "tenant_name": "ABC 企业",
    "matched": true
  },
  "message": "ok"
}
```

**兜底响应**（域名未配置，返回 default 租户）：
```json
{
  "code": 0,
  "data": {
    "tenant_code": "default",
    "tenant_name": "默认租户",
    "matched": false
  },
  "message": "ok"
}
```

**`domain` 缺失时**：返回 400

#### 管理端 CRUD 接口

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/tenant-domains | 分页列表 | JWT + 权限 |
| POST | /api/v1/admin/tenant-domains | 创建 | JWT + 权限 |
| PUT | /api/v1/admin/tenant-domains/:id | 更新 | JWT + 权限 |
| DELETE | /api/v1/admin/tenant-domains/:id | 删除（软删除） | JWT + 权限 |

**列表查询参数**：`page`, `page_size`, `domain`（模糊）, `tenant_id`（精确筛选）

**创建/更新请求体**：
```json
{
  "domain": "abc.example.com",
  "tenant_id": "2079132216964681728",
  "remark": "ABC 企业专属入口"
}
```
更新时需额外传 `"version": 1`（乐观锁）

**列表响应**：
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "xxx",
        "domain": "abc.example.com",
        "match_type": "EXACT",
        "tenant_id": "xxx",
        "tenant_code": "abc",
        "tenant_name": "ABC 企业",
        "remark": "",
        "version": 1,
        "created_at": "2026-07-24T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 数据库设计

```sql
CREATE TABLE tenant_domain (
    id          BIGINT       NOT NULL PRIMARY KEY,   -- 雪花 ID
    domain      VARCHAR(255) NOT NULL,               -- 域名，如 abc.example.com
    match_type  VARCHAR(16)  NOT NULL DEFAULT 'EXACT', -- EXACT | WILDCARD（预留）
    tenant_id   BIGINT       NOT NULL,               -- 关联 admin_tenant.id
    remark      VARCHAR(256) NOT NULL DEFAULT '',    -- 备注
    version     INT          NOT NULL DEFAULT 1,     -- 乐观锁
    created_at  DATETIME(3)  NOT NULL,
    updated_at  DATETIME(3)  NOT NULL,
    deleted_at  DATETIME(3)  NULL,
    -- 注意：不使用 (domain, deleted_at) 联合唯一索引
    -- MySQL 中 NULL 不参与唯一约束比较，无法防止软删除记录+活跃记录并存
    -- 唯一性由 Service 层业务校验保证（Create/Update 前先查活跃记录）
    INDEX idx_domain (domain),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**关联关系**：`tenant_domain.tenant_id → admin_tenant.id`（逻辑外键，不建物理 FK）

**唯一性保障**：MySQL 中 NULL 值不参与唯一约束比较，`(domain, deleted_at)` 联合唯一索引无法真正防重。改为普通索引 + Service 层在 Create/Update 时先查询 `WHERE domain=? AND deleted_at IS NULL` 确认无活跃记录，存在则返回 409。

---

### 核心数据结构

#### model/tenant_domain.go

```go
type TenantDomain struct {
    ID        int64          `json:"id,string" gorm:"primaryKey;autoIncrement:false"`
    Domain    string         `json:"domain" gorm:"size:255;not null;index:idx_domain"`  // 普通索引，唯一性由 Service 层业务校验
    MatchType string         `json:"match_type" gorm:"size:16;not null;default:EXACT"` // EXACT | WILDCARD
    TenantID  int64          `json:"tenant_id,string" gorm:"not null;index:idx_tenant_id"`
    Remark    string         `json:"remark" gorm:"size:256;default:''"`
    Version   int            `json:"version" gorm:"not null;default:1"`
    CreatedAt *time.Time     `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt *time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
func (TenantDomain) TableName() string { return "tenant_domain" }
```

#### service DTO

```go
// QueryDomainResult 域名查询结果
type QueryDomainResult struct {
    TenantCode string `json:"tenant_code"`
    TenantName string `json:"tenant_name"`
    Matched    bool   `json:"matched"` // false=兜底 default 租户
}

// CreateTenantDomainRequest
type CreateTenantDomainRequest struct {
    Domain   string `json:"domain" binding:"required"`
    TenantID int64  `json:"tenant_id,string" binding:"required"`
    Remark   string `json:"remark"`
}

// UpdateTenantDomainRequest
// Domain 和 TenantID 使用指针类型，区分 nil（未传，不更新）和零值
type UpdateTenantDomainRequest struct {
    Domain   *string `json:"domain"`             // nil=不更新
    TenantID *int64  `json:"tenant_id,string"`   // nil=不更新
    Remark   *string `json:"remark"`             // nil=不更新
    Version  int     `json:"version" binding:"required"` // 乐观锁，必传
}
```

---

### 核心逻辑

#### QueryByDomain 查询流程

```
func (s *TenantDomainService) QueryByDomain(domain string) (*QueryDomainResult, error)

1. 参数校验：domain 不能为空
2. 查缓存：key = "tenant_domain:" + domain
   命中 → 直接返回
3. 缓存 miss → 两次独立 DB 查询（遵循 Repository 单表原则）：
   3a. TenantDomainRepository.FindByDomain(domain)
       → SELECT * FROM tenant_domain
         WHERE domain = ? AND deleted_at IS NULL AND match_type = 'EXACT'
         LIMIT 1
   3b. 找到记录 → 用 tenantID 查 TenantRepository.FindByID(tenantID)
       → SELECT * FROM admin_tenant WHERE id = ? AND deleted_at IS NULL
4. 找到 → result = {TenantCode: tenant.TenantCode, TenantName: tenant.Name}，result.Matched = true
   未找到（step 3a 无记录，或 step 3b 租户已删除）：
     → TenantRepository.FindByCode("default")
     → 找到：result.Matched = false，使用 DB 中的 default 租户信息
     → 未找到（default 租户被删除）：硬编码 {TenantCode:"default", TenantName:"默认租户"}，result.Matched = false
5. 结果写入缓存（TTL 5min）
6. 返回 result
```

**注意**：QueryByDomain 不需要 gin.Context，直接传 domain string，不走 TenantIsolationCallback。`tenant_domain` 表不含 `tenant_id` 字段，TenantIsolationCallback 使用 `db.Statement.Schema.LookUpField("TenantID")` 检测，找不到该字段则自动跳过注入，无需额外处理（RG-3 的原理）。

#### 缓存失效策略

```
// 所有 cache.Delete 必须在 DB 写操作确认成功后才执行，写失败不失效缓存
Create(domain)  → DB 写入成功后 → cache.Delete("tenant_domain:" + domain)
Update(id, req) → 先查旧 domain，DB 更新成功后 → cache.Delete("tenant_domain:" + oldDomain)
                   若 domain 变更，还需 → cache.Delete("tenant_domain:" + *req.Domain)
Delete(id)      → 先查 domain，DB 软删除成功后 → cache.Delete("tenant_domain:" + domain)
```

#### SUPER_ADMIN 权限双重校验

```go
// Service 层在 Create/Update/Delete 前显式校验，不完全依赖 permission_code
// 防止租户管理员因被错误赋予 permission_code 而跨租户操作域名
func (s *TenantDomainService) assertSuperAdmin(authCtx *middleware.AuthContext) error {
    // 查询当前用户的角色，验证至少有一个 SUPER_ADMIN role_type
    var count int64
    s.db.Model(&model.UserRole{}).
        Joins("JOIN admin_role r ON r.id = user_roles.role_id").
        Where("user_roles.user_id = ? AND r.role_type = 'SUPER_ADMIN'", authCtx.UserID).
        Count(&count)
    if count == 0 {
        return errors.NewAuthError(errors.ErrPermissionDenied, "仅超级管理员可操作域名绑定")
    }
    return nil
}
```

#### localhost/127.0.0.1 保护

```go
var reservedDomains = map[string]bool{
    "localhost": true,
    "127.0.0.1": true,
    "0.0.0.0":   true,
}

// Create/Update 时校验（顺序：1.长度 2.保留域名 3.业务唯一性）
func validateDomain(domain string) error {
    if len(domain) > 255 {
        return errors.NewAuthError(errors.ErrInvalidInput, "域名长度超限，最大 255 字符")
    }
    if reservedDomains[domain] {
        return errors.NewAuthError(errors.ErrInvalidInput, "保留域名不允许绑定")
    }
    return nil
}
```

---

### 路由注册

在 `common/auth/router.go` 的 `RegisterRoutes` 函数中，新增两个路由组：

```go
// 公开路由（无任何中间件）
public := rg.Group("/api/v1/public")
{
    public.GET("/tenant-domain", deps.TenantDomainHandler.QueryByDomain)
}

// 管理端路由（已在 admin Group 内，复用已有中间件链）
// 在 admin Group 的花括号内添加：
if deps.TenantDomainHandler != nil {
    deps.TenantDomainHandler.RegisterRoutes(admin)
}
```

---

### 前端改造：dev-web-user 登录页

**vite.config.ts proxy 补充**（修复 #6：/api/v1/public 未代理）：
```typescript
proxy: {
  '/api': {
    target: 'http://localhost:8000',
    changeOrigin: true
  }
  // /api 前缀已覆盖 /api/v1/public，无需单独配置
}
```
> 注意：现有 proxy 规则 `'/api'` 已能匹配 `/api/v1/public/...`，Vite 会正确转发，无需额外配置。

```typescript
// src/utils/tenant.ts — 使用封装的 service 实例，利用已配置的 proxy
import service from './request'  // 复用 request.ts 的 axios 实例（含 baseURL=/api/v1/user）

export async function resolveTenantCode(): Promise<string> {
  const hostname = window.location.hostname
  try {
    // 注意：此接口不在 /api/v1/user 下，需使用完整路径覆盖 baseURL
    const resp = await service.get('/api/v1/public/tenant-domain', {
      params: { domain: hostname },
      baseURL: '',   // 覆盖 service 的 baseURL，使用相对路径（走 Vite proxy）
      timeout: 3000
    })
    if (resp && (resp as any).tenant_code) {
      return (resp as any).tenant_code
    }
  } catch {
    // 超时或网络异常，fallback
  }
  return import.meta.env.VITE_TENANT_CODE || 'default'
}

```typescript
// login/index.vue
const tenantCode = ref<string>('')
const resolvingTenant = ref(true)  // 解析中，禁用发送按钮

onMounted(async () => {
  tenantCode.value = await resolveTenantCode()
  resolvingTenant.value = false
})
```

发送验证码按钮绑定：
```html
<el-button
  :disabled="countdown > 0 || sendingCode || resolvingTenant"
>
  {{ resolvingTenant ? '加载中…' : (countdown > 0 ? `${countdown}s` : '获取验证码') }}
</el-button>
```

---

### 前端改造：dev-web-admin 域名管理页

**路由**：`system/tenant-domain`（加入 static-routes.ts）

**页面结构**：
- 搜索栏：域名（模糊）+ 租户下拉筛选
- 表格列：域名 | 绑定租户 | 备注 | 创建时间 | 操作（编辑/删除）
- 新增/编辑弹窗：域名输入框 + 租户 el-select（从 /api/v1/admin/tenants 加载）
  - **编辑时 `version` 以 hidden field 形式存在**（从列表接口返回的 version 读取，submit 时携带，不显示给用户）
- 删除：二次确认，确认文案含域名

**API 模块**：`src/api/tenant-domain.ts`

---

### Seed 变更

在 `seed.go` 新增 `SeedCR6Menus` 函数（幂等）：

```go
// SeedCR6Menus 补充 CR-6 域名管理菜单
func SeedCR6Menus(db *gorm.DB) error {
    // path = /system/tenant-domain
    // parent: /system
    // permission_code: system:tenant-domain:list
}
```

在 `auth.go` 的 `Init` 中调用。

---

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 管理端登录流程不受影响 | /auth/login 正常返回 token |
| RG-2 | 现有 C端认证流程不受影响 | /api/v1/user/auth/login 正常工作 |
| RG-3 | TenantIsolationCallback 不影响 tenant_domain 查询 | `tenant_domain` 表不含 `tenant_id` 字段，GORM Callback 通过 `db.Statement.Schema.LookUpField("TenantID")` 检测，找不到字段则自动跳过注入，无需额外配置 |
| RG-4 | 其他公开接口（send-code/login）不受新路由注册影响 | /api/v1/user/auth/send-code 正常返回 |
| RG-5 | 现有 LocalCache 使用者（permission 缓存等）不受影响 | TenantDomain 使用独立 LocalCache 实例，缓存 key 前缀 `tenant_domain:` 不与其他模块冲突 |

---

## 正确性属性

- `domain` 活跃记录中全局唯一（Service 层业务校验，非 DB 唯一索引）
- `localhost` / `127.0.0.1` / `0.0.0.0` 不允许被绑定（validateDomain 校验）
- 域名长度不超过 255 字符（validateDomain 校验）
- 公开查询接口在任何情况下都返回 200（未找到时返回 default 租户；default 不存在时硬编码兜底，永不 500）
- 缓存失效在 DB 写操作**成功后**执行，写失败不触发缓存失效
- `tenant_id` 必须引用实际存在的 `admin_tenant` 记录（Service 层校验，不依赖 DB 外键）
- `match_type` 当前只接受 EXACT（Service 层硬编码，预留 WILDCARD 枚举值但不处理）
- CRUD 操作必须通过 Service 层 `assertSuperAdmin` 校验，不完全依赖 permission_code
- UpdateRequest 中指针类型字段：nil 表示不更新，非 nil 表示更新为新值
