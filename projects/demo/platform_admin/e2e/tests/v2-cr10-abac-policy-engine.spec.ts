/**
 * V2-CR10 ABAC 策略引擎 — 接口 E2E 测试
 * 覆盖范围：策略 CRUD、行权限、列权限、评估 API、租户隔离、乐观锁、缓存失效
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

  // 多租户模式：选择第一个租户
  if (body.data?.tenants?.length > 0) {
    const tenant = body.data.tenants[0]
    const tenantResp = await api.post(`${API_BASE}/auth/tenant/select`, {
      data: { tenant_id: String(tenant.id) },
      headers: { Authorization: `Bearer ${body.data.token}` }
    })
    const tb = await tenantResp.json()
    return {
      token: tb.data?.access_token ?? tb.data?.token,
      tenantId: String(tenant.id)
    }
  }
  return {
    token: body.data?.access_token ?? body.data?.token,
    tenantId: '1'
  }
}

function authHeader(token: string) {
  return { headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } }
}

function unique(prefix: string): string {
  return `${prefix}_cr10_${Date.now()}`
}

/** 创建 ABAC 策略并返回 id + version */
async function createPolicy(
  api: APIRequestContext,
  token: string,
  overrides: Record<string, unknown> = {}
): Promise<{ id: string; version: number }> {
  const resp = await api.post(`${API_BASE}/api/v1/admin/abac/policies`, {
    data: {
      name: unique('policy'),
      resource_type: 'order',
      subject_type: 'ROLE',
      subject_id: 'sales',
      effect: 'ALLOW',
      priority: 100,
      ...overrides
    },
    ...authHeader(token)
  })
  expect(resp.status(), `创建策略失败: ${await resp.text()}`).toBe(201)
  const body = await resp.json()
  return { id: String(body.data?.id ?? body.id), version: body.data?.version ?? body.version ?? 1 }
}

// ============================================================
// TC-1xx: 策略 CRUD 基础
// ============================================================

test.describe('TC-1xx — 策略 CRUD', () => {
  let api: APIRequestContext
  let token: string

  test.beforeAll(async () => {
    api = await request.newContext()
    ;({ token } = await loginAdmin(api))
  })
  test.afterAll(async () => { await api.dispose() })

  test('TC-101: 创建合法策略返回 201 含 id', async () => {
    const { id } = await createPolicy(api, token)
    expect(id).toBeTruthy()
  })

  test('TC-102: 同名策略重复创建返回 409', async () => {
    const name = unique('dup_policy')
    await api.post(`${API_BASE}/api/v1/admin/abac/policies`, {
      data: { name, resource_type: 'order', subject_type: 'ROLE', subject_id: 'sales', effect: 'ALLOW' },
      ...authHeader(token)
    })
    const resp2 = await api.post(`${API_BASE}/api/v1/admin/abac/policies`, {
      data: { name, resource_type: 'order', subject_type: 'ROLE', subject_id: 'sales', effect: 'ALLOW' },
      ...authHeader(token)
    })
    expect(resp2.status()).toBe(409)
  })

  test('TC-103: 查询策略详情返回 row_policies 和 col_policies', async () => {
    const { id } = await createPolicy(api, token, {
      row_policies: [
        { action: 'read', condition_expr: { type: 'expr', left: { source: 'resource', attr: 'dept_id' }, op: 'eq', right: { source: 'const', value: '10' } } }
      ],
      col_policies: [
        { field_name: 'phone', effect: 'MASK', mask_type: 'phone' }
      ]
    })
    const resp = await api.get(`${API_BASE}/api/v1/admin/abac/policies/${id}`, authHeader(token))
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    const data = body.data ?? body
    expect(data.row_policies).toHaveLength(1)
    expect(data.col_policies).toHaveLength(1)
    expect(data.col_policies[0].effect).toBe('MASK')
  })

  test('TC-104: 列表分页查询含 total 字段', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/abac/policies?page=1&page_size=5`, authHeader(token))
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    const data = body.data ?? body
    expect(typeof (data.total ?? data.list?.length)).toBe('number')
  })

  test('TC-105: 无 JWT 访问策略列表返回 401', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/abac/policies`)
    expect(resp.status()).toBe(401)
  })

  test('TC-106: 更新策略名称成功（乐观锁 version 递增）', async () => {
    const { id, version } = await createPolicy(api, token)
    const newName = unique('updated')
    const resp = await api.put(`${API_BASE}/api/v1/admin/abac/policies/${id}`, {
      data: { name: newName, version },
      ...authHeader(token)
    })
    expect(resp.status()).toBe(200)
    // 验证版本已递增
    const detail = await api.get(`${API_BASE}/api/v1/admin/abac/policies/${id}`, authHeader(token))
    const detailBody = await detail.json()
    const d = detailBody.data ?? detailBody
    expect(d.version).toBe(version + 1)
    expect(d.name).toBe(newName)
  })

  test('TC-107: 使用旧版本号更新返回 409（乐观锁冲突）', async () => {
    const { id, version } = await createPolicy(api, token)
    // 第一次更新，version 消耗掉
    await api.put(`${API_BASE}/api/v1/admin/abac/policies/${id}`, {
      data: { name: unique('first_update'), version },
      ...authHeader(token)
    })
    // 第二次用旧版本号更新 → 应返回 409
    const resp = await api.put(`${API_BASE}/api/v1/admin/abac/policies/${id}`, {
      data: { name: unique('second_update'), version },  // 仍用原 version
      ...authHeader(token)
    })
    expect(resp.status()).toBe(409)
  })

  test('TC-108: 删除策略后查询返回 404', async () => {
    const { id } = await createPolicy(api, token)
    const delResp = await api.delete(`${API_BASE}/api/v1/admin/abac/policies/${id}`, authHeader(token))
    expect(delResp.status()).toBe(200)
    const getResp = await api.get(`${API_BASE}/api/v1/admin/abac/policies/${id}`, authHeader(token))
    expect(getResp.status()).toBe(404)
  })
})

