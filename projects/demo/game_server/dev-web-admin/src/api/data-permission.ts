/**
 * 数据权限 API（对接 go-admin）
 * 后端路由：/api/v1/roledatascope（数据范围配置）
 *          /api/v1/roleDeptTreeselect/:roleId（角色部门树）
 *          /api/v1/deptTree（部门树）
 */
import request from '@/utils/request'

export interface OptionItem {
  id: number
  name: string
}

export interface DataPermissionOptions {
  areas: OptionItem[]
  communities: OptionItem[]
  buildings: OptionItem[]
}

/**
 * 获取数据权限可选范围
 * 注意：go-admin 标准版无此接口，使用部门树模拟
 * 如后端未实现则返回空数据
 */
export async function getDataPermissionOptions(): Promise<DataPermissionOptions> {
  try {
    const res: any = await request.get('/deptTree')
    // 将部门树适配为选项格式
    const items: OptionItem[] = Array.isArray(res) ? res.map((d: any) => ({ id: d.id || d.deptId, name: d.label || d.deptName })) : []
    return { areas: items, communities: [], buildings: [] }
  } catch {
    return { areas: [], communities: [], buildings: [] }
  }
}

/** 获取角色部门树（含已选中节点） */
export function getRoleDeptTreeSelect(roleId: number) {
  return request.get(`/roleDeptTreeselect/${roleId}`)
}

/** 获取部门树 */
export function getDeptTree() {
  return request.get('/deptTree')
}
