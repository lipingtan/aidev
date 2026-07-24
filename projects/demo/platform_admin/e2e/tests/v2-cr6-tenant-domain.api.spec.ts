/**
 * V2-CR6 域名-租户映射管理 — API 接口测试
 * 关联用例: test_cases.md (TC-001 ~ TC-013, TC-N01 ~ TC-N09, TC-B01 ~ TC-B04, TC-R01 ~ TC-R05)
 */
import { test, expect, APIRequestContext, request } from '@playwright/test';

const API_BASE = 'http://localhost:8000';
const ADMIN_USER = 'admin';
const ADMIN_PASS = 'admin123';

// ─── 辅助函数 ───

async function loginAdmin(api: APIRequestContext): Promise<{ token: string; tenantId: string }> {
  const resp = await api.post(`${API_BASE}/auth/login`, {
    data: { username: ADMIN_USER, password: ADMIN_PASS },
  });
  const body = await resp.json();
  if (body.data?.tenants?.length > 0) {
    const tenant = body.data.tenants[0];
    const selResp = await api.post(`${API_BASE}/auth/tenant/select`, {
      data: { tenant_id: String(tenant.id) },
      headers: { Authorization: `Bearer ${body.data.token}` },
    });
    const selBody = await selResp.json();
    return { token: selBody.data?.access_token ?? selBody.data?.token, tenantId: String(tenant.id) };
  }
  return { token: body.data?.access_token ?? body.data?.token, tenantId: '1' };
}

function auth(token: string) {
  return { headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } };
}

function uniqueDomain(prefix = 'test'): string {
  return `${prefix}-${Date.now()}.example.com`;
}

// ============================================================
// 公开接口测试
// ============================================================

