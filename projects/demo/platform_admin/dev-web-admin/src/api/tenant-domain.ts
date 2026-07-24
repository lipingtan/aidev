/**
 * 域名-租户映射管理 API
 * 后端路由：/api/v1/admin/tenant-domains（CRUD）
 */
import request from '@/utils/request'

/** 域名映射查询参数 */
export interface TenantDomainQuery {
  domain?: string
  tenant_id?: string
  page: number
  page_size: number
}

/** 域名映射列表项 */
export interface TenantDomainItem {
  id: string
  domain: string
  match_type: string
  tenant_id: string
  tenant_code: string
  tenant_name: string
  remark: string
  version: number
  created_at: string
}

/** 创建域名映射参数 */
export interface CreateTenantDomainData {
  domain: string
  tenant_id: string
  remark?: string
}

/** 更新域名映射参数 */
export interface UpdateTenantDomainData {
  domain?: string
  tenant_id?: string
  remark?: string
  version: number
}

/** 分页查询域名映射列表 */
export function listTenantDomains(params: TenantDomainQuery) {
  return request.get('/api/v1/admin/tenant-domains', { params })
}

/** 创建域名映射 */
export function createTenantDomain(data: CreateTenantDomainData) {
  return request.post('/api/v1/admin/tenant-domains', data)
}

/** 更新域名映射 */
export function updateTenantDomain(id: string, data: UpdateTenantDomainData) {
  return request.put(`/api/v1/admin/tenant-domains/${id}`, data)
}

/** 删除域名映射 */
export function deleteTenantDomain(id: string) {
  return request.delete(`/api/v1/admin/tenant-domains/${id}`)
}
