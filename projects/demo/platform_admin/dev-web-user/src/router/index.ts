/**
 * 路由配置
 * - /login → 登录页（无需认证）
 * - / → 基础布局（需要 token，无 token 重定向 /login）
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
    name: 'Layout',
    component: () => import('@/views/layout/index.vue'),
    meta: { requiresAuth: true },
    redirect: '/home',
    children: [
      {
        path: 'home',
        name: 'Home',
        component: () => import('@/views/home/index.vue'),
        meta: { title: '首页' }
      }
    ]
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
