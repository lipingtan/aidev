import { test, expect } from '@playwright/test'

const API_BASE = 'http://localhost:8000'

test.describe('CRUD 回归 (RG-3, RG-4)', () => {
  let accessToken: string
  let headers: Record<string, string>

  test.beforeAll(async ({ request }) => {
    const resp = await request.post(`${API_BASE}/auth/login`, {
      data: { username: 'admin', password: 'admin123' },
    })
    const body = await resp.json()
    accessToken = body.data.access_token
    headers = { Authorization: `Bearer ${accessToken}` }
  })

  test('TC-R03: 角色 CRUD 完整流程', async ({ request }) => {
    const code = `e2e_role_${Date.now()}`
    // 创建
    const createResp = await request.post(`${API_BASE}/api/v1/admin/roles`, {
      headers,
      data: { role_code: code, role_name: 'E2E测试角色' },
    })
    expect(createResp.status()).toBe(200)
    const createBody = await createResp.json()
    expect(createBody.code).toBe(0)
    const roleId = createBody.data.id

    // 查询
    const listResp = await request.get(`${API_BASE}/api/v1/admin/roles`, { headers })
    expect(listResp.status()).toBe(200)
    const listBody = await listResp.json()
    expect(listBody.code).toBe(0)

    // 更新
    const updateResp = await request.put(`${API_BASE}/api/v1/admin/roles/${roleId}`, {
      headers,
      data: { role_name: 'E2E更新角色', version: 1 },
    })
    expect(updateResp.status()).toBe(200)

    // 删除
    const delResp = await request.delete(`${API_BASE}/api/v1/admin/roles/${roleId}`, { headers })
    expect(delResp.status()).toBe(200)
    expect((await delResp.json()).code).toBe(0)
  })

  test('TC-R04: 用户创建 + 关联租户 + 删除', async ({ request }) => {
    const username = `e2e_user_${Date.now()}`
    // 创建用户
    const createResp = await request.post(`${API_BASE}/api/v1/admin/users`, {
      headers,
      data: { username, password: 'Test1234', nickname: 'E2E用户' },
    })
    expect(createResp.status()).toBe(200)
    const createBody = await createResp.json()
    expect(createBody.code).toBe(0)
    const userId = createBody.data.id

    // 查询用户列表
    const listResp = await request.get(`${API_BASE}/api/v1/admin/users`, { headers })
    expect(listResp.status()).toBe(200)

    // 删除
    const delResp = await request.delete(`${API_BASE}/api/v1/admin/users/${userId}`, { headers })
    expect(delResp.status()).toBe(200)
  })
})
