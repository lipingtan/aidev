# 设计计划：V2-CR2 权限体系增强

## 设计方向

基于已确认的 11 个功能需求（FR-1 ~ FR-11），分 4 个 Phase 递进实现。技术方案沿用现有 Gin + GORM + 分层架构，在现有 `role_service.go` 上增量修改（Phase 1-2），新增独立模块（Phase 3-4）。

**核心技术决策：**
1. 角色继承校验：写时校验（AssignResources/AssignApis 入口），不做读时校验
2. 级联裁剪：同步递归（角色树通常 ≤ 3 层，无需异步）
3. 权限集合并：在权限缓存层合并，DynamicPermissionMiddleware 读取时已包含 PERMISSION_SET 权限
4. 字段过滤：Gin middleware 级别的 response interceptor（或 Handler 层显式调用 FieldFilter）
5. 记录共享：GORM Callback 中扩展 OR 子查询，需 object_code 上下文注入
6. 字段对象注册：struct tag 反射 + 数据库持久化（admin_field_object / admin_field_definition 两张元数据表）

## 技术选型

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 字段过滤时机 | Handler 层显式调用 FieldFilter.Filter() | 比 middleware 全局拦截更精确，避免对非业务接口（如树形、批量）的误过滤 |
| 字段元数据存储 | 独立表 admin_field_object + admin_field_definition | 支持手动注册 + 自动注册合并；管理员可修改描述 |
| 记录共享 WHERE 注入 | GORM Callback 中通过 context 获取 object_code | 与现有 DataScopeCallback 统一机制，业务代码无感知 |
| 权限集缓存 | 权限缓存 key 包含用户所有 roleIDs（含 PERMISSION_SET），缓存值已合并 | 无需额外缓存层 |

## 澄清问题

- [Question-1] FieldFilter 的调用方式：是每个 Handler 返回前手动调用 `FieldFilter.Filter(ctx, objectCode, data)`，还是通过 Gin middleware 自动拦截所有 JSON 响应并按路由元数据匹配 objectCode？前者更精确但侵入性高，后者自动化但可能误伤非业务接口。
  [Answer-1]
自动化处理，增加误伤后的手动管理修改逻辑
- [Question-2] 字段元数据表设计：是用一张表（admin_field_definition 含 object_code 字段）还是拆成两张表（admin_field_object 对象级 + admin_field_definition 字段级）？两张表更清晰但多一次关联查询。
  [Answer-2]
2张表
- [Question-3] 记录共享的 DataScopeCallback 扩展中，object_code 如何传递给 GORM Callback？方案 A：通过 context 注入（Handler 设置）；方案 B：通过 GORM Statement.Context 的自定义 key（Service 层设置）。
  [Answer-3]
方案B
- [Question-4] 前端「权限集」Tab 的位置：是在角色列表页顶部增加 Tab 切换（普通角色 / 权限集），还是在同一列表中用 tag 标识区分？
  [Answer-4]
Tab切换
- [Question-5] 级联裁剪发生时，是否需要通知受影响的子角色管理员（如操作日志记录 / 事件通知）？还是静默裁剪？
  [Answer-5]
静默裁剪，但给消息通知
## 风险点

- [Risk-1] 级联裁剪的事务范围：多级递归在同一事务中可能锁行较多。但角色树通常 ≤ 3 层 × 每层 ≤ 10 角色，影响可控。
- [Risk-2] 字段过滤对嵌套 JSON 的处理：如果响应包含嵌套对象（如关联查询返回的 tenant 信息内嵌在 user 中），FieldFilter 需支持递归或仅处理顶层字段。建议 CR-2 仅处理顶层字段，嵌套场景留后续优化。
- [Risk-3] 记录共享的 OR 子查询性能：大量共享规则时子查询可能变慢。可通过限制单记录共享数量（如 ≤ 50 条）+ 索引优化缓解。
- [Risk-4] 现有数据兼容：如果已有子角色权限超出父角色范围，启用校验后修改子角色会报错。需提供一次性数据修复脚本。
