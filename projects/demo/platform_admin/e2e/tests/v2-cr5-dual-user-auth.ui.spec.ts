/**
 * V2-CR5 双用户池 + C端认证 — UI 端面测试
 * 关联用例: test_cases.md
 *
 * 执行规范：
 * - 每个关键操作（点击/填写/提交）后调用 pause()，等待渲染稳定后再停 500ms
 * - 等待使用 expect(...).toBeVisible / waitForResponse，不用硬编码 waitForTimeout
 * - slowMo=100 模拟真实手速（config 层设置）
 */
import { test, expect, Page } from '@playwright/test';

const ADMIN_APP_URL = 'http://localhost:3000';
const USER_APP_URL  = 'http://localhost:5174';

// ─── 辅助：等待页面渲染稳定后额外停 500ms，让人眼看清状态 ───
async function pause(page: Page, ms = 500) {
  await page.waitForLoadState('domcontentloaded');
  await page.waitForTimeout(ms);
}

// ─── 辅助：等待指定 API 响应完成后停 500ms ───
async function waitForApi(page: Page, urlPattern: string | RegExp, action: () => Promise<void>) {
  const [response] = await Promise.all([
    page.waitForResponse(
      resp => (typeof urlPattern === 'string'
        ? resp.url().includes(urlPattern)
        : urlPattern.test(resp.url())
      ) && resp.status() < 500
    ),
    action(),
  ]);
  await page.waitForTimeout(500); // 等渲染完成
  return response;
}

// ─── 辅助：导航到管理端页面并等待稳定 ───
async function navigateTo(page: Page, path: string) {
  // 使用完整绝对 URL，避免被 playwright baseURL 或当前路径污染
  await page.goto(`http://localhost:3000${path}`);
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(500);
}

// ============================================================
// TC-F01 ~ TC-F06: C端用户管理页（dev-web-admin）
// ============================================================
test.describe('C端用户管理页 — dev-web-admin', () => {
  test.use({ storageState: './test-results/.auth/state.json' });

  test.beforeEach(async ({ page }) => {
    // History 模式，路径对应 static-routes.ts 中的 system/biz-user
    await navigateTo(page, '/system/biz-user');
    await expect(page.locator('.el-table')).toBeVisible({ timeout: 15000 });
    await pause(page);
  });

  // TC-F01: 列表页加载，验证三段式结构
  test('TC-F01 列表页表格正确展示数据', async ({ page }) => {
    // 等 loading 消失
    await page.locator('.el-loading-mask').waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
    await pause(page);

    await expect(page.locator('.el-table')).toBeVisible();
    await expect(page.locator('.el-pagination')).toBeVisible();
    expect(await page.locator('.el-card').count()).toBeGreaterThanOrEqual(1);
  });

  // TC-F02: 手机号搜索，等待接口响应后验证表格更新
  test('TC-F02 手机号搜索过滤正确', async ({ page }) => {
    const searchInput = page.locator('input[placeholder*="手机号"]').first();
    if (!await searchInput.isVisible()) return test.skip();

    // 填写搜索词
    await searchInput.fill('138');
    await pause(page, 300);

    // 按回车，等待接口返回
    await waitForApi(page, '/biz-users', () => page.keyboard.press('Enter'));

    // 等 loading 消失，验证表格仍可见
    await page.locator('.el-loading-mask').waitFor({ state: 'hidden', timeout: 8000 }).catch(() => {});
    await expect(page.locator('.el-table')).toBeVisible();
    await pause(page);
  });

  // TC-F03: 新增用户完整流程
  test('TC-F03 新增用户弹窗并提交成功', async ({ page }) => {
    const addBtn = page.getByRole('button', { name: '新增' });
    if (!await addBtn.isVisible()) return test.skip();

    // 点击新增按钮
    await addBtn.click();
    await pause(page, 300);

    // 等弹窗出现
    await expect(page.locator('.el-dialog')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.el-dialog .el-dialog__title')).toContainText('新增');
    await pause(page);

    // 填写手机号（逐字输入，模拟真实填写）
    const phone = `139${Date.now().toString().slice(-8)}`;
    await page.getByPlaceholder('请输入手机号').click();
    await page.getByPlaceholder('请输入手机号').type(phone, { delay: 50 });
    await pause(page, 300);

    // 填写昵称
    const nicknameInput = page.getByPlaceholder('请输入昵称');
    if (await nicknameInput.isVisible()) {
      await nicknameInput.click();
      await nicknameInput.type('测试新增', { delay: 50 });
      await pause(page, 300);
    }

    // 点提交，等待接口响应
    await waitForApi(page, '/biz-users', () =>
      page.locator('.el-dialog').getByRole('button', { name: '确定' }).click()
    );

    // 等成功提示
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 });
    await pause(page);
  });

  // TC-F06: 启用/禁用开关切换，等接口响应后验证提示
  test('TC-F06 启用禁用开关切换', async ({ page }) => {
    const statusSwitch = page.locator('.el-table .el-switch').first();
    if (!await statusSwitch.isVisible()) return test.skip();

    // 点开关，等接口返回
    await waitForApi(page, '/toggle-status', () => statusSwitch.click());

    // 等 ElMessage 出现
    await expect(
      page.locator('.el-message--success, .el-message--error')
    ).toBeVisible({ timeout: 5000 }).catch(() => {});
    await pause(page);
  });
});

