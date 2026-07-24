# 任务列表：V2-CR5 双用户池 + C端认证 + 认证策略接口 + TenantIsolationCallback

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 14 |
| 已完成 | 14 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 14/14 (100%) |
| 当前阶段 | ✅ 全部完成 |

---

## Phase 1: 认证基础设施重构

### Task 1: 抽取 JWT 工具到 common/auth/strategy/ 包 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `common/auth/strategy/strategy.go`（新增）
  - `common/auth/strategy/claims.go`（新增）
  - `common/auth/strategy/token.go`（新增）
  - `common/auth/service/auth_service.go`（修改：Login/SelectTenant/Refresh/Logout 改为调用 strategy 包公共函数）
  - `common/auth/middleware/auth_middleware.go`（修改：解析 token 改用 strategy.ParseAccessToken）
- 涉及模块: common/auth/strategy, common/auth/service, common/auth/middleware
- 不触碰: app/ 下所有文件、common/auth/handler/、common/auth/repository/

**Constraints（约束）:**
- AccessClaims 新增 UserPool 和 TokenVersion 字段，旧 token（无此字段）解析为 UserPool="" 视为 "admin"、TokenVersion=0
- GenerateAccessToken 使用 Options struct 模式（AccessTokenOptions）
- 现有 AuthService 的 Login/SelectTenant/Refresh/Logout 行为必须 1:1 不变
- 先跑通现有 auth_integration_test.go 确认 baseline 后再改

**Acceptance（验证标准）:**
- AC: `common/auth/strategy/` 包编译通过，对外暴露 Strategy 接口 + Claims + Token 函数
- AC: AuthService.Login 调用 strategy.GenerateAccessToken/GeneratePlatformToken 输出与原逻辑一致
- AC: AuthService.SelectTenant/Refresh/Logout 行为不变
- AC: auth_middleware.go 使用 strategy.ParseAccessToken 解析 token
- AC: AuthContext 新增 UserPool/TokenVersion 字段
- AC: go build ./... 零错误
- AC: go vet ./... 无警告
- AC: 【回归】现有 auth_integration_test.go 全部 PASS（RG-1~5, RG-10）

**自测:**
- 测试文件: `common/auth/strategy/token_test.go`
- ST: GenerateAccessToken(opts={admin}) → 解析后 UserPool="admin"
- ST: GenerateAccessToken(opts={user, version=3}) → 解析后 TokenVersion=3
- ST: GeneratePlatformToken → 解析后 Tenants 列表正确
- ST: 解析旧格式 token（无 UserPool 字段）→ UserPool="" 不报错

---

### Task 2: 实现 PasswordStrategy + StrategyRouter ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `common/auth/strategy/password.go`（新增）
  - `common/auth/strategy/router.go`（新增）
  - `common/auth/handler/auth_handler.go`（修改：Login handler 增加 grant_type 路由）
- 涉及模块: common/auth/strategy, common/auth/handler
- 不触碰: app/user_auth/（SmsStrategy 在后续 Task）

**Constraints（约束）:**
- PasswordStrategy 封装现有 Login 逻辑（查 admin_user、验 bcrypt、查租户、签发 token）
- grant_type 不传或 ="password" 时走 PasswordStrategy
- grant_type 不在支持列表时返回 400
- Captcha 校验在 PasswordStrategy 内部执行
- StrategyRouter 通过 Register(strategy) 动态注册策略

**Acceptance（验证标准）:**
- AC: POST /auth/login 不传 grant_type → 响应与现有完全一致（RG-1）
- AC: POST /auth/login grant_type=password → 响应与不传一致（RG-2）
- AC: POST /auth/login grant_type=unknown → 400 错误
- AC: go build ./... 零错误
- AC: 【回归】SelectTenant/Refresh/Logout 不受影响（RG-3~5）

**自测:**
- 测试文件: `common/auth/strategy/password_test.go`, `common/auth/handler/auth_handler_test.go`
- ST: PasswordStrategy.Authenticate(正确凭证) → 返回 AuthResult（UserPool=admin）
- ST: PasswordStrategy.Authenticate(错误密码) → 返回错误
- ST: StrategyRouter.Route("password") → 返回 PasswordStrategy
- ST: StrategyRouter.Route("unknown") → 返回 ErrUnsupported
- ST: 【回归】Login handler 无 grant_type → 原有响应结构不变

