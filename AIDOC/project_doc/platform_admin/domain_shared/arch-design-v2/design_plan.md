# 设计计划：Platform Admin 架构设计 V2

## 设计方向

基于 V1 架构 review 的 8+2 个功能需求，V2 设计采用以下核心架构理念：

### 核心设计理念：「一切扩展皆应用」

将 Application 提升为平台的一等公民和唯一扩展单元。内置模块、插件、外部集成统一视为「应用」，共享同一套权限、配置、生命周期管理机制。

### 多端架构（Multi-Platform）

平台部署模型：**1 后端 + 2 前端 + N 插件前端**

```
Go Backend（1 个服务，统一 API）
    ├── dev-web-admin（管理端，PC + H5）
    ├── dev-web-user（用户端/C 端，PC + H5）
    └── Plugin Frontends（每个插件可选择性提供 admin/user 的 PC/H5 bundle）
```

引入 `platform` 维度：
- 前端平台：`admin`（管理端）、`user`（用户端）
- 终端形态：`pc`、`h5`
- 组合：`admin:pc`、`admin:h5`、`user:pc`、`user:h5`

影响的模型：
- `admin_application.platforms` — 声明应用在哪些平台有 UI
- `admin_resource.platform` — 菜单/资源归属哪个前端平台
- 插件 Manifest `frontends` — 声明各平台的 bundle 路径
- 前端 GetUserMenu 接口增加 `platform` 查询参数

后端 API 不按平台拆分，权限/业务逻辑统一；前端差异仅在 UI 层。

### 应用级 API 路由与权限识别（核心设计）

**问题：** 权限跟着租户订阅的应用走，后端如何识别请求属于哪个应用？

**方案：** 应用路由前缀绑定 + AppResolveMiddleware

```
请求: GET /api/v1/billing/invoices
    │
    ├─ 1. AuthMiddleware → 解析 JWT → userId + tenantId + roles
    │
    ├─ 2. AppResolveMiddleware → URL 前缀匹配 → app_code = "billing"
    │      → 查 admin_tenant_app 校验租户订阅 → 未订阅则 403
    │
    ├─ 3. DynamicPermissionMiddleware → 查该 app_code 下的 permission_code
    │      → 仅在当前应用范围内检查用户权限
    │
    └─ 4. 业务 Handler
```

**路由前缀约定：**
- 内置平台管理：`/api/v1/admin/...`（platform_admin 应用）
- 业务应用：`/api/v1/{app_code}/...`
- 插件：`/api/v1/plugin/{name}/...` 或自定义 route_prefix

**admin_application 新增字段：**
- `route_prefix` — 该应用 API 的路由前缀
- `platforms` — 该应用在哪些前端平台有 UI

**AppResolveMiddleware 启动时加载 route_prefix → app_code 映射表，请求时 O(1) 匹配。**

**dev-web-admin vs dev-web-user 区分：**
- 不通过 URL 区分前端来源，而是通过 Token 中的角色决定权限
- GetUserMenu 加 `?platform=admin` / `?platform=user` 返回对应平台菜单
- 管理角色 → 有 admin 平台的应用权限
- 普通用户角色 → 有 user 平台的应用权限

### 各模块设计方向

**1. 统一应用模型**
- `admin_application` 增加 `app_type`（BUILTIN/PLUGIN/EXTERNAL）+ `app_config` JSON 列
- 插件安装时自动创建 Application 记录；卸载时标记删除
- EXTERNAL 类型存储 OAuth2 client_id/secret/回调地址等配置

**2. 插件 ↔ RBAC 统一**
- 废弃独立的 `sys_menu` 和 `plugin.Registry` 的独立权限存储
- 插件启动时通过 `PluginResourceSyncer` 将菜单 → `admin_resource`、API → `admin_api_permission`
- 卸载时反向清理，保证幂等

**3. 角色继承 — 权限上界模型**
- `admin_role.parent_id` 语义明确为「权限天花板」
- 子角色可配置权限 ⊆ 父角色已有权限
- 实现方式：AssignResources/AssignApis 前先查父角色权限集做交集校验
- 父角色权限缩减时级联裁剪子角色超出部分