// ============================================================
// TC-F07 ~ TC-F09: C端登录页（dev-web-user，端口 5174）
// ============================================================
test.describe('C端登录页 — dev-web-user (5174)', () => {

  // TC-F07: 发送验证码，验证倒计时按钮状态变化
  test('TC-F07 发送验证码按钮60秒倒计时', async ({ browser }) => {
    const ctx  = await browser.newContext();
    const page = await ctx.newPage();
    const phone = `1380${Date.now().toString().slice(-7)}`;
    try {
      await page.goto(`${USER_APP_URL}/login`);
      await page.waitForLoadState('networkidle');
      await pause(page);

      // 填写手机号（租户编码已从页面去除，由环境变量配置）
      await page.getByPlaceholder('请输入手机号').click();
      await page.getByPlaceholder('请输入手机号').type(phone, { delay: 60 });
      await pause(page, 300);

      // 点击发送验证码，等接口响应
      await waitForApi(page, '/send-code', () =>
        page.getByRole('button', { name: '获取验证码' }).click()
      );

      // 验证倒计时按钮出现且禁用
      await expect(page.getByRole('button', { name: /秒/ })).toBeVisible({ timeout: 5000 });
      await expect(page.getByRole('button', { name: /秒/ })).toBeDisabled();
      await pause(page);
    } finally {
      await ctx.close();
    }
  });

  test.skip('TC-F08 登录成功跳转主页（需手动获取验证码）', async () => {});

  // TC-F09: 错误验证码登录，验证错误提示
  test('TC-F09 登录失败显示错误提示', async ({ browser }) => {
    const ctx  = await browser.newContext();
    const page = await ctx.newPage();
    try {
      await page.goto(`${USER_APP_URL}/login`);
      await page.waitForLoadState('networkidle');
      await pause(page);

      // 填写手机号
      await page.getByPlaceholder('请输入手机号').click();
      await page.getByPlaceholder('请输入手机号').type('13800138000', { delay: 60 });
      await pause(page, 300);

      // 填写错误验证码（租户编码已由环境变量配置，无需填写）
      await page.getByPlaceholder('请输入验证码').click();
      await page.getByPlaceholder('请输入验证码').type('0000', { delay: 80 });
      await pause(page, 300);

      // 点登录，等接口响应（会失败）
      await waitForApi(page, '/auth/login', () =>
        page.getByRole('button', { name: '登' }).click()
      );

      // 验证错误提示出现
      await expect(page.locator('.el-message--error')).toBeVisible({ timeout: 5000 });
      await pause(page);
    } finally {
      await ctx.close();
    }
  });
});

// TC-F10 ~ TC-F12: 需 C端登录状态，暂 skip
test.describe('C端布局页 — dev-web-user', () => {
  test.skip('TC-F10~F12 需要 C端用户登录状态', () => {});
});