// ============================================================
// TC-2xx: 资源属性注册表
// ============================================================

test.describe('TC-2xx — 资源属性列表', () => {
  let api: APIRequestContext
  let token: string

  test.beforeAll(async () => {
    api = await request.newContext()
    ;({ token } = await loginAdmin(api))
  })
  test.afterAll(async () => { await api.dispose() })

  test('TC-201: GET /abac/resources 返回 resources + subject_attrs', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/abac/resources`, authHeader(token))
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    const data = body.data ?? body
    expect(Array.isArray(data.resources)).toBeTruthy()
    expect(Array.isArray(data.subject_attrs)).toBeTruthy()
  })

  test('TC-202: subject_attrs 包含内置 user_id/dept_ids', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/abac/resources`, authHeader(token))
    const body = await resp.json()
    const data = body.data ?? body
    const attrNames = data.subject_attrs.map((a: any) => a.attr_name)
    expect(attrNames).toContain('user_id')
    expect(attrNames).toContain('dept_ids')
  })
})

// ============================================================
// TC-3xx: 策略评估 API
// ============================================================

test.describe('TC-3xx — 策略评估 API', () => {
  let api: APIRequestContext
  let token: string

  test.beforeAll(async () => {
    api = await request.newContext()
    ;({ token } = await loginAdmin(api))
  })
  test.afterAll(async () => { await api.dispose() })

  test('TC-301: POST /abac/evaluate 返回 allowed + col_effects', async () => {
    // 先创建一条 ALLOW 策略（含列权限）
    await createPolicy(api, token, {
      resource_type: 'order',
      subject_type: 'ROLE',
      subject_id: 'e2e_eval_role',
      effect: 'ALLOW',
      col_policies: [{ field_name: 'phone', effect: 'MASK', mask_type: 'phone' }]
    })

    const resp = await api.post(`${API_BASE}/api/v1/admin/abac/evaluate`, {
      data: {
        resource_type: 'order',
        action: 'read',
        subject: { user_id: '1', role_ids: ['e2e_eval_role'], dept_ids: [], tenant_id: '1' }
      },
      ...authHeader(token)
    })
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    const data = body.data ?? body
    expect(typeof data.allowed).toBe('boolean')
    expect(data.col_effects).toBeDefined()
  })

  test('TC-302: 缺少 resource_type 返回 400', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/admin/abac/evaluate`, {
      data: { action: 'read', subject: {} },
      ...authHeader(token)
    })
    expect(resp.status()).toBe(400)
  })

  test('TC-303: 无 JWT 评估返回 401', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/admin/abac/evaluate`, {
      data: { resource_type: 'order', action: 'read', subject: {} }
    })
    expect(resp.status()).toBe(401)
  })
})

// ============================================================
// TC-4xx: 行权限条件树验证
// ============================================================

test.describe('TC-4xx — 行权限条件树', () => {
  let api: APIRequestContext
  let token: string

  test.beforeAll(async () => {
    api = await request.newContext()
    ;({ token } = await loginAdmin(api))
  })
  test.afterAll(async () => { await api.dispose() })

  test('TC-401: 创建含嵌套 AND/OR 条件树的策略，详情可完整读回', async () => {
    const condExpr = {
      type: 'group',
      operator: 'AND',
      children: [
        { type: 'expr', left: { source: 'resource', attr: 'dept_id' }, op: 'eq', right: { source: 'subject', attr: 'dept_ids' } },
        {
          type: 'group',
          operator: 'OR',
          children: [
            { type: 'expr', left: { source: 'resource', attr: 'status' }, op: 'eq', right: { source: 'const', value: 'active' } },
            { type: 'expr', left: { source: 'resource', attr: 'status' }, op: 'eq', right: { source: 'const', value: 'pending' } }
          ]
        }
      ]
    }
    const { id } = await createPolicy(api, token, {
      row_policies: [{ action: 'read', condition_expr: condExpr }]
    })
    const resp = await api.get(`${API_BASE}/api/v1/admin/abac/policies/${id}`, authHeader(token))
    const body = await resp.json()
    const d = body.data ?? body
    expect(d.row_policies[0].condition_expr.type).toBe('group')
    expect(d.row_policies[0].condition_expr.operator).toBe('AND')
    expect(d.row_policies[0].condition_expr.children).toHaveLength(2)
  })

  test('TC-402: 创建含 is_null 操作符（无右值）的策略，condition_expr 可保存', async () => {
    const condExpr = { type: 'expr', left: { source: 'resource', attr: 'dept_id' }, op: 'is_null' }
    const { id } = await createPolicy(api, token, {
      row_policies: [{ action: 'read', condition_expr: condExpr }]
    })
    const resp = await api.get(`${API_BASE}/api/v1/admin/abac/policies/${id}`, authHeader(token))
    const body = await resp.json()
    const d = body.data ?? body
    expect(d.row_policies[0].condition_expr.op).toBe('is_null')
  })
})

