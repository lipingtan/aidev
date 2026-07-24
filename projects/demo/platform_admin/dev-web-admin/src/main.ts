import { createApp } from 'vue'
import App from './App.vue'

// Element Plus 全量引入
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
// Element Plus 图标全局注册（侧边栏动态菜单依赖）
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

// 全局样式
import './styles/global.css'

// Pinia 状态管理
import pinia from './store'

// 路由
import router from './router'
import { setupRouterGuard } from './router/guard'
import { useAppStore } from './store/modules/app'

// 国际化
import i18n from './locales'

const app = createApp(App)

// 注册插件
app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.use(i18n)

// 全局注册所有 Element Plus 图标（侧边栏动态菜单图标渲染依赖）
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

const appStore = useAppStore()
appStore.initTheme()

// 注册路由守卫
setupRouterGuard(router)

app.mount('#app')
