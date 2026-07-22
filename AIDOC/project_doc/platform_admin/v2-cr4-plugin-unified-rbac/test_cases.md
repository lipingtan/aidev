# 测试用例：V2-CR4 插件系统统一 RBAC + 通信契约

**关联文档**: requirements.md / design.md / tasks.md
**编写日期**: 2025-07-15
**测试类型**: 功能测试 / 接口测试 / 回归测试 / 数据验证
**测试方法**: 等价类划分 + 边界值分析 + 状态转换测试 + 错误推测 + 场景法

---

## 测试矩阵概览

| 模块 | 正向 | 反向 | 边界 | 回归 | 接口 | 数据验证 | 合计 |
|------|------|------|------|------|------|----------|------|
| Manifest V2 解析 + 安装 | 5 | 4 | 3 | 1 | 8 | 3 | 24 |
| PluginResourceSyncer | 4 | 3 | 2 | 2 | 0 | 3 | 14 |
| CallPlugin 版本校验 | 3 | 3 | 2 | 1 | 0 | 0 | 9 |
| EventBus + Proxy | 3 | 2 | 1 | 2 | 0 | 0 | 8 |
| 租户应用订阅 | 4 | 3 | 2 | 1 | 16 | 2 | 28 |
| 前端交互 | 4 | 3 | 0 | 0 | 0 | 0 | 7 |
| 插件状态转换 | 3 | 2 | 0 | 0 | 0 | 2 | 7 |
| **总计** | **26** | **20** | **10** | **7** | **24** | **10** | **97** |

---

## 一、正向测试（Happy Path）

### TC-001: 插件安装自动创建 admin_application 记录

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 平台管理员已登录；合法插件包 billing.zip（含完整 plugin.json V2） |
| **测试步骤** | 1. POST /api/v1/admin/plugins/upload 上传 billing.zip <br> 2. 查询 admin_application WHERE app_code='billing' |
| **预期结果** | HTTP 200 安装成功；admin_application 含记录：app_code=billing, app_type=PLUGIN, platforms=["admin:pc","user:pc"], modules 含 invoice/payment |
| **关联需求** | FR-1 |
| **设计方法** | 场景法 |

### TC-002: plugin.json V2 完整字段解析

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | plugin.json 含 modules/platforms/frontends/exposedActions/subscribedEvents/menus/apiPermissions 全部字段 |
| **测试步骤** | 1. 调用 ParseManifest(pluginDir) <br> 2. 验证返回的 ManifestV2 各字段 |
| **预期结果** | ManifestV2.Modules 含 3 项；Platforms 含 3 项；ExposedActions 含 getInvoice；SubscribedEvents 含 order.created/payment.completed；Menus 含树形菜单；ApiPermissions 含 GROUP+ENDPOINT |
| **关联需求** | FR-2 |
| **设计方法** | 等价类（有效完整输入） |

### TC-003: plugin.json 缺少 displayName 时 fallback 到 name

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | plugin.json 仅含 name="billing"，无 displayName 字段 |
| **测试步骤** | 1. 调用 ParseManifest(pluginDir) |
| **预期结果** | ManifestV2.DisplayName == "billing"（fallback） |
| **关联需求** | FR-2 |
| **设计方法** | 等价类（缺省值） |

### TC-004: PluginResourceSyncer.SyncOnStart 正常同步

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 插件 billing 已安装；plugin.json 含 2 个菜单 + 3 个 API 权限 |
| **测试步骤** | 1. 启动插件 billing <br> 2. 查询 admin_resource WHERE app_code='billing' <br> 3. 查询 admin_api_permission WHERE app_code='billing' |
| **预期结果** | admin_resource 含 2 条记录（含 parent_id 递归）；admin_api_permission 含 3 条记录（含 GROUP 树形） |
| **关联需求** | FR-3 |
| **设计方法** | 场景法 |

### TC-005: SyncOnStart 幂等执行

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 插件 billing 已启动一次（admin_resource 有数据） |
| **测试步骤** | 1. 再次启动插件 billing（触发 SyncOnStart） <br> 2. 查询 admin_resource 记录数 |
| **预期结果** | 记录数不变（先清后写），无重复数据 |
| **关联需求** | FR-3 |
| **设计方法** | 错误推测（重复操作） |

### TC-006: CallPlugin 版本兼容正常转发

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件暴露 getInvoice action version="1.0"；billing 插件运行中 |
| **测试步骤** | 1. 调用 CallPlugin(target="billing", method="getInvoice", action_version="1.5") |
| **预期结果** | 正常转发调用成功（major=1 相同即兼容） |
| **关联需求** | FR-5 |
| **设计方法** | 等价类（兼容版本） |

### TC-007: CallPlugin 不带版本号向后兼容

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件运行中；旧版调用不含 action_version 字段 |
| **测试步骤** | 1. 调用 CallPlugin(target="billing", method="getInvoice", action_version="") |
| **预期结果** | 正常转发，不进行版本校验 |
| **关联需求** | FR-5 |
| **设计方法** | 等价类（向后兼容） |

### TC-008: EventBus 插件启动自动订阅事件

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件 plugin.json 含 subscribedEvents: ["order.created", "payment.completed"] |
| **测试步骤** | 1. 启动 billing 插件 <br> 2. 发布 order.created 事件 |
| **预期结果** | billing 插件收到 order.created 事件回调 |
| **关联需求** | FR-6 |
| **设计方法** | 场景法 |

### TC-009: 插件停止后事件订阅自动注销

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件运行中，已订阅 order.created |
| **测试步骤** | 1. 停止 billing 插件 <br> 2. 发布 order.created 事件 |
| **预期结果** | billing 插件不再收到事件广播 |
| **关联需求** | FR-6 |
| **设计方法** | 状态转换测试 |

### TC-010: PluginProxy 正常代理转发

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件运行中 |
| **测试步骤** | 1. 发送 GET /api/v1/billing/invoices |
| **预期结果** | 请求被正常代理到 billing 插件，返回业务数据 |
| **关联需求** | FR-7 |
| **设计方法** | 等价类（正向代理） |

