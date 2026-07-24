# 任务列表：V2-CR6 域名-租户映射管理模块

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 9 |
| 已完成 | 9 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 9/9 (100%) |
| 当前阶段 | 全部完成 |

---

## 任务依赖图

```
Task 1（Model + DDL）
    │
Task 2（Repository）
    │
Task 3（Service：查询 + 缓存）── Task 4（Service：CRUD + 校验）
    │                               │
Task 5（Handler + 路由注册）────────┘
    │
Task 6（Seed 菜单 + auth.go 接入）
    │
    ├── Task 7（dev-web-admin：域名管理页）
    │
    └── Task 8（dev-web-user：登录页改造）
            │
        Task 9（vite.config.ts 代理 + utils/tenant.ts）
```

Task 7、8 可在 Task 6 完成后并行执行（无文件交叉）。

---

## Phase 1：后端基础层

### Task 1: TenantDomain Model + DDL ✅

**复杂度**: 高（新增表 + AutoMigrate）

**Scope（边界）:**
- 涉及文件:
  - `common/auth/model/tenant_domain.go`（新建）
  - `common/auth/auth.go`（修改 autoMigrate 列表）
- 不触碰: 其他 model 文件、现有迁移逻辑

**Constraints（约束）:**
- `Domain` 字段使用普通索引（`index:idx_domain`），不使用唯一索引——唯一性由 Service 层业务校验
- `DeletedAt` 使用 `gorm.DeletedAt` 类型，配合软删除
- 雪花 ID：`BeforeCreate` 钩子 + `autoIncrement:false`
- `json:",string"` tag：`ID`、`TenantID` 两个 int64 字段必须加
- `match_type` 默认值 `EXACT`，预留 `WILDCARD` 枚举值

**Acceptance（验证标准）:**
- AC: `tenant_domain` 表结构包含所有设计字段（id/domain/match_type/tenant_id/remark/version/created_at/updated_at/deleted_at）
- AC: GORM tag 与 DDL 设计完全一致（普通索引、无唯一约束）
- AC: AutoMigrate 执行后 `tenant_domain` 表在 DB 中正确创建
- AC: `go build ./common/auth/...` 零错误
- AC: 【回归】现有表结构不受影响（RG-1/RG-2）

**自测:**
- 测试文件: `common/auth/model/tenant_domain_test.go`
- ST: TableName() → 返回 `"tenant_domain"`
- ST: BeforeCreate 未设 ID → 自动生成非零雪花 ID
- ST: 字段 json tag 检查 → ID/TenantID 含 `,string`

---

### Task 2: TenantDomainRepository ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `common/auth/repository/tenant_domain_repo.go`（新建）
- 不触碰: 其他 repository 文件

**Constraints（约束）:**
- Repository 只操作单表 `tenant_domain`，不做跨表 JOIN——遵循分层边界
- `FindByDomain` 只返回 `TenantDomain` 记录（含 tenant_id），不 JOIN admin_tenant
- tenant_code / tenant_name 的查询由 Service 层通过 TenantRepository 补充
- 定义专用查询参数 struct `TenantDomainListParams`（page/page_size/domain/tenant_id）

**Acceptance（验证标准）:**
- AC: `TenantDomainRepository` 接口定义：`FindByDomain`, `FindByID`, `Create`, `Update`, `SoftDelete`, `List`
- AC: `FindByDomain` 过滤 `deleted_at IS NULL`，只返回 `TenantDomain` 结构（含 tenant_id），不 JOIN 其他表
- AC: `List` 支持 domain 模糊搜索、tenant_id 精确筛选、分页
- AC: `go build ./common/auth/repository/...` 零错误

**自测:**
- 测试文件: `common/auth/repository/tenant_domain_repo_test.go`（使用 SQLite 内存 DB）
- ST: FindByDomain 存在且未删除 → 返回正确 TenantDomain 记录
- ST: FindByDomain 不存在 → 返回 nil, nil
- ST: FindByDomain 已软删除 → 返回 nil, nil
- ST: List domain 模糊搜索 → 只返回匹配域名的记录
- ST: SoftDelete → 记录 deleted_at 非空，FindByDomain 找不到

---

### Task 3: TenantDomainService — 查询与缓存 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `common/auth/service/tenant_domain_service.go`（新建，先实现查询部分）
- 不触碰: 其他 service 文件、LocalCache 实现

