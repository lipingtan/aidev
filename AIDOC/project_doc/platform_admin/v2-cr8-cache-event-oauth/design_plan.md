# 设计计划：V2-CR8 缓存事件驱动 + OAuth2 预留 + 扩展能力收尾

## 设计方向

### 核心架构决策

CR-8 涉及四个独立子系统，设计方向如下：

**1. Redis L2 缓存层**

当前 `DynamicPermissionMiddleware` 中 `getUserPermCodes()` 每次都执行 3 表 JOIN 查 DB。CR-8 引入 Redis 作为 L2 缓存层：

```
请求 → DynamicPermissionMiddleware
        │
        ▼
    Redis L2 (key = perm:{tenantID}:{userID})
        │ 命中 → 返回权限码列表
        │ 未命中 ↓
    DB 查询 → 写 Redis L2 (TTL=10min) → 返回
```

利用 go-admin 框架已有的 `sdk.Runtime.GetCacheAdapter()` 作为 Redis 客户端。`config.CacheConfig.Setup()` 已支持根据 settings.yml 中 `cache` 配置自动切换 memory/redis 适配器。

**与现有 PermissionCache 的关系：** 现有 `common/auth/cache/PermissionCache` 是 L1(角色级) + L2(用户级) 的进程内内存缓存，用于角色权限的组合计算。CR-8 新增的是**独立的 Redis 缓存层**，专用于 `DynamicPermissionMiddleware` 中 `getUserPermCodes` 的 DB 查询缓存，key 格式和生命周期完全不同。两套缓存独立运行，互不干扰。

**2. 事件驱动缓存失效**

利用现有 `common/event/DefaultBus` 发布 `PermissionChanged` 事件。权限变更操作（AssignResources/AssignApis/用户角色绑定变更）在事务提交成功后发布事件，EventBus 订阅方删除对应 Redis key。

```
RoleService.AssignResources() 
    → tx.Commit() 成功
    → event.DefaultBus.Publish("PermissionChanged", payload)
    → 订阅方: 遍历 affected userIDs，删除 perm:{tenantID}:{userID}
```

**跨 Pod 一致性：** 所有 Pod 读同一个 Redis 实例，key 删除后所有 Pod 下次请求都走 DB 重建缓存。EventBus 是进程内的，但 Redis key 删除是全局生效的。

**3. OAuth2/LDAP 骨架**

在 `common/auth/strategy/` 下新增 `oauth2_strategy.go` 和 `ldap_strategy.go`，实现 `AuthenticationStrategy` 接口，Authenticate 方法返回 HTTP 501 错误。在 StrategyRouter 注册 `grant_type=oauth2` 和 `grant_type=ldap`。

**4. 扩展能力**

- ext_fields: 在 4 张表的 GORM Model 中新增 `ExtFields` 字段（`datatypes.JSON`）
- admin_custom_field: 新增 Model + AutoMigrate，不实现 CRUD
- 操作日志 risk_level: 新增字段 + 高风险白名单判断逻辑嵌入操作日志中间件
- 安装向导 Redis 配置: `SetupRequest` 扩展 Redis 字段，`writeConfig` 条件写入 cache 节

## 技术选型

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| 直接用 `sdk.Runtime.GetCacheAdapter()` 操作 Redis | 框架原生支持，无额外依赖，settings.yml 配置即切换 | 接口为 `interface{}` 序列化，需 JSON 编解码 | ✓ |
| 引入独立 go-redis 客户端 | 更灵活的 Redis 操作（Pipeline/Lua） | 新增依赖，与框架缓存体系割裂 | ✗ |
| 在现有 PermissionCache 内加 Redis 支持 | 复用现有 TwoLevelCache 抽象 | PermissionCache 是角色粒度设计，和用户权限码缓存粒度不同，改动侵入大 | ✗ |
| 新增独立 `PermCodeCache` 结构专用于中间件 | 职责单一，与现有 PermissionCache 互不干扰 | 新增一个文件 | ✓ |

## 澄清问题

### [Question-1] Redis CacheAdapter 的序列化格式

go-admin 框架的 CacheAdapter（Redis 模式）使用 `encoding/json` 序列化 value。用户权限码列表 `[]string` 序列化后存为 JSON 数组字符串。是否接受这种方式？还是需要自定义编码（如逗号分隔）减小存储空间？

