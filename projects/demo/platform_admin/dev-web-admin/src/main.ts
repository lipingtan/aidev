import { createApp } from 'vue'
import App from './App.vue'

// Element Plus 全量引入
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
// Element Plus 图标全局注册（侧边栏动态菜单依赖）
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

// H5 布局组件依赖 Vant（运行时 isMobile() 检测后自动使用 H5Layout）
import Vant from 'vant'
import 'vant/lib/index.css'

// 全局样式
import './styles/global.css'

// Pinia 状态管理
import pinia from './store'

// 路由
import router from './router'
import { setupRouterGuard } from './router/guard'
import { useAppStore } from './store/modules/app'
import { initRouteManager } from './plugin-loader/route-manager'

// 国际化
import i18n from './locales'

// 暴露宿主库（供 UMD 格式插件共享依赖，PC 端和 H5 端均需要）
import * as Vue from 'vue'
import * as VueRouter from 'vue-router'
import * as VantLib from 'vant'
;(window as any).__PLATFORM_ADMIN_LIBS__ = { Vue, ElementPlus, VueRouter, Vant: VantLib }

const app = createApp(App)

// 注册插件
app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.use(Vant)
app.use(i18n)

// 全局注册所有 Element Plus 图标（侧边栏动态菜单图标渲染依赖）
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

const appStore = useAppStore()
appStore.initTheme()

// 注册路由守卫
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
