/**
 * 认证 API（对接 go-admin 后端）
 * 登录和验证码接口返回格式特殊，使用独立 axios 实例
 */
import axios from 'axios'
import request from '@/utils/request'

// 认证相关接口使用独立 axios 实例（不走响应拦截器的 .data 解包）
const authRequest = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' }
})

export interface LoginParams {
  username: string
  password: string
  code: string   // 验证码
  uuid: string   // 验证码 key
}

export interface TokenResult {
  token: string
  expire: string
}

export interface CaptchaResult {
  id: string
  data: string  // base64 图片
}

/** 获取图形验证码 */
export async function getCaptcha(): Promise<CaptchaResult> {
  const res = await authRequest.get('/captcha')
  const body = res.data
  return { id: body.id, data: body.data }
}

/** 用户名密码登录 */
export async function login(data: LoginParams): Promise<TokenResult> {
  const res = await authRequest.post('/login', data)
  const body = res.data
  if (body.code !== 200) {
    throw new Error(body.msg || '登录失败')
  }
  return { token: body.token, expire: body.expire }
}

/** 刷新 Token */
export function refreshToken(): Promise<TokenResult> {
  return request.get('/refresh_token')
}

/** 登出（前端清除 token 即可） */
export function logout(): Promise<void> {
  return Promise.resolve()
}