### TC-011: 前端插件列表展示各状态

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 存在 3 个插件：billing(运行中)、report(已停止)、analytics(异常) |
| **测试步骤** | 1. 访问插件管理页 |
| **预期结果** | billing 显示绿色状态标签 + "停止"按钮；report 显示橙色标签 + "启动"按钮；analytics 显示红色标签 + "重启"按钮 |
| **关联需求** | FR-8 |
| **设计方法** | 等价类（状态枚举） |

### TC-012: 前端升级流程二次确认

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件 v1.0 运行中；新版本 v2.0 breakingUpgrade=true |
| **测试步骤** | 1. 上传 billing-v2.0.zip 升级 <br> 2. 后端返回 needConfirm=true + migrationNotes <br> 3. 前端弹窗展示迁移说明 <br> 4. 用户确认 <br> 5. 前端发送 confirm=true 再次请求 |
| **预期结果** | 弹窗含 migrationNotes 内容 + "⚠️ 升级期间插件不可用"提示；确认后升级执行成功 |
| **关联需求** | FR-8 |
| **设计方法** | 场景法（端到端升级流程） |

### TC-013: 租户管理员浏览应用目录

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 平台有 2 个 BUILTIN 应用 + 1 个 PLUGIN 应用(运行中)；当前租户已订阅 1 个 BUILTIN |
| **测试步骤** | 1. GET /api/v1/admin/app-catalog |
| **预期结果** | HTTP 200, 返回 3 个应用；已订阅的标注 subscribed=true；PLUGIN 类型含 pluginStatus 字段 |
| **关联需求** | FR-9 |
| **设计方法** | 场景法 |

### TC-014: 租户订阅应用

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件运行中且未被当前租户订阅；租户配额未满 |
| **测试步骤** | 1. POST /api/v1/admin/app-subscriptions body: {"app_code":"billing"} |
| **预期结果** | HTTP 201；admin_tenant_app 含 tenant_id + app_code=billing 记录；enabled_modules=NULL |
| **关联需求** | FR-9, FR-10 |
| **设计方法** | 场景法 |

### TC-015: 租户退订应用 + 级联清理

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 当前租户已订阅 billing；该租户角色已绑定 billing 的资源 |
| **测试步骤** | 1. DELETE /api/v1/admin/app-subscriptions/billing |
| **预期结果** | HTTP 200；admin_tenant_app 记录删除；admin_role_resource 中该租户引用 billing 资源的绑定被清理 |
| **关联需求** | FR-9, FR-10 |
| **设计方法** | 场景法（级联操作） |

### TC-016: 卸载插件级联清理全部数据

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件已安装并被多个租户订阅；角色已绑定资源 |
| **测试步骤** | 1. DELETE /api/v1/admin/plugins/billing |
| **预期结果** | admin_resource/admin_api_permission/admin_role_resource/admin_role_api/admin_tenant_app 全部清理；admin_application 软删除 |
| **关联需求** | FR-1, FR-3 |
| **设计方法** | 场景法（完整卸载流程） |

### TC-017: 废弃 RegisterMenus — 安装不写 sys_menu

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 全新安装插件 billing |
| **测试步骤** | 1. 安装并启动 billing 插件 <br> 2. 查询 sys_menu WHERE menu_name LIKE 'billing%' |
| **预期结果** | sys_menu 中无 billing 相关新记录 |
| **关联需求** | FR-4 |
| **设计方法** | 等价类（新行为验证） |

### TC-018: GetUserMenu 从 admin_resource 获取插件菜单

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件已启动；admin_resource 含插件菜单；用户角色已绑定该资源 |
| **测试步骤** | 1. 调用 GetUserMenu 接口 |
| **预期结果** | 返回的菜单树中含 billing 插件菜单（来源为 admin_resource，非 sys_menu） |
| **关联需求** | FR-4 |
| **设计方法** | 场景法 |

### TC-019: 事件仅广播到运行中的已订阅插件

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | billing(运行中,订阅order.created)；report(已停止,订阅order.created) |
| **测试步骤** | 1. 发布 order.created 事件 |
| **预期结果** | billing 收到事件；report 不收到事件 |
| **关联需求** | FR-6 |
| **设计方法** | 状态转换测试 |

### TC-020: 插件启动后 AutoDiscover 刷新

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | billing 插件安装后含 apiPermissions |
| **测试步骤** | 1. 启动 billing 插件 <br> 2. 访问 billing 的 API 端点（无权限用户） |
| **预期结果** | API 返回 403（说明 api_code_map 已刷新，权限中间件能识别新 API） |
| **关联需求** | FR-3 |
| **设计方法** | 场景法 |

### TC-021: 角色配置页统一展示 BUILTIN + PLUGIN 资源树

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 存在 BUILTIN 应用(platform_admin) + PLUGIN 应用(billing)；billing 运行中 |
| **测试步骤** | 1. 访问角色权限配置页 |
| **预期结果** | 资源树中无差别展示 platform_admin 和 billing 的菜单/权限节点 |
| **关联需求** | FR-8 |
| **设计方法** | 等价类 |

### TC-022: 模块配置功能

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 租户已订阅 billing（modules 含 invoice/payment/refund） |
| **测试步骤** | 1. PUT /api/v1/admin/app-subscriptions/billing/modules body: {"enabled_modules":["invoice","payment"]} |
| **预期结果** | HTTP 200；admin_tenant_app.enabled_modules 更新为 ["invoice","payment"] |
| **关联需求** | FR-10 |
| **设计方法** | 等价类 |

