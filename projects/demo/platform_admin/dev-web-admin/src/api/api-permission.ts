/**
 * 接口权限管理 API
 * 对接后端 /api/v1/api-permissions 系列接口
 */
import request from '@/utils/request'

/** 接口权限节点类型 */
export type ApiPermissionNodeType = 'GROUP' | 'ENDPOINT'

/** 接口权限树节点 */
export interface ApiPermissionNode {
  id: string
  parent_id: string | null
  type: ApiPermissionNodeType
  name: string
  display_name?: string
  permission_code?: string
  /** ENDPOINT 专属字段 */
  http_method?: string
  url_pattern?: string
  app_code?: string
  status: string
  visible: number
  sort_order: number
  children?: ApiPermissionNode[]
}

/** 新增/编辑请求体 */
export interface ApiPermissionForm {
  parent_id: string | null
  type: ApiPermissionNodeType
  name: string
  permission_code?: string
  http_method?: string
  url_pattern?: string
  app_code: string
  sort_order?: number
}

/** 移动请求体 */
export interface MovePayload {
  target_parent_id: string
  sort_order?: number
}

/** 获取接口权限树 */
export function getApiPermissionTree(appCode: string): Promise<ApiPermissionNode[]> {
  return request.get('/api/v1/api-permissions/tree', { params: { app_code: appCode } })
}

/** 获取未分配 endpoint 列表 */
export function getUnassignedEndpoints(appCode: string): Promise<ApiPermissionNode[]> {
  return request.get('/api/v1/api-permissions/unassigned', { params: { app_code: appCode } })
}

/** 新增接口权限节点 */
export function createApiPermission(data: ApiPermissionForm): Promise<ApiPermissionNode> {
  return request.post('/api/v1/api-permissions', data)
}

/** 编辑接口权限节点 */
export function updateApiPermission(id: string, data: Partial<ApiPermissionForm>): Promise<ApiPermissionNode> {
  return request.put(`/api/v1/api-permissions/${id}`, data)
}

/** 删除接口权限节点 */
export function deleteApiPermission(id: string): Promise<void> {
  return request.delete(`/api/v1/api-permissions/${id}`)
}

/** 移动节点到目标 GROUP */
export function moveApiPermission(id: string, data: MovePayload): Promise<void> {
  return request.put(`/api/v1/api-permissions/${id}/move`, data)
}

/** 切换接口权限节点显示/隐藏 */
export function toggleApiPermissionVisible(id: string, visible: number): Promise<void> {
  return request.put(`/api/v1/api-permissions/${id}/visible`, { visible })
}
