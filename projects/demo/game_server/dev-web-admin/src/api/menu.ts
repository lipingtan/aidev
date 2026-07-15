/**
 * 菜单管理 API（对接 go-admin）
 */
import request from '@/utils/request'

export interface MenuTreeItem {
  menuId: number
  menuName: string
  title: string
  parentId: number
  menuType: string
  path: string
  component: string
  permission: string
  icon: string
  sort: number
  visible: string
  children?: MenuTreeItem[]
}

/** 获取当前用户角色菜单（侧边栏用） */
export function getUserMenuTree(): Promise<MenuTreeItem[]> {
  return request.get('/menurole')
}

/** 菜单列表（管理用） */
export function getMenuList(params?: { menuName?: string; status?: number }): Promise<MenuTreeItem[]> {
  return request.get('/menu', { params })
}

/** 菜单列表（兼容旧引用） */
export const getMenuTree = getMenuList

/** 新增菜单 */
export function createMenu(data: any): Promise<void> {
  return request.post('/menu', data)
}

/** 编辑菜单 */
export function updateMenu(id: number, data: any): Promise<void> {
  return request.put(`/menu/${id}`, data)
}

/** 删除菜单 */
export function deleteMenu(id: number): Promise<void> {
  return request.delete(`/menu/${id}`)
}
