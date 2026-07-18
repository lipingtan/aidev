/**
 * 日志 API（对接 auth-rbac）
 * 后端路由：/api/v1/login-logs（登录日志）
 *          /api/v1/operation-logs（操作日志）
 */
import request from '@/utils/request'

export interface LoginLogQuery {
  username?: string
  ip?: string
  status?: number
  page: number
  page_size: number
}

export interface LoginLogItem {
  id: string
  user_id: string
  username: string
  ip: string
  location: string
  browser: string
  os: string
  login_time: string
  status: number
  message: string
}

export interface OperationLogQuery {
  module?: string
  action?: string
  user_id?: string
  target_type?: string
  page: number
  page_size: number
}

export interface OperationLogItem {
  id: string
  user_id: string
  tenant_id: string
  module: string
  action: string
  target_type: string
  target_id: string
  summary: string
  old_value?: any
  new_value?: any
  client_ip: string
  user_agent: string
  created_at: string
}

/** 登录日志查询 */
export function listLoginLogs(params: LoginLogQuery) {
  return request.get('/api/v1/login-logs', { params })
}

/** 删除登录日志 */
export function deleteLoginLog(id: string): Promise<void> {
  return request.delete(`/api/v1/login-logs/${id}`)
}

/** 操作日志查询 */
export function listOperationLogs(params: OperationLogQuery) {
  return request.get('/api/v1/operation-logs', { params })
}
