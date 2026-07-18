# V2 架构实现路线图

## 拆分原则

1. **每段 CR 可独立交付、独立测试**：完成后系统保持可运行状态
2. **依赖顺序清晰**：后序 CR 依赖前序 CR 的产出
3. **风险前置**：核心模型变更优先，衍生功能靠后
4. **粒度适度**：每段 CR 约 3-7 天工作量，涵盖后端+前端+测试

---

## 总览（8 段 CR）

```
CR-1: 应用模型升级 + AppResolveMiddleware + API 路由重构
  │
CR-2: 权限体系增强（角色继承 + 权限集 + 字段权限 + 记录共享）
  │
CR-3: 数据权限 scope_type + 三级配置 + 租户生命周期 + 配额
  │
CR-4: 插件系统统一 RBAC + 通信契约
  │
CR-5: 双用户池（biz_user）+ C端认证 + 用户端权限策略
  │
CR-6: 多端前端架构（platform/device + Plugin SDK V2）
  │
CR-7: 审批流引擎
  │
CR-8: 缓存事件驱动 + 认证策略可插拔 + 扩展能力收尾
```

---

## CR-1: 应用模型升级 + 中间件链重构

**目标**：建立 V2 核心骨架 — 统一应用模型 + 请求级应用识别

**范围**：
- DDL 变更：
  - `admin_application` 增加 app_type / route_prefix / platforms / modules / app_config / icon
  - `admin_resource` 增加 platform / module_code
  - `admin_api_permission` 增加 module_code
  - `admin_tenant_app` 增加 enabled_modules / updated_at
  - `admin_tenant` 增加 timezone / locale / currency / expired_at，status 扩展为四态
- 后端：
  - 实现 `AppResolveMiddleware`（URL 前缀 → app_code + 租户订阅 + 模块启用校验）
  - API 路由从 `/api/v1/` 迁移到 `/api/v1/admin/`（platform_admin 应用前缀）
  - 新增 `/api/v1/common/user-menu` 公共接口
  - `GetUserMenu` 增加 platform 参数过滤 + enabled_modules 过滤
  - 租户状态四态支持（AuthMiddleware 识别 READ_ONLY）
  - 管理员分级保护规则（SUPER_ADMIN 删除保护等）
- 前端：调整所有 API 请求路径适配新前缀
- Seed：更新种子数据适配新字段

**验收标准**：
- 现有功能正常运行（路径迁移无断裂）
- AppResolveMiddleware 正确识别应用 + 校验订阅 + 模块过滤
- GetUserMenu 根据 platform 参数返回对应菜单
- 租户 READ_ONLY 状态仅允许 GET
- 删除最后一个 SUPER_ADMIN 被阻止

**预估**：6-8 天

---

## CR-2: 权限体系增强

**目标**：角色继承为权限上界、权限集叠加、字段级权限、记录共享

**范围**：
- DDL 新增：
  - `admin_field_permission`（字段级权限）
  - `admin_record_share`（记录共享规则）
  - `admin_role.role_type` COMMENT 更新含 PERMISSION_SET
- 后端：
  - `AssignResources` / `AssignApis` 增加父角色子集校验
  - 实现 `cascadeTrimChildren` 级联裁剪
  - 新增 `GET /roles/:id/assignable-resources` 和 `assignable-apis` 接口
  - Permission Set 类型角色的创建/分配（不受 parent_id 约束）
  - 权限计算合并：角色权限 ∪ Permission Set 权限
  - FieldFilter 组件实现（Handler 返回时按角色过滤字段）
  - 字段权限 CRUD 接口
  - 记录共享 CRUD 接口
  - DataScopeCallback 扩展 OR 共享规则命中
- 前端：
  - 角色权限配置页增加"可分配范围"约束提示
  - 新增「权限集」管理 Tab
  - 字段权限配置界面（选择对象 → 配置字段 access）
  - 记录共享入口（业务详情页"共享"按钮）

