/**
 * 三级配置 API（admin_config 表）
 * 后端路由：/api/v1/admin/configs
 */
import request from '@/utils/request'

/** 配置项 */
export interface AdminConfigItem {
  id: string
  config_key: string
  config_value: string
  config_type: string
  scope: string
  scope_id: string
  tenant_id: string
  display_name: string
  description: string
  is_feature_flag: number
  status: number
  created_at: string
  updated_at: string
}

/** 查询参数 */
export interface AdminConfigQuery {
  scope?: string
  key?: string
  is_feature_flag?: number
  page: number
  page_size: number
}

/** 创建参数 */
export interface AdminConfigCreateParams {
  config_key: string
  config_value: string
  config_type?: string
  scope: string
  scope_id: string
  tenant_id: string
  display_name?: string
  description?: string
  is_feature_flag?: number
}

/** 功能开关项 */
export interface FeatureFlagItem {
  key: string
  display_name: string
  enabled: boolean
}

/** 配置列表 */
export async function listAdminConfigs(params: AdminConfigQuery) {
  const res: any = await request.get('/api/v1/admin/configs', { params })
  return res?.data || res || { list: [], total: 0 }
}

/** 创建配置 */
export function createAdminConfig(data: AdminConfigCreateParams) {
  return request.post('/api/v1/admin/configs', data)
}

/** 更新配置 */
export function updateAdminConfig(id: string, data: Partial<AdminConfigCreateParams>) {
  return request.put(`/api/v1/admin/configs/${id}`, data)
}

/** 删除配置 */
export function deleteAdminConfig(id: string) {
  return request.delete(`/api/v1/admin/configs/${id}`)
}

/** 按 key 解析有效值（三级合并） */
export function resolveConfig(key: string) {
  return request.get(`/api/v1/admin/configs/resolve/${key}`)
}

/** 获取功能开关列表 */
export async function getFeatureFlags(): Promise<FeatureFlagItem[]> {
  const res: any = await request.get('/api/v1/admin/configs/feature-flags')
  return res?.data || res || []
}
