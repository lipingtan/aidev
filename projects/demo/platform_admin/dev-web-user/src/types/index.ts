/**
 * 全局类型定义
 */

/** 后端统一响应结构 */
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

/** 菜单项 */
export interface MenuItem {
  id: string
  parentId: string | null
  name: string
  path: string
  component: string
  title: string
  icon?: string
  sort: number
  hidden: boolean
  children?: MenuItem[]
}

/** 用户信息 */
export interface UserInfo {
  phone: string
  token: string
}
