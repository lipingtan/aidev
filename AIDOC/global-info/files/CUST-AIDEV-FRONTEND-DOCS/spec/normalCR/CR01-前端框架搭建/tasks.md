# CR01 - 前端框架搭建 实施任务列表

| 项目 | 内容 |
|------|------|
| 需求编号 | FE-CR01 |
| 需求名称 | 前端主体页面框架搭建 |
| 版本 | 1.0 |
| 创建日期 | 2026-04-17 |
| 状态 | 待执行 |

---

## 概述

本任务列表基于 `requirements.md` 和 `design.md`，将前端框架搭建拆解为可独立执行和验证的编码任务。PC 管理端（spmp-web-pc）和 H5 业主端（spmp-web-h5）为两个完全独立的项目，部分基础设施逻辑一致，可并行开发。

**技术栈**：Vue 3.4 + Vite 5 + TypeScript 5 + Vue Router 4 + Pinia 2 + Axios

---

## 并行分组说明

| 并行组 | 任务 | 说明 |
|--------|------|------|
| A | 1, 2 | PC 端和 H5 端工程初始化可并行 |
| B | 3, 4 | PC 端和 H5 端公共基础设施可并行 |
| C | 6, 7 | PC 端和 H5 端布局组件可并行 |
| D | 9, 10 | PC 端和 H5 端路由与守卫可并行 |
| E | 12, 13 | PC 端和 H5 端页面开发可并行 |

---

## 任务列表

### 并行组 A：工程骨架初始化

- [ ] 1. 初始化 PC 管理端（spmp-web-pc）工程骨架
  - [ ] 1.1 使用 Vite 5 创建 Vue 3 + TypeScript 项目，配置 `package.json` 依赖（Vue 3.4、Element Plus 2.7、Vue Router 4、Pinia 2、Axios 1.7、vue-i18n 9）
    - 项目位于 `src/spmp-web-pc/`
    - 开发端口 3000
    - _需求: US-01, US-03_
  - [ ] 1.2 配置 `tsconfig.json`（strict 模式、路径别名 `@/*`）和 `tsconfig.node.json`
    - _需求: US-03（TypeScript 严格模式）_
  - [ ] 1.3 配置 `vite.config.ts`（端口 3000、路径别名、proxy 代理 `/api` → `localhost:8080`）
    - _需求: US-01, US-03_
  - [ ] 1.4 配置 `.eslintrc.cjs`（vue3-recommended + typescript-eslint + prettier）和 `.prettierrc`
    - _需求: US-03（ESLint + Prettier 代码规范）_
  - [ ] 1.5 创建环境变量文件 `.env.development` 和 `.env.production`（`VITE_API_BASE_URL=/api/v1`）
    - _需求: US-03_
  - [ ] 1.6 创建 `index.html` 入口文件和 `src/main.ts` 应用入口（注册 Vue、Router、Pinia、Element Plus、i18n）
    - _需求: US-01_
  - [ ] 1.7 创建 `src/App.vue`（仅包含 `<router-view />`）
    - _需求: US-01_

- [ ] 2. 初始化 H5 业主端（spmp-web-h5）工程骨架
  - [ ] 2.1 使用 Vite 5 创建 Vue 3 + TypeScript 项目，配置 `package.json` 依赖（Vue 3.4、Vant 4、Vue Router 4、Pinia 2、Axios 1.7、vue-i18n 9、postcss-px-to-viewport-8-plugin）
    - 项目位于 `src/spmp-web-h5/`
    - 开发端口 3001
    - _需求: US-02, US-03_
  - [ ] 2.2 配置 `tsconfig.json`（strict 模式、路径别名 `@/*`）和 `tsconfig.node.json`
    - _需求: US-03（TypeScript 严格模式）_
  - [ ] 2.3 配置 `vite.config.ts`（端口 3001、路径别名、proxy 代理 `/api` → `localhost:8080`）
    - _需求: US-02, US-03_
  - [ ] 2.4 配置 `postcss.config.cjs`（postcss-px-to-viewport-8-plugin，viewportWidth: 375，排除 Vant）
    - _需求: US-02（适配 375px 基准宽度）_
  - [ ] 2.5 配置 `.eslintrc.cjs` 和 `.prettierrc`
    - _需求: US-03（ESLint + Prettier 代码规范）_
  - [ ] 2.6 创建环境变量文件 `.env.development` 和 `.env.production`
    - _需求: US-03_
  - [ ] 2.7 创建 `index.html` 入口文件和 `src/main.ts` 应用入口（注册 Vue、Router、Pinia、Vant、i18n）
    - _需求: US-02_
  - [ ] 2.8 创建 `src/App.vue`（包含 `<router-view />` 和过渡动画容器）
    - _需求: US-02（页面切换有过渡动画）_


