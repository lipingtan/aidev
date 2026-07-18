/**
 * 角色管理 API
 * 对接后端 /api/v1/roles 系列接口
 */
import request from '@/utils/request'

// ==================== 类型定义 ====================

/** 角色节点（树结构） */
export interface RoleItem {
  id: string
  role_name: string
  role_code: string
  parent_id: string | null
  sort_order: number
  status: number
  role_type: string
  version: number
  created_at: string
  children?: RoleItem[]
}

/** 角色创建参数 */
export interface RoleCreateParams {
  role_name: string
  role_code: string
  role_type?: string
  parent_id?: string
  sort_order?: number
  status?: number
}

/** 角色更新参数 */
export interface RoleUpdateParams {
  role_name?: string
  role_type?: string
  parent_id?: string
  sort_order?: number
  status?: number
  version: number
}

/** 资源树节点 */
export interface ResourceTreeNode {
  id: string
  name: string
  type?: string
  path?: string
  icon?: string
  children?: ResourceTreeNode[]
}

/** 接口权限树节点 */
export interface ApiPermTreeNode {
  id: string
  name: string
  display_name?: string
  type?: string
  http_method?: string
  url_pattern?: string
  visible?: number
  children?: ApiPermTreeNode[]
}

/** 数据权限维度配置 */
export interface DataScopeConfig {
  dimension: string
  dimension_label: string
  values: string[]
}

/** 数据权限维度选项（后端返回的可选维度） */
export interface DataScopeDimension {
  id: string
  dimension_name: string
  display_name: string
  value_source?: string
  status?: number
  options?: { value: string; label: string }[]
}

/** 应用信息 */
export interface AppItem {
  id: string
  name: string
  app_code: string
  description?: string
}

// ==================== 角色 CRUD ====================

/** 获取角色列表（树形） */
export function getRoleList(): Promise<RoleItem[]> {
  return request.get('/api/v1/roles')
}

/** 创建角色 */
export function createRole(data: RoleCreateParams): Promise<void> {
  return request.post('/api/v1/roles', data)
}

/** 更新角色 */
export function updateRole(id: string, data: RoleUpdateParams): Promise<void> {
  return request.put(`/api/v1/roles/${id}`, data)
}

/** 删除角色 */
export function deleteRole(id: string): Promise<void> {
  return request.delete(`/api/v1/roles/${id}`)
}

// ==================== 权限分配 ====================

/** 获取资源树 */
export function getResourceTree(appCode: string): Promise<ResourceTreeNode[]> {
  return request.get('/api/v1/resources/tree', { params: { app_code: appCode } })
}

/** 获取角色已分配的资源 ID 列表 */
export function getRoleResources(roleId: string): Promise<string[]> {
  return request.get(`/api/v1/roles/${roleId}/resources`)
}

/** 分配资源（菜单）权限 */
export function assignRoleResources(roleId: string, resourceIds: string[]): Promise<void> {
  return request.put(`/api/v1/roles/${roleId}/resources`, { resource_ids: resourceIds })
}

/** 获取接口权限树 */
export function getApiPermTree(appCode: string): Promise<ApiPermTreeNode[]> {
  return request.get('/api/v1/api-permissions/tree', { params: { app_code: appCode } })
}

/** 获取角色已分配的接口权限 ID 列表 */
export function getRoleApis(roleId: string): Promise<string[]> {
  return request.get(`/api/v1/roles/${roleId}/apis`)
}

/** 分配接口权限 */
export function assignRoleApis(roleId: string, apiIds: string[]): Promise<void> {
  return request.put(`/api/v1/roles/${roleId}/apis`, { api_permission_ids: apiIds })
}

/** 获取数据权限维度配置选项 */
export function getDataScopeConfigs(): Promise<DataScopeDimension[]> {
  return request.get('/api/v1/data-scope-configs')
}

/** 获取角色已配置的数据权限 */
export function getRoleDataScopes(roleId: string): Promise<DataScopeConfig[]> {
  return request.get(`/api/v1/roles/${roleId}/data-scopes`)
}

/** 配置数据权限 */
export function assignRoleDataScopes(roleId: string, scopes: DataScopeConfig[]): Promise<void> {
  return request.put(`/api/v1/roles/${roleId}/data-scopes`, { scopes })
}

/** 获取租户下应用列表 */
export function getTenantApps(tenantId: string): Promise<AppItem[]> {
  return request.get(`/api/v1/tenants/${tenantId}/apps`)
}

/** 获取角色已绑定的应用编码列表 */
export function getRoleApps(roleId: string): Promise<string[]> {
  return request.get(`/api/v1/roles/${roleId}/apps`)
}

/** 绑定应用（按 app_code） */
export function assignRoleApps(roleId: string, appCodes: string[]): Promise<void> {
  return request.put(`/api/v1/roles/${roleId}/apps`, { app_codes: appCodes })
}

// ==================== 权限汇总 ====================

/** 应用权限统计摘要 */
export interface AppPermissionSummary {
  app_code: string
  app_name: string
  menu_count: number
  api_count: number
  data_scope_count: number
  bound: boolean
}

/** 获取角色权限汇总（每应用统计） */
export function getRolePermissionSummary(roleId: string): Promise<AppPermissionSummary[]> {
  return request.get(`/api/v1/roles/${roleId}/permission-summary`)
}

// ==================== 兼容旧引用 ====================

/** 角色全量列表（兼容 UserForm 等组件引用） */
export async function listAllRoles(): Promise<{ id: string; roleName: string; roleKey: string }[]> {
  const list = await getRoleList()
  const result: { id: string; roleName: string; roleKey: string }[] = []
  function flatten(nodes: RoleItem[]) {
    for (const n of nodes) {
      result.push({ id: n.id, roleName: n.role_name, roleKey: n.role_code })
      if (n.children) flatten(n.children)
    }
  }
  flatten(list)
  return result
}
