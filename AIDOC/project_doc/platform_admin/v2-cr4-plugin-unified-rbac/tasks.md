# 任务：V2-CR4 插件系统统一 RBAC + 通信契约

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 16 |
| 已完成 | 16 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 16/16 (100%) |
| 当前阶段 | 全部完成 |

---

## Phase 1: 基础设施层（Manifest + Syncer + ActionRegistry）

### Task 1: ManifestV2 结构体 + 解析函数 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: `backend/common/plugin/manifest.go`（新建）
- 不触碰: installer.go, manager.go

**Constraints（约束）:**
- ManifestV2 结构体严格对齐 design.md 中定义的 JSON 格式
- ParseManifest(pluginDir string) (*ManifestV2, error) 从 plugin.json 读取解析
- 缺少 displayName 时 fallback 到 name
- breakingUpgrade/minUpgradeFrom/migrationNotes 字段支持

**Acceptance（验证标准）:**
- AC: ManifestV2 含 Name/Version/DisplayName/Description/RoutePrefix/Platforms/Modules/Frontends/Menus/ApiPermissions/ExposedActions/SubscribedEvents/BreakingUpgrade/MinUpgradeFrom/MigrationNotes
- AC: ParseManifest 正确解析完整 plugin.json 和最小 plugin.json
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/common/plugin/manifest_test.go`
- ST: 完整 plugin.json 解析 → 所有字段正确填充
- ST: 最小 plugin.json（仅 name）→ displayName fallback 到 name
- ST: name 为空 → 返回错误
- ST: 文件不存在 → 返回错误

### Task 2: PluginResourceSyncer 实现 ✅

**复杂度**: 高

**依赖**: Task 1

**Scope（边界）:**
- 涉及文件: `backend/common/plugin/syncer.go`（新建）
- 涉及模块: plugin, auth/model
- 不触碰: manager.go, installer.go, auth/service

**Constraints（约束）:**
- SyncOnStart: 单事务内先清后写，硬删除（Unscoped）+ 批量 INSERT（非逐条 Create）
- 事务内硬删除保证：事务提交前外部查询看到旧数据，提交后看到新数据，无中间态
- ManifestMenu → model.Resource 转换：递归处理 children，正确设置 parent_id
- ManifestApiPermission → model.ApiPermission 转换：树形递归，GROUP 设 parent_id=nil
- 清理孤儿角色绑定：LEFT JOIN 查出引用已不存在的 resource_id/api_permission_id 的绑定
- SyncOnUninstall: 级联删除资源+绑定+tenant_app+软删除 application
- syncApplicationMeta: 更新 admin_application 的 name/description/route_prefix/platforms/modules
- 事务提交后同步调用 AutoDiscover.Refresh()
- 需要传入 *gorm.DB 实例

**Acceptance（验证标准）:**
- AC: SyncOnStart 在事务中执行，中途失败自动回滚
- AC: 重复调用 SyncOnStart 产出幂等（数据量不增）
- AC: SyncOnUninstall 级联清理所有租户的角色绑定
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/common/plugin/syncer_test.go`
- ST: SyncOnStart 写入菜单 → admin_resource 含正确记录（含 parent_id 递归）
- ST: SyncOnStart 写入 API → admin_api_permission 含正确记录（含 GROUP 树形）
- ST: SyncOnStart 重复调用 → 记录数不增（幂等）
- ST: SyncOnStart 中途 DB 错误 → 事务回滚，无残留数据
- ST: SyncOnUninstall → admin_resource/admin_api_permission/admin_role_resource/admin_role_api/admin_tenant_app 全部清理
- ST: SyncOnUninstall → admin_application 被软删除
- ST: syncApplicationMeta → admin_application 的 name/description/platforms/modules 被正确更新

### Task 3: ActionRegistry 实现 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: `backend/common/plugin/action_registry.go`（新建）
- 不触碰: event_bus.go（后续 Task 集成）

**Constraints（约束）:**
- 线程安全（sync.RWMutex）
- Register(pluginName, actions []ManifestAction)
- Unregister(pluginName)
- IsCompatible(pluginName, actionName, requestVersion string) bool — major 版本匹配
- FindAction(pluginName, actionName string) (*ActionDescriptor, bool)
- 版本比较逻辑：解析 "major.minor" 格式，仅比较 major 部分

