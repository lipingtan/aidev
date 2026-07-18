/**
 * 菜单管理 API（对接 go-admin resource 接口）
 */
import request from '@/utils/request'
import type { MenuItem } from '@/store/modules/menu'

export interface MenuTreeItem {
  id: string
  name: string
  parent_id: string | null
  type: string
  path: string
  component: string
  permission_code: string
  icon: string
  sort_order: number
  status: number
  app_code: string
  children?: MenuTreeItem[]
}

/** 获取当前用户动态菜单（侧边栏渲染用） */
export function fetchUserMenu(): Promise<MenuItem[]> {
  return request.get('/api/v1/resources/user-menu')
}

/** 获取当前用户角色菜单（侧边栏用） */
export function getUserMenuTree(): Promise<MenuTreeItem[]> {
  // 新 RBAC 接口，失败时返回空数组不阻塞
  return request.get('/api/v1/resources/user-menu').catch(() => [])
}

/** 菜单列表（管理用）— 实际是 resource tree */
export function getMenuList(params?: { menuName?: string; status?: number }): Promise<MenuTreeItem[]> {
  return request.get('/api/v1/resources/tree', { params })
}

/** 菜单列表（兼容旧引用） */
export const getMenuTree = getMenuList

/** 新增菜单（实际是创建 resource） */
export function createMenu(data: any): Promise<void> {
  return request.post('/api/v1/resources', data)
}

/** 编辑菜单（实际是更新 resource） */
export function updateMenu(id: string, data: any): Promise<void> {
  return request.put(`/api/v1/resources/${id}`, data)
}

/** 删除菜单（实际是删除 resource） */
export function deleteMenu(id: string): Promise<void> {
  return request.delete(`/api/v1/resources/${id}`)
}
