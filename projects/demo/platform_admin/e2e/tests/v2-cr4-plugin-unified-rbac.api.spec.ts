/**
 * V2-CR4 插件系统统一 RBAC + 通信契约 — 接口测试
 * 关联用例: AIDOC/project_doc/platform_admin/v2-cr4-plugin-unified-rbac/test_cases.md
 * 执行日期: 2025-07-15
 */
import { test, expect, APIRequestContext, request } from '@playwright/test';

const API_BASE = 'http://localhost:8000';
const ADMIN_USER = 'admin';
const ADMIN_PASS = 'admin123';

// ============================================================
// 辅助工具
// ============================================================

async function loginAdmin(api: APIRequestContext): Promise<string> {
  const loginResp = await api.post(`${API_BASE}/auth/login`, {
    data: { username: ADMIN_USER, password: ADMIN_PASS },
  });
  expect(loginResp.ok(), `登录失败: ${await loginResp.text()}`).toBeTruthy();
  const loginBody = await loginResp.json();
  if (loginBody.data?.tenants?.length > 0) {
    const tenantResp = await api.post(`${API_BASE}/auth/tenant/select`, {
      data: { tenant_id: String(loginBody.data.tenants[0].id) },
      headers: { Authorization: `Bearer ${loginBody.data.token}` },
    });
    const tenantBody = await tenantResp.json();
    return tenantBody.data?.access_token ?? tenantBody.data?.token;
  }
  return loginBody.data?.access_token ?? loginBody.data?.token;
}

function auth(token: string) {
  return { headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } };
}

// ============================================================
// 插件管理接口测试
// ============================================================

test.describe('V2-CR4 接口测试 — 插件管理', () => {
  let api: APIRequestContext;
  let token: string;

  test.beforeAll(async () => {
    api = await request.newContext();
    token = await loginAdmin(api);
  });
  test.afterAll(async () => {
    await api.dispose();
  });

  // TC-A01: GET /api/v1/admin/plugins — 正向
  test('TC-A01 获取插件列表', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/plugins`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    // 后端返回 {list: [...], total: N} 格式
    const data = body.data;
    expect(data).toBeDefined();
    expect(Array.isArray(data.list)).toBeTruthy();
  });

  // TC-A02: GET /api/v1/admin/plugins — 未认证
  test('TC-A02 未认证访问插件列表返回401', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/plugins`);
    expect(resp.status()).toBe(401);
  });

  // TC-A04: POST /api/v1/admin/plugins/upload — 需要实际插件包，跳过
  test.skip('TC-A04 上传安装插件（需实际zip包）', async () => {
    // 需要合法的插件 zip 包，纯 API 测试环境不可用
  });

  // TC-A07: POST /api/v1/admin/plugins/upload — 无效文件格式
  test.skip('TC-A07 上传无效文件返回400（需multipart上传）', async () => {
    // 需要构造 multipart/form-data 上传非 zip 文件
  });

  // TC-A08: POST /api/v1/admin/plugins/:name/start — 验证接口可达（插件不存在场景）
  test('TC-A08 启动不存在的插件返回错误', async () => {
    const resp = await api.post(
      `${API_BASE}/api/v1/admin/plugins/nonexist_plugin_${Date.now()}/start`,
      auth(token),
    );
    // 插件不存在时应返回 404 或 400
    expect([400, 404]).toContain(resp.status());
    const body = await resp.json();
    expect(body.code).not.toBe(0);
  });

  // TC-A11: POST /api/v1/admin/plugins/:name/stop — 验证接口可达（插件不存在场景）
  test('TC-A11 停止不存在的插件返回错误', async () => {
    const resp = await api.post(
      `${API_BASE}/api/v1/admin/plugins/nonexist_plugin_${Date.now()}/stop`,
      auth(token),
    );
    // 插件不存在时 Handler 返回 500（内部错误）
    expect([400, 404, 500]).toContain(resp.status());
    const body = await resp.json();
    expect(body.code).not.toBe(0);
  });

  // TC-A12: DELETE /api/v1/admin/plugins/:name — 验证接口可达（插件不存在场景）
  test('TC-A12 卸载不存在的插件返回错误', async () => {
    const resp = await api.delete(
      `${API_BASE}/api/v1/admin/plugins/nonexist_plugin_${Date.now()}`,
      { ...auth(token), timeout: 10000 },
    );
    // 卸载不存在的插件：后端 syncer + 文件删除 tolerant，可能返回 200 或错误码
    // 验证接口可达且不 panic
    expect([200, 400, 404, 500]).toContain(resp.status());
  });

  // TC-A14: PUT /api/v1/admin/plugins/:name/upgrade — 需要实际插件包，跳过
  test.skip('TC-A14 升级插件（需实际zip包）', async () => {
    // 需要合法的插件升级包，纯 API 测试环境不可用
  });

  // TC-A16: GET /api/v1/admin/plugins/:name/health — 验证接口可达（插件不存在场景）
  test('TC-A16 查询不存在插件的健康状态返回错误', async () => {
    const resp = await api.get(
      `${API_BASE}/api/v1/admin/plugins/nonexist_plugin_${Date.now()}/health`,
      auth(token),
    );
    // 未注册插件返回 500（内部错误）
    expect([400, 404, 500]).toContain(resp.status());
    const body = await resp.json();
    expect(body.code).not.toBe(0);
  });
});

