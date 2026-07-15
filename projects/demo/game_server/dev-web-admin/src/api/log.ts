/**
 * 日志 API（对接 go-admin）
 * 后端路由：/api/v1/sys-login-log（登录日志）
 *          /api/v1/sys-opera-log（操作日志）
 */
import request from '@/utils/request'

export interface LoginLogQuery {
  username?: string
  ipaddr?: string
  loginLocation?: string
  status?: string
  pageNum: number
  pageSize: number
}

export interface LoginLogItem {
  infoId: number
  username: string
  ipaddr: string
  loginLocation: string
  browser: string
  os: string
  loginTime: string
  status: string
  msg: string
}

export interface OperationLogQuery {
  title?: string
  operName?: string
  businessType?: string
  status?: string
  pageNum: number
  pageSize: number
}

export interface OperationLogItem {
  operId: number
  title: string
  businessType: string
  method: string
  requestMethod: string
  operName: string
  operUrl: string
  operIp: string
  operLocation: string
  operParam: string
  jsonResult: string
  status: string
  operTime: string
  latencyTime: string
}

/** 登录日志查询 */
export function listLoginLogs(params: LoginLogQuery) {
  return request.get('/sys-login-log', { params })
}

/** 登录日志详情 */
export function getLoginLog(id: number) {
  return request.get(`/sys-login-log/${id}`)
}

/** 删除登录日志 */
export function deleteLoginLogs(ids: number[]): Promise<void> {
  return request.delete('/sys-login-log', { data: { ids } })
}

/** 操作日志查询 */
export function listOperationLogs(params: OperationLogQuery) {
  return request.get('/sys-opera-log', { params })
}

/** 操作日志详情 */
export function getOperationLog(id: number) {
  return request.get(`/sys-opera-log/${id}`)
}

/** 删除操作日志 */
export function deleteOperationLogs(ids: number[]): Promise<void> {
  return request.delete('/sys-opera-log', { data: { ids } })
}
