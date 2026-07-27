import request from '@/utils/request'

export interface AppCatalogItem {
  id: number
  app_code: string
  name: string
  description: string
  app_type: 'BUILTIN' | 'PLUGIN' | 'EXTERNAL'
  icon: string
  platforms: string[]
  modules: Array<{ code: string; name: string }>
  subscribed: boolean
  plugin_status?: 'RUNNING' | 'STOPPED' | 'ERROR' | 'NOT_INSTALLED'
  subscription_mode?: 'direct' | 'approval_required'
  subscription_status?: 'active' | 'pending_approval' | 'rejected'
}

export interface SubscriptionItem {
  id: number
  tenant_id: number
  app_code: string
  enabled_modules: string[] | null
  app_name: string
  app_type: string
  description: string
  icon: string
}

/** 获取应用目录 */
export function getAppCatalog() {
  return request.get<AppCatalogItem[]>('/api/v1/admin/app-catalog')
}

/** 获取已订阅列表 */
export function getSubscriptions() {
  return request.get<SubscriptionItem[]>('/api/v1/admin/app-subscriptions')
}

/** 订阅应用 */
export function subscribeApp(appCode: string) {
  return request.post('/api/v1/admin/app-subscriptions', { app_code: appCode })
}

/** 退订应用 */
export function unsubscribeApp(appCode: string) {
  return request.delete(`/api/v1/admin/app-subscriptions/${appCode}`)
}

/** 更新启用模块 */
export function updateModules(appCode: string, enabledModules: string[]) {
  return request.put(`/api/v1/admin/app-subscriptions/${appCode}/modules`, { enabled_modules: enabledModules })
}
