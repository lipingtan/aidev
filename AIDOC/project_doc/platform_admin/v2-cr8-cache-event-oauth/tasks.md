# 任务列表：V2-CR8 缓存事件驱动 + OAuth2 预留 + 扩展能力收尾

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 12 |
| 已完成 | 12 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 12/12 (100%) |
| 当前阶段 | 全部完成 |

---

## Phase 1: Redis L2 缓存 + 事件驱动失效

### Task 1: 新增 ErrNotImplemented 错误码 ⬜

**复杂度**: 低

**Scope（边界）:**
- 涉及文件: common/auth/errors/errors.go
- 不触碰: 其他文件

**Acceptance（验证标准）:**
- AC: 新增 `ErrNotImplemented = 50101` 到错误码常量区
- AC: go build ./... 零错误

---

### Task 2: 实现 PermCodeCache 结构 ⬜

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: common/auth/cache/perm_code_cache.go（新建）
- 涉及模块: common/auth/cache
- 不触碰: permission_cache.go、two_level_cache.go

**Constraints（约束）:**
- 使用 `sdk.Runtime.GetCacheAdapter()` 返回的适配器操作 Redis
- key 格式: `perm:{tenantID}:{userID}`
- TTL 默认 10 分钟
- adapter 为 nil 时所有方法 no-op（Get 返回 false，Set/Invalidate 不操作）
- 序列化使用 JSON（`encoding/json`）

**Acceptance（验证标准）:**
- AC: PermCodeCache.Get() 缓存命中时返回 []string, true
- AC: PermCodeCache.Get() 未命中时返回 nil, false
- AC: PermCodeCache.Set() 写入后 Get 可命中
- AC: PermCodeCache.Invalidate() 删除后 Get 不命中
- AC: adapter=nil 时 Get 返回 false，Set/Invalidate 不 panic
- AC: go build ./... 零错误

---

### Task 3: 新增 PermissionChanged 事件定义 ⬜

**复杂度**: 低

**Scope（边界）:**
- 涉及文件: common/event/bus.go
- 不触碰: 其他事件的 payload 定义

**Acceptance（验证标准）:**
- AC: 新增 `EventPermissionChanged = "permission.changed"` 常量
- AC: 新增 `PermissionChangedEvent` 和 `AffectedUser` 结构体
- AC: go build ./... 零错误

---

### Task 4: DynamicPermissionMiddleware 接入 PermCodeCache ⬜

**复杂度**: 高

**依赖**: Task 2

**Scope（边界）:**
- 涉及文件: common/auth/middleware/dynamic_permission_middleware.go
- 涉及模块: common/auth/middleware
- 不触碰: skipPaths 白名单、codeMap 逻辑、isSuperAdminByDB 逻辑

**Constraints（约束）:**
- 中间件构造函数新增可选参数 `permCodeCache *cache.PermCodeCache`
- getUserPermCodes 调用前先查缓存，未命中再查 DB 并写回
- permCodeCache 为 nil 时行为与现有完全一致（直接查 DB）

**Acceptance（验证标准）:**
- AC: 缓存命中时不执行 DB 查询
- AC: 缓存未命中时查 DB 并写入缓存
- AC: permCodeCache=nil 时直接查 DB（向后兼容）
- AC: go build ./... 零错误
- AC: 【回归】RG-3 Redis 不可用时降级
- AC: 【回归】RG-4 SUPER_ADMIN 放行不变

---

### Task 5: RoleService 权限变更后发布事件 ⬜

**复杂度**: 高

**依赖**: Task 3

**Scope（边界）:**
- 涉及文件: common/auth/service/role_service.go
- 涉及模块: common/auth/service
- 不触碰: 创建/更新/列表等非权限分配方法

**Constraints（约束）:**
- 事件必须在事务 Commit 成功后发布，不得在事务内
- AssignResources：收集当前角色 + 被裁剪子角色关联的 userIDs
- AssignApis：同上
- DeleteRole（PERMISSION_SET）：已有 affectedUserIDs

**Acceptance（验证标准）:**
- AC: AssignResources 事务成功后发布 PermissionChangedEvent
- AC: AssignApis 事务成功后发布 PermissionChangedEvent
- AC: DeleteRole(PERMISSION_SET) 删除成功后发布事件
- AC: 事务回滚时不发布事件
- AC: go build ./... 零错误

---

### Task 6: UserRoleService 角色绑定变更后发布事件 ⬜

**复杂度**: 中

**依赖**: Task 3

**Scope（边界）:**
- 涉及文件: common/auth/service/user_role_service.go
- 不触碰: 校验逻辑、SUPER_ADMIN 保护逻辑

**Constraints（约束）:**
- AssignRoles：方法返回 nil 后发布事件
- ReplaceRoles：事务回调返回 nil（Commit 成功）后发布事件
- payload: [{UserID, TenantID}]

**Acceptance（验证标准）:**
- AC: AssignRoles 成功后发布 PermissionChangedEvent
- AC: ReplaceRoles 事务成功后发布 PermissionChangedEvent
- AC: 操作失败时不发布事件
- AC: go build ./... 零错误

---

### Task 7: 注册事件订阅（缓存失效） ⬜

**复杂度**: 中

**依赖**: Task 2, Task 3

**Scope（边界）:**
- 涉及文件: common/auth/init.go 或 auth-rbac 模块初始化入口
- 不触碰: 其他事件订阅

**Constraints（约束）:**
- 订阅 `EventPermissionChanged`，遍历 AffectedUsers 调用 PermCodeCache.Invalidate
- 初始化时机：auth-rbac 模块注册路由时一并注册订阅

