/**
 * 用户管理 API
 * 后端路由前缀：/api/v1/users
 */
import request from '@/utils/request'

// ==================== 类型定义 ====================

/** 用户查询参数 */
export interface UserQuery {
  username?: string
  email?: string
  phone?: string
  status?: number
  page: number
  page_size: number
}

/** 用户列表项 */
export interface UserPageItem {
  id: string
  username: string
  email: string
  phone: string
  status: number
  version: number
  created_at: string
  updated_at: string
}

/** 用户新增参数 */
export interface UserCreateParams {
  username: string
  password: string
  email?: string
  phone?: string
}

/** 用户编辑参数 */
export interface UserUpdateParams {
  username?: string
  email?: string
  phone?: string
  status?: number
  version: number
}

/** 用户关联的租户 */
export interface UserTenant {
  id: string
  name: string
}

/** 用户关联租户参数 */
export interface UserTenantParams {
  tenant_ids: (string | number)[]
}

/** 用户角色分配参数 */
export interface UserRoleParams {
  tenant_id: string
  roles: { role_id: string }[]
}

/** 分页响应 */
export interface UserPageResult {
  list: UserPageItem[]
  total: number
  page: number
  page_size: number
}

// ==================== API 方法 ====================

/** 用户分页查询 */
export function listUsers(params: UserQuery) {
  return request.get<any, UserPageResult>('/api/v1/admin/users', { params })
}

/** 新增用户 */
export function createUser(data: UserCreateParams): Promise<void> {
  return request.post('/api/v1/admin/users', data)
}

/** 编辑用户 */
export function updateUser(id: string, data: UserUpdateParams): Promise<void> {
  return request.put(`/api/v1/admin/users/${id}`, data)
}

/** 删除用户 */
export function deleteUser(id: string): Promise<void> {
  return request.delete(`/api/v1/admin/users/${id}`)
}

/** 获取用户关联的租户列表 */
export function getUserTenants(userId: string): Promise<UserTenant[]> {
  return request.get(`/api/v1/admin/users/${userId}/tenants`)
}

/** 关联用户到租户 */
export function addUserTenants(userId: string, data: UserTenantParams): Promise<void> {
  return request.post(`/api/v1/admin/users/${userId}/tenants`, data)
}

/** 移除用户与租户的关联 */
export function removeUserTenant(userId: string, tenantId: string): Promise<void> {
  return request.delete(`/api/v1/admin/users/${userId}/tenants/${tenantId}`)
}

/** 分配用户角色（当前租户下） */
export function assignUserRoles(userId: string, data: UserRoleParams): Promise<void> {
  return request.post(`/api/v1/admin/users/${userId}/roles`, data)
}

/** 强制下线 */
export function forceOfflineUser(userId: string): Promise<void> {
  return request.post(`/api/v1/admin/users/${userId}/force-offline`)
}