**验收标准**：
- 子角色超出父角色范围 → 报错
- 父角色缩减 → 子角色级联裁剪
- Permission Set 叠加后用户获得额外权限
- HIDDEN 字段在 API 响应中被移除
- 被共享的记录对目标用户可见

**依赖**：CR-1（module_code、platform 字段）

**预估**：6-7 天

---

## CR-3: 数据权限增强 + 三级配置 + 配额

**目标**：数据权限 5 种 scope_type、三级配置链、功能开关、配额管理

**范围**：
- DDL 变更：
  - `admin_data_scope` 增加 scope_type 字段
  - `admin_data_scope_config` 增加 supported_scope_types
  - 新建 `admin_config`（替代 sys_config）
- 后端：
  - DataScopeCallback 按 scope_type 分支注入 WHERE（ALL/SELF/DEPT/DEPT_TREE/CUSTOM）
  - 激活 OrganizationProvider SPI（DEPT/DEPT_TREE）
  - ConfigService 实现三级合并查询（USER > TENANT > SYSTEM）
  - ConfigHandler：resolve/:key、feature-flags 接口
  - 功能开关：is_feature_flag 标识 + 租户级开关设置
  - 配额校验：用户数/角色数/应用数超限拦截
  - TenantIsolationCallback（业务表自动注入 tenant_id）
- 前端：
  - 数据权限配置支持 scope_type 选择
  - 系统配置页改造（三级配置 + scope 切换）
  - 功能开关管理界面

**验收标准**：
- 5 种 scope_type 正确注入 WHERE 条件
- 三级配置合并正确
- 配额超限拒绝操作
- 功能开关关闭后对应模块 403

**依赖**：CR-1（enabled_modules 运行时过滤）

**预估**：5-6 天

---

## CR-4: 插件系统统一 RBAC + 通信契约

**目标**：插件权限纳入 admin_resource/admin_api_permission、废弃 sys_menu、插件通信版本化

**范围**：
- 后端：
  - 实现 `PluginResourceSyncer`（SyncOnStart / SyncOnUninstall）
  - 插件安装时自动创建 `admin_application`（app_type=PLUGIN）
  - 插件 Manifest V2 解析（modules / exposedActions / platforms / frontends）
  - `CallPluginRequest` 增加 ActionVersion 校验
  - 废弃 `Installer.RegisterMenus`（sys_menu 逻辑）
  - PluginProxy 插件停止时返回 503
  - 插件 EventBus 订阅基于 Manifest.subscribedEvents 自动注册
- 前端：
  - 插件管理页适配新模型（状态感知矩阵）
  - 角色权限配置统一展示插件应用权限
  - 应用目录页（租户侧浏览/申请订阅入口）

**验收标准**：
- 插件启动后菜单/API 出现在 admin_resource/admin_api_permission
- 插件卸载后资源和角色绑定被清理
- 插件停止 → 前端展示"维护中"
- CallPlugin 版本不兼容返回明确错误
- 租户可在应用目录中看到可订阅应用

**依赖**：CR-1（应用模型 + AppResolveMiddleware）、CR-2（权限配置联动）

**预估**：6-7 天

---

## CR-5: 双用户池 + C 端认证 + 用户端权限

**目标**：biz_user 独立表、独立认证流程、简化权限模型

**范围**：
- DDL 新增：`biz_user` 表
- 后端：
  - `/auth/user/login`（sms/wechat 端点）
  - `/auth/user/register`（C端注册端点）
  - Token Claims 含 UserPool 字段
  - 中间件 pool=user 时跳过 PermissionMiddleware
  - C 端 GetUserMenu（返回租户订阅 × 已启用模块 × platform=user 的菜单）
  - 管理端 biz_user CRUD 接口（/api/v1/admin/biz-users）
  - BizUserProvider SPI（手机号/微信 openid 查询）