test.describe('公开查询接口 — /api/v1/public/tenant-domain', () => {
  let api: APIRequestContext;

  test.beforeAll(async () => { api = await request.newContext(); });
  test.afterAll(async () => { await api.dispose(); });

  // TC-007: 无需认证即可访问
  test('TC-007 无需认证即可访问', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/public/tenant-domain?domain=anything.com`);
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    expect(body.data).toHaveProperty('tenant_code');
    expect(body.data).toHaveProperty('tenant_name');
    expect(body.data).toHaveProperty('matched');
  });

  // TC-006: 未配置域名返回 default
  test('TC-006 未配置域名返回 default', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/public/tenant-domain?domain=nonexistent-${Date.now()}.xyz`);
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    expect(body.data.tenant_code).toBe('default');
    expect(body.data.matched).toBe(false);
  });

  // TC-N06: 缺少 domain 参数返回 400
  test('TC-N06 缺少 domain 参数返回 400', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/public/tenant-domain`);
    expect(resp.status()).toBe(400);
  });

  // TC-005: 已配置域名返回正确租户（依赖 CRUD 先创建）
  test('TC-005 已配置域名返回 matched=true', async () => {
    // 先登录创建一条映射
    const { token, tenantId } = await loginAdmin(api);
    const domain = uniqueDomain('pub');
    const createResp = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, {
      ...auth(token),
      data: { domain, tenant_id: tenantId },
    });
    expect(createResp.status()).toBe(200);

    // 公开查询
    const resp = await api.get(`${API_BASE}/api/v1/public/tenant-domain?domain=${domain}`);
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data.matched).toBe(true);
    expect(body.data.tenant_code).toBeTruthy();

    // 清理
    const createBody = await createResp.json();
    await api.delete(`${API_BASE}/api/v1/admin/tenant-domains/${String(createBody.data?.id)}`, auth(token));
  });
});

// ============================================================
// 管理端 CRUD 接口测试
// ============================================================

test.describe('管理端 CRUD — /api/v1/admin/tenant-domains', () => {
  let api: APIRequestContext;
  let token: string;
  let tenantId: string;
  const createdIds: string[] = [];

  test.beforeAll(async () => {
    api = await request.newContext();
    const info = await loginAdmin(api);
    token = info.token;
    tenantId = info.tenantId;
  });
  test.afterAll(async () => {
    // 清理所有测试创建的记录
    for (const id of createdIds) {
      await api.delete(`${API_BASE}/api/v1/admin/tenant-domains/${id}`, auth(token)).catch(() => {});
    }
    await api.dispose();
  });

  // TC-N08: 未认证返回 401
  test('TC-N08 未认证返回 401', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenant-domains`);
    expect(resp.status()).toBe(401);
  });

  // TC-001: 创建成功
  test('TC-001 创建域名映射成功', async () => {
    const domain = uniqueDomain('cr');
    const resp = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, {
      ...auth(token),
      data: { domain, tenant_id: tenantId, remark: '测试创建' },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    expect(body.data.domain).toBe(domain);
    expect(body.data.version).toBe(1);
    expect(typeof body.data.id).toBe('string');
    createdIds.push(String(body.data.id));
  });

  // TC-002: 查询列表
  test('TC-002 查询列表（分页+搜索）', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenant-domains?page=1&page_size=10`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    expect(Array.isArray(body.data.list)).toBeTruthy();
    expect(typeof body.data.total).toBe('number');
  });

  // TC-013: 列表含 tenant_code/tenant_name
  test('TC-013 列表含 tenant_code 和 tenant_name', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenant-domains?page=1&page_size=10`, auth(token));
    const body = await resp.json();
    if (body.data.list.length > 0) {
      const item = body.data.list[0];
      expect(item).toHaveProperty('tenant_code');
      expect(item).toHaveProperty('tenant_name');
    }
  });

  // TC-003: 更新成功
  test('TC-003 更新域名映射成功', async () => {
    // 先创建
    const domain = uniqueDomain('upd');
    const createResp = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, {
      ...auth(token),
      data: { domain, tenant_id: tenantId },
    });
    const createBody = await createResp.json();
    const id = String(createBody.data.id);
    createdIds.push(id);

    // 更新
    const newDomain = uniqueDomain('upd-new');
    const resp = await api.put(`${API_BASE}/api/v1/admin/tenant-domains/${id}`, {
      ...auth(token),
      data: { domain: newDomain, version: 1 },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data.domain).toBe(newDomain);
    expect(body.data.version).toBe(2);
  });

  // TC-004: 删除成功
  test('TC-004 删除域名映射成功', async () => {
    const domain = uniqueDomain('del');
    const createResp = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, {
      ...auth(token),
      data: { domain, tenant_id: tenantId },
    });
    const id = String((await createResp.json()).data.id);

    const resp = await api.delete(`${API_BASE}/api/v1/admin/tenant-domains/${id}`, auth(token));
    expect(resp.status()).toBe(200);
  });

  // TC-N01: 重复域名返回 409
  test('TC-N01 重复域名返回冲突', async () => {
    const domain = uniqueDomain('dup');
    const r1 = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, {
      ...auth(token),
      data: { domain, tenant_id: tenantId },
    });
    createdIds.push(String((await r1.json()).data.id));

    const r2 = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, {
      ...auth(token),
      data: { domain, tenant_id: tenantId },
    });
    expect([400, 409]).toContain(r2.status());
  });

  // TC-N02: 保留域名返回 400
  test('TC-N02 保留域名（localhost）返回 400', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, {
      ...auth(token),
      data: { domain: 'localhost', tenant_id: tenantId },
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.message).toMatch(/保留域名/);
  });

  // TC-N03: 超长域名返回 400
  test('TC-N03 超长域名（256字符）返回 400', async () => {
    const longDomain = 'a'.repeat(256) + '.com';
    const resp = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, {
      ...auth(token),
      data: { domain: longDomain, tenant_id: tenantId },
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.message).toMatch(/255/);
  });

  // TC-B01: 255 字符域名创建成功
  test('TC-B01 域名恰好 255 字符创建成功', async () => {
    const domain255 = 'a'.repeat(243) + '.example.com'; // 243 + 12 = 255
    const resp = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, {
      ...auth(token),
      data: { domain: domain255, tenant_id: tenantId },
    });
    expect(resp.status()).toBe(200);
    createdIds.push(String((await resp.json()).data.id));
  });

  // TC-B03: 同一租户绑定多个域名
  test('TC-B03 同一租户绑定多个域名', async () => {
    const d1 = uniqueDomain('multi1');
    const d2 = uniqueDomain('multi2');
    const r1 = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, { ...auth(token), data: { domain: d1, tenant_id: tenantId } });
    const r2 = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, { ...auth(token), data: { domain: d2, tenant_id: tenantId } });
    expect(r1.status()).toBe(200);
    expect(r2.status()).toBe(200);
    createdIds.push(String((await r1.json()).data.id));
    createdIds.push(String((await r2.json()).data.id));
  });

  // TC-N05: version 不匹配返回冲突
  test('TC-N05 更新 version 不匹配返回冲突', async () => {
    const domain = uniqueDomain('ver');
    const createResp = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, { ...auth(token), data: { domain, tenant_id: tenantId } });
    const id = String((await createResp.json()).data.id);
    createdIds.push(id);

    // 用错误的 version
    const resp = await api.put(`${API_BASE}/api/v1/admin/tenant-domains/${id}`, {
      ...auth(token),
      data: { domain: 'changed.com', version: 99 },
    });
    expect([400, 409]).toContain(resp.status());
  });

  // TC-N07: tenant_id 不存在返回错误
  test('TC-N07 tenant_id 不存在返回 404', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, {
      ...auth(token),
      data: { domain: uniqueDomain('notenant'), tenant_id: '9999999999999' },
    });
    expect([400, 404]).toContain(resp.status());
  });

  // TC-N09: 删除不存在的 ID
  test('TC-N09 删除不存在的 ID 返回 404', async () => {
    const resp = await api.delete(`${API_BASE}/api/v1/admin/tenant-domains/9999999999999`, auth(token));
    expect([400, 404]).toContain(resp.status());
  });
});

