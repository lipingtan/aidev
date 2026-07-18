/**
 * Axios 请求封装
 * - 创建 Axios 实例（baseURL 从环境变量读取、超时 15000ms）
 * - 请求拦截器：注入 access_token
 * - 响应拦截器：统一错误处理、401 自动刷新 token、刷新失败跳转登录页
 */

import axios from 'axios'
import type { AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiResponse } from '@/types'
import { showLoading, hideLoading } from './loading'

/** 创建 Axios 实例 */
const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' }
})

/** 是否正在刷新 token */
let isRefreshing = false
/** 等待刷新完成的请求队列 */
let refreshQueue: Array<(token: string) => void> = []

/** 执行 token 刷新 */
async function doRefreshToken(): Promise<string> {
  const platformToken = localStorage.getItem('platform_token')
  if (!platformToken) {
    throw new Error('无 platform_token')
  }
  const tenantId = localStorage.getItem('current_tenant_id') || ''
  const res = await axios.post(
    `${import.meta.env.VITE_API_BASE_URL}/auth/refresh`,
    { tenant_id: tenantId },
    { headers: { Authorization: `Bearer ${platformToken}` } }
  )
  const body = res.data
  if (body.code !== 0 && body.code !== 200) {
    throw new Error(body.message || '刷新失败')
  }
  const newToken = body.data.access_token
  localStorage.setItem('access_token', newToken)
  return newToken
}

/** 跳转登录页并清除所有 token */
function redirectToLogin(): void {
  localStorage.removeItem('access_token')
  localStorage.removeItem('platform_token')
  localStorage.removeItem('current_tenant_id')
  if (!window.location.pathname.includes('/login') && !window.location.pathname.includes('/init')) {
    window.location.href = '/login'
  }
}

/** 请求拦截器：从 localStorage 读取 Token 注入 Authorization 头 */
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('access_token')
    if (token && !config.headers.Authorization) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

/** 响应拦截器：检查业务状态码、处理 401 自动刷新 */
service.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const res = response.data
    // 业务状态码：0 或 200 均为成功
    if (res.code !== 0 && res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    // 分页响应：包含 total 字段时返回完整对象（保留分页元数据）
    if ((res as any).total !== undefined) {
      return res as any
    }
    return res.data as any
  },
  async (error) => {
    const originalConfig = error.config

    // 网络异常或超时：提示错误（不做页面跳转避免循环）
    if (!error.response || error.code === 'ECONNABORTED') {
      ElMessage.error('网络连接失败，请检查后端服务是否启动')
      return Promise.reject(error)
    }

    // 非 401 错误直接报错
    if (error.response?.status !== 401) {
      ElMessage.error(error.response?.data?.message || error.message || '网络异常')
      return Promise.reject(error)
    }

    // 401 处理：尝试用 platform_token 刷新 access_token
    const platformToken = localStorage.getItem('platform_token')

    // 无 platform_token 或刷新接口本身返回 401，直接跳登录
    if (!platformToken || originalConfig._isRetry) {
      redirectToLogin()
      return Promise.reject(error)
    }

    // 标记为重试请求，防止无限循环
    originalConfig._isRetry = true

    if (isRefreshing) {
      // 已有刷新请求在进行中，排队等待
      return new Promise((resolve) => {
        refreshQueue.push((newToken: string) => {
          originalConfig.headers.Authorization = `Bearer ${newToken}`
          resolve(service(originalConfig))
        })
      })
    }

    isRefreshing = true
    showLoading('正在刷新认证...')
    try {
      const newToken = await doRefreshToken()
      // 刷新成功，执行队列中的请求
      refreshQueue.forEach((cb) => cb(newToken))
      refreshQueue = []
      // 重试原始请求
      originalConfig.headers.Authorization = `Bearer ${newToken}`
      return service(originalConfig)
    } catch {
      // 刷新失败，清除队列并跳转登录
      refreshQueue = []
      redirectToLogin()
      return Promise.reject(error)
    } finally {
      isRefreshing = false
      hideLoading()
    }
  }
)

export default service
