/**
 * 系统配置 API（对接 go-admin）
 * 后端路由：/api/v1/config（CRUD）
 *          /api/v1/configKey/:configKey（按 key 查询）
 *          /api/v1/set-config（全局设置）
 */
import request from '@/utils/request'
import { adaptPageResponse, type PageQuery, type PageResult } from './helpers'

export interface ConfigItem {
  configId: number
  configName: string
  configKey: string
  configValue: string
  configType: number
  remark: string
  createTime: string
}

export interface ConfigCreateParams {
  configName: string
  configKey: string
  configValue: string
  configType?: number
  remark?: string
}

/** 获取配置分页列表 */
export async function listConfigs(params: PageQuery): Promise<PageResult<ConfigItem>> {
  const res = await request.get('/config', { params })
  return adaptPageResponse<ConfigItem>(res)
}

/** 获取配置详情 */
export function getConfig(id: number) {
  return request.get(`/config/${id}`)
}

/** 按 configKey 查询配置值 */
export function getConfigByKey(configKey: string) {
  return request.get(`/configKey/${configKey}`)
}

/** 创建配置 */
export function createConfig(data: ConfigCreateParams) {
  return request.post('/config', data)
}

/** 更新配置 */
export function updateConfig(id: number, data: Partial<ConfigCreateParams>) {
  return request.put(`/config/${id}`, data)
}

/** 删除配置 */
export function deleteConfig(id: number) {
  return request.delete(`/config/${id}`)
}