### TC-023: 完整生命周期流程 — 安装→启动→使用→升级→卸载

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 无已安装的 billing 插件 |
| **测试步骤** | 1. POST /plugins/upload 安装 billing v1.0 <br> 2. POST /plugins/billing/start 启动 <br> 3. 租户订阅 billing <br> 4. 分配 billing 资源给角色 <br> 5. PUT /plugins/billing/upgrade 升级到 v1.1 <br> 6. 验证租户订阅保留 <br> 7. DELETE /plugins/billing 卸载 <br> 8. 验证所有数据清理 |
| **预期结果** | 每步执行正确；升级后订阅保留；卸载后全部清理 |
| **关联需求** | FR-1, FR-3, FR-5, FR-10 |
| **设计方法** | 场景法（端到端） |

### TC-024: 升级保留租户订阅关系

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing v1.0 运行中；租户 A/B 已订阅 |
| **测试步骤** | 1. PUT /plugins/billing/upgrade 上传 v1.1 <br> 2. 查询 admin_tenant_app WHERE app_code='billing' |
| **预期结果** | 租户 A/B 的订阅记录保留不变；admin_application.id 不变 |
| **关联需求** | FR-1 |
| **设计方法** | 场景法 |

### TC-025: 插件停止后 admin_resource 记录保留

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | billing 插件运行中，admin_resource 含数据 |
| **测试步骤** | 1. POST /plugins/billing/stop <br> 2. 查询 admin_resource WHERE app_code='billing' |
| **预期结果** | admin_resource 记录仍存在（停止不清理资源） |
| **关联需求** | FR-3 |
| **设计方法** | 状态转换测试 |

### TC-026: syncApplicationMeta 更新元数据

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | billing 插件已安装；plugin.json 更新了 description 和 modules |
| **测试步骤** | 1. 重新启动 billing 插件（触发 SyncOnStart） <br> 2. 查询 admin_application WHERE app_code='billing' |
| **预期结果** | admin_application.description/modules/platforms 已更新为 plugin.json 最新值 |
| **关联需求** | FR-3 |
| **设计方法** | 等价类 |

---

## 二、反向测试（Negative）

### TC-N01: 同名插件安装被拒绝 — sys_plugin 冲突

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件已安装 |
| **测试步骤** | 1. POST /api/v1/admin/plugins/upload 再次上传 billing.zip |
| **预期结果** | HTTP 409/400, msg="插件名称 billing 与已有应用冲突，无法安装" |
| **关联需求** | FR-1 |
| **设计方法** | 错误推测（重复安装） |

### TC-N02: 同名插件安装被拒绝 — admin_application 冲突

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | admin_application 中已存在 app_code="billing"（BUILTIN 类型） |
| **测试步骤** | 1. POST /api/v1/admin/plugins/upload 上传 name=billing 的插件包 |
| **预期结果** | HTTP 409/400, msg="插件名称 billing 与已有应用冲突，无法安装" |
| **关联需求** | FR-1 |
| **设计方法** | 错误推测（名称冲突） |

### TC-N03: CallPlugin major 版本不匹配

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 暴露 getInvoice version="1.0" |
| **测试步骤** | 1. CallPlugin(target="billing", method="getInvoice", action_version="2.0") |
| **预期结果** | 返回 ErrActionVersionIncompatible 错误 |
| **关联需求** | FR-5 |
| **设计方法** | 等价类（无效 major） |

### TC-N04: CallPlugin action 不存在

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件运行中，未暴露 nonExistAction |
| **测试步骤** | 1. CallPlugin(target="billing", method="nonExistAction", action_version="1.0") |
| **预期结果** | 返回 ErrActionNotFound 错误 |
| **关联需求** | FR-5 |
| **设计方法** | 错误推测（无效 action） |

### TC-N05: CallPlugin 目标插件未运行

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件已停止 |
| **测试步骤** | 1. CallPlugin(target="billing", method="getInvoice", action_version="1.0") |
| **预期结果** | 返回 ErrPluginNotRunning / 503 错误 |
| **关联需求** | FR-5 |
| **设计方法** | 状态转换测试 |

### TC-N06: PluginProxy 插件未注册返回 404

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 无名为 ghost 的插件 |
| **测试步骤** | 1. GET /api/v1/ghost/anything |
| **预期结果** | HTTP 404, body: {"code":40400,"msg":"插件不存在"} |
| **关联需求** | FR-7 |
| **设计方法** | 错误推测 |

### TC-N07: PluginProxy 插件已停止返回 503

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 插件已停止 |
| **测试步骤** | 1. GET /api/v1/billing/invoices |
| **预期结果** | HTTP 503, body: {"code":50301,"msg":"插件维护中，请稍后再试","data":{"plugin":"billing","status":"stopped"}} |
| **关联需求** | FR-7 |
| **设计方法** | 状态转换测试 |

### TC-N08: 退订 BUILTIN 类型应用被拒绝

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | platform_admin 为 BUILTIN 类型应用 |
| **测试步骤** | 1. DELETE /api/v1/admin/app-subscriptions/platform_admin |
| **预期结果** | HTTP 400, body: {"code":40001,"msg":"内置应用不允许退订"} |
| **关联需求** | FR-10 |
| **设计方法** | 错误推测（禁止操作） |

### TC-N09: 订阅超配额被拒绝

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 租户 quota.max_apps=5；当前已订阅 5 个应用 |
| **测试步骤** | 1. POST /api/v1/admin/app-subscriptions body: {"app_code":"new_plugin"} |
| **预期结果** | HTTP 400/403, msg 含配额超限信息 |
| **关联需求** | FR-10 |
| **设计方法** | 边界值分析（上限） |

### TC-N10: plugin.json name 为空拒绝解析

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | plugin.json 中 name="" 或缺失 |
| **测试步骤** | 1. 调用 ParseManifest(pluginDir) |
| **预期结果** | 返回错误："name 字段不能为空" |
| **关联需求** | FR-2 |
| **设计方法** | 等价类（无效输入） |

### TC-N11: plugin.json 文件不存在

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 插件目录下无 plugin.json 文件 |
| **测试步骤** | 1. 调用 ParseManifest(emptyDir) |
| **预期结果** | 返回错误信息（文件不存在） |
| **关联需求** | FR-2 |
| **设计方法** | 错误推测（文件缺失） |