### 并行组 B：公共基础设施（类型定义 + 样式 + 请求封装 + 国际化）

- [ ] 3. PC 端公共基础设施
  - [ ] 3.1 创建 TypeScript 类型定义 `src/types/index.ts`
    - 定义 `ApiResponse<T>`、`PageParams`、`PageResult<T>`、`LoginParams`、`LoginResult`、`UserInfo`、`MenuItem`、`TagView`、`AppState` 等类型
    - 扩展 `vue-router` 的 `RouteMeta` 类型（title、icon、hidden、requiresAuth、affix、permission、noCache）
    - _需求: US-03, Design 五_
  - [ ] 3.2 创建 CSS 变量文件 `src/styles/variables.css` 和全局样式 `src/styles/global.css`
    - CSS 变量包含主色 `#1890FF`、功能色、中性色、业务状态色、间距、字体、PC 端布局尺寸、圆角、阴影
    - 全局样式包含 box-sizing 重置、滚动条美化
    - _需求: US-01（主题色和布局符合 UI 规范）, US-03（统一的 CSS 变量定义）_
  - [ ] 3.3 创建 Axios 请求封装 `src/utils/request.ts`
    - 创建 Axios 实例（baseURL 从环境变量读取、timeout 15000ms）
    - 请求拦截器：从 localStorage 读取 `access_token` 注入 `Authorization: Bearer {token}`
    - 响应拦截器：业务错误使用 `ElMessage.error()` 提示、401 清除 Token 跳转登录页
    - _需求: US-03（Axios 请求封装、统一拦截器、错误处理、Token 注入）_
  - [ ] 3.4 创建国际化配置 `src/locales/index.ts`、`src/locales/zh-CN.ts`、`src/locales/en-US.ts`
    - 使用 vue-i18n 9 Composition API 模式，默认 zh-CN
    - 语言包包含 common 和 login 两个模块的占位文案
    - _需求: Design 4.6（国际化预留）_

- [ ] 4. H5 端公共基础设施
  - [ ] 4.1 创建 TypeScript 类型定义 `src/types/index.ts`
    - 定义 `ApiResponse<T>`、`PageParams`、`PageResult<T>`、`LoginParams`、`LoginResult`、`UserInfo`、`AppState` 等类型
    - 扩展 `vue-router` 的 `RouteMeta` 类型（title、requiresAuth、tabBar）
    - _需求: US-03, Design 五_
  - [ ] 4.2 创建 CSS 变量文件 `src/styles/variables.css` 和全局样式 `src/styles/global.css`
    - CSS 变量包含主色、功能色、中性色、H5 端布局尺寸（navbar 44px、tabbar 50px）
    - 全局样式包含 box-sizing 重置、安全区域适配（safe-area-inset-bottom）
    - _需求: US-02（适配 375px 基准宽度）, US-03（统一的 CSS 变量定义）_
  - [ ] 4.3 创建 Axios 请求封装 `src/utils/request.ts`
    - 逻辑与 PC 端一致，错误提示使用 Vant 的 `showToast()`
    - _需求: US-03（Axios 请求封装）_
  - [ ] 4.4 创建国际化配置 `src/locales/index.ts`、`src/locales/zh-CN.ts`、`src/locales/en-US.ts`
    - _需求: Design 4.6（国际化预留）_

- [ ] 5. 检查点 — 确保两个项目工程骨架可正常启动
  - 确保 `pnpm install` 和 `pnpm dev` 可正常运行，如有问题请向用户确认。


### 并行组 C：状态管理 + 布局组件

