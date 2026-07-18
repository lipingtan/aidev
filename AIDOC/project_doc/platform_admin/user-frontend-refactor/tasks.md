# 任务列表：插件化架构规范（Plugin Architecture）

## 执行顺序

```
Phase 1: 后端插件 SDK
  Task 1 → Task 2 → Task 3 → Task 4

Phase 2: 前端插件 SDK + 加载机制
  Task 5 → Task 6 → Task 7 → Task 8

Phase 3: 管理端插件管理页面
  Task 9 → Task 10

Phase 4: dev-web-user 框架瘦身
  Task 11 → Task 12

Phase 5: 现有模块拆为插件（示例）
  Task 13
```

依赖关系：
- Phase 2 依赖 Phase 1（前端需后端提供插件列表 API）
- Phase 3 依赖 Phase 1 + Phase 2（管理页面需调用后端 API + 使用前端框架）
- Phase 4 与 Phase 2 可并行（独立项目）
- Phase 5 依赖 Phase 1 + Phase 2（需 SDK 就绪）

---

## Phase 1: 后端插件 SDK

### Task 1: 定义 Plugin Interface 和 Manifest 结构

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `plugin-sdk/plugin.go`（新增）
  - `plugin-sdk/manifest.go`（新增）
  - `plugin-sdk/permission.go`（新增）
- 涉及模块: plugin-sdk
- 不触碰: backend 现有业务代码

**Acceptance（验证标准）:**
- AC: Plugin interface 定义完整（Name/Version/Manifest/RegisterRoutes/Permissions/OnEnable/OnDisable/OnInstall/OnUninstall/MigrateUp/MigrateDown）
- AC: Manifest struct 包含 name/version/displayName/description/dependencies/platforms/frontendEntry 字段
- AC: Permission struct 包含 code/displayName/group 字段
- AC: go build ./... 零错误

---

### Task 2: 实现 Plugin Manager + LocalAdapter

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `plugin-sdk/manager.go`（新增）
  - `plugin-sdk/adapter_local.go`（新增）
  - `plugin-sdk/adapter.go`（新增：PluginAdapter 接口定义）
  - `plugin-sdk/event.go`（新增：EventPublisher 接口）
- 涉及模块: plugin-sdk
- 不触碰: backend 现有业务代码、现有路由

**Constraints（约束）:**
- Manager 面向 PluginAdapter 接口编程，不直接依赖 LocalAdapter 实现
- EventPublisher 接口预留 Redis Pub/Sub，初期用 noop 实现
- Manager 方法需线程安全（sync.RWMutex 保护插件 map）

**Acceptance（验证标准）:**
- AC: PluginAdapter 接口定义（Load/Unload）
- AC: LocalAdapter 实现 Register/Load/Unload
- AC: Manager 实现 EnablePlugin/DisablePlugin/ListPlugins
- AC: EventPublisher 接口定义 + NoopPublisher 实现
- AC: go build ./... 零错误
- AC: go vet ./... 无警告

---

### Task 3: 创建插件相关数据库表 + Migration

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `backend/database/migrations/xxx_create_plugin_tables.sql`（新增）
  - `plugin-sdk/registry.go`（新增：PluginRegistry DB 操作）
- 涉及模块: plugin-sdk, database
- 不触碰: 现有业务表

**Acceptance（验证标准）:**
- AC: sys_plugin 表创建成功（含 name/version/status/config/frontend_entry/timestamps）
- AC: sys_plugin_version 表创建成功（含 plugin_name/version/snapshot_path/migration_ver）
- AC: sys_tenant_plugin 表创建成功（含 tenant_id/plugin_name/enabled）
- AC: sys_plugin_permission 表创建成功（含 plugin_name/code/display_name/group_name）
- AC: PluginRegistry CRUD 方法实现（GetByName/List/UpdateStatus/CreateVersion）
- AC: go build ./... 零错误

---

### Task 4: 插件生命周期 API

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `backend/app/plugin/apis/plugin.go`（新增）
  - `backend/app/plugin/router/router.go`（新增）
  - `backend/app/plugin/service/plugin.go`（新增）
- 涉及模块: app/plugin
- 不触碰: 现有 app/ 下其他模块

**Constraints（约束）:**
- 所有接口需 JWT 认证 + admin 角色校验
- Install/Upgrade 需执行 MigrateUp，失败自动回滚
- Rollback 需执行 MigrateDown
- 状态变更后调用 EventPublisher 广播

