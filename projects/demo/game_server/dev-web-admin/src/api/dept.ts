/**
 * 部门管理 API（对接 go-admin）
 * 后端路由：/api/v1/dept（CRUD）
 *          /api/v1/deptTree（部门树）
 */
import request from '@/utils/request'

export interface DeptItem {
  deptId: number
  deptName: string
  parentId: number
  sort: number
  leader: string
  status: number
  children?: DeptItem[]
  createTime: string
}

export interface DeptCreateParams {
  deptName: string
  parentId: number
  sort: number
  leader?: string
  status?: number
}

/** 获取部门列表（树形） */
export function listDepts(params?: { deptName?: string }) {
  return request.get<DeptItem[]>('/dept', { params })
}

/** 获取部门详情 */
export function getDept(id: number) {
  return request.get<DeptItem>(`/dept/${id}`)
}

/** 创建部门 */
export function createDept(data: DeptCreateParams) {
  return request.post('/dept', data)
}

/** 更新部门 */
export function updateDept(id: number, data: Partial<DeptCreateParams>) {
  return request.put(`/dept/${id}`, data)
}

/** 删除部门 */
export function deleteDept(id: number) {
  return request.delete(`/dept/${id}`)
}

/** 获取部门树（用于下拉选择） */
export function getDeptTree() {
  return request.get<DeptItem[]>('/deptTree')
}
