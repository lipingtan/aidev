/**
 * 用户信息 API（对接 go-admin）
 * 后端接口：GET /api/v1/getinfo
 */
import request from '@/utils/request'

export interface ProfileInfo {
  userId: string
  username: string
  realName: string
  phone?: string
  avatar: string
  roles: string[]
  permissions: string[]
  tenantId?: string
}

/** 获取个人信息 */
export async function getProfile(): Promise<ProfileInfo> {
  // TODO: 对接新 auth-rbac 用户信息接口，当前返回 localStorage 中的基本信息
  const tenantId = localStorage.getItem('current_tenant_id') || ''
  return {
    userId: '0',
    username: 'admin',
    realName: '超级管理员',
    avatar: '',
    roles: ['SUPER_ADMIN'],
    permissions: ['*'],
    tenantId,
  }
}

/** 获取个人详细信息 */
export function getProfileDetail() {
  return request.get('/user/profile')
}

/** 修改个人信息（V2 用户更新接口） */
export async function updateProfile(userId: string, data: { nickname?: string; phone?: string; version: number }): Promise<void> {
  return request.put(`/api/v1/admin/users/${userId}`, data)
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
