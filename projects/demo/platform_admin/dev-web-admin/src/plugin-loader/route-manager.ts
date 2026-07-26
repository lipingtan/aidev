import type { Router, RouteRecordRaw } from 'vue-router'
import type { PluginRoute } from '../plugin-sdk/types'

// 保存每个插件注册的路由名称
const pluginRouteNames = new Map<string, string[]>()

let _router: Router | null = null

/**
 * 初始化 route-manager，注入 Vue Router 实例
 * 必须在 main.ts/main-h5.ts 的 app.use(router) 之后调用
 */
export function initRouteManager(router: Router): void {
  _router = router
}

/**
 * 注册插件路由
 */
export function registerPluginRoutes(pluginName: string, routes: PluginRoute[]): void {
  if (!_router) {
    console.warn('[route-manager] router 未初始化，跳过注册', pluginName)
    return
  }
  const names: string[] = []
  for (const route of routes) {
    const record: RouteRecordRaw = {
      path: route.path,
      name: route.name,
      component: route.component,
      meta: route.meta
    }
    _router.addRoute(record)
    names.push(route.name)
  }
  pluginRouteNames.set(pluginName, names)
}

/**
 * 移除插件注册的所有路由
 */
export function unregisterPluginRoutes(pluginName: string): void {
  if (!_router) return
  const names = pluginRouteNames.get(pluginName) || []
  for (const name of names) {
    _router.removeRoute(name)
  }
  pluginRouteNames.delete(pluginName)
}
