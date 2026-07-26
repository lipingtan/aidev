/**
 * 认证相关 API
 * - 发送验证码
 * - 登录（手机号 + 验证码 + 租户编码）
 * - 登出
 * - 获取菜单
 */

import request from '@/utils/request'
import type { MenuItem } from '@/types'

/** 发送验证码 */
export function sendCode(phone: string, tenantCode: string) {
  return request.post('/auth/send-code', { phone, tenant_code: tenantCode })
}

/** 登录 */
export function login(phone: string, code: string, tenantCode: string) {
  return request.post<any, { token: string }>('/auth/login', { phone, code, tenant_code: tenantCode })
}

/** 登出 */
export function logout() {
  return request.post('/auth/logout')
}

/** 获取当前用户菜单 */
export function getMenu() {
  return request.get<any, MenuItem[]>('/api/v1/common/user-menu', {
    params: { platform: 'user' }
  })
}