**Acceptance（验证标准）:**
- AC: POST /api/v1/plugin/install 安装插件
- AC: PUT /api/v1/plugin/{name}/enable 启用
- AC: PUT /api/v1/plugin/{name}/disable 禁用
- AC: DELETE /api/v1/plugin/{name} 卸载
- AC: PUT /api/v1/plugin/{name}/upgrade 升级
- AC: PUT /api/v1/plugin/{name}/rollback 回滚
- AC: GET /api/v1/plugins 获取插件列表（支持按 tenant_id 过滤）
- AC: go build ./... 零错误
- AC: 【回归】现有 API 不受影响（RG-1）

---

## Phase 2: 前端插件 SDK + 加载机制

### Task 5: 前端 Plugin SDK 模块

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/plugin-sdk/index.ts`（新增）
  - `dev-web-admin/src/plugin-sdk/types.ts`（新增）
  - `dev-web-admin/src/plugin-sdk/context.ts`（新增）
- 涉及模块: plugin-sdk（前端）
- 不触碰: 现有页面组件和路由

**Acceptance（验证标准）:**
- AC: 导出 definePlugin(config) 函数
- AC: 导出 usePluginContext() 组合式函数
- AC: 导出所有 TypeScript 类型（PluginManifest/PluginRoute/PluginMenu/PluginConfig/PluginContext/EventBus）
- AC: 导出 registerExtension(pointName, component, sort?) 函数
- AC: vue-tsc --noEmit 类型检查通过

---

### Task 6: EventBus 实现

**复杂度**: 低

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/core/event-bus.ts`（新增）
  - `dev-web-admin/package.json`（修改：添加 mitt 依赖）
- 不触碰: 其他模块

**Acceptance:**
- AC: 安装 mitt 依赖
- AC: 导出 pluginEventBus 对象实现 EventBus 接口（emit/on/off）
- AC: vue-tsc --noEmit 通过

---

### Task 7: Plugin Loader 实现

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/core/plugin-loader.ts`（新增）
  - `dev-web-admin/src/api/plugin.ts`（新增：获取已启用插件列表 API）
- 涉及模块: core, api
- 不触碰: 现有路由和页面

**Constraints（约束）:**
- 支持 source 和 runtime 两种加载模式
- 单个插件加载失败不影响其他插件
- 使用 import.meta.glob 实现源码模式的动态导入

**Acceptance（验证标准）:**
- AC: loadSource(name) 从 src/plugins/{name}/index.ts 加载
- AC: loadRuntime(name) 从 /static/plugins/{name}/index.js 动态 import
- AC: loadAll(enabledPlugins) 批量加载并容错
- AC: API 调用 GET /api/v1/plugins 获取已启用列表
- AC: vue-tsc --noEmit 通过

---

### Task 8: 动态路由与菜单合并

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/core/plugin-router.ts`（新增）
  - `dev-web-admin/src/core/plugin-menu.ts`（新增）
  - `dev-web-admin/src/router/index.ts`（修改：登录后调用插件加载和路由合并）
  - `dev-web-admin/src/layout/AppSidebar.vue`（修改：合并插件菜单）
- 涉及模块: core, router, layout
- 不触碰: 现有业务页面组件

**Constraints（约束）:**
- 插件路由添加到已有 AppLayout 下（router.addRoute）
- 菜单合并按 sort 排序
- 用户无权限时过滤菜单项
- 不破坏已有静态路由

**Acceptance（验证标准）:**
- AC: mergePluginRoutes 正确将插件路由注册到 Vue Router
- AC: mergePluginMenus 按 sort 排序 + 权限过滤
- AC: 登录后自动加载插件并注册路由/菜单
- AC: 禁用插件后菜单和路由不可见
- AC: vue-tsc --noEmit 通过
- AC: vite build 零错误
- AC: 【回归】已有页面正常访问（RG-3）

---

## Phase 3: 管理端插件管理页面

### Task 9: 插件管理列表页面

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/views/system/plugin/PluginList.vue`（新增）
  - `dev-web-admin/src/api/plugin.ts`（修改：增加 enable/disable/uninstall API）
- 涉及模块: system/plugin
- 不触碰: 其他 system 子模块

**Acceptance（验证标准）:**
- AC: 表格展示插件列表（名称、版本、状态、描述）
- AC: 支持启用/禁用切换操作
- AC: 支持卸载操作（确认弹窗）
- AC: 展示插件版本历史（展开行或弹窗）
- AC: 页面样式遵循 dev-web-admin 规范
- AC: vue-tsc --noEmit 通过

---

### Task 10: 租户插件分配页面

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/views/system/plugin/TenantPluginList.vue`（新增）
  - `dev-web-admin/src/api/plugin.ts`（修改：增加租户插件分配 API）
- 涉及模块: system/plugin
- 不触碰: 其他模块

