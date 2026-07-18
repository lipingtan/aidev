/**
 * 插件管理 API（对接 go-admin）
 * 后端路由：/api/v1/plugins
 */
import request from '@/utils/request'

/** 已启用插件信息（后端返回） */
export interface EnabledPluginInfo {
  name: string
  version: string
  displayName: string
  frontendPath: string
  mode: 'source' | 'runtime'
}

/** 获取当前租户已启用的插件列表（从全部列表中过滤 status=1） */
export async function listEnabledPlugins(): Promise<EnabledPluginInfo[]> {
  const res: any = await request.get('/api/v1/plugins')
  const all = Array.isArray(res) ? res : (res?.data || res?.list || [])
  return all
    .filter((p: any) => p.status === 1 && p.frontendPath)
    .map((p: any) => ({
      name: p.name,
      version: p.version || '',
      displayName: p.description || p.name,
      frontendPath: p.frontendPath || p.frontend_path || '',
      mode: 'runtime' as const
    }))
}

/** 获取所有插件列表（管理用） */
export async function listAllPlugins() {
  return request.get('/api/v1/plugins')
}

/** 启用插件 */
export function enablePlugin(name: string) {
  return request.put(`/api/v1/plugins/${name}/start`)
}

/** 禁用插件 */
export function disablePlugin(name: string) {
  return request.put(`/api/v1/plugins/${name}/stop`)
}

/** 卸载插件 */
export function uninstallPlugin(name: string) {
  return request.delete(`/api/v1/plugins/${name}`)
}

/** 安装插件（URL 方式） */
export function installPluginByUrl(data: { name: string; url: string }) {
  return request.post('/api/v1/plugins/install', data)
}

/** 安装插件（文件上传方式） */
export function installPluginByFile(name: string, file: File) {
  const formData = new FormData()
  formData.append('name', name)
  formData.append('file', file)
  return request.post('/api/v1/plugins/install', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

/** 健康检查 */
export function healthCheckPlugin(name: string) {
  return request.get(`/api/v1/plugins/${name}/health`)
}
