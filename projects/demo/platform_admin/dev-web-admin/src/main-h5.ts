import { createApp } from 'vue'
import App from './App.vue'

// Element Plus 全量引入（H5 仍需要，业务表单使用）
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
// Element Plus 图标全局注册
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

// H5 专用：Vant
import Vant from 'vant'
import 'vant/lib/index.css'

// Pinia 状态管理
import pinia from './store'

// 路由
import router from './router'
import { setupRouterGuard } from './router/guard'
import { useAppStore } from './store/modules/app'
import { initRouteManager } from './plugin-loader/route-manager'

// 国际化
import i18n from './locales'

import * as Vue from 'vue'
import * as VueRouter from 'vue-router'

// 暴露宿主库（H5 端额外暴露 Vant）
;(window as any).__PLATFORM_ADMIN_LIBS__ = {
  Vue,
  ElementPlus,
  VueRouter,
  Vant,
}

const app = createApp(App)

app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.use(Vant)
app.use(i18n)

// 全局注册所有 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

const appStore = useAppStore()
appStore.initTheme()

setupRouterGuard(router)

// 初始化插件路由管理器
initRouteManager(router)

app.mount('#app')

// DEV 模式：暴露 plugin-loader 到 window，供 E2E 测试调用
if (import.meta.env.DEV) {
  import('./plugin-loader').then(({ loadPlugin, unloadPlugin }) => {
    ;(window as any).loadPlugin = loadPlugin
    ;(window as any).unloadPlugin = unloadPlugin
  })
}
