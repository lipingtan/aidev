# 需求计划：auth-rbac（RBAC 权限管理框架 Go 版）

## 需求理解

基于 Java 版 `auth-pivot-java.md` 的设计，全新实现 Go 版企业级 RBAC 权限管理框架，作为平台基础框架的核心能力模块。前后端同时实现。

- 目标：全新实现（不复用 go-admin 现有权限代码），避免历史包袱
- 范围：完整实现全部功能 + 前端管理页面
- 技术栈：Go 1.24 + Gin + GORM + MySQL + Redis（可选）
- 代码位置：`backend/common/auth/`（在现有 backend 项目内新增模块）

## 关键决策

| 决策项 | 结论 |
|--------|------|
| SPI 扩展机制 | interface + 构造函数注入（显式，无框架） |
| 数据迁移 | 不迁移，全新数据、全新表 |
| 前端 | 同步实现（前后端一起做） |
| API 自动发现 | 三层策略：自动扫描 gin.Routes() 兜底 + 代码声明 RegisterAPIs() + 管理后台手动补充 |
| 代码目录 | `backend/common/auth/`（业务模块内） |
| 表名前缀 | `admin_`（非 sys_） |
| 模块独立性 | 用户权限模块不依赖框架其他模块，所有功能自包含 |
| 用户与租户关系 | 用户为全局实体（无 tenant_id），通过 admin_user_tenant M:N 关联租户 |
| 租户管理 | 框架管理租户完整生命周期（创建/启用/禁用/删除/初始化） |
| 组织结构与租户 | OrganizationProvider 按 tenant_id 隔离，每个租户有独立组织树 |

## 多租户 + 应用隔离关系

- 多租户（tenant_id）：数据隔离维度，不同租户完全隔离
- **用户为全局实体**：admin_user 不含 tenant_id，仅代表可登录身份
- **用户-租户 M:N 关联**：一个用户可关联多个租户，关联后才能获得对应身份和权限
- 应用隔离（app_code）：同一 tenant 下可有多个应用
- 角色对 tenant 隔离：每个 tenant 有自己的角色集（初始化时预置常用角色）
- 平台管理员角色：全局唯一，跨 tenant
- 同一角色可关联多个不同应用的权限
- **认证流程**：登录签发含可用租户列表的 Token → 选择/切换租户 → 签发含 tenantId 的访问 Token
- **缓存 Key 含租户维度**：user:{userId}:tenant:{tenantId}:app:{appCode}

## 核心功能清单

| # | 功能 | 说明 |
|---|------|------|
| 1 | 用户管理 | 全局用户，UserProvider 接口可替换用户来源 |
| 2 | 租户管理 | 租户生命周期（创建/启用/禁用/删除/初始化） |
| 3 | 角色管理 | 层级继承 parent_id、临时授权时间窗口、租户隔离 |
| 4 | 菜单/按钮资源管理 | 树形 admin_resource，含 app_code 归属，tenant_id 隔离 |
| 5 | 接口权限管理 | 独立树 admin_api_permission，按业务对象分组，tenant_id 隔离 |
| 6 | API 自动发现 | 三层策略（自动扫描 + 代码声明 + 手动补充） |
| 7 | 数据权限 | 维度注册 + GORM Scope 拦截 + target_entity 绑定 |
| 8 | 认证体系 | JWT 为主（可扩展 Session/OAuth2），含租户切换 |
| 9 | 两级缓存 | L1 角色权限 + L2 用户合并权限（本地/Redis），Key 含 tenant 维度 |
| 10 | 应用隔离 | admin_application + admin_role_app M:N |
| 11 | 组织结构 SPI | 框架不管理组织，提供 OrganizationProvider 接口，按 tenant 隔离 |
| 12 | 权限表达式引擎 | module:object:action 通配符匹配 |
| 13 | JWT 黑名单 | 强制下线能力 |
| 14 | 前端管理页面 | 租户/用户/角色/菜单/接口权限/数据权限/应用 管理页面 |

## 核心数据模型（表）

| 表名 | 职责 |
|------|------|
| admin_tenant | 租户（含状态、配置） |
| admin_user | 用户（全局，不绑定租户，代表登录身份） |
| admin_user_tenant | 用户-租户关联 M:N |
| admin_role | 角色（含层级 parent_id，tenant_id 隔离） |
| admin_application | 应用实体 |
| admin_resource | 菜单/按钮资源树（type: MENU/BUTTON，含 app_code，tenant_id 隔离） |
| admin_api_permission | 接口权限树（type: GROUP/ENDPOINT，含 app_code，tenant_id 隔离） |
| admin_user_role | 用户-角色关联（含 tenant_id、临时授权时间窗口） |
| admin_role_resource | 角色-菜单/按钮关联 |
| admin_role_api | 角色-接口权限关联（仅叶子节点） |
| admin_role_app | 角色-应用关联 M:N |
| admin_data_scope | 数据权限维度配置（角色绑定） |
| admin_data_scope_config | 数据权限维度注册表 |

## API 自动发现三层策略

```
层级 1：自动扫描（兜底）
  - 启动时遍历 gin.Engine.Routes()
  - 发现未注册的 endpoint → 插入 admin_api_permission（status=UNASSIGNED）
  - 保证不漏

层级 2：代码声明（精确）
  - 开发者在路由注册时调用 authrbac.RegisterAPIs(metadata)
  - 带完整元数据（title/group/permission）直接注册为 ACTIVE
  - 灵活度最高

层级 3：管理后台手动（补充）
  - 管理员在后台对 UNASSIGNED 的 endpoint 补充标题/分组
  - 拖入业务对象 GROUP 下
```

## 假设列表

- [假设-1] 现有 go-admin 的 sys_user/sys_role/sys_menu 表将被废弃，由新表完全替代
- [假设-2] 现有的 JWT 中间件和 Casbin 权限检查逻辑将被新模块替换
- [假设-3] 前端页面完全重做用户/角色/菜单/接口权限管理（替换当前 UserList/RoleList/MenuList）
- [假设-4] RBAC 模块启动时自动建表（GORM AutoMigrate）
- [假设-5] 平台管理员（super_admin）拥有全部权限，不受 RBAC 限制

## 非功能需求

- **性能**：权限判断热路径 O(1)（缓存命中）
- **安全**：密码 bcrypt 加密、JWT 过期 + 黑名单、SQL 注入防护
- **可扩展**：10+ 个 interface 扩展点
- **可配置**：每个子功能独立开关
- **零侵入默认**：引入模块 + 零配置即可运行
- **模块独立**：不依赖框架其他业务模块，仅依赖基础库（Gin、GORM、Redis client）

## 实施分阶段建议

| Phase | 内容 |
|-------|------|
| Phase 1 | 核心模型 + 租户管理 + 用户/角色 CRUD + JWT 认证 |
| Phase 2 | 菜单/按钮资源 + 接口权限树 + API 自动发现 |
| Phase 3 | 权限表达式引擎 + 缓存 + 中间件集成 |
| Phase 4 | 数据权限 + 应用隔离 + 组织 SPI |
| Phase 5 | 前端管理页面 |
| Phase 6 | 临时授权 + JWT 黑名单 + 高级功能 |