---

### Task 3: TenantIsolationCallback 实现 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `common/auth/middleware/tenant_isolation_callback.go`（新增）
  - `common/auth/middleware/tenant_isolation_callback_test.go`（新增）
  - `common/auth/auth.go`（修改：注册 Callback）
- 涉及模块: common/auth/middleware
- 不触碰: data_scope_callback.go（仅调整注册顺序）

**Constraints（约束）:**
- 使用 GORM Schema 反射检测 tenant_id 字段（`db.Statement.Schema.LookUpField("TenantID")`）
- Create：tenant_id > 0 时自动填充，=0 时跳过
- Query/Update/Delete：tenant_id > 0 时注入 WHERE，=0 时跳过
- SystemOp 标记跳过（与 DataScope 共用机制）
- 注册顺序：tenant_query Before("auth:data_scope")，确保先于 DataScope 执行

**Acceptance（验证标准）:**
- AC: 含 tenant_id 的表 Create 操作自动填充 tenant_id
- AC: 含 tenant_id 的表 Query 操作自动注入 WHERE tenant_id=?
- AC: 不含 tenant_id 的表操作不受影响（RG-9）
- AC: tenant_id=0（超管）不注入过滤条件
- AC: DataScopeCallback 仍然正常工作（RG-7）
- AC: go build ./... 零错误

**自测:**
- 测试文件: `common/auth/middleware/tenant_isolation_callback_test.go`
- ST: Query 含 tenant_id 表 + ctx.tenantID=1 → SQL 包含 WHERE tenant_id=1
- ST: Query 含 tenant_id 表 + ctx.tenantID=0 → SQL 不包含 tenant_id 条件
- ST: Query 不含 tenant_id 表 → SQL 无变化
- ST: Create 含 tenant_id 表 + ctx.tenantID=1 → 记录 tenant_id=1
- ST: Create + ctx.tenantID=0 → 记录 tenant_id 不被填充
- ST: SystemOp 标记 → 跳过所有注入

---

### Task 4: DynamicPermissionMiddleware 改造（UserPool 跳过） ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `common/auth/middleware/dynamic_permission_middleware.go`（修改）
  - `common/auth/middleware/auth_middleware.go`（修改：注入 UserPool/TokenVersion 到 AuthContext）
- 不触碰: permission_engine.go、data_scope_middleware.go

**Acceptance（验证标准）:**
- AC: UserPool="user" 的请求跳过权限检查直接 c.Next()
- AC: UserPool="admin" 或 UserPool="" 的请求走原有权限检查逻辑（RG-6）
- AC: AuthContext.UserPool 和 TokenVersion 字段正确注入
- AC: go build ./... 零错误

**自测:**
- 测试文件: `common/auth/middleware/permission_middleware_test.go`
- ST: 请求带 UserPool=user token → 中间件直接放行
- ST: 请求带 UserPool=admin token → 走原有权限逻辑
- ST: 请求带旧 token（无 UserPool）→ 走原有权限逻辑

---

## Phase 2: C端认证域

### Task 5: biz_user DDL + Model + Repository ✅

**复杂度**: 高
**依赖**: 无

**Scope（边界）:**
- 涉及文件:
  - `app/user_auth/model/biz_user.go`（新增）
  - `app/user_auth/repository/biz_user_repo.go`（新增）
  - DDL 脚本（新增，放 config/ 或 migration/）
- 涉及模块: app/user_auth
- 不触碰: common/auth/model/（不复用 admin_user 结构）

**Constraints（约束）:**
- biz_user 含 token_version 字段（INT NOT NULL DEFAULT 1）
- 唯一约束 uk_tenant_phone(tenant_id, phone)
- 索引：idx_biz_user_tenant, idx_biz_user_phone, idx_biz_user_deleted
- ID 使用雪花算法（int64 + json:",string" tag）
- 所有 CRUD 方法通过 *gorm.DB 参数传入（便于上层控制事务/context）