**Acceptance（验证标准）:**
- AC: 发布 PermissionChangedEvent 后对应用户缓存被清除
- AC: go build ./... 零错误
- AC: 【回归】RG-5 现有 PermissionCache 功能不变

---

## Phase 2: 操作日志风险分级

### Task 8: OperationLog Model 新增 risk_level + 高风险白名单 ⬜

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: common/auth/model/operation_log.go, common/auth/middleware/risk_level.go（新建）
- 不触碰: sys_opera_log 相关代码

**Constraints（约束）:**
- risk_level 字段: `gorm:"type:varchar(16);default:LOW;index:idx_log_risk"`
- 高风险白名单硬编码 map[string]bool（method:fullPath → true）
- GetRiskLevel 函数：命中白名单返回 HIGH，非 GET 返回 LOW，GET 返回空

**Acceptance（验证标准）:**
- AC: admin_operation_log 表 AutoMigrate 后包含 risk_level 列
- AC: GetRiskLevel("DELETE", "/api/v1/admin/tenants/:id") 返回 "HIGH"
- AC: GetRiskLevel("PUT", "/api/v1/admin/users/:id") 返回 "LOW"
- AC: GetRiskLevel("GET", "/api/v1/admin/users") 返回 ""
- AC: go build ./... 零错误
- AC: 【回归】RG-6 sys_opera_log 不受影响

---

### Task 9: 操作日志查询增加 risk_level 过滤 + 多租户隔离 ⬜

**复杂度**: 中

**依赖**: Task 8

**Scope（边界）:**
- 涉及文件: common/auth/handler/operation_log_handler.go（或对应查询接口）
- 不触碰: 日志写入逻辑

**Acceptance（验证标准）:**
- AC: SUPER_ADMIN 可查所有租户日志
- AC: TENANT_ADMIN 仅查本租户日志
- AC: 普通用户仅查本人日志
- AC: 支持 ?risk_level=HIGH 过滤
- AC: 前端操作日志列表页新增 risk_level 筛选下拉框（如已有日志列表页）
- AC: go build ./... 零错误
- AC: 【回归】RG-1 现有登录流程不受影响
- AC: 【回归】RG-2 租户隔离不被破坏

---

## Phase 3: OAuth2/LDAP 骨架

### Task 10: OAuth2Strategy + LDAPStrategy 骨架 + 注册 ⬜

**复杂度**: 中

**依赖**: Task 1

**Scope（边界）:**
- 涉及文件: common/auth/strategy/oauth2_strategy.go（新建）, common/auth/strategy/ldap_strategy.go（新建）, auth-rbac 初始化入口（StrategyRouter.Register 调用处）
- 不触碰: password_strategy.go, sms_strategy.go, router.go 接口定义

**Constraints（约束）:**
- 实现 AuthenticationStrategy 接口
- Authenticate 返回 AuthError{Code: ErrNotImplemented, Message: "...暂未对接..."}
- 注册到 StrategyRouter（key=oauth2, key=ldap）

**Acceptance（验证标准）:**
- AC: grant_type=oauth2 返回 501 + 错误信息
- AC: grant_type=ldap 返回 501 + 错误信息
- AC: 现有 password/sms 策略不受影响
- AC: go build ./... 零错误
- AC: 【回归】RG-8 现有策略正常工作

---

## Phase 4: 扩展能力收尾

### Task 11: 4 张表新增 ext_fields + admin_custom_field DDL ⬜

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: common/auth/model/user.go, common/auth/model/tenant.go, common/auth/model/application.go, app/user_auth/model/biz_user.go, common/auth/model/custom_field.go（新建）
- 不触碰: 现有字段定义、TableName、BeforeCreate

**Constraints（约束）:**
- ext_fields: `datatypes.JSON \`gorm:"type:json" json:"ext_fields"\``
- admin_custom_field 仅定义 Model + TableName，加入 AutoMigrate 列表
- 不实现 CRUD Handler

**Acceptance（验证标准）:**
- AC: AutoMigrate 后 4 张表均有 ext_fields 列（JSON DEFAULT NULL）
- AC: AutoMigrate 后 admin_custom_field 表存在
- AC: admin_custom_field 有 (tenant_id, object_code) 联合索引
- AC: 现有 INSERT 不受影响（ext_fields 不传不报错）
- AC: go build ./... 零错误

---

### Task 12: 安装向导 Redis 可选配置 + 前端表单 ⬜

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - 后端: app/setup/setup.go（SetupRequest + writeConfig）
  - 前端: dev-web-admin 安装向导页面（如内嵌 HTML 模板或独立前端组件）
- 不触碰: createDatabaseIfNotExists、runMigrations、SeedInitialData

**Constraints（约束）:**
- SetupRequest 新增 RedisAddr/RedisPassword/RedisDB 字段（非 required）
- writeConfig：RedisAddr 非空时写入 settings.cache 节（driver:redis, addr, password, db）
- 前端：安装向导新增 Redis 配置区域（可折叠，标注"可选"）

**Acceptance（验证标准）:**
- AC: 填写 Redis 配置后 settings.yml 包含 cache 节
- AC: 不填写 Redis 配置时 settings.yml 无 cache 节，框架降级内存缓存
- AC: 安装向导前端展示 Redis 配置输入框（host:port/password/db）
- AC: go build ./... 零错误
- AC: 【回归】RG-7 不配置 Redis 时系统正常启动
