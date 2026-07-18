/**
 * API 公共适配工具
 * 适配 go-admin 后端 PageOK 响应格式
 */

/** 分页查询参数 */
export interface PageQuery {
  pageNum: number
  pageSize: number
  [key: string]: any
}

/** 适配后的分页响应 */
export interface PageResult<T> {
  list: T[]
  total: number
}

/**
 * 适配 go-admin PageOK 响应格式
 * 后端返回: { code: 200, data: { list: T[], count: number }, message: '' }
 * 经 axios 拦截器解包后为: { list: T[], count: number }
 * 适配为: { list: T[], total: number }
 */
export function adaptPageResponse<T>(res: any): PageResult<T> {
  if (res && Array.isArray(res.list)) {
    return { list: res.list, total: res.count || res.list.length }
  }
  if (Array.isArray(res)) {
    return { list: res, total: res.length }
  }
  return { list: [], total: 0 }
}
