/**
 * 用户状态管理
 * - token / 用户信息 / 权限 / 菜单
 * - 对接后端 auth-rbac 认证和个人中心 API
 */

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { getProfile } from '@/api/profile'
import { getUserMenuTree } from '@/api/menu'
import type { ProfileInfo } from '@/api/profile'
import type { MenuTreeItem } from '@/api/menu'
import request from '@/utils/request'

/** 验证码结果 */
export interface CaptchaResult {
  captcha_key: string
  captcha_image: string  // base64 图片
}

export const useUserStore = defineStore('user', () => {
  const router = useRouter()

  const token = ref<string>(localStorage.getItem('access_token') || '')
  const refreshTokenVal = ref<string>(localStorage.getItem('refresh_token') || '')
  const userId = ref<number>(0)
  const username = ref<string>('')
  const realName = ref<string>('')
  const avatar = ref<string>('')
  const roles = ref<string[]>([])
  const permissions = ref<string[]>([])
  const menus = ref<MenuTreeItem[]>([])

  /** 获取用户信息 */
  async function fetchUserInfo(): Promise<ProfileInfo> {
    const info = await getProfile()
    userId.value = info.userId
    username.value = info.username
    realName.value = info.realName
    avatar.value = info.avatar || ''
    roles.value = info.roles || []
    permissions.value = info.permissions || []
    return info
  }

  /** 获取用户菜单 */
  async function fetchMenus(): Promise<MenuTreeItem[]> {
    const tree = await getUserMenuTree()
    menus.value = tree
    return tree
  }

  /** 获取验证码（新 auth-rbac 接口） */
  async function fetchCaptcha(): Promise<CaptchaResult> {
    return request.get('/auth/captcha')
  }

  /** 登出 */
  async function logout(): Promise<void> {
    resetState()
    router.push('/login')
  }

  /** 重置状态 */
  function resetState(): void {
    token.value = ''
    refreshTokenVal.value = ''
    userId.value = 0
    username.value = ''
    realName.value = ''
    avatar.value = ''
    roles.value = []
    permissions.value = []
    menus.value = []
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  /** 检查权限 */
  function hasPermission(perm: string): boolean {
    return permissions.value.includes(perm)
  }

  return {
    token, refreshTokenVal, userId, username, realName, avatar,
    roles, permissions, menus,
    fetchUserInfo, fetchMenus, fetchCaptcha, logout, resetState, hasPermission
  }
})
