# 任务列表：auth-rbac（RBAC 权限管理框架 Go 版）

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 34 |
| 已完成 | 34 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 34/34 (100%) |
| 当前阶段 | ✅ 全部完成 |

### 状态说明

| 标记 | 状态 | 含义 |
|------|------|------|
| ✅ | 已完成 | 代码已实现 + 测试已通过 + go build 零错误 |
| 🔨 | 进行中 | 正在编码或测试中 |
| ⬜ | 未开始 | 尚未启动实现 |
| 🚫 | 阻塞 | 前置依赖未满足，无法执行 |

**产品验收状态（附加在 ✅ 后）：**

| 标记 | 状态 | 含义 |
|------|------|------|
| 📋 | 待验收 | 开发完成，等待产品经理 Review |
| 🎯 | 验收通过 | 产品经理确认符合预期 |
| 🔄 | 需整改 | 产品经理提出修改意见，待修复 |

**本 CR 验收结论：** 验收通过（发现 18 个优化点，已拆分为新 CR `auth-rbac-ux-polish`）

## 全局约束（对所有任务生效）

**自测强制要求：**
- 每个任务编码完成后，必须同步编写测试验证逻辑（单元测试或集成测试）
- 测试验证通过后才算任务完成
- 后端任务：编写对应的 `_test.go` 文件，覆盖核心逻辑和边界条件，`go test ./...` 通过
- 前端任务：编写页面级别的交互验证脚本或手动验证清单，所有页面功能点可正常操作
- 涉及 API 的任务：编写 HTTP 接口测试（可用 httptest 或实际请求验证），覆盖正常流程 + 异常流程（参数错误/权限不足/数据不存在）

## Phase 1: 核心模型 + 租户管理 + 用户/角色 CRUD + JWT 认证

### Task 1: 模块骨架 + 配置结构 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: backend/common/auth/auth.go, backend/common/auth/config/config.go, backend/common/auth/errors/errors.go
- 涉及模块: auth（新建）
- 不触碰: 现有业务代码

**Acceptance（验证标准）:**
- AC: 创建 backend/common/auth/ 目录结构（model/repository/service/handler/middleware/cache/spi/engine/discovery/config/errors）
- AC: config.go 定义完整配置结构体（对应 YAML 配置项），含默认值
- AC: errors.go 定义全部错误码常量和 AuthError 结构体
- AC: auth.go 提供模块初始化入口函数 Init(cfg *Config, db *gorm.DB, engine *gin.Engine)
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 config_test.go：验证默认值正确加载、YAML 覆盖生效
- 编写 errors_test.go：验证错误码不重复
- `go test ./backend/common/auth/...` 全部通过

---

### Task 2: GORM 模型定义（全部表） ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/model/*.go
- 不触碰: 现有 model 文件

**Constraints（约束）:**
- 主键使用雪花算法（GORM tag: `gorm:"primaryKey"` + 自定义 BeforeCreate hook）
- 软删除使用 gorm.DeletedAt
- 时间字段使用 time.Time
- JSON 字段使用 datatypes.JSON (gorm.io/datatypes)

**Acceptance（验证标准）:**
- AC: 定义全部 14 张表的 GORM 模型（admin_tenant, admin_user, admin_user_tenant, admin_role, admin_user_role, admin_application, admin_tenant_app, admin_role_app, admin_resource, admin_role_resource, admin_api_permission, admin_role_api, admin_data_scope_config, admin_data_scope, admin_operation_log）
- AC: 每个模型含正确的 GORM tag（column/type/index/unique）
- AC: TableName() 方法返回 admin_ 前缀表名
- AC: AutoMigrate 可正确建表
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 model_test.go：连接测试 DB，执行 AutoMigrate，验证全部 15 张表创建成功
- 验证 TableName() 返回值正确（assert admin_ 前缀）
- 验证 BeforeCreate hook 自动生成雪花 ID
- `go test ./backend/common/auth/model/...` 全部通过

---

### Task 3: 雪花 ID 生成器 ✅

**复杂度**: 低

**Acceptance:**
- AC: 封装 bwmarrin/snowflake，提供 NextID() int64
- AC: 在 model BeforeCreate hook 中自动生成主键
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 snowflake_test.go：生成 1000 个 ID 验证唯一性、递增趋势
- `go test ./backend/common/auth/...` 全部通过

---

### Task 4: 租户管理 CRUD ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/repository/tenant_repo.go, backend/common/auth/service/tenant_service.go, backend/common/auth/handler/tenant_handler.go
- 涉及模块: auth

**Constraints（约束）:**
- 所有租户操作仅 SUPER_ADMIN 可执行
- 创建租户后自动执行初始化（预置角色 + 订阅默认应用）
- 禁用租户不删除数据，仅修改 status