**4. 数据权限增强**
- `admin_data_scope` 增加 `scope_type` 枚举：ALL/SELF/DEPT/DEPT_TREE/CUSTOM
- GORM Callback 根据 scope_type 分支生成 WHERE 条件
- 激活 `OrganizationProvider` SPI 支撑 DEPT/DEPT_TREE

**5. 三级配置链**
- 新增 `admin_config` 表（替代 sys_config），增加 `scope`（SYSTEM/TENANT/USER）+ `scope_id` 列
- 查询时按 USER → TENANT → SYSTEM 优先级合并
- 功能开关作为特殊 config_type，租户管理员可在界面上启用/禁用

**6. 缓存性能优化**
- 引入事件驱动缓存失效：权限变更 → 发布事件 → 缓存层订阅并主动刷新
- `DynamicPermissionMiddleware` 中的 `getUserPermCodes` 结果走 L2 缓存
- `codeMap` 改为带版本号的缓存，API 权限变更时 bump version 触发重建

**7. 插件通信契约**
- 插件注册时声明 `ExposedActions []ActionDescriptor`（含 name + version + inputSchema + outputSchema）
- `CallPlugin` 增加 action version 校验
- 不兼容时返回明确 error（非静默失败）

**8. 前端 Plugin SDK 生命周期**
- `definePlugin` 增加 `teardown` 回调
- `registerExtension` 返回 `unregister` 函数
- 框架在插件禁用/卸载/热更新时统一调用 teardown
- Plugin SDK 同时提供给 admin-frontend 和 user-frontend 使用（共享包）

**9. 认证策略可插拔**
- 抽象 `AuthenticationStrategy` 接口：`Authenticate(ctx, credentials) → AuthResult`
- 内置实现：PasswordStrategy、OAuth2Strategy
- 通过配置选择默认策略，支持同时启用多种（如密码 + 企业微信 SSO）
- 管理端和用户端可配置不同的认证策略组合

**10. 租户订阅与功能开关联动**
- `admin_tenant_app` 增加 `enabled_modules` JSON 列
- 插件应用声明多个功能模块时，租户可选择性启用部分
- 角色权限配置时仅展示租户已启用模块的权限节点

**11. 多端资源隔离**
- `admin_resource` 增加 `platform` 字段（admin/user），标识菜单归属哪个前端
- GetUserMenu 接口增加 `?platform=admin` 查询参数
- 插件安装时按 manifest.frontends 声明将 bundle 分别部署到对应前端静态目录
- 插件 Manifest 扩展支持多端 bundle 声明：
  ```json
  {
    "frontends": {
      "admin": { "pc": "dist/admin-pc/", "h5": "dist/admin-h5/" },
      "user": { "pc": "dist/user-pc/", "h5": "dist/user-h5/" }
    }
  }
  ```

## 技术选型

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 缓存失效 | 事件驱动（EventPublisher SPI）| 比 TTL 过期更精确，减少脏读窗口 |
| 配置合并 | 单表多 scope 设计 | 比三张表简洁，查询用 COALESCE 模式 |
| 插件权限同步 | 启动时全量同步 + 增量事件 | 简单可靠，避免复杂的 diff 逻辑 |
| 角色继承校验 | 写时校验（assign 时检查）| 比读时计算性能好，避免运行时递归 |
| 认证策略 | Strategy 模式 + 配置路由 | 经典模式，易扩展 |

## 风险点

- [Risk-1] 角色继承级联裁剪逻辑复杂度高，父角色权限变更时需递归处理所有后代角色
- [Risk-2] 插件 ResourceSyncer 需处理插件升级场景（资源增减的 diff 合并）
- [Risk-3] 三级配置合并在高并发下需考虑缓存一致性

## 产出物

本 CR 最终产出 `design_v2.md`，包含：
- 完整 DDL（全量建表语句）
- 完整 API 设计表
- 核心逻辑伪代码/流程图
- 实体关系图
- 缓存策略
- 前端架构变更
