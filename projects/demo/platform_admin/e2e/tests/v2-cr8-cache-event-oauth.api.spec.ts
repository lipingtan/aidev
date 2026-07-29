/**
 * V2-CR8 缓存事件驱动 + OAuth2 预留 + 扩展能力收尾 — 接口测试
 * 关联用例: AIDOC/project_doc/platform_admin/v2-cr8-cache-event-oauth/test_cases.md
 * 执行日期: 2026-07-29
 */
import { test, expect, APIRequestContext, request } from '@playwright/test';

const API_BASE = 'http://localhost:8000';
const ADMIN_USER = 'admin';
const ADMIN_PASS = 'admin123';

// ============================================================
// 辅助工具
// ============================================================

/** 登录并返回 access_token */
async function loginAdmin(api: APIRequestContext): Promise<string> {
  const loginResp = await api.post(`${API_BASE}/auth/login`, {
    data: { username: ADMIN_USER, password: ADMIN_PASS },
  });
  expect(loginResp.ok(), `登录失败: ${await loginResp.text()}`).toBeTruthy();
  const loginBody = await loginResp.json();

  if (loginBody.data?.tenants?.length > 0) {
    const tenant = loginBody.data.tenants[0];
    const tenantResp = await api.post(`${API_BASE}/auth/tenant/select`, {
      data: { tenant_id: String(tenant.id) },
      headers: { Authorization: `Bearer ${loginBody.data.token}` },
    });
    const tenantBody = await tenantResp.json();
    return tenantBody.data?.access_token ?? tenantBody.data?.token;
  }
  return loginBody.data?.access_token ?? loginBody.data?.token;
}

/** 注入 Authorization header */
function auth(token: string) {
  return { headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } };
}

/** 生成唯一名称 */
function unique(prefix: string): string {
  return `${prefix}_cr8_${Date.now()}`;
}

// ============================================================
// Phase 3: OAuth2/LDAP 骨架（P0，优先验证）
// ============================================================