**Acceptance（验证标准）:**
- AC: DDL 可执行无语法错误
- AC: Model 字段、GORM tag、JSON tag 定义正确
- AC: Repository 包含 Create/FindByID/FindByTenantPhone/Update/Delete/ListByTenant/IncrementTokenVersion
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/user_auth/repository/biz_user_repo_test.go`
- ST: Create biz_user → DB 有记录 + tenant_id 正确
- ST: FindByTenantPhone 存在 → 返回正确记录
- ST: FindByTenantPhone 不存在 → 返回 nil/ErrNotFound
- ST: IncrementTokenVersion → token_version + 1
- ST: 同租户重复 phone Create → 唯一约束错误

---

### Task 6: SmsService（验证码生成/校验/限频） ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `app/user_auth/service/sms_service.go`（新增）
  - `app/user_auth/spi/sms_sender.go`（新增：接口 + mock 实现）
- 涉及模块: app/user_auth
- 不触碰: common/

**Acceptance（验证标准）:**
- AC: 生成 4 位数字验证码
- AC: 验证码 5 分钟有效，过期校验失败
- AC: 60 秒限频，重复请求返回剩余等待时间
- AC: SmsSender 接口定义 + ConsoleMockSender 实现（打印验证码）
- AC: CodeStore 接口定义 + MemoryCodeStore 实现
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/user_auth/service/sms_service_test.go`
- ST: SendCode → 生成 4 位数字
- ST: VerifyCode 正确码 → 通过
- ST: VerifyCode 错误码 → 失败
- ST: VerifyCode 过期码 → 失败
- ST: 60 秒内重复 SendCode → 返回 ErrTooFrequent + 剩余时间

---

### Task 7: BizUserService（CRUD + 重置密码 + 强制登出 + 启用禁用） ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `app/user_auth/service/biz_user_service.go`（新增）
  - `app/user_auth/dto/biz_user_dto.go`（新增）
- 涉及模块: app/user_auth
- 不触碰: common/auth/service/user_service.go

**Constraints（约束）:**
- 创建时密码可选，有值则 bcrypt 哈希
- 重置密码生成 8 位随机字符串，bcrypt 存储，明文返回
- 强制登出：调用 repo.IncrementTokenVersion + 失效 token_version 缓存
- 启用/禁用：切换 status 字段（1/0）
- 所有写操作需乐观锁（version 字段校验）

**Acceptance（验证标准）:**
- AC: CreateBizUser 正确创建 + 密码哈希
- AC: ResetPassword 返回明文密码 + DB 中是 bcrypt
- AC: ForceLogout 使 token_version +1
- AC: ToggleStatus 切换 status
- AC: 并发创建同 phone → 仅一条记录（唯一约束）
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/user_auth/service/biz_user_service_test.go`
- ST: CreateBizUser(phone="13800138000") → DB 有记录
- ST: CreateBizUser 重复 phone → 返回错误
- ST: ResetPassword → 返回明文 + bcrypt.Compare 通过
- ST: ForceLogout → token_version 递增
- ST: ToggleStatus(id, 0) → status=0

---

### Task 8: SmsStrategy + C端认证 Handler ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `app/user_auth/strategy/sms_strategy.go`（新增）
  - `app/user_auth/handler/user_auth_handler.go`（新增：SendCode/Login/Logout）
  - `app/user_auth/dto/auth_dto.go`（新增）
- 涉及模块: app/user_auth
- 不触碰: common/auth/handler/auth_handler.go

**Constraints（约束）:**
- SmsStrategy 实现 strategy.AuthenticationStrategy 接口
- 登录时 tenant_code 无效/租户非 ACTIVE → 400
- 手机号未注册时自动创建（单事务保证原子性）
- 唯一约束冲突时捕获错误并重新查询已有记录（upsert 语义），事务使用默认隔离级别
- 签发 access_token 使用 cfg.JWT.UserAccessTokenTTL
- UserPool = "user"

**Acceptance（验证标准）:**
- AC: POST /api/v1/user/auth/send-code → 返回 expires_in
- AC: POST /api/v1/user/auth/send-code 60秒内重复 → 429 + retry_after
- AC: POST /api/v1/user/auth/login 正确码 + 已注册 → 200 + access_token
- AC: POST /api/v1/user/auth/login 正确码 + 未注册 → 自动注册 + access_token
- AC: POST /api/v1/user/auth/login 错误码 → 401
- AC: POST /api/v1/user/auth/login 无效 tenant_code → 400
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/user_auth/handler/user_auth_handler_test.go`
- ST: SendCode("13800138000", "valid_tenant") → 200 + 验证码生成
- ST: Login(正确码, 已存在用户) → 200 + token.UserPool="user"
- ST: Login(正确码, 新用户) → 200 + DB 新增 biz_user
- ST: Login(错误码) → 401
- ST: Login(无效 tenant_code) → 400

