# 需求：V2-CR8 缓存事件驱动 + OAuth2 预留 + 扩展能力收尾

## 背景

Platform Admin V2 路线图的收尾 CR。CR-5 建立了双用户认证策略体系，CR-4 实现了插件统一 RBAC，但权限检查每次请求仍查 DB，无法满足高并发场景下 P99 < 5ms 的目标。CR-8 引入 Redis L2 缓存 + 事件驱动失效机制，同时预留 OAuth2/LDAP 认证策略骨架，并为核心业务表预留扩展字段能力。

## 用户故事

- 作为**平台运维**，我希望权限检查从缓存命中而非每次查 DB，以便高并发时保持低延迟
- 作为**平台运维**，我希望权限变更后缓存自动失效，以便用户不需要手动刷新即可获得最新权限
- 作为**平台运维**，我希望在安装向导中可选配置 Redis，以便小型部署无需强依赖 Redis
- 作为**超级管理员**，我希望查看所有租户的高风险操作日志，以便及时发现安全隐患
- 作为**租户管理员**，我希望只看到本租户的操作日志，以便数据隔离
- 作为**平台架构师**，我希望 OAuth2/LDAP 接口骨架已就绪，以便未来对接第三方认证无需大改接口
- 作为**插件开发者**，我希望核心表预留 ext_fields JSON 列，以便在不改 DDL 的前提下存储扩展数据

## 功能需求

### Phase 1: Redis L2 缓存 + 事件驱动失效

#### FR-1: 用户权限 Redis 缓存

**描述：** DynamicPermissionMiddleware 中为用户权限码列表引入 Redis L2 缓存（通过 go-admin CacheAdapter），缓存命中时跳过 DB 查询。

**验收标准：**
- WHEN 用户首次请求且缓存未命中 THEN 系统 SHALL 查询 DB 获取权限码列表，写入 Redis 缓存（key = `perm:{tenantID}:{userID}`，TTL = 10 分钟）
- WHEN 用户再次请求且缓存命中 THEN 系统 SHALL 直接从 Redis 读取权限码列表，不查 DB
- WHEN Redis 不可用（连接失败/超时）THEN 系统 SHALL 降级为直接查 DB，不中断服务
- WHEN 安装向导未配置 Redis（降级为内存缓存）THEN 系统 SHALL 使用框架内存 CacheAdapter，功能不变

#### FR-2: 事件驱动缓存失效

**描述：** 权限变更操作完成后，通过 EventBus 发布失效事件，订阅方清除对应用户的缓存。

**验收标准：**
- WHEN 角色权限变更（AssignResources/AssignApis/级联裁剪）事务提交成功后 THEN 系统 SHALL 发布 `PermissionChanged` 事件，携带受影响的 userID 列表
- WHEN 用户-角色绑定变更事务提交成功后 THEN 系统 SHALL 发布 `PermissionChanged` 事件
- WHEN EventBus 收到 `PermissionChanged` 事件 THEN 系统 SHALL 删除 Redis 中对应的 `perm:{tenantID}:{userID}` 缓存 key
- WHEN 缓存失效在事务未提交时触发 THEN 系统 SHALL 视为 BUG（失效必须在事务提交成功后）

#### FR-3: 安装向导 Redis 可选配置

**描述：** 安装向导新增 Redis 配置步骤（主机/端口/密码/DB 编号），Redis 为可选项。

**验收标准：**
- WHEN 安装向导提交时包含 Redis 配置（host/port/password/db）THEN 系统 SHALL 写入 settings.yml 的 cache 节，启用 Redis 缓存
- WHEN 安装向导提交时未填写 Redis 配置 THEN 系统 SHALL 不写入 cache 节，框架降级为内存缓存
- WHEN Redis 配置写入后系统重启 THEN 系统 SHALL 通过 `sdk.Runtime.GetCacheAdapter()` 获取到 Redis 实例

### Phase 2: 操作日志风险分级

#### FR-4: 操作日志 risk_level 字段

**描述：** admin_operation_log 表新增 `risk_level` 字段（枚举：LOW/MEDIUM/HIGH），通过代码硬编码白名单为高风险操作打标。

