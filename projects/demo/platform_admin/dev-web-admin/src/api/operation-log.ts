/**
 * 操作日志 API
 * 后端路由：GET /api/v1/operation-logs
 */
import request from '@/utils/request'

/** 操作日志查询参数 */
export interface OperationLogQuery {
  module?: string
  action?: string
  user_id?: string
  start_time?: string
  end_time?: string
  page: number
  page_size: number
}

/** 操作日志列表项 */
export interface OperationLogItem {
  id: string
  module: string
  action: string
  target_type?: string
  target_id?: string
  user_id: string
  summary?: string
  old_value?: Record<string, any> | null
  new_value?: Record<string, any> | null
  client_ip?: string
  user_agent?: string
  created_at: string
}

/** 操作日志分页响应 */
export interface OperationLogPageResult {
  code: number
  message: string
  data: OperationLogItem[]
  total: number
}

/** 获取操作日志列表（分页） */
export function listOperationLogs(params: OperationLogQuery): Promise<OperationLogPageResult> {
  return request.get('/api/v1/operation-logs', { params })
}
