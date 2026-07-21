/**
 * 组织架构 API
 * 后端路由：/api/v1/admin/org-units
 */
import request from '@/utils/request'

/** 组织节点 */
export interface OrgUnitItem {
  id: string
  tenant_id: string
  parent_id: string | null
  node_type: string
  name: string
  code: string
  sort_order: number
  status: number
  version: number
  created_at: string
  updated_at: string
  children?: OrgUnitItem[]
}

/** 创建组织节点参数 */
export interface OrgUnitCreateParams {
  parent_id?: string | null
  node_type: string
  name: string
  code?: string
  sort_order?: number
}

/** 更新组织节点参数 */
export interface OrgUnitUpdateParams {
  parent_id?: string | null
  node_type?: string
  name?: string
  code?: string
  sort_order?: number
  status?: number
  version: number
}

/** 用户-组织关联 */
export interface UserOrgItem {
  id: string
  user_id: string
  org_unit_id: string
  tenant_id: string
  is_primary: number
}

/** 设置节点用户参数 */
export interface SetNodeUsersParams {
  user_ids: string[]
  is_primary?: number
}

/** 获取组织架构树 */
export function getOrgTree(): Promise<OrgUnitItem[]> {
  return request.get('/api/v1/admin/org-units/tree')
}

/** 创建组织节点 */
export function createOrgUnit(data: OrgUnitCreateParams) {
  return request.post('/api/v1/admin/org-units', data)
}

/** 更新组织节点 */
export function updateOrgUnit(id: string, data: OrgUnitUpdateParams) {
  return request.put(`/api/v1/admin/org-units/${id}`, data)
}

/** 删除组织节点 */
export function deleteOrgUnit(id: string) {
  return request.delete(`/api/v1/admin/org-units/${id}`)
}

/** 获取节点下用户 */
export function getNodeUsers(id: string): Promise<UserOrgItem[]> {
  return request.get(`/api/v1/admin/org-units/${id}/users`)
}

/** 设置节点用户 */
export function setNodeUsers(id: string, data: SetNodeUsersParams) {
  return request.put(`/api/v1/admin/org-units/${id}/users`, data)
}