### TC-N12: 升级版本过低被拒绝（minUpgradeFrom）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing 当前 v0.5；新包 v2.0 minUpgradeFrom="1.0.0" |
| **测试步骤** | 1. PUT /api/v1/admin/plugins/billing/upgrade 上传 v2.0 |
| **预期结果** | HTTP 400, msg="当前版本过低，请先升级到 1.0.0 再执行此升级" |
| **关联需求** | FR-1 |
| **设计方法** | 边界值分析 |

### TC-N13: 升级包 name 与 URL :name 不一致

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | URL 为 /plugins/billing/upgrade；上传包 plugin.json.name="report" |
| **测试步骤** | 1. PUT /api/v1/admin/plugins/billing/upgrade 上传 report 包 |
| **预期结果** | HTTP 400, msg 含 name 不一致错误 |
| **关联需求** | FR-1 |
| **设计方法** | 错误推测 |

### TC-N14: SyncOnStart 事务中途 DB 错误回滚

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 模拟 DB 写入 admin_api_permission 时报错 |
| **测试步骤** | 1. 启动插件触发 SyncOnStart <br> 2. 检查 admin_resource 和 admin_api_permission |
| **预期结果** | 事务回滚，admin_resource 无残留数据（原子性保证） |
| **关联需求** | FR-3 |
| **设计方法** | 错误推测（事务回滚） |

### TC-N15: 升级失败自动回滚

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | billing v1.0 运行中；新版本 v1.1 启动失败 |
| **测试步骤** | 1. PUT /plugins/billing/upgrade 上传 v1.1（启动会失败） <br> 2. 查询 sys_plugin.version |
| **预期结果** | .bak 恢复为正式目录；sys_plugin.version 恢复为 1.0；返回错误信息 |
| **关联需求** | FR-1 |
| **设计方法** | 场景法（回滚流程） |

### TC-N16: 前端已停止插件权限节点红字提示

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | billing 插件已停止 |
| **测试步骤** | 1. 访问角色权限配置页 |
| **预期结果** | billing 应用节点名后追加红字"(插件已停止)"；权限分配操作不受影响 |
| **关联需求** | FR-8 |
| **设计方法** | 状态转换测试 |

### TC-N17: 订阅不存在的应用

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 无名为 ghost_app 的应用 |
| **测试步骤** | 1. POST /api/v1/admin/app-subscriptions body: {"app_code":"ghost_app"} |
| **预期结果** | HTTP 404/400, msg 含应用不存在错误 |
| **关联需求** | FR-10 |
| **设计方法** | 错误推测 |

### TC-N18: 重复订阅同一应用

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 租户已订阅 billing |
| **测试步骤** | 1. POST /api/v1/admin/app-subscriptions body: {"app_code":"billing"} |
| **预期结果** | HTTP 409/400, msg 含已订阅错误 |
| **关联需求** | FR-10 |
| **设计方法** | 错误推测（重复操作） |

### TC-N19: 卸载运行中的插件

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | billing 插件运行中 |
| **测试步骤** | 1. DELETE /api/v1/admin/plugins/billing |
| **预期结果** | 系统先停止插件再执行卸载流程（或返回需先停止的提示），最终卸载成功 |
| **关联需求** | FR-1 |
| **设计方法** | 错误推测 |

### TC-N20: ActionRegistry 空版本字符串

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **前置条件** | billing 暴露 getInvoice version="1.0" |
| **测试步骤** | 1. IsCompatible("billing", "getInvoice", "") |
| **预期结果** | 返回 false |
| **关联需求** | FR-5 |
| **设计方法** | 边界值分析 |

---

## 三、边界测试（Boundary）

### TC-B01: 配额边界 — 恰好达到上限

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | quota.max_apps=5；当前已订阅 4 个 |
| **测试步骤** | 1. POST /app-subscriptions 订阅第 5 个 |
| **预期结果** | HTTP 201，订阅成功（恰好达到上限） |
| **关联需求** | FR-10 |
| **设计方法** | 边界值分析（max） |

### TC-B02: 配额边界 — 超过上限 1 个

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | quota.max_apps=5；当前已订阅 5 个 |
| **测试步骤** | 1. POST /app-subscriptions 订阅第 6 个 |
| **预期结果** | HTTP 400/403，配额超限拒绝 |
| **关联需求** | FR-10 |
| **设计方法** | 边界值分析（max+1） |

### TC-B03: 版本号边界 — major 版本 0

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | 注册 action version="0.1"；请求 version="0.9" |
| **测试步骤** | 1. IsCompatible("billing", "getInvoice", "0.9") |
| **预期结果** | true（major=0 相同） |
| **关联需求** | FR-5 |
| **设计方法** | 边界值分析（最小 major） |

### TC-B04: 菜单层级深度 — 3 级嵌套

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | plugin.json menus 含 3 级 children 嵌套 |
| **测试步骤** | 1. 安装并启动插件 <br> 2. 查询 admin_resource |
| **预期结果** | 所有 3 级菜单正确写入，parent_id 链路正确 |
| **关联需求** | FR-3 |
| **设计方法** | 边界值分析（嵌套深度） |

### TC-B05: 大量菜单 + API 性能边界

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | plugin.json 含 50 个菜单 + 100 个 API 权限 |
| **测试步骤** | 1. 启动插件触发 SyncOnStart <br> 2. 记录执行时间 |
| **预期结果** | SyncOnStart 在单事务中完成，耗时 < 2s |
| **关联需求** | FR-3（非功能需求） |
| **设计方法** | 边界值分析（性能边界） |

### TC-B06: 插件名称长度边界

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | name="a"（1 字符） / name="a"*128（128 字符） |
| **测试步骤** | 1. ParseManifest 解析 1 字符名称 <br> 2. ParseManifest 解析 128 字符名称 |
| **预期结果** | 1 字符成功；128 字符根据 DB 字段限制判定（超长则拒绝） |
| **关联需求** | FR-2 |
| **设计方法** | 边界值分析 |

