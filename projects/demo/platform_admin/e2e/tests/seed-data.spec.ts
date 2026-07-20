import { test, expect } from '@playwright/test'

const API_BASE = 'http://localhost:8000'

test.describe('Seed 数据验证 (Task 4)', () => {
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

  test('TC-D01: platform_admin 应用字段完整', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/admin/applications`, { headers })
    const body = await resp.json()
    const app = body.data.find((a: any) => a.app_code === 'platform_admin')

    expect(app).toBeTruthy()
    expect(app.app_type).toBe('BUILTIN')
    expect(app.route_prefix).toBe('/api/v1/admin')
    // platforms 可能已被 JSON 反序列化为数组或字符串
    const platforms = typeof app.platforms === 'string' ? JSON.parse(app.platforms) : app.platforms
    expect(platforms).toContain('admin:pc')
  })

  test('TC-D02: 菜单全部 platform=admin, app_code=platform_admin', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/common/user-menu?platform=admin`, { headers })
    const body = await resp.json()

    function checkNodes(nodes: any[]) {
      for (const node of nodes) {
        expect(node.platform).toBe('admin')
        expect(node.app_code).toBe('platform_admin')
        if (node.children) checkNodes(node.children)
      }
    }
    checkNodes(body.data)
  })

  test('TC-D03: 默认租户 timezone/locale/currency 正确', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/admin/tenants`, { headers })
    const body = await resp.json()
    const tenant = body.data.list.find((t: any) => t.tenant_code === 'default')

    expect(tenant).toBeTruthy()
    expect(tenant.timezone).toBe('Asia/Shanghai')
    expect(tenant.locale).toBe('zh-CN')
    expect(tenant.currency).toBe('CNY')
  })

  test('TC-D09: 菜单资源数量 ≥ 15', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/common/user-menu?platform=admin`, { headers })
    const body = await resp.json()

    let count = 0
    function countNodes(nodes: any[]) {
      for (const n of nodes) {
        count++
        if (n.children) countNodes(n.children)
      }
    }
    countNodes(body.data)
    expect(count).toBeGreaterThanOrEqual(15)
  })
})