**Acceptance（验证标准）:**
- AC: GET /api/v1/tenants 分页查询，支持 status/name 筛选
- AC: POST /api/v1/tenants 创建租户，自动初始化：预置角色 + 管理员账号（指定或自动创建）+ 关联租户 + 分配角色 + 订阅应用 + 同步全局 API 权限模板
- AC: PUT /api/v1/tenants/:id 更新租户信息
- AC: PUT /api/v1/tenants/:id/status 启用/禁用
- AC: DELETE /api/v1/tenants/:id 软删除
- AC: 创建时自动订阅 config.tenant.default-apps 中的应用
- AC: 操作日志记录（含原参数/新参数）
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 tenant_handler_test.go（httptest）：
  - 测试创建租户→返回 200 + 自动生成预置角色
  - 测试禁用租户→status 变更
  - 测试非 SUPER_ADMIN 调用→返回 403
  - 测试重复 tenant_code→返回 400
- `go test ./backend/common/auth/...` 全部通过

### Task 5: 用户管理 CRUD ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/repository/user_repo.go, backend/common/auth/service/user_service.go, backend/common/auth/handler/user_handler.go
- 涉及模块: auth

**Constraints（约束）:**
- admin_user 为全局表，无 tenant_id
- 密码 bcrypt 加密存储
- 查询用户列表时按租户上下文过滤（通过 admin_user_tenant JOIN）
- 解除用户-租户关联时，级联删除该用户在该租户下的 admin_user_role 记录 + 清 L2 缓存（RG-6）

**Acceptance（验证标准）:**
- AC: POST /api/v1/users 创建全局用户，密码 bcrypt 加密
- AC: GET /api/v1/users 租户上下文下返回已关联用户列表（分页）
- AC: PUT /api/v1/users/:id 更新用户信息（乐观锁）
- AC: DELETE /api/v1/users/:id 软删除
- AC: POST /api/v1/users/:id/tenants 关联用户到租户
- AC: DELETE /api/v1/users/:id/tenants/:tenantId 解除关联（级联删除该租户下角色绑定 + 清 L2 缓存）
- AC: GET /api/v1/users/:id/tenants 查询用户已关联租户列表
- AC: 操作日志记录
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 user_handler_test.go（httptest）：
  - 测试创建用户→密码不明文存储
  - 测试关联租户→admin_user_tenant 记录创建
  - 测试租户上下文查询→仅返回已关联用户
  - 测试乐观锁冲突→返回错误
  - 测试软删除→记录仍存在但 deleted_at 有值
- `go test ./backend/common/auth/...` 全部通过

---

### Task 6: 角色管理 CRUD ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/repository/role_repo.go, backend/common/auth/service/role_service.go, backend/common/auth/handler/role_handler.go
- 涉及模块: auth

**Constraints（约束）:**
- 角色归属 tenant_id，查询按租户隔离
- parent_id 需校验循环引用（遍历祖先链）
- parent_id 继承深度不超过 config.role.max-depth（默认 5），超过时返回 ErrMaxHierarchyDepth
- 删除角色前检查是否有用户绑定

**Acceptance（验证标准）:**
- AC: GET /api/v1/roles 返回当前租户角色列表/树
- AC: POST /api/v1/roles 创建角色（关联当前 tenant_id）
- AC: PUT /api/v1/roles/:id 更新角色（乐观锁）
- AC: DELETE /api/v1/roles/:id 软删除（有用户绑定时拒绝）
- AC: 设置 parent_id 时校验无循环引用
- AC: 操作日志记录
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 role_handler_test.go（httptest）：
  - 测试创建角色→关联正确 tenant_id
  - 测试循环引用检测→设置 A.parent=B, B.parent=A 返回 400
  - 测试有用户绑定时删除→返回 400
  - 测试租户隔离→租户 A 看不到租户 B 的角色
- `go test ./backend/common/auth/...` 全部通过

---

### Task 7: 用户-角色分配 ✅

**依赖**: Task 5, Task 6

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: backend/common/auth/service/user_role_service.go, backend/common/auth/handler/user_handler.go（扩展）
- 不触碰: role_handler.go

**Acceptance（验证标准）:**
- AC: POST /api/v1/users/:id/roles 为用户分配角色（支持 effective_start/effective_end 临时授权）
- AC: PUT /api/v1/users/:id/roles 全量替换用户角色
- AC: 分配前校验：用户已关联当前租户、角色属于当前租户
- AC: 操作日志记录（原角色列表 → 新角色列表）
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 user_role_test.go（httptest）：
  - 测试分配角色→admin_user_role 记录创建
  - 测试用户未关联租户时分配→返回 400
  - 测试角色不属于当前租户→返回 400
  - 测试全量替换→旧记录删除、新记录创建
- `go test ./backend/common/auth/...` 全部通过

---

### Task 8: JWT 认证（两阶段 Token） ✅

**依赖**: Task 5

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/service/auth_service.go, backend/common/auth/handler/auth_handler.go, backend/common/auth/middleware/auth_middleware.go
- 涉及模块: auth