// ============================================================
// 缓存测试
// ============================================================

test.describe('缓存行为', () => {
  let api: APIRequestContext;
  let token: string;
  let tenantId: string;

  test.beforeAll(async () => {
    api = await request.newContext();
    const info = await loginAdmin(api);
    token = info.token;
    tenantId = info.tenantId;
  });
  test.afterAll(async () => { await api.dispose(); });

  // TC-009: 更新后缓存失效
  test('TC-009 更新后查询返回新值', async () => {
    const domain = uniqueDomain('cache');
    // 创建
    const createResp = await api.post(`${API_BASE}/api/v1/admin/tenant-domains`, {
      ...auth(token),
      data: { domain, tenant_id: tenantId },
    });
    const id = String((await createResp.json()).data.id);

    // 查一次（填充缓存）
    await api.get(`${API_BASE}/api/v1/public/tenant-domain?domain=${domain}`);

    // 删除（缓存失效）
    await api.delete(`${API_BASE}/api/v1/admin/tenant-domains/${id}`, auth(token));

    // 再查（应返回 default）
    const resp = await api.get(`${API_BASE}/api/v1/public/tenant-domain?domain=${domain}`);
    const body = await resp.json();
    expect(body.data.matched).toBe(false);
    expect(body.data.tenant_code).toBe('default');
  });
});

// ============================================================
// 回归测试
// ============================================================

test.describe('回归测试', () => {
  let api: APIRequestContext;

  test.beforeAll(async () => { api = await request.newContext(); });
  test.afterAll(async () => { await api.dispose(); });

  // TC-R01: 管理端登录不受影响
  test('TC-R01 管理端登录正常', async () => {
    const resp = await api.post(`${API_BASE}/auth/login`, {
      data: { username: ADMIN_USER, password: ADMIN_PASS },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
  });

  // TC-R02: C端认证不受影响
  test('TC-R02 C端 send-code 正常', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/user/auth/send-code`, {
      data: { phone: `136${Date.now().toString().slice(-8)}`, tenant_code: 'default' },
    });
    expect(resp.status()).toBe(200);
  });

  // TC-R04: 其他公开接口不受影响
  test('TC-R04 公开接口 send-code 不是 404', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/user/auth/send-code`, {
      data: { phone: '13800000000', tenant_code: 'default' },
    });
    expect(resp.status()).not.toBe(404);
  });
});