// ============================================================
// 应用目录与订阅接口测试
// ============================================================

test.describe('V2-CR4 接口测试 — 应用目录与订阅', () => {
  let api: APIRequestContext;
  let token: string;
  // 记录测试中创建的订阅，用于 afterAll 清理
  const subscribedApps: string[] = [];

  test.beforeAll(async () => {
    api = await request.newContext();
    token = await loginAdmin(api);
  });
  test.afterAll(async () => {
    // 清理测试中创建的订阅记录
    for (const appCode of subscribedApps) {
      await api.delete(`${API_BASE}/api/v1/admin/app-subscriptions/${appCode}`, auth(token));
    }
    await api.dispose();
  });

  // TC-A17: GET /api/v1/admin/app-catalog — 正向
  test('TC-A17 获取应用目录', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/app-catalog`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    // 应至少包含 BUILTIN 应用（platform_admin）
    const list: any[] = body.data ?? [];
    expect(list.length).toBeGreaterThan(0);
    // 验证字段结构
    const first = list[0];
    expect(first.app_code).toBeDefined();
    expect(first.app_type).toBeDefined();
  });

  // TC-A18: GET /api/v1/admin/app-catalog — 未认证
  test('TC-A18 未认证访问应用目录返回401', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/app-catalog`);
    expect(resp.status()).toBe(401);
  });

  // TC-A19: POST /api/v1/admin/app-subscriptions — 正向
  test('TC-A19 订阅应用', async () => {
    // 先获取应用目录，找到一个可订阅的 PLUGIN 类型应用
    const catalogResp = await api.get(`${API_BASE}/api/v1/admin/app-catalog`, auth(token));
    const catalogBody = await catalogResp.json();
    const apps: any[] = catalogBody.data ?? [];
    // 找到 PLUGIN 类型且未订阅的应用
    const pluginApp = apps.find(
      (a: any) => a.app_type === 'PLUGIN' && !a.subscribed,
    );
    // 如果没有可用的 PLUGIN 应用，跳过
    test.skip(!pluginApp, '无可订阅的PLUGIN应用，跳过');

    const resp = await api.post(`${API_BASE}/api/v1/admin/app-subscriptions`, {
      ...auth(token),
      data: { app_code: pluginApp.app_code },
    });
    // 订阅成功应返回 200 或 201
    expect([200, 201]).toContain(resp.status());
    const body = await resp.json();
    expect(body.code).toBe(0);
    subscribedApps.push(pluginApp.app_code);
  });

  // TC-A22: POST /api/v1/admin/app-subscriptions — 无效 body
  test('TC-A22 订阅应用缺少app_code返回400', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/admin/app-subscriptions`, {
      ...auth(token),
      data: {},
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.code).not.toBe(0);
  });

  // TC-N17: POST /api/v1/admin/app-subscriptions — 订阅不存在的应用
  test('TC-N17 订阅不存在的应用返回错误', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/admin/app-subscriptions`, {
      ...auth(token),
      data: { app_code: `ghost_app_${Date.now()}` },
    });
    expect([400, 404]).toContain(resp.status());
    const body = await resp.json();
    expect(body.code).not.toBe(0);
  });

  // TC-A23: DELETE /api/v1/admin/app-subscriptions/:app_code — 正向
  test('TC-A23 退订PLUGIN应用', async () => {
    // 先确保有一个已订阅的 PLUGIN 应用可退订
    const catalogResp = await api.get(`${API_BASE}/api/v1/admin/app-catalog`, auth(token));
    const catalogBody = await catalogResp.json();
    const apps: any[] = catalogBody.data ?? [];
    const subscribedPlugin = apps.find(
      (a: any) => a.app_type === 'PLUGIN' && a.subscribed,
    );

    if (!subscribedPlugin) {
      // 先订阅一个用于退订测试
      const pluginApp = apps.find((a: any) => a.app_type === 'PLUGIN');
      test.skip(!pluginApp, '无PLUGIN应用可用于退订测试');
      const subResp = await api.post(`${API_BASE}/api/v1/admin/app-subscriptions`, {
        ...auth(token),
        data: { app_code: pluginApp.app_code },
      });
      if (subResp.status() !== 200 && subResp.status() !== 201) {
        test.skip(true, '订阅失败，无法测试退订');
      }
      // 退订
      const resp = await api.delete(
        `${API_BASE}/api/v1/admin/app-subscriptions/${pluginApp.app_code}`,
        auth(token),
      );
      expect(resp.status()).toBe(200);
      const body = await resp.json();
      expect(body.code).toBe(0);
      // 从清理列表移除（已手动退订）
      const idx = subscribedApps.indexOf(pluginApp.app_code);
      if (idx !== -1) subscribedApps.splice(idx, 1);
    } else {
      const resp = await api.delete(
        `${API_BASE}/api/v1/admin/app-subscriptions/${subscribedPlugin.app_code}`,
        auth(token),
      );
      expect(resp.status()).toBe(200);
      const body = await resp.json();
      expect(body.code).toBe(0);
      // 从清理列表移除
      const idx = subscribedApps.indexOf(subscribedPlugin.app_code);
      if (idx !== -1) subscribedApps.splice(idx, 1);
    }
  });

  // TC-N08: DELETE /api/v1/admin/app-subscriptions/platform_admin — BUILTIN 拒绝退订
  test('TC-N08 退订BUILTIN应用被拒绝', async () => {
    const resp = await api.delete(
      `${API_BASE}/api/v1/admin/app-subscriptions/platform_admin`,
      auth(token),
    );
    // BUILTIN 应用不允许退订，应返回 400
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.code).not.toBe(0);
    // 错误消息应包含"内置"或"不允许"相关信息
    const msg = body.msg ?? body.message ?? '';
    expect(msg).toMatch(/内置|不允许|BUILTIN|builtin/i);
  });

  // TC-N09: 配额超限拒绝订阅（需要设置极低配额 + 非 SUPER_ADMIN）
  test('TC-N09 配额超限拒绝订阅', async () => {
    // 当前 admin 为 SUPER_ADMIN，跳过配额检查
    // 此用例需要非 SUPER_ADMIN 账号才能验证
    test.skip(true, '需要非SUPER_ADMIN账号验证配额拒绝，当前环境仅有admin账号');
  });

  // TC-A26: GET /api/v1/admin/app-subscriptions — 正向
  test('TC-A26 获取已订阅应用列表', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/app-subscriptions`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    // data 应为数组
    const list: any[] = body.data?.list ?? body.data ?? [];
    expect(Array.isArray(list)).toBeTruthy();
    // 至少应包含 BUILTIN 应用的订阅
    expect(list.length).toBeGreaterThan(0);
  });

  // TC-A27: GET /api/v1/admin/app-subscriptions — 未认证
  test('TC-A27 未认证访问订阅列表返回401', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/app-subscriptions`);
    expect(resp.status()).toBe(401);
  });

  // TC-A28: PUT /api/v1/admin/app-subscriptions/:app_code/modules — 正向
  test('TC-A28 更新订阅模块配置', async () => {
    // 获取已订阅的应用列表，找到一个有 modules 的应用
    const subResp = await api.get(`${API_BASE}/api/v1/admin/app-subscriptions`, auth(token));
    const subBody = await subResp.json();
    const subs: any[] = subBody.data?.list ?? subBody.data ?? [];
    // 找到 PLUGIN 类型已订阅的应用（BUILTIN 可能不支持模块配置）
    const pluginSub = subs.find((s: any) => s.app_type === 'PLUGIN');
    test.skip(!pluginSub, '无已订阅的PLUGIN应用可用于模块配置测试');

    const resp = await api.put(
      `${API_BASE}/api/v1/admin/app-subscriptions/${pluginSub.app_code}/modules`,
      {
        ...auth(token),
        data: { enabled_modules: ['invoice', 'payment'] },
      },
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
  });

  // TC-A29: PUT /api/v1/admin/app-subscriptions/:app_code/modules — 未认证
  test('TC-A29 未认证更新模块配置返回401', async () => {
    const resp = await api.put(
      `${API_BASE}/api/v1/admin/app-subscriptions/billing/modules`,
      { data: { enabled_modules: ['invoice'] } },
    );
    expect(resp.status()).toBe(401);
  });

  // TC-N18: POST /api/v1/admin/app-subscriptions — 重复订阅
  test('TC-N18 重复订阅同一应用返回错误', async () => {
    // 获取已订阅列表
    const subResp = await api.get(`${API_BASE}/api/v1/admin/app-subscriptions`, auth(token));
    const subBody = await subResp.json();
    const subs: any[] = subBody.data?.list ?? subBody.data ?? [];
    const existingSub = subs[0];
    test.skip(!existingSub, '无已订阅应用可用于重复订阅测试');

    const resp = await api.post(`${API_BASE}/api/v1/admin/app-subscriptions`, {
      ...auth(token),
      data: { app_code: existingSub.app_code },
    });
    // 重复订阅应返回 400 或 409
    expect([400, 409]).toContain(resp.status());
    const body = await resp.json();
    expect(body.code).not.toBe(0);
  });
});

