/**
 * 路由配置 — 模式 A（访问时触发登录）
 *
 * requiresAuth 三态：
 *   false   → 公开路由，匿名直接访问
 *   'soft'  → 软鉴权，可访问但触碰受限操作时弹登录框（默认值）
 *   true    → 强制登录，无 token 直接跳登录页
 */

import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录', requiresAuth: false }
  },
  {
    path: '/',
    redirect: '/home'
  },
  // 公开路由 — 匿名可访问
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/home/index.vue'),
    meta: { title: '首页', requiresAuth: false }
  },
  {
    path: '/bulletin',
    name: 'Bulletin',
    component: () => import('@/views/bulletin/index.vue'),
    meta: { title: '公告栏', requiresAuth: false }
  },
  // 强制登录路由 — 无 token 直接跳登录页
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/home/index.vue'), // 占位，后续替换
    meta: { title: '个人中心', requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

/**
 * 从路径第一段推断插件名
 * /notice → "notice"
 * /plugin/notice/list → "notice"
 */
function inferPluginName(path: string): string {
  const segments = path.split('/').filter(Boolean)
  if (segments[0] === 'plugin') return segments[1] || ''
  return segments[0] || ''
}

/**
 * 路由守卫
 * 1. requiresAuth === true 且无 token → 跳登录页
 * 2. 路由不存在时 → 尝试加载插件 bundle 后重新导航（插件懒加载兜底）
 */
router.beforeEach(async (to, _from, next) => {
  const token = localStorage.getItem('access_token')

  // 强制登录路由
  if (to.meta.requiresAuth === true && !token) {
    next({ path: '/login', query: { redirect: to.fullPath } })
    return
  }

  // 路由已知，直接放行
  if (to.name) {
    next()
    return
  }

  // 路由未知（可能是插件路由还未注册）
  // 尝试推断插件名，动态加载 bundle 后重新导航
  const pluginName = inferPluginName(to.path)
  if (pluginName && pluginName !== 'login') {
    try {
      const module = await import(/* @vite-ignore */ `/static/plugins/${pluginName}/index.js`)
      if (module.routes && Array.isArray(module.routes)) {
        for (const route of module.routes) {
          try { router.addRoute(route) } catch {}
        }
        // 重新导航，此时路由已注册
        next({ path: to.fullPath, replace: true })
        return
      }
    } catch {
      // 插件 bundle 不存在，继续走 404
    }
  }

  next()
})

export default router
