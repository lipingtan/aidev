# CR-8 需求计划：缓存事件驱动 + OAuth2 预留 + 扩展能力收尾

## 需求理解陈述

CR-8 是整个 V2 路线图的收尾 CR，目标是做三件事：

1. **缓存性能优化**：权限检查目前每次都查 DB，CR-8 引入 L2 缓存（本进程内存缓存），并通过事件驱动（权限变更 → 发布事件 → 缓存失效）保证缓存一致性。
2. **认证策略扩展骨架**：预留 OAuth2 和 LDAP 接入口，骨架可编译但不完整对接三方服务。
3. **扩展能力收尾**：操作日志增加 risk_level 字段和分级可见性；关键业务表预留 ext_fields JSON 列；admin_custom_field 表 DDL 创建（不实现业务逻辑）。

---

## 假设列表

- [假设-1] L2 缓存使用 go-admin 框架已集成的 `CacheAdapter`（底层 Redis），通过 `sdk.Runtime.GetCacheAdapter()` 读写，不引入额外依赖。Redis 连接在安装向导阶段配置，支持可选跳过（跳过则降级为框架内存缓存）。
- [假设-2] 缓存失效通过已有的 EventBus（common/event/bus.go，DefaultBus）发布/订阅触发，EventBus 通知本 Pod 清理 L1 进程内缓存；跨 Pod 一致性由 Redis 缓存本身保证（所有 Pod 读同一个 Redis Key，Key 失效即全部失效）。
- [假设-3] OAuth2Strategy 骨架仅实现接口定义 + 空方法，返回"未实现"错误，确保 `go build` 通过
- [假设-4] LDAP 同上，骨架级别
- [假设-5] 操作日志的 risk_level 字段新增到 `admin_operation_log` 表，现有 `sys_opera_log`（go-admin 原生）不改动
- [假设-6] 日志分级可见性规则：SUPER_ADMIN 可查所有租户操作日志，TENANT_ADMIN 只能查本租户
- [假设-7] ext_fields 列以 `JSON NULL` 类型预留到 admin_tenant、biz_user 等核心表，不影响现有查询
- [假设-8] admin_custom_field 表只建 DDL（表结构），不实现任何 CRUD 接口
- [假设-9] 安装向导新增 Redis 配置步骤（主机/端口/密码/DB编号），Redis 为可选：填写则启用 Redis 缓存并写入 settings.yml；不填写则降级为内存缓存。`SetupRequest` 和 `writeConfig` 同步扩展。

---

## 澄清问题

### [Question-1] 缓存 key 粒度与失效策略

缓存方案已确定使用 **Redis（通过 go-admin CacheAdapter）**，key 粒度有两个选项：

- **选项A（用户粒度）**：key = `perm:{tenantID}:{userID}`，变更某用户权限时只失效该用户缓存。精准，跨 Pod 强一致。
- **选项B（租户粒度）**：key = `perm:{tenantID}`，任何权限变更都失效整个租户缓存。范围大但实现简单。

**推荐**：选项A（用户粒度），和 CR-5 已有的 `InvalidateBizUserCache(userID)` 保持一致风格。

[Answer-1]
选项A
---

### [Question-2] 操作日志 risk_level 的打标方式

如何判断一个操作是"高风险"：

- **选项A（接口粒度预设）**：在代码里为特定接口（如删除租户、修改超级管理员、重置密码）硬编码 `risk_level=HIGH`，其余默认 `LOW`。
- **选项B（配置驱动）**：在 `admin_config` 里配置哪些 API pattern 属于高风险，运行时动态判断。

行业做法：初期大多用选项A，够用且零运维成本；需要灵活调整时再升级到选项B。

**推荐**：选项A，代码里定义一个 `highRiskPaths []string` 白名单，覆盖最核心的危险操作。

[Answer-2]
选项A
---

### [Question-3] ext_fields 预留到哪些表

候选表：

| 表名 | 说明 |
|------|------|
| `admin_user`（auth-rbac） | 管理员用户 |
| `admin_tenant` | 租户 |
| `biz_user` | C 端用户 |
| `admin_application` | 应用 |

**推荐**：admin_tenant + biz_user 这两个最有可能被业务扩展字段的表。admin_user 和 admin_application 暂不加，减少 DDL 变更范围。

[Answer-3]
都加
---

### [Question-4] DynamicPermissionMiddleware L2 缓存的命中策略

目前 `DynamicPermissionMiddleware` 每次请求都调用 `getUserPermCodes(db, userID, tenantID)` 查 DB。L2 缓存的设计：

- **选项A（请求内缓存）**：token 不变则权限不变，缓存 key = `{userID}:{tenantID}`，TTL = 5-10 分钟，权限变更时主动失效。
- **选项B（仅 code_map 缓存）**：只缓存 method:path → permission_code 映射（已有 `codeMap`），不缓存用户权限（每次还是查 DB）。

当前代码已经对 `codeMap` 做了内存缓存，选项B 意味着不做用户权限缓存（效果有限）。

**推荐**：选项A，加用户权限的 TTL 缓存 + 事件驱动失效，这才是 roadmap 要求的"P99 < 5ms"目标。

[Answer-4]
A
---

### [Question-5] OAuth2Strategy 骨架的接口定义范围

`AuthenticationStrategy` 接口在 CR-5 已定义，OAuth2 骨架需要：

- 实现 `AuthenticationStrategy` 接口（`Authenticate(ctx, req) (*Claims, error)` 方法）
- 注册到 `StrategyRouter`（key = `"oauth2"`）
- 但实际 OAuth2 授权码流/token 交换等不实现

这样 `grant_type=oauth2` 会返回"OAuth2 策略暂未对接，请联系管理员"，而不是空指针 panic。

**推荐**：按此范围实现，占位骨架的意义就是防止未来新增时需要大改接口。

[Answer-5]
同意
---

## 非功能性需求建议

- **性能**：缓存命中时 `getUserPermCodes` 不查 DB，目标 P99 < 5ms
- **安全**：缓存失效必须在权限变更事务提交成功后触发，不得在事务中失效（避免缓存清空但事务回滚的 window）
- **多租户隔离**：操作日志查询必须按 tenant_id 隔离，SUPER_ADMIN 可跨租户查询
- **向后兼容**：ext_fields 列 DEFAULT NULL，不影响现有 INSERT 语句

---

## 影响范围预判

| 模块 | 文件/目录 | 影响类型 |
|------|----------|---------|
| `common/auth/middleware/` | `dynamic_permission_middleware.go` | 增加 Redis L2 缓存读写 |
| `common/auth/service/` | 权限相关 service | 权限变更后发布失效事件 |
| `common/event/bus.go` | 已有 EventBus | 新增订阅 |
| `common/auth/strategy/` | 新增 `oauth2_strategy.go`、`ldap_strategy.go` | 新文件 |
| `common/auth/service/auth_service.go` | StrategyRouter 注册 | 扩展注册逻辑 |
| `common/auth/model/` | admin_operation_log | 新增 risk_level 字段 |
| `common/auth/handler/` | OperationLogHandler | 增加权限过滤 |
| `app/setup/setup.go` | SetupRequest + writeConfig | 新增 Redis 可选配置项 |
| `config/settings.yml` | cache 节 | 新增 Redis 连接配置（安装后生成） |
| DB DDL | admin_tenant、biz_user、admin_custom_field | 新增字段/表 |
