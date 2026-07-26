# 设计：V2-CR6 多端前端架构 + Plugin SDK V2

## 技术方案概述

| 决策点 | 选型 |
|--------|------|
| 前端构建 | Vite 多入口（共享 src/，不同 main-pc.ts / main-h5.ts） |
| admin H5 布局 | 汉堡菜单 + 顶部标题栏（Vant NavBar + Popup） |
| user H5 布局 | 底部 Tab 导航（Vant Tabbar） |
| admin H5 组件 | 布局层 Vant，业务表单 Element Plus |
| 插件 bundle 格式 | UMD |
| 插件依赖共享 | 宿主暴露 `window.__PLATFORM_ADMIN_LIBS__` |
| teardown | `onTeardown(fn)` 注册模式 + 5s 超时 |
| H5 性能验证 | Lighthouse CI |

---

## 一、前端工程结构改造

### 1.1 dev-web-admin 双入口结构

```
dev-web-admin/
├── src/
│   ├── views/              # 业务页面（PC/H5 共享，不改动）
│   ├── api/                # API 模块（共享，不改动）
│   ├── store/              # Pinia Store（共享，不改动）
│   ├── router/             # 路由（共享，不改动）
│   ├── locales/            # i18n（共享，不改动）
│   ├── components/         # 通用组件（共享，不改动）
│   ├── layout/             # 现有布局（重构为 PC 专用）
│   │   ├── pc/             # PC 布局（侧边栏+顶栏，现有逻辑迁入）
│   │   │   └── PcLayout.vue
│   │   └── h5/             # H5 布局（汉堡菜单+顶部标题栏，新增）
│   │       ├── H5Layout.vue
│   │       ├── H5Header.vue      # 顶部标题栏（Vant NavBar）
│   │       └── H5MenuDrawer.vue  # 侧滑菜单（Vant Popup）
│   ├── plugin-sdk/         # Plugin SDK V2（扩展）
│   ├── plugin-loader/      # 插件加载器（新增）
│   ├── main.ts             # 保留（或重命名为 main-pc.ts 入口）
│   └── main-h5.ts          # H5 专用入口（新增）
├── index.html              # PC 入口 HTML
├── index-h5.html           # H5 入口 HTML（新增）
└── vite.config.ts          # 修改为多入口配置
```

### 1.2 dev-web-user 双入口结构

```
dev-web-user/
├── src/
│   ├── views/              # 业务页面（共享）
│   ├── api/                # API 模块（共享）
│   ├── stores/             # Pinia Store（共享）
│   ├── router/             # 路由（共享）
│   ├── layout/
│   │   ├── pc/
│   │   │   └── UserPcLayout.vue   # PC 布局（Element Plus，现有逻辑迁入）
│   │   └── h5/
│   │       └── UserH5Layout.vue   # H5 布局（Vant Tabbar，新增）
│   ├── main.ts             # PC 入口（重命名参考，实际保留）
│   └── main-h5.ts          # H5 专用入口（新增）
├── index.html
├── index-h5.html           # H5 入口 HTML（新增）
└── vite.config.ts          # 修改为多入口配置
```

**dev-web-user 依赖说明：** `package.json` 需新增 `vant` 依赖（dev-web-user 当前 node_modules 中有 `@vant`，确认版本为 vant@4.x）。

---

## 二、Vite 多入口配置

### 2.1 产物分离策略（重要）

Vite 的 `rollupOptions.input` 多入口模式会将所有产物输出到同一 `outDir`，无法天然分离到 `dist-pc/` 和 `dist-h5/`。**正确实现方式：分次构建，每次指定不同 `outDir`**。

```typescript
// vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig(({ mode }) => {
  const isH5 = mode === 'h5'
  return {
    plugins: [vue()],
    resolve: {
      alias: { '@': resolve(__dirname, 'src') }
    },
    build: {
      // 根据 mode 切换输出目录和入口
      outDir: isH5 ? 'dist-h5' : 'dist-pc',
      rollupOptions: {
        input: isH5
          ? resolve(__dirname, 'index-h5.html')
          : resolve(__dirname, 'index.html')
      }
    },
    server: {
      port: 3000,
      proxy: {
        '/api':   { target: 'http://localhost:8000', changeOrigin: true },
        '/auth':  { target: 'http://localhost:8000', changeOrigin: true },
        '/setup': { target: 'http://localhost:8000', changeOrigin: true },
        '/static':{ target: 'http://localhost:8000', changeOrigin: true }
      }
    }
  }
})
```

