/**
 * 插件管理 API（对接新后端 /api/v1/admin/plugins）
 */
import request from '@/utils/request'

export interface PluginItem {
  id: number
  name: string
  version: string
  description: string
  status: number           // 0=已安装 1=运行中 2=已停止
  binaryPath: string
  frontendPath: string
  runStatus: 'running' | 'stopped' | 'error' | 'unknown'
}

export interface PluginListResponse {
  list: PluginItem[]
  total: number
}

export interface UpgradeResponse {
  needConfirm?: boolean
  migrationNotes?: string
  oldVersion?: string
  newVersion?: string
  message?: string
}

/** 获取插件列表 */
export function getPluginList() {
  return request.get<PluginListResponse>('/api/v1/admin/plugins')
}

/** 上传安装插件 */
export function uploadPlugin(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/api/v1/admin/plugins/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

/** 启动插件 */
export function startPlugin(name: string) {
  return request.post(`/api/v1/admin/plugins/${name}/start`)
}

/** 停止插件 */
export function stopPlugin(name: string) {
  return request.post(`/api/v1/admin/plugins/${name}/stop`)
}

/** 卸载插件 */
export function uninstallPlugin(name: string) {
  return request.delete(`/api/v1/admin/plugins/${name}`)
}

/** 升级插件 */
export function upgradePlugin(name: string, file: File, confirm = false) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('confirm', String(confirm))
  return request.put<UpgradeResponse>(`/api/v1/admin/plugins/${name}/upgrade`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

/** 健康检查 */
export function getPluginHealth(name: string) {
  return request.get(`/api/v1/admin/plugins/${name}/health`)
}

// ========== 兼容旧 API ==========

export interface EnabledPluginInfo {
  name: string
  version: string
  displayName: string
  frontendPath: string
  mode: 'source' | 'runtime'
}

/** 获取当前运行中的插件列表（供插件容器使用） */
export async function listEnabledPlugins(): Promise<EnabledPluginInfo[]> {
  const res: any = await request.get('/api/v1/admin/plugins')
  const data = res?.data || res
  const all = data?.list || []
  return all
    .filter((p: any) => p.runStatus === 'running' && p.frontendPath)
    .map((p: any) => ({
      name: p.name,
      version: p.version || '',
      displayName: p.description || p.name,
      frontendPath: p.frontendPath || '',
      mode: 'runtime' as const
    }))
}

// ========== 旧版 API 兼容别名 ==========

/** @deprecated 使用 getPluginList 替代 */
export const listAllPlugins = getPluginList

/** @deprecated 使用 startPlugin 替代 */
export const enablePlugin = startPlugin

/** @deprecated 使用 stopPlugin 替代 */
export const disablePlugin = stopPlugin

/** @deprecated 使用 uploadPlugin 替代 */
export function installPluginByUrl(data: { name: string; url: string }) {
  return request.post('/api/v1/admin/plugins/upload', data)
}

/** @deprecated 使用 uploadPlugin 替代 */
export function installPluginByFile(name: string, file: File) {
  return uploadPlugin(file)
}