### TC-B07: modules 为空数组

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | plugin.json modules=[] |
| **测试步骤** | 1. ParseManifest 解析 <br> 2. 安装插件 |
| **预期结果** | 安装成功；admin_application.modules 为空数组 |
| **关联需求** | FR-2 |
| **设计方法** | 边界值分析（空集合） |

### TC-B08: exposedActions 为空 — 无 action 暴露

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | plugin.json exposedActions=[] |
| **测试步骤** | 1. 启动插件 <br> 2. CallPlugin 调用任何 action |
| **预期结果** | 启动正常；CallPlugin 返回 ErrActionNotFound |
| **关联需求** | FR-5 |
| **设计方法** | 边界值分析（空集合） |

### TC-B09: subscribedEvents 为空 — 无事件订阅

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | plugin.json subscribedEvents=[] |
| **测试步骤** | 1. 启动插件 <br> 2. 发布任何事件 |
| **预期结果** | 启动正常；不收到任何事件广播 |
| **关联需求** | FR-6 |
| **设计方法** | 边界值分析（空集合） |

### TC-B10: 并发安装同名插件

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 两个请求同时上传 name="billing" 的插件包 |
| **测试步骤** | 1. 并发发送两个 POST /plugins/upload（同名包） |
| **预期结果** | 仅一个成功，另一个返回冲突错误；无脏数据 |
| **关联需求** | FR-1 |
| **设计方法** | 错误推测（并发竞态） |

---

## 四、回归测试（Regression）

> 回归测试定义：验证本次 CR 改动**未破坏**已有功能。
> 执行范围：每次代码变更后运行本章节全部用例。

### TC-R01: 现有插件安装/卸载文件操作不受影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-1 |
| **验证步骤** | 1. 上传合法插件 zip <br> 2. 验证正确解压到 plugins/ 目录 <br> 3. 卸载后验证文件目录被删除 |
| **预期结果** | 文件操作行为与改动前完全一致 |
| **设计方法** | 回归验证 |

### TC-R02: 现有插件 gRPC 通信正常

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-2 |
| **验证步骤** | 1. 启动已有插件 <br> 2. 调用 Plugin.Register() 获取 PluginInfo <br> 3. 发送 HandleRequest 验证响应 |
| **预期结果** | gRPC 通信正常，返回数据无异常 |
| **设计方法** | 回归验证 |

### TC-R03: 现有角色权限配置页 BUILTIN 应用展示不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-3 |
| **验证步骤** | 1. 访问角色权限配置页 <br> 2. 查看 platform_admin（BUILTIN）应用资源树 |
| **预期结果** | 资源树展示结构、权限节点与改动前完全一致 |
| **设计方法** | 回归验证 |

### TC-R04: 现有租户订阅逻辑不被破坏

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-4 |
| **验证步骤** | 1. 已订阅租户访问系统 <br> 2. 验证可正常使用已订阅应用的功能 |
| **预期结果** | admin_tenant_app 原有记录有效，功能正常 |
| **设计方法** | 回归验证 |

### TC-R05: 现有路由注册不受影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-5 |
| **验证步骤** | 1. 访问所有原有 /api/v1/admin/ 接口（auth、role、tenant 等） |
| **预期结果** | 所有原有接口正常响应，状态码/功能无变化 |
| **设计方法** | 回归验证 |

### TC-R06: 现有 AutoDiscover 逻辑不受影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-6 |
| **验证步骤** | 1. 系统启动 <br> 2. 验证内置应用的 api_code_map 正确生成 <br> 3. 访问内置应用受保护 API（无权限用户） |
| **预期结果** | 返回 403（说明权限中间件正常拦截） |
| **设计方法** | 回归验证 |

### TC-R07: EventBus 现有事件广播行为不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-7 |
| **验证步骤** | 1. 使用不带 action_version 的 CallPlugin 调用 <br> 2. 使用 PublishEvent 广播事件 |
| **预期结果** | 旧版 CallPlugin 正常转发；PublishEvent 正常广播到已订阅且运行中的插件 |
| **设计方法** | 回归验证 |

---

## 五、接口测试（API Level）

### 5.1 插件管理接口

#### TC-A01: GET /api/v1/admin/plugins — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/plugins` |
| **Headers** | Authorization: Bearer {platform_admin_token} |
| **预期响应** | HTTP 200, `{"code":0,"data":[{"name":"billing","status":"running","platforms":["admin:pc"],"modules":[...]}]}` |
| **关联需求** | FR-8 |

#### TC-A02: GET /api/v1/admin/plugins — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/plugins`（无 Authorization） |
| **预期响应** | HTTP 401, code=40101 |
| **关联需求** | FR-8 |

#### TC-A03: GET /api/v1/admin/plugins — 无权限

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/plugins`（普通用户 token，无 platform:plugin:list） |
| **预期响应** | HTTP 403, code=40301 |
| **关联需求** | FR-8 |

#### TC-A04: POST /api/v1/admin/plugins/upload — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/plugins/upload` multipart/form-data file=billing.zip |
| **Headers** | Authorization: Bearer {platform_admin_token} |
| **预期响应** | HTTP 200, `{"code":0,"data":{"name":"billing","version":"1.0.0"}}` |
| **关联需求** | FR-1 |

#### TC-A05: POST /api/v1/admin/plugins/upload — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/plugins/upload`（无 Authorization） |
| **预期响应** | HTTP 401, code=40101 |
| **关联需求** | FR-1 |

#### TC-A06: POST /api/v1/admin/plugins/upload — 无权限

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/plugins/upload`（无 platform:plugin:install） |
| **预期响应** | HTTP 403, code=40301 |
| **关联需求** | FR-1 |

#### TC-A07: POST /api/v1/admin/plugins/upload — 无效文件

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/plugins/upload` file=invalid.txt（非 zip） |
| **预期响应** | HTTP 400, msg 含文件格式错误 |
| **关联需求** | FR-1 |