**Constraints（约束）:**
- 使用 golang-jwt/jwt/v5
- platform_token TTL 7天, access_token TTL 2小时
- 单租户用户登录后直接返回 access_token（跳过选择步骤）
- AuthMiddleware 解析 access_token 注入 AuthContext 到 gin.Context

**Acceptance（验证标准）:**
- AC: POST /auth/login 验证凭证后签发 platform_token（含可用租户列表）
- AC: POST /auth/tenant/select 验证 platform_token + 租户关联后签发 access_token
- AC: POST /auth/refresh 用 platform_token 刷新 access_token
- AC: POST /auth/logout 将 access_token 加入黑名单
- AC: 单租户优化：tenants.length==1 时 login 直接返回 access_token
- AC: AuthMiddleware 解析 Token、检查黑名单、检查租户状态、注入 AuthContext
- AC: 未认证请求返回 401
- AC: 操作日志记录（LOGIN/LOGOUT）
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 auth_handler_test.go（httptest）：
  - 测试正确凭证登录→返回 platform_token + 租户列表
  - 测试错误密码→返回 401
  - 测试选择租户→返回 access_token
  - 测试选择未关联租户→返回 400
  - 测试单租户优化→直接返回 access_token
  - 测试 Token 过期→返回 401
  - 测试黑名单 Token→返回 401
  - 测试刷新 Token→返回新 access_token
- `go test ./backend/common/auth/...` 全部通过

---

### Task 9: SPI 接口定义 + 默认实现注册 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: backend/common/auth/spi/*.go
- 不触碰: 业务代码

**Acceptance（验证标准）:**
- AC: 定义全部 SPI 接口（UserProvider, RoleProvider, CacheAdapter, AuthProvider, TokenBlacklistStore, OrganizationProvider, DataScopeHandler, ApiDiscoveryStrategy, EventPublisher, OperationLogger）
- AC: auth.go Init() 中注册默认实现（可被外部注入覆盖）
- AC: 构造函数注入模式，无全局变量/单例
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 spi_test.go：验证每个 SPI 接口有对应的默认实现、可被 mock 替换
- `go test ./backend/common/auth/spi/...` 全部通过

---

### Task 10: 操作日志基础设施 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: backend/common/auth/service/operation_log_service.go, backend/common/auth/handler/operation_log_handler.go, backend/common/auth/repository/operation_log_repo.go
- 不触碰: 已有 handler

**Acceptance（验证标准）:**
- AC: OperationLogger 默认实现：异步写入（buffered channel + worker goroutine）
- AC: 写入失败降级到标准日志输出（不阻塞主业务）
- AC: GET /api/v1/operation-logs 分页查询，支持 module/action/user_id/target_type/start_time/end_time 筛选
- AC: 各 service 中调用 logger.Log() 的统一模式（helper 函数封装 AuthContext 提取）
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 operation_log_test.go：
  - 测试异步写入→调用 Log() 后记录最终出现在 DB
  - 测试写入失败降级→模拟 DB 错误不阻塞主流程
  - 测试查询接口→按 module/action/时间筛选返回正确结果
- `go test ./backend/common/auth/...` 全部通过

---

## Phase 2: 菜单/按钮资源 + 接口权限树 + API 自动发现

### Task 11: 菜单/按钮资源 CRUD ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/repository/resource_repo.go, backend/common/auth/service/resource_service.go, backend/common/auth/handler/resource_handler.go
- 涉及模块: auth

**Constraints（约束）:**
- 资源按 tenant_id 隔离
- 树形结构，parent_id 校验循环引用
- 支持按 app_code 过滤

**Acceptance（验证标准）:**
- AC: GET /api/v1/resources/tree 返回当前租户资源树（支持 app_code 过滤）
- AC: POST /api/v1/resources 创建资源（MENU/BUTTON）
- AC: PUT /api/v1/resources/:id 更新资源
- AC: DELETE /api/v1/resources/:id 删除资源（含子节点检查）
- AC: PUT /api/v1/resources/sort 拖拽排序（批量更新 sort_order + parent_id）
- AC: GET /api/v1/resources/user-menu 返回当前用户有权限的菜单树
- AC: 操作日志记录
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 resource_handler_test.go（httptest）：
  - 测试创建菜单→tree 接口返回新节点
  - 测试循环引用→返回 400
  - 测试 user-menu 接口→仅返回有权限的菜单
  - 测试拖拽排序→sort_order 更新正确
  - 测试租户隔离→租户 A 看不到租户 B 的资源
- `go test ./backend/common/auth/...` 全部通过

---

### Task 12: 角色-菜单权限分配 ✅

**依赖**: Task 6, Task 11

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: backend/common/auth/service/role_service.go（扩展）, backend/common/auth/handler/role_handler.go（扩展）
- 不触碰: resource_handler.go

