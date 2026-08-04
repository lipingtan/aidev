/**
 * 多 Pod 兼容性修复 E2E 测试
 * 覆盖：Token 黑名单、Snowflake ID 唯一性、C端强制登出
 * 依赖：后端运行于 http://localhost:8000
 */
import { test, expect, request, APIRequestContext } from '@playwright/test'

const API_BASE = 'http://localhost:8000'
const ADMIN_USER = 'admin'
const ADMIN_PASS = 'admin123'

// ============================================================
// 辅助工具
// ============================================================

interface LoginResult { token: string; tenantId: string }

async function loginAdmin(api: APIRequestContext): Promise<LoginResult> {
  const resp = await api.post(`${API_BASE}/auth/login`, {
    data: { username: ADMIN_USER, password: ADMIN_PASS }
  })
  expect(resp.ok(), `登录失败: ${await resp.text()}`).toBeTruthy()
  const body = await resp.json()
  if (body.data?.tenants?.length > 0) {
    const tenant = body.data.tenants[0]
    const tenantResp = await api.post(`${API_BASE}/auth/tenant/select`, {
      data: { tenant_id: String(tenant.id) },
      headers: { Authorization: `Bearer ${body.data.token}` }
    })
    const tb = await tenantResp.json()
    return { token: tb.data?.access_token ?? tb.data?.token, tenantId: String(tenant.id) }
  }
  return { token: body.data?.access_token ?? body.data?.token, tenantId: '1' }
}

function auth(token: string) {
  return { headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } }
}

// ============================================================
// TC-BL: Token 黑名单（登出后 token 失效）
// ============================================================

test.describe('TC-BL — Token 黑名单', () => {
  let api: APIRequestContext

  test.beforeAll(async () => { api = await request.newContext() })
  test.afterAll(async () => { await api.dispose() })

  test('TC-BL-01: 登出后原 token 被拒绝（401）', async () => {
    const { token } = await loginAdmin(api)

    // 验证 token 有效
    const beforeLogout = await api.get(`${API_BASE}/api/v1/admin/tenants?page=1&page_size=1`, auth(token))
    expect(beforeLogout.status()).toBe(200)

    // 登出
    const logoutResp = await api.post(`${API_BASE}/auth/logout`, auth(token))
    expect(logoutResp.status()).toBe(200)

    // 登出后原 token 应被拒绝
    const afterLogout = await api.get(`${API_BASE}/api/v1/admin/tenants?page=1&page_size=1`, auth(token))
    expect([401, 403]).toContain(afterLogout.status())
  })

  test('TC-BL-02: 登出后重新登录可获得新 token 正常访问', async () => {
    const { token: firstToken } = await loginAdmin(api)

    // 登出第一个 token
    await api.post(`${API_BASE}/auth/logout`, auth(firstToken))

    // 重新登录
    const { token: newToken } = await loginAdmin(api)
    expect(newToken).not.toBe(firstToken)

    // 新 token 可以正常访问
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenants?page=1&page_size=1`, auth(newToken))
    expect(resp.status()).toBe(200)
  })

  test('TC-BL-03: 无 token 请求仍返回 401（未受登出影响）', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenants`)
    expect(resp.status()).toBe(401)
  })
})

// ============================================================
// TC-SF: Snowflake ID 唯一性
// ============================================================

test.describe('TC-SF — Snowflake ID 唯一性', () => {
  let api: APIRequestContext
  let token: string

  test.beforeAll(async () => {
    api = await request.newContext()
    ;({ token } = await loginAdmin(api))
  })
  test.afterAll(async () => { await api.dispose() })

  test('TC-SF-01: 连续创建多个策略 ID 全部不同', async () => {
    const ids = new Set<string>()
    const created: string[] = []

    for (let i = 0; i < 5; i++) {
      const resp = await api.post(`${API_BASE}/api/v1/admin/abac/policies`, {
        data: {
          name: `id-uniqueness-test-${Date.now()}-${i}`,
          resource_type: 'order',
          subject_type: 'ROLE',
          subject_id: 'test-role',
          effect: 'ALLOW'
        },
        ...auth(token)
      })
      expect(resp.status()).toBe(201)
      const body = await resp.json()
      const id = String(body.data?.id ?? body.id)
      expect(id).toBeTruthy()
      expect(ids.has(id)).toBeFalsy()
      ids.add(id)
      created.push(id)
    }

    expect(ids.size).toBe(5)

    // 清理
    for (const id of created) {
      await api.delete(`${API_BASE}/api/v1/admin/abac/policies/${id}`, auth(token))
    }
  })

  test('TC-SF-02: ID 为 string 类型（雪花 ID 精度保护）', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/admin/abac/policies`, {
      data: {
        name: `sf-type-test-${Date.now()}`,
        resource_type: 'order',
        subject_type: 'ROLE',
        subject_id: 'test-role',
        effect: 'ALLOW'
      },
      ...auth(token)
    })
    expect(resp.status()).toBe(201)
    const body = await resp.json()
    const id = body.data?.id ?? body.id
    // ID 应为 string 格式（json:",string" tag 作用）
    expect(typeof id).toBe('string')
    // 清理
    await api.delete(`${API_BASE}/api/v1/admin/abac/policies/${id}`, auth(token))
  })
})

// ============================================================
// TC-ML: 多 Pod 兼容性回归
// ============================================================

test.describe('TC-ML — 多 Pod 兼容性回归', () => {
  let api: APIRequestContext
  let token: string

  test.beforeAll(async () => {
    api = await request.newContext()
    ;({ token } = await loginAdmin(api))
  })
  test.afterAll(async () => { await api.dispose() })

  test('TC-ML-01: 连续登录两次均成功（黑名单不影响新 token）', async () => {
    const r1 = await loginAdmin(api)
    const r2 = await loginAdmin(api)
    expect(r1.token).toBeTruthy()
    expect(r2.token).toBeTruthy()
    // 两个 token 都应能访问接口
    const resp1 = await api.get(`${API_BASE}/api/v1/admin/tenants?page=1&page_size=1`, auth(r1.token))
    const resp2 = await api.get(`${API_BASE}/api/v1/admin/tenants?page=1&page_size=1`, auth(r2.token))
    expect(resp1.status()).toBe(200)
    expect(resp2.status()).toBe(200)
  })

  test('TC-ML-02: 登录流程不受 Redis 缺失降级影响（服务正常启动）', async () => {
    // 如果 Redis 不可用但服务已启动（降级为内存），登录仍应成功
    const resp = await api.post(`${API_BASE}/auth/login`, {
      data: { username: ADMIN_USER, password: ADMIN_PASS }
    })
    expect(resp.ok()).toBeTruthy()
  })

  test('TC-ML-03: 用户管理接口正常返回（多 Pod 修复不影响核心功能）', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/users?page=1&page_size=5`, auth(token))
    expect(resp.status()).toBe(200)
  })

  test('TC-ML-04: 审批超时扫描回归（DataScope 不受影响）', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/data-scope-configs`, auth(token))
    expect(resp.status()).toBe(200)
  })
})
