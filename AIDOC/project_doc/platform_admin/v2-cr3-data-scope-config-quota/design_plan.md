# 设计计划：数据权限增强 + 三级配置 + 配额管理（V2-CR3）

## 设计方向

基于已确认的需求，CR-3 技术方案分三个核心模块：

1. **DataScopeCallback 增强**：在现有 GORM Callback 中增加 scope_type 分支路由，将 SELF/DEPT/DEPT_TREE 的 WHERE 注入逻辑与现有 CUSTOM 逻辑并行
2. **三级配置服务**：新建 admin_config 表 + AdminConfigService（替代 ConfigService），实现三级合并查询 + 本地缓存
3. **配额/功能开关**：基于 AdminConfigService 的 Resolve 能力，在 Service 层植入配额校验和功能开关拦截

## 技术选型

| 选项 | 方案 | 推荐 |
|------|------|------|
| 组织架构模型 | 统一组织节点表 admin_org_unit + node_type 区分层级 + admin_user_org 多对多 | ✓ |
| 配置缓存 | sync.Map 本地缓存 + TTL 过期 | ✓ |
| | 引入 redis/外部缓存 | ✗（过早引入外部依赖） |
| 功能开关拦截 | 嵌入 DynamicPermissionMiddleware 中 | ✓ |
| 配额校验 | Service 层 Create 方法内主动调用 | ✓ |

## 澄清问题

- [Question-1] 组织架构数据模型设计：采用什么表结构支持 DEPT/DEPT_TREE 数据权限？
  **决定**：采用统一组织节点表（admin_org_unit）+ node_type 区分层级类型，组织架构和部门分开但共用一张树形表。用户通过 admin_user_org 多对多关联（is_primary 标注主归属）。
  [Answer-1]
  统一组织节点表 admin_org_unit（node_type 区分 COMPANY/BRANCH/DEPARTMENT/GROUP/TEAM）+ admin_user_org 多对多关联表（含 is_primary），保留灵活度。

- [Question-2] admin_user 如何关联组织节点？
  **决定**：使用 admin_user_org 中间表（多对多），支持兼任多节点，is_primary=1 标注主归属供数据权限使用。
  [Answer-2]
  admin_user_org 中间表（user_id + org_unit_id + is_primary + tenant_id），不在 admin_user 上加 dept_id 字段。

- [Question-3] DataScopeContext 扩展策略：当前 DataScopeContext 仅含 Dimensions[]，scope_type 信息如何传入 GORM Callback？
  **推荐 A**：扩展 DataScopeDimension 增加 ScopeType 字段 → Callback 内部按 ScopeType 分支。改动最小。
  [Answer-3]
  方案 A，扩展 DataScopeDimension 增加 ScopeType 字段。

- [Question-4] 功能开关 40302 的拦截时机？
  **决定**：嵌入 DynamicPermissionMiddleware 中，在权限码校验前先检查 module_code 对应的 feature_flag，关闭则返回 40302。
  [Answer-4]
  嵌入 DynamicPermissionMiddleware 中，权限码校验前检查 feature_flag。

## 风险点

- [Risk-1] DataScopeCallback 变更影响所有查询性能 — 需确保 scope_type=ALL 时零开销（直接 continue）
- [Risk-2] 三级配置缓存一致性 — 配置更新后缓存需及时失效，否则用户修改配置不生效
- [Risk-3] admin_org_unit 表初始为空时，DEPT/DEPT_TREE 返回空结果 — 需文档说明，避免用户配置后看不到数据
- [Risk-4] sys_config → admin_config 迁移中断 — 需幂等（重复执行不报错）
