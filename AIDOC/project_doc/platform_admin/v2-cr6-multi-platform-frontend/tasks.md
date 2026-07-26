# 任务列表：V2-CR6 多端前端架构 + Plugin SDK V2

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 12 |
| 已完成 | 12 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 12/12 (100%) |
| 当前阶段 | Phase 2: Plugin SDK V2 |

---

## 任务依赖关系

```
Task 1 (admin Vite 配置)  ─┐
Task 2 (user Vite 配置)   ─┤─→ Task 5 (admin Plugin SDK V2)
Task 3 (admin H5 布局)    ─┤─→ Task 6 (plugin-loader 辅助模块)
Task 4 (user H5 布局)     ─┘─→ Task 7 (plugin-loader 主模块)  ─→ Task 8 (前端联调)
                                                                      │
Task 9 (后端 manifest V2 扩展) ─→ Task 10 (后端 bundle 部署) ─→ Task 11 (路由验证)
                                                                      │
                                         Task 7 + Task 9 + Task 10 + Task 11 ─→ Task 12 (全链路联调)

可并行执行：Task 1 & 2、Task 3 & 4、Task 5 & 9、Task 6 & 10
```

---

## Phase 1：前端基础改造（双入口构建）

### Task 1: dev-web-admin 双入口 Vite 配置改造 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/vite.config.ts`（修改）
  - `dev-web-admin/package.json`（修改，新增构建命令）
  - `dev-web-admin/index-h5.html`（新增）
  - `dev-web-admin/src/main.ts`（修改，补充 `__PLATFORM_ADMIN_LIBS__` 暴露）
  - `dev-web-admin/src/main-h5.ts`（新增）
- 涉及模块: dev-web-admin
- 不触碰: `src/views/`、`src/api/`、`src/store/`、`src/router/`、`src/locales/`

**Acceptance（验证标准）:**
- AC: `npm run build:pc` 产物输出到 `dist-pc/` 目录
- AC: `npm run build:h5` 产物输出到 `dist-h5/` 目录
- AC: `dist-pc/` 和 `dist-h5/` 产物各自可独立访问（index.html 存在）
- AC: `main.ts`（PC）中 `window.__PLATFORM_ADMIN_LIBS__` 包含 `{ Vue, ElementPlus, VueRouter }`
- AC: `main-h5.ts` 中 `window.__PLATFORM_ADMIN_LIBS__` 额外包含 `Vant`
- AC: `index-h5.html` 的 viewport meta 包含 `user-scalable=no`
- AC: 【回归】`npm run dev` 仍正常启动（RG-2）

**自测:**
- 测试文件: 构建产物人工验证（前端构建无 Go 测试）
- ST: `npm run build:pc` → `dist-pc/index.html` 存在，内容引用 PC 入口 bundle
- ST: `npm run build:h5` → `dist-h5/index.html` 存在，内容引用 H5 入口 bundle
- ST: 浏览器打开 PC 产物 → Element Plus 侧边栏布局正常渲染
- ST: 【回归】dev server 启动后 `/api/v1/common/user-menu` 代理正常（RG-1）

---

### Task 2: dev-web-user 双入口 Vite 配置改造 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-user/vite.config.ts`（修改）
  - `dev-web-user/package.json`（修改，新增构建命令 + vant 依赖）
  - `dev-web-user/index-h5.html`（新增）
  - `dev-web-user/src/main.ts`（修改，补充 `__PLATFORM_USER_LIBS__` 暴露）
  - `dev-web-user/src/main-h5.ts`（新增）
- 涉及模块: dev-web-user
- 不触碰: `src/views/`、`src/api/`、`src/stores/`、`src/router/`

**Acceptance（验证标准）:**
- AC: `npm run build:pc` 产物输出到 `dist-pc/`
- AC: `npm run build:h5` 产物输出到 `dist-h5/`
- AC: `main.ts`（PC）中 `window.__PLATFORM_USER_LIBS__` 包含 `{ Vue, ElementPlus, VueRouter }`
- AC: `main-h5.ts` 中 `window.__PLATFORM_USER_LIBS__` 额外包含 `Vant`
- AC: Vant 依赖已在 `package.json` 中声明（vant@4.x）
- AC: 【回归】PC 入口 Element Plus 风格正常（RG-3）

