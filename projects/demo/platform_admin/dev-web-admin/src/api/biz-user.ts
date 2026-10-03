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
  phone: string | null
  email: string | null
  nickname: string
  avatar: string
  status: number
  disabled_at: string | null
  disabled_by: string
  disabled_reason: string
  disabled_reason_user: string
  enabled_at: string | null
  enabled_by: string
  token_version: number
  last_login_at: string | null
  last_login_ip: string
  created_at: string
  updated_at: string
  version: number
  access_token_expire_minutes?: number | null
  refresh_token_expire_minutes?: number | null
}

/** C端用户查询参数 */
export interface BizUserQuery {
  page: number
  page_size: number
  phone?: string
  email?: string
}

/** C端用户新增参数 */
export interface BizUserCreateParams {
  phone?: string
  email?: string
  nickname?: string
  password?: string
  // 键名 = 后端 DTO json tag（snake_case），驼峰会被 gin 丢弃
  access_token_expire_minutes?: number | null
  refresh_token_expire_minutes?: number | null
}

/** C端用户编辑参数 */
export interface BizUserUpdateParams {
  phone?: string
  email?: string
  nickname?: string
  avatar?: string
  version: number
  // 键名 = 后端 DTO json tag（snake_case），驼峰会被 gin 丢弃
  access_token_expire_minutes?: number | null
  refresh_token_expire_minutes?: number | null
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

/** 重置密码（可指定密码，留空由服务端随机生成；返回明文密码） */
export function resetPassword(id: string, password = ''): Promise<ResetPasswordResult> {
  return request.post(`/api/v1/admin/biz-users/${id}/reset-password`, password ? { password } : {})
}

/** 强制登出 */
export function forceLogout(id: string): Promise<void> {
  return request.post(`/api/v1/admin/biz-users/${id}/force-logout`)
}

/** 切换启用/禁用状态（FR-21：禁用时需携带管理端原因与用户展示原因） */
export function toggleStatus(id: string, status: number, adminReason = '', userReason = ''): Promise<void> {
  return request.put(`/api/v1/admin/biz-users/${id}/toggle-status`, { status, admin_reason: adminReason, user_reason: userReason })
}