**Acceptance（验证标准）:**
- AC: Register/Unregister 正确管理内存 map
- AC: IsCompatible("billing", "getInvoice", "1.5") 与注册的 "1.0" 兼容（major=1 相同）
- AC: IsCompatible("billing", "getInvoice", "2.0") 与注册的 "1.0" 不兼容（major 不同）
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/common/plugin/action_registry_test.go`
- ST: Register → FindAction 能找到已注册的 action
- ST: Unregister → FindAction 返回 false
- ST: IsCompatible major 相同 → true
- ST: IsCompatible major 不同 → false
- ST: IsCompatible action 不存在 → false
- ST: IsCompatible 空版本字符串 → false

---

## Phase 2: 核心改造（Manager + Installer + EventBus + Proxy）

### Task 4: PluginManager 集成 Syncer + ActionRegistry + 自动订阅 ✅

**复杂度**: 高

**依赖**: Task 1, Task 2, Task 3

**Scope（边界）:**
- 涉及文件: `backend/common/plugin/manager.go`
- 不触碰: installer.go, proxy.go

**Constraints（约束）:**
- NewPluginManager 增加参数：db *gorm.DB, pluginsDir string
- Start/StartProcess 成功后：
  1. ParseManifest(pluginsDir + "/" + name)
  2. syncer.SyncOnStart(name, manifest)
  3. actionRegistry.Register(name, manifest.ExposedActions)
  4. eventBus.Subscribe(name, manifest.SubscribedEvents)
- Stop 时：actionRegistry.Unregister(name)
- PluginManager 持有 syncer/actionRegistry 实例
- 启动失败时不调用 syncer（保持旧状态）

**Acceptance（验证标准）:**
- AC: 插件启动后 admin_resource 含该插件菜单
- AC: 插件启动后 admin_api_permission 含该插件 API
- AC: 插件停止后 ActionRegistry 清除该插件 actions
- AC: 插件停止后 EventBus 取消该插件订阅
- AC:【回归】现有 gRPC 通信不变（RG-2）
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/common/plugin/manager_test.go`
- ST: Start（进程内模式，有 manifest）→ syncer.SyncOnStart 被调用，admin_resource 有数据
- ST: Start（进程内模式，无 manifest）→ 插件正常启动，仅 log 警告
- ST: Start 后 → ActionRegistry 含该插件 actions
- ST: Start 后 → EventBus 含该插件订阅的事件
- ST: Start syncer 失败 → 状态回滚为 StatusError，registry 注销
- ST: Stop → ActionRegistry 不含该插件 actions
- ST: Stop → EventBus 不含该插件订阅
- ST: StopAll → 所有插件 ActionRegistry/EventBus 均清除

### Task 5: Installer 改造 — 安装创建 application + 升级接口 + 废弃 RegisterMenus ✅

**复杂度**: 高

**依赖**: Task 1, Task 2

**Scope（边界）:**
- 涉及文件: `backend/common/plugin/installer.go`
- 涉及模块: plugin, auth/model
- 不触碰: manager.go, proxy.go

**Constraints（约束）:**
- InstallFromFile 改造：
  - pluginJSON 结构体替换为 ManifestV2（解析完整字段）
  - 增加 admin_application 冲突检查（app_code 重名）
  - 创建 admin_application 记录（app_type=PLUGIN）
  - 不再调用 RegisterMenus
- 新增 Upgrade 方法：
  - 校验目标插件存在 + name 一致
  - breakingUpgrade 检测 → 返回 needConfirm
  - minUpgradeFrom 校验
  - 安全文件替换：新包部署到 .new → rename 旧目录为 .bak → rename .new 为正式
  - 更新 sys_plugin + admin_application
  - **不直接调用 manager.Start**（由 Handler 层编排：installer.Upgrade → manager.Start）
  - 提供 Rollback 方法：rename .bak 恢复
  - 启动时检测残留 .new/.bak 目录自动恢复
- Uninstall 改造：调用 syncer.SyncOnUninstall
- RegisterMenus/UnregisterMenus 标记 Deprecated（保留代码不删除，不再被调用）