**自测:**
- 测试文件: 构建产物人工验证
- ST: `npm run build:pc` → `dist-pc/index.html` 存在
- ST: `npm run build:h5` → `dist-h5/index.html` 存在
- ST: 【回归】PC 模式下 Element Plus 组件正常加载（RG-3）

---

### Task 3: dev-web-admin H5 布局组件实现 ✅

**依赖**: Task 1

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/layout/h5/H5Layout.vue`（新增）
  - `dev-web-admin/src/layout/h5/H5Header.vue`（新增）
  - `dev-web-admin/src/layout/h5/H5MenuDrawer.vue`（新增）
  - `dev-web-admin/src/layout/pc/PcLayout.vue`（新增，将现有 layout/ 下 PC 布局迁入）
  - `dev-web-admin/src/App.vue`（修改，根据构建模式注入对应 Layout）
- 涉及模块: dev-web-admin/layout
- 不触碰: `src/views/` 下所有业务组件

**Constraints（约束）:**
- `App.vue` 根据构建模式注入布局的机制：使用 `import.meta.env.MODE` 判断（`'h5'` 为 H5 模式），或通过 Vite `define` 注入编译期常量 `__IS_H5__`
- 推荐方案：在 `vite.config.ts` 中 `define: { __IS_H5__: isH5 }`，然后 App.vue 用 `if (__IS_H5__)` 条件导入对应 Layout，实现 tree-shaking

**Acceptance（验证标准）:**
- AC: H5 模式运行时，顶部显示 Vant NavBar（含标题和汉堡图标）
- AC: 点击汉堡图标，Vant Popup 从左侧滑出，显示菜单树
- AC: 点击菜单项后 Popup 关闭，路由正确跳转
- AC: PC 模式运行时，仍显示原有侧边栏 + 顶栏布局（RG-2）
- AC: H5 布局下 `src/views/` 业务页面正常渲染（代码共享验证）
- AC: `dist-h5/` 产物中不包含 PC 布局代码，`dist-pc/` 产物中不包含 H5/Vant 布局代码

**自测:**
- 测试文件: 浏览器 H5 DevTools 验证
- ST: H5 模式访问首页 → van-nav-bar 存在于 DOM
- ST: 点击 wap-nav 图标 → van-popup 出现，class 包含 left 方向
- ST: 【回归】PC 模式访问首页 → 侧边栏布局存在（RG-2）

---

### Task 4: dev-web-user H5 布局组件实现 ✅

**依赖**: Task 2

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-user/src/layout/h5/UserH5Layout.vue`（新增）
  - `dev-web-user/src/layout/pc/UserPcLayout.vue`（新增，将现有 PC 布局迁入）
  - `dev-web-user/src/App.vue`（修改，根据构建模式注入对应 Layout）
- 涉及模块: dev-web-user/layout
- 不触碰: `src/views/` 下所有业务组件

**Acceptance（验证标准）:**
- AC: H5 模式运行时，底部显示 Vant Tabbar（route 模式）
- AC: Tabbar 菜单项来源于后端 `/api/v1/common/user-menu?platform=user` 返回的顶级菜单
- AC: 点击 Tabbar 项，路由正确跳转，选中状态正确高亮
- AC: PC 模式运行时，仍显示 Element Plus 风格布局（RG-3）

**自测:**
- 测试文件: 浏览器 H5 DevTools 验证
- ST: H5 模式访问 → van-tabbar 存在于底部
- ST: 点击 Tabbar 项 → router-view 内容更新，active tab 正确
- ST: 【回归】PC 模式访问 → Element Plus 风格布局正常（RG-3）

---

## Phase 2：Plugin SDK V2

### Task 5: Plugin SDK V2 类型扩展 ✅

