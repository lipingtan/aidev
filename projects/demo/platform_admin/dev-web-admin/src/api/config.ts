/**
 * 系统配置 API（对接 auth-rbac）
 * 后端路由：/api/v1/configs
 */
import request from '@/utils/request'

export interface ConfigItem {
  id: string
  config_name: string
  config_key: string
  config_value: string
  config_type: number
  remark: string
  status: number
  created_at: string
}

export interface ConfigCreateParams {
  config_name: string
  config_key: string
  config_value: string
  config_type?: number
  remark?: string
}

export interface ConfigQuery {
  config_name?: string
  config_key?: string
  page: number
  page_size: number
}

export interface ConfigPageResult {
  list: ConfigItem[]
  total: number
}

/** 获取配置分页列表 */
export async function listConfigs(params: ConfigQuery): Promise<ConfigPageResult> {
  const res: any = await request.get('/api/v1/admin/configs', { params })
  if (res && Array.isArray(res.list)) {
    return { list: res.list, total: res.count || res.total || res.list.length }
  }
  if (Array.isArray(res)) {
    return { list: res, total: res.length }
  }
  return { list: [], total: 0 }
}

/** 获取配置详情 */
export function getConfig(id: string) {
  return request.get(`/api/v1/admin/configs/${id}`)
}

/** 按 configKey 查询配置值 */
export function getConfigByKey(configKey: string) {
  return request.get(`/api/v1/admin/configs/key/${configKey}`)
}

/** 创建配置 */
export function createConfig(data: ConfigCreateParams) {
  return request.post('/api/v1/admin/configs', data)
}

/** 更新配置 */
export function updateConfig(id: string, data: Partial<ConfigCreateParams>) {
  return request.put(`/api/v1/admin/configs/${id}`, data)
}

/** 删除配置 */
export function deleteConfig(id: string) {
  return request.delete(`/api/v1/admin/configs/${id}`)
}