**Acceptance（验证标准）:**
- AC: 展示所有租户及其已授权的插件列表
- AC: 支持为租户启用/禁用特定插件
- AC: 页面样式遵循 dev-web-admin 规范
- AC: vue-tsc --noEmit 通过

---

## Phase 4: dev-web-user 框架瘦身

### Task 11: 清理 dev-web-user 旧业务模块

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `dev-web-user/src/views/workorder/`（删除）
  - `dev-web-user/src/views/billing/`（删除）
  - `dev-web-user/src/views/notice/`（删除）
  - `dev-web-user/src/views/certify/`（删除）
  - `dev-web-user/src/views/repair/`（删除）
  - `dev-web-user/src/api/workorder.ts`（删除）
  - `dev-web-user/src/api/billing.ts`（删除）
  - `dev-web-user/src/api/notice.ts`（删除）
  - `dev-web-user/src/api/certification.ts`（删除）
  - `dev-web-user/src/api/repair.ts`（删除）
  - `dev-web-user/src/api/owner.ts`（删除）
  - `dev-web-user/src/api/base.ts`（删除）
  - `dev-web-user/src/api/base.spec.ts`（删除）
  - `dev-web-user/src/api/family.ts`（删除）
  - `dev-web-user/src/api/property.ts`（删除）
  - `dev-web-user/src/router/static-routes.ts`（修改）
- 涉及模块: views, api, router
- 不触碰: layout/, login/, register/, mine/(个人信息保留), home/, error/

**Constraints（约束）:**
- 保留 TabBarLayout 框架骨架
- 保留 login/register/个人中心（profile）/home/error 页面
- 移除所有物业业务相关路由

**Acceptance（验证标准）:**
- AC: workorder/billing/notice/certify/repair 的 views + API 全部移除
- AC: owner/base/family/property API 文件移除
- AC: static-routes.ts 中只保留 login/register/home/mine/error 路由
- AC: vite build 零错误
- AC: vue-tsc --noEmit 通过

---

### Task 12: dev-web-user 集成插件容器 + 响应式布局

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `dev-web-user/src/plugin-sdk/`（新增：复用 admin 的 SDK 类型）
  - `dev-web-user/src/core/plugin-loader.ts`（新增）
  - `dev-web-user/src/core/event-bus.ts`（新增）
  - `dev-web-user/src/views/plugin/PluginContainer.vue`（新增）
  - `dev-web-user/src/layout/TabBarLayout.vue`（修改：响应式适配）
  - `dev-web-user/src/router/static-routes.ts`（修改：添加插件容器路由）
- 涉及模块: plugin-sdk, core, views/plugin, layout, router
- 不触碰: login/register 页面逻辑

**Constraints（约束）:**
- 响应式布局基于 Vant + CSS media query（breakpoint: 768px）
- PC 端（≥768px）内容区域拉宽，可选侧边导航
- H5 端（<768px）保持 TabBar 底部导航

**Acceptance（验证标准）:**
- AC: PluginContainer 正确加载插件页面（source + runtime 模式）
- AC: PC 端内容区域自适应宽屏布局
- AC: H5 端保持 TabBar 底部导航
- AC: 插件菜单动态注入到导航中
- AC: vite build 零错误
- AC: vue-tsc --noEmit 通过

---

## Phase 5: 现有模块拆为插件（示例）

### Task 13: 将字典管理模块拆为插件（验证架构）

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/plugins/dict-manager/`（新增整个目录）
  - `dev-web-admin/src/plugins/dict-manager/manifest.json`（新增）
  - `dev-web-admin/src/plugins/dict-manager/index.ts`（新增）
  - `dev-web-admin/src/views/system/dict/`（移动到插件目录）
  - `dev-web-admin/src/api/dict.ts`（移动到插件目录）
  - `dev-web-admin/src/router/static-routes.ts`（修改：移除字典路由）
  - `backend/plugins/dict-manager/`（新增：后端插件实现）
- 涉及模块: plugins/dict-manager（前后端）
- 不触碰: 其他已有模块

**Constraints（约束）:**
- 后端插件实现 Plugin Interface 全部方法
- 前端插件使用 definePlugin() 标准入口
- manifest.json 包含完整的 permissions/menus/routes 声明
- 提供 up/down migration

**Acceptance（验证标准）:**
- AC: 字典管理功能通过插件方式加载，功能与重构前完全一致
- AC: 禁用插件后字典管理菜单和页面不可访问
- AC: 启用插件后功能恢复
- AC: 后端 go build ./... 零错误
- AC: 前端 vue-tsc --noEmit + vite build 通过
- AC: 【回归】其他模块不受影响（RG-3）