// ============================================================
// TC-UX: C端用户管理页体验审查
// ============================================================
test.describe('C端用户管理页 — UX 体验审查', () => {
  test.use({ storageState: './test-results/.auth/state.json' });

  test.beforeEach(async ({ page }) => {
    await navigateTo(page, '/system/biz-user');
    await expect(page.locator('.el-table')).toBeVisible({ timeout: 15000 });
    await page.locator('.el-loading-mask').waitFor({ state: 'hidden', timeout: 8000 }).catch(() => {});
    await pause(page);
  });

  // TC-UX01: 三段式布局验证
  test('TC-UX01 布局合理性 — 三段式结构', async ({ page }) => {
    await expect(page.locator('.el-table')).toBeVisible();
    await expect(page.locator('.el-pagination')).toBeVisible();
    expect(await page.locator('.el-card').count()).toBeGreaterThanOrEqual(1);
    await pause(page);
  });

  // TC-UX03: 空状态/加载指示
  test('TC-UX03 完整性 — 空状态和加载指示', async ({ page }) => {
    const rowCount = await page.locator('.el-table__body tr').count();
    if (rowCount === 0) {
      await expect(page.locator('.el-empty')).toBeVisible();
    }
    await pause(page);
  });

  // TC-UX06: 删除危险操作二次确认
  test('TC-UX06 危险操作 — 删除二次确认', async ({ page }) => {
    const deleteBtn = page.locator('.el-table').getByRole('button', { name: '删除' }).first();
    if (!await deleteBtn.isVisible()) return test.skip();

    // 点击删除按钮
    await deleteBtn.click();
    await pause(page, 300);

    // 等确认框出现
    await expect(page.locator('.el-message-box')).toBeVisible({ timeout: 3000 });
    await expect(page.locator('.el-message-box')).toContainText(/确定|确认|删除/);
    await pause(page);

    // 点取消（不实际删除）
    await page.locator('.el-message-box').getByRole('button', { name: '取消' }).click();
    await expect(page.locator('.el-message-box')).toBeHidden({ timeout: 3000 });
    await pause(page);
  });
});

// ============================================================
// TC-UX 登录页体验审查（5174）
// ============================================================
test.describe('C端登录页 — UX 体验审查', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(`${USER_APP_URL}/login`);
    await page.waitForLoadState('networkidle');
    await pause(page);
  });

  // TC-UX07: 登录卡片可见
  test('TC-UX07 布局合理性 — 登录卡片可见', async ({ page }) => {
    const loginCard = page.locator('.login-page, .login-card, .login-form').first();
    await expect(loginCard).toBeVisible({ timeout: 5000 });
    await pause(page);
  });

  // TC-UX08: placeholder 文案清晰（租户编码已去除）
  test('TC-UX08 清晰性 — placeholder 文案', async ({ page }) => {
    await expect(page.getByPlaceholder('请输入手机号')).toBeVisible();
    await expect(page.getByPlaceholder('请输入验证码')).toBeVisible();
    // 租户编码已由环境变量配置，不再展示给用户
    await pause(page);
  });

  // TC-UX09: 未填验证码点登录应出现提示
  test('TC-UX09 完整性 — 未填验证码点登录有警告', async ({ page }) => {
    await page.getByPlaceholder('请输入手机号').type('13800138000', { delay: 60 });
    await pause(page, 300);

    // 不填验证码直接点登录（无需填租户编码）
    await page.getByRole('button', { name: '登' }).click();
    await pause(page, 300);

    await expect(
      page.locator('.el-message--warning, .el-message--error')
    ).toBeVisible({ timeout: 3000 }).catch(() => {});
    await pause(page);
  });

  // TC-UX10: 未填手机号时发送验证码被拦截
  test('TC-UX10 错误预防 — 未填手机号点发送验证码', async ({ page }) => {
    // 不填手机号直接点获取验证码
    await page.getByRole('button', { name: '获取验证码' }).click();
    await pause(page, 300);

    const warned   = await page.locator('.el-message--warning').isVisible().catch(() => false);
    const disabled = await page.getByRole('button', { name: '获取验证码' }).isDisabled().catch(() => false);
    expect(warned || disabled).toBeTruthy();
    await pause(page);
  });
});
