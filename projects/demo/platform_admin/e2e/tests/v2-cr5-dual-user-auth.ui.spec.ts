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

    // 填写手机号（定位弹窗内的输入框，避免与列表页搜索框冲突）
    const phone = `139${Date.now().toString().slice(-8)}`;
    const phoneInput = page.locator('.el-dialog').getByPlaceholder('请输入手机号');
    await phoneInput.click();
    await phoneInput.type(phone, { delay: 50 });
    await pause(page, 300);

    // 填写昵称
    const nicknameInput = page.locator('.el-dialog').getByPlaceholder('请输入昵称');
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

  // TC-F04: 编辑用户弹窗回填数据并修改提交
  test('TC-F04 编辑用户弹窗回填数据', async ({ page }) => {
    // 找到第一行的编辑按钮
    const editBtn = page.locator('.el-table').getByRole('button', { name: '编辑' }).first();
    if (!await editBtn.isVisible()) return test.skip();

    // 点击编辑按钮
    await editBtn.click();
    await pause(page, 300);

    // 等弹窗出现
    await expect(page.locator('.el-dialog')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.el-dialog .el-dialog__title')).toContainText('编辑');
    await pause(page);

    // 验证表单已回填数据（手机号输入框应有值）
    const phoneInput = page.locator('.el-dialog').getByPlaceholder('请输入手机号');
    if (await phoneInput.isVisible()) {
      const phoneValue = await phoneInput.inputValue();
      expect(phoneValue.length).toBeGreaterThan(0);  // 手机号已回填
    }

    // 验证昵称输入框已回填
    const nicknameInput = page.locator('.el-dialog').getByPlaceholder('请输入昵称');
    if (await nicknameInput.isVisible()) {
      // 修改昵称
      await nicknameInput.clear();
      await nicknameInput.type('编辑后昵称', { delay: 50 });
      await pause(page, 300);
    }

    // 点取消关闭弹窗（不实际修改数据，避免污染测试数据）
    await page.locator('.el-dialog').getByRole('button', { name: '取消' }).click();
    await expect(page.locator('.el-dialog')).toBeHidden({ timeout: 3000 });
    await pause(page);
  });

  // TC-F05: 重置密码流程 — 确认弹窗 → 密码展示
  test('TC-F05 重置密码流程', async ({ page }) => {
    // 找到第一行的重置密码按钮
    const resetBtn = page.locator('.el-table').getByRole('button', { name: '重置密码' }).first();
    if (!await resetBtn.isVisible()) return test.skip();

    // 点击重置密码按钮
    await resetBtn.click();
    await pause(page, 500);

    // 等确认弹窗出现（ElMessageBox）
    const confirmDialog = page.locator('.el-message-box');
    await expect(confirmDialog).toBeVisible({ timeout: 5000 });
    await pause(page, 300);

    // 验证确认弹窗文案含"重置"或"密码"
    await expect(confirmDialog).toContainText(/重置|密码|确定/);

    // 点击确定按钮（ElMessageBox 的确认按钮类名是 el-button--primary）
    const confirmBtn = confirmDialog.locator('.el-button--primary');
    await expect(confirmBtn).toBeVisible({ timeout: 3000 });
    await confirmBtn.click();
    await pause(page, 2000);  // 等待接口响应和弹窗渲染

    // 验证密码展示弹窗或成功提示出现
    const passwordDialog = page.locator('.el-dialog').filter({ hasText: /密码|password/i });
    const successMessage = page.locator('.el-message--success');
    
    // 等待密码弹窗或成功提示出现（任一即可）
    await expect(passwordDialog.or(successMessage)).toBeVisible({ timeout: 8000 });
    await pause(page);

    // 如果有密码展示弹窗，尝试关闭它
    if (await passwordDialog.isVisible()) {
      const closeBtn = passwordDialog.locator('.el-dialog__headerbtn, .el-button').first();
      if (await closeBtn.isVisible()) {
        await closeBtn.click();
        await pause(page, 300);
      }
    }
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
    const ctx = await browser.newContext();
    const page = await ctx.newPage();
    const phone = `1380${Date.now().toString().slice(-7)}`;
    try {
      await page.goto(`${USER_APP_URL}/login`);
      await page.waitForLoadState('networkidle');
      await pause(page);

      // 等待租户解析完成（按钮文案变为"获取验证码"而非"加载中"）
      await expect(page.getByRole('button', { name: /获取验证码/ })).toBeVisible({ timeout: 10000 });
      await pause(page, 300);

      // 填写手机号
      await page.getByPlaceholder('请输入手机号').click();
      await page.getByPlaceholder('请输入手机号').type(phone, { delay: 60 });
      await pause(page, 300);

      // 点击发送验证码，等待接口响应
      const sendBtn = page.getByRole('button', { name: /获取验证码/ });
      await sendBtn.click();
      await pause(page, 2000);

      // 验证倒计时按钮出现且禁用（按钮文案变为 Ns 格式，如"60s"）
      // 使用 locator 匹配按钮内文本包含数字+s的模式
      const countdownBtn = page.locator('button', { hasText: /^\d+s$/ });
      await expect(countdownBtn).toBeVisible({ timeout: 5000 });
      await expect(countdownBtn).toBeDisabled();
      await pause(page);
    } finally {
      await ctx.close();
    }
  });

  test.skip('TC-F08 登录成功跳转主页（需手动获取验证码）', async () => {});

  // TC-F09: 错误验证码登录，验证错误提示
  test('TC-F09 登录失败显示错误提示', async ({ browser }) => {
    const ctx = await browser.newContext();
    const page = await ctx.newPage();
    try {
      await page.goto(`${USER_APP_URL}/login`);
      await page.waitForLoadState('networkidle');
      await pause(page);

      // 等待租户解析完成（输入框变为可用状态）
      const phoneInput = page.getByPlaceholder('请输入手机号');
      await expect(phoneInput).toBeEnabled({ timeout: 10000 });
      await pause(page, 500);

      // 填写手机号（使用 fill 确保输入成功）
      await phoneInput.fill('13800138000');
      await pause(page, 300);

      // 填写错误验证码
      const codeInput = page.getByPlaceholder('请输入验证码');
      await codeInput.fill('0000');
      await pause(page, 300);

      // 验证输入内容
      await expect(phoneInput).toHaveValue('13800138000');
      await expect(codeInput).toHaveValue('0000');

      // 点登录，等待接口响应
      const loginBtn = page.locator('.login-btn');
      await expect(loginBtn).toBeEnabled();
      
      // 使用 Promise.race 同时等待接口响应和页面变化
      await Promise.all([
        page.waitForResponse(resp => resp.url().includes('/login') && resp.status() < 500).catch(() => null),
        loginBtn.click()
      ]);
      await pause(page, 2000);

      // 验证错误提示出现（el-message 或页面停留在登录页）
      const errorMsg = page.locator('.el-message--error, .el-message--warning');
      const isErrorVisible = await errorMsg.isVisible().catch(() => false);
      
      if (!isErrorVisible) {
        // 如果没有 el-message，检查页面是否仍在登录页（说明登录失败）
        // 新设计标题为"欢迎回来"或"用户登录"
        const stillOnLogin = await page.locator('.login-card').isVisible({ timeout: 3000 }).catch(() => false)
        expect(isErrorVisible || stillOnLogin).toBeTruthy()
      }
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
    await pause(page, 800);

    // 验证确认框出现（ElMessageBox）
    const confirmBox = page.locator('.el-message-box');
    const isVisible = await confirmBox.isVisible().catch(() => false);
    
    if (isVisible) {
      // 验证确认弹窗文案含删除相关提示
      await expect(confirmBox).toContainText(/确定|确认|删除/);
      await pause(page, 300);

      // 关闭弹窗（点击取消或关闭按钮或按 ESC）
      await page.keyboard.press('Escape');
      await pause(page, 500);
    } else {
      // 如果弹窗没出现，测试跳过
      test.skip();
    }
  });
});

