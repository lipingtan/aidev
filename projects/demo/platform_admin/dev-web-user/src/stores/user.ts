/**
 * 用户状态管理 — 模式 A
 * - 静态菜单（无需登录）始终显示
 * - 登录后追加动态菜单
 * - 提供 requireLogin() 方法供页面内触发登录弹窗
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, logout as logoutApi, getMenu } from '@/api/auth'
import { loadPluginRoutesFromMenus, clearPluginRouteCache } from '@/utils/plugin-loader'
import type { MenuItem } from '@/types'
import router from '@/router'

/** 静态菜单：无需登录也始终显示 */
const STATIC_MENUS: MenuItem[] = [
  { id: 'home',     title: '首页',  path: '/home',     icon: '' },
  { id: 'bulletin', title: '公告栏', path: '/bulletin', icon: '' },
]

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('access_token') || '')
  const phone = ref<string>(localStorage.getItem('user_phone') || '')

  /** 动态菜单（登录后追加） */
  const dynamicMenus = ref<MenuItem[]>([])

  /** 最终展示的菜单 = 静态 + 动态（去重） */
  const menus = computed<MenuItem[]>(() => {
    const dynamicPaths = new Set(dynamicMenus.value.map(m => m.path))
    const filtered = STATIC_MENUS.filter(m => !dynamicPaths.has(m.path))
    return [...filtered, ...dynamicMenus.value]
  })

  /** 是否已登录 */
  const isLoggedIn = computed(() => !!token.value)

  /** 登录弹窗控制（由布局层监听） */
  const showLoginDialog = ref(false)

  /**
   * requireLogin — 模式 A 核心方法
   * 已登录：执行回调
   * 未登录：弹出登录框，登录成功后执行回调
   */
  function requireLogin(callback?: () => void) {
    if (token.value) {
      callback?.()
    } else {
      showLoginDialog.value = true
      if (callback) {
        pendingCallback.value = callback
      }
    }
  }

  const pendingCallback = ref<(() => void) | null>(null)

  /** 登录成功后执行 pending 回调 */
  function onLoginSuccess() {
    showLoginDialog.value = false
    const cb = pendingCallback.value
    pendingCallback.value = null
    cb?.()
  }

  /** 登录 */
  async function login(phoneVal: string, code: string, tenantCode: string) {
    const res = await loginApi(phoneVal, code, tenantCode)
    token.value = res.token
    phone.value = phoneVal
    localStorage.setItem('access_token', res.token)
    localStorage.setItem('user_phone', phoneVal)
    await fetchMenu()
    onLoginSuccess()
  }

  /** 登出 */
  async function logout() {
    try { await logoutApi() } catch {}
    token.value = ''
    phone.value = ''
    dynamicMenus.value = []
    clearPluginRouteCache()
    localStorage.removeItem('access_token')
    localStorage.removeItem('user_phone')
    router.push('/home')  // 登出后回首页而不是登录页
  }

  /** 获取动态菜单（登录后调用），并自动加载插件路由 */
  async function fetchMenu() {
    // 已登录：从认证接口获取完整菜单（含权限过滤）
    // 未登录：从公开接口获取 requiresAuth=false 的菜单（插件动态菜单）
    try {
      let data: MenuItem[]
      if (token.value) {
        data = await getMenu()
      } else {
        // 公开菜单接口（无需 token），获取 platform=user 的所有菜单
        const res = await fetch('/api/v1/public/user-menu?platform=user')
        const json = await res.json()
        const raw: any[] = (json as any)?.data ?? []
        // 后端返回 name 字段，前端 MenuItem 用 title 字段，做映射
        data = raw.map(item => ({
          id: String(item.id),
          parentId: item.parent_id,
          name: item.name,
          title: item.name,        // 关键：映射 name → title
          path: item.path,
          component: item.component || 'plugin/container',
          icon: item.icon || '',
          sort: item.sort_order || 0,
          hidden: item.hidden || false,
          children: []
        }))
      }
      dynamicMenus.value = data
      await loadPluginRoutesFromMenus(data, router)
    } catch {
      dynamicMenus.value = []
    }
  }

  return {
    token,
    phone,
    menus,
    isLoggedIn,
    showLoginDialog,
    requireLogin,
    login,
    logout,
    fetchMenu
  }
})