**Constraints（约束）:**
- 使用独立的 `*cache.LocalCache` 实例（不复用其他模块的缓存）
- 缓存 key 格式：`tenant_domain:{domain}`
- TTL 固定 5 分钟
- 缓存失效在写操作成功后显式调用（不用 defer）
- `QueryByDomain` 不依赖 gin.Context，直接接收 `domain string`
- Service 层查询到 TenantDomain 记录后，再调用 TenantRepository 查 tenant_code/name（不依赖 Repository JOIN）
- default 租户不存在时硬编码兜底：`{TenantCode:"default", TenantName:"默认租户"}`
- **前置依赖**：`TenantDomainRepository` 接口必须由 Task 2 先行定义，Service mock 依赖该接口编写

**Acceptance（验证标准）:**
- AC: `QueryByDomain` 命中缓存 → 不查 DB，直接返回
- AC: `QueryByDomain` 缓存 miss 精确匹配 → Service 查 TenantRepository 补充 tenant_code/name，`Matched=true`，结果写入缓存
- AC: `QueryByDomain` 未配置域名 → 查 default 租户，`Matched=false`，结果写入缓存
- AC: `QueryByDomain` domain 为空 → 返回 error
- AC: default 租户不存在时 → 硬编码兜底，不返回 error
- AC: `go build ./common/auth/service/...` 零错误

**自测:**
- 测试文件: `common/auth/service/tenant_domain_service_test.go`（mock repository）
- ST: 缓存命中 → repo 未被调用
- ST: 精确匹配 → Matched=true，结果一致
- ST: 无匹配 → Matched=false，TenantCode="default"
- ST: domain 为空 → 返回 error
- ST: default 租户查不到 → 返回硬编码 default，不 panic

---

### Task 4: TenantDomainService — CRUD + 校验 ✅

**依赖**: Task 3（同文件，续写）

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `common/auth/service/tenant_domain_service.go`（续写 CRUD 方法）
- 不触碰: Task 3 已实现的查询方法

**Constraints（约束）:**
- `Create`/`Update` 前必须调用 `validateDomain`（长度 ≤255 + 保留域名检查）
- `Create`/`Update` 前必须调用 `assertSuperAdmin(userID int64)`
  - 入参为 `userID int64`，由 Handler 从 `middleware.GetAuthContext(c).UserID` 提取后传入
  - Service 层不依赖 gin.Context，保持分层纯洁
- `Create` 前查活跃记录唯一性：`WHERE domain=? AND deleted_at IS NULL`，存在则返回 `errDuplicateDomain`
- `Update` 使用指针类型字段：nil 不更新，非 nil 更新；乐观锁 version 校验
- 写操作成功后显式调用 `cache.Delete("tenant_domain:" + domain)`
- 若 domain 变更，同时失效旧 domain 和新 domain 的缓存
- `tenant_id` 必须引用存在的 admin_tenant（Service 层查库校验）

**Acceptance（验证标准）:**
- AC: Create 保留域名 → 返回 `errInvalidInput` 类型 error（HTTP 状态码由 Task 5 Handler 验证）
- AC: Create 超长域名（>255） → 返回 `errInvalidInput` 类型 error
- AC: Create 重复活跃域名 → 返回 `errDuplicateDomain` 类型 error
- AC: Create 成功 → DB 记录存在，缓存失效
- AC: Update version 不匹配 → 返回 version conflict error
- AC: Update domain 变更 → 旧新 domain 缓存均失效
- AC: Delete 成功 → 软删除，缓存失效
- AC: 非 SUPER_ADMIN userID 调用 → 返回 `errPermissionDenied` 类型 error
- AC: `go build ./common/auth/service/...` 零错误

**自测:**
- 测试文件: `common/auth/service/tenant_domain_service_test.go`（续写）
- ST: Create localhost → error 类型为 errInvalidInput，message 含"保留域名"
- ST: Create domain >255 字符 → error 类型为 errInvalidInput，message 含"超限"
- ST: Create 重复 → error 类型为 errDuplicateDomain
- ST: Create 成功 → repo.Create 被调用，缓存被失效
- ST: Update version 错误 → error 非 nil
- ST: Update 变更 domain → 两个 cache key 均被 Delete
- ST: Delete → repo.SoftDelete 被调用，缓存被失效
- ST: 非 SUPER_ADMIN userID → error 类型为 errPermissionDenied

---

### Task 5: TenantDomainHandler + 路由注册 ✅

**依赖**: Task 3, Task 4

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `common/auth/handler/tenant_domain_handler.go`（新建）
  - `common/auth/router.go`（修改：注册 public 路由组、admin CRUD 路由，Dependencies struct 增加 TenantDomainHandler 字段）
