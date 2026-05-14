# CR01 - 前端框架搭建 设计计划

| 项目 | 内容 |
|------|------|
| 需求编号 | FE-CR01 |
| 需求名称 | 前端主体页面框架搭建 |
| 版本 | 1.0 |
| 创建日期 | 2026-04-17 |
| 状态 | 待审批 |

---

## 一、设计目标

基于 `requirements.md` 中定义的三个用户故事（US-01 PC 管理端框架、US-02 H5 业主端框架、US-03 公共基础设施），制定完整的前端框架技术设计方案，输出 `design.md`。

---

## 二、设计步骤

### 步骤 1：概述与架构设计
- [x] 1.1 编写设计概述（项目定位、设计范围、技术选型确认）
- [x] 1.2 绘制整体架构图（Mermaid C4 模型，展示两个项目的关系和分层）
- [x] 1.3 定义项目工程结构（monorepo vs 独立项目的组织方式）

### 步骤 2：PC 管理端（spmp-web-pc）组件与接口设计
- [x] 2.1 布局组件设计（AppLayout / AppHeader / AppSidebar / AppBreadcrumb）
- [x] 2.2 路由架构设计（静态路由 + 动态路由占位、路由守卫逻辑）
- [x] 2.3 菜单配置数据结构设计（7 个业务模块的菜单定义）
- [x] 2.4 登录页与 Mock 登录流程设计
- [x] 2.5 占位页面与 404 页面设计

### 步骤 3：H5 业主端（spmp-web-h5）组件与接口设计
- [x] 3.1 布局组件设计（TabBarLayout / NavBarLayout）
- [x] 3.2 路由架构设计（Tab 路由 + 子页面路由、过渡动画）
- [x] 3.3 首页快捷入口卡片设计
- [x] 3.4 登录页与 Mock 登录流程设计
- [x] 3.5 移动端适配方案设计（postcss-px-to-viewport 配置）

### 步骤 4：公共基础设施设计
- [x] 4.1 Axios 请求封装设计（拦截器、错误处理、Token 注入）
- [x] 4.2 Pinia 状态管理设计（user store / app store 结构）
- [x] 4.3 路由守卫框架设计（登录检查、权限检查占位）
- [x] 4.4 CSS 变量与全局样式设计（与 UI 规范对齐）
- [x] 4.5 ESLint + Prettier + TypeScript 配置方案

### 步骤 5：数据模型定义
- [x] 5.1 TypeScript 类型定义（用户信息、菜单项、路由元信息、API 响应等）

### 步骤 6：正确性属性分析
- [x] 6.1 使用 prework 工具分析验收标准的可测试性
- [x] 6.2 编写正确性属性（Correctness Properties）

### 步骤 7：错误处理与测试策略
- [x] 7.1 错误处理策略（网络错误、认证失败、路由异常等）
- [x] 7.2 测试策略（单元测试 + 属性测试方案）

---

## 三、澄清问题

在开始设计前，需要确认以下问题：

### Q1：项目组织方式
[Question] 两个前端项目（spmp-web-pc 和 spmp-web-h5）是否采用 monorepo 方式管理（如 pnpm workspace），还是完全独立的两个项目？需求文档中目录结构写的是 `src/spmp-web-pc/` 和 `src/spmp-web-h5/`，看起来是在同一个 workspace 下。请确认：
- 方案 A：pnpm workspace monorepo，共享公共代码（utils、types、styles）
- 方案 B：完全独立的两个项目，各自独立的 package.json，公共代码各自维护一份

[Answer] 完全独立

### Q2：多标签页（Tags View）功能
[Question] UI 规范中提到 PC 端"支持多标签页（Tags View），可关闭、关闭其他、关闭全部"。在框架搭建阶段是否需要实现 Tags View 功能，还是仅作为占位，后续 CR 再实现？

[Answer] 需要实现

### Q3：动态路由与权限
[Question] 需求中提到"路由守卫框架（登录检查占位、权限检查占位）"。框架搭建阶段的菜单和路由是否全部静态配置（硬编码 7 个模块），后续再改为从后端 API 动态获取？还是现在就需要预留动态路由注册的完整机制？

[Answer] 预留动态路由注册

### Q4：Mock 登录的实现方式
[Question] Mock 登录是使用前端硬编码的方式（直接在代码中模拟 Token 生成和存储），还是需要引入 Mock 服务（如 MSW / vite-plugin-mock）来模拟后端 API 响应？

[Answer] 代码中模拟

### Q5：CSS 方案确认
[Question] 需求中提到 UnoCSS 为"可选"。请确认：
- 是否在框架搭建阶段引入 UnoCSS？
- 还是仅使用 CSS 变量 + 组件库自带样式，保持简洁？

[Answer] 保持简洁

### Q6：环境变量与代理配置
[Question] 开发环境下的 API 代理配置，是否需要在 vite.config.ts 中预配置 proxy 规则指向后端服务地址（如 `localhost:8080`），还是仅定义环境变量占位即可？

[Answer] 预配置proxy

### Q7：国际化（i18n）
[Question] 框架搭建阶段是否需要预留国际化支持（vue-i18n），还是当前阶段不考虑，后续有需求再引入？

[Answer] 预留国际化

---

## 四、参考文档

| 文档 | 路径 |
|------|------|
| 需求规格说明书 | `requirements.md`（同目录） |
| UI 设计规范 | `CUST-AIDEV-FRONTEND-DOCS/global-info/frontend-ui-specification.md` |
| 前端路由说明 | `CUST-AIDEV-FRONTEND-DOCS/global-info/frontend-pages-route.md` |
| 前端 API 对应关系 | `CUST-AIDEV-FRONTEND-DOCS/global-info/frontend-page-with-api.md` |
| 技术栈规范 | `.trae/rules/tech.md` |

---

## 五、审批

请审阅以上设计计划和澄清问题，确认后我将按步骤逐步执行，输出 `design.md`。
