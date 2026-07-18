/**
 * 资源/菜单管理 API
 * 对接后端 /api/v1/resources 接口
 */
import request from '@/utils/request'

/** 资源类型 */
export type ResourceType = 'MENU' | 'BUTTON'

/** 资源树节点 */
export interface ResourceTreeNode {
  id: string
  parent_id: string | null
  name: string
  type: ResourceType
  path: string
  component: string
  icon: string
  permission_code: string
  app_code: string
  sort_order: number
  status: number
  children?: ResourceTreeNode[]
}

/** 创建/编辑资源参数 */
export interface ResourceForm {
  parent_id?: string | null
  name: string
  type: ResourceType
  path?: string
  component?: string
  icon?: string
  permission_code?: string
  app_code: string
  sort_order?: number
}

/** 排序项 */
export interface ResourceSortItem {
  id: string
  parent_id: string | null
  sort_order: number
}

/** 获取资源树 */
export function getResourceTree(params: { app_code: string }): Promise<ResourceTreeNode[]> {
  return request.get('/api/v1/resources/tree', { params })
}

/** 创建资源 */
export function createResource(data: ResourceForm): Promise<void> {
  return request.post('/api/v1/resources', data)
}

/** 编辑资源 */
export function updateResource(id: string, data: ResourceForm): Promise<void> {
  return request.put(`/api/v1/resources/${id}`, data)
}

/** 删除资源 */
export function deleteResource(id: string): Promise<void> {
  return request.delete(`/api/v1/resources/${id}`)
}

/** 拖拽排序 */
export function sortResources(data: ResourceSortItem[]): Promise<void> {
  return request.put('/api/v1/resources/sort', data)
}