- 不触碰: 其他 handler 文件、现有路由注册逻辑

**Constraints（约束）:**
- 公开路由 `GET /api/v1/public/tenant-domain` 必须在不带任何中间件的 Group 中注册（在 `RegisterRoutes` 开头注册）
- 管理端 CRUD 路由 `/api/v1/admin/tenant-domains` 注册在已有 `admin` Group 内（复用认证+权限中间件链）
- `domain` query 参数长度 Handler 层校验（≤255），超长返回 400
- Handler 调用 CRUD 方法时，从 `middleware.GetAuthContext(c).UserID` 提取 userID 传给 Service 的 `assertSuperAdmin`
- 响应格式与项目统一（`Success(c, data)` / `Error(c, err)`）
- 公开接口响应体字段：`tenant_code`, `tenant_name`, `matched`（不含 tenant_id）
- errInvalidInput → 400，errDuplicateDomain → 409，errPermissionDenied → 403

**Acceptance（验证标准）:**
- AC: `GET /api/v1/public/tenant-domain?domain=xxx` 无 token 可访问，返回正确响应
- AC: `GET /api/v1/public/tenant-domain` 缺少 domain → 返回 400
- AC: `GET /api/v1/public/tenant-domain?domain=` + 超长字符串 → 返回 400
- AC: `GET /api/v1/admin/tenant-domains` 无 token → 返回 401
- AC: Create 保留域名 → 返回 HTTP 400（errInvalidInput 映射）
- AC: Create 重复域名 → 返回 HTTP 409（errDuplicateDomain 映射）
- AC: 非 SUPER_ADMIN token → 返回 HTTP 403（errPermissionDenied 映射）
- AC: CRUD 路由均在 admin 中间件链保护下
- AC: `go build ./common/auth/...` 零错误
- AC: 【回归】/auth/login 正常（RG-1）
- AC: 【回归】/api/v1/user/auth/send-code 正常（RG-4）

**自测:**
- 测试文件: `common/auth/handler/tenant_domain_handler_test.go`
- ST: GET /public/tenant-domain?domain=xxx（无 token）→ 200 + 正确 JSON
- ST: GET /public/tenant-domain 无 domain → 400
- ST: GET /admin/tenant-domains 无 token → 401
- ST: POST /admin/tenant-domains 创建成功 → 200 + 返回 id

---

### Task 6: Seed 菜单 + auth.go 依赖接入 ✅

**依赖**: Task 5

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `common/auth/seed.go`（新增 `SeedCR6Menus` 函数）
  - `common/auth/auth.go`（`buildDependencies` 中实例化 `TenantDomainService`/`Handler`，`Init` 中调用 `SeedCR6Menus`）
- 不触碰: 其他 Seed 函数、已有依赖

**Acceptance（验证标准）:**
- AC: 后端重启后，`admin_resource` 中存在 path=`/system/tenant-domain` 的菜单记录（幂等，可多次执行）
- AC: `TenantDomainHandler` 在 `Dependencies` struct 中有对应字段，且在 `buildDependencies` 中正确实例化
- AC: `GET /api/v1/admin/tenant-domains` 携带有效 token 返回 200（验证路由注册成功、Handler 已接入）
- AC: `TenantDomainService` 使用的 `LocalCache` 实例在服务关闭时调用 `Stop()`，与 LocalBlacklistStore 保持一致
- AC: `go build ./common/auth/...` 零错误

**自测:**
- 测试文件: `common/auth/seed_test.go`（续写）
- ST: SeedCR6Menus 首次执行 → admin_resource 新增 tenant-domain 菜单记录
- ST: SeedCR6Menus 重复执行 → admin_resource 记录数不变（幂等）

---

## Phase 2：前端

### Task 7: dev-web-admin — 域名管理页 ✅