---

### Task 9: biz_user 管理端 Handler（CRUD + 管理操作） ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `app/user_auth/handler/biz_user_handler.go`（新增）
- 涉及模块: app/user_auth
- 不触碰: common/auth/handler/user_handler.go

**Acceptance（验证标准）:**
- AC: GET /api/v1/admin/biz-users → 分页列表 + tenant_id 隔离
- AC: GET /api/v1/admin/biz-users/:id → 返回详情
- AC: POST /api/v1/admin/biz-users → 创建
- AC: PUT /api/v1/admin/biz-users/:id → 更新（乐观锁）
- AC: DELETE /api/v1/admin/biz-users/:id → 软删除
- AC: POST .../reset-password → 返回明文密码
- AC: POST .../force-logout → token_version +1
- AC: POST .../toggle-status → status 切换
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/user_auth/handler/biz_user_handler_test.go`
- ST: GET /api/v1/admin/biz-users tenant_id=1 → 仅返回该租户数据
- ST: GET /api/v1/admin/biz-users/:id 存在 → 200 + 正确数据
- ST: POST .../reset-password → 响应含明文密码
- ST: POST .../force-logout → DB token_version 递增
- ST: POST .../toggle-status → DB status 变更

---

### Task 10: C端 GetUserMenu + 租户级缓存 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `app/user_auth/service/user_menu_service.go`（新增）
  - `app/user_auth/handler/user_auth_handler.go`（修改：增加 GetUserMenu 路由）
- 涉及模块: app/user_auth
- 不触碰: common/auth/service/resource_service.go（可引用其底层查询逻辑但不修改）

**Acceptance（验证标准）:**
- AC: GET /api/v1/user/menu → 返回该租户 platform=user 的菜单树
- AC: 重复请求命中缓存（第二次无 DB 查询）
- AC: InvalidateUserMenuCache(tenantID) 后下次请求重新查库
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/user_auth/service/user_menu_service_test.go`
- ST: GetUserMenu(tenantID=1) → 返回该租户的 user 菜单
- ST: 调用两次 → 第二次命中缓存
- ST: Invalidate 后调用 → 重新查库

---

### Task 11: app/user_auth 路由注册 + 配置扩展 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `app/user_auth/router.go`（新增）
  - `common/auth/config/config.go`（修改：新增 UserAccessTokenTTL）
  - `common/auth/auth.go`（修改：注册 user_auth 路由组 + SmsStrategy 注册到 StrategyRouter）
  - `config/settings.yml`（修改：新增 userAccessTokenTTL 配置项）
  - `config/biz_user_api_permission_seed.sql`（新增：biz_user 管理接口权限记录）
- 涉及模块: app/user_auth, common/auth/config
- 不触碰: common/auth/router.go 中现有管理端路由

