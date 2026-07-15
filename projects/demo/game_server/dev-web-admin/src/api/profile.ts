/**
 * 用户信息 API（对接 go-admin）
 * 后端接口：GET /api/v1/getinfo
 */
import request from '@/utils/request'

export interface ProfileInfo {
  userId: number
  username: string
  realName: string
  phone: string
  avatar: string
  roles: string[]
  permissions: string[]
}

/** 获取个人信息 */
export async function getProfile(): Promise<ProfileInfo> {
  const res: any = await request.get('/getinfo')
  return {
    userId: res?.userId || res?.user_id || 0,
    username: res?.userName || res?.username || '',
    realName: res?.realName || res?.nick_name || '',
    phone: res?.phone || '',
    avatar: res?.avatar || '',
    roles: res?.roles || [],
    permissions: res?.permissions || []
  }
}

/** 获取个人详细信息 */
export function getProfileDetail() {
  return request.get('/user/profile')
}

/** 修改个人信息（go-admin 用 SysUser Update 接口） */
export function updateProfile(data: { realName?: string; phone?: string }): Promise<void> {
  return request.put('/sys-user', data)
}

/** 更新头像 */
export function updateAvatar(file: File): Promise<void> {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/user/avatar', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

/** 修改密码 */
export function updatePassword(data: { oldPassword: string; newPassword: string }): Promise<void> {
  return request.put('/user/pwd/set', data)
}