- [ ] 6. PC 端状态管理与布局组件
  - [ ] 6.1 创建 Pinia Store：`src/store/index.ts`、`src/store/modules/user.ts`、`src/store/modules/app.ts`、`src/store/modules/tags-view.ts`
    - `useUserStore`：token / username / avatar / roles 状态，login()（Mock 生成 Token + 硬编码用户信息）、logout()（清除状态跳转登录页）、getUserInfo()（预留）
    - `useAppStore`：sidebarCollapsed / locale 状态，toggleSidebar()，持久化到 localStorage
    - `useTagsViewStore`：visitedViews 列表，addView() / closeView() / closeOtherViews() / closeAllViews()，首页标签 affix 不可关闭
    - _需求: US-01, US-03（Pinia 状态管理）, Design 2.6, 4.2_
  - [ ]* 6.2 编写 TagsView Store 单元测试
    - 测试添加/关闭/关闭其他/关闭全部标签操作
    - 测试 affix 标签不可关闭
    - _需求: US-01, Design 8.3_
  - [ ] 6.3 创建布局组件 `src/layout/AppLayout.vue`
    - 使用 el-container / el-header / el-aside / el-main 组合布局
    - 通过 appStore.sidebarCollapsed 控制侧边栏宽度（220px / 64px）
    - _需求: US-01（顶部导航栏 + 左侧菜单栏 + 右侧内容区）_
  - [ ] 6.4 创建 `src/layout/AppHeader.vue` 顶部导航栏
    - 高度 64px，左侧 Logo + 系统名称「智慧物业管理平台」+ 折叠按钮，右侧用户头像 + 用户名 + 下拉菜单（退出登录）
    - 背景色 #FFFFFF，底部 1px 边框 #D9D9D9
    - _需求: US-01（顶部导航栏：Logo + 系统名称 + 用户头像）_
  - [ ] 6.5 创建 `src/layout/AppSidebar.vue` 侧边菜单栏
    - 使用 el-menu router 模式，从 asyncRoutes children 中提取菜单项（过滤 hidden）
    - 支持折叠，高亮当前路由
    - _需求: US-01（左侧菜单栏可折叠、7 个业务模块菜单占位）_
  - [ ] 6.6 创建 `src/layout/AppBreadcrumb.vue` 面包屑导航
    - 根据 route.matched 自动生成，首项固定「首页」链接到 /home，末项不可点击
    - 使用 el-breadcrumb 组件
    - _需求: US-01（包含面包屑导航）_
  - [ ] 6.7 创建 `src/layout/TagsView.vue` 多标签页
    - 使用 el-tag 渲染标签列表，点击切换页面
    - 关闭单个标签（首页不可关闭）、右键菜单（关闭当前/关闭其他/关闭全部）
    - 标签栏水平滚动，当前标签高亮主色 #1890FF
    - _需求: US-01, Design 2.6_

- [ ] 7. H5 端状态管理与布局组件
  - [ ] 7.1 创建 Pinia Store：`src/store/index.ts`、`src/store/modules/user.ts`、`src/store/modules/app.ts`
    - `useUserStore`：与 PC 端逻辑一致
    - `useAppStore`：locale 状态（无 sidebarCollapsed）
    - _需求: US-02, US-03（Pinia 状态管理）_
  - [ ] 7.2 创建 `src/layout/TabBarLayout.vue` 底部 Tab 栏布局
    - 顶部 van-nav-bar 显示当前路由 meta.title
    - 底部 van-tabbar 5 个 Tab（首页/报修/缴费/公告/我的），safe-area-inset-bottom 适配
    - 内容区 router-view 包裹 transition 过渡动画
    - _需求: US-02（底部 Tab 栏、页面切换有过渡动画）_
  - [ ] 7.3 创建 `src/layout/NavBarLayout.vue` 顶部导航栏布局
    - van-nav-bar 左侧返回箭头 + 标题居中，点击返回调用 router.back()
    - 无底部 Tab 栏
    - _需求: US-02, Design 3.1_
  - [ ] 7.4 创建页面过渡动画 CSS（slide-left / slide-right）
    - Tab 页面根据索引判断滑动方向，子页面进入 slide-left、返回 slide-right
    - _需求: US-02（页面切换有过渡动画）_

- [ ] 8. 检查点 — 确保布局组件渲染正常
  - 确保所有测试通过，如有问题请向用户确认。


### 并行组 D：路由配置与路由守卫

