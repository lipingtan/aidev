# 任务：CR-1 应用模型升级 + 中间件链重构

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 10 |
| 已完成 | 10 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 10/10 (100%) |
| 当前阶段 | CR-1 全部完成 ✅ |

---

## Step A: 纯迁移（先稳定基础）

### Task 1: Model 层字段扩展 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `backend/common/auth/model/application.go`
  - `backend/common/auth/model/resource.go`
  - `backend/common/auth/model/api_permission.go`
  - `backend/common/auth/model/tenant.go`
  - `backend/common/auth/model/tenant_app.go`
- 不触碰: service/handler/middleware 层

**Constraints（约束）:**
- 新字段必须有合理默认值（platform 默认 'admin'，app_type 默认 'BUILTIN' 等）
- 使用 GORM tag 定义字段，AutoMigrate 自动加列
- JSON 类型字段使用 `datatypes.JSON`
- `admin_tenant.status` COMMENT 更新为四态说明

**Acceptance（验证标准）:**
- AC: `go build ./...` 零错误
- AC: 启动后 AutoMigrate 成功为已有表添加新列
- AC: 新列有正确的默认值和 COMMENT

---

### Task 2: 路由前缀迁移（后端） ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `backend/common/auth/router.go` — 路由注册重构
  - `backend/common/auth/handler/*.go` — 各 Handler 的 RegisterRoutes 方法（如有硬编码路径）
  - `backend/app/plugin/router/*.go` — 插件路由（如有）
- 不触碰: service 层逻辑、middleware 逻辑

