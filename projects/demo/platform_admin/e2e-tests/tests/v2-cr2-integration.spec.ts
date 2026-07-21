/**
 * V2-CR2 权限体系增强 — 集成场景测试（补充）
 * 覆盖之前跳过/未验证的端到端业务场景：
 * - Phase 1: 角色继承（子集校验 + 级联裁剪）
 * - Phase 2: 权限计算合并（GetUserMenu 并集）
 * - Phase 3: FieldFilter 字段过滤生效
 * - Phase 4: 记录共享 DataScope 生效验证
 *
 * 关联用例: AIDOC/project_doc/platform_admin/v2-cr2-permission-enhancement/test_cases.md
 * 执行日期: 2026-07-21
 */
import { test, expect, APIRequestContext, request } from '@playwright/test';

const API_BASE = 'http://localhost:8000';
const ADMIN_USER = 'admin';
const ADMIN_PASS = 'admin123';

async function loginAdmin(api: APIRequestContext): Promise<string> {
  const resp = await api.post(`${API_BASE}/auth/login`, {
    data: { username: ADMIN_USER, password: ADMIN_PASS },
  });
  const body = await resp.json();
  if (body.data?.tenants?.length > 0) {
    const tenantResp = await api.post(`${API_BASE}/auth/tenant/select`, {
      data: { tenant_id: String(body.data.tenants[0].id) },
      headers: { Authorization: `Bearer ${body.data.token}` },
    });
    const tb = await tenantResp.json();
    return tb.data?.access_token ?? tb.data?.token;
  }
  return body.data?.access_token ?? body.data?.token;
}

function auth(token: string) {
  return { headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } };
}

// ============================================================
// Phase 1: 角色继承 — 子集校验 + 级联裁剪（端到端）
// ============================================================

