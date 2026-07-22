/**
 * 全局认证 Setup：登录一次并保存浏览器状态（cookie + localStorage）
 * 后续所有 UI 测试复用此状态，无需重复登录
 */
import { test as setup, expect } from '@playwright/test';
import { fileURLToPath } from 'url';
import path from 'path';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const AUTH_FILE = path.join(__dirname, '../test-results/.auth/state.json');

const BASE_URL = 'http://localhost:3000';
const API_BASE = 'http://localhost:8000';
const ADMIN_USER = 'admin';
const ADMIN_PASS = 'admin123';

setup('全局登录', async ({ page }) => {
  // 1. 通过 API 获取 token（更快更稳定）
  const loginResp = await page.request.post(`${API_BASE}/auth/login`, {
    data: { username: ADMIN_USER, password: ADMIN_PASS },
  });
  expect(loginResp.ok()).toBeTruthy();
  const loginBody = await loginResp.json();

  let accessToken: string;
  const platformToken = loginBody.data?.token;

  // 2. 选择租户获取 access_token
  if (loginBody.data?.tenants?.length > 0) {
    const tenantResp = await page.request.post(`${API_BASE}/auth/tenant/select`, {
      data: { tenant_id: String(loginBody.data.tenants[0].id) },
      headers: { Authorization: `Bearer ${platformToken}` },
    });
    const tenantBody = await tenantResp.json();
    accessToken = tenantBody.data?.access_token ?? tenantBody.data?.token;
  } else {
    accessToken = loginBody.data?.access_token ?? platformToken;
  }

  // 3. 访问前端页面并注入 token 到 localStorage
  await page.goto(BASE_URL);
  await page.evaluate((token) => {
    // 按项目前端存储格式注入（根据 request.ts 中的读取方式）
    const userInfo = JSON.stringify({
      token: token,
      access_token: token,
    });
    localStorage.setItem('user-info', userInfo);
    localStorage.setItem('access_token', token);
  }, accessToken);

  // 4. 刷新页面使 token 生效
  await page.goto(`${BASE_URL}/home`);
  // 等待页面加载完成（不再跳转到登录页）
  await page.waitForURL(/\/(home|system|tenant-select)/, { timeout: 15000 });

  // 如果跳转到租户选择页，选择第一个
  if (page.url().includes('tenant-select')) {
    await page.click('.tenant-item >> nth=0, .tenant-card >> nth=0, [class*="tenant"] >> nth=0');
    await page.waitForURL('**/home', { timeout: 10000 });
  }

  // 5. 保存浏览器状态
  await page.context().storageState({ path: AUTH_FILE });
});
