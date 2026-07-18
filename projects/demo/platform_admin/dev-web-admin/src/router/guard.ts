/**
 * 路由守卫
 * - 未认证页面直接放行
 * - 无 access_token 重定向 /login（携带 redirect）
 * - 已登录访问 /login 重定向 /home
 * - 登录后自动加载用户信息和菜单
 * - 路由切换时自动添加 TagsView 标签
 */

import type { Router } from 'vue-router'
import { useTagsViewStore } from '@/store/modules/tags-view'
import { useUserStore } from '@/store/modules/user'
import { checkInitStatus } from '@/api/init'
import { ElLoading } from 'element-plus'

/** 初始化状态缓存 */
let initChecked = false
let isInitialized = true
/** 用户信息是否已加载 */
let userInfoLoaded = false

/** 全局 loading 实例（路由切换专用，不用引用计数） */
let routeLoading: ReturnType<typeof ElLoading.service> | null = null

function startLoading() {
  if (!routeLoading) {
    routeLoading = ElLoading.service({ fullscreen: true, text: '加载中...', background: 'rgba(0,0,0,0.2)' })
  }
}

function stopLoading() {
  routeLoading?.close()
  routeLoading = null
}

/** 无需认证的白名单路径 */
const WHITE_LIST = ['/login', '/init', '/tenant-select']

export function setupRouterGuard(router: Router): void {
  router.beforeEach(async (to, _from, next) => {
    startLoading()

    try {
      // 初始化页面直接放行
      if (to.path === '/init') {
        return next()
      }

      // 首次访问时检查初始化状态（超时 3s，不阻塞）
      if (!initChecked) {
        try {
          const controller = new AbortController()
          const timer = setTimeout(() => controller.abort(), 3000)
          const status = await checkInitStatus()
          clearTimeout(timer)
          isInitialized = status.initialized
        } catch {
          isInitialized = true
        }
        initChecked = true
      }

      // 未初始化，重定向到初始化页面
      if (!isInitialized) {
        stopLoading()
        return next({ path: '/init' })
      }

      const accessToken = localStorage.getItem('access_token')
      const platformToken = localStorage.getItem('platform_token')

      // 不需要认证的页面直接放行
      if (to.meta.requiresAuth === false) {
        if (to.path === '/login' && accessToken) {
          stopLoading()
          return next({ path: '/home' })
        }
        return next()
      }

      // 未登录
      if (!accessToken) {
        if (platformToken && to.path !== '/tenant-select') {
          stopLoading()
          return next({ path: '/tenant-select' })
        }
        if (WHITE_LIST.includes(to.path)) {
          return next()
        }
        stopLoading()
        return next({ path: '/login', query: { redirect: to.fullPath } })
      }

      // 已登录访问登录页
      if (to.path === '/login') {
        stopLoading()
        return next({ path: '/home' })
      }

      // 登录后首次访问，加载用户信息和菜单
      if (!userInfoLoaded) {
        try {
          const userStore = useUserStore()
          await userStore.fetchUserInfo()
          await userStore.fetchMenus()
          // 插件加载（非阻塞）
          try {
            const { loadAllPlugins } = await import('@/core/plugin-loader')
            const { mergePluginRoutes } = await import('@/core/plugin-router')
            const { usePluginStore } = await import('@/store/modules/plugin')
            const pluginStore = usePluginStore()
            const loadedConfigs = await loadAllPlugins()
            loadedConfigs.forEach(config => {
              pluginStore.addLoadedModule(config.manifest.name, config)
            })
            mergePluginRoutes(router, loadedConfigs)
          } catch (e) {
            console.error('[Guard] 插件加载失败:', e)
          }
          userInfoLoaded = true
        } catch {
          localStorage.removeItem('access_token')
          localStorage.removeItem('platform_token')
          localStorage.removeItem('current_tenant_id')
          userInfoLoaded = false
          stopLoading()
          return next({ path: '/login', query: { redirect: to.fullPath } })
        }
      }

      // 路由切换时添加标签
      const tagsViewStore = useTagsViewStore()
      tagsViewStore.addView(to)

      next()
    } catch (e) {
      console.error('[Guard] 路由守卫异常:', e)
      stopLoading()
      next()
    }
  })

  router.afterEach(() => {
    stopLoading()
  })
}

/** 重置守卫状态（登出时调用） */
export function resetGuardState(): void {
  userInfoLoaded = false
}