**Acceptance（验证标准）:**
- AC: PUT /api/v1/roles/:id/resources 全量替换角色菜单权限（接收 resource_id 数组）
- AC: 校验 resource_id 属于当前租户
- AC: 操作日志记录（原权限列表 → 新权限列表）
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 role_resource_test.go（httptest）：
  - 测试分配菜单权限→admin_role_resource 记录正确
  - 测试跨租户 resource_id→返回 400
  - 测试全量替换→旧记录清除
- `go test ./backend/common/auth/...` 全部通过

---

### Task 13: 接口权限树 CRUD ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/repository/api_permission_repo.go, backend/common/auth/service/api_permission_service.go, backend/common/auth/handler/api_permission_handler.go
- 涉及模块: auth

**Constraints（约束）:**
- 按 tenant_id 隔离
- 区分 GROUP 和 ENDPOINT 类型
- ENDPOINT 带 url_pattern + http_method + permission_code

**Acceptance（验证标准）:**
- AC: GET /api/v1/api-permissions/tree 返回接口权限树
- AC: POST /api/v1/api-permissions 创建节点（GROUP 或 ENDPOINT）
- AC: PUT /api/v1/api-permissions/:id 更新节点
- AC: DELETE /api/v1/api-permissions/:id 删除节点（含子节点检查）
- AC: PUT /api/v1/api-permissions/:id/move 移动节点到指定 GROUP 下（状态变 ACTIVE）
- AC: GET /api/v1/api-permissions/unassigned 返回未分组 ENDPOINT 列表
- AC: 操作日志记录
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 api_permission_handler_test.go（httptest）：
  - 测试创建 GROUP + ENDPOINT→tree 返回正确层级
  - 测试 move 操作→status 变 ACTIVE、parent_id 更新
  - 测试 unassigned 接口→仅返回未分组 ENDPOINT
  - 测试删除有子节点的 GROUP→返回 400
- `go test ./backend/common/auth/...` 全部通过

---

### Task 14: 角色-接口权限分配 ✅

**依赖**: Task 6, Task 13

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: backend/common/auth/service/role_service.go（扩展）, backend/common/auth/handler/role_handler.go（扩展）
- 不触碰: api_permission_handler.go

**Acceptance（验证标准）:**
- AC: PUT /api/v1/roles/:id/apis 全量替换角色接口权限（接收 api_permission_id 数组）
- AC: 仅允许绑定 ENDPOINT 类型叶子节点
- AC: 传入 GROUP id 时自动展开为所有子 ENDPOINT id
- AC: 操作日志记录（原权限 → 新权限）
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 role_api_test.go（httptest）：
  - 测试绑定 ENDPOINT→admin_role_api 记录正确
  - 测试绑定 GROUP id→自动展开为子 ENDPOINT
  - 测试绑定非 ENDPOINT 类型→返回 400
- `go test ./backend/common/auth/...` 全部通过

---

### Task 15: API 自动发现 ✅

**依赖**: Task 13

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/discovery/auto_discover.go, backend/common/auth/discovery/register_api.go
- 涉及模块: auth

**Constraints（约束）:**
- 自动发现注册为全局模板（tenant_id=0）
- 路由数 < 500 同步执行，否则异步
- RegisterAPIs() 供开发者代码声明使用

**Acceptance（验证标准）:**
- AC: 启动时扫描 gin.Engine.Routes()，新 endpoint 插入 admin_api_permission（tenant_id=0, status=UNASSIGNED）
- AC: 已存在的 endpoint 不重复插入（幂等）
- AC: RegisterAPIs(metadata) 函数可带完整元数据注册为 ACTIVE
- AC: 路由数超过阈值时异步执行（goroutine + 延迟）
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 auto_discover_test.go：
  - 测试首次扫描→新 endpoint 插入 DB（status=UNASSIGNED）
  - 测试重复扫描→不产生重复记录（幂等）
  - 测试 RegisterAPIs()→status=ACTIVE + 完整元数据
- `go test ./backend/common/auth/discovery/...` 全部通过

---

## Phase 3: 权限表达式引擎 + 缓存 + 中间件集成

### Task 16: 权限表达式引擎 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/engine/permission_engine.go, backend/common/auth/engine/pattern.go
- 涉及模块: auth

**Constraints（约束）:**
- * 匹配恰好一层（不含冒号）
- ** 匹配一层或多层
- 精确码 HashSet O(1)，通配符 O(n)

**Acceptance（验证标准）:**
- AC: "sys:user:*" 匹配 "sys:user:add" ✓，不匹配 "sys:user:detail:view" ✗
- AC: "sys:**" 匹配 "sys:user:add" ✓，匹配 "sys:role:list" ✓
- AC: "**" 匹配一切 ✓
- AC: 精确码 O(1) 查找
- AC: 提供 NewPermissionEngine(codes []string) 构造，预编译模式
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 permission_engine_test.go：
  - 表驱动测试覆盖：精确匹配、* 单层、** 多层、边界场景（空串、单层、超长链）
  - 性能基准测试：10000 次匹配 < 10ms