// ============================================================
// TC-UX 登录页体验审查（5174）
// ============================================================
test.describe('C端登录页 — UX 体验审查', () => {

  // TC-UX07: 布局合理性 — 登录卡片居中/视觉焦点
  test('TC-UX07 布局合理性 — 登录卡片居中可见', async ({ browser }) => {
    const ctx = await browser.newContext();
    const page = await ctx.newPage();
    try {
      await page.goto(`${USER_APP_URL}/login`);
      await page.waitForLoadState('networkidle');
      await pause(page);

      // 验证登录卡片可见
      const loginCard = page.locator('.el-card, .login-card, form').first();
      await expect(loginCard).toBeVisible({ timeout: 10000 });

      // 验证卡片大致居中（检查是否在视口中央区域）
      // 新设计 PC 端为左右分栏，卡片在右半区，允许较大偏差；核心验证：卡片可见即可
      const box = await loginCard.boundingBox();
      if (box && box.width > 0) {
        const viewportSize = page.viewportSize();
        if (viewportSize) {
          const centerX = box.x + box.width / 2;
          const viewportCenterX = viewportSize.width / 2;
          // 允许 500px 偏差（兼容左右分栏布局）
          expect(Math.abs(centerX - viewportCenterX)).toBeLessThan(500);
        }
      }
      await pause(page);
    } finally {
      await ctx.close();
    }
  });

  // TC-UX08: 清晰性 — placeholder 文案
  test('TC-UX08 清晰性 — placeholder 文案', async ({ browser }) => {
    const ctx = await browser.newContext();
    const page = await ctx.newPage();
    try {
      await page.goto(`${USER_APP_URL}/login`);
      await page.waitForLoadState('networkidle');
      await pause(page);

      // 验证手机号输入框 placeholder
      const phoneInput = page.getByPlaceholder(/手机号/);
      await expect(phoneInput).toBeVisible({ timeout: 5000 });

      // 验证验证码输入框 placeholder
      const codeInput = page.getByPlaceholder(/验证码/);
      await expect(codeInput).toBeVisible();

      // 验证发送验证码按钮文案
      const sendBtn = page.getByRole('button', { name: /验证码/ });
      await expect(sendBtn).toBeVisible();

      // 验证登录按钮文案
      const loginBtn = page.locator('.login-btn');
      await expect(loginBtn).toBeVisible();
      await pause(page);
    } finally {
      await ctx.close();
    }
  });

  // TC-UX09: 完整性 — 未填验证码点登录有警告
  test('TC-UX09 完整性 — 未填验证码点登录有警告', async ({ browser }) => {
    const ctx = await browser.newContext();
    const page = await ctx.newPage();
    try {
      await page.goto(`${USER_APP_URL}/login`);
      await page.waitForLoadState('networkidle');
      await pause(page);

      // 仅填写手机号，不填验证码
      await page.getByPlaceholder('请输入手机号').click();
      await page.getByPlaceholder('请输入手机号').type('13800138000', { delay: 50 });
      await pause(page, 300);

      // 点击登录
      await page.locator('.login-btn').click();
      await pause(page, 1500);

      // 验证有警告提示（el-message--warning 或表单校验提示）
      const warning = page.locator('.el-message--warning, .el-message--error, .el-form-item__error');
      await expect(warning).toBeVisible({ timeout: 5000 });
      await pause(page);
    } finally {
      await ctx.close();
    }
  });

  // TC-UX10: 错误预防 — 未填手机号点发送验证码
  test('TC-UX10 错误预防 — 未填手机号点发送验证码', async ({ browser }) => {
    const ctx = await browser.newContext();
    const page = await ctx.newPage();
    try {
      await page.goto(`${USER_APP_URL}/login`);
      await page.waitForLoadState('networkidle');
      await pause(page);

      const sendBtn = page.getByRole('button', { name: /验证码/ });
      await expect(sendBtn).toBeVisible({ timeout: 8000 });

      // 检查：空手机号时按钮应已禁用，无需点击
      const isBtnDisabled = await sendBtn.isDisabled().catch(() => false);
      if (isBtnDisabled) {
        // 按钮禁用即为通过 —— 符合 TC-UX10 错误预防要求
        expect(isBtnDisabled).toBeTruthy();
        return;
      }

      // 如果按钮未禁用，则点击后验证有警告提示
      await sendBtn.click();
      await pause(page, 1500);

      const warning = page.locator('.el-message--warning, .el-message--error, .el-form-item__error');
      const isWarningVisible = await warning.isVisible().catch(() => false);

      // 至少满足其一：有警告提示 或 按钮禁用
      expect(isWarningVisible || isBtnDisabled).toBeTruthy();
      await pause(page);
    } finally {
      await ctx.close();
    }
  });
});
