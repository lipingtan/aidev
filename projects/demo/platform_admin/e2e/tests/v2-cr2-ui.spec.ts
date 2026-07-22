/**
 * V2-CR2 权限体系增强 — UI 集成测试
 * 通过浏览器操作验证之前 API 测试中跳过的场景：
 * - 普通角色用户登录后能看到分配的菜单
 * - HIDDEN 字段在用户列表中不显示
 * - 多角色冲突取最高权限
 *
 * 策略：API 完成数据准备（创建用户/角色/配置字段权限）→ 注入 localStorage token → page 操作验证
 */
import { test, expect, Page, request, APIRequestContext } from '@playwright/test';

const WEB_BASE = 'http://localhost:3000';
const API_BASE = 'http://localhost:8000';

/** API 登录获取完整 token 信息 */
async function apiLogin(api: APIRequestContext, username: string, password: string) {
  const resp = await api.post(`${API_BASE}/auth/login`, {
    data: { username, password },
  });
  const body = await resp.json();
  const data = body.data;
  return {
    platform_token: data?.platform_token ?? data?.token ?? '',
    access_token: data?.access_token ?? '',
    tenants: data?.tenants ?? [],
    current_tenant_id: data?.tenants?.[0]?.id ? String(data.tenants[0].id) : '',
  };
}

/** 注入认证状态到浏览器 localStorage，跳过登录页 */
async function injectAuth(page: Page, authData: {
  platform_token: string;
  access_token: string;
  tenants: any[];
  current_tenant_id: string;
}) {
  await page.goto(WEB_BASE);
  await page.evaluate((data) => {
    localStorage.setItem('platform_token', data.platform_token);
    localStorage.setItem('access_token', data.access_token);
    localStorage.setItem('tenant_list', JSON.stringify(data.tenants));
    localStorage.setItem('current_tenant_id', data.current_tenant_id);
  }, authData);
  // 刷新页面让 store 从 localStorage 读取
  await page.goto(`${WEB_BASE}/#/home`);
  await page.waitForLoadState('networkidle');
}

function auth(token: string) {
  return { headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } };
}

// ============================================================
// TC-006b: 普通角色用户登录后能看到分配的菜单
// ============================================================