- `go test ./backend/common/auth/engine/...` 全部通过

---

### Task 17: 两级缓存实现 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/cache/local_cache.go, backend/common/auth/cache/redis_cache.go, backend/common/auth/cache/two_level_cache.go, backend/common/auth/cache/permission_cache.go
- 涉及模块: auth

**Constraints（约束）:**
- L1 key: role:{roleId}, L2 key: user:{userId}:tenant:{tenantId}:app:{appCode}
- 本地缓存使用 go-cache
- Redis 使用 go-redis/redis/v9
- 失效：分批 100 个/批 + 随机延迟 50~200ms

**Acceptance（验证标准）:**
- AC: CacheAdapter 接口的 local/redis/two-level 三种实现
- AC: PermissionCache 封装 L1/L2 逻辑（查 L2 → 未命中查 L1 合并 → 写 L2）
- AC: 角色权限变更时清 L1 + 关联用户 L2（分批异步）
- AC: 用户角色变更时清该用户 L2
- AC: 根据 config.cache-type 自动选择实现
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 cache_test.go：
  - 测试 local：Set/Get/Delete/DeleteByPrefix 基本操作
  - 测试 L1/L2 逻辑：L2 未命中→查 L1→写 L2
  - 测试失效：角色变更后 L1 清除、关联用户 L2 清除
  - （Redis 测试可 skip 如无 Redis 环境，标注 //go:build integration）
- `go test ./backend/common/auth/cache/...` 全部通过

---

### Task 18: 授权中间件（PermissionMiddleware） ✅

**依赖**: Task 8, Task 16, Task 17

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/middleware/permission_middleware.go
- 涉及模块: auth

**Constraints（约束）:**
- 从缓存获取用户权限集
- 使用 PermissionEngine 匹配
- enforcement-strategy 配置支持（annotation-first/url-first/both/any）
- SUPER_ADMIN 跳过权限检查

**Acceptance（验证标准）:**
- AC: 中间件从路由元数据获取所需 permission_code
- AC: 从 L2 缓存获取用户权限集，调用 PermissionEngine.HasPermission()
- AC: SUPER_ADMIN 角色直接放行
- AC: 无权限返回 403 + ErrPermissionDenied
- AC: 临时授权过期的角色不纳入权限集
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 permission_middleware_test.go（httptest）：
  - 测试有权限→请求通过
  - 测试无权限→返回 403
  - 测试 SUPER_ADMIN→始终通过
  - 测试临时授权过期→权限不生效
- `go test ./backend/common/auth/middleware/...` 全部通过

---

## Phase 4: 数据权限 + 应用隔离 + 组织 SPI

### Task 19: 数据权限维度管理 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: backend/common/auth/repository/data_scope_repo.go, backend/common/auth/service/data_scope_service.go, backend/common/auth/handler/data_scope_handler.go
- 不触碰: middleware/

**Acceptance（验证标准）:**
- AC: GET /api/v1/data-scope-configs 查询维度注册列表
- AC: POST /api/v1/data-scope-configs 注册维度
- AC: PUT /api/v1/data-scope-configs/:id 更新维度
- AC: DELETE /api/v1/data-scope-configs/:id 删除维度
- AC: PUT /api/v1/roles/:id/data-scopes 角色配置数据权限（维度+target_entity+values）
- AC: 操作日志记录
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 data_scope_handler_test.go（httptest）：
  - 测试注册维度→DB 记录创建
  - 测试角色绑定数据权限→admin_data_scope 记录正确
  - 测试删除维度→记录删除
- `go test ./backend/common/auth/...` 全部通过

---

### Task 20: 数据权限 GORM Scope 拦截 ✅

**依赖**: Task 19

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/middleware/data_scope_middleware.go, backend/common/auth/middleware/data_scope_callback.go
- 涉及模块: auth

**Constraints（约束）:**
- 注册为 GORM Callback（gorm:query before）
- 从 context 读取 DataScope 配置
- target_entity 匹配当前查询 Model
- 多角色同维度取并集
- SystemOp 标记跳过

**Acceptance（验证标准）:**
- AC: 自动为匹配的查询拼接 WHERE {column} IN (?) 条件
- AC: 多角色同维度值取并集
- AC: target_entity 不匹配时不注入
- AC: SystemOpContext 标记的查询跳过注入
- AC: config.data-scope.enabled=false 时不注册 Callback
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 data_scope_callback_test.go：
  - 测试自动注入 WHERE 条件→查询 SQL 含 IN 子句
  - 测试多角色并集→合并值正确
  - 测试 target_entity 不匹配→不注入
  - 测试 SystemOp 标记→跳过注入