- [ ] 9. PC 端路由与守卫
  - [ ] 9.1 创建静态路由配置 `src/router/static-routes.ts`
    - 定义 `constantRoutes`：登录页（/login，requiresAuth: false）、404 页面（/:pathMatch(.*)*, requiresAuth: false）
    - 定义 `asyncRoutes`：AppLayout 布局下 8 个子路由（home、base-data、user-auth、owner、workorder、billing、notice、access），每个路由配置 meta（title、icon），业务模块指向 PlaceholderView
    - 首页路由 meta.affix = true（标签页固定不可关闭）
    - _需求: US-01（7 个业务模块菜单占位、点击菜单切换占位页面）_
  - [ ] 9.2 创建路由实例 `src/router/index.ts`
    - 使用 createWebHistory 模式，注册 constantRoutes + asyncRoutes
    - 预留 `setupDynamicRoutes()` 动态路由注册方法（注释说明后续从后端 API 获取菜单并使用 router.addRoute()）
    - _需求: US-01, Design 2.2（动态路由注册预留）_
  - [ ] 9.3 创建路由守卫 `src/router/guard.ts`
    - beforeEach 守卫：requiresAuth === false 直接放行；无 Token 重定向 /login（携带 redirect 参数）；已登录访问 /login 重定向 /home；预留权限检查注释
    - 路由切换时自动调用 tagsViewStore.addView() 添加标签
    - _需求: US-03（路由守卫框架、登录检查占位、权限检查占位）_
  - [ ]* 9.4 编写路由守卫单元测试
    - 测试未登录重定向、已登录放行、登录页重定向
    - _需求: US-03, Design 8.3_
  - [ ]* 9.5 编写属性测试 — Property 6: 路由守卫认证检查
    - **Property 6: 路由守卫认证检查**
    - 对于任意路由跳转，当目标路由 meta.requiresAuth !== false 且无 Token 时，守卫应重定向到 /login 并携带 redirect 参数；当 Token 存在时应放行
    - **验证: 需求 US-03（3.2）**

- [ ] 10. H5 端路由与守卫
  - [ ] 10.1 创建静态路由配置 `src/router/static-routes.ts`
    - 定义 `constantRoutes`：登录页、404 页面
    - 定义 `tabRoutes`：TabBarLayout 布局下 5 个 Tab 子路由（home、workorder、billing、notice、mine），每个路由 meta.tabBar = true
    - _需求: US-02（底部 Tab 栏：首页/报修/缴费/公告/我的）_
  - [ ] 10.2 创建路由实例 `src/router/index.ts`
    - 使用 createWebHistory 模式，注册 constantRoutes + tabRoutes
    - _需求: US-02_
  - [ ] 10.3 创建路由守卫 `src/router/guard.ts`
    - 逻辑与 PC 端一致（登录检查 + 权限预留）
    - 在 beforeEach 中根据 Tab 索引设置 transitionName（slide-left / slide-right）
    - _需求: US-02（页面切换有过渡动画）, US-03（路由守卫框架）_

- [ ] 11. 检查点 — 确保路由跳转和守卫逻辑正常
  - 确保所有测试通过，如有问题请向用户确认。


### 并行组 E：页面开发

- [ ] 12. PC 端页面开发
  - [ ] 12.1 创建登录页 `src/views/login/LoginView.vue`
    - 使用 el-form + el-input + el-button，表单校验（用户名非空、密码非空）
    - 调用 userStore.login()，成功后 router.replace(redirect || '/home')
    - 页面居中卡片布局，标题「智慧物业管理平台」
    - _需求: US-01（展示登录页、mock 登录）_
  - [ ]* 12.2 编写属性测试 — Property 1: Mock 登录 Token 存储往返
    - **Property 1: Mock 登录 Token 存储往返**
    - 对于任意非空用户名和密码，调用 userStore.login() 后，localStorage 中应存在非空 Token，且 userStore.token 与之一致
    - **验证: 需求 US-01（1.2）, US-02（2.2）**
  - [ ] 12.3 创建首页 `src/views/home/HomeView.vue`
    - 仪表盘占位页面，居中显示「首页 — 仪表盘」
    - 使用 el-card 展示欢迎信息
    - _需求: US-01_
  - [ ] 12.4 创建占位页面 `src/views/placeholder/PlaceholderView.vue`
    - 根据 route.meta.title 动态显示「{模块名} — 开发中」
    - 使用 el-empty 组件
    - _需求: US-01（点击菜单切换到对应占位页面）_
  - [ ] 12.5 创建 404 页面 `src/views/error/NotFoundView.vue`
    - 使用 el-result 组件，sub-title「页面不存在」，提供「返回首页」按钮跳转 /home
    - _需求: US-01（包含 404 页面）_
  - [ ]* 12.6 编写属性测试 — Property 2: 导航项路由映射正确性
    - **Property 2: 导航项路由映射正确性**
    - 对于任意静态路由配置中的菜单项，其 path 应对应已注册路由，且路由 meta.title 与菜单显示标题一致
    - **验证: 需求 US-01（1.5）**
  - [ ]* 12.7 编写属性测试 — Property 4: 未知路径匹配 404
    - **Property 4: 未知路径匹配 404**
    - 对于任意不在路由配置中的路径字符串，路由解析结果应匹配到 not-found 路由
    - **验证: 需求 US-01（1.7）**