**依赖**: Task 1（需要确认 Vite 构建后 plugin-sdk 正常导出）

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/plugin-sdk/types.ts`（修改）
  - `dev-web-admin/src/plugin-sdk/index.ts`（修改，补充新类型导出）
- 涉及模块: dev-web-admin/plugin-sdk
- 不触碰: 其他 src/ 下文件

**Acceptance（验证标准）:**
- AC: `PluginContext` 接口新增 `platform`、`device`、`getPlatform()`、`getDevice()`、`onTeardown()`、`i18n.mergeLocale()`
- AC: `registerExtension` 返回值类型从 `void` 变更为 `UnregisterFn`
- AC: `PluginManifest` 新增 `frontends` 可选字段（数组形式）
- AC: `PluginConfig` 新增可选 `teardown` 字段（V1 兼容）
- AC: TypeScript 编译零错误（`tsc --noEmit`）
- AC: `UnregisterFn` 和 `TeardownFn` 类型已导出

**自测:**
- 测试文件: TypeScript 编译验证（`tsc --noEmit`）
- ST: 导入 `PluginContext` → IDE 类型提示包含所有 V2 新字段
- ST: `registerExtension` 赋值到 `UnregisterFn` 变量 → 编译通过
- ST: `PluginManifest.frontends` 赋值数组 → 编译通过，结构匹配

---

### Task 6: plugin-loader 辅助模块实现 ✅

**依赖**: Task 5

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/plugin-loader/extension-registry.ts`（新增）
  - `dev-web-admin/src/plugin-loader/route-manager.ts`（新增）
  - `dev-web-admin/src/plugin-loader/event-bus.ts`（新增）
- 涉及模块: dev-web-admin/plugin-loader
- 不触碰: 现有 router/ 下文件（route-manager 通过参数接收 router 实例，不直接 import）

**Acceptance（验证标准）:**
- AC: `doRegisterExtension(point, component, sort)` 将组件加入扩展点 Map，返回 `UnregisterFn`
- AC: 调用返回的 `UnregisterFn` 后，该组件从 Map 中移除
- AC: `registerPluginRoutes(pluginName, routes)` 将路由动态添加到 Vue Router
- AC: `unregisterPluginRoutes(pluginName)` 移除该插件注册的所有路由
- AC: `globalEventBus` 使用 `mitt` 实现，支持 `emit`/`on`/`off`
- AC: TypeScript 编译零错误

**自测:**
- 测试文件: `dev-web-admin/src/plugin-loader/__tests__/extension-registry.test.ts`
- ST: `doRegisterExtension('menu.extra', MockComp)` → 扩展点 Map 有该组件
- ST: 调用返回的 `UnregisterFn` → 扩展点 Map 中该组件已移除
- ST: `globalEventBus.on('test', fn)` → `globalEventBus.emit('test', data)` 触发 fn

---

### Task 7: plugin-loader 主模块实现 ✅

