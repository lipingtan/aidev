/**
 * 认证状态管理（多租户模式）
 * - platform_token：登录后获取，用于租户选择和 token 刷新
 * - access_token：选择租户后获取，用于业务 API 请求
 * - 登录 → 单租户自动跳过 → 多租户展示选择页
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request from '@/utils/request'

/** 租户信息 */
export interface TenantInfo {
  id: string
  name: string
  logo?: string
  description?: string
}

/** 登录响应 */
export interface LoginResponse {
  platform_token: string
  access_token?: string
  tenants: TenantInfo[]
}

/** 租户选择响应 */
export interface TenantSelectResponse {
  access_token: string
  expires_in: number
}

/** 刷新 Token 响应 */
export interface RefreshResponse {
  access_token: string
  expires_in: number
}

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const platformToken = ref<string>(localStorage.getItem('platform_token') || '')
  const accessToken = ref<string>(localStorage.getItem('access_token') || '')
  const tenants = ref<TenantInfo[]>(JSON.parse(localStorage.getItem('tenant_list') || '[]'))
  const currentTenantId = ref<string>(localStorage.getItem('current_tenant_id') || '')

  // 计算属性
  const isLoggedIn = computed(() => !!accessToken.value)
  const hasPlatformToken = computed(() => !!platformToken.value)

  /**
   * 登录：调用 POST /auth/login
   * 返回 platform_token 和租户列表
   */
  async function login(username: string, password: string, captchaKey?: string, captchaCode?: string): Promise<LoginResponse> {
    const payload: Record<string, string> = { username, password }
    if (captchaKey) payload.captcha_key = captchaKey
    if (captchaCode) payload.captcha_code = captchaCode
    const res = await request.post<any, LoginResponse>('/auth/login', payload)
    platformToken.value = res.platform_token
    tenants.value = res.tenants
    localStorage.setItem('platform_token', res.platform_token)
    localStorage.setItem('tenant_list', JSON.stringify(res.tenants))
    // 单租户时后端直接返回 access_token，直接存储
    if (res.access_token) {
      accessToken.value = res.access_token
      localStorage.setItem('access_token', res.access_token)
      if (res.tenants?.length === 1) {
        currentTenantId.value = String(res.tenants[0].id)
        localStorage.setItem('current_tenant_id', String(res.tenants[0].id))
      }
    }
    return res
  }

  /**
   * 选择租户：调用 POST /auth/tenant/select
   * 返回 access_token
   */
  async function selectTenant(tenantId: string): Promise<TenantSelectResponse> {
    const res = await request.post<any, TenantSelectResponse>(
      '/auth/tenant/select',
      { tenant_id: tenantId },
      { headers: { Authorization: `Bearer ${platformToken.value}` } }
    )
    accessToken.value = res.access_token
    currentTenantId.value = tenantId
    localStorage.setItem('access_token', res.access_token)
    localStorage.setItem('current_tenant_id', tenantId)
    return res
  }

  /**
   * 刷新 Token：调用 POST /auth/refresh
   * 使用 platform_token 刷新 access_token
   */
  async function refresh(): Promise<RefreshResponse> {
    const res = await request.post<any, RefreshResponse>(
      '/auth/refresh',
      {},
      { headers: { Authorization: `Bearer ${platformToken.value}` } }
    )
    accessToken.value = res.access_token
    localStorage.setItem('access_token', res.access_token)
    return res
  }

  /** 登出：清除所有认证状态 */
  function logout(): void {
    platformToken.value = ''
    accessToken.value = ''
    tenants.value = []
    currentTenantId.value = ''
    localStorage.removeItem('platform_token')
    localStorage.removeItem('access_token')
    localStorage.removeItem('current_tenant_id')
    localStorage.removeItem('tenant_list')
  }

  return {
    platformToken,
    accessToken,
    tenants,
    currentTenantId,
    isLoggedIn,
    hasPlatformToken,
    login,
    selectTenant,
    refresh,
    logout
  }
})
