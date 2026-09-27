import { createApp } from 'vue'
import App from './App.vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import Vant from 'vant'
import 'vant/lib/index.css'
import * as Vue from 'vue'
import * as VueRouter from 'vue-router'
import * as VantLib from 'vant'
import { createPinia } from 'pinia'
import router from './router'
import '@/styles/reset.css'
import { useUserStore } from '@/stores/user'
import {
  ElMessage, ElMessageBox, ElNotification,
  ElButton, ElInput, ElSelect, ElOption, ElForm, ElFormItem,
  ElTable, ElTableColumn, ElPagination, ElDialog, ElDrawer,
  ElCard, ElTag, ElBadge, ElAvatar, ElImage,
  ElDropdown, ElDropdownMenu, ElDropdownItem,
  ElPopconfirm, ElPopover, ElTooltip,
  ElDatePicker, ElCheckbox, ElCheckboxGroup,
  ElRadio, ElRadioGroup, ElSwitch, ElUpload,
  ElProgress, ElEmpty, ElResult, ElAlert,
  ElRow, ElCol, ElDivider, ElScrollbar,
  ElTabs, ElTabPane, ElTree, ElCascader,
  ElMenu, ElMenuItem, ElSubMenu,
  ElBreadcrumb, ElBreadcrumbItem
} from 'element-plus'

// User 端宿主库（供插件 bundle 共享依赖，避免双实例问题）
// ElementPlus 必须是组件映射对象，与 element-plus.js shim 结构对齐
;(window as any).__PLATFORM_USER_LIBS__ = {
  Vue,
  ElementPlus: {
    ElMessage, ElMessageBox, ElNotification,
    ElButton, ElInput, ElSelect, ElOption, ElForm, ElFormItem,
    ElTable, ElTableColumn, ElPagination, ElDialog, ElDrawer,
    ElCard, ElTag, ElBadge, ElAvatar, ElImage,
    ElDropdown, ElDropdownMenu, ElDropdownItem,
    ElPopconfirm, ElPopover, ElTooltip,
    ElDatePicker, ElCheckbox, ElCheckboxGroup,
    ElRadio, ElRadioGroup, ElSwitch, ElUpload,
    ElProgress, ElEmpty, ElResult, ElAlert,
    ElRow, ElCol, ElDivider, ElScrollbar,
    ElTabs, ElTabPane, ElTree, ElCascader,
    ElMenu, ElMenuItem, ElSubMenu,
    ElBreadcrumb, ElBreadcrumbItem
  },
  VueRouter,
  Vant: VantLib
}

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)
app.use(ElementPlus)
app.use(Vant)

// 注册所有 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 在 router install 之前预加载插件路由：
// fetchMenu() 会从公开接口拉取 user 端菜单，并将插件路由注册到 router。
// 这样 app.use(router) 执行时插件路由已就位，避免 "No match found" warn。
const userStore = useUserStore(pinia)
await userStore.fetchMenu()

app.use(router)

app.mount('#app')