#### TC-A08: POST /api/v1/admin/plugins/:name/start — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/plugins/billing/start` |
| **Headers** | Authorization: Bearer {platform_admin_token} |
| **预期响应** | HTTP 200, `{"code":0,"msg":"启动成功"}` |
| **关联需求** | FR-8 |

#### TC-A09: POST /api/v1/admin/plugins/:name/start — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/plugins/billing/start`（无 Authorization） |
| **预期响应** | HTTP 401, code=40101 |
| **关联需求** | FR-8 |

#### TC-A10: POST /api/v1/admin/plugins/:name/start — 无权限

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/plugins/billing/start`（无 platform:plugin:start） |
| **预期响应** | HTTP 403, code=40301 |
| **关联需求** | FR-8 |

#### TC-A11: POST /api/v1/admin/plugins/:name/stop — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/plugins/billing/stop` |
| **Headers** | Authorization: Bearer {platform_admin_token} |
| **预期响应** | HTTP 200, `{"code":0,"msg":"停止成功"}` |
| **关联需求** | FR-8 |

#### TC-A12: DELETE /api/v1/admin/plugins/:name — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `DELETE /api/v1/admin/plugins/billing` |
| **Headers** | Authorization: Bearer {platform_admin_token} |
| **预期响应** | HTTP 200, `{"code":0,"msg":"卸载成功"}` |
| **关联需求** | FR-1 |

#### TC-A13: DELETE /api/v1/admin/plugins/:name — 无权限

| 字段 | 内容 |
|------|------|
| **请求** | `DELETE /api/v1/admin/plugins/billing`（无 platform:plugin:uninstall） |
| **预期响应** | HTTP 403, code=40301 |
| **关联需求** | FR-1 |

#### TC-A14: PUT /api/v1/admin/plugins/:name/upgrade — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/plugins/billing/upgrade` multipart file=billing-v1.1.zip |
| **Headers** | Authorization: Bearer {platform_admin_token} |
| **预期响应** | HTTP 200, `{"code":0,"data":{"old_version":"1.0.0","new_version":"1.1.0"}}` |
| **关联需求** | FR-1 |

#### TC-A15: PUT /api/v1/admin/plugins/:name/upgrade — breakingUpgrade 返回确认

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/plugins/billing/upgrade` file=billing-v2.0.zip（breakingUpgrade=true） |
| **预期响应** | HTTP 200, `{"code":0,"data":{"needConfirm":true,"migrationNotes":"V2 重构了订单表结构..."}}` |
| **关联需求** | FR-1, FR-8 |

#### TC-A16: GET /api/v1/admin/plugins/:name/health — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/plugins/billing/health` |
| **Headers** | Authorization: Bearer {platform_admin_token} |
| **预期响应** | HTTP 200, `{"code":0,"data":{"status":"running","uptime":"..."}}` |
| **关联需求** | FR-8 |

### 5.2 应用目录 + 订阅接口

#### TC-A17: GET /api/v1/admin/app-catalog — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/app-catalog` |
| **Headers** | Authorization: Bearer {tenant_admin_token} |
| **预期响应** | HTTP 200, data 含 BUILTIN + PLUGIN 应用列表，每项含 subscribed/pluginStatus 字段 |
| **关联需求** | FR-9, FR-10 |

#### TC-A18: GET /api/v1/admin/app-catalog — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/app-catalog`（无 Authorization） |
| **预期响应** | HTTP 401, code=40101（需要登录但免权限检查） |
| **关联需求** | FR-10 |

#### TC-A19: POST /api/v1/admin/app-subscriptions — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/app-subscriptions` |
| **Headers** | Authorization: Bearer {tenant_admin_token} |
| **Body** | `{"app_code":"billing"}` |
| **预期响应** | HTTP 201, `{"code":0,"msg":"订阅成功"}` |
| **关联需求** | FR-10 |

#### TC-A20: POST /api/v1/admin/app-subscriptions — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/app-subscriptions`（无 Authorization） |
| **预期响应** | HTTP 401, code=40101 |
| **关联需求** | FR-10 |

#### TC-A21: POST /api/v1/admin/app-subscriptions — 无权限

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/app-subscriptions`（无 tenant:app:subscribe） |
| **预期响应** | HTTP 403, code=40301 |
| **关联需求** | FR-10 |

#### TC-A22: POST /api/v1/admin/app-subscriptions — 无效 body

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/app-subscriptions` body: `{}` |
| **预期响应** | HTTP 400, msg 含 app_code 必填 |
| **关联需求** | FR-10 |

#### TC-A23: DELETE /api/v1/admin/app-subscriptions/:app_code — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `DELETE /api/v1/admin/app-subscriptions/billing` |
| **Headers** | Authorization: Bearer {tenant_admin_token} |
| **预期响应** | HTTP 200, `{"code":0,"msg":"退订成功"}` |
| **关联需求** | FR-10 |

#### TC-A24: DELETE /api/v1/admin/app-subscriptions/:app_code — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `DELETE /api/v1/admin/app-subscriptions/billing`（无 Authorization） |
| **预期响应** | HTTP 401, code=40101 |
| **关联需求** | FR-10 |

#### TC-A25: DELETE /api/v1/admin/app-subscriptions/:app_code — 无权限

| 字段 | 内容 |
|------|------|
| **请求** | `DELETE /api/v1/admin/app-subscriptions/billing`（无 tenant:app:unsubscribe） |
| **预期响应** | HTTP 403, code=40301 |
| **关联需求** | FR-10 |

#### TC-A26: GET /api/v1/admin/app-subscriptions — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/app-subscriptions` |
| **Headers** | Authorization: Bearer {tenant_admin_token} |
| **预期响应** | HTTP 200, data 含当前租户已订阅应用列表 |
| **关联需求** | FR-10 |

#### TC-A27: GET /api/v1/admin/app-subscriptions — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/app-subscriptions`（无 Authorization） |
| **预期响应** | HTTP 401, code=40101 |
| **关联需求** | FR-10 |

