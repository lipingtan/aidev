/**
 * 系统初始化 API
 * 对接后端 /setup/* 接口
 */

import axios from 'axios'

// 初始化接口使用独立的 axios 实例（不走 JWT 拦截器，不带 /api/v1 前缀）
const setupRequest = axios.create({
  baseURL: '',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
})

/** 安装请求参数（与后端 SetupRequest 一致） */
export interface SetupConfig {
  dbHost: string
  dbPort: number
  dbName: string
  dbUser: string
  dbPassword: string
  dbCharset?: string
  redisHost?: string
  redisPort?: number
  redisPassword?: string
  redisDb?: number
  appName?: string
  appPort?: number
}

/** 测试数据库连接参数 */
export interface TestDBParams {
  dbHost: string
  dbPort: number
  dbName: string
  dbUser: string
  dbPassword: string
}

/** 初始化状态 */
export interface InitStatus {
  initialized: boolean
}

/** 检查初始化状态（不需要认证） */
export async function checkInitStatus(): Promise<InitStatus> {
  try {
    const res = await setupRequest.get('/setup/status')
    return { initialized: res.data?.installed ?? true }
  } catch (e: any) {
    // 503 = 后端未初始化（InstallMiddleware 拦截）
    if (e.response?.status === 503) {
      return { initialized: false }
    }
    // 其他错误（网络不通、401等）视为已初始化，走正常登录流程
    return { initialized: true }
  }
}

/** 测试数据库连接 */
export async function testDBConnection(params: TestDBParams): Promise<{ code: number; msg: string }> {
  const res = await setupRequest.post('/setup/test-db', params)
  return res.data
}

/** 执行初始化安装 */
export async function executeInit(config: SetupConfig): Promise<{ code: number; msg: string; data?: any }> {
  const res = await setupRequest.post('/setup/init', config)
  return res.data
}
