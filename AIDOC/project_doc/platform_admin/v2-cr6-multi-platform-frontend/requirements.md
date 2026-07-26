# 需求：V2-CR6 多端前端架构 + Plugin SDK V2

## 背景

Platform Admin V2 架构需要支持管理端（dev-web-admin）和用户端（dev-web-user）同时在 PC 和 H5 设备上运行。同时，插件系统需要升级到 SDK V2，支持多端感知和资源清理机制。本 CR 是 V2 架构路线图中第 6 段，依赖 CR-4（插件 Manifest V2）和 CR-5（dev-web-user 基础框架）。

## 用户故事

- 作为**管理端用户**，我希望能在手机浏览器上访问管理后台，以便在移动场景下处理紧急事务
- 作为**C端用户**，我希望能在 PC 和手机上都能使用用户端功能，以便获得一致的使用体验
- 作为**插件开发者**，我希望能为不同设备提供定制化的前端界面，以便优化各端的用户体验
- 作为**插件开发者**，我希望在插件停止或卸载时能清理已注册的资源，以便避免内存泄漏和副作用残留
- 作为**平台运维人员**，我希望 PC 和 H5 能独立部署到不同域名，以便灵活配置 CDN 和缓存策略

## 功能需求

### FR-1: dev-web-admin PC/H5 双入口改造

**描述：** 管理端前端工程支持 PC 和 H5 两种构建模式，共享业务代码，仅布局层差异化。

**验收标准：**
- WHEN 执行 `npm run build:pc` THEN 系统 SHALL 生成 PC 版本产物到 `dist-pc/` 目录
- WHEN 执行 `npm run build:h5` THEN 系统 SHALL 生成 H5 版本产物到 `dist-h5/` 目录
- WHEN 以 PC 模式运行 THEN 系统 SHALL 使用侧边栏 + 顶栏布局
- WHEN 以 H5 模式运行 THEN 系统 SHALL 使用移动端适配布局（底部导航或汉堡菜单）
- WHEN 修改 `src/views/` 下的业务组件 THEN 系统 SHALL 同时影响 PC 和 H5 版本（代码共享）

### FR-2: dev-web-user PC/H5 双入口改造

**描述：** 用户端前端工程支持 PC 和 H5 两种构建模式，结构与管理端对称。

**验收标准：**
- WHEN 执行 `npm run build:pc` THEN 系统 SHALL 生成 PC 版本产物到 `dist-pc/` 目录
- WHEN 执行 `npm run build:h5` THEN 系统 SHALL 生成 H5 版本产物到 `dist-h5/` 目录
- WHEN 以 H5 模式运行 THEN 系统 SHALL 使用 Vant 风格的移动端组件
- WHEN 以 PC 模式运行 THEN 系统 SHALL 使用 Element Plus 风格的桌面端组件

### FR-3: Plugin SDK V2 — 基础 API

**描述：** 提供插件前端开发所需的 SDK，包含平台感知、扩展点注册等核心能力。

**验收标准：**
- WHEN 插件调用 `sdk.getPlatform()` THEN 系统 SHALL 返回当前平台标识（`"admin"` 或 `"user"`）
- WHEN 插件调用 `sdk.getDevice()` THEN 系统 SHALL 返回当前设备标识（`"pc"` 或 `"h5"`）
- WHEN 插件调用 `sdk.registerExtension(name, component)` THEN 系统 SHALL 返回 `UnregisterFn` 函数
- WHEN 调用返回的 `UnregisterFn` THEN 系统 SHALL 从扩展点注册表中移除该组件
- WHEN 插件调用 `sdk.i18n.mergeLocale(locale, messages)` THEN 系统 SHALL 将语言包合入主应用 i18n 实例

### FR-4: Plugin SDK V2 — teardown 机制

**描述：** 插件停止或卸载时，自动调用已注册的清理函数，释放资源。

**验收标准：**
- WHEN 插件调用 `sdk.onTeardown(cleanupFn)` THEN 系统 SHALL 将 cleanupFn 注册到插件的清理队列
- WHEN 插件被 stop（禁用）THEN 系统 SHALL 依次调用该插件所有已注册的 teardown 函数
- WHEN 插件被 uninstall（卸载）THEN 系统 SHALL 依次调用该插件所有已注册的 teardown 函数
- WHEN teardown 函数执行完成 THEN 系统 SHALL 清空该插件的扩展点注册（自动 unregister）

### FR-5: plugin-loader 实现

**描述：** 实现插件前端 bundle 的动态加载和卸载机制。

**验收标准：**
- WHEN 调用 `loadPlugin(pluginName)` THEN 系统 SHALL 根据当前 platform-device 加载对应 bundle
- WHEN bundle 路径为 `/plugins/{name}/admin-pc/bundle.js` THEN 系统 SHALL 正确解析并加载
- WHEN 调用 `unloadPlugin(pluginName)` THEN 系统 SHALL 触发该插件的 teardown 流程
- WHEN unloadPlugin 完成 THEN 系统 SHALL 从 DOM 和内存中清理插件相关资源
- WHEN 插件 bundle 加载失败 THEN 系统 SHALL 返回明确错误信息，不影响主应用运行

### FR-6: 后端插件 bundle 部署

**描述：** 插件安装时，后端根据 manifest.frontends 配置将前端 bundle 部署到对应静态目录。

**验收标准：**
- WHEN 安装含 frontends 配置的插件 THEN 系统 SHALL 解析 frontends 数组中的 platform-device 组合
- WHEN frontends 包含 `{"platform":"admin","device":"pc"}` THEN 系统 SHALL 部署 bundle 到 `/static/plugins/{name}/admin-pc/`
- WHEN frontends 包含 `{"platform":"user","device":"h5"}` THEN 系统 SHALL 部署 bundle 到 `/static/plugins/{name}/user-h5/`
- WHEN 插件卸载 THEN 系统 SHALL 删除该插件的所有静态资源目录
- WHEN 请求 `/static/plugins/{name}/{platform}-{device}/bundle.js` THEN 系统 SHALL 返回对应文件

## 非功能需求

- **NFR-1 性能**：H5 首屏加载时间 < 3s（3G 网络）；插件 bundle 按需加载，不阻塞主应用首屏
- **NFR-2 兼容性**：H5 支持 iOS Safari 14+、Android Chrome 80+；PC 支持 Chrome 88+、Edge 88+、Firefox 78+
- **NFR-3 可维护性**：PC/H5 业务代码 100% 共享，差异仅在布局层；Plugin SDK 提供完整 TypeScript 类型定义
- **NFR-4 部署灵活性**：PC 和 H5 产物支持独立域名部署（如 admin.example.com / m-admin.example.com）

## 术语表

| 术语 | 定义 |
|------|------|
| platform | 前端平台，取值 `admin`（管理端）或 `user`（用户端） |
| device | 设备类型，取值 `pc` 或 `h5` |
| teardown | 插件清理函数，在插件停止/卸载时执行 |
| UnregisterFn | 取消注册函数，调用后移除已注册的扩展点 |
| bundle | 插件前端构建产物，通常为单个 JS 文件 |
