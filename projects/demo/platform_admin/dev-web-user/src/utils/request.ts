/**
 * Axios 请求封装
 * - baseURL: /api/v1/user
 * - 请求拦截器：自动附加 Authorization: Bearer {token}
 * - 响应拦截器：401 时跳转登录页
 */

import axios from 'axios'
import type { AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'

const service = axios.create({
  baseURL: '/api/v1/user',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' }
})

/** 请求拦截器：注入 Token */
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('access_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

/** 响应拦截器：统一错误处理 */
service.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data
    // 后端统一成功码为 0
    if (res.code !== 0) {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res.data
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('access_token')
      window.location.href = '/login'
      return Promise.reject(error)
    }
    ElMessage.error(error.response?.data?.message || error.message || '网络异常')
    return Promise.reject(error)
  }
)

export default service
