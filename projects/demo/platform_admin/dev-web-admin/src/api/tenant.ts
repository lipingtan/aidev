/**
 * 租户管理 API
 * 后端路由：/api/v1/tenants（CRUD + 状态切换）
 */
import request from '@/utils/request'

/** 租户查询参数 */
export interface TenantQuery {
  name?: string
  status?: number
  page: number
  page_size: number
}

/** 租户列表项 */
export interface TenantItem {
  id: string
  name: string
  tenant_code: string
  status: number
  config?: Record<string, any>
  version: number
  created_at: string
  updated_at: string
}

/** 租户创建/编辑参数 */
export interface TenantFormData {
  name: string
  tenant_code: string
  config?: Record<string, any>
}
/** 租户分页查询 */
export function listTenants(params: TenantQuery) {
  return request.get('/api/v1/admin/tenants', { params })
}

/** 获取租户详情 */
export function getTenant(id: string) {
  return request.get(`/api/v1/admin/tenants/${id}`)
}

/** 新增租户 */
export function createTenant(data: TenantFormData): Promise<void> {
  return request.post('/api/v1/admin/tenants', data)
}

/** 编辑租户 */
export function updateTenant(id: string, data: TenantFormData): Promise<void> {
  return request.put(`/api/v1/admin/tenants/${id}`, data)
}

/** 切换租户状态（启用/禁用） */
export function updateTenantStatus(id: string, status: number): Promise<void> {
  return request.put(`/api/v1/admin/tenants/${id}/status`, { status })
}

/** 删除租户 */
export function deleteTenant(id: string): Promise<void> {
  return request.delete(`/api/v1/admin/tenants/${id}`)
}