dev-web-user 的 `vite.config.ts` 采用相同模式（port 改为 5174）。

### 2.2 package.json 新增构建命令

```json
{
  "scripts": {
    "dev":        "vite",
    "dev:h5":     "vite --mode h5",
    "build:pc":   "vite build --mode pc",
    "build:h5":   "vite build --mode h5",
    "build":      "npm run build:pc && npm run build:h5"
  }
}
```

---

## 三、H5 入口文件

### 3.1 dev-web-admin/index-h5.html

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no" />
    <title>管理后台</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main-h5.ts"></script>
  </body>
</html>
```

### 3.2 dev-web-admin/src/main.ts（PC 入口，补充 `__PLATFORM_ADMIN_LIBS__`）

PC 入口同样需要暴露宿主库，供 PC 端插件共享依赖：

```typescript
import { createApp } from 'vue'
import App from './App.vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import * as Vue from 'vue'
import * as VueRouter from 'vue-router'
import pinia from './store'
import router from './router'
import { setupRouterGuard } from './router/guard'
import { useAppStore } from './store/modules/app'
import i18n from './locales'
import './styles/global.css'

// 暴露宿主库（供 UMD 格式插件共享依赖，PC 端和 H5 端均需要）
;(window as any).__PLATFORM_ADMIN_LIBS__ = { Vue, ElementPlus, VueRouter }

const app = createApp(App)
app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.use(i18n)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
const appStore = useAppStore()
appStore.initTheme()
setupRouterGuard(router)
app.mount('#app')
```

### 3.3 dev-web-admin/src/main-h5.ts（H5 入口）

```typescript
import { createApp } from 'vue'
import App from './App.vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import Vant from 'vant'
import 'vant/lib/index.css'
import * as Vue from 'vue'
import * as VueRouter from 'vue-router'
import pinia from './store'
import router from './router'
import { setupRouterGuard } from './router/guard'
import { useAppStore } from './store/modules/app'
import i18n from './locales'

// H5 入口额外暴露 Vant（PC 入口不引入 Vant）
;(window as any).__PLATFORM_ADMIN_LIBS__ = { Vue, ElementPlus, VueRouter, Vant }

const app = createApp(App)
app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.use(Vant)
app.use(i18n)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
const appStore = useAppStore()
appStore.initTheme()
setupRouterGuard(router)
app.mount('#app')
```

### 3.4 dev-web-user/src/main-h5.ts（User 端 H5 入口，新增）

```typescript
import { createApp } from 'vue'
import App from './App.vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import Vant from 'vant'
import 'vant/lib/index.css'
import * as Vue from 'vue'
import * as VueRouter from 'vue-router'
import { createPinia } from 'pinia'
import router from './router'
import '@/styles/reset.css'

// User 端 H5 宿主库（供插件共享，platform='user', device='h5'）
;(window as any).__PLATFORM_USER_LIBS__ = { Vue, ElementPlus, VueRouter, Vant }

const app = createApp(App)
app.use(createPinia())
app.use(ElementPlus)
app.use(Vant)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
app.use(router)
app.mount('#app')
```

### 3.5 dev-web-user/src/main.ts（User 端 PC 入口，补充 `__PLATFORM_USER_LIBS__`）

PC 入口同样需要暴露宿主库，供 user-pc 类型插件共享依赖：

```typescript
import { createApp } from 'vue'
import App from './App.vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import * as Vue from 'vue'
import * as VueRouter from 'vue-router'
import { createPinia } from 'pinia'
import router from './router'
import '@/styles/reset.css'

// User 端 PC 宿主库（供插件共享，platform='user', device='pc'）
;(window as any).__PLATFORM_USER_LIBS__ = { Vue, ElementPlus, VueRouter }

const app = createApp(App)
app.use(createPinia())
app.use(ElementPlus)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
app.use(router)
app.mount('#app')
```

> **注意**：user 端使用 `window.__PLATFORM_USER_LIBS__` 而非 `__PLATFORM_ADMIN_LIBS__`，插件开发时根据 platform 选择读取哪个命名空间。

---

## 四、H5 布局组件

### 4.1 dev-web-admin/src/layout/h5/H5Layout.vue（核心结构）

```vue
<template>
  <!-- Vant NavBar 顶部标题栏 -->
  <van-nav-bar
    :title="currentPageTitle"
    left-arrow
    @click-left="onBack"
  >
    <template #right>
      <van-icon name="wap-nav" size="20" @click="showMenu = true" />
    </template>
  </van-nav-bar>

  <!-- 业务内容区 -->
  <div class="h5-content">
    <router-view />
  </div>

  <!-- 侧滑菜单 Popup -->
  <van-popup v-model:show="showMenu" position="left" :style="{ width: '75%', height: '100%' }">
    <H5MenuDrawer :menus="menus" @close="showMenu = false" />
  </van-popup>
