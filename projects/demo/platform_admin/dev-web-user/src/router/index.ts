/**
 * 路由配置
 * - /login → 登录页（无需认证）
 * - / → 首页（需要 token，无 token 重定向 /login）
 * Layout 布局由 App.vue 根据构建模式注入
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
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/home/index.vue'),
    meta: { title: '首页', requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

/** 路由守卫：无 token 时重定向到登录页 */
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('access_token')

  if (to.meta.requiresAuth === false) {
    // 不需要认证的页面直接放行
    next()
  } else if (!token) {
    // 无 token，跳转登录
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else {
    next()
  }
})

export default router