- `go test ./backend/common/auth/middleware/...` 全部通过

---

### Task 21: 应用管理 CRUD + 租户订阅 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/repository/application_repo.go, backend/common/auth/service/application_service.go, backend/common/auth/handler/application_handler.go
- 涉及模块: auth

**Constraints（约束）:**
- admin_application 全局表（SUPER_ADMIN 管理）
- admin_tenant_app 控制租户可用应用
- 角色绑定应用时校验租户已订阅

**Acceptance（验证标准）:**
- AC: GET /api/v1/applications 全局应用列表
- AC: POST /api/v1/applications 创建应用（SUPER_ADMIN）
- AC: PUT /api/v1/applications/:id 更新应用
- AC: DELETE /api/v1/applications/:id 删除应用
- AC: PUT /api/v1/tenants/:id/apps 租户订阅应用（设置可用应用列表，取消订阅时级联清除角色-应用绑定 RG-7）
- AC: GET /api/v1/tenants/:id/apps 租户已订阅应用列表
- AC: PUT /api/v1/roles/:id/apps 角色绑定应用（校验租户已订阅）
- AC: 操作日志记录
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 application_handler_test.go（httptest）：
  - 测试创建应用→DB 记录存在
  - 测试租户订阅→admin_tenant_app 记录正确
  - 测试角色绑定未订阅应用→返回 400 (ErrTenantAppNotSubscribed)
  - 测试非 SUPER_ADMIN 创建应用→返回 403
- `go test ./backend/common/auth/...` 全部通过

---

### Task 22: 组织结构 SPI + NoOp 默认实现 ✅

**复杂度**: 低

**Acceptance:**
- AC: OrganizationProvider 接口定义（GetOrgIds/GetOrgPath/GetSubOrgIds，均含 tenantId 参数）
- AC: NoOpOrganizationProvider 默认实现（返回空）
- AC: 数据权限 value_source=user_attr 时调用 OrganizationProvider
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 organization_provider_test.go：NoOp 实现返回空切片、不报错
- `go test ./backend/common/auth/spi/...` 全部通过

---

## Phase 5: 前端管理页面

### Task 23: 前端基础框架 + 登录/租户选择页 ✅

**依赖**: Task 8

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/views/login/, dev-web-admin/src/views/tenant-select/(新建), dev-web-admin/src/store/modules/auth.ts, dev-web-admin/src/utils/request.ts
- 涉及模块: 前端 auth 模块
- 注意: 前端代码生成到 `dev-web-admin/`（`frontend/` 已废弃）

**Constraints（约束）:**
- 登录页调用 POST /auth/login 获取 platform_token + 租户列表
- 单租户自动跳过选择，多租户展示选择页
- 选择租户后调用 POST /auth/tenant/select 获取 access_token
- request.ts 拦截器自动附加 Token、处理 401 刷新/跳转

**Acceptance（验证标准）:**
- AC: 登录页（用户名/密码输入 + 登录按钮）
- AC: 租户选择页（展示可用租户列表/卡片，点击进入）
- AC: 单租户时登录后自动进入主页
- AC: Token 存储（localStorage/sessionStorage）
- AC: 请求拦截器附加 Authorization header
- AC: 401 响应自动跳转登录页
- AC: access_token 过期自动用 platform_token 刷新

**自测（验证通过才算完成）:**
- 手动验证清单：
  - [ ] 输入正确凭证→成功进入租户选择/主页
  - [ ] 输入错误密码→提示错误
  - [ ] 多租户用户→展示选择页
  - [ ] 单租户用户→直接进入主页
  - [ ] Token 过期后→自动刷新（无感）或跳转登录
  - [ ] 关闭浏览器重开→platform_token 有效期内免登录

---

### Task 24: 前端布局框架 + 动态菜单 ✅

**依赖**: Task 23, Task 11

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/layout/, dev-web-admin/src/router/, dev-web-admin/src/store/modules/menu.ts
- 不触碰: 页面组件

**Acceptance（验证标准）:**
- AC: 主布局（侧边栏 + 顶栏 + 内容区）
- AC: 侧边栏菜单从 GET /api/v1/resources/user-menu 动态加载
- AC: 路由守卫检查认证状态
- AC: 顶栏显示当前用户/租户，支持切换租户
- AC: 登出按钮调用 POST /auth/logout

**自测（验证通过才算完成）:**
- 手动验证清单：
  - [ ] 侧边栏菜单根据后端返回动态渲染
  - [ ] 无权限菜单不显示
  - [ ] 切换租户→菜单刷新
  - [ ] 未登录访问→跳转登录页
  - [ ] 登出→清除 Token + 跳转登录页

---

### Task 25: 前端页面（租户管理） ✅

**依赖**: Task 24

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/views/system/tenant/(新建)
- 涉及模块: 前端 tenant 模块