**Acceptance（验证标准）:**
- AC: 安装后 admin_application 含 PLUGIN 记录
- AC: 同名安装被拒绝（含 admin_application 冲突检查）
- AC: 升级保留租户订阅和数据表
- AC: 破坏性升级返回 needConfirm + migrationNotes
- AC: 升级失败自动回滚到旧版本
- AC: 卸载后 admin_application 被软删除 + 资源清理
- AC:【回归】文件操作不变（RG-1）
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/common/plugin/installer_test.go`
- ST: InstallFromFile 成功 → admin_application 含 app_type=PLUGIN 记录
- ST: InstallFromFile 重名 → 返回冲突错误
- ST: InstallFromFile admin_application 同 app_code → 返回冲突错误
- ST: Upgrade breakingUpgrade=true → 返回 needConfirm + migrationNotes
- ST: Upgrade minUpgradeFrom 不满足 → 返回版本过低错误
- ST: Upgrade 正常 → sys_plugin/admin_application 版本更新，旧目录 .bak 存在
- ST: Rollback → .bak 恢复为正式目录
- ST: Uninstall → syncer.SyncOnUninstall 被调用，文件目录被删除
- ST: RegisterMenus 标记 Deprecated 但不报编译错误

### Task 6: EventBus.CallPlugin 增加版本校验 ✅

**复杂度**: 中

**依赖**: Task 3

**Scope（边界）:**
- 涉及文件: `backend/common/plugin/event_bus.go`
- 不触碰: manager.go, proxy.go

**Constraints（约束）:**
- CallPlugin 方法增加 ActionVersion 校验逻辑
- proto.CallPluginRequest 增加 action_version 字段（需修改 .proto + 重新生成）
- 需要 ActionRegistry 引用（通过构造参数或 PluginManager 传入）
- 不带 ActionVersion 的旧调用直接放行（向后兼容）
- 新增错误变量：ErrActionNotFound, ErrActionVersionIncompatible, ErrPluginNotRunning

**Acceptance（验证标准）:**
- AC: major 版本不匹配 → ErrActionVersionIncompatible
- AC: action 不存在 → ErrActionNotFound
- AC: 不带版本的调用正常放行
- AC:【回归】现有 CallPlugin 行为不变（RG-7）
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/common/plugin/event_bus_test.go`
- ST: CallPlugin 带 action_version 且 major 不匹配 → ErrActionVersionIncompatible
- ST: CallPlugin 带 action_version 且 action 不存在 → ErrActionNotFound
- ST: CallPlugin 不带 action_version → 正常转发（向后兼容）
- ST: CallPlugin 目标插件未运行 → ErrPluginNotRunning
- ST: 【回归】PublishEvent 行为不变（RG-7）

### Task 7: PluginProxy 503 增强 ✅

**复杂度**: 低

**Scope（边界）:**
- 涉及文件: `backend/common/plugin/proxy.go`
- 不触碰: manager.go, installer.go

**Acceptance（验证标准）:**
- AC: 插件未注册 → 404 + `{"code":40400,"msg":"插件不存在"}`
- AC: 插件已停止 → 503 + `{"code":50301,"msg":"插件维护中，请稍后再试","data":{"plugin":"xxx","status":"stopped"}}`
- AC: 插件运行中 → 正常代理
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/common/plugin/proxy_test.go`
- ST: 请求不存在的插件 → 404 + code=40400
- ST: 请求已停止的插件 → 503 + code=50301 + data 含 plugin/status
- ST: 请求运行中的插件 → 正常代理转发

---

## Phase 3: 后端路由 + Handler

### Task 8: 插件管理 Handler + 路由注册 ✅

**复杂度**: 高

**依赖**: Task 4, Task 5

**Scope（边界）:**
- 涉及文件: `backend/common/auth/handler/plugin_handler.go`（新建）, `backend/common/auth/router.go`
- 不触碰: 其他 handler

**Constraints（约束）:**
- 7 个接口：list/upload/start/stop/uninstall/upgrade/health
- 路由注册到 admin group（/api/v1/admin/plugins）
- permission_code 按 design.md 定义
- upgrade 接口支持 confirm 参数（二次确认）
- upload/upgrade 使用 multipart/form-data 文件上传
- upgrade Handler 编排：installer.Upgrade → manager.Start → 成功清理 .bak / 失败调 installer.Rollback
- AutoDiscover endpointPermissionCodes/endpointDisplayNames 需同步更新

**Acceptance（验证标准）:**
- AC: 7 个接口全部注册并可编译
- AC: upgrade 接口 breakingUpgrade 时返回 needConfirm
- AC:【回归】现有路由不受影响（RG-5）
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/common/auth/handler/plugin_handler_test.go`
- ST: GET /api/v1/admin/plugins → 200 + 返回插件列表
- ST: POST /api/v1/admin/plugins/:name/start → 启动成功返回 200
- ST: POST /api/v1/admin/plugins/:name/stop → 停止成功返回 200
- ST: DELETE /api/v1/admin/plugins/:name → 卸载成功返回 200
- ST: PUT /api/v1/admin/plugins/:name/upgrade breakingUpgrade → 返回 needConfirm
- ST: PUT /api/v1/admin/plugins/:name/upgrade confirm=true → 执行升级
- ST: GET /api/v1/admin/plugins/:name/health → 返回健康状态

