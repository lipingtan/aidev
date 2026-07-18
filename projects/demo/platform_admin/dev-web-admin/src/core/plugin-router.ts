import type { Router } from 'vue-router'
import type { PluginConfig } from '@/plugin-sdk'

/**
 * 将已加载插件的路由动态注册到 Vue Router
 * 所有插件路由添加到 AppLayout 子路由下
 */
export function mergePluginRoutes(router: Router, plugins: PluginConfig[]) {
  for (const plugin of plugins) {
    if (!plugin.routes || plugin.routes.length === 0) continue
    for (const route of plugin.routes) {
      // 检查路由是否已存在（避免重复注册）
      if (router.hasRoute(route.name)) continue
      // 添加到根 layout 路由的 children 中
      router.addRoute({
        path: '/',
        component: () => import('@/layout/AppLayout.vue'),
        children: [{
          path: route.path,
          name: route.name,
          component: route.component,
          meta: {
            ...route.meta,
            pluginName: plugin.manifest.name
          }
        }]
      })
    }
  }
}

/**
 * 移除指定插件注册的所有路由
 */
export function removePluginRoutes(router: Router, pluginName: string) {
  const routes = router.getRoutes()
  for (const route of routes) {
    if (route.meta?.pluginName === pluginName && route.name) {
      router.removeRoute(route.name)
    }
  }
}
