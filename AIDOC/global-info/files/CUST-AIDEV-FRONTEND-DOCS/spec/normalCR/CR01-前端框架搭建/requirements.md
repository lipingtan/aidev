# CR01 - 前端框架搭建 需求规格说明书

| 项目 | 内容 |
|------|------|
| 需求编号 | FE-CR01 |
| 需求名称 | 前端主体页面框架搭建 |
| 版本 | 1.0 |
| 创建日期 | 2026-04-17 |
| 状态 | 草稿 |

---

## 一、需求概述

搭建 SPMP 智慧物业管理平台的前端主体页面框架，包括 PC 管理端（spmp-web-pc）和 H5 业主端（spmp-web-h5）两个项目。要求创建完整的工程骨架、布局组件、路由框架、请求封装和基础样式，能够直接运行并看到页面效果。

---

## 二、用户故事

### US-01：PC 管理端框架

**作为** 前端开发人员，
**我希望** PC 管理端工程能直接运行，展示完整的管理后台布局（顶部导航 + 侧边栏 + 内容区），
**以便** 后续业务模块开发时只需在框架内添加页面。

验收标准：
- 执行 `pnpm dev` 后可在浏览器访问
- 展示登录页（mock 登录：输入任意用户名密码，模拟 Token 存储后跳转主页）
- 登录后进入管理后台布局：顶部导航栏（Logo + 系统名称 + 用户头像）、左侧菜单栏（可折叠）、右侧内容区
- 侧边栏展示 7 个业务模块的菜单占位（基础数据、用户权限、业主管理、工单管理、缴费管理、公告管理、门禁管理）
- 侧边栏菜单按路由路径前缀自动分组为子菜单（如 `base/` 前缀归入"基础数据"分组，`owner/` 前缀归入"业主管理"分组），具体分组规则见 `CUST-AIDEV-FRONTEND-DOCS/global-info/frontend-pages-route.md` 中的"PC 管理端侧边栏菜单分组规范"章节
- 点击菜单可切换到对应的占位页面（显示"XX模块 - 开发中"）
- 包含面包屑导航
- 包含 404 页面
- 主题色和布局符合 UI 规范文档

### US-02：H5 业主端框架

**作为** 前端开发人员，
**我希望** H5 业主端工程能直接运行，展示移动端布局（底部 Tab 栏 + 内容区），
**以便** 后续业务模块开发时只需在框架内添加页面。

验收标准：
- 执行 `pnpm dev` 后可在浏览器访问（建议使用 Chrome 移动端模拟）
- 展示登录页（mock 登录）
- 登录后进入主页面：底部 Tab 栏（首页/报修/缴费/公告/我的）
- 每个 Tab 对应一个占位页面
- 首页展示快捷入口卡片（报修、缴费、公告、访客预约）
- 页面切换有过渡动画
- 适配 375px 基准宽度

### US-03：公共基础设施

**作为** 前端开发人员，
**我希望** 两个项目都包含统一的基础设施代码，
**以便** 后续开发时有一致的开发体验。

验收标准：
- Axios 请求封装（统一拦截器、错误处理、Token 注入占位）
- 路由守卫框架（登录检查占位、权限检查占位）
- Pinia 状态管理（user store 占位、app store 占位）
- 统一的 CSS 变量定义（主题色、字号、间距，与 UI 规范一致）
- ESLint + Prettier 代码规范配置
- TypeScript 严格模式

---

## 三、技术栈

### PC 管理端（spmp-web-pc）

| 技术 | 版本 | 说明 |
|------|------|------|
| Vue | 3.4.x | 使用 Composition API + `<script setup>` |
| Vite | 5.x | 构建工具 |
| TypeScript | 5.x | 类型安全 |
| Element Plus | 2.7.x | UI 组件库 |
| Vue Router | 4.x | 路由（history 模式） |
| Pinia | 2.x | 状态管理 |
| Axios | 1.7.x | HTTP 客户端 |
| UnoCSS | 0.60.x | 原子化 CSS（可选） |

### H5 业主端（spmp-web-h5）

| 技术 | 版本 | 说明 |
|------|------|------|
| Vue | 3.4.x | 使用 Composition API + `<script setup>` |
| Vite | 5.x | 构建工具 |
| TypeScript | 5.x | 类型安全 |
| Vant | 4.x | 移动端 UI 组件库 |
| Vue Router | 4.x | 路由（history 模式） |
| Pinia | 2.x | 状态管理 |
| Axios | 1.7.x | HTTP 客户端 |
| postcss-px-to-viewport | - | 移动端适配（375px 基准） |

---

## 四、目录结构要求

### PC 管理端

```
src/spmp-web-pc/
├── src/
│   ├── api/                    # API 接口定义（占位）
│   ├── assets/                 # 静态资源（图标、图片）
│   ├── components/             # 公共组件
│   ├── layout/                 # 布局组件
│   │   ├── AppLayout.vue       # 主布局（顶部+侧边栏+内容区）
│   │   ├── AppHeader.vue       # 顶部导航
│   │   ├── AppSidebar.vue      # 侧边菜单
│   │   └── AppBreadcrumb.vue   # 面包屑
│   ├── router/                 # 路由配置
│   │   └── index.ts
│   ├── store/                  # Pinia 状态管理
│   │   ├── modules/
│   │   │   ├── user.ts         # 用户状态（占位）
│   │   │   └── app.ts          # 应用状态
│   │   └── index.ts
│   ├── styles/                 # 全局样式
│   │   ├── variables.css       # CSS 变量（主题色、间距）
│   │   └── global.css          # 全局样式重置
│   ├── utils/                  # 工具函数
│   │   └── request.ts          # Axios 封装
│   ├── views/                  # 页面组件
│   │   ├── login/              # 登录页
│   │   ├── home/               # 首页（仪表盘占位）
│   │   ├── error/              # 404 页面
│   │   └── placeholder/        # 模块占位页
│   ├── App.vue
│   └── main.ts
├── index.html
├── vite.config.ts
├── tsconfig.json
├── package.json
├── .eslintrc.cjs
└── .prettierrc
```

### H5 业主端

```
src/spmp-web-h5/
├── src/
│   ├── api/
│   ├── assets/
│   ├── components/
│   ├── layout/
│   │   ├── TabBarLayout.vue    # 底部 Tab 栏布局
│   │   └── NavBarLayout.vue    # 顶部导航栏布局
│   ├── router/
│   │   └── index.ts
│   ├── store/
│   │   ├── modules/
│   │   │   ├── user.ts
│   │   │   └── app.ts
│   │   └── index.ts
│   ├── styles/
│   │   ├── variables.css
│   │   └── global.css
│   ├── utils/
│   │   └── request.ts
│   ├── views/
│   │   ├── login/
│   │   ├── home/              # 首页（快捷入口）
│   │   ├── workorder/         # 报修 Tab 占位
│   │   ├── billing/           # 缴费 Tab 占位
│   │   ├── notice/            # 公告 Tab 占位
│   │   ├── mine/              # 我的 Tab 占位
│   │   └── error/
│   ├── App.vue
│   └── main.ts
├── index.html
├── vite.config.ts
├── tsconfig.json
├── package.json
├── .eslintrc.cjs
└── .prettierrc
```

---

## 五、非功能性需求

- 两个项目都能通过 `pnpm install` + `pnpm dev` 直接运行
- PC 端开发服务器端口：3000
- H5 端开发服务器端口：3001
- 首屏加载时间 < 2s（开发环境）
- 代码通过 ESLint 检查无报错