- 前端（dev-web-user 初始化）：
  - 基础框架搭建（Vue3 + Vite）
  - 登录页（手机验证码 / 微信授权）
  - 菜单加载 + 动态路由
  - 基础布局（PC + H5 双入口骨架）

**验收标准**：
- C 端注册 → biz_user 创建并绑定 tenant_id
- C 端 token pool=user → 跳过 RBAC
- C 端 GetUserMenu 正确返回 user 平台菜单
- 管理端可 CRUD C 端用户
- dev-web-user 可运行并加载菜单

**依赖**：CR-1（platform 字段）、CR-3（租户状态/TenantIsolationCallback）

**预估**：6-7 天

---

## CR-6: 多端前端架构 + Plugin SDK V2

**目标**：双前端支持 PC+H5、Plugin SDK 支持 teardown 和多端感知

**范围**：
- 前端（dev-web-admin）：
  - 工程改造支持 PC/H5 双入口（共享 src，不同布局层）
- 前端（dev-web-user）：
  - 工程改造支持 PC/H5 双入口
- 前端共享包：
  - Plugin SDK V2（teardown / platform / device / UnregisterFn / i18n.mergeLocale）
  - plugin-loader 实现（loadPlugin / unloadPlugin）
  - 插件 bundle 按 platform-device 路径加载
- 后端：
  - 插件安装时按 manifest.frontends 部署 bundle 到对应静态目录

**验收标准**：
- dev-web-admin PC 和 H5 均可运行
- dev-web-user PC 和 H5 均可运行
- 插件前端按 platform-device 正确加载
- teardown 在插件禁用/卸载时被调用
- registerExtension 返回 unregister 函数

**依赖**：CR-4（插件 Manifest V2）、CR-5（dev-web-user 基础框架）

**预估**：5-6 天

---

## CR-7: 审批流引擎

**目标**：轻量级审批流，支持单级/会签/或签，预留多级扩展

**范围**：
- DDL 新增：`admin_approval_flow`、`admin_approval`、`admin_approval_node`
- 后端：
  - 审批流程定义 CRUD
  - 发起审批 → 创建实例 + 按 flow_config 创建节点
  - 审批/驳回/撤销操作
  - 节点类型处理（SINGLE / AND_SIGN / OR_SIGN）
  - 审批人确定（USER / ROLE / DEPT_HEAD / APPLICANT_HEAD）
  - 审批完成后发布事件（approval.completed）→ 触发后续动作
  - 应用订阅审批流集成（subscription_mode=approval_required 时走审批）
  - 租户可覆盖全局流程定义
- 前端：
  - 审批管理页面（待我审批/我发起的）
  - 审批详情 + 审批/驳回操作
  - 应用订阅时触发审批流程的交互

**验收标准**：
- 发起审批 → 节点自动创建
- 单人审批通过 → 状态流转为 APPROVED
- 会签需所有人通过
- 或签任一人通过即可
- 应用订阅审批模式下，订阅需审批后生效
- 租户自定义流程覆盖全局默认

**依赖**：CR-1（应用订阅模型）、CR-3（OrganizationProvider）

**预估**：5-6 天

---

## CR-8: 缓存优化 + 认证策略 + 扩展收尾

**目标**：性能收尾、认证扩展性、杂项

**范围**：
- 后端缓存：
  - EventPublisher SPI 实现（进程内事件总线）
  - 权限变更时发布事件 → 缓存订阅失效
  - api_code_map 版本化替换
  - DynamicPermissionMiddleware 改用 L2 缓存
- 后端认证：
  - `AuthenticationStrategy` 接口 + `StrategyRouter`
  - `PasswordStrategy` 实现（重构现有逻辑）
  - `/auth/login` 统一入口 grant_type 路由
  - OAuth2Strategy 骨架（预留）