### Task 9: 应用目录 Handler + 订阅接口 ✅

**复杂度**: 高

**依赖**: Task 4

**Scope（边界）:**
- 涉及文件: `backend/common/auth/handler/app_catalog_handler.go`（新建）, `backend/common/auth/router.go`, `backend/common/auth/service/tenant_service.go`
- 不触碰: 其他 service

**Constraints（约束）:**
- GET /app-catalog: auth_required=0（免检，仅需登录）
- POST /app-subscriptions: 检查配额 quota.max_apps
- DELETE /app-subscriptions/:app_code: 拒绝退订 BUILTIN 类型应用 + 级联清理该租户下该 app_code 的角色绑定
- GET /app-subscriptions: 返回当前租户已订阅列表
- 权限码 tenant:app:subscribe / tenant:app:unsubscribe
- 订阅时 enabled_modules 默认 NULL（全部启用）
- app-catalog 列表中 BUILTIN 应用不展示退订按钮（前端 + 后端双重校验）

**Acceptance（验证标准）:**
- AC: 4 个接口全部注册并可编译
- AC: app-catalog 含 BUILTIN + PLUGIN 应用，附带订阅状态和插件运行状态
- AC: 订阅超配额 → 拒绝
- AC: 退订级联清理角色绑定
- AC:【回归】现有 tenant_app 逻辑不变（RG-4）
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/common/auth/handler/app_catalog_handler_test.go`
- ST: GET /app-catalog → 200 + 含 BUILTIN 和 PLUGIN 应用 + 订阅状态字段
- ST: POST /app-subscriptions 合法 → 201 + admin_tenant_app 有记录
- ST: POST /app-subscriptions 超配额 → 403/400 + 配额超限信息
- ST: DELETE /app-subscriptions/:app_code PLUGIN 类型 → 200 + 角色绑定被清理
- ST: DELETE /app-subscriptions/:app_code BUILTIN 类型 → 400 + 禁止退订
- ST: GET /app-subscriptions → 200 + 当前租户已订阅列表

### Task 10: sys_menu 旧数据清理 + AutoDiscover 刷新触发 ✅

**复杂度**: 低

**依赖**: Task 4

**Scope（边界）:**
- 涉及文件: `backend/common/auth/auth.go`（启动时调用清理）
- 不触碰: syncer.go

**Acceptance（验证标准）:**
- AC: 启动后 sys_menu 中 plugin_% 前缀的记录被软删除
- AC: 幂等（重复启动不报错）
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/common/auth/auth_test.go`
- ST: 执行清理函数 → sys_menu 中 plugin_% 记录 deleted_at 非空
- ST: 重复执行 → 不报错且不产生副作用

---

## Phase 4: 集成验证

### Task 11: PluginManager 依赖注入链路更新 ✅

**复杂度**: 中

**依赖**: Task 4, Task 8, Task 9

**Scope（边界）:**
- 涉及文件: `backend/common/auth/auth.go`, `backend/common/auth/router.go`
- 不触碰: 其他模块初始化

**Constraints（约束）:**
- auth.go 中 PluginManager 创建时传入 db + pluginsDir
- PluginHandler / AppCatalogHandler 需要 PluginManager + Installer 实例
- Installer 需要 db + syncer 实例
- 确保依赖注入顺序正确（db → syncer → manager → installer → handler）

**Acceptance（验证标准）:**
- AC: 应用启动后 PluginManager 正确初始化
- AC: 所有 Handler 正确接收依赖
- AC: go build ./... 零错误
- AC: go vet ./... 零警告

**自测:**
- 测试文件: `backend/common/auth/auth_integration_test.go`
- ST: Init 调用后 PluginManager 非 nil 且持有 db/pluginsDir
- ST: PluginHandler/AppCatalogHandler 正确接收 Manager/Installer 依赖
- ST: 依赖注入顺序正确（db → syncer → manager → installer → handler）无 nil panic

---

## Phase 5: 前端

### Task 12: 前端 — 插件管理页改造 ✅

**复杂度**: 高

