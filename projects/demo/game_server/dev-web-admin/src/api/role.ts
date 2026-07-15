/**
 * 角色管理 API（对接 go-admin）
 * 后端路由：/api/v1/role（CRUD）
 *          /api/v1/role-status（状态切换）
 *          /api/v1/roledatascope（数据权限配置）
 *          /api/v1/roleMenuTreeselect/:roleId（角色菜单树）
 */
import request from '@/utils/request'

export interface RoleQuery {
  roleName?: string
  roleKey?: string
  /** 兼容视图使用 roleCode */
  roleCode?: string
  status?: string | number
  pageNum: number
  pageSize: number
}

export interface RolePageItem {
  roleId: number
  /** 兼容视图层使用 row.id */
  id: number
  roleName: string
  roleKey: string
  roleSort: number
  status: string
  dataScope: string
  /** 兼容视图层使用 dataPermissionLevel */
  dataPermissionLevel: string
  remark: string
  createTime: string
}

export interface RoleSimple {
  roleId: number
  /** 兼容视图使用 r.id */
  id: number
  roleName: string
  roleKey: string
}

export interface RoleCreateParams {
  roleName: string
  roleKey?: string
  /** 兼容视图使用 roleCode */
  roleCode?: string
  roleSort?: number
  /** 兼容视图使用 sort */
  sort?: number
  status?: string
  remark?: string
  menuIds?: number[]
  /** 兼容视图使用 dataPermissionLevel */
  dataPermissionLevel?: string
}

export interface RoleUpdateParams {
  roleId?: number
  roleName?: string
  roleKey?: string
  roleSort?: number
  /** 兼容视图使用 sort */
  sort?: number
  status?: string
  remark?: string
  menuIds?: number[]
  /** 兼容视图使用 dataPermissionLevel */
  dataPermissionLevel?: string
}

export interface DataScopeConfig {
  roleId: number
  dataScope: string
  deptIds?: number[]
}

/** 角色分页查询 */
export function listRoles(params: RoleQuery) {
  return request.get('/role', { params })
}

/** 获取角色详情 */
export function getRole(id: number) {
  return request.get(`/role/${id}`)
}

/** 新增角色 */
export function createRole(data: RoleCreateParams): Promise<void> {
  return request.post('/role', data)
}

/** 编辑角色 */
export function updateRole(id: number, data: RoleUpdateParams): Promise<void> {
  return request.put(`/role/${id}`, { ...data, roleId: id })
}

/** 删除角色 */
export function deleteRole(id: number): Promise<void> {
  return request.delete('/role', { data: { ids: [id] } })
}

/** 角色状态切换 */
export function updateRoleStatus(data: { roleId: number; status: string }): Promise<void> {
  return request.put('/role-status', data)
}

/** 查询角色菜单树（含已选中节点） */
export function getRoleMenuTreeSelect(roleId: number) {
  return request.get(`/roleMenuTreeselect/${roleId}`)
}

/** 配置数据权限 */
export function configDataScope(data: DataScopeConfig): Promise<void> {
  return request.put('/roledatascope', data)
}

// ===== 兼容旧视图引用 =====

/** 角色全量列表（兼容 UserForm 引用） */
export function listAllRoles(): Promise<RoleSimple[]> {
  return request.get('/role')
}

/** 查询角色菜单 ID 列表（兼容 MenuAssign 引用） */
export async function getRoleMenuIds(roleId: number): Promise<number[]> {
  const res: any = await request.get(`/roleMenuTreeselect/${roleId}`)
  return res?.checkedKeys || res?.menuIds || []
}

/** 分配菜单权限（兼容 MenuAssign 引用） */
export function assignMenus(roleId: number, menuIds: number[]): Promise<void> {
  return request.put(`/role/${roleId}`, { roleId, menuIds })
}

/** 配置数据权限（兼容 DataPermissionConfig 引用） */
export function configDataPermission(roleId: number, data: { dataPermissionLevel: string; dataIds: number[] }): Promise<void> {
  return request.put('/roledatascope', { roleId, dataScope: data.dataPermissionLevel, deptIds: data.dataIds })
}
