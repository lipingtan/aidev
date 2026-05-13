import { createApp } from 'vue'
import { ElMessage } from 'element-plus'

import Cookies from 'js-cookie'

import 'normalize.css/normalize.css' // a modern alternative to CSS resets

import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'

import '@/styles/index.scss' // global css
import '@/styles/admin.scss'

import { Codemirror } from 'vue-codemirror'

import App from './App'
import store from './store'
import router from './router'
import permission from './directive/permission'

import { getDicts } from '@/api/admin/dict/data'
import { getItems, setItems } from '@/api/table'
import { getConfigKey } from '@/api/admin/sys-config'
import { parseTime, resetForm, addDateRange, selectDictLabel, /* download,*/ selectItemsLabel } from '@/utils/costum'
import { dialogDrag } from '@/utils/dialog' // dialog directive
import { setupErrorHandler } from '@/utils/error-log' // error log

import SvgIcon from './icons' // icon
import './permission' // permission control

// import Viser from 'viser-vue'
// Note: viser-vue 不支持 Vue 3，需要后续处理

import * as filters from './filters' // global filters

import Pagination from '@/components/Pagination'
import BasicLayout from '@/layout/BasicLayout'

// import VueParticles from 'vue-particles'
// Note: vue-particles 不支持 Vue 3，需要后续处理

import '@/utils/dialog'

// import VueDND from 'awe-dnd'
// Note: awe-dnd 不支持 Vue 3，已改用 vue3-dnd

import 'remixicon/fonts/remixicon.css'

console.info(`欢迎使用go-admin，谢谢您对我们的支持，在使用过程中如果有什么问题，
请访问https://github.com/go-admin-team/go-admin 或者
 https://github.com/go-admin-team/go-admin-ui 向我们反馈，
 谢谢！`)

// 创建 Vue 应用实例
const app = createApp(App)

// 全局方法挂载
app.config.globalProperties.$getDicts = getDicts
app.config.globalProperties.$getItems = getItems
app.config.globalProperties.$setItems = setItems
app.config.globalProperties.$getConfigKey = getConfigKey
app.config.globalProperties.$parseTime = parseTime
app.config.globalProperties.$resetForm = resetForm
app.config.globalProperties.$addDateRange = addDateRange
app.config.globalProperties.$selectDictLabel = selectDictLabel
app.config.globalProperties.$selectItemsLabel = selectItemsLabel
// app.config.globalProperties.$download = download

// 全局过滤器改为全局方法
app.config.globalProperties.$filters = filters

// 消息提示方法
app.config.globalProperties.$msgSuccess = function(msg) {
  ElMessage({ showClose: true, message: msg, type: 'success' })
}

app.config.globalProperties.$msgError = function(msg) {
  ElMessage({ showClose: true, message: msg, type: 'error' })
}

app.config.globalProperties.$msgInfo = function(msg) {
  ElMessage.info(msg)
}

// 全局组件注册
app.component('AppPagination', Pagination)
app.component('BasicLayout', BasicLayout)
app.component('CodeEditor', Codemirror)
app.component('SvgIcon', SvgIcon)

// 注册插件
app.use(store)
app.use(router)
app.use(permission)
app.use(ElementPlus, {
  locale: zhCn,
  size: Cookies.get('size') || 'default'
})

// 注册自定义指令
app.directive('dialogDrag', dialogDrag)

// 设置错误处理器
setupErrorHandler(app)

// 挂载应用
app.mount('#app')