**依赖**: Task 5, Task 6

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/plugin-loader/index.ts`（新增）
  - `dev-web-admin/src/main.ts`（修改，初始化 globalEventBus）
  - `dev-web-admin/src/main-h5.ts`（修改，初始化 globalEventBus）
- 涉及模块: dev-web-admin/plugin-loader
- 不触碰: `src/store/`、`src/router/` 内部实现（只通过外部 API 调用）

**Constraints（约束）:**
- `loadPlugin` 必须防重复加载（`data-plugin` 属性检查）
- `unloadPlugin` 必须处理 V1 兼容（`PluginConfig.teardown` 字段）
- teardown 超时时间固定为 5000ms，超时后强制继续卸载
- 插件 bundle 加载失败（onerror）必须 reject 并携带明确错误信息
- 约定 UMD 全局变量名格式：`__PLUGIN_{NAME_UPPER}__`（短横线转下划线）
- `route-manager.ts` 在 `main.ts` 和 `main-h5.ts` 初始化阶段通过 `initRouteManager(router)` 注入 Vue Router 实例，`plugin-loader/index.ts` 直接调用已初始化的 route-manager，不在模块顶层 import router

**Acceptance（验证标准）:**
- AC: `loadPlugin('game', 'admin', 'pc')` 请求路径为 `/static/plugins/game/admin-pc/bundle.js`
- AC: bundle 加载失败时 `loadPlugin` reject，主应用继续正常运行（FR-5 错误隔离）
- AC: `unloadPlugin` 执行后 `extensionRegistry` 和 `teardownQueues` 中无该插件记录
- AC: teardown 耗时超过 5s 时，`unloadPlugin` 在 5s 后强制继续（不永久挂起）
- AC: V1 插件（有 `PluginConfig.teardown` 字段）卸载时该函数被调用（RG-4）
- AC: 【回归】现有 game 插件 setup 函数调用不报错（RG-4）

**自测:**
- 测试文件: `dev-web-admin/src/plugin-loader/__tests__/plugin-loader.test.ts`（vitest）
- ST: `loadPlugin` 成功路径 → `window.__PLUGIN_GAME__` 中 setup 被调用
- ST: `loadPlugin` 失败路径（bundle 404）→ reject 含路径信息，不 throw 到宿主
- ST: `unloadPlugin` 后 → `extensionRegistry.has('game')` === false
- ST: teardown 函数耗时 > 5s → `unloadPlugin` 在 5s 后正常 resolve（超时警告日志出现）
- ST: V1 插件（有 teardown 字段）`unloadPlugin` → teardown 字段被调用

---

### Task 8: 前端联调验证（H5 + 插件加载） ✅

**依赖**: Task 3, Task 4, Task 7

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: 无代码修改，仅验证
- 涉及模块: dev-web-admin, dev-web-user

**Acceptance（验证标准）:**
- AC: dev-web-admin H5 模式访问后台，汉堡菜单正常显示系统菜单
- AC: dev-web-user H5 模式，底部 Tabbar 显示用户端菜单；菜单为空时显示空态提示，不报错
- AC: PC 产物部署后各功能正常（RG-2, RG-3）
- AC: NFR-3：`src/views/` 修改一处，PC 和 H5 构建产物均体现该修改
- AC: Lighthouse CI 对 H5 产物首屏 LCP ≤ 3000ms（模拟 3G 节流，满足 NFR-1）

---

## Phase 3：后端插件 bundle 部署

### Task 9: 后端 ManifestV2 扩展 frontends 字段 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `backend/common/plugin/manifest.go`（修改，调整 `Frontends` 字段结构）
- 涉及模块: backend/common/plugin
- 不触碰: `installer.go`、`manager.go`、其他 plugin 文件

**Constraints（约束）:**
- 当前 `ManifestV2.Frontends` 类型为 `map[string]map[string]string`，需改为结构化数组以匹配新设计
- 必须保持向后兼容（V1 插件无 `frontends` 字段，解析不报错）
- 新类型：`[]FrontendConfig`，`FrontendConfig{ Platform, Device, Entry string }`

**Acceptance（验证标准）:**
- AC: `ManifestV2.Frontends` 类型改为 `[]FrontendConfig`
- AC: `FrontendConfig` struct 包含 `Platform`、`Device`、`Entry` 字段，JSON tag 为小写
- AC: 解析不含 `frontends` 字段的旧 plugin.json → `Frontends` 为空切片，不报错
- AC: `go build ./...` 零错误
- AC: `go vet ./...` 无警告

**自测:**
- 测试文件: `backend/common/plugin/manifest_test.go`（新增测试用例）
- ST: 解析含 `frontends` 数组的 plugin.json → `Frontends` 切片长度正确
- ST: 解析不含 `frontends` 字段的 plugin.json → `Frontends` 为 nil/空切片，无 error
- ST: `FrontendConfig.Platform` 值为 "admin"，`Device` 为 "pc" → `go build` 通过

---

### Task 10: 后端插件 bundle 部署逻辑实现 ✅

**依赖**: Task 9

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `backend/common/plugin/installer.go`（修改，新增 `deployFrontendBundles` 和 `removeFrontendBundles`，修改 `InstallFromFile` 和 `Uninstall`）
- 涉及模块: backend/common/plugin
- 不触碰: `manager.go`、`syncer.go`、`proxy.go`，不修改其他函数签名

**Constraints（约束）:**
- `deployFrontendBundles` 目标路径格式严格为 `{staticDir}/plugins/{name}/{platform}-{device}/bundle.js`，platform 和 device 均小写
- 源文件路径从插件目录 `frontend/{entry}` 解析（`entry` 来自 `FrontendConfig.Entry`）
- `removeFrontendBundles` 使用 `os.RemoveAll` 删除整个插件静态目录
- `InstallFromFile` 调用 `deployFrontendBundles` 替代原有的单路径前端 bundle 处理逻辑
- `Uninstall` 已有 `os.RemoveAll(staticPluginDir)` 逻辑保留，不重复删除

**Acceptance（验证标准）:**
- AC: 安装含 `frontends: [{platform:"admin",device:"pc",entry:"admin-pc/bundle.js"}]` 的插件 → `{staticDir}/plugins/{name}/admin-pc/bundle.js` 文件存在
- AC: 安装含多个 frontends 条目的插件 → 每个对应目录均创建正确
- AC: `frontends` 为空的插件安装 → 不报错，不创建任何 bundle 目录
- AC: 卸载插件 → `{staticDir}/plugins/{name}/` 目录被删除
- AC: bundle 源文件不存在时（entry 路径错误）→ 安装返回明确错误，不部分部署
- AC: `go build ./...` 零错误
- AC: 【回归】现有 `InstallFromFile` 的 zip/tar.gz 解压逻辑不受影响（RG-6）

**自测:**
- 测试文件: `backend/common/plugin/installer_test.go`（新增测试用例）
- ST: `deployFrontendBundles` 单 frontend → 目标文件存在，内容与源一致
- ST: `deployFrontendBundles` 多 frontend → 所有目标文件存在
- ST: `removeFrontendBundles` → 整个插件静态目录不存在
- ST: bundle 源文件不存在 → 返回包含路径信息的 error
- ST: 【回归】安装不含 frontends 的旧插件 → 安装成功，无报错

---

### Task 11: 后端静态文件路由验证 ✅

**依赖**: Task 10

**复杂度**: 低

**说明：** 经代码确认，`backend/cmd/api/server.go` 中 `registerStaticFiles` 函数已包含 `/static/plugins/*` 路由处理逻辑（匹配 `strings.HasPrefix(fp, "/plugins/")` 时从磁盘读取 `./static` + fp）。本 Task 为验证任务，确认该逻辑与新的多端 bundle 路径格式兼容，无需修改代码。

**Scope（边界）:**
- 涉及文件: `backend/cmd/api/server.go`（只读确认，无修改）
- 实际路由处理位于 `registerStaticFiles` 函数的 `/static/*filepath` 路由
- 不触碰: 其他路由、中间件配置

**Acceptance:**
- AC: GET `/static/plugins/game/admin-pc/bundle.js` → 返回对应文件内容（200），Content-Type 为 `application/javascript`
- AC: GET `/static/plugins/game/user-h5/bundle.js` → 同上（验证多端路径格式）
- AC: GET `/api/v1/admin/plugins` → 正常响应，`/api/` 路由不受影响（RG-6）
- AC: `go build ./...` 零错误

---

### Task 12: 前后端联调验证（插件 bundle 部署） ✅

**依赖**: Task 7, Task 9, Task 10, Task 11

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: 无代码修改，验证 + 修复 game 插件 plugin.json（如需）
- 涉及模块: dev-web-admin, backend/common/plugin

**Acceptance（验证标准）:**
- AC: 将 game 插件 plugin.json 添加 `frontends` 配置，重新安装后 bundle 正确部署到静态目录
- AC: `loadPlugin('game', 'admin', 'pc')` 成功加载 bundle，game 插件菜单出现在管理端
- AC: 插件卸载后 `unloadPlugin` 触发 teardown，菜单和路由消失（RG-5）
- AC: 【回归】现有 `/api/v1/admin/plugins` 列表接口正常返回（RG-6）
- AC: 【备注-低优】V1 兼容 teardown 中 ctx 的 platform/device 为硬编码，不影响现有插件但需人工确认 teardown 行为符合预期

---

## 一致性自检

**执行结果（tasks.md 生成后自动检查）：**

| # | 检查项 | 结果 |
|---|--------|------|
| 1 | 需求覆盖度 | FR-1→Task1/3, FR-2→Task2/4, FR-3→Task5, FR-4→Task5/7, FR-5→Task6/7, FR-6→Task9/10/11 ✅ |
| 2 | 设计对齐度 | vite.config→T1/2, H5布局→T3/4, SDK types→T5, loader辅助→T6, loader主→T7, manifest→T9, installer→T10, 路由→T11 ✅ |
| 3 | 文件路径真实性 | `installer.go`、`manifest.go`、`dev-web-admin/src/plugin-sdk/` 均已确认存在 ✅ |
| 4 | 回归覆盖 | RG-1→T1, RG-2→T1/3, RG-3→T2/4, RG-4→T7, RG-5→T12, RG-6→T11/12 ✅ |
| 5 | 内部一致性 | 无循环依赖，并行 Task 无文件交叉 ✅ |
| 6 | 前后端对称性 | 后端 bundle 部署(T9/10/11) 对应前端 loadPlugin(T7)，bundle 路径格式一致 ✅ |