**Acceptance（验证标准）:**
- AC: 租户列表页（表格 + 搜索/状态筛选/分页）
- AC: 租户新增弹窗（名称/编码/配置）
- AC: 租户编辑弹窗
- AC: 启用/禁用切换（确认对话框）
- AC: 删除操作（确认对话框）
- AC: 仅 SUPER_ADMIN 角色可见此菜单

**自测（验证通过才算完成）:**
- 手动验证清单：
  - [ ] 新增租户→列表刷新显示
  - [ ] 编辑租户→保存成功
  - [ ] 禁用→状态变更 + 确认对话框
  - [ ] 非 SUPER_ADMIN 账号→菜单不可见
  - [ ] 搜索/分页正常工作

---

### Task 26: 前端页面（用户管理） ✅

**依赖**: Task 24

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/views/system/user/(新建)
- 涉及模块: 前端 user 模块

**Acceptance（验证标准）:**
- AC: 用户列表页（表格 + 搜索/分页）
- AC: 用户新增弹窗（用户名/密码/邮箱/手机）
- AC: 用户编辑弹窗
- AC: 用户删除（确认对话框）
- AC: 用户-租户关联管理（穿梭框或多选）
- AC: 用户-角色分配弹窗（当前租户角色勾选 + 临时授权时间选择器）
- AC: 强制下线按钮（确认对话框）

**自测（验证通过才算完成）:**
- 手动验证清单：
  - [ ] 新增用户→列表显示
  - [ ] 编辑用户→保存成功
  - [ ] 关联租户→穿梭框操作正确
  - [ ] 分配角色→勾选保存后生效
  - [ ] 临时授权→时间选择器工作
  - [ ] 强制下线→确认后目标用户被踢出
  - [ ] 搜索/分页正常工作

---

### Task 27: 前端页面（角色管理） ✅

**依赖**: Task 24, Task 12, Task 14

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/views/system/role/(新建)
- 涉及模块: 前端 role 模块

**Acceptance（验证标准）:**
- AC: 角色列表/树展示
- AC: 角色新增/编辑弹窗（含 parent 选择下拉树）
- AC: 角色删除（有用户绑定时提示不可删）
- AC: Tab 1 — 菜单权限分配（资源树 checkbox 勾选）
- AC: Tab 2 — 接口权限分配（业务对象树 checkbox 勾选，勾选 GROUP 全选子节点）
- AC: Tab 3 — 数据权限配置（维度列表 + 值输入/选择）
- AC: Tab 4 — 应用绑定（已订阅应用 checkbox 勾选）

**自测（验证通过才算完成）:**
- 手动验证清单：
  - [ ] 角色树正确渲染层级
  - [ ] 新增角色→选择 parent 后保存
  - [ ] 菜单权限 Tab→勾选/取消保存后接口返回正确
  - [ ] 接口权限 Tab→勾选 GROUP 全选子节点
  - [ ] 数据权限 Tab→配置维度值保存成功
  - [ ] 应用绑定 Tab→仅显示租户已订阅应用
  - [ ] 删除有绑定用户的角色→提示不可删

---

### Task 28: 前端页面（菜单/资源管理） ✅

**依赖**: Task 24, Task 11

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/views/system/resource/(新建)
- 涉及模块: 前端 resource 模块

**Acceptance（验证标准）:**
- AC: 菜单树展示（按 app_code 下拉切换）
- AC: 拖拽排序（调用 PUT /api/v1/resources/sort）
- AC: 菜单/按钮新增弹窗（类型/名称/路径/图标/权限码/app_code）
- AC: 编辑/删除操作

**自测（验证通过才算完成）:**
- 手动验证清单：
  - [ ] 菜单树按 app_code 切换正确
  - [ ] 拖拽排序→顺序更新
  - [ ] 新增菜单/按钮→树刷新
  - [ ] 编辑→保存成功
  - [ ] 删除→确认后移除

---

### Task 29: 前端页面（接口权限管理） ✅

**依赖**: Task 24, Task 13

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/views/system/api-permission/(新建)
- 涉及模块: 前端 api-permission 模块

**Acceptance（验证标准）:**
- AC: 接口权限树展示（按业务对象分组）
- AC: UNASSIGNED endpoint 列表面板
- AC: 拖拽 UNASSIGNED endpoint 到 GROUP 节点（调用 move API）
- AC: GROUP/ENDPOINT 新增/编辑/删除弹窗

**自测（验证通过才算完成）:**
- 手动验证清单：
  - [ ] 接口权限树正确渲染
  - [ ] UNASSIGNED 列表显示未分组 endpoint
  - [ ] 拖拽到 GROUP→状态变 ACTIVE
  - [ ] 新增 GROUP/ENDPOINT→正确插入
  - [ ] 删除→确认后移除

---

### Task 30: 前端页面（应用管理 + 数据权限 + 操作日志） ✅