**推荐：** 直接用 JSON 数组。权限码列表通常 <100 个，JSON 开销可忽略（<2KB），代码最简洁。

[Answer-1]
同意
---

### [Question-2] 缓存失效的事件粒度

当 `AssignResources` 触发级联裁剪时，affected 可能涉及多个子角色、多个用户。事件 payload 的粒度：

- **选项A（聚合事件）**：一次操作发一个事件，payload 包含所有受影响的 `[]userID + tenantID` 对
- **选项B（逐用户事件）**：每个受影响用户发一个独立事件

**推荐：** 选项A。EventBus 是进程内同步分发（handler 在 goroutine 中执行），聚合发送减少 goroutine 数量，订阅方内部循环删除 Redis key 即可。

[Answer-2]
同意
---

### [Question-3] 操作日志 admin_operation_log 是否已存在

当前代码中 go-admin 原生的操作日志是 `sys_opera_log`（通过 `SaveOperaLog` 消费消息队列写入）。CR-8 的 `admin_operation_log` 是新表还是在 `sys_opera_log` 基础上改造？

需求计划假设-5 说"新增到 admin_operation_log 表，现有 sys_opera_log 不改动"。设计上是否新建独立表 + 独立中间件记录？

**推荐：** 新建 `admin_operation_log` 表 + 新中间件 `OperationLogMiddleware`（替代原生的 `LoggerMiddleware`），仅记录写操作（POST/PUT/DELETE），包含 risk_level 字段。原生 sys_opera_log 保持不动。

[Answer-3]
按新表
---

### [Question-4] 安装向导 Redis 配置写入 settings.yml 的格式

go-admin 框架 `config.CacheConfig` 读取 settings.yml 的 `settings.cache` 节。Redis 模式的标准配置格式为：

```yaml
settings:
  cache:
    driver: redis
    addr: 127.0.0.1:6379
    password: ""
    db: 0
```

不填写时不写入 `cache` 节，框架默认使用 memory 适配器。这个格式是否确认？

**推荐：** 遵循框架约定格式。

[Answer-4]
同意
---

### [Question-5] 用户-角色绑定变更的入口在哪里

需要在"用户-角色绑定变更"事务后发布缓存失效事件。当前用户-角色分配的 Service 方法名是什么？需要确认入口以便设计事件发布点。

从代码看 `RoleService` 有 `AssignResources`/`AssignApis`，但用户-角色绑定应该在 UserService 或 UserRoleService 中。请确认是否为 `common/auth/service/` 下的 `user_role_service.go` 或类似文件。

[Answer-5]
入口为 `common/auth/service/user_role_service.go` 中的两个方法：
- `UserRoleService.AssignRoles(userID, req)` — 追加模式，无显式事务，逐条 Create 成功即完成
- `UserRoleService.ReplaceRoles(userID, req)` — 全量替换，有 `db.Transaction` 包裹

事件发布点设计：
- `AssignRoles`：方法返回 nil 后，在调用方（或方法尾部）发布 `PermissionChanged` 事件
- `ReplaceRoles`：事务回调返回 nil（Commit 成功）后发布事件

payload 为 `{UserID, TenantID}`，订阅方删除 `perm:{tenantID}:{userID}` 缓存 key。

---

## 风险点

- [Risk-1] `sdk.Runtime.GetCacheAdapter()` 在安装完成前为 nil（Setup 未调用时）。DynamicPermissionMiddleware 启动时需 nil 检查，降级为直接查 DB。
- [Risk-2] EventBus.Publish 是异步 goroutine 执行。如果 Publish 后立即有同一用户的请求进来，可能缓存还未清除（时间窗口 < 1ms）。可接受，最多 TTL 后自然过期。
- [Risk-3] 安装向导新增 Redis 配置后，需要触发 `storage.Setup()` 重新初始化 CacheAdapter。需确认 `onInstalledCallbacks` 中包含 storage 初始化。
- [Risk-4] ext_fields 列用 `datatypes.JSON` 类型（GORM 官方 datatypes 包），需确认项目已引入 `gorm.io/datatypes` 依赖。
- [Risk-5] admin_operation_log 新中间件与原生 LoggerMiddleware 共存，需确保不重复记录。设计上新中间件只记录变更操作（非 GET），原生仅记录到 sys_opera_log。