**依赖**: Task 8

**Scope（边界）:**
- 涉及文件: `dev-web-admin/src/views/plugin/`（改造）, `dev-web-admin/src/api/plugin.ts`（改造）
- 不触碰: 其他 views

**Constraints（约束）:**
- 数据源改为 /api/v1/admin/plugins（返回含 admin_application 信息）
- 状态感知 UI：运行中(绿)/已停止(橙)/已安装(灰)/异常(红)
- 操作按钮根据状态动态展示（启动/停止/重启/卸载/升级）
- 升级：文件上传 → 如返回 needConfirm → 弹窗展示 migrationNotes + "⚠️ 升级期间插件不可用" → 确认后再次请求（confirm=true）
- 升级执行后前端轮询 GET /plugins/:name/health 直到插件恢复 Running，期间展示 loading 状态
- 展示 platforms/modules 信息

**Acceptance（验证标准）:**
- AC: 插件列表正确展示各状态
- AC: 升级流程含二次确认弹窗
- AC: 操作按钮按状态正确显示/隐藏
- AC: API 路径带 /api/v1/admin/ 前缀

### Task 13: 前端 — 应用目录页 ✅

**复杂度**: 高

**依赖**: Task 9

**Scope（边界）:**
- 涉及文件: `dev-web-admin/src/views/app-catalog/`（新建）, `dev-web-admin/src/api/app-catalog.ts`（新建）
- 不触碰: 其他 views

**Constraints（约束）:**
- 卡片式展示应用列表（图标 + 名称 + 描述 + 状态标签）
- 已订阅显示"已开通"绿色标签 + "模块配置"按钮
- "模块配置"点击弹出 Drawer：展示应用 modules 列表（从 app-catalog 接口返回的 modules 字段），Switch 控制各模块启停，保存时 PUT /app-subscriptions/:app_code/modules
- 未订阅显示"申请开通"按钮；BUILTIN 类型已订阅时不展示"退订"按钮
- 退订需二次确认弹窗（提示将清理权限配置）
- PLUGIN 类型应用展示运行状态（停止时显示"维护中"标签）
- 配额超限时弹窗提示

**Acceptance（验证标准）:**
- AC: 展示 BUILTIN + PLUGIN 应用
- AC: 订阅/退订操作正常
- AC: 退订二次确认
- AC: 配额超限提示
- AC: API 路径带 /api/v1/admin/ 前缀

### Task 14: 前端 — 角色权限配置页插件状态提示 ✅

**复杂度**: 中

**依赖**: Task 4

**Scope（边界）:**
- 涉及文件: `dev-web-admin/src/views/permission/`（改造，权限配置组件）
- 不触碰: 后端

**Constraints（约束）:**
- 资源树中如果某 app_code 对应的插件状态为 stopped → 该应用节点名后追加红字"(插件已停止)"
- 数据源：前端在加载资源树时，同时请求 GET /api/v1/admin/plugins 获取插件状态列表，在前端按 app_code join 匹配
- 不影响权限分配操作（仍可分配/取消）

**Acceptance（验证标准）:**
- AC: 已停止插件的权限节点显示红字提示
- AC: 正常插件无额外标记
- AC: 分配操作不受影响

---

## Phase 6: 规范更新

### Task 15: go-plugin-development.md 规范更新 ✅

**复杂度**: 低

**依赖**: Task 5

**Scope（边界）:**
- 涉及文件: `AIDOC/global-info/knowledge/steering/software/go/go-plugin-development.md`
- 不触碰: 代码

**Acceptance（验证标准）:**
- AC: plugin.json 格式示例更新为 V2（含 menus/apiPermissions/exposedActions/subscribedEvents/breakingUpgrade）
- AC: "菜单注册机制"章节更新为 admin_resource 方式（废弃 sys_menu 描述）
- AC: 新增"插件升级"章节（breakingUpgrade/minUpgradeFrom/回滚机制）
- AC: 新增"Action 版本声明"章节

### Task 16: scaffold 模板更新 ✅

**复杂度**: 低

**依赖**: Task 1

**Scope（边界）:**
- 涉及文件: `projects/demo/platform_admin/scaffold/templates/plugin.json.tmpl`
- 不触碰: 其他模板

**Acceptance（验证标准）:**
- AC: 模板含 V2 完整字段结构（platforms/modules/menus/apiPermissions/subscribedEvents）
- AC: go build ./... 零错误（如模板涉及 Go 代码）