</template>
```

### 4.2 dev-web-user/src/layout/h5/UserH5Layout.vue（底部 Tab）

```vue
<template>
  <div class="user-h5-container">
    <div class="user-h5-content">
      <router-view />
    </div>
    <!-- Vant Tabbar 底部导航 -->
    <van-tabbar v-model="activeTab" route>
      <van-tabbar-item
        v-for="item in tabMenus"
        :key="item.path"
        :to="item.path"
        :icon="item.icon"
      >{{ item.title }}</van-tabbar-item>
    </van-tabbar>
  </div>
</template>
```

---

## 五、Plugin SDK V2

### 5.1 扩展后的类型定义（types.ts 变更）

```typescript
// 新增：取消注册函数
export type UnregisterFn = () => void

// 新增：teardown 清理函数
export type TeardownFn = () => void | Promise<void>

// 扩展 PluginContext（在现有基础上新增字段）
export interface PluginContext {
  currentUser: Readonly<UserInfo>
  currentTenant: Readonly<TenantInfo>
  permissions: Readonly<string[]>
  eventBus: EventBus
  // V1 已有，返回值变更：void → UnregisterFn
  registerExtension: (point: string, component: any, sort?: number) => UnregisterFn
  // V2 新增
  platform: 'admin' | 'user'
  device: 'pc' | 'h5'
  getPlatform: () => 'admin' | 'user'
  getDevice: () => 'pc' | 'h5'
  onTeardown: (fn: TeardownFn) => void
  i18n: {
    mergeLocale: (locale: string, messages: Record<string, any>) => void
  }
}

// 扩展 PluginConfig：setup 兼容 V1，teardown 废弃（改用 onTeardown 注册）
export interface PluginConfig {
  manifest: PluginManifest
  routes?: PluginRoute[]
  menus?: PluginMenu[]
  permissions?: string[]
  extensions?: Record<string, any>
  setup?: (ctx: PluginContext) => void | Promise<void>
  // teardown 字段保留用于兼容旧写法，推荐用 ctx.onTeardown()
  teardown?: (ctx: PluginContext) => void | Promise<void>
}

// 扩展 PluginManifest：新增 frontends 字段
export interface PluginManifest {
  name: string
  version: string
  displayName: string
  description: string
  platforms: ('admin' | 'user' | 'both')[]
  dependencies?: Record<string, string>
  frontendEntry?: string  // V1 兼容保留
  // V2 新增：多端 bundle 配置
  frontends?: Array<{
    platform: 'admin' | 'user'
    device: 'pc' | 'h5'
    entry: string  // 相对路径，如 'admin-pc/bundle.js'
  }>
}
```

### 5.2 plugin-loader 核心实现

新文件：`dev-web-admin/src/plugin-loader/index.ts`

**依赖说明：**
- `doRegisterExtension(point, component, sort)` — 实现在 `src/plugin-loader/extension-registry.ts`，维护全局扩展点 Map，返回 `UnregisterFn`
- `registerPluginRoutes(pluginName, routes)` / `unregisterPluginRoutes(pluginName)` — 实现在 `src/plugin-loader/route-manager.ts`，调用 `router.addRoute` / `router.removeRoute`
- `globalEventBus` — 实现在 `src/plugin-loader/event-bus.ts`，使用 `mitt` 创建，在 `main.ts` 初始化后挂载

```typescript
import type { PluginConfig, PluginContext, TeardownFn, UnregisterFn } from '../plugin-sdk/types'
import { useAuthStore } from '../store/modules/auth'
import { useI18n } from 'vue-i18n'
import { doRegisterExtension } from './extension-registry'
import { registerPluginRoutes, unregisterPluginRoutes } from './route-manager'
import { globalEventBus } from './event-bus'

// 扩展点注册表（plugin-name → unregisterFn[]）
const extensionRegistry = new Map<string, UnregisterFn[]>()

// teardown 队列（plugin-name → TeardownFn[]）
const teardownQueues = new Map<string, TeardownFn[]>()

// 已加载插件的 PluginConfig 缓存（用于 V1 兼容）
const loadedConfigs = new Map<string, PluginConfig>()

