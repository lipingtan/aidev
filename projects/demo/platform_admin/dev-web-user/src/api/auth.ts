/**
 * 认证相关 API
 * - 发送验证码
 * - 登录（手机号 + 验证码 + 租户编码）
 * - 登出
 * - 获取菜单
 */

import request from '@/utils/request'
import axios from 'axios'
import type { MenuItem } from '@/types'

/** 发送验证码 */
export function sendCode(phone: string, tenantCode: string) {
  return request.post('/auth/send-code', { phone, tenant_code: tenantCode })
}

/** 登录 */
export function login(phone: string, code: string, tenantCode: string) {
  return request.post<any, { access_token: string }>('/auth/login', {
    account: phone,   // 后端 SmsLoginRequest.Account 字段
    code,
    tenant_code: tenantCode
  })
}

/** 登出 */
export function logout() {
  return request.post('/auth/logout')
}

/** 获取当前用户菜单（接口在 /api/v1/common 路由组，需绕过 baseURL） */
export function getMenu() {
  const token = localStorage.getItem('access_token')
  return axios.get<any, { data: { data: any[] } }>('/api/v1/common/user-menu', {
    params: { platform: 'user' },
    headers: token ? { Authorization: `Bearer ${token}` } : {}
  }).then(res => {
    const raw: any[] = (res.data as any)?.data ?? []
    // 后端返回 name 字段，前端 MenuItem 用 title 字段，做映射
    return raw.map(item => ({
      id: String(item.id),
      parentId: item.parent_id,
      name: item.name,
      title: item.name,        // 关键：映射 name → title
      path: item.path,
      component: item.component || 'plugin/container',
      icon: item.icon || '',
      sort: item.sort_order || 0,
      hidden: item.hidden || false,
      children: []
    }))
  })
}
