/**
 * C端用户（biz_user）管理 API
 * 后端路由前缀：/api/v1/admin/biz-users
 */
import request from '@/utils/request'

// ==================== 类型定义 ====================

/** C端用户实体 */
export interface BizUser {
  id: string
  tenant_id: string
  phone: string
  nickname: string
  avatar: string
  status: number
  token_version: number
  last_login_at: string | null
  last_login_ip: string
  created_at: string
  updated_at: string
  version: number
}

/** C端用户查询参数 */
export interface BizUserQuery {
  page: number
  page_size: number
  phone?: string
}

/** C端用户新增参数 */
export interface BizUserCreateParams {
  phone: string
  nickname?: string
  password?: string
}

/** C端用户编辑参数 */
export interface BizUserUpdateParams {
  phone?: string
  nickname?: string
  avatar?: string
  version: number
}

/** C端用户分页响应 */
export interface BizUserPageResult {
  list: BizUser[]
  total: number
  page: number
  page_size: number
}

/** 重置密码响应 */
export interface ResetPasswordResult {
  password: string
}

// ==================== API 方法 ====================

/** C端用户分页查询 */
export function getBizUsers(params: BizUserQuery) {
  return request.get<any, BizUserPageResult>('/api/v1/admin/biz-users', { params })
}

/** 获取C端用户详情 */
export function getBizUserById(id: string) {
  return request.get<any, BizUser>(`/api/v1/admin/biz-users/${id}`)
}

/** 新增C端用户 */
export function createBizUser(data: BizUserCreateParams): Promise<void> {
  return request.post('/api/v1/admin/biz-users', data)
}

/** 编辑C端用户 */
export function updateBizUser(id: string, data: BizUserUpdateParams): Promise<void> {
  return request.put(`/api/v1/admin/biz-users/${id}`, data)
}

/** 删除C端用户 */
export function deleteBizUser(id: string): Promise<void> {
  return request.delete(`/api/v1/admin/biz-users/${id}`)
}

/** 重置密码（返回明文密码） */
export function resetPassword(id: string): Promise<ResetPasswordResult> {
  return request.post(`/api/v1/admin/biz-users/${id}/reset-password`)
}

/** 强制登出 */
export function forceLogout(id: string): Promise<void> {
  return request.post(`/api/v1/admin/biz-users/${id}/force-logout`)
}

/** 切换启用/禁用状态 */
export function toggleStatus(id: string, status: number): Promise<void> {
  return request.put(`/api/v1/admin/biz-users/${id}/toggle-status`, { status })
}