**验收标准：**
- WHEN 请求命中高风险操作白名单（删除租户、修改超级管理员、重置密码、删除角色等）THEN 系统 SHALL 记录操作日志时设置 risk_level = HIGH
- WHEN 请求为写操作但不在白名单 THEN 系统 SHALL 设置 risk_level = LOW
- WHEN 查询操作日志时 THEN 系统 SHALL 支持按 risk_level 过滤

#### FR-5: 操作日志多租户隔离查询

**描述：** 操作日志查询接口按角色分级可见。

**验收标准：**
- WHEN SUPER_ADMIN 查询操作日志 THEN 系统 SHALL 返回所有租户的日志（可按 tenant_id 过滤）
- WHEN TENANT_ADMIN 查询操作日志 THEN 系统 SHALL 仅返回本租户的日志
- WHEN 普通用户查询操作日志 THEN 系统 SHALL 仅返回本人操作的日志

### Phase 3: OAuth2/LDAP 认证策略骨架

#### FR-6: OAuth2Strategy 骨架

**描述：** 实现 `AuthenticationStrategy` 接口的 OAuth2 骨架，注册到 StrategyRouter，实际调用时返回"未对接"错误。

**验收标准：**
- WHEN 登录请求 grant_type=oauth2 THEN 系统 SHALL 返回错误提示"OAuth2 策略暂未对接，请联系管理员"（HTTP 501）
- WHEN OAuth2Strategy 被注册到 StrategyRouter THEN 系统 SHALL 编译通过，不影响现有 password/sms 策略
- WHEN 未来实现 OAuth2 对接 THEN 只需填充 OAuth2Strategy 的方法体，无需修改接口或路由

#### FR-7: LDAPStrategy 骨架

**描述：** 实现 `AuthenticationStrategy` 接口的 LDAP 骨架，注册到 StrategyRouter，实际调用时返回"未对接"错误。

**验收标准：**
- WHEN 登录请求 grant_type=ldap THEN 系统 SHALL 返回错误提示"LDAP 策略暂未对接，请联系管理员"（HTTP 501）
- WHEN LDAPStrategy 被注册到 StrategyRouter THEN 系统 SHALL 编译通过，不影响现有策略

### Phase 4: 扩展能力收尾

#### FR-8: 核心表 ext_fields JSON 预留

**描述：** 为 admin_user、admin_tenant、biz_user、admin_application 四张表新增 `ext_fields` JSON 列。

**验收标准：**
- WHEN 执行 DDL 迁移后 THEN admin_user、admin_tenant、biz_user、admin_application 四张表 SHALL 各包含 `ext_fields JSON DEFAULT NULL` 列
- WHEN 现有 INSERT/UPDATE 语句未指定 ext_fields THEN 系统 SHALL 正常执行（DEFAULT NULL 不影响）
- WHEN 通过 API 写入 ext_fields THEN 系统 SHALL 接受任意合法 JSON 对象并持久化

#### FR-9: admin_custom_field 表 DDL

**描述：** 创建 admin_custom_field 表结构（仅 DDL，不实现业务 CRUD）。

**验收标准：**
- WHEN AutoMigrate 执行后 THEN 数据库 SHALL 存在 admin_custom_field 表，包含字段：id、tenant_id、object_code、field_name、field_type、field_label、sort_order、is_required、options(JSON)、created_at、updated_at
- WHEN 查看表结构 THEN 系统 SHALL 有 (tenant_id, object_code) 联合索引
- WHEN 本 CR 范围内 THEN 系统 SHALL 不实现该表的 CRUD 接口（留给后续 CR）

## 非功能需求

- **性能**：缓存命中时权限检查 P99 < 5ms；缓存未命中时不超过 50ms
- **可用性**：Redis 不可用时自动降级为 DB 查询，不中断服务
- **安全**：缓存失效必须在权限变更事务提交成功后触发，避免事务回滚导致缓存不一致
- **多租户隔离**：操作日志查询严格按 tenant_id 隔离；缓存 key 包含 tenantID 防止跨租户命中
- **向后兼容**：ext_fields 列 DEFAULT NULL 不影响现有 INSERT；未配置 Redis 时所有功能正常运行（内存缓存降级）
- **可观测性**：缓存命中/未命中/失效事件应可通过日志追踪（DEBUG 级别）
