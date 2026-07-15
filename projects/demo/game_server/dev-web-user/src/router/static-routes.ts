/**
 * H5 端静态路由配置
 * - constantRoutes：不需要布局的路由（登录页、个人信息、404）
 * - tabRoutes：TabBarLayout 布局下的 Tab 子路由
 */

import type { RouteRecordRaw } from 'vue-router'

/** 不需要布局的路由 */
export const constantRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'h5-login',
    component: () => import('@/views/login/LoginView.vue'),
    meta: { title: '登录', requiresAuth: false }
  },
  {
    path: '/register',
    name: 'h5-register',
    component: () => import('@/views/register/index.vue'),
    meta: { title: '注册', requiresAuth: false }
  },
  {
    path: '/mine/profile',
    name: 'h5-profile',
    component: () => import('@/views/mine/profile.vue'),
    meta: { title: '个人信息' }
  },
  {
    path: '/plugin/:pluginName/:pathMatch(.*)*',
    name: 'h5-plugin-container',
    component: () => import('@/views/plugin/PluginContainer.vue'),
    meta: { title: '插件页面' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'h5-not-found',
    component: () => import('@/views/error/NotFoundView.vue'),
    meta: { title: '页面不存在', requiresAuth: false }
  }
]

/** Tab 栏路由 */
export const tabRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layout/TabBarLayout.vue'),
    redirect: '/home',
    children: [
      {
        path: 'home',
        name: 'h5-home',
        component: () => import('@/views/home/HomeView.vue'),
        meta: { title: '首页', tabBar: true, requiresAuth: false }
      },
      {
        path: 'mine',
        name: 'h5-mine',
        component: () => import('@/views/mine/MineView.vue'),
        meta: { title: '我的', tabBar: true, requiresAuth: false }
      }
    ]
  }
]

/**
 * Tab 路由路径与索引映射
 * 用于路由守卫中判断滑动方向
 */
export const TAB_INDEX_MAP: Record<string, number> = {
  '/home': 0,
  '/mine': 1
}
