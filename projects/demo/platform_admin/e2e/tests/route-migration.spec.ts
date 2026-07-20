import { test, expect } from '@playwright/test'

const API_BASE = 'http://localhost:8000'

test.describe('路由前缀迁移验证 (Task 2)', () => {
  let accessToken: string

  test.beforeAll(async ({ request }) => {
    const resp = await request.post(`${API_BASE}/auth/login`, {
      data: { username: 'admin', password: 'admin123' },
    })
    const body = await resp.json()
    accessToken = body.data.access_token
  })

  test('TC-N01: 旧路径 /api/v1/tenants 返回 404', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/tenants`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    expect(resp.status()).toBe(404)
  })

  test('TC-N02: 旧路径 /api/v1/resources/user-menu 返回 404', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/resources/user-menu`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    expect(resp.status()).toBe(404)
  })

  test('TC-006: 新路径 /api/v1/admin/tenants 正常', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/admin/tenants`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    expect(body.code).toBe(0)
    expect(body.data.list.length).toBeGreaterThanOrEqual(1)
  })

  test('TC-007: 新路径 /api/v1/common/user-menu 正常', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/common/user-menu?platform=admin`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    expect(body.code).toBe(0)
    expect(body.data.length).toBeGreaterThanOrEqual(4)
  })

  test('TC-A03: 未认证返回 401', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/admin/tenants`)
    expect(resp.status()).toBe(401)
  })

  test('TC-R07: SUPER_ADMIN 访问所有接口不被拦截', async ({ request }) => {
    const endpoints = [
      '/api/v1/admin/tenants',
      '/api/v1/admin/users',
      '/api/v1/admin/roles',
      '/api/v1/admin/applications',
      '/api/v1/admin/configs',
      '/api/v1/admin/operation-logs',
      '/api/v1/admin/login-logs',
    ]
    for (const ep of endpoints) {
      const resp = await request.get(`${API_BASE}${ep}`, {
        headers: { Authorization: `Bearer ${accessToken}` },
      })
      expect(resp.status(), `${ep} 应返回 200`).toBe(200)
      const body = await resp.json()
      expect(body.code, `${ep} code 应为 0`).toBe(0)
    }
  })
})
