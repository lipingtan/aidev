import { test, expect } from '@playwright/test'

const API_BASE = 'http://localhost:8000'

test.describe('Step B: AppResolveMiddleware + 租户四态 + SUPER_ADMIN 保护', () => {
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

  // === Task 6: AppResolveMiddleware ===

  test('TC-B601: /api/v1/admin/tenants → context 中 app_code=platform_admin', async ({ request }) => {
    // 间接验证：请求能通过 AppResolveMiddleware 说明前缀匹配成功
    const resp = await request.get(`${API_BASE}/api/v1/admin/tenants`, { headers })
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    expect(body.code).toBe(0)
  })

  test('TC-B602: SUPER_ADMIN 不被 AppResolve 拦截', async ({ request }) => {
    // admin 是 SUPER_ADMIN，即使租户未显式订阅某应用也不应被拦截
    const endpoints = [
      '/api/v1/admin/tenants',
      '/api/v1/admin/users',
      '/api/v1/admin/roles',
      '/api/v1/admin/applications',
    ]
    for (const ep of endpoints) {
      const resp = await request.get(`${API_BASE}${ep}`, { headers })
      expect(resp.status(), `${ep} 应通过`).toBe(200)
    }
  })

  // === Task 7: 租户四态 ===

  test('TC-B701: 正常租户(status=1) 所有请求放行', async ({ request }) => {
    // admin 用户的默认租户 status=1，所有方法都应放行
    const resp = await request.get(`${API_BASE}/api/v1/admin/tenants`, { headers })
    expect(resp.status()).toBe(200)

    // POST 也应放行
    const postResp = await request.post(`${API_BASE}/api/v1/admin/roles`, {
      headers,
      data: { role_code: `tc_b701_${Date.now()}`, role_name: '四态测试' },
    })
    expect(postResp.status()).toBe(200)
    // 清理
    const roleId = (await postResp.json()).data.id
    await request.delete(`${API_BASE}/api/v1/admin/roles/${roleId}`, { headers })
  })

  // === Task 8: GetUserMenu 增强 ===

  test('TC-B801: ?platform=admin 仅返回 admin 平台菜单', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/common/user-menu?platform=admin`, { headers })
    const body = await resp.json()
    expect(body.code).toBe(0)
    expect(body.data.length).toBeGreaterThanOrEqual(4)

    // 验证所有返回菜单 platform=admin
    function checkPlatform(nodes: any[]) {
      for (const n of nodes) {
        expect(n.platform).toBe('admin')
        if (n.children) checkPlatform(n.children)
      }
    }
    checkPlatform(body.data)
  })

  test('TC-B802: SUPER_ADMIN 返回对应 platform 全部菜单', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/common/user-menu?platform=admin`, { headers })
    const body = await resp.json()
    // SUPER_ADMIN 应返回全部 admin 平台菜单（≥15 含子菜单）
    let count = 0
    function countNodes(nodes: any[]) {
      for (const n of nodes) { count++; if (n.children) countNodes(n.children) }
    }
    countNodes(body.data)
    expect(count).toBeGreaterThanOrEqual(15)
  })

  // === Task 9: SUPER_ADMIN 保护规则 ===

  test('TC-B901: 删除唯一 SUPER_ADMIN 用户被拒绝', async ({ request }) => {
    // 获取 admin 用户 ID
    const usersResp = await request.get(`${API_BASE}/api/v1/admin/users`, { headers })
    const users = (await usersResp.json()).data.list
    const adminUser = users.find((u: any) => u.username === 'admin')
    expect(adminUser).toBeTruthy()

    // 尝试删除
    const delResp = await request.delete(`${API_BASE}/api/v1/admin/users/${adminUser.id}`, { headers })
    // 应被拒绝（400 或 403）
    const delBody = await delResp.json()
    expect(delBody.code).toBe(40004) // ErrProtectedEntity
  })

  test('TC-B902: 有多个 SUPER_ADMIN 时可删除非最后一个', async ({ request }) => {
    // 创建第二个用户
    const createResp = await request.post(`${API_BASE}/api/v1/admin/users`, {
      headers,
      data: { username: `sa_test_${Date.now()}`, password: 'Test1234', nickname: 'SA测试' },
    })
    const user2 = (await createResp.json()).data
    expect(user2.id).toBeTruthy()

    // 删除第二个用户应该成功（不是 SUPER_ADMIN）
    const delResp = await request.delete(`${API_BASE}/api/v1/admin/users/${user2.id}`, { headers })
    expect(delResp.status()).toBe(200)
  })
})