**依赖**: Task 24, Task 21, Task 19, Task 10

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/views/system/application/(新建), dev-web-admin/src/views/system/data-scope/(新建), dev-web-admin/src/views/system/operation-log/(新建)
- 涉及模块: 前端

**Acceptance（验证标准）:**
- AC: 应用列表 CRUD + 租户订阅管理（穿梭框）
- AC: 数据权限维度列表 + 角色维度值配置弹窗
- AC: 操作日志列表页（表格 + 模块/操作人/时间范围筛选 + 分页）
- AC: 日志详情弹窗（展示原参数/新参数 JSON diff）

**自测（验证通过才算完成）:**
- 手动验证清单：
  - [ ] 应用 CRUD 正常
  - [ ] 租户订阅穿梭框操作正确
  - [ ] 数据权限维度配置保存成功
  - [ ] 操作日志列表筛选+分页正常
  - [ ] 日志详情弹窗展示 old_value/new_value 对比

---

## Phase 6: 临时授权 + JWT 黑名单 + 高级功能

### Task 31: 临时授权时间窗口生效逻辑 ✅

**依赖**: Task 7, Task 17

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: backend/common/auth/service/permission_service.go（扩展）
- 不触碰: auth_handler.go

**Acceptance（验证标准）:**
- AC: 查询用户角色时过滤 effective_start/effective_end（NULL 表示永久有效）
- AC: 当前时间 < effective_start 或 > effective_end 的角色不纳入权限集
- AC: 缓存 L2 TTL 不超过最近一个 effective_end 到期时间（确保过期后缓存自动失效）
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 temporal_auth_test.go：
  - 测试 effective_end 已过期的角色→不纳入权限集
  - 测试 effective_start 未到的角色→不纳入权限集
  - 测试 NULL 时间窗口→永久有效
  - 测试缓存 TTL 计算→不超过最近 effective_end
- `go test ./backend/common/auth/...` 全部通过

---

### Task 32: JWT 黑名单（强制下线） ✅

**依赖**: Task 8

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: backend/common/auth/cache/blacklist_local.go, backend/common/auth/cache/blacklist_redis.go, backend/common/auth/handler/user_handler.go（扩展）
- 涉及模块: auth

**Acceptance（验证标准）:**
- AC: TokenBlacklistStore local 实现（sync.Map + 定时清理过期条目）
- AC: TokenBlacklistStore redis 实现（SET + EXPIRE）
- AC: POST /api/v1/users/:id/force-offline 将用户所有活跃 Token 加入黑名单
- AC: AuthMiddleware 在验证 Token 时检查黑名单
- AC: 黑名单条目 TTL = Token 剩余有效期
- AC: 操作日志记录（FORCE_OFFLINE）
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 blacklist_test.go：
  - 测试 Add→Contains 返回 true
  - 测试过期后→Contains 返回 false（模拟时间或短 TTL）
  - 测试 force-offline 接口→后续请求返回 401
- `go test ./backend/common/auth/...` 全部通过

---

### Task 33: 租户禁用联动（Token 即时失效） ✅

**依赖**: Task 8, Task 17

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: backend/common/auth/middleware/auth_middleware.go（扩展）, backend/common/auth/service/tenant_service.go（扩展）
- 不触碰: 其他中间件

**Acceptance（验证标准）:**
- AC: AuthMiddleware 验证 Token 时额外检查 admin_tenant.status（可缓存，短 TTL）
- AC: 租户禁用时清除该租户下所有 L2 缓存
- AC: 禁用后该租户下用户请求返回 ErrTenantDisabled (401)
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 tenant_disable_test.go（httptest）：
  - 测试禁用租户后→该租户 access_token 请求返回 401 + ErrTenantDisabled
  - 测试启用租户后→恢复正常
- `go test ./backend/common/auth/...` 全部通过

---

### Task 34: 集成测试 + 路由注册 ✅

**依赖**: Task 8, Task 18, Task 17

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: backend/common/auth/auth.go（完善）, backend/common/auth/router.go
- 涉及模块: auth

**Acceptance（验证标准）:**
- AC: auth.Init() 完成全部初始化（AutoMigrate + 注册路由 + 注册中间件 + 启动 API 发现）
- AC: 路由注册使用路由组，可独立挂载到任意 gin.Engine
- AC: auth.enabled=false 时 Init() 为 no-op
- AC: 完整端到端流程可走通：创建租户→创建用户→关联租户→登录→选择租户→访问受保护资源
- AC: go build ./... 零错误

**自测（验证通过才算完成）:**
- 编写 integration_test.go（端到端）：
  - 完整流程：Init→创建租户→创建用户→关联租户→登录→选择租户→请求受保护 API→成功
  - 反向流程：未认证→401、无权限→403、租户禁用→401
  - auth.enabled=false 时→所有 auth 路由不存在
- `go test ./backend/common/auth/... -tags=integration` 全部通过