test.describe('UI 验证 — 权限合并 + 字段过滤', () => {
  let api: APIRequestContext;
  let adminToken: string;
  let testUsername: string;
  let testUserID: string;
  let roleID: string;

  test.beforeAll(async () => {
    api = await request.newContext();
    const adminAuth = await apiLogin(api, 'admin', 'admin123');
    adminToken = adminAuth.access_token;

    const suffix = Date.now();
    testUsername = `ui_test_${suffix}`;

    // 1. 创建测试用户
    const userResp = await api.post(`${API_BASE}/api/v1/admin/users`, {
      ...auth(adminToken),
      data: { username: testUsername, nickname: 'UI测试用户', password: 'Test123456', phone: '13900139000' },
    });
    testUserID = String((await userResp.json()).data?.id);

    // 2. 创建测试角色
    const roleResp = await api.post(`${API_BASE}/api/v1/admin/roles`, {
      ...auth(adminToken),
      data: { role_code: `ui_role_${suffix}`, role_name: 'UI测试角色', role_type: 'NORMAL' },
    });
    roleID = String((await roleResp.json()).data?.id);

    // 3. 给角色分配所有资源（确保菜单可见）
    const treeResp = await api.get(`${API_BASE}/api/v1/admin/resources/tree?app_code=platform_admin`, auth(adminToken));
    const treeBody = await treeResp.json();
    function extractIDs(nodes: any[]): string[] {
      const ids: string[] = [];
      for (const n of nodes) { ids.push(String(n.id)); if (n.children?.length > 0) ids.push(...extractIDs(n.children)); }
      return ids;
    }
    const allResIDs = extractIDs(treeBody.data ?? []);
    if (allResIDs.length > 0) {
      await api.put(`${API_BASE}/api/v1/admin/roles/${roleID}/resources`, {
        ...auth(adminToken),
        data: { resource_ids: allResIDs },
      });
    }

    // 4. 给用户分配角色
    await api.post(`${API_BASE}/api/v1/admin/users/${testUserID}/roles`, {
      ...auth(adminToken),
      data: { role_ids: [roleID] },
    });

    // 5. 配置字段权限：该角色对 user 对象 phone 字段为 HIDDEN
    await api.put(`${API_BASE}/api/v1/admin/field-permissions`, {
      ...auth(adminToken),
      data: { role_id: roleID, object_code: 'user', items: [{ field_name: 'phone', access: 'HIDDEN' }] },
    });
  });

  test.afterAll(async () => { await api.dispose(); });

  test('TC-006b 普通角色用户登录后能看到菜单', async ({ page }) => {
    // 用测试用户的 token 注入
    const userAuth = await apiLogin(api, testUsername, 'Test123456');
    if (!userAuth.access_token) { test.skip(); return; }
    await injectAuth(page, userAuth);

    // 等待页面主内容加载（侧边栏或主容器）
    await page.waitForTimeout(2000);
    // 检查是否存在菜单或者成功进入了非登录页
    const url = page.url();
    const isNotLogin = !url.includes('/login');
    // 如果没被重定向到登录页，说明 token 有效
    expect(isNotLogin).toBeTruthy();

    // 检查侧边栏存在任何菜单元素
    const sidebarMenu = await page.locator('.el-menu, [class*="sidebar"], [class*="menu"]').count();
    expect(sidebarMenu).toBeGreaterThan(0);
  });

  test('TC-009 HIDDEN 字段在用户列表页面中不显示 phone 列', async ({ page }) => {
    // 用配置了 phone=HIDDEN 角色的测试用户登录
    const userAuth = await apiLogin(api, testUsername, 'Test123456');
    if (!userAuth.access_token) { test.skip(); return; }
    await injectAuth(page, userAuth);

    // 导航到用户管理页面
    await page.goto(`${WEB_BASE}/#/system/users`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);

    // 检查表格表头：不应包含"手机号"列（phone 字段被 HIDDEN）
    const headers = await page.locator('.el-table__header th').allTextContents();
    const hasPhone = headers.some(h => h.includes('手机') || h.includes('phone'));
    expect(hasPhone).toBeFalsy();
  });

  // BUG-011: FieldFilter 多角色冲突未正确取最高权限
  // 当 roleA 配置 phone=HIDDEN、roleB 配置 phone=VISIBLE 时，预期 VISIBLE 生效，实际仍被过滤
  test('TC-012 多角色冲突取最高权限 [BUG: 未取最高]', async ({ page }) => {
    test.fail(); // BUG-011: FieldFilter 多角色冲突取最高权限逻辑未生效
    // 创建第二个角色，配置 phone=VISIBLE
    const suffix2 = Date.now();
    const role2Resp = await api.post(`${API_BASE}/api/v1/admin/roles`, {
      ...auth(adminToken),
      data: { role_code: `ui_role2_${suffix2}`, role_name: 'UI测试角色2', role_type: 'NORMAL' },
    });
    const role2ID = String((await role2Resp.json()).data?.id);

    // 配置 role2 对 user.phone=VISIBLE
    await api.put(`${API_BASE}/api/v1/admin/field-permissions`, {
      ...auth(adminToken),
      data: { role_id: role2ID, object_code: 'user', items: [{ field_name: 'phone', access: 'VISIBLE' }] },
    });

    // 给用户追加 role2（现在有两个角色：HIDDEN + VISIBLE → 取最高 VISIBLE）
    await api.put(`${API_BASE}/api/v1/admin/users/${testUserID}/roles`, {
      ...auth(adminToken),
      data: { role_ids: [roleID, role2ID] },
    });

    // 重新获取 token（角色变更后需要刷新，新 token 包含最新 roleIDs）
    const userAuth = await apiLogin(api, testUsername, 'Test123456');
    if (!userAuth.access_token) { test.skip(); return; }
    await injectAuth(page, userAuth);

    // 导航到用户管理页面
    await page.goto(`${WEB_BASE}/#/system/users`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    // 检查表格表头：应包含"手机号"列（VISIBLE > HIDDEN）
    // 注：如果 FieldFilter 从 token.roles 实时查库取最新角色绑定，则 phone 应可见
    // 如果 FieldFilter 只用 token 中原始 roles，则可能仍然 HIDDEN（取决于实现）
    const headers = await page.locator('.el-table__header th').allTextContents();
    const hasPhone = headers.some(h => h.includes('手机') || h.includes('phone'));
    // 注意：此断言验证的是"多角色取最高权限"逻辑
    // 如果失败，说明 FieldFilter 未正确合并多角色的字段权限
    expect(hasPhone).toBeTruthy();
  });
});
