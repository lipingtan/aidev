/**
 * 记录共享 API
 * 后端路由前缀：/api/v1/admin/record-shares
 */
import request from '@/utils/request'

// ==================== 类型定义 ====================

/** 记录共享规则 */
export interface RecordShareItem {
  id: string
  tenant_id: string
  object_code: string
  record_id: string
  share_to_type: string
  share_to_id: string
  access_level: string
  expire_at: string | null
  created_by: string
  created_at: string
}

/** 创建共享规则参数 */
export interface RecordShareCreateParams {
  object_code: string
  record_id: string
  share_to_type: string
  share_to_id: string
  access_level: string
  expire_at?: string | null
}

// ==================== API 方法 ====================

/** 创建共享规则 */
export function createRecordShare(data: RecordShareCreateParams): Promise<RecordShareItem> {
  return request.post('/api/v1/admin/record-shares', data)
}

/** 查询记录的共享规则列表 */
export function listRecordShares(objectCode: string, recordId: string): Promise<RecordShareItem[]> {
  return request.get('/api/v1/admin/record-shares', { params: { object_code: objectCode, record_id: recordId } })
}

/** 删除共享规则 */
export function deleteRecordShare(id: string): Promise<void> {
  return request.delete(`/api/v1/admin/record-shares/${id}`)
}
