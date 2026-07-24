/**
 * 用户状态管理
 * - token / phone / menus 管理
 * - login / logout / fetchMenu actions
 */

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi, logout as logoutApi, getMenu } from '@/api/auth'
import type { MenuItem } from '@/types'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('access_token') || '')
  const phone = ref<string>(localStorage.getItem('user_phone') || '')
  const menus = ref<MenuItem[]>([])

  /** 登录 */
  async function login(phoneVal: string, code: string, tenantCode: string) {
    const res = await loginApi(phoneVal, code, tenantCode)
    token.value = res.token
    phone.value = phoneVal
    localStorage.setItem('access_token', res.token)
    localStorage.setItem('user_phone', phoneVal)
  }

  /** 登出 */
  async function logout() {
    try {
      await logoutApi()
    } finally {
      token.value = ''
      phone.value = ''
      menus.value = []
      localStorage.removeItem('access_token')
      localStorage.removeItem('user_phone')
      router.push('/login')
    }
  }

  /** 获取菜单 */
  async function fetchMenu() {
    const data = await getMenu()
    menus.value = data
  }

  return {
    token,
    phone,
    menus,
    login,
    logout,
    fetchMenu
  }
})
