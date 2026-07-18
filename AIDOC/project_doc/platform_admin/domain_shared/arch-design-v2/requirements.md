# 需求：Platform Admin 架构设计 V2

## 背景

Platform Admin V1 已实现多租户认证、RBAC 权限控制、应用管理、插件系统等基础能力。经架构 review 发现 8 个需改进的架构问题，需产出 V2 架构设计文档，在未上生产的窗口期完成架构优化。

## 用户故事

- 作为平台架构师，我希望获得一份完整的 V2 架构设计文档，以便指导后续各模块的实现重构
- 作为业务开发者，我希望新增业务应用/插件时无需修改平台核心代码，以便快速交付业务需求
- 作为运营人员，我希望不同租户可独立配置功能开关和参数，以便灵活服务不同客户

## 功能需求

### FR-1: 统一应用模型（一切扩展皆应用）

**描述：** 将 Application 模型扩展为平台的核心扩展单元，统一内置模块、插件、外部集成的管理方式

**验收标准：**
- WHEN 设计新的 admin_application 表 THEN 必须包含 app_type 字段区分 BUILTIN/PLUGIN/EXTERNAL
- WHEN 插件安装 THEN 系统 SHALL 自动创建对应 Application 记录
- WHEN 外部应用接入 THEN 系统 SHALL 通过 EXTERNAL 类型 + OAuth2 配置支持
- WHEN 查询应用列表 THEN 系统 SHALL 统一返回所有类型应用

### FR-2: 插件权限统一到 RBAC 体系

**描述：** 插件注册的菜单、API、权限码统一纳入 admin_resource 和 admin_api_permission 体系

**验收标准：**
- WHEN 插件启动注册 THEN 系统 SHALL 将插件声明的菜单写入 admin_resource（app_code = 插件 app_code）
- WHEN 插件启动注册 THEN 系统 SHALL 将插件声明的 API 写入 admin_api_permission
- WHEN 插件卸载 THEN 系统 SHALL 清理该插件注册的所有资源和 API 权限记录
- WHEN 管理员配置角色权限 THEN 插件权限 SHALL 与内置应用权限在同一界面统一管理

### FR-3: 插件租户级订阅控制

**描述：** 不同租户可选择性启用/禁用插件应用，支持插件功能子集的租户级控制

**验收标准：**
- WHEN 租户订阅插件应用 THEN 该租户下的角色才能配置该插件的权限
- WHEN 租户取消订阅插件 THEN 系统 SHALL 级联清除该租户角色对该插件的权限绑定
- WHEN 插件声明多个功能模块 THEN 系统 SHALL 支持租户级选择性启用部分模块

### FR-4: 角色继承 — 权限上界约束

**描述：** 子角色可配置的权限范围不能超过父角色的权限范围（子集关系）

**验收标准：**
- WHEN 为子角色分配菜单/API 权限 THEN 系统 SHALL 校验所选权限 ⊆ 父角色已有权限
- WHEN 父角色缩减权限 THEN 系统 SHALL 级联缩减子角色超出范围的权限
- WHEN 查询子角色可分配权限树 THEN 系统 SHALL 仅返回父角色权限范围内的节点
- WHEN 角色无 parent_id THEN 该角色为顶级角色，可分配范围 = 租户订阅应用的全部权限

### FR-5: 数据权限模型增强

**描述：** 支持规则表达式类型的数据权限，覆盖「全部」「仅本人」「本部门」「本部门及下级」「自定义值」等场景

**验收标准：**
- WHEN 数据权限维度绑定使用 scope_type=ALL THEN 系统 SHALL 不注入任何 WHERE 条件
- WHEN scope_type=SELF THEN 系统 SHALL 注入 `create_by = currentUserId`
- WHEN scope_type=DEPT THEN 系统 SHALL 通过 OrganizationProvider 获取用户部门并注入过滤
- WHEN scope_type=DEPT_TREE THEN 系统 SHALL 获取本部门及所有下级部门 ID 注入过滤
- WHEN scope_type=CUSTOM THEN 系统 SHALL 使用 dimension_values 中的固定值列表

### FR-6: 租户配置能力增强（三级配置链）

**描述：** 支持系统默认 → 租户级覆盖 → 用户级偏好的三级配置链，包括功能开关

**验收标准：**
- WHEN 查询配置项 THEN 系统 SHALL 按优先级合并：用户级 > 租户级 > 系统默认
- WHEN 租户管理员设置功能开关（如关闭某应用模块）THEN 该租户下该功能 SHALL 不可见
- WHEN 配置项不存在租户/用户覆盖 THEN 系统 SHALL 返回系统默认值

### FR-7: 缓存与权限中间件性能优化

**描述：** 优化 DynamicPermissionMiddleware 的缓存策略，消除每请求数据库查询

**验收标准：**
- WHEN 权限 codeMap 加载后有 admin_api_permission 变更 THEN 系统 SHALL 主动刷新缓存
- WHEN 用户权限已缓存 THEN DynamicPermissionMiddleware SHALL 不查数据库
- WHEN 缓存命中 THEN 权限检查链路 P99 SHALL < 5ms

### FR-8: 插件间通信契约

**描述：** 插件间调用增加版本契约和接口声明，避免无类型约束的裸通信

**验收标准：**
- WHEN 插件注册 THEN 系统 SHALL 要求声明对外暴露的 Action 列表（含版本号）
- WHEN 调用其他插件 THEN 系统 SHALL 校验目标 Action 是否存在且版本兼容
- WHEN 目标插件升级后移除某 Action THEN 调用方 SHALL 获得明确错误而非静默失败

### FR-9: 前端插件 SDK 生命周期管理

**描述：** 前端 Plugin SDK 支持完整的安装/卸载/热更新生命周期

**验收标准：**
- WHEN registerExtension 被调用 THEN 系统 SHALL 返回 unregister 清理函数
- WHEN 插件被禁用/卸载 THEN 框架 SHALL 调用插件 teardown 清理所有注册的扩展点
- WHEN 插件热更新 THEN 框架 SHALL 先 teardown 旧版本再 setup 新版本

### FR-10: 认证策略可扩展

**描述：** 认证方式从硬编码 JWT password 流程抽象为可插拔策略

**验收标准：**
- WHEN 系统配置认证方式为 password THEN 走现有用户名密码 + JWT 流程
- WHEN 系统配置认证方式为 oauth2 THEN 系统 SHALL 通过 OAuth2 授权码流程完成认证
- WHEN 新增认证方式（如 LDAP）THEN 仅需实现 AuthenticationStrategy 接口并注册

## 非功能需求

- 性能：权限检查链路 P99 < 5ms（缓存命中时）
- 扩展性：新增业务应用/插件时无需修改平台核心代码
- 设计自由度：可重新设计任何模块，追求最优架构（未上生产）
- 产出物：完整 DDL + 接口设计 + 核心逻辑伪代码
