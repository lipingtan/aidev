/**
 * 字段权限 API
 * 后端路由前缀：/api/v1/admin/field-objects, /api/v1/admin/field-permissions
 */
import request from '@/utils/request'

// ==================== 类型定义 ====================

/** 字段对象 */
export interface FieldObject {
  id: string
  object_code: string
  object_name: string
  source: string
}

/** 字段定义 */
export interface FieldDefinition {
  id: string
  object_code: string
  field_name: string
  description: string
  source: string
}

/** 字段权限配置项 */
export interface FieldPermission {
  id: string
  role_id: string
  object_code: string
  field_name: string
  access: string
}

// ==================== API 方法 ====================

/** 获取已注册的字段对象列表 */
export function listFieldObjects(): Promise<FieldObject[]> {
  return request.get('/api/v1/admin/field-objects')
}

/** 获取对象的字段定义列表 */
export function listFieldDefinitions(objectCode: string): Promise<FieldDefinition[]> {
  return request.get(`/api/v1/admin/field-objects/${objectCode}/fields`)
}

/** 修改字段描述 */
export function updateFieldDescription(objectCode: string, fieldName: string, description: string): Promise<void> {
  return request.put(`/api/v1/admin/field-objects/${objectCode}/fields/${fieldName}`, { description })
}

/** 查询角色字段权限配置 */
export function getFieldPermissions(roleId: string, objectCode: string): Promise<FieldPermission[]> {
  return request.get('/api/v1/admin/field-permissions', { params: { role_id: roleId, object_code: objectCode } })
}

/** 批量设置角色字段权限 */
export function setFieldPermissions(
  roleId: string,
  objectCode: string,
  items: { field_name: string; access: string }[]
): Promise<void> {
  return request.put('/api/v1/admin/field-permissions', { role_id: roleId, object_code: objectCode, items })
}

/** 删除字段权限配置 */
export function deleteFieldPermission(id: string): Promise<void> {
  return request.delete(`/api/v1/admin/field-permissions/${id}`)
}

/** 手动注册字段对象 */
export function manualRegisterObject(objectCode: string, objectName: string): Promise<void> {
  return request.post('/api/v1/admin/field-objects', { object_code: objectCode, object_name: objectName })
}

/** 手动注册字段 */
export function manualRegisterField(objectCode: string, fieldName: string, description: string): Promise<void> {
  return request.post(`/api/v1/admin/field-objects/${objectCode}/fields`, { field_name: fieldName, description })
}
