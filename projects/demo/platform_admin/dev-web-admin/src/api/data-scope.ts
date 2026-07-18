/**
 * 数据权限维度配置 API
 * 后端路由：/api/v1/data-scope-configs（CRUD）
 */
import request from '@/utils/request'

/** 数据权限维度配置项 */
export interface DataScopeConfigItem {
  id: string
  dimension_name: string
  display_name: string
  value_source: string
  description?: string
  status: number
  created_at: string
  updated_at: string
}

/** 数据权限维度创建/编辑参数 */
export interface DataScopeConfigFormData {
  dimension_name: string
  display_name: string
  value_source: string
  description?: string
  status?: number
}

/** 获取数据权限维度列表 */
export function listDataScopeConfigs() {
  return request.get<any, DataScopeConfigItem[]>('/api/v1/data-scope-configs')
}

/** 新增数据权限维度 */
export function createDataScopeConfig(data: DataScopeConfigFormData): Promise<void> {
  return request.post('/api/v1/data-scope-configs', data)
}

/** 编辑数据权限维度 */
export function updateDataScopeConfig(id: string, data: DataScopeConfigFormData): Promise<void> {
  return request.put(`/api/v1/data-scope-configs/${id}`, data)
}

/** 删除数据权限维度 */
export function deleteDataScopeConfig(id: string): Promise<void> {
  return request.delete(`/api/v1/data-scope-configs/${id}`)
}