- 扩展收尾：
  - 操作日志增加 risk_level 字段 + 高风险操作标记
  - 日志分级可见性（SUPER_ADMIN 全部 / TENANT_ADMIN 本租户）
  - ext_fields JSON 列预留到关键业务表
  - admin_custom_field 表 DDL 预留（不实现逻辑）

**验收标准**：
- 权限变更后缓存 1s 内失效
- 缓存命中时权限检查 P99 < 5ms
- /auth/login grant_type 正确路由
- 现有密码登录不受影响
- 高风险操作日志可独立查询

**依赖**：CR-1 ~ CR-5 稳定

**预估**：5-6 天

---

## 时间线总览

```
Week 1-2:   CR-1 应用模型 + 中间件链（地基）
Week 2-3:   CR-2 权限体系增强
Week 3-4:   CR-3 数据权限 + 配置 + 配额    ← 可与 CR-2 后半段并行
Week 4-5:   CR-4 插件统一 RBAC
Week 5-6:   CR-5 双用户池 + C端
Week 6-7:   CR-6 多端前端
Week 7-8:   CR-7 审批流引擎              ← 可与 CR-6 并行
Week 8-9:   CR-8 缓存 + 认证策略 + 收尾
```

**总计约 8-9 周**。CR-7 与 CR-6 可并行（无依赖交叉）。

---

## 依赖关系图

```
CR-1 (地基)
 ├──▶ CR-2 (权限体系增强)
 ├──▶ CR-3 (数据权限 + 配置)
 │     └──▶ CR-5 (双用户池)    ← 依赖 CR-3 的 TenantIsolationCallback
 ├──▶ CR-4 (插件统一)          ← 依赖 CR-1 + CR-2
 │     └──▶ CR-6 (多端前端)    ← 依赖 CR-4 + CR-5
 ├──▶ CR-7 (审批流)            ← 依赖 CR-1 + CR-3（可与 CR-6 并行）
 └──▶ CR-8 (收尾)              ← 依赖全部稳定
```

---

## 每段 CR 交付标准

| 维度 | 要求 |
|------|------|
| 编译 | `go build ./...` 零错误 |
| 静态检查 | `go vet ./...` 零警告 |
| 功能测试 | 核心流程端到端可走通 |
| 回归 | 前序 CR 的功能不被破坏 |
| 文档 | 更新 design_v2.md 中实际实现细节（如有偏差） |

---

## design_v2.md 章节 → CR 映射

| 章节 | 对应 CR |
|------|--------|
| 1. 概述 | CR-1 |
| 2. 核心概念模型 | CR-1 建立，后续 CR 扩展 |
| 3. 认证体系（3.1-3.5） | CR-8 |
| 3.6 双用户池 | CR-5 |
| 4.1-4.5 中间件+角色继承 | CR-1(中间件) + CR-2(继承) |
| 4.6 数据权限 | CR-3 |
| 4.7 权限引擎 | CR-1(已有) |
| 4.8 字段级权限 | CR-2 |
| 4.9 记录共享 | CR-2 |
| 4.10 权限集 | CR-2 |
| 5. 多租户架构 | CR-1(框架) + CR-3(配置/配额) + CR-5(C端) |
| 6. 应用管理 | CR-1(模型) + CR-4(插件联动) |
| 7. 插件系统 | CR-4 |
| 8. 缓存策略 | CR-8 |
| 9. 数据库设计 | 各 CR 按需创建对应表 |
| 10. API 设计 | 各 CR 按需实现对应接口 |
| 11. 前端架构 | CR-6 |
| 12. 初始化与种子 | CR-1 |
| 13. SPI 接口 | CR-3(Org) + CR-8(Auth/Event) |
| 14. 数据隔离规范 | CR-3 |
| 15. 配额管理 | CR-3 |
| 16. 操作日志 | CR-8 |
| 17. 审批流引擎 | CR-7 |
| 18. 扩展能力 | CR-8 |
| 19. 通知与告警 | 后续独立 CR |
| 20. V1→V2 对照 | 文档维护 |
