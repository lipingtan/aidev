# 需求：V2-CR4 插件系统统一 RBAC + 通信契约

## 背景

当前插件系统的菜单注册依赖 `sys_menu` 旧表，权限管理与 V2 的 `admin_resource` / `admin_api_permission` 体系脱节。插件的权限无法在统一的角色配置页中管理，也无法享受 V2 的 AppResolveMiddleware、权限继承、数据权限等能力。

本 CR 将插件权限纳入统一 RBAC 体系，同时升级插件通信契约（Manifest V2 + Action 版本化），为后续多端前端架构（CR-6）奠定基础。

## 用户故事

- 作为**平台管理员**，我希望安装插件后它的菜单和 API 权限自动出现在统一的权限配置体系中，以便用同一套角色配置管理所有应用权限
- 作为**租户管理员**，我希望能浏览平台可用的应用列表并申请订阅，以便按需开通功能模块
- 作为**插件开发者**，我希望通过 plugin.json 声明插件的模块、平台、暴露动作和事件订阅，以便 Host 自动完成资源注册和事件分发
- 作为**插件开发者**，我希望调用其他插件时有版本兼容性校验，以便避免因接口不兼容导致运行时错误

## 功能需求

### FR-1: 插件安装自动创建 admin_application

**描述：** 插件安装时，Host 解析 plugin.json（Manifest V2），自动在 `admin_application` 表创建一条 `app_type=PLUGIN` 的记录。

**验收标准：**
- WHEN 插件安装成功 THEN 系统 SHALL 在 admin_application 表创建记录（app_code=插件 name，app_type=PLUGIN，platforms/modules/route_prefix 从 plugin.json 读取）
- WHEN 插件卸载 THEN 系统 SHALL 软删除对应的 admin_application 记录
- WHEN 同名插件已存在 THEN 系统 SHALL 拒绝安装并返回错误信息

### FR-2: plugin.json Manifest V2 格式

**描述：** 扩展 plugin.json 格式，新增 V2 描述性字段。proto 层面保持不变（仅用于运行时 RPC），所有元数据描述由 plugin.json 承载。

**验收标准：**
- WHEN plugin.json 包含 modules 字段 THEN 系统 SHALL 解析并写入 admin_application.modules
- WHEN plugin.json 包含 platforms 字段 THEN 系统 SHALL 解析并写入 admin_application.platforms
- WHEN plugin.json 包含 frontends 字段 THEN 系统 SHALL 按 platform-device 路径部署前端 bundle
- WHEN plugin.json 包含 exposedActions 字段 THEN 系统 SHALL 注册到 ActionRegistry 供 CallPlugin 版本校验使用
- WHEN plugin.json 包含 subscribedEvents 字段 THEN 系统 SHALL 在插件启动时自动注册 EventBus 订阅
- WHEN plugin.json 包含 menus 字段 THEN 系统 SHALL 用于 PluginResourceSyncer 同步到 admin_resource（替代 proto.PluginInfo.Menus）
- WHEN plugin.json 包含 apiPermissions 字段 THEN 系统 SHALL 用于同步到 admin_api_permission

### FR-3: PluginResourceSyncer 实现

**描述：** 插件启动时，将 plugin.json 中声明的菜单和 API 权限同步到 `admin_resource` 和 `admin_api_permission` 表。

**验收标准：**
- WHEN 插件启动 THEN 系统 SHALL 调用 SyncOnStart，将菜单写入 admin_resource（app_code=插件 name）
- WHEN 插件启动 THEN 系统 SHALL 将 API 权限写入 admin_api_permission（app_code=插件 name）
- WHEN 插件重复启动 THEN 系统 SHALL 幂等执行（先清后写），不产生重复数据
- WHEN 插件卸载 THEN 系统 SHALL 调用 SyncOnUninstall，删除资源 + 级联清理角色绑定（admin_role_resource/admin_role_api）
- WHEN 同步完成 THEN 系统 SHALL 触发 AutoDiscover 刷新 api_code_map（确保新 API 可被权限中间件识别）

### FR-4: 废弃 Installer.RegisterMenus

**描述：** 移除向 sys_menu 注册插件菜单的旧逻辑，改由 PluginResourceSyncer 写入 admin_resource。

**验收标准：**
- WHEN 插件安装/启动 THEN 系统 SHALL NOT 向 sys_menu 表写入任何记录
- WHEN GetUserMenu 被调用 THEN 系统 SHALL 从 admin_resource 获取插件菜单（而非 sys_menu）
- WHEN 升级后首次启动 THEN 系统 SHALL 对 sys_menu 中已有的插件菜单执行软删除清理

### FR-5: CallPlugin Action 版本兼容性校验

