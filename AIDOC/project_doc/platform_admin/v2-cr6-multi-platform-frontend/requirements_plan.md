# 需求计划：V2-CR6 多端前端架构 + Plugin SDK V2

## 需求理解

根据 V2 架构 implementation_roadmap.md 中 CR-6 的定义，本 CR 的目标是：

- **目标**：实现双前端（dev-web-admin / dev-web-user）支持 PC + H5 双入口，以及 Plugin SDK V2 支持插件卸载清理和多端感知
- **范围**：
  - dev-web-admin 工程改造支持 PC/H5 双入口（共享 src，不同布局层）
  - dev-web-user 工程改造支持 PC/H5 双入口
  - Plugin SDK V2（teardown / platform / device / UnregisterFn / i18n.mergeLocale）
  - plugin-loader 实现（loadPlugin / unloadPlugin）
  - 插件 bundle 按 platform-device 路径加载
  - 后端：插件安装时按 manifest.frontends 部署 bundle 到对应静态目录
- **预期效果**：
  - 管理端和用户端都能以 PC 模式和 H5 模式运行
  - 插件前端能正确按端加载，禁用/卸载时能正确清理资源
  - 开发者使用 Plugin SDK V2 可以注册扩展点并获得取消注册函数

## 假设列表

- [假设-1] CR-4（插件 Manifest V2）已完成，manifest 中已有 `frontends` 字段定义插件前端 bundle
- [假设-2] CR-5（dev-web-user 基础框架）已完成，dev-web-user 工程已存在基础骨架
- [假设-3] PC 和 H5 共享大部分业务组件和逻辑，仅布局层（Layout）和响应式样式不同
- [假设-4] 插件前端 bundle 以独立 JS 文件形式存在，通过动态 import 加载
- [假设-5] 后端静态目录结构为 `/static/plugins/{plugin_name}/{platform}/{device}/`

## 澄清问题

- [Question-1] PC 和 H5 的构建策略是什么？是否使用单入口多配置（一次构建生成 PC+H5）还是双入口独立构建？
  **行业实践**：
  - 方案 A：单入口 + CSS 媒体查询 — 一套代码适配所有设备，构建简单但 H5 包体较大
  - 方案 B：双入口独立构建 — `vite build --mode pc` / `vite build --mode h5`，包体更小但构建复杂
  - 方案 C：共享 src + 不同 main.ts 入口 — 折中方案，推荐
  **推荐**：方案 C，平衡开发效率和产物体积
  [Answer-1] 
方案C
- [Question-2] H5 端是否需要支持独立部署域名（如 m.example.com），还是与 PC 同域通过 UA/路径区分？
  **行业实践**：
  - 独立域名（m.xxx.com）：利于 CDN 缓存策略独立、SEO 区分，但需额外部署配置
  - 同域路径区分（/h5/xxx）：运维简单，共享登录态，但 Vite 需配置 base
  [Answer-2] 
独立域名
- [Question-3] Plugin SDK V2 的 teardown 机制期望在哪些场景触发？
  **选项**：
  - A. 仅在插件卸载时触发
  - B. 在插件禁用（stop）和卸载（uninstall）时都触发
  - C. 在路由离开插件页面时也触发（SPA 内部清理）
  **推荐**：B（stop/uninstall 时都触发），C 可由插件自行实现 onUnmounted
  [Answer-3] 
B
- [Question-4] 插件前端 bundle 的多端 variant 文件结构如何设计？
  **候选方案**：
  - A. `/plugins/{name}/admin-pc/bundle.js` + `/plugins/{name}/admin-h5/bundle.js`
  - B. `/plugins/{name}/bundle.{platform}.{device}.js`（单目录多文件）
  **推荐**：方案 A，目录隔离更清晰，利于 CDN 缓存策略
  [Answer-4] 
方案A
- [Question-5] i18n.mergeLocale 是否需要支持热更新（插件启用后动态合入语言包），还是仅支持启动时合入？
  **推荐**：启动时合入即可，热更新增加复杂度且实际需求不高
  [Answer-5] 
启动时合入
## 非功能需求建议

- **性能**：
  - H5 首屏加载时间 < 3s（3G 网络）
  - 插件 bundle 按需加载，不阻塞主应用
  - PC/H5 共享组件应 tree-shaking 友好
- **兼容性**：
  - H5 支持 iOS Safari 14+、Android Chrome 80+
  - PC 支持 Chrome 88+、Edge 88+、Firefox 78+
- **可维护性**：
  - PC/H5 差异应控制在布局层，业务逻辑 100% 共享
  - Plugin SDK 需有 TypeScript 类型定义

## 影响范围预判

- 涉及模块：
  - 前端 dev-web-admin（工程结构改造）
  - 前端 dev-web-user（工程结构改造）
  - 前端 Plugin SDK 包（新建或重构）
  - 后端 common/plugin/installer.go（bundle 部署逻辑）
  - 后端 common/plugin/proxy.go（静态资源路由）
- 涉及文件（预估）：
  - `dev-web-admin/vite.config.ts`（双入口配置）
  - `dev-web-admin/src/main-pc.ts` / `main-h5.ts`（入口拆分）
  - `dev-web-admin/src/layouts/pc/` 和 `layouts/h5/`（布局组件）
  - `dev-web-user/` 对称改造
  - `dev-web-shared/plugin-sdk/`（SDK 包）
  - `backend/common/plugin/installer.go`
- 可能的副作用：
  - 现有 dev-web-admin 的单入口构建脚本需调整
  - 插件前端开发者需适配新 SDK API
