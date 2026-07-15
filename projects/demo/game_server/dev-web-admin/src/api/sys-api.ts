/**
 * 接口管理 API（对接 go-admin）
 * 后端路由：/api/v1/sys-api
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
  const res = await request.get('/sys-api', { params })
  return adaptPageResponse<SysApiItem>(res)
}

/** 获取接口详情 */
export function getSysApi(id: number) {
  return request.get(`/sys-api/${id}`)
}

/** 更新接口 */
export function updateSysApi(id: number, data: Partial<SysApiItem>) {
  return request.put(`/sys-api/${id}`, data)
}