**描述：** 插件间调用时，校验目标插件暴露的 Action 版本与请求版本的兼容性（major 版本相同即兼容）。

**验收标准：**
- WHEN CallPlugin 请求的 actionVersion major 版本与目标插件 exposedAction 的 major 版本相同 THEN 系统 SHALL 正常转发调用
- WHEN major 版本不匹配 THEN 系统 SHALL 返回 ErrActionVersionIncompatible 错误
- WHEN 目标插件未暴露请求的 action THEN 系统 SHALL 返回 ErrActionNotFound 错误
- WHEN 目标插件未运行 THEN 系统 SHALL 返回 503 错误

### FR-6: EventBus 基于 Manifest 自动订阅

**描述：** 插件启动时，Host 从 plugin.json 的 subscribedEvents 读取事件类型列表，自动注册到 EventBus。

**验收标准：**
- WHEN 插件启动且 plugin.json 含 subscribedEvents THEN 系统 SHALL 自动调用 EventBus.Subscribe 注册订阅
- WHEN 插件停止 THEN 系统 SHALL 注销该插件的所有事件订阅
- WHEN 事件发布 THEN 系统 SHALL 仅广播到已订阅且运行中的插件

### FR-7: PluginProxy 状态感知增强

**描述：** 插件停止时，PluginProxy 对该插件的 HTTP 请求返回 503 + 结构化错误信息。

**验收标准：**
- WHEN 插件处于停止状态且收到 HTTP 请求 THEN 系统 SHALL 返回 HTTP 503，body 含 `{"code":50301,"msg":"插件维护中，请稍后再试","data":{"plugin":"xxx","status":"stopped"}}`
- WHEN 插件未注册（不存在）THEN 系统 SHALL 返回 HTTP 404

### FR-8: 前端 — 插件管理页改造（平台侧）

**描述：** 改造现有插件管理页面，适配新的应用模型和状态感知矩阵。

**验收标准：**
- WHEN 查看插件列表 THEN 系统 SHALL 展示插件状态（已安装/运行中/已停止/异常）+ 所属模块 + 平台支持
- WHEN 插件运行中 THEN 系统 SHALL 显示"停止"按钮
- WHEN 插件已停止 THEN 系统 SHALL 显示"启动"按钮 + 状态标签为橙色
- WHEN 插件异常 THEN 系统 SHALL 显示"重启"按钮 + 状态标签为红色
- WHEN 角色权限配置页加载 THEN 系统 SHALL 统一展示 BUILTIN + PLUGIN 应用的资源树（无差别）
- WHEN 插件已停止 THEN 角色权限配置页中该插件的权限节点 SHALL 显示红字提示"插件已停止"

### FR-9: 前端 — 租户侧应用目录页

**描述：** 租户管理员可浏览平台可用的应用列表（运行中 + 对外可见），并申请订阅。

**验收标准：**
- WHEN 租户管理员访问应用目录 THEN 系统 SHALL 展示所有 status=运行中 的应用列表（含 BUILTIN + PLUGIN）
- WHEN 应用已被当前租户订阅 THEN 系统 SHALL 显示"已开通"标签 + 模块开关配置入口
- WHEN 应用未被订阅 THEN 系统 SHALL 显示"申请开通"按钮
- WHEN 点击"申请开通" THEN 系统 SHALL 创建 admin_tenant_app 订阅记录（self_service 模式直接生效）
- WHEN 退订应用 THEN 系统 SHALL 二次确认后删除 admin_tenant_app 记录 + 清理该租户下相关角色绑定

### FR-10: 后端 — 租户应用订阅接口

**描述：** 提供租户侧的应用订阅/退订/列表接口。

**验收标准：**
- WHEN GET /api/v1/admin/app-catalog THEN 系统 SHALL 返回当前租户可见的应用列表（含订阅状态）
- WHEN POST /api/v1/admin/app-subscriptions THEN 系统 SHALL 为当前租户订阅指定应用
- WHEN DELETE /api/v1/admin/app-subscriptions/:app_code THEN 系统 SHALL 退订并级联清理权限绑定
- WHEN 订阅成功 THEN 系统 SHALL 使该租户的角色配置页可见该应用的资源

## 非功能需求

- **性能**：PluginResourceSyncer.SyncOnStart 在单事务中完成，50 菜单 + 100 API 场景下 < 2s
- **幂等性**：SyncOnStart 先清后写，重复调用不产生脏数据
- **安全**：插件 API 权限默认 auth_required=1，必须通过角色分配才能访问
- **多租户**：插件资源为平台级（无 tenant_id），通过 admin_tenant_app 订阅关系控制租户可见性
- **可扩展**：plugin.json 作为元数据载体，后续扩展新字段只需修改 Host 解析逻辑