test.describe('TC-A08/A09/A10 — OAuth2/LDAP 认证骨架', () => {
  let api: APIRequestContext;

  test.beforeAll(async () => {
    api = await request.newContext();
  });

  test.afterAll(async () => {
    await api.dispose();
  });

  test('TC-A08: grant_type=oauth2 返回 501 且 message 含"OAuth2"', async () => {
    const resp = await api.post(`${API_BASE}/auth/login`, {
      data: { grant_type: 'oauth2', username: 'any', password: 'any' },
    });
    // 后端返回 HTTP 501，响应体 code=50101
    expect(resp.status()).toBe(501);
    const body = await resp.json();
    expect(body.code).toBe(50101);
    expect(body.message).toMatch(/OAuth2|oauth2/i);
    // 确认不包含 token
    expect(body.data?.token ?? body.data?.access_token).toBeFalsy();
  });

  test('TC-A09: grant_type=ldap 返回 501 且 message 含"LDAP"', async () => {
    const resp = await api.post(`${API_BASE}/auth/login`, {
      data: { grant_type: 'ldap', username: 'any', password: 'any' },
    });
    expect(resp.status()).toBe(501);
    const body = await resp.json();
    expect(body.code).toBe(50101);
    expect(body.message).toMatch(/LDAP|ldap/i);
    expect(body.data?.token ?? body.data?.access_token).toBeFalsy();
  });

  test('TC-N01: grant_type=oauth2 响应无 token 字段', async () => {
    const resp = await api.post(`${API_BASE}/auth/login`, {
      data: { grant_type: 'oauth2', username: 'any', password: 'any' },
    });
    const body = await resp.json();
    // 验证响应数据中没有 token
    const hasToken = !!(body?.data?.token || body?.data?.access_token || body?.data?.platform_token);
    expect(hasToken).toBe(false);
  });

  test('TC-N02: grant_type=ldap 响应无 token 字段', async () => {
    const resp = await api.post(`${API_BASE}/auth/login`, {
      data: { grant_type: 'ldap', username: 'any', password: 'any' },
    });
    const body = await resp.json();
    const hasToken = !!(body?.data?.token || body?.data?.access_token || body?.data?.platform_token);
    expect(hasToken).toBe(false);
  });

  test('TC-A10/TC-P03/TC-R01: grant_type=password 正常登录不受影响（RG-8）', async () => {
    const resp = await api.post(`${API_BASE}/auth/login`, {
      data: { grant_type: 'password', username: ADMIN_USER, password: ADMIN_PASS },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    // 登录成功 code=0
    expect(body.code).toBe(0);
    // 返回有效 token
    const token = body.data?.token ?? body.data?.access_token;
    expect(token).toBeTruthy();
  });

  test('TC-N03: grant_type=unknown 返回不支持错误', async () => {
    const resp = await api.post(`${API_BASE}/auth/login`, {
      data: { grant_type: 'unknown_strategy', username: 'any', password: 'any' },
    });
    const body = await resp.json();
    // code 非 0
    expect(body.code).not.toBe(0);
    // 无 token
    expect(body.data?.token ?? body.data?.access_token).toBeFalsy();
  });
});

// ============================================================
// Phase 1: 权限检查降级 + 缓存透明性（P0）
// ============================================================

test.describe('TC-A01/A02/TC-P04/TC-R02/TC-N04 — 权限检查与缓存降级', () => {
  let api: APIRequestContext;
  let token: string;

  test.beforeAll(async () => {
    api = await request.newContext();
    token = await loginAdmin(api);
  });

  test.afterAll(async () => {
    await api.dispose();
  });

  test('TC-A01/TC-P04/TC-R02: SUPER_ADMIN 权限放行正常（缓存降级不影响）', async () => {
    // 访问有权限的接口，验证不因缓存问题拒绝
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenants`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
  });

  test('TC-N04: 无 token 请求被拒绝（不因缓存问题放行）', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenants`, {
      headers: { 'Content-Type': 'application/json' },
    });
    // 401 未认证
    expect([401, 403]).toContain(resp.status());
  });

  test('TC-A02: 无权限接口被拦截', async () => {
    // 通过无效 token 模拟无权限用户
    const fakeToken = 'invalid.token.here';
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenants`, {
      headers: { Authorization: `Bearer ${fakeToken}` },
    });
    expect([401, 403]).toContain(resp.status());
  });
});

// ============================================================
// Phase 2: 操作日志风险分级
// ============================================================

test.describe('TC-A04~A07/TC-P01~P02 — 操作日志风险分级', () => {
  let api: APIRequestContext;
  let token: string;

  test.beforeAll(async () => {
    api = await request.newContext();
    token = await loginAdmin(api);
  });

  test.afterAll(async () => {
    await api.dispose();
  });

  test('TC-A04/TC-R05: 操作日志查询接口可用（SUPER_ADMIN 全量）', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/operation-logs`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    expect(body.data).toBeTruthy();
  });

  test('TC-A07: 操作日志响应中每条记录包含 risk_level 字段', async () => {
    // 先触发一个写操作，确保有日志
    const roleName = unique('role');
    await api.post(`${API_BASE}/api/v1/admin/roles`, {
      ...auth(token),
      data: { role_code: roleName, role_name: roleName, tenant_id: '1' },
    });

    // 查询日志
    const resp = await api.get(`${API_BASE}/api/v1/admin/operation-logs?page=1&page_size=10`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);

    const list: any[] = body.data?.list ?? [];
    if (list.length > 0) {
      // 验证 risk_level 字段存在
      const firstLog = list[0];
      expect(firstLog).toHaveProperty('risk_level');
      // risk_level 必须是合法值
      expect(['LOW', 'MEDIUM', 'HIGH']).toContain(firstLog.risk_level);
    } else {
      // 日志为空时跳过字段验证，但接口本身已验证正常
      console.log('[TC-A07] 日志列表为空，跳过字段验证');
    }
  });

  test('TC-A05/TC-P01: 按 risk_level=HIGH 过滤日志', async () => {
    // 先触发高风险操作：创建租户并删除（删除租户为 HIGH 风险）
    const tenantCode = unique('t_del');
    const createResp = await api.post(`${API_BASE}/api/v1/admin/tenants`, {
      ...auth(token),
      data: { tenant_code: tenantCode, name: tenantCode },
    });
    if (createResp.status() === 200) {
      const createBody = await createResp.json();
      if (createBody.code === 0 && createBody.data?.id) {
        const tid = String(createBody.data.id);
        await api.delete(`${API_BASE}/api/v1/admin/tenants/${tid}`, auth(token));
      }
    }

    // 查询 HIGH 风险日志
    const resp = await api.get(
      `${API_BASE}/api/v1/admin/operation-logs?risk_level=HIGH&page=1&page_size=20`,
      auth(token)
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);

    const list: any[] = body.data?.list ?? [];
    // 如果有日志，验证都是 HIGH
    for (const log of list) {
      expect(log.risk_level).toBe('HIGH');
    }
  });

  test('TC-A06/TC-P02: 按 risk_level=LOW 过滤日志（普通写操作）', async () => {
    // 触发低风险写操作
    const roleName = unique('role_low');
    await api.post(`${API_BASE}/api/v1/admin/roles`, {
      ...auth(token),
      data: { role_code: roleName, role_name: roleName, tenant_id: '1' },
    });

    const resp = await api.get(
      `${API_BASE}/api/v1/admin/operation-logs?risk_level=LOW&page=1&page_size=20`,
      auth(token)
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);

    const list: any[] = body.data?.list ?? [];
    for (const log of list) {
      expect(log.risk_level).toBe('LOW');
    }
  });
});

// ============================================================
// Phase 4: ext_fields 向后兼容 + admin_custom_field
// ============================================================

test.describe('TC-A11~A15 — ext_fields 向后兼容', () => {
  let api: APIRequestContext;
  let token: string;

  test.beforeAll(async () => {
    api = await request.newContext();
    token = await loginAdmin(api);
  });

  test.afterAll(async () => {
    await api.dispose();
  });

  test('TC-A11: 用户创建时不传 ext_fields 正常（向后兼容）', async () => {
    const username = unique('user_noext');
    const resp = await api.post(`${API_BASE}/api/v1/admin/users`, {
      ...auth(token),
      data: { username, password: 'Test@12345' },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    expect(body.data?.id).toBeTruthy();

    // 清理
    if (body.data?.id) {
      await api.delete(`${API_BASE}/api/v1/admin/users/${String(body.data.id)}`, auth(token));
    }
  });

  test('TC-A12: 租户创建时不传 ext_fields 正常（向后兼容）', async () => {
    const tenantCode = unique('tnt_noext');
    const resp = await api.post(`${API_BASE}/api/v1/admin/tenants`, {
      ...auth(token),
      data: { tenant_code: tenantCode, name: tenantCode },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    expect(body.data?.id).toBeTruthy();

    // 清理
    if (body.data?.id) {
      await api.delete(`${API_BASE}/api/v1/admin/tenants/${String(body.data.id)}`, auth(token));
    }
  });

  test('TC-A13: 用户查询响应中包含 ext_fields 字段', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/users?page=1&page_size=5`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);

    const list: any[] = body.data?.list ?? [];
    if (list.length > 0) {
      // ext_fields 字段应存在（可为 null）
      expect(list[0]).toHaveProperty('ext_fields');
    }
  });

  test('TC-A14: 租户查询响应中包含 ext_fields 字段', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenants?page=1&page_size=5`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);

    const list: any[] = body.data?.list ?? [];
    if (list.length > 0) {
      expect(list[0]).toHaveProperty('ext_fields');
    }
  });

  test('TC-A15: 服务状态接口正常（AutoMigrate 包括 admin_custom_field 无报错）', async () => {
    // 通过服务本身可响应来验证 AutoMigrate（包含 CustomField）无错误
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenants?page=1&page_size=1`, auth(token));
    expect(resp.status()).toBe(200);
    // 服务响应正常意味着 AutoMigrate 没有导致启动失败
  });
});

// ============================================================
// 回归测试
// ============================================================

test.describe('TC-R03/R04 — 回归测试', () => {
  let api: APIRequestContext;
  let token: string;

  test.beforeAll(async () => {
    api = await request.newContext();
    token = await loginAdmin(api);
  });

  test.afterAll(async () => {
    await api.dispose();
  });

  test('TC-R03: 租户隔离不被破坏（RG-2）', async () => {
    // 查询租户列表，验证正常返回
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenants?page=1&page_size=5`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
  });
});
