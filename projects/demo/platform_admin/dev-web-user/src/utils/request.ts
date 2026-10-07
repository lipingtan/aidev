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
    // 后端成功码：0 或 200 均视为成功
    if (res.code !== 0 && res.code !== 200) {
      // 提取 error_code 附到 Error 对象，方便上层组件精细化处理
      const errorCode: number | undefined = (res as any).error_code ?? res.code
      const msg = (res as any).msg || res.message || '请求失败'
      ElMessage.error(msg)
      const err = Object.assign(new Error(msg), { error_code: errorCode, response })
      return Promise.reject(err)
    }
    return res.data
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('access_token')
      window.location.href = '/login'
      return Promise.reject(error)
    }
    ElMessage.error(error.response?.data?.msg || error.response?.data?.message || error.message || '网络异常')
    return Promise.reject(error)
  }
)

export default service
