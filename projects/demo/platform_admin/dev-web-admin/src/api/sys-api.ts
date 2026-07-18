/**
 * 接口管理 API（对接 auth-rbac）
 * 后端路由：/api/v1/sys-apis
 */
import request from '@/utils/request'
import { adaptPageResponse, type PageQuery, type PageResult } from './helpers'

export interface SysApiItem {
  id: number
  path: string
  method: string
  title: string
  group: string
}

/** 获取接口分页列表 */
export async function listSysApis(params: PageQuery): Promise<PageResult<SysApiItem>> {
  const res = await request.get('/api/v1/sys-apis', { params })
  return adaptPageResponse<SysApiItem>(res)
}

/** 获取接口详情 */
export function getSysApi(id: number) {
  return request.get(`/api/v1/sys-apis/${id}`)
}

/** 更新接口 */
export function updateSysApi(id: number, data: Partial<SysApiItem>) {
  return request.put(`/api/v1/sys-apis/${id}`, data)
}
