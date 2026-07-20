/**
 * 应用管理 API
 * 后端路由：/api/v1/applications（CRUD）+ /api/v1/tenants/:id/apps（租户订阅）
 */
import request from '@/utils/request'

/** 应用列表项 */
export interface ApplicationItem {
  id: string
  name: string
  app_code: string
  description?: string
  status: number
  version: number
  created_at: string
  updated_at: string
}

/** 应用创建/编辑参数 */
export interface ApplicationFormData {
  name: string
  app_code: string
  description?: string
  status?: number
}

/** 获取应用列表 */
export function listApplications() {
  return request.get<any, ApplicationItem[]>('/api/v1/admin/applications')
}

/** 新增应用 */
export function createApplication(data: ApplicationFormData): Promise<void> {
  return request.post('/api/v1/admin/applications', data)
}

/** 编辑应用 */
export function updateApplication(id: string, data: ApplicationFormData): Promise<void> {
  return request.put(`/api/v1/admin/applications/${id}`, data)
}

/** 删除应用 */
export function deleteApplication(id: string): Promise<void> {
  return request.delete(`/api/v1/admin/applications/${id}`)
}

/** 获取租户已订阅的应用 ID 列表 */
export function getTenantApps(tenantId: string): Promise<ApplicationItem[]> {
  return request.get(`/api/v1/admin/tenants/${tenantId}/apps`)
}

/** 更新租户订阅的应用列表（按 app_code） */
export function updateTenantApps(tenantId: string, appCodes: string[]): Promise<void> {
  return request.put(`/api/v1/admin/tenants/${tenantId}/apps`, { app_codes: appCodes })
}
