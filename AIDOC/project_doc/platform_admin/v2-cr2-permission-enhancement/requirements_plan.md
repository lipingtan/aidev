# 需求计划：V2-CR2 权限体系增强

## 需求理解

- 目标：在 CR-1 建立的应用模型基础上，增强 RBAC 权限体系，实现角色继承约束、权限集叠加、字段级权限、记录共享四大特性
- 范围：`backend/common/auth/` 下的 model / service / handler / repository / middleware 层 + 前端角色权限配置相关页面
- 预期效果：
  - 子角色权限不可超越父角色范围（写时校验 + 级联裁剪）
  - PERMISSION_SET 类型角色可叠加额外权限给用户
  - 不同角色看到同一对象的不同字段（字段级安全）
  - 特定记录可额外共享给指定用户/角色/部门

## 假设列表

- [假设-1] CR-1 已完成，`admin_resource.platform`、`admin_resource.module_code`、`admin_api_permission.module_code` 字段已存在且有数据
- [假设-2] 当前 `admin_role.role_type` 的值域为 `SUPER_ADMIN / TENANT_ADMIN / NORMAL`，CR-2 新增 `PERMISSION_SET`
- [假设-3] 字段权限和记录共享为全新模块，不影响现有功能，现有接口在未配置字段权限时行为不变
- [假设-4] 级联裁剪采用同步方式（角色层级不深，通常 ≤ 3 层），无需异步队列
- [假设-5] 记录共享的 `object_code` 由业务应用定义，CR-2 只建立框架和 CRUD，不对具体业务表做集成
- [假设-6] 前端改动范围限定在 dev-web-admin（管理端），C 端不涉及

## 澄清问题

- [Question-1] 角色继承的分步交付顺序确认：先做「角色继承 + 级联裁剪」→ 再做「权限集」→ 最后做「字段权限 + 记录共享」，这个顺序可以接受？还是有调整需要？
  [Answer-1]
可以
- [Question-2] `assignable-resources` / `assignable-apis` 接口的行为：是返回父角色已有权限的子集（即子角色可选范围），还是只做写时校验不增加新查询接口？
  [Answer-2]
回父角色已有权限的子集
- [Question-3] 权限集（PERMISSION_SET）的分配方式：是在「用户-角色分配」界面同时展示普通角色和权限集供勾选，还是独立入口？
  [Answer-3]
同时展示普通角色和权限集供勾选
- [Question-4] 字段权限的 `object_code` 注册机制：是需要一个管理页面让管理员手动定义哪些对象有哪些字段，还是由代码硬编码对象字段列表供选择？
  [Answer-4]
代码硬编码对象字段列， 是否有办法自动注册？
- [Question-5] 记录共享的前端入口设计：design_v2.md 提到"业务详情页'共享'按钮"，CR-2 范围内是否需要实现一个通用的共享对话框组件？还是只做后端 API + 管理端的共享规则列表页？
  [Answer-5]
需要实现一个通用的共享对话框组件
- [Question-6] DataScopeCallback 扩展（OR 共享规则命中）：这个改动是在 CR-2 中做，还是留到 CR-3 中一起做？路线图说"CR-2 的 DataScopeCallback 扩展（OR 共享规则）"但同时 CR-3 也有 DataScopeCallback 改动。
  [Answer-6]
CR-2做
## 非功能需求建议

- 性能：权限计算（含权限集合并）应有缓存支持，P99 < 10ms；级联裁剪为低频操作，不做性能优化
- 安全：字段权限过滤必须在服务端执行（前端仅做展示优化），防止绕过
- 多租户：字段权限和记录共享均含 tenant_id 隔离
- 兼容性：未配置字段权限的对象/字段默认不限制（EDITABLE），零配置时系统行为与 CR-1 完全一致

## 影响范围预判

- 涉及模块：`backend/common/auth/` (model, service, handler, repository, middleware)
- 涉及文件（预估）：
  - 新增：`model/field_permission.go`、`model/record_share.go`、`service/field_permission_service.go`、`service/record_share_service.go`、对应 handler/repository
  - 修改：`service/role_service.go`（AssignResources/AssignApis 增加子集校验 + 级联裁剪）、`middleware/data_scope.go`（OR 共享规则）、`cache/` 相关、前端角色管理页
- 新增 DDL：`admin_field_permission`、`admin_record_share` 两张表
- 可能的副作用：AssignResources/AssignApis 增加校验后，如果现有数据中子角色权限已超出父角色范围，需要做数据修复或兼容处理