const TEARDOWN_TIMEOUT_MS = 5000

/**
 * 创建插件上下文
 */
function createPluginContext(
  pluginName: string,
  platform: 'admin' | 'user',
  device: 'pc' | 'h5'
): PluginContext {
  const authStore = useAuthStore()
  const { mergeLocaleMessage } = useI18n()

  teardownQueues.set(pluginName, [])
  extensionRegistry.set(pluginName, [])

  return {
    currentUser: authStore.userInfo as any,
    currentTenant: authStore.tenantInfo as any,
    permissions: authStore.permissions,
    eventBus: globalEventBus,
    platform,
    device,
    getPlatform: () => platform,
    getDevice: () => device,
    registerExtension: (point, component, sort) => {
      const unregister = doRegisterExtension(point, component, sort)
      extensionRegistry.get(pluginName)!.push(unregister)
      return unregister
    },
    onTeardown: (fn: TeardownFn) => {
      teardownQueues.get(pluginName)!.push(fn)
    },
    i18n: {
      mergeLocale: (locale, messages) => {
        mergeLocaleMessage(locale, messages)
      }
    }
  }
}

/**
 * 加载插件 bundle（UMD 格式，通过 <script> 标签注入）
 */
export async function loadPlugin(
  pluginName: string,
  platform: 'admin' | 'user',
  device: 'pc' | 'h5'
): Promise<void> {
  const bundlePath = `/static/plugins/${pluginName}/${platform}-${device}/bundle.js`

  return new Promise((resolve, reject) => {
    // 避免重复加载
    if (document.querySelector(`script[data-plugin="${pluginName}"]`)) {
      resolve()
      return
    }

    const script = document.createElement('script')
    script.src = bundlePath
    script.setAttribute('data-plugin', pluginName)

    script.onload = async () => {
      const globalKey = `__PLUGIN_${pluginName.toUpperCase().replace(/-/g, '_')}__`
      const pluginConfig: PluginConfig = (window as any)[globalKey]

      if (!pluginConfig) {
        reject(new Error(`[plugin-loader] 插件 ${pluginName} 未正确注册到 window.${globalKey}`))
        return
      }

      loadedConfigs.set(pluginName, pluginConfig)
      const ctx = createPluginContext(pluginName, platform, device)

      try {
        if (pluginConfig.setup) {
          await pluginConfig.setup(ctx)
        }
        if (pluginConfig.routes) {
          registerPluginRoutes(pluginName, pluginConfig.routes)
        }
        resolve()
      } catch (e) {
        reject(e)
      }
    }

    script.onerror = () => {
      reject(new Error(`[plugin-loader] 加载插件 bundle 失败: ${bundlePath}`))
    }

    document.head.appendChild(script)
  })
}

/**
 * 卸载插件：执行 teardown（含 V1 兼容）→ 清理扩展点 → 移除路由
 */
