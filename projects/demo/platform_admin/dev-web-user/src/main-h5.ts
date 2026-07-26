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
