# 设计计划：V2-CR6 多端前端架构 + Plugin SDK V2

## 设计方向

基于已确认的需求，拟采用以下技术方案：

1. **前端双入口构建**：使用 Vite 多入口配置，PC 和 H5 共享 `src/` 业务代码，仅布局层（`layouts/`）和入口文件（`main-pc.ts` / `main-h5.ts`）差异化
2. **Plugin SDK V2**：在现有 SDK 基础上扩展 `teardown` 机制和平台感知 API
3. **plugin-loader**：实现动态加载/卸载逻辑，按 `{platform}-{device}` 路径加载 bundle
4. **后端静态资源部署**：插件安装时解析 manifest.frontends，将 bundle 部署到对应目录

## 技术选型

### 前端构建方案

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| Vite 多入口 + 不同 main.ts | 配置简单、HMR 快、产物分离 | 需维护两套入口配置 | ✓ |
| 运行时判断 window.innerWidth | 零构建改动 | 无法 tree-shake、bundle 体积大、SEO 差 | ✗ |
| 两个独立 Vite 项目 | 完全隔离 | 代码重复、维护成本高 | ✗ |

### 布局组件策略

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| 抽象 Layout 组件 + 具体实现注入 | 业务代码零感知、可测试 | 需定义布局接口 | ✓ |
| 条件渲染 `v-if (isPC)` | 实现简单 | 布局代码耦合、bundle 冗余 | ✗ |

### Plugin SDK teardown 实现

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| `onTeardown(fn)` 注册模式 | 插件灵活控制、支持多个清理函数 | 插件需主动调用 | ✓ |
| 框架自动清理 | 插件零代码 | 无法清理插件自定义资源（定时器、监听器等） | ✗ |

## 澄清问题

- [Question-1] **H5 布局选型**：管理端 H5 采用哪种导航布局？
  - A. 底部 Tab 导航（类似 APP，一级菜单固定在底部）
  - B. 顶部汉堡菜单（点击展开侧滑菜单）
  - C. 顶部标题栏 + 面包屑导航（适合深层级页面）
  - **行业实践**：企业级管理端 H5 多采用「汉堡菜单 + 顶部标题栏」组合（如钉钉后台、飞书后台），底部 Tab 更适合 C 端高频功能入口
  - **推荐**：B（汉堡菜单）
  [Answer-1]
B
- [Question-2] **Element Plus vs Vant 组件混用策略**：dev-web-admin H5 模式是否允许混用 Vant 组件（H5 专用）？
  - A. 允许混用，H5 布局层使用 Vant（Popup/Tabbar），业务表单仍用 Element Plus
  - B. 仅使用 Element Plus，通过 CSS 媒体查询适配移动端
  - C. H5 模式完全使用 Vant，PC 模式完全使用 Element Plus
  - **行业实践**：混用会增加 bundle 体积（Vant + Element Plus），但能获得原生级移动体验。纯 Element Plus 移动端体验一般但包体积小
  - **推荐**：A（混用，布局层用 Vant，业务层用 Element Plus）
  [Answer-2]
方案A
- [Question-3] **插件 bundle 打包格式**：插件前端 bundle 采用什么模块格式？
  - A. ES Module（`import()`动态加载）
  - B. UMD（兼容性好，支持 `<script>` 标签加载）
  - C. SystemJS（适合复杂依赖场景）
  - **行业实践**：现代浏览器均支持 ES Module，配合 Vite 动态导入性能最优。UMD 适合需兼容旧浏览器场景
  - **推荐**：A（ES Module）
  [Answer-3]
B
- [Question-4] **插件依赖共享策略**：插件如何使用宿主的 Vue/Element Plus/Vant？
  - A. 宿主暴露全局变量（`window.Vue`、`window.ElementPlus`）
  - B. Vite externals + CDN（插件 bundle 不打包这些依赖）
  - C. 宿主通过 Plugin SDK 注入依赖（`ctx.libs.vue`）
  - **行业实践**：qiankun/micro-app 等微前端框架多采用「共享依赖」模式避免重复打包。宿主暴露全局变量最简单，SDK 注入最灵活
  - **推荐**：A（全局变量，简单且兼容性好）
  [Answer-4]
A 用命名空间 window.__PLATFORM_ADMIN_LIBS__ 而非裸 window.Vue，避免与其他可能存在的全局变量冲突。
- [Question-5] **H5 首屏性能目标**：NFR-1 要求 H5 首屏 < 3s（3G），如何验证？
  - A. 使用 Lighthouse throttling 模拟 3G 自动化测试
  - B. 真机测试（实际 3G 网络）
  - C. 开发阶段不验证，上线后通过 APM 监控
  - **行业实践**：Lighthouse 模拟足够精确且可自动化纳入 CI，真机测试成本高但更真实
  - **推荐**：A（Lighthouse CI）
  [Answer-5]
A
- [Question-6] **插件 teardown 超时处理**：如果插件 teardown 函数执行时间过长（如有异步清理），是否设置超时？
  - A. 设置超时（如 5s），超时后强制继续卸载
  - B. 不设超时，等待 teardown 完成
  - C. 异步 teardown 返回 Promise，宿主 await 但不阻塞 UI
  - **行业实践**：VSCode 扩展卸载有超时机制防止卡死。Chrome 扩展卸载是同步的，超时会被强制终止
  - **推荐**：A（5s 超时）
  [Answer-6]
A
## 风险点

- [Risk-1] **bundle 体积膨胀**：如果混用 Vant + Element Plus，需监控总 bundle 大小，可能需要按需引入
- [Risk-2] **插件内存泄漏**：即使有 teardown 机制，插件如果未正确清理（如忘记移除全局事件监听），仍会泄漏
- [Risk-3] **CSS 样式冲突**：插件样式可能与宿主冲突，需考虑 CSS 隔离方案（如 scoped CSS 或 shadow DOM）

## 影响范围预判

### 前端改动
- `dev-web-admin/`：新增 `main-pc.ts`、`main-h5.ts`、`layouts/PcLayout.vue`、`layouts/H5Layout.vue`、修改 Vite 配置
- `dev-web-user/`：同上结构
- `dev-web-admin/src/plugin-sdk/`：扩展 `PluginContext` 接口、实现 `onTeardown`
- `dev-web-admin/src/plugin-loader/`：实现 `loadPlugin`/`unloadPlugin`

### 后端改动
- `backend/common/plugin/installer.go`：解析 manifest.frontends，部署 bundle 到静态目录
- `backend/router/static.go`：注册 `/static/plugins/*` 静态文件路由
- 可能涉及 `sys_plugin` 表结构扩展（记录已部署的 frontend 组合）
