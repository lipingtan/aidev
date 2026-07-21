# 需求计划：数据权限增强 + 三级配置 + 配额管理（V2-CR3）

## 需求理解

- 目标：实现 5 种 scope_type 的数据权限注入、三级配置链替代 sys_config、功能开关、配额管理
- 范围：DataScopeCallback 增强、admin_config 新表及服务、OrganizationProvider SPI 激活、配额校验
- 预期效果：
  - 角色可配置 ALL/SELF/DEPT/DEPT_TREE/CUSTOM 五种数据权限类型
  - 配置支持 SYSTEM > TENANT > USER 三级覆盖
  - 租户可设置配额（最大用户数/角色数/应用数）并在操作前校验
  - 功能开关（is_feature_flag）可按租户开关功能模块

## 当前系统状态

| 模块 | 现状 |
|------|------|
| DataScopeCallback | 已实现维度+记录共享，但仅支持 CUSTOM（固定值列表）方式 |
| admin_data_scope | 存在但无 scope_type 字段 |
| admin_data_scope_config | 存在但无 supported_scope_types 字段 |
| OrganizationProvider SPI | 已定义接口 + NoOp 实现，未实际消费 |
| ConfigService | 基于 sys_config 单表，无 scope/三级逻辑 |
| admin_config 表 | 不存在，需新建 |
| 配额校验 | 无 |
| 功能开关 | enabled_modules 已存在（CR-1），但无 is_feature_flag 配置项机制 |

## 假设列表

- [假设-1] CR-2 的 DataScopeCallback 记录共享扩展已合入（当前代码已实现 `injectRecordShareScope`）
- [假设-2] DEPT/DEPT_TREE 依赖 OrganizationProvider 返回的部门 ID 列表，当前 NoOp 实现在未接入组织架构时返回空切片 → DEPT/DEPT_TREE 等效于"无数据"而非"所有数据"
- [假设-3] 三级配置 admin_config 替代 sys_config，但需要考虑旧数据迁移兼容
- [假设-4] 配额的默认值通过 SYSTEM scope 配置存储，未配置时取代码内硬编码默认值（如 999999 = 不限制）

## 澄清问题

- [Question-1] DEPT/DEPT_TREE 的 OrganizationProvider 在 CR-3 中是否只实现 SPI 消费逻辑（DataScopeCallback 中调用 provider），还是需要同时提供一个基于 admin_department 表的默认实现？
  **背景**：roadmap 写"激活 OrganizationProvider SPI"。如果只是激活消费，则 DEPT/DEPT_TREE 需要外部应用注册具体 provider 才能工作；如果提供默认实现，需要新建 admin_department 表。
  [Answer-1]

- [Question-2] sys_config 现有数据的迁移策略：是在本 CR 中自动迁移到 admin_config（scope=SYSTEM），还是两套表并存一段时间后再废弃？
  **行业做法**：多数 SaaS 采用"双写一段时间 + 后续 CR 彻底废弃"来降低风险。推荐方案：本 CR 新建 admin_config 并将现有 sys_config 数据一次性导入到 SYSTEM scope，ConfigService 统一读取 admin_config，但保留 sys_config 表不删。
  [Answer-2]

- [Question-3] 配额校验的拦截点：仅在 Service 层代码中手动调用，还是需要一个通用中间件/装饰器统一拦截？
  **分析**：配额类型有限（用户数/角色数/应用数），每种校验逻辑不同，中间件方式过于通用难以覆盖。推荐方案：Service 层在 Create 方法内调用 `configService.ResolveInt(tenantID, "quota.xxx", defaultValue)` 进行检查。
  [Answer-3]

- [Question-4] 功能开关（is_feature_flag）关闭后的 API 行为：返回 403 还是特定错误码（如 40302 功能未开启）？
  **行业做法**：AWS/Azure 等返回专用状态码以区分"无权限"和"功能未开启"。推荐：返回 403 但 message 区分（"权限不足" vs "该功能未开启"），code 可用 40302。
  [Answer-4]

- [Question-5] 三级配置的前端管理界面范围：是完整实现 SYSTEM/TENANT/USER 三个 scope 的 CRUD 管理页面，还是本 CR 只做后端 + SYSTEM scope 管理页（TENANT/USER scope 留后续）？
  [Answer-5]

## 非功能需求建议

- 性能：配置读取应加本地缓存（频繁调用 ResolveInt/ResolveString），避免每次请求查库
- 安全：TENANT/USER scope 配置写入需校验 tenant_id 归属
- 回归：现有 DataScopeCallback 维度+记录共享行为不被破坏
- 兼容：sys_config 接口保持可用（至少读取兼容）

## 影响范围预判

- 涉及模块：
  - `common/auth/middleware/` — DataScopeCallback 增强
  - `common/auth/model/` — DataScopeConfig、DataScope 模型增加字段，新增 AdminConfig 模型
  - `common/auth/service/` — ConfigService 重构、DataScopeService 增强、配额校验
  - `common/auth/handler/` — ConfigHandler 新增接口
  - `common/auth/spi/` — OrganizationProvider 消费
  - `common/auth/router.go` — 新路由注册
- 涉及文件（预估）：
  - `middleware/data_scope_callback.go` — scope_type 分支逻辑
  - `middleware/data_scope_context.go` — 数据权限上下文扩展
  - `model/data_scope_config.go` — 增加 supported_scope_types
  - `model/data_scope.go` — 增加 scope_type 字段
  - `model/admin_config.go` — 新建
  - `service/config_service.go` — 重构为三级配置
  - `service/data_scope_service.go` — scope_type 处理
  - `handler/config_handler.go` — 新接口
  - `discovery/auto_discover.go` — 新增接口 permission_code
- 可能的副作用：
  - ConfigService API 变更可能影响现有 handler 调用
  - admin_data_scope 增加 scope_type 字段，旧数据需设默认值 'CUSTOM'