export async function unloadPlugin(pluginName: string): Promise<void> {
  const fns = teardownQueues.get(pluginName) ?? []

  // V1 兼容：如果插件使用了旧式 PluginConfig.teardown 字段，也加入队列
  const config = loadedConfigs.get(pluginName)
  const ctx = { getPlatform: () => 'admin' as const, getDevice: () => 'pc' as const } as any
  if (config?.teardown) {
    fns.push(() => config.teardown!(ctx))
  }

  const teardownPromise = Promise.all(fns.map(fn => fn()))
  const timeoutPromise = new Promise<void>((_, reject) =>
    setTimeout(() => reject(new Error(`[plugin-loader] 插件 ${pluginName} teardown 超时`)), TEARDOWN_TIMEOUT_MS)
  )

  try {
    await Promise.race([teardownPromise, timeoutPromise])
  } catch (e) {
    console.warn(e)
    // 超时后强制继续卸载
  }

  // 清理扩展点
  const unregisters = extensionRegistry.get(pluginName) ?? []
  unregisters.forEach(fn => fn())
  extensionRegistry.delete(pluginName)
  teardownQueues.delete(pluginName)
  loadedConfigs.delete(pluginName)

  // 移除路由
  unregisterPluginRoutes(pluginName)

  // 清理 window 全局变量
  const globalKey = `__PLUGIN_${pluginName.toUpperCase().replace(/-/g, '_')}__`
  delete (window as any)[globalKey]
}
```

### 5.3 辅助模块说明

| 文件 | 职责 |
|------|------|
| `src/plugin-loader/extension-registry.ts` | 维护全局扩展点 Map（`point → ComponentDef[]`），提供 `doRegisterExtension` |
| `src/plugin-loader/route-manager.ts` | 封装 `router.addRoute` / `router.removeRoute`，按插件名分组管理路由 |
| `src/plugin-loader/event-bus.ts` | 用 `mitt` 创建 `globalEventBus`，在 `main.ts` 创建 app 前初始化 |

---

## 六、后端插件 bundle 部署

### 6.1 manifest.frontends 字段（plugin.json 扩展）

```json
{
  "name": "game",
  "version": "1.0.0",
  "frontends": [
    { "platform": "admin", "device": "pc",  "entry": "admin-pc/bundle.js" },
    { "platform": "admin", "device": "h5",  "entry": "admin-h5/bundle.js" },
    { "platform": "user",  "device": "pc",  "entry": "user-pc/bundle.js" }
  ]
}
```

### 6.2 backend/common/plugin/installer.go 新增逻辑

```go
// deployFrontendBundles 安装插件时部署前端 bundle 到静态目录
func (inst *Installer) deployFrontendBundles(pluginName string, pluginDir string, frontends []FrontendConfig) error {
    for _, fe := range frontends {
        // 源文件路径（插件包解压目录下）
        src := filepath.Join(pluginDir, "frontend", fe.Entry)
        // 目标路径：/static/plugins/{name}/{platform}-{device}/bundle.js
        destDir := filepath.Join(inst.staticDir, "plugins", pluginName, fe.Platform+"-"+fe.Device)
        destFile := filepath.Join(destDir, "bundle.js")

        if err := os.MkdirAll(destDir, 0755); err != nil {
            return fmt.Errorf("创建目录失败 %s: %w", destDir, err)
        }
        if err := copyFile(src, destFile); err != nil {
            return fmt.Errorf("复制 bundle 失败 %s → %s: %w", src, destFile, err)
        }
    }
    return nil
}

// removeFrontendBundles 卸载插件时删除所有前端 bundle
func (inst *Installer) removeFrontendBundles(pluginName string) error {
    pluginStaticDir := filepath.Join(inst.staticDir, "plugins", pluginName)
    return os.RemoveAll(pluginStaticDir)
}
```

### 6.3 静态文件路由注册

在 `backend/cmd/api/server.go`（或路由注册处）新增：

```go
// 插件前端 bundle 静态文件服务
r.Static("/static/plugins", "./static/plugins")
```

---

## 七、API 设计（无新增后端接口）

本 CR 无新增 REST API，后端改动仅为：
1. `installer.go` 新增 `deployFrontendBundles` / `removeFrontendBundles` 逻辑
2. 路由层新增 `/static/plugins/*` 静态文件服务

---

## 八、不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 现有 PC 端登录、菜单加载、权限校验流程不受影响 | `/auth/login` + `/api/v1/common/user-menu?platform=admin` 正常响应 |
| RG-2 | dev-web-admin PC 构建产物与改造前功能等价 | `npm run build:pc` 后 dist-pc/ 可正常部署访问 |
| RG-3 | dev-web-user PC 构建产物与改造前功能等价 | `npm run build:pc` 后 dist-pc/ 可正常部署访问 |
| RG-4 | 现有插件（V1 格式）setup 函数正常调用 | 已安装 game 插件加载不报错 |
| RG-5 | 插件 stop/uninstall 后菜单和路由不残留 | 操作后刷新页面确认插件菜单消失 |
| RG-6 | 后端静态文件服务不影响现有 `/api/` 路由 | 现有接口响应不变 |

---

## 九、正确性属性

- 双入口构建：`npm run build:pc` 产物只包含 PC 布局代码；`npm run build:h5` 产物只包含 H5 布局代码
- 依赖共享：插件 bundle 加载后使用的 Vue 实例 === 宿主 `window.__PLATFORM_ADMIN_LIBS__.Vue`，`===` 严格相等
- teardown 完整性：`unloadPlugin` 完成后，`extensionRegistry` 和 `teardownQueues` 中不存在该插件的记录
- 超时保护：teardown 超过 5000ms 后，`unloadPlugin` 必须继续执行后续清理，不得永久挂起
- 错误隔离：单个插件 bundle 加载失败（404/语法错误），不影响宿主主应用和其他插件正常运行
- bundle 路径规范：后端部署路径格式严格为 `/static/plugins/{name}/{platform}-{device}/bundle.js`，platform 和 device 均为小写