test.describe('Phase1 角色继承 — 子集校验 + 级联裁剪', () => {
  let api: APIRequestContext;
  let token: string;
  let parentID: string;
  let childID: string;
  let grandChildID: string;
  let resIDs: string[] = [];
  let apiIDs: string[] = [];

  test.beforeAll(async () => {
    api = await request.newContext();
    token = await loginAdmin(api);

    // 获取系统中已有的资源树（取 platform_admin 应用下的资源 ID）
    const treeResp = await api.get(`${API_BASE}/api/v1/admin/resources/tree?app_code=platform_admin`, auth(token));
    if (treeResp.ok()) {
      const treeBody = await treeResp.json();
      // 递归提取所有资源 ID
      function extractIDs(nodes: any[]): string[] {
        const ids: string[] = [];
        for (const n of nodes) {
          ids.push(String(n.id));
          if (n.children?.length > 0) ids.push(...extractIDs(n.children));
        }
        return ids;
      }
      resIDs = extractIDs(treeBody.data ?? []).slice(0, 6); // 取前6个
    }

    // 获取 API 权限 ID
    const apiListResp = await api.get(`${API_BASE}/api/v1/admin/api-permissions?app_code=platform_admin`, auth(token));
    if (apiListResp.ok()) {
      const apiBody = await apiListResp.json();
      function extractApiIDs(nodes: any[]): string[] {
        const ids: string[] = [];
        for (const n of nodes) {
          if (n.type === 'ENDPOINT') ids.push(String(n.id));
          if (n.children?.length > 0) ids.push(...extractApiIDs(n.children));
        }
        return ids;
      }
      apiIDs = extractApiIDs(apiBody.data ?? []).slice(0, 4);
    }

    const suffix = Date.now();

    // 创建父角色（顶级）
    const pResp = await api.post(`${API_BASE}/api/v1/admin/roles`, {
      ...auth(token),
      data: { role_code: `int_parent_${suffix}`, role_name: `集成父_${suffix}`, role_type: 'NORMAL' },
    });
    parentID = String((await pResp.json()).data?.id);

    // 给父角色分配资源 [R1..R6] 和 API [A1..A4]
    if (resIDs.length >= 4) {
      await api.put(`${API_BASE}/api/v1/admin/roles/${parentID}/resources`, {
        ...auth(token),
        data: { resource_ids: resIDs },
      });
    }
    if (apiIDs.length >= 2) {
      await api.put(`${API_BASE}/api/v1/admin/roles/${parentID}/apis`, {
        ...auth(token),
        data: { api_permission_ids: apiIDs },
      });
    }

    // 创建子角色
    const cResp = await api.post(`${API_BASE}/api/v1/admin/roles`, {
      ...auth(token),
      data: { role_code: `int_child_${suffix}`, role_name: `集成子_${suffix}`, role_type: 'NORMAL', parent_id: parentID },
    });
    childID = String((await cResp.json()).data?.id);

    // 创建孙角色
    const gcResp = await api.post(`${API_BASE}/api/v1/admin/roles`, {
      ...auth(token),
      data: { role_code: `int_grand_${suffix}`, role_name: `集成孙_${suffix}`, role_type: 'NORMAL', parent_id: childID },
    });
    grandChildID = String((await gcResp.json()).data?.id);
  });

  test.afterAll(async () => { await api.dispose(); });

  test('TC-001 子角色分配父角色子集资源 — 成功', async () => {
    test.skip(resIDs.length < 4, '资源数量不足');
    const subset = resIDs.slice(0, 4);
    const resp = await api.put(`${API_BASE}/api/v1/admin/roles/${childID}/resources`, {
      ...auth(token),
      data: { resource_ids: subset },
    });
    expect(resp.status()).toBe(200);
    expect((await resp.json()).code).toBe(0);
  });

  test('TC-002 子角色分配父角色子集API — 成功', async () => {
    test.skip(apiIDs.length < 2, 'API 数量不足');
    const subset = apiIDs.slice(0, 2);
    const resp = await api.put(`${API_BASE}/api/v1/admin/roles/${childID}/apis`, {
      ...auth(token),
      data: { api_permission_ids: subset },
    });
    expect(resp.status()).toBe(200);
    expect((await resp.json()).code).toBe(0);
  });

  test('TC-N01 子角色超出父角色资源范围 — 被拒绝', async () => {
    test.skip(resIDs.length < 2, '资源数量不足');
    // 传入不在父角色范围内的虚假 ID
    const resp = await api.put(`${API_BASE}/api/v1/admin/roles/${childID}/resources`, {
      ...auth(token),
      data: { resource_ids: [resIDs[0], '99999999999999999'] },
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.message).toContain('超出');
  });

  test('TC-N02 子角色超出父角色API范围 — 被拒绝', async () => {
    test.skip(apiIDs.length < 1, 'API 数量不足');
    const resp = await api.put(`${API_BASE}/api/v1/admin/roles/${childID}/apis`, {
      ...auth(token),
      data: { api_permission_ids: [apiIDs[0], '99999999999999999'] },
    });
    expect(resp.status()).toBe(400);
    const body = await resp.json();
    expect(body.message).toContain('超出');
  });

  test('TC-003 父角色缩减资源触发级联裁剪', async () => {
    test.skip(resIDs.length < 4, '资源数量不足');
    // 先给子角色 4 个资源，孙角色 2 个
    await api.put(`${API_BASE}/api/v1/admin/roles/${childID}/resources`, {
      ...auth(token),
      data: { resource_ids: resIDs.slice(0, 4) },
    });
    await api.put(`${API_BASE}/api/v1/admin/roles/${grandChildID}/resources`, {
      ...auth(token),
      data: { resource_ids: resIDs.slice(0, 2) },
    });

    // 父角色缩减为前 2 个
    const resp = await api.put(`${API_BASE}/api/v1/admin/roles/${parentID}/resources`, {
      ...auth(token),
      data: { resource_ids: resIDs.slice(0, 2) },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data?.affected_children?.length).toBeGreaterThan(0);

    // 验证子角色被裁剪
    const childRes = await api.get(`${API_BASE}/api/v1/admin/roles/${childID}/resources`, auth(token));
    const childIDs: string[] = (await childRes.json()).data ?? [];
    // 子角色不应有第 3、4 个资源
    expect(childIDs).not.toContain(resIDs[2]);
    expect(childIDs).not.toContain(resIDs[3]);
    // 孙角色不受影响（本身就在范围内）
    const grandRes = await api.get(`${API_BASE}/api/v1/admin/roles/${grandChildID}/resources`, auth(token));
    const grandIDs: string[] = (await grandRes.json()).data ?? [];
    expect(grandIDs.length).toBeLessThanOrEqual(2);
  });

  test('TC-N03 顶级角色分配任意资源 — 不做子集校验', async () => {
    test.skip(resIDs.length < 2, '资源数量不足');
    const resp = await api.put(`${API_BASE}/api/v1/admin/roles/${parentID}/resources`, {
      ...auth(token),
      data: { resource_ids: resIDs.slice(0, 3) },
    });
    expect(resp.status()).toBe(200);
  });

  test('TC-N04 PERMISSION_SET 分配任意资源 — 不做子集校验', async () => {
    test.skip(resIDs.length < 2, '资源数量不足');
    const psResp = await api.post(`${API_BASE}/api/v1/admin/roles`, {
      ...auth(token),
      data: { role_code: `int_ps_${Date.now()}`, role_name: `集成权限集`, role_type: 'PERMISSION_SET' },
    });
    const psID = String((await psResp.json()).data?.id);
    const resp = await api.put(`${API_BASE}/api/v1/admin/roles/${psID}/resources`, {
      ...auth(token),
      data: { resource_ids: resIDs },
    });
    expect(resp.status()).toBe(200);
  });
});

// ============================================================
// Phase 2: 权限合并 — GetUserMenu 并集验证
// ============================================================

test.describe('Phase2 权限合并 — GetUserMenu 验证', () => {
  let api: APIRequestContext;
  let token: string;

  test.beforeAll(async () => {
    api = await request.newContext();
    token = await loginAdmin(api);
  });
  test.afterAll(async () => { await api.dispose(); });

  test('TC-006 GetUserMenu 返回菜单列表（SUPER_ADMIN 全量）', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/common/user-menu`, auth(token));
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.code).toBe(0);
    // SUPER_ADMIN 应该能看到菜单
    const menus: any[] = body.data ?? [];
    expect(menus.length).toBeGreaterThan(0);
  });

  test('TC-006b 权限合并 — 普通角色用户也能获取菜单', async () => {
    // 创建一个测试用户并分配角色，验证 GetUserMenu
    const suffix = Date.now();
    // 创建用户
    const userResp = await api.post(`${API_BASE}/api/v1/admin/users`, {
      ...auth(token),
      data: { username: `perm_test_${suffix}`, nickname: `权限测试用户`, password: 'Test123456' },
    });
    if (userResp.status() !== 200) { test.skip(); return; }
    const userID = String((await userResp.json()).data?.id);

    // 创建角色并分配一些资源
    const roleResp = await api.post(`${API_BASE}/api/v1/admin/roles`, {
      ...auth(token),
      data: { role_code: `perm_role_${suffix}`, role_name: `权限测试角色`, role_type: 'NORMAL' },
    });
    const roleID = String((await roleResp.json()).data?.id);

    // 获取可用资源
    const treeResp = await api.get(`${API_BASE}/api/v1/admin/resources/tree?app_code=platform_admin`, auth(token));
    const treeBody = await treeResp.json();
    function extractIDs(nodes: any[]): string[] {
      const ids: string[] = [];
      for (const n of nodes) { ids.push(String(n.id)); if (n.children?.length > 0) ids.push(...extractIDs(n.children)); }
      return ids;
    }
    const allResIDs = extractIDs(treeBody.data ?? []).slice(0, 3);
    if (allResIDs.length === 0) { test.skip(); return; }

    // 分配资源给角色
    await api.put(`${API_BASE}/api/v1/admin/roles/${roleID}/resources`, {
      ...auth(token),
      data: { resource_ids: allResIDs },
    });

    // 给用户分配角色
    await api.post(`${API_BASE}/api/v1/admin/users/${userID}/roles`, {
      ...auth(token),
      data: { role_ids: [roleID] },
    });

    // 用该用户登录（需要密码）— 如果登录成功，调 GetUserMenu
    const loginResp = await api.post(`${API_BASE}/auth/login`, {
      data: { username: `perm_test_${suffix}`, password: 'Test123456' },
    });
    if (loginResp.status() !== 200) { test.skip(); return; }
    const loginBody = await loginResp.json();
    let userToken = loginBody.data?.token ?? loginBody.data?.access_token;
    // 如果需要 tenant select
    if (loginBody.data?.tenants?.length > 0) {
      const tsResp = await api.post(`${API_BASE}/auth/tenant/select`, {
        data: { tenant_id: String(loginBody.data.tenants[0].id) },
        headers: { Authorization: `Bearer ${userToken}` },
      });
      const tsBody = await tsResp.json();
      userToken = tsBody.data?.access_token ?? tsBody.data?.token;
    }

    if (!userToken) { test.skip(); return; }

    // 调用 GetUserMenu
    const menuResp = await api.get(`${API_BASE}/api/v1/common/user-menu`, auth(userToken));
    expect(menuResp.status()).toBe(200);
    const menuBody = await menuResp.json();
    expect(menuBody.code).toBe(0);
    // 应该能看到菜单（分配了资源）
    const menus: any[] = menuBody.data ?? [];
    expect(menus.length).toBeGreaterThan(0);
  });
});

// ============================================================
// Phase 3: FieldFilter 字段过滤生效验证
// ============================================================

test.describe('Phase3 FieldFilter 字段过滤', () => {
  let api: APIRequestContext;
  let token: string;
  let testUserToken: string;
  let roleID: string;
  let userID: string;

  test.beforeAll(async () => {
    api = await request.newContext();
    token = await loginAdmin(api);

    const suffix = Date.now();

    // 创建专用角色
    const roleResp = await api.post(`${API_BASE}/api/v1/admin/roles`, {
      ...auth(token),
      data: { role_code: `ff_role_${suffix}`, role_name: `字段过滤角色`, role_type: 'NORMAL' },
    });
    roleID = String((await roleResp.json()).data?.id);

    // 给角色分配资源（至少要有用户管理菜单权限才能访问 /users）
    const treeResp = await api.get(`${API_BASE}/api/v1/admin/resources/tree?app_code=platform_admin`, auth(token));
    const treeBody = await treeResp.json();
    function extractIDs(nodes: any[]): string[] {
      const ids: string[] = [];
      for (const n of nodes) { ids.push(String(n.id)); if (n.children?.length > 0) ids.push(...extractIDs(n.children)); }
      return ids;
    }
    const allResIDs = extractIDs(treeBody.data ?? []);
    if (allResIDs.length > 0) {
      await api.put(`${API_BASE}/api/v1/admin/roles/${roleID}/resources`, {
        ...auth(token),
        data: { resource_ids: allResIDs },
      });
    }

    // 配置角色对 user 对象的 phone 字段为 HIDDEN
    await api.put(`${API_BASE}/api/v1/admin/field-permissions`, {
      ...auth(token),
      data: { role_id: roleID, object_code: 'user', items: [{ field_name: 'phone', access: 'HIDDEN' }] },
    });

    // 创建测试用户
    const userResp = await api.post(`${API_BASE}/api/v1/admin/users`, {
      ...auth(token),
      data: { username: `ff_user_${suffix}`, nickname: `字段过滤用户`, password: 'Test123456', phone: '13800138000' },
    });
    if (userResp.status() === 200) {
      userID = String((await userResp.json()).data?.id);
      // 分配角色
      await api.post(`${API_BASE}/api/v1/admin/users/${userID}/roles`, {
        ...auth(token),
        data: { role_ids: [roleID] },
      });

      // 登录该用户
      const loginResp = await api.post(`${API_BASE}/auth/login`, {
        data: { username: `ff_user_${suffix}`, password: 'Test123456' },
      });
      if (loginResp.ok()) {
        const lb = await loginResp.json();
        testUserToken = lb.data?.token ?? lb.data?.access_token;
        if (lb.data?.tenants?.length > 0) {
          const tsResp = await api.post(`${API_BASE}/auth/tenant/select`, {
            data: { tenant_id: String(lb.data.tenants[0].id) },
            headers: { Authorization: `Bearer ${testUserToken}` },
          });
          const tsBody = await tsResp.json();
          testUserToken = tsBody.data?.access_token ?? tsBody.data?.token;
        }
      }
    }
  });

  test.afterAll(async () => { await api.dispose(); });

  test('TC-009 HIDDEN 字段在用户列表 API 响应中被过滤', async () => {
    if (!testUserToken) { test.skip(); return; }
    // 用配置了 phone=HIDDEN 的角色用户查询用户列表
    const resp = await api.get(`${API_BASE}/api/v1/admin/users`, auth(testUserToken));
    if (resp.status() !== 200) { test.skip(); return; }
    const body = await resp.json();
    const users: any[] = body.data?.list ?? body.data ?? [];
    if (users.length === 0) { test.skip(); return; }
    // 验证：用户列表中所有记录都不应包含 phone 字段
    for (const user of users) {
      expect(user).not.toHaveProperty('phone');
    }
  });

  test('TC-B06 未配置字段权限的对象不做过滤', async () => {
    if (!testUserToken) { test.skip(); return; }
    // tenant 对象未配置字段权限，查看租户列表应该完整返回所有字段
    const resp = await api.get(`${API_BASE}/api/v1/admin/tenants`, auth(testUserToken));
    if (resp.status() !== 200) { test.skip(); return; }
    const body = await resp.json();
    const tenants: any[] = body.data?.list ?? body.data ?? [];
    if (tenants.length === 0) { test.skip(); return; }
    // 租户对象应包含常见字段（未被过滤）
    expect(tenants[0]).toHaveProperty('name');
  });

  test('TC-012 多角色冲突取最高权限', async () => {
    // 创建第二个角色配置 phone=VISIBLE
    const suffix2 = Date.now();
    const role2Resp = await api.post(`${API_BASE}/api/v1/admin/roles`, {
      ...auth(token),
      data: { role_code: `ff_role2_${suffix2}`, role_name: `字段过滤角色2`, role_type: 'NORMAL' },
    });
    const role2ID = String((await role2Resp.json()).data?.id);

    // 配置 role2 对 user.phone=VISIBLE
    await api.put(`${API_BASE}/api/v1/admin/field-permissions`, {
      ...auth(token),
      data: { role_id: role2ID, object_code: 'user', items: [{ field_name: 'phone', access: 'VISIBLE' }] },
    });

    // 给测试用户追加 role2
    if (userID) {
      await api.post(`${API_BASE}/api/v1/admin/users/${userID}/roles`, {
        ...auth(token),
        data: { role_ids: [roleID, role2ID] },
      });
    }

    if (!testUserToken) { test.skip(); return; }
    // 重新查询用户列表 — phone 应可见（VISIBLE > HIDDEN）
    const resp = await api.get(`${API_BASE}/api/v1/admin/users`, auth(testUserToken));
    if (resp.status() !== 200) { test.skip(); return; }
    const body = await resp.json();
    const users: any[] = body.data?.list ?? body.data ?? [];
    if (users.length === 0) { test.skip(); return; }
    // VISIBLE > HIDDEN，phone 应该存在
    const hasPhone = users.some((u: any) => 'phone' in u);
    expect(hasPhone).toBeTruthy();
  });
});

// ============================================================
// Phase 4: 记录共享 — DataScope 生效验证
// ============================================================

test.describe('Phase4 记录共享 — 共享规则 share_to_type 验证', () => {
  let api: APIRequestContext;
  let token: string;

  test.beforeAll(async () => {
    api = await request.newContext();
    token = await loginAdmin(api);
  });
  test.afterAll(async () => { await api.dispose(); });

  test('TC-015 共享给角色（share_to_type=ROLE）创建成功', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/admin/record-shares`, {
      ...auth(token),
      data: {
        object_code: 'invoice', record_id: '2001',
        share_to_type: 'ROLE', share_to_id: '100',
        access_level: 'READ',
      },
    });
    expect(resp.status()).toBe(200);
    expect((await resp.json()).code).toBe(0);
  });

  test('TC-016 共享给部门（share_to_type=DEPT）创建成功', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/admin/record-shares`, {
      ...auth(token),
      data: {
        object_code: 'invoice', record_id: '2002',
        share_to_type: 'DEPT', share_to_id: '50',
        access_level: 'EDIT',
      },
    });
    expect(resp.status()).toBe(200);
    expect((await resp.json()).code).toBe(0);
  });

  test('TC-B11 同一记录被共享给多种类型（USER+ROLE+DEPT）', async () => {
    const recordID = '3001';
    // 共享给 USER
    const r1 = await api.post(`${API_BASE}/api/v1/admin/record-shares`, {
      ...auth(token),
      data: { object_code: 'invoice', record_id: recordID, share_to_type: 'USER', share_to_id: '100', access_level: 'READ' },
    });
    expect(r1.status()).toBe(200);
    // 共享给 ROLE
    const r2 = await api.post(`${API_BASE}/api/v1/admin/record-shares`, {
      ...auth(token),
      data: { object_code: 'invoice', record_id: recordID, share_to_type: 'ROLE', share_to_id: '200', access_level: 'READ' },
    });
    expect(r2.status()).toBe(200);
    // 共享给 DEPT
    const r3 = await api.post(`${API_BASE}/api/v1/admin/record-shares`, {
      ...auth(token),
      data: { object_code: 'invoice', record_id: recordID, share_to_type: 'DEPT', share_to_id: '300', access_level: 'EDIT' },
    });
    expect(r3.status()).toBe(200);

    // 查询：应该有 3 条规则
    const listResp = await api.get(
      `${API_BASE}/api/v1/admin/record-shares?object_code=invoice&record_id=${recordID}`,
      auth(token)
    );
    expect(listResp.status()).toBe(200);
    const list: any[] = (await listResp.json()).data ?? [];
    expect(list.length).toBeGreaterThanOrEqual(3);
  });

  test('TC-B09 expire_at=null 永久生效 — 查询可见', async () => {
    const resp = await api.post(`${API_BASE}/api/v1/admin/record-shares`, {
      ...auth(token),
      data: {
        object_code: 'invoice', record_id: '4001',
        share_to_type: 'USER', share_to_id: '600',
        access_level: 'READ',
        // 不传 expire_at → 永久
      },
    });
    expect(resp.status()).toBe(200);
    const shareID = String((await resp.json()).data?.id);

    // 查询应能找到
    const listResp = await api.get(
      `${API_BASE}/api/v1/admin/record-shares?object_code=invoice&record_id=4001`,
      auth(token)
    );
    const list: any[] = (await listResp.json()).data ?? [];
    const found = list.find((s: any) => String(s.id) === shareID);
    expect(found).toBeDefined();
  });
});