- [ ] 13. H5 端页面开发
  - [ ] 13.1 创建登录页 `src/views/login/LoginView.vue`
    - 使用 van-field + van-button，表单校验（用户名非空、密码非空）
    - 调用 userStore.login()，成功后 router.replace('/home')
    - _需求: US-02（展示登录页、mock 登录）_
  - [ ] 13.2 创建首页 `src/views/home/HomeView.vue`
    - 顶部欢迎语 + 小区名称（硬编码占位）
    - 中部 van-grid 2×2 快捷入口卡片（报修 orders-o → /workorder、缴费 balance-o → /billing、公告 bell → /notice、访客预约 friends-o → /visitor 占位）
    - 下部最新公告摘要列表（占位显示「暂无公告」）
    - _需求: US-02（首页展示快捷入口卡片）_
  - [ ] 13.3 创建 Tab 占位页面
    - `src/views/workorder/WorkorderView.vue`（报修 — 开发中）
    - `src/views/billing/BillingView.vue`（缴费 — 开发中）
    - `src/views/notice/NoticeView.vue`（公告 — 开发中）
    - `src/views/mine/MineView.vue`（我的 — 开发中）
    - 使用 van-empty 组件展示占位信息
    - _需求: US-02（每个 Tab 对应一个占位页面）_
  - [ ] 13.4 创建 404 页面 `src/views/error/NotFoundView.vue`
    - 使用 van-empty 组件，提供「返回首页」按钮
    - _需求: US-02_

- [ ] 14. 检查点 — 确保所有页面可正常访问和交互
  - 确保所有测试通过，如有问题请向用户确认。


### 串行：集成联调与补充测试

- [ ] 15. 集成联调与端到端验证
  - [ ] 15.1 PC 端集成联调
    - 验证完整流程：登录 → 主布局渲染（Header + Sidebar + TagsView + Breadcrumb + Content）→ 菜单点击切换占位页 → 标签页操作 → 退出登录
    - 确保侧边栏折叠/展开正常，面包屑随路由更新
    - _需求: US-01 全部验收标准_
  - [ ] 15.2 H5 端集成联调
    - 验证完整流程：登录 → Tab 栏布局渲染 → Tab 切换 → 首页快捷入口点击 → 页面过渡动画 → 退出登录
    - 确保 375px 基准宽度适配正常，安全区域适配正常
    - _需求: US-02 全部验收标准_
  - [ ]* 15.3 编写属性测试 — Property 3: 面包屑根据路由链生成
    - **Property 3: 面包屑根据路由链生成**
    - 对于任意已注册路由，面包屑生成函数应根据 route.matched 产出面包屑项列表，第一项始终为「首页」，最后一项为当前路由 meta.title
    - **验证: 需求 US-01（1.6）**
  - [ ]* 15.4 编写属性测试 — Property 5: 请求拦截器 Token 注入
    - **Property 5: 请求拦截器 Token 注入**
    - 对于任意 HTTP 请求，当 localStorage 中存在 access_token 时，请求头应包含 Authorization: Bearer {token}；不存在时不应包含
    - **验证: 需求 US-03（3.1）**
  - [ ]* 15.5 编写 Axios 请求封装单元测试
    - 测试 Token 注入、401 处理、业务错误处理
    - _需求: US-03, Design 8.3_

- [ ] 16. 最终检查点 — 确保所有测试通过，项目可正常运行
  - 确保 PC 端 `pnpm dev` 端口 3000 可正常访问
  - 确保 H5 端 `pnpm dev` 端口 3001 可正常访问
  - 确保 ESLint 检查无报错
  - 确保所有单元测试和属性测试通过
  - 如有问题请向用户确认。

---

## 依赖关系

```
并行组 A（任务 1, 2）→ 并行组 B（任务 3, 4）→ 检查点 5
                                                    ↓
                                              并行组 C（任务 6, 7）→ 检查点 8
                                                                        ↓
                                                                  并行组 D（任务 9, 10）→ 检查点 11
                                                                                              ↓
                                                                                        并行组 E（任务 12, 13）→ 检查点 14
                                                                                                                      ↓
                                                                                                                任务 15 → 最终检查点 16
```

## 备注

- 标记 `*` 的子任务为可选测试任务，可跳过以加速 MVP 交付
- 每个任务引用了具体的需求条目，确保需求可追溯
- 属性测试验证设计文档中定义的 6 个正确性属性
- 单元测试和属性测试互补，单元测试捕获具体 Bug，属性测试验证通用正确性
- PC 端和 H5 端为完全独立项目，同一并行组内的任务可分配给不同开发人员同时执行