#### TC-A28: PUT /api/v1/admin/app-subscriptions/:app_code/modules — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/app-subscriptions/billing/modules` |
| **Headers** | Authorization: Bearer {tenant_admin_token} |
| **Body** | `{"enabled_modules":["invoice","payment"]}` |
| **预期响应** | HTTP 200, `{"code":0,"msg":"模块配置更新成功"}` |
| **关联需求** | FR-10 |

#### TC-A29: PUT /api/v1/admin/app-subscriptions/:app_code/modules — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/app-subscriptions/billing/modules`（无 Authorization） |
| **预期响应** | HTTP 401, code=40101 |
| **关联需求** | FR-10 |

#### TC-A30: PUT /api/v1/admin/app-subscriptions/:app_code/modules — 无权限

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/app-subscriptions/billing/modules`（无 tenant:app:subscribe） |
| **预期响应** | HTTP 403, code=40301 |
| **关联需求** | FR-10 |

---

## 六、数据验证

### TC-D01: 安装后 admin_application 数据正确

| 字段 | 内容 |
|------|------|
| **触发操作** | 安装 billing 插件（含完整 plugin.json V2） |
| **验证 SQL** | ```sql SELECT app_code, app_type, name, route_prefix, platforms, modules FROM admin_application WHERE app_code = 'billing' AND deleted_at IS NULL``` |
| **预期结果** | app_type='PLUGIN', name='计费管理', route_prefix='/api/v1/billing', platforms 含 admin:pc, modules 含 invoice/payment/refund |
| **关联需求** | FR-1 |

### TC-D02: SyncOnStart 后 admin_resource 树形正确

| 字段 | 内容 |
|------|------|
| **触发操作** | 启动 billing 插件 |
| **验证 SQL** | ```sql SELECT id, parent_id, name, permission_code, app_code, resource_type FROM admin_resource WHERE app_code = 'billing' ORDER BY parent_id, sort``` |
| **预期结果** | 顶层菜单 parent_id=NULL；子按钮 parent_id 指向父菜单 ID；resource_type 正确（menu/button） |
| **关联需求** | FR-3 |

### TC-D03: SyncOnStart 后 admin_api_permission 树形正确

| 字段 | 内容 |
|------|------|
| **触发操作** | 启动 billing 插件 |
| **验证 SQL** | ```sql SELECT id, parent_id, type, name, permission_code, url_pattern, http_method, app_code FROM admin_api_permission WHERE app_code = 'billing'``` |
| **预期结果** | GROUP 类型 parent_id=NULL；ENDPOINT 类型 parent_id 指向 GROUP；permission_code/url_pattern/http_method 与 plugin.json 一致 |
| **关联需求** | FR-3 |

### TC-D04: 卸载后全表清理验证

| 字段 | 内容 |
|------|------|
| **触发操作** | 卸载 billing 插件 |
| **验证 SQL** | ```sql SELECT COUNT(*) FROM admin_resource WHERE app_code = 'billing'; SELECT COUNT(*) FROM admin_api_permission WHERE app_code = 'billing'; SELECT COUNT(*) FROM admin_role_resource WHERE resource_id IN (原 billing 资源 IDs); SELECT COUNT(*) FROM admin_tenant_app WHERE app_code = 'billing'; SELECT deleted_at FROM admin_application WHERE app_code = 'billing';``` |
| **预期结果** | 前 4 个 COUNT=0；admin_application.deleted_at 非空 |
| **关联需求** | FR-3 |

### TC-D05: sys_menu 旧数据清理验证

| 字段 | 内容 |
|------|------|
| **触发操作** | 系统升级后首次启动 |
| **验证 SQL** | ```sql SELECT COUNT(*) FROM sys_menu WHERE menu_name LIKE 'plugin_%' AND deleted_at IS NULL``` |
| **预期结果** | COUNT=0（所有 plugin_ 前缀菜单已被软删除） |
| **关联需求** | FR-4 |

### TC-D06: 订阅后 admin_tenant_app 数据正确

| 字段 | 内容 |
|------|------|
| **触发操作** | 租户 A 订阅 billing |
| **验证 SQL** | ```sql SELECT tenant_id, app_code, enabled_modules FROM admin_tenant_app WHERE tenant_id = ? AND app_code = 'billing'``` |
| **预期结果** | 记录存在；enabled_modules=NULL（全部启用） |
| **关联需求** | FR-10 |

### TC-D07: 模块配置更新后数据正确

| 字段 | 内容 |
|------|------|
| **触发操作** | PUT /app-subscriptions/billing/modules body: {"enabled_modules":["invoice"]} |
| **验证 SQL** | ```sql SELECT enabled_modules FROM admin_tenant_app WHERE tenant_id = ? AND app_code = 'billing'``` |
| **预期结果** | enabled_modules = '["invoice"]' |
| **关联需求** | FR-10 |

### TC-D08: 退订级联清理角色绑定验证

| 字段 | 内容 |
|------|------|
| **触发操作** | 租户 A 退订 billing（该租户角色已绑定 billing 资源 R1/R2） |
| **验证 SQL** | ```sql SELECT COUNT(*) FROM admin_role_resource rr JOIN admin_role r ON rr.role_id = r.id WHERE r.tenant_id = ? AND rr.resource_id IN (SELECT id FROM admin_resource WHERE app_code = 'billing')``` |
| **预期结果** | COUNT=0（该租户下引用 billing 资源的角色绑定全部清理） |
| **关联需求** | FR-10 |

### TC-D09: 升级后 admin_application.id 不变

| 字段 | 内容 |
|------|------|
| **触发操作** | 升级 billing v1.0 → v1.1 |
| **验证 SQL** | ```sql SELECT id, app_code FROM admin_application WHERE app_code = 'billing' AND deleted_at IS NULL``` |
| **预期结果** | ID 与升级前相同（升级不改变 application ID） |
| **关联需求** | FR-1 |

### TC-D10: SyncOnStart 孤儿角色绑定清理

| 字段 | 内容 |
|------|------|
| **触发操作** | 插件升级后菜单减少（v1.0 有 R1/R2/R3，v1.1 只有 R1/R2），角色绑定了 R3 |
| **验证 SQL** | ```sql SELECT COUNT(*) FROM admin_role_resource WHERE resource_id NOT IN (SELECT id FROM admin_resource)``` |
| **预期结果** | COUNT=0（孤儿绑定已被清理） |
| **关联需求** | FR-3 |

---

## 插件状态转换覆盖矩阵

| 编号 | 起始状态 | 触发动作 | 目标状态 | 对应用例 |
|------|----------|----------|----------|----------|
| ST-1 | 未安装 | 安装 | 已安装 | TC-001 |
| ST-2 | 已安装 | 启动成功 | 运行中 | TC-004, TC-A08 |
| ST-3 | 已安装 | 启动失败 | 异常 | TC-N14 |
| ST-4 | 运行中 | 停止 | 已停止 | TC-009, TC-A11 |
| ST-5 | 已停止 | 重新启动 | 运行中 | TC-005 |
| ST-6 | 运行中 | 卸载 | 已卸载 | TC-016 |
| ST-7 | 已停止 | 卸载 | 已卸载 | TC-A12 |
| ST-8（非法） | 运行中 | 安装同名 | 拒绝 | TC-N01 |

---

## 执行结果记录（测试执行时填写）

| 用例编号 | 结果 | 执行人 | 日期 | 备注 |
|----------|------|--------|------|------|
| TC-001 | ⬜ | | | |
| TC-002 | ⬜ | | | |
| TC-003 | ⬜ | | | |
| TC-004 | ⬜ | | | |
| TC-005 | ⬜ | | | |
| TC-006 | ⬜ | | | |
| TC-007 | ⬜ | | | |
| TC-008 | ⬜ | | | |
| TC-009 | ⬜ | | | |
| TC-010 | ⬜ | | | |
| TC-011 | ⬜ | | | |
| TC-012 | ⬜ | | | |
| TC-013 | ⬜ | | | |
| TC-014 | ⬜ | | | |
| TC-015 | ⬜ | | | |
| TC-016 | ⬜ | | | |
| TC-017 | ⬜ | | | |
| TC-018 | ⬜ | | | |
| TC-019 | ⬜ | | | |
| TC-020 | ⬜ | | | |
| TC-021 | ⬜ | | | |
| TC-022 | ⬜ | | | |
| TC-023 | ⬜ | | | |
| TC-024 | ⬜ | | | |
| TC-025 | ⬜ | | | |
| TC-026 | ⬜ | | | |
| TC-N01 | ⬜ | | | |
| TC-N02 | ⬜ | | | |
| TC-N03 | ⬜ | | | |
| TC-N04 | ⬜ | | | |
| TC-N05 | ⬜ | | | |
| TC-N06 | ⬜ | | | |
| TC-N07 | ⬜ | | | |
| TC-N08 | ⬜ | | | |
| TC-N09 | ⬜ | | | |
| TC-N10 | ⬜ | | | |
| TC-N11 | ⬜ | | | |
| TC-N12 | ⬜ | | | |
| TC-N13 | ⬜ | | | |
| TC-N14 | ⬜ | | | |
| TC-N15 | ⬜ | | | |
| TC-N16 | ⬜ | | | |
| TC-N17 | ⬜ | | | |
| TC-N18 | ⬜ | | | |
| TC-N19 | ⬜ | | | |
| TC-N20 | ⬜ | | | |
| TC-B01 | ⬜ | | | |
| TC-B02 | ⬜ | | | |
| TC-B03 | ⬜ | | | |
| TC-B04 | ⬜ | | | |
| TC-B05 | ⬜ | | | |
| TC-B06 | ⬜ | | | |
| TC-B07 | ⬜ | | | |
| TC-B08 | ⬜ | | | |
| TC-B09 | ⬜ | | | |
| TC-B10 | ⬜ | | | |
| TC-R01 | ⬜ | | | |
| TC-R02 | ⬜ | | | |
| TC-R03 | ⬜ | | | |
| TC-R04 | ⬜ | | | |
| TC-R05 | ⬜ | | | |
| TC-R06 | ⬜ | | | |
| TC-R07 | ⬜ | | | |
| TC-A01 | ⬜ | | | |
| TC-A02 | ⬜ | | | |
| TC-A03 | ⬜ | | | |
| TC-A04 | ⬜ | | | |
| TC-A05 | ⬜ | | | |
| TC-A06 | ⬜ | | | |
| TC-A07 | ⬜ | | | |
| TC-A08 | ⬜ | | | |
| TC-A09 | ⬜ | | | |
| TC-A10 | ⬜ | | | |
| TC-A11 | ⬜ | | | |
| TC-A12 | ⬜ | | | |
| TC-A13 | ⬜ | | | |
| TC-A14 | ⬜ | | | |
| TC-A15 | ⬜ | | | |
| TC-A16 | ⬜ | | | |
| TC-A17 | ⬜ | | | |
| TC-A18 | ⬜ | | | |
| TC-A19 | ⬜ | | | |
| TC-A20 | ⬜ | | | |
| TC-A21 | ⬜ | | | |
| TC-A22 | ⬜ | | | |
| TC-A23 | ⬜ | | | |
| TC-A24 | ⬜ | | | |
| TC-A25 | ⬜ | | | |
| TC-A26 | ⬜ | | | |
| TC-A27 | ⬜ | | | |
| TC-A28 | ⬜ | | | |
| TC-A29 | ⬜ | | | |
| TC-A30 | ⬜ | | | |
| TC-D01 | ⬜ | | | |
| TC-D02 | ⬜ | | | |
| TC-D03 | ⬜ | | | |
| TC-D04 | ⬜ | | | |
| TC-D05 | ⬜ | | | |
| TC-D06 | ⬜ | | | |
| TC-D07 | ⬜ | | | |
| TC-D08 | ⬜ | | | |
| TC-D09 | ⬜ | | | |
| TC-D10 | ⬜ | | | |