**Constraints（约束）:**
- 路由组结构: `/auth/`（公开）+ `/api/v1/common/`（仅认证）+ `/api/v1/admin/`（完整中间件链）
- 认证路由（/auth/*）保持不变
- `/api/v1/common/user-menu` 新增（从原 `/api/v1/resources/user-menu` 迁移）
- 所有业务接口从 `/api/v1/xxx` 改为 `/api/v1/admin/xxx`
- DynamicPermissionMiddleware 仍挂在 admin group 上

**Acceptance（验证标准）:**
- AC: `go build ./...` 零错误
- AC: 所有业务接口在 `/api/v1/admin/` 前缀下可访问
- AC: `/api/v1/common/user-menu` 可访问
- AC: 旧路径 `/api/v1/tenants` 返回 404
- AC: `/auth/login` 正常可用
- AC:【回归】RG-1, RG-2 登录/租户选择正常

---

### Task 3: AutoDiscover 适配新前缀 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `backend/common/auth/discovery/auto_discover.go`
- 不触碰: 其他模块

**Constraints（约束）:**
- 扫描 `/api/v1/` 前缀路由（覆盖 admin 和 common）
- 根据路由路径推导 app_code（`/api/v1/admin/` → platform_admin）
- 推导 module_code（从路径第四段推断，如 `/api/v1/admin/tenants` → tenant-mgmt）
- 注册时写入新字段 app_code 和 module_code
- 幂等写入（按 method + url_pattern 去重）

**Acceptance（验证标准）:**
- AC: 启动后 admin_api_permission 表中所有记录的 app_code = 'platform_admin'
- AC: 记录包含合理的 module_code 值
- AC: 重复启动不产生重复记录
- AC: `go build ./...` 零错误

---

### Task 4: Seed 数据更新 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `backend/common/auth/seed.go`（或对应种子数据文件）
- 不触碰: 路由/中间件/service

**Constraints（约束）:**
- 种子数据必须包含: admin 用户 + default 租户(含 timezone/locale/currency) + SUPER_ADMIN 角色 + platform_admin 应用(含 app_type/route_prefix/platforms) + 租户订阅 + 管理端菜单(含 platform/module_code/app_code)
- 所有菜单 app_code='platform_admin', platform='admin'
- platform_admin 应用: app_type='BUILTIN', route_prefix='/api/v1/admin', platforms='["admin:pc","admin:h5"]'
- Seed 幂等：admin_user 表无记录时才执行

**Acceptance（验证标准）:**
- AC: 空库启动后 admin_user/admin_tenant/admin_role/admin_application/admin_tenant_app/admin_resource 有正确数据
- AC: admin_resource 中每条记录 platform='admin', app_code='platform_admin'
- AC: admin_application 中 platform_admin 记录字段完整
- AC: 系统可使用 admin/admin123 登录
- AC: `go build ./...` 零错误

---

### Task 5: 前端 API 路径迁移 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/api/*.ts` — 全部 API 文件
  - `dev-web-admin/src/views/` — 如有直接调用 API 路径的地方
  - `dev-web-admin/vite.config.ts` — proxy 配置（如需调整）
- 不触碰: 后端代码

**Constraints（约束）:**
- 替换规则: `/api/v1/{resource}` → `/api/v1/admin/{resource}`
- 特殊处理: `/api/v1/resources/user-menu` → `/api/v1/common/user-menu?platform=admin`
- 全局搜索确保无遗漏（搜索 `/api/v1/` 排除已迁移的）
- vite proxy 规则保持 `/api` 代理即可（不受影响）

**Acceptance（验证标准）:**
- AC: 前端项目编译通过（`npm run build` 零错误）
- AC: 前端无任何 API 调用使用旧路径 `/api/v1/tenants` 等
- AC: 前端登录 + 菜单加载 + 各页面 CRUD 正常
- AC: 应用管理页展示新字段（app_type/route_prefix/platforms）
- AC:【回归】RG-3 ~ RG-8 全部正常

---

## Step B: 新能力（在稳定基础上增强）

依赖: Task 1-5 全部完成

### Task 6: AppResolveMiddleware 实现 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `backend/common/auth/middleware/app_resolve.go`（新建）
  - `backend/common/auth/router.go`（注册中间件）
  - `backend/common/auth/auth.go`（初始化时加载 prefixMap）
- 不触碰: DynamicPermissionMiddleware（本 task 不改它）

**Constraints（约束）:**
- 启动时从 admin_application 表加载 route_prefix → app_code 映射
- **加载顺序必须保证**：AutoMigrate → Seed → RegisterRoutes → AutoDiscover → LoadModuleCodeMap → LoadAppPrefixMap（确保 module_code 数据已写入后再加载缓存）
- 请求时最长前缀匹配（倒序排列，O(n) 遍历，n < 100 可接受）
- 公共白名单: `/auth/*`, `/api/v1/common/*`
- SUPER_ADMIN 跳过订阅校验
- 匹配失败且非白名单 → 403 code=40303
- 租户未订阅 → 403 code=40302
- enabled_modules 非 NULL 且 module_code 不在列表中 → 403 code=40304
- module_code 通过 admin_api_permission 表的 module_code 字段查找（启动时缓存 method:path → module_code 映射）

**Acceptance（验证标准）:**
- AC: 请求 `/api/v1/admin/tenants` → context 中 app_code='platform_admin'
- AC: 非白名单未知路径 → 403
- AC: 模拟租户未订阅应用 → 403 code=40302
- AC: 模拟 module 未启用 → 403 code=40304
- AC: SUPER_ADMIN 不被拦截
- AC:【回归】RG-1 ~ RG-8 全部正常
- AC: `go build ./...` 零错误

---

### Task 7: AuthMiddleware 租户四态 + 自动降级 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `backend/common/auth/middleware/auth_middleware.go`
  - `backend/common/auth/service/auth_service.go`（新增 AutoDegradeToReadOnly 方法）
- 不触碰: 其他中间件

**Constraints（约束）:**
- status=1(ACTIVE) → 放行
- status=2(READ_ONLY) → 仅 GET/HEAD 放行，其余 403 code=40305
- status=0(DISABLED) / status=3(CANCELLING) → 403 code=40105
- 自动降级: status=1 且 expired_at 非空且已过期 → CAS 更新 `WHERE id=? AND status=1 AND expired_at < NOW()` → 设为 2
- 避免并发多次 UPDATE（CAS 条件保证）

**Acceptance（验证标准）:**
- AC: 正常租户(status=1) → 所有请求放行
- AC: READ_ONLY 租户 → GET 放行，POST/PUT/DELETE 返回 403
- AC: DISABLED 租户 → 所有请求 403
- AC: 设置 expired_at 为过去时间 + status=1 → 首次请求后自动变为 status=2
- AC:【回归】RG-1, RG-2 正常
- AC: `go build ./...` 零错误

---

### Task 8: GetUserMenu 增强（platform + modules 过滤） ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `backend/common/auth/service/resource_service.go` — GetUserMenu 方法
  - `backend/common/auth/handler/resource_handler.go` — 接收 platform 参数
- 不触碰: 其他 service

**Constraints（约束）:**
- 新增 platform 参数（从 query param 获取，必传）
- 过滤条件: `WHERE ... AND platform = ?`
- enabled_modules 过滤: 对每个已订阅应用，如果 enabled_modules 非 NULL，排除该应用中 module_code 不在列表中的资源
- SUPER_ADMIN 仍返回对应 platform 的所有菜单
- 接口路径: `GET /api/v1/common/user-menu?platform=admin`

**Acceptance（验证标准）:**
- AC: `?platform=admin` 仅返回 platform=admin 的菜单
- AC: 模拟 enabled_modules=["user-mgmt"] → 仅返回 user-mgmt 模块的菜单
- AC: enabled_modules=NULL → 返回全部
- AC: SUPER_ADMIN → 返回对应 platform 全部菜单
- AC:【回归】RG-5 菜单分配后返回正确
- AC: `go build ./...` 零错误

---

### Task 9: SUPER_ADMIN 保护规则 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `backend/common/auth/service/user_service.go` — DeleteUser 增加保护
  - `backend/common/auth/service/role_service.go` — DeleteRole 增加保护
  - `backend/common/auth/service/user_role_service.go` — UpdateUserRoles 增加保护
- 不触碰: handler/middleware 层

**Constraints（约束）:**
- 删除用户时: 如果该用户是最后一个 SUPER_ADMIN → 拒绝
- 删除角色时: 如果是最后一个 SUPER_ADMIN 角色 → 拒绝
- 更新用户角色时: 如果是 SUPER_ADMIN 用户移除自身 SUPER_ADMIN 角色 → 拒绝
- 判定逻辑通过数据库 COUNT 查询

**Acceptance（验证标准）:**
- AC: 尝试删除唯一 SUPER_ADMIN 用户 → 返回错误
- AC: 有多个 SUPER_ADMIN 时删除其中一个 → 成功
- AC: SUPER_ADMIN 尝试移除自身角色 → 返回错误
- AC: 删除最后一个 SUPER_ADMIN 角色 → 返回错误
- AC: `go build ./...` 零错误

---

### Task 10: 全链路验证（空库初始化 + E2E 回归） ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及: 整个系统端到端验证
- 不触碰: 不修改代码，纯验证

**Acceptance（验证标准）:**
- AC: drop 全部 admin_* 表 → 启动 → AutoMigrate 建表成功
- AC: Seed 写入完整种子数据
- AC: AutoDiscover 注册 API 权限（含 app_code + module_code）
- AC: admin/admin123 登录成功 → 选租户 → 获取菜单 → 全部正常
- AC: 前端登录 → 各页面 CRUD 正常
- AC: RG-1 ~ RG-8 全部验证通过
- AC: AppResolveMiddleware 正确拦截未订阅/未启用模块
- AC: 租户到期自动降级正常
- AC: SUPER_ADMIN 保护规则正常

---

## 任务依赖关系

```
Task 1 (Model) ─┐
Task 2 (路由)  ─┤── 可并行 ──→ Task 5 (前端) ─┐
Task 3 (Discover)┤                             │
Task 4 (Seed)  ─┘                             │
                                               ▼
                                        Step A 验收
                                               │
                    ┌──────────────────────────┼───────────────┐
                    ▼                          ▼               ▼
              Task 6 (AppResolve)    Task 7 (四态)    Task 9 (保护规则)
                    │                          │
                    └──────────┬───────────────┘
                               ▼
                        Task 8 (GetUserMenu)
                               │
                               ▼
                        Task 10 (全链路验证)
```

**并行说明**:
- Task 1/2/3/4 可并行（无文件交叉）
- Task 5 依赖 Task 2（需要后端路由先改好）
- Task 6/7/9 可并行（不同中间件文件）
- Task 8 依赖 Task 6（需要 AppResolve 先工作）
- Task 10 依赖全部