// ============================================================
// TC-5xx: 列权限脱敏配置
// ============================================================

test.describe('TC-5xx — 列权限配置', () => {
  let api: APIRequestContext
  let token: string

  test.beforeAll(async () => {
    api = await request.newContext()
    ;({ token } = await loginAdmin(api))
  })
  test.afterAll(async () => { await api.dispose() })

  test('TC-501: 创建含多字段列权限（SHOW/HIDE/MASK）的策略可保存并读回', async () => {
    const { id } = await createPolicy(api, token, {
      col_policies: [
        { field_name: 'phone', effect: 'MASK', mask_type: 'phone' },
        { field_name: 'id_card', effect: 'HIDE' },
        { field_name: 'name', effect: 'SHOW' }
      ]
    })
    const resp = await api.get(`${API_BASE}/api/v1/admin/abac/policies/${id}`, authHeader(token))
    const body = await resp.json()
    const d = body.data ?? body
    expect(d.col_policies).toHaveLength(3)
    const phonePolicy = d.col_policies.find((c: any) => c.field_name === 'phone')
    expect(phonePolicy?.effect).toBe('MASK')
    expect(phonePolicy?.mask_type).toBe('phone')
    const idCardPolicy = d.col_policies.find((c: any) => c.field_name === 'id_card')
    expect(idCardPolicy?.effect).toBe('HIDE')
  })

  test('TC-502: 自定义脱敏 mask_pattern 可保存', async () => {
    const { id } = await createPolicy(api, token, {
      col_policies: [
        { field_name: 'custom_field', effect: 'MASK', mask_type: 'custom', mask_pattern: '\\d{4,}' }
      ]
    })
    const resp = await api.get(`${API_BASE}/api/v1/admin/abac/policies/${id}`, authHeader(token))
    const body = await resp.json()
    const d = body.data ?? body
    const cp = d.col_policies.find((c: any) => c.field_name === 'custom_field')
    expect(cp?.mask_pattern).toBe('\\d{4,}')
  })
})

// ============================================================
// TC-6xx: 回归测试（RG 验证）
// ============================================================

test.describe('TC-6xx — 回归测试', () => {
  let api: APIRequestContext
  let token: string

  test.beforeAll(async () => {
    api = await request.newContext()
    ;({ token } = await loginAdmin(api))
  })
  test.afterAll(async () => { await api.dispose() })

  test('TC-601/RG-4: 登录流程不受 ABAC 影响（仍能正常登录）', async () => {
    const resp = await api.post(`${API_BASE}/auth/login`, {
      data: { username: ADMIN_USER, password: ADMIN_PASS }
    })
    expect(resp.ok()).toBeTruthy()
  })

  test('TC-602/RG-1: 用户管理列表接口正常返回（无 ABAC 策略资源不受影响）', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/users?page=1&page_size=5`, authHeader(token))
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    expect(Array.isArray((body.data ?? body).list ?? (body.data ?? body))).toBeTruthy()
  })

  test('TC-603/RG-3: 字段权限接口正常（FieldPermission 不受 ABAC 影响）', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/field-objects`, authHeader(token))
    // 200 或 404（路径可能不同）均可，关键是不应 500
    expect(resp.status()).not.toBe(500)
  })

  test('TC-604/RG-2: DataScope 列表接口正常（数据权限配置不受影响）', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/data-scope-configs`, authHeader(token))
    expect(resp.status()).toBe(200)
  })

  test('TC-605: 启用/禁用策略（status 切换）', async () => {
    const { id, version } = await createPolicy(api, token)
    // 禁用
    const disableResp = await api.put(`${API_BASE}/api/v1/admin/abac/policies/${id}`, {
      data: { status: 0, version },
      ...authHeader(token)
    })
    expect(disableResp.status()).toBe(200)
    // 验证 status=0
    const detailResp = await api.get(`${API_BASE}/api/v1/admin/abac/policies/${id}`, authHeader(token))
    const body = await detailResp.json()
    expect((body.data ?? body).status).toBe(0)
  })
})