**Acceptance（验证标准）:**
- AC: C 端路由 /api/v1/user/auth/* 正确注册
- AC: 管理端路由 /api/v1/admin/biz-users/* 正确注册且受权限保护
- AC: cfg.JWT.UserAccessTokenTTL 可从 settings.yml 读取，默认 7 天
- AC: SmsStrategy 注册到 StrategyRouter 后 grant_type=sms 可路由
- AC: biz_user 管理接口的 admin_api_permission 记录已注册，DynamicPermissionMiddleware 可正确识别
- AC: go build ./... 零错误
- AC: 【回归】现有管理端路由不受影响

**自测:**
- 测试文件: `app/user_auth/router_test.go`
- ST: 启动路由 → /api/v1/user/auth/login 路径存在
- ST: 启动路由 → /api/v1/admin/biz-users 路径存在
- ST: grant_type=sms 路由 → 到达 SmsStrategy

---

### Task 12: AuthMiddleware token_version + status 校验（C端用户） ✅

**复杂度**: 中
**依赖**: Task 1, Task 5

**Scope（边界）:**
- 涉及文件:
  - `common/auth/middleware/auth_middleware.go`（修改：UserPool=user 时校验 token_version + status）
- 不触碰: BizUserService（通过接口注入调用）

**Acceptance（验证标准）:**
- AC: UserPool=user + token_version 匹配 + status=1 → 放行
- AC: UserPool=user + token_version 不匹配 → 401
- AC: UserPool=user + status=0（被禁用） → 401（"账号已禁用"）
- AC: UserPool=admin → 不校验 token_version（走原黑名单逻辑）
- AC: 缓存 miss → fallback 查库
- AC: 查库失败 → 拒绝（fail-closed）
- AC: go build ./... 零错误
- AC: 【回归】管理端 token 验证流程不变（RG-1）

**自测:**
- 测试文件: `common/auth/middleware/auth_middleware_test.go`
- ST: user token + version=1 + DB version=1 + status=1 → 200
- ST: user token + version=1 + DB version=2 → 401
- ST: user token + version=1 + status=0 → 401（账号已禁用）
- ST: admin token → 不查 token_version
- ST: user token + DB 查询失败 → 401（fail-closed）

---

## Phase 3: 前端

### Task 13: dev-web-admin biz_user 管理页面 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `projects/demo/platform_admin/dev-web-admin/src/api/biz-user.ts`（新增）
  - `projects/demo/platform_admin/dev-web-admin/src/views/biz-user/index.vue`（新增）
  - `projects/demo/platform_admin/dev-web-admin/src/views/biz-user/components/`（新增）
- 涉及模块: dev-web-admin 前端
- 不触碰: 后端代码

**Acceptance（验证标准）:**
- AC: biz_user 列表页展示数据（表格 + 分页 + 搜索）
- AC: 新增/编辑弹窗功能正常
- AC: 重置密码按钮 → 弹窗展示明文密码
- AC: 强制登出按钮 → 确认后调用接口
- AC: 启用/禁用开关 → 调用 toggle-status
- AC: ID 字段用 string 类型（雪花 ID 精度保护）
- AC: API 请求带 /api/v1/admin/ 前缀

**自测:**
- 前端无自动化测试要求（手工验证）

---

### Task 14: dev-web-user 前端骨架搭建 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `projects/demo/platform_admin/dev-web-user/`（新建整个工程）
- 涉及模块: dev-web-user 前端
- 不触碰: 后端代码、dev-web-admin

**Acceptance（验证标准）:**
- AC: Vue3 + Vite + Element Plus + Vue Router + Pinia 工程可启动
- AC: 登录页展示手机号 + 验证码输入 + 发送验证码按钮
- AC: 登录成功后跳转主布局页
- AC: 主布局加载菜单（调用 /api/v1/user/menu）
- AC: token 存储 + 请求拦截器自动附加 Authorization header
- AC: 工程位于 `projects/demo/platform_admin/dev-web-user/`

**自测:**
- 前端无自动化测试要求（手工验证）

---

## 任务依赖关系

```
Task 1 ──→ Task 2 ──→ Task 8
  │                     ↑
  ├──→ Task 3          │
  │                     │
  ├──→ Task 4          │
  │                     │
  └──→ Task 12         │
                        │
Task 5 ──→ Task 6 ──→ Task 7 ──→ Task 8
  │                     │
  ├──→ Task 12         └──→ Task 9
  │                     │
  │                     └──→ Task 10
  │

Task 8 + Task 9 + Task 10 ──→ Task 11（路由总装配 + API 权限 seed）

Task 9 ──→ Task 13（管理端前端依赖管理端 API）
Task 8 + Task 11 ──→ Task 14（C端前端依赖 C端 API）
```

**显式依赖标注：**
- Task 2 依赖 Task 1
- Task 8 依赖 Task 2, Task 7
- Task 12 依赖 Task 1, Task 5
- Task 11 依赖 Task 8, Task 9, Task 10
- Task 13 依赖 Task 9, Task 11
- Task 14 依赖 Task 8, Task 11

**可并行的任务组：**
- Phase 1 内：Task 3 和 Task 4 可与 Task 2 并行（Task 3/4 依赖 Task 1 完成后即可开始），但 Task 12 不可与 Task 1 并行（文件交叉 auth_middleware.go）
- Phase 2 内：Task 5+6+7 可与 Task 1+2 并行（不同模块）
- Phase 3 内：Task 13 和 Task 14 可并行