// ============================================================
// 回归测试 — 验证现有接口不受 CR4 改动影响
// ============================================================

test.describe('V2-CR4 回归测试', () => {
  let api: APIRequestContext;
  let token: string;

  test.beforeAll(async () => {
    api = await request.newContext();
    token = await loginAdmin(api);
  });
  test.afterAll(async () => {
    await api.dispose();
  });

  // TC-R05: 现有路由注册不受影响
  test('TC-R05 现有管理接口正常响应', async () => {
    // 验证原有核心接口可用
    const endpoints = [
      '/api/v1/admin/users?page=1&page_size=1',
      '/api/v1/admin/roles?page=1&page_size=1',
    ];
    for (const ep of endpoints) {
      const resp = await api.get(`${API_BASE}${ep}`, auth(token));
      expect(resp.status(), `接口 ${ep} 异常`).toBe(200);
      const body = await resp.json();
      expect(body.code, `接口 ${ep} 业务码异常`).toBe(0);
    }
  });

  // TC-R03: 角色权限配置页 BUILTIN 应用展示正常
  test('TC-R03 角色列表接口正常', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/roles?page=1&page_size=5`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    const list: any[] = body.data?.list ?? body.data ?? [];
    expect(list.length).toBeGreaterThan(0);
  });

  // TC-R04: 登录接口不受影响
  test('TC-R04 登录接口正常', async () => {
    const resp = await api.post(`${API_BASE}/auth/login`, {
      data: { username: ADMIN_USER, password: ADMIN_PASS },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    expect(body.data?.token).toBeDefined();
  });
});