**依赖**: Task 6（后端接口就绪）

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/api/tenant-domain.ts`（新建）
  - `dev-web-admin/src/views/system/tenant-domain/index.vue`（新建）
  - `dev-web-admin/src/router/static-routes.ts`（新增路由）
- 不触碰: 其他 views 文件、现有 API 模块

**Constraints（约束）:**
- API 路径前缀 `/api/v1/admin/tenant-domains`（走 Vite proxy `/api`）
- 表格列 `prop` 值必须与后端 JSON 字段名完全一致
- `id`/`tenant_id` 字段类型为 `string`（后端加 `json:",string"`）
- 编辑弹窗：`version` 字段以隐藏形式从列表数据读取，提交时携带，不展示给用户
- 删除确认文案必须含域名
- 分页参数：`page` / `page_size`（与后端一致）
- 租户下拉调用 `GET /api/v1/admin/tenants?page_size=100`（已有接口），受管理端中间件保护，与域名管理页权限一致
- 错误提示：通过 axios 响应拦截器统一展示后端返回的 `message` 字段，不自定义前端文案

**Acceptance（验证标准）:**
- AC: 进入 `/system/tenant-domain` 页面，表格正确展示域名列表
- AC: 新增弹窗：域名输入框 + 租户下拉（从 /api/v1/admin/tenants 加载），提交后列表刷新
- AC: 编辑弹窗：回填当前数据，version 隐藏携带，提交成功
- AC: 删除：二次确认文案含域名，确认后软删除，列表刷新
- AC: 域名搜索、租户筛选正常工作
- AC: 错误操作（重复域名、保留域名）显示 ElMessage.error 提示

---

### Task 8: dev-web-user — 登录页 tenant_code 自动解析 ✅

**依赖**: Task 9（utils/tenant.ts 就绪）

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-user/src/views/login/index.vue`（修改：接入 resolveTenantCode，加载状态）
- 不触碰: `stores/user.ts`、`router/index.ts`、`utils/request.ts`

**Constraints（约束）:**
- `onMounted` 立即将 `resolvingTenant` 设为 `true`，发送验证码按钮 disabled
- 查询完成（成功或 fallback）后才将 `resolvingTenant` 设为 `false`
- 按钮文案：解析中显示"加载中…"，解析完成后显示"获取验证码"
- 不在界面上展示 tenantCode 任何值

**Acceptance（验证标准）:**
- AC: 页面加载时调用 `resolveTenantCode()`，发送按钮禁用且显示"加载中…"
- AC: 查询完成后按钮恢复"获取验证码"，可点击
- AC: 点击发送验证码，使用自动解析的 tenantCode（而非硬编码值）
- AC: 界面上不显示任何租户编码输入框或文本

---

### Task 9: dev-web-user — vite.config.ts + utils/tenant.ts ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-user/vite.config.ts`（确认 proxy 规则覆盖 `/api/v1/public`）
  - `dev-web-user/src/utils/tenant.ts`（改为 API 调用，删除静态映射表）
- 不触碰: `src/utils/request.ts`、其他工具文件

**Constraints（约束）:**
- Vite proxy 使用 `/api` 前缀（已覆盖 `/api/v1/public`），无需额外配置，但需在注释中说明
- `resolveTenantCode` 函数签名：`async function resolveTenantCode(): Promise<string>`（无参数，内部自取 `window.location.hostname`）
- 使用 `axios`（直接 import，不经过 request.ts 封装的 baseURL）调用 `/api/v1/public/tenant-domain?domain={hostname}`
- timeout 3000ms
- 失败时返回 `import.meta.env.VITE_TENANT_CODE || 'default'`
- 删除 `tenant.ts` 中原有的静态映射表 `DOMAIN_TENANT_MAP`

**Acceptance（验证标准）:**
- AC: `vite.config.ts` proxy 中 `/api` 规则能正确转发 `/api/v1/public/tenant-domain` 到后端（注释说明覆盖关系）
- AC: `resolveTenantCode()` 无参调用 → 内部取 hostname，调用后端，返回对应 tenant_code
- AC: 后端不可达时 → 3 秒超时后返回 fallback 值，不抛出未捕获异常
- AC: `import.meta.env.VITE_TENANT_CODE` 未设置时 fallback 为 `'default'`

---

## 一致性自检

**自检结果**：

- [修正] ✅ 需求覆盖度：FR-1~FR-6 所有 AC 均有对应 Task
- [修正] ✅ 设计对齐：design.md 中所有新建/修改文件均在 Task Scope 中覆盖
- [修正] ✅ 文件路径：所有 Scope 文件路径已与工程目录结构核对
- [修正] ✅ 回归覆盖：RG-1~RG-5 在 Task 5 Acceptance 中引用
- [修正] ✅ 内部一致性：Task 依赖顺序无循环，并行 Task（7/8）无文件交叉
- [修正] ✅ 前后端对称：后端接口（Task 1-6）与前端调用（Task 7-9）一一对应

---

> **注意**：Task 9 应先于 Task 8 执行（Task 8 依赖 tenant.ts 函数签名就绪）。
