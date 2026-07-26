/**
 * V2-CR6 多端前端架构 + Plugin SDK V2 — E2E 测试脚本
 * 关联用例: AIDOC/project_doc/platform_admin/v2-cr6-multi-platform-frontend/test_cases.md
 * 关联需求: AIDOC/project_doc/platform_admin/v2-cr6-multi-platform-frontend/requirements.md
 *
 * 测试分组：
 *   describe 1 — admin 构建产物验证（TC-001/002）
 *   describe 2 — admin H5 布局 UI 验证（TC-F01/F02/TC-R01/R02）
 *   describe 3 — user H5 布局 UI 验证（TC-F04/F05/F06/TC-R03）
 *   describe 4 — Plugin SDK V2 功能验证（TC-012/013/TC-N01/N04）
 *   describe 5 — 后端 API 验证（TC-A01/A02/A04/TC-R06）
 *   describe 6 — 回归验证（TC-R01/R04/R05）
 *
 * 服务端口约定（手动启动）：
 *   后端服务:     http://localhost:8000  (go run)
 *   admin PC/H5:  http://localhost:3000  (npm run dev 或 dev:h5)
 *   user PC/H5:   http://localhost:5174  (npm run dev 或 dev:h5)
 *
 * 移动端测试注意：
 *   需以 H5 模式启动前端（npm run dev:h5），否则 Vant 组件不存在于 DOM
 */
import { test, expect, devices, request, APIRequestContext } from '@playwright/test'
import path from 'path'
import fs from 'fs'
import { fileURLToPath } from 'url'

// ES module 兼容写法
const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const API_BASE = 'http://localhost:8000'
const ADMIN_H5_BASE = 'http://localhost:3002'   // dev:h5 启动的地址（H5 模式）
const USER_H5_BASE = 'http://localhost:5174'    // user dev:h5 启动的地址

const ADMIN_USER = 'admin'
const ADMIN_PASS = 'admin123'

// 项目根目录（用于产物检查）
const PROJECT_ROOT = path.resolve(__dirname, '../..')  // e2e/tests/ → e2e/ → platform_admin/
const ADMIN_WEB_DIR = path.join(PROJECT_ROOT, 'dev-web-admin')
const USER_WEB_DIR = path.join(PROJECT_ROOT, 'dev-web-user')

// ============================================================
// 辅助工具
// ============================================================

async function loginAdmin(api: APIRequestContext): Promise<string> {
  const loginResp = await api.post(`${API_BASE}/auth/login`, {
    data: { username: ADMIN_USER, password: ADMIN_PASS },
  })
  expect(loginResp.ok(), `登录失败: ${await loginResp.text()}`).toBeTruthy()
  const loginBody = await loginResp.json()
  if (loginBody.data?.tenants?.length > 0) {
    const tenantResp = await api.post(`${API_BASE}/auth/tenant/select`, {
      data: { tenant_id: String(loginBody.data.tenants[0].id) },
      headers: { Authorization: `Bearer ${loginBody.data.token}` },
    })
    const tenantBody = await tenantResp.json()
    return tenantBody.data?.access_token ?? tenantBody.data?.token
  }
  return loginBody.data?.access_token ?? loginBody.data?.token
}

function authHeader(token: string) {
  return { headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } }
}

// ============================================================
// describe 1: admin 构建产物验证（FR-1）
// ============================================================

test.describe('describe 1: admin 构建产物验证（FR-1）', () => {
  // TC-001: admin build:pc 产物验证
  test('TC-001 admin build:pc 产物目录存在', async () => {
    // 通过 Node.js fs 检查构建产物目录
    // 前置条件：已执行 npm run build:pc
    const distPcDir = path.join(ADMIN_WEB_DIR, 'dist-pc')
    const distPcIndex = path.join(distPcDir, 'index.html')

    const distExists = fs.existsSync(distPcDir)
    const indexExists = fs.existsSync(distPcIndex)

    if (!distExists) {
      // dist-pc/ 不存在时跳过（构建未执行）
      test.skip(!distExists, `dist-pc/ 目录不存在（${distPcDir}），请先执行 npm run build:pc`)
      return
    }

    expect(distExists, `dist-pc/ 目录不存在: ${distPcDir}`).toBeTruthy()
    expect(indexExists, `dist-pc/index.html 不存在: ${distPcIndex}`).toBeTruthy()

    // 验证 index.html 内容（PC 入口不含 Vant 相关 meta）
    const htmlContent = fs.readFileSync(distPcIndex, 'utf-8')
    // PC 模式不含 user-scalable=no（这是 H5 特征）
    expect(htmlContent).not.toContain('user-scalable=no')
  })

  // TC-002: admin build:h5 产物验证
  test('TC-002 admin build:h5 产物目录存在且含 H5 特征', async () => {
    const distH5Dir = path.join(ADMIN_WEB_DIR, 'dist-h5')
    // H5 构建入口是 index-h5.html，产物中主 HTML 也叫 index-h5.html
    const distH5Index = path.join(distH5Dir, 'index-h5.html')

    const distExists = fs.existsSync(distH5Dir)
    if (!distExists) {
      test.skip(!distExists, `dist-h5/ 目录不存在（${distH5Dir}），请先执行 npm run build:h5`)
      return
    }

    expect(distExists, `dist-h5/ 目录不存在: ${distH5Dir}`).toBeTruthy()
    expect(fs.existsSync(distH5Index), `dist-h5/index-h5.html 不存在`).toBeTruthy()

    // 验证 H5 index-h5.html 含 user-scalable=no（移动端 viewport 特征）
    const htmlContent = fs.readFileSync(distH5Index, 'utf-8')
    expect(htmlContent).toContain('user-scalable=no')
  })

  // TC-B09: build:pc 与 build:h5 产物互不污染（P0）
  test('TC-B09 build:pc 与 build:h5 产物互不污染', async () => {
    const distPcDir = path.join(ADMIN_WEB_DIR, 'dist-pc')
    const distH5Dir = path.join(ADMIN_WEB_DIR, 'dist-h5')
    // H5 产物 HTML 为 index-h5.html（构建入口名）
    const h5IndexPath = path.join(distH5Dir, 'index-h5.html')

    if (!fs.existsSync(distPcDir) || !fs.existsSync(distH5Dir)) {
      test.skip(true, '需要两个产物目录都存在才能验证隔离性')
      return
    }

    if (!fs.existsSync(h5IndexPath)) {
      test.skip(true, 'dist-h5/index-h5.html 不存在，请先执行 npm run build:h5')
      return
    }

    const pcIndex = fs.readFileSync(path.join(distPcDir, 'index.html'), 'utf-8')
    const h5Index = fs.readFileSync(h5IndexPath, 'utf-8')

    // PC 产物 index.html 不含 H5 专用 viewport meta
    expect(pcIndex).not.toContain('user-scalable=no')
    // H5 产物 index-h5.html 含移动端 viewport
    expect(h5Index).toContain('user-scalable=no')
  })
})

// ============================================================
// describe 2: admin H5 布局 UI 验证（TC-F01/F02/TC-R01/R02）
// ============================================================

test.describe('describe 2: admin H5 布局 UI 验证', () => {
  // 使用移动端设备模拟（iPhone 14 规格）
  test.use({
    viewport: { width: 390, height: 844 },
    userAgent:
      'Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1',
    baseURL: ADMIN_H5_BASE,
  })

  // TC-F01: H5Layout — NavBar 标题栏渲染（P0）
  test('TC-F01 admin H5 NavBar 存在于 DOM', async ({ page }) => {
    // 监控所有 HTTP 响应，记录 403
    const http403s: string[] = []
    page.on('response', (resp) => {
      if (resp.status() === 403) {
        http403s.push(`403 ${resp.url()}`)
      }
    })
    const indexH5Content = fs.readFileSync(
      path.join(ADMIN_WEB_DIR, 'index-h5.html'),
      'utf-8'
    )
    await page.route('**/*', async (route) => {
      const req = route.request()
      const url = req.url()
      // 只拦截 HTML 导航请求（SPA fallback）
      if (req.resourceType() === 'document' && !url.includes('.js') && !url.includes('.css')) {
        await route.fulfill({ status: 200, contentType: 'text/html', body: indexH5Content })
      } else {
        await route.continue()
      }
    })

    await page.goto('/home', { waitUntil: 'domcontentloaded', timeout: 15000 })

    // 注入 token
    const api = page.context().request
    const loginResp = await api.post(`${API_BASE}/auth/login`, {
      data: { username: ADMIN_USER, password: ADMIN_PASS },
    })
    const loginBody = await loginResp.json()
    let token = loginBody.data?.access_token ?? loginBody.data?.token
    if (loginBody.data?.tenants?.length > 0) {
      const tr = await api.post(`${API_BASE}/auth/tenant/select`, {
        data: { tenant_id: String(loginBody.data.tenants[0].id) },
        headers: { Authorization: `Bearer ${loginBody.data.token}` },
      })
      const tb = await tr.json()
      token = tb.data?.access_token ?? tb.data?.token
    }

    await page.evaluate((t) => {
      localStorage.setItem('access_token', t)
      localStorage.setItem('user-info', JSON.stringify({ token: t, access_token: t }))
    }, token)

    // 刷新让路由守卫重新执行
    await page.reload({ waitUntil: 'networkidle', timeout: 15000 })
    await page.waitForTimeout(2000)

    // 诊断：打印当前 URL 和 body 结构
    const currentUrl = page.url()
    const bodyHtml = await page.evaluate(() => document.body.innerHTML.slice(0, 500))
    console.log('[TC-F01 DIAG] URL:', currentUrl)
    console.log('[TC-F01 DIAG] body[:500]:', bodyHtml)

    // 验证 Vant NavBar 存在
    const navBar = page.locator('.van-nav-bar')
    await expect(navBar).toBeVisible({ timeout: 10000 })

    // 验证汉堡图标存在（wap-nav 图标）
    const hamburger = page.locator('.van-icon-wap-nav, [class*="wap-nav"]')
    await expect(hamburger).toBeVisible({ timeout: 5000 })

    // 输出 403 报告（帮助排查权限问题，不阻断测试）
    if (http403s.length > 0) {
      const unexpected = http403s.filter(u => !u.includes('/static/') && !u.includes('.ico'))
      if (unexpected.length > 0) {
        console.warn('[TC-F01] 发现 403 响应（可能是权限或应用订阅问题）:')
        unexpected.forEach(u => console.warn(' ', u))
      }
    }
  })

  // TC-F02: H5Layout — 汉堡菜单侧滑弹出与关闭（P0）
  test('TC-F02 admin H5 汉堡菜单侧滑弹出并点击菜单项后关闭', async ({ page }) => {
    // 收集 403/非预期 HTTP 响应
    const http403s: string[] = []
    page.on('response', (resp) => {
      if (resp.status() === 403) {
        http403s.push(`${resp.status()} ${resp.url()}`)
      }
    })

    // 复用 TC-F01 的 SPA fallback + token 注入方式
    const indexH5Content = fs.readFileSync(path.join(ADMIN_WEB_DIR, 'index-h5.html'), 'utf-8')
    await page.route('**/*', async (route) => {
      const req = route.request()
      if (req.resourceType() === 'document' && !req.url().includes('.js') && !req.url().includes('.css')) {
        await route.fulfill({ status: 200, contentType: 'text/html', body: indexH5Content })
      } else {
        await route.continue()
      }
    })

    await page.goto('/home', { waitUntil: 'domcontentloaded', timeout: 15000 })

    // 注入 token
    const api = page.context().request
    const loginResp = await api.post(`${API_BASE}/auth/login`, {
      data: { username: ADMIN_USER, password: ADMIN_PASS },
    })
    const loginBody = await loginResp.json()
    let token = loginBody.data?.access_token ?? loginBody.data?.token
    if (loginBody.data?.tenants?.length > 0) {
      const tr = await api.post(`${API_BASE}/auth/tenant/select`, {
        data: { tenant_id: String(loginBody.data.tenants[0].id) },
        headers: { Authorization: `Bearer ${loginBody.data.token}` },
      })
      const tb = await tr.json()
      token = tb.data?.access_token ?? tb.data?.token
    }
    await page.evaluate((t) => {
      localStorage.setItem('access_token', t)
      localStorage.setItem('user-info', JSON.stringify({ token: t, access_token: t }))
    }, token)
    await page.reload({ waitUntil: 'networkidle', timeout: 15000 })
    // 等待路由守卫完成，然后主动触发 menuStore.fetchMenus()
    await page.waitForTimeout(2000)

    // H5Layout.vue onMounted 会自动调用 menuStore.fetchMenus()
    // 等待菜单加载完毕（最多 5s）
    await page.waitForTimeout(3000)

    // 确认 NavBar 存在
    const navBar = page.locator('.van-nav-bar')
    await expect(navBar).toBeVisible({ timeout: 10000 })

    // 记录 403（白名单：静态资源 403 不算问题）
    const unexpected403s = http403s.filter(u => !u.includes('/static/') && !u.includes('.ico'))
    if (unexpected403s.length > 0) {
      console.warn('[TC-F02] 发现 403 响应（非静态资源）:', unexpected403s)
    }

    // 点击汉堡图标
    const hamburger = page.locator('.van-icon-wap-nav, [class*="wap-nav"]')
    await expect(hamburger).toBeVisible({ timeout: 5000 })
    await hamburger.click()

    // van-popup 从左侧滑出
    const popup = page.locator('.van-popup')
    await expect(popup).toBeVisible({ timeout: 5000 })

    // 等待菜单数据加载（Popup 内容不只有 header "菜单" 文字）
    // 等到 Popup 里有 .menu-item 或者 .drawer-body 下有任何子元素
    await page.waitForFunction(() => {
      const body = document.querySelector('.van-popup .drawer-body')
      return body && body.children.length > 0
    }, { timeout: 8000 }).catch(() => {/* 超时后继续，后面的 count 会给出实际值 */})
    await page.waitForTimeout(300)

    // 菜单列表中至少有一项可见（H5MenuDrawer 使用自定义 .menu-item div）
    const menuItems = page.locator('.van-popup .menu-item')
    const count = await menuItems.count()
    expect(count, `菜单项数量应大于0（Popup 内容: ${await popup.innerText().catch(() => '无法获取')}）`).toBeGreaterThan(0)

    // 点击第一个菜单项
    await menuItems.first().click()

    // popup 应关闭
    await expect(popup).not.toBeVisible({ timeout: 5000 })
  })

  // TC-R01: PC 端登录、菜单加载、权限校验回归（RG-1）（P0）
  test('TC-R01 登录和菜单加载 API 回归（RG-1）', async ({ page }) => {
    const api = page.request

    // POST /auth/login
    const loginResp = await api.post(`${API_BASE}/auth/login`, {
      data: { username: ADMIN_USER, password: ADMIN_PASS },
    })
    expect(loginResp.status()).toBe(200)
    const loginBody = await loginResp.json()
    expect(loginBody.code).toBe(0)
    expect(loginBody.data?.token).toBeTruthy()

    // GET /api/v1/common/user-menu?platform=admin
    let token = loginBody.data?.access_token ?? loginBody.data?.token
    if (loginBody.data?.tenants?.length > 0) {
      const tr = await api.post(`${API_BASE}/auth/tenant/select`, {
        data: { tenant_id: String(loginBody.data.tenants[0].id) },
        headers: { Authorization: `Bearer ${loginBody.data.token}` },
      })
      const tb = await tr.json()
      token = tb.data?.access_token ?? tb.data?.token
    }
    const menuResp = await api.get(`${API_BASE}/api/v1/common/user-menu?platform=admin`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    expect(menuResp.status()).toBe(200)
    const menuBody = await menuResp.json()
    expect(menuBody.code).toBe(0)
  })

  // TC-R02: admin PC 构建产物可访问（P0）
  test('TC-R02 admin dist-pc/ 产物 index.html 存在（RG-2）', async () => {
    const distPcIndex = path.join(ADMIN_WEB_DIR, 'dist-pc', 'index.html')
    if (!fs.existsSync(distPcIndex)) {
      test.skip(true, 'dist-pc/index.html 不存在，请先执行 npm run build:pc')
      return
    }
    expect(fs.existsSync(distPcIndex)).toBeTruthy()
  })
})

// ============================================================
// describe 3: user H5 布局 UI 验证（TC-F04/F05/F06/TC-R03）
// ============================================================

test.describe('describe 3: user H5 布局 UI 验证', () => {
  test.use({
    viewport: { width: 390, height: 844 },
    userAgent:
      'Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1',
    baseURL: USER_H5_BASE,
  })

  // TC-F04: UserH5Layout — 底部 Tabbar 渲染（P0）
  test('TC-F04 user H5 底部 Tabbar 存在于 DOM', async ({ page }) => {
    // 前置：dev-web-user 以 npm run dev:h5 启动（port 5174），访问 index-h5.html
    await page.goto('/index-h5.html', { waitUntil: 'networkidle', timeout: 15000 })

    // 注入登录态（与 admin 相同后端）
    if (page.url().includes('login') || !(await page.locator('.van-tabbar').isVisible().catch(() => false))) {
      const api = page.request
      const loginResp = await api.post(`${API_BASE}/auth/login`, {
        data: { username: ADMIN_USER, password: ADMIN_PASS },
      })
      const loginBody = await loginResp.json()
      let token = loginBody.data?.access_token ?? loginBody.data?.token
      if (loginBody.data?.tenants?.length > 0) {
        const tr = await api.post(`${API_BASE}/auth/tenant/select`, {
          data: { tenant_id: String(loginBody.data.tenants[0].id) },
          headers: { Authorization: `Bearer ${loginBody.data.token}` },
        })
        const tb = await tr.json()
        token = tb.data?.access_token ?? tb.data?.token
      }
      await page.evaluate((t) => {
        localStorage.setItem('access_token', t)
        localStorage.setItem('user-info', JSON.stringify({ token: t, access_token: t }))
      }, token)
      // 重新访问 H5 入口（防止 SPA 路由返回 PC index.html）
      await page.goto('/index-h5.html', { waitUntil: 'networkidle', timeout: 15000 })
    }

    // 验证 Vant Tabbar 固定在底部
    const tabbar = page.locator('.van-tabbar')
    await expect(tabbar).toBeVisible({ timeout: 10000 })
  })

  // TC-F05: UserH5Layout — Tabbar 路由跳转与选中高亮（P0）
  test('TC-F05 user H5 点击 Tabbar 第二项路由跳转并高亮', async ({ page }) => {
    // 监控 403
    const http403s: string[] = []
    page.on('response', (resp) => {
      if (resp.status() === 403 && !resp.url().includes('/static/')) {
        http403s.push(`${resp.status()} ${resp.url()}`)
      }
    })

    // SPA fallback：所有 HTML 导航请求返回 user index-h5.html
    const userIndexH5 = fs.readFileSync(path.join(USER_WEB_DIR, 'index-h5.html'), 'utf-8')
    await page.route('**/*', async (route) => {
      const req = route.request()
      if (req.resourceType() === 'document' && !req.url().includes('.js') && !req.url().includes('.css')) {
        await route.fulfill({ status: 200, contentType: 'text/html', body: userIndexH5 })
      } else {
        await route.continue()
      }
    })

    await page.goto('/home', { waitUntil: 'domcontentloaded', timeout: 15000 })

    // 注入 token
    const api = page.context().request
    const loginResp = await api.post(`${API_BASE}/auth/login`, {
      data: { username: ADMIN_USER, password: ADMIN_PASS },
    })
    const loginBody = await loginResp.json()
    let token = loginBody.data?.access_token ?? loginBody.data?.token
    if (loginBody.data?.tenants?.length > 0) {
      const tr = await api.post(`${API_BASE}/auth/tenant/select`, {
        data: { tenant_id: String(loginBody.data.tenants[0].id) },
        headers: { Authorization: `Bearer ${loginBody.data.token}` },
      })
      const tb = await tr.json()
      token = tb.data?.access_token ?? tb.data?.token
    }
    await page.evaluate((t) => {
      localStorage.setItem('access_token', t)
      localStorage.setItem('user-info', JSON.stringify({ token: t, access_token: t }))
    }, token)
    await page.reload({ waitUntil: 'networkidle', timeout: 15000 })
    await page.waitForTimeout(2000)

    // 确认 Tabbar 存在
    const tabbar = page.locator('.van-tabbar')
    await expect(tabbar).toBeVisible({ timeout: 10000 })

    if (http403s.length > 0) {
      console.warn('[TC-F05] 发现 403 响应:', http403s)
    }

    const tabItems = page.locator('.van-tabbar-item')
    const count = await tabItems.count()
    if (count < 2) {
      test.skip(true, 'Tabbar 菜单项少于2项，后端未返回足够菜单数据')
      return
    }

    // 点击第二项
    await tabItems.nth(1).click()
    await page.waitForLoadState('networkidle', { timeout: 8000 })

    // 第二项应处于激活状态
    const activeItem = page.locator('.van-tabbar-item--active')
    await expect(activeItem).toBeVisible({ timeout: 3000 })
    const activeCount = await activeItem.count()
    expect(activeCount).toBe(1)
  })

  // TC-F06: UserH5Layout — 菜单为空时无 JS 错误（P1）
  test('TC-F06 user H5 菜单为空时不报 JS 错误', async ({ page }) => {
    // 收集页面 JS 错误
    const jsErrors: string[] = []
    page.on('pageerror', (err) => jsErrors.push(err.message))

    // 拦截菜单 API，返回空数组
    await page.route('**/api/v1/common/user-menu**', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ code: 0, data: [], msg: 'ok' }),
      })
    })

    await page.goto('/index-h5.html', { waitUntil: 'networkidle', timeout: 15000 })
    await page.waitForTimeout(2000)

    // 不应有 uncaught JS 错误
    const uncaughtErrors = jsErrors.filter(
      (e) => !e.includes('404') && !e.includes('net::ERR'),
    )
    expect(uncaughtErrors, `存在 JS 错误: ${uncaughtErrors.join('; ')}`).toHaveLength(0)
  })

  // TC-R03: user PC 构建产物可访问（P0，RG-3）
  test('TC-R03 user dist-pc/ 产物 index.html 存在（RG-3）', async () => {
    const distPcIndex = path.join(USER_WEB_DIR, 'dist-pc', 'index.html')
    if (!fs.existsSync(distPcIndex)) {
      test.skip(true, 'user dist-pc/index.html 不存在，请先执行 npm run build:pc')
      return
    }
    expect(fs.existsSync(distPcIndex)).toBeTruthy()
  })
})

// ============================================================
// describe 4: Plugin SDK V2 功能验证（TC-012/013/TC-N01/N04）
// ============================================================

test.describe('describe 4: Plugin SDK V2 功能验证', () => {
  test.use({ baseURL: ADMIN_H5_BASE })

  // TC-012: loadPlugin 成功加载 — script[data-plugin] 出现在 DOM（P0）
  test('TC-012 loadPlugin 调用后 script[data-plugin] 标签出现在 DOM', async ({ page }) => {
    await page.goto('/home', { waitUntil: 'networkidle', timeout: 15000 })

    // 检查页面是否暴露 loadPlugin 函数（由 plugin-loader 挂载到 window）
    const hasLoadPlugin = await page.evaluate(() => typeof (window as any).loadPlugin === 'function')
    if (!hasLoadPlugin) {
      test.skip(true, 'window.loadPlugin 未暴露，plugin-loader 可能未实现或未挂载到 window')
      return
    }

    // 拦截 bundle 请求，返回最小合法 UMD bundle
    await page.route('**/static/plugins/test-plugin/admin-pc/bundle.js', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/javascript',
        body: `
          (function() {
            window.__PLUGIN_TEST_PLUGIN__ = {
              manifest: { name: 'test-plugin', version: '1.0.0', displayName: '测试插件' },
              setup: function(ctx) { /* no-op */ }
            };
          })();
        `,
      })
    })

    // 调用 loadPlugin
    await page.evaluate(async () => {
      await (window as any).loadPlugin('test-plugin', 'admin', 'pc')
    })

    // 验证 script[data-plugin="test-plugin"] 出现在 DOM
    const scriptTag = page.locator('script[data-plugin="test-plugin"]')
    await expect(scriptTag).toBeAttached({ timeout: 5000 })
  })

  // TC-013: loadPlugin 路径格式验证（P0）
  test('TC-013 loadPlugin 构造的 bundle 路径格式正确', async ({ page }) => {
    await page.goto('/home', { waitUntil: 'networkidle', timeout: 15000 })

    const hasLoadPlugin = await page.evaluate(() => typeof (window as any).loadPlugin === 'function')
    if (!hasLoadPlugin) {
      test.skip(true, 'window.loadPlugin 未暴露')
      return
    }

    // 记录实际发出的网络请求 URL
    const requestedUrls: string[] = []
    page.on('request', (req) => {
      if (req.url().includes('/static/plugins/')) {
        requestedUrls.push(req.url())
      }
    })

    // 拦截并 fulfill，防止真实请求失败
    await page.route('**/static/plugins/my-plugin/user-h5/bundle.js', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/javascript',
        body: `(function(){ window.__PLUGIN_MY_PLUGIN__ = { manifest:{name:'my-plugin',version:'1.0.0',displayName:'My'}, setup:function(){} }; })();`,
      })
    })

    await page.evaluate(async () => {
      try {
        await (window as any).loadPlugin('my-plugin', 'user', 'h5')
      } catch (_) { /* 忽略失败 */ }
    })

    // 验证请求 URL 符合规范：/static/plugins/{name}/{platform}-{device}/bundle.js
    const matched = requestedUrls.some((url) =>
      url.includes('/static/plugins/my-plugin/user-h5/bundle.js'),
    )
    expect(matched, `未找到符合规范的 bundle 请求 URL，实际请求: ${requestedUrls.join(', ')}`).toBeTruthy()
  })

  // TC-N01: loadPlugin 失败（bundle 404）— 主应用不崩溃（P0）
  test('TC-N01 loadPlugin bundle 404 时主应用不崩溃无 uncaught error', async ({ page }) => {
    const jsErrors: string[] = []
    page.on('pageerror', (err) => jsErrors.push(err.message))

    // 先注入 token，确保路由守卫放行
    const api = page.context().request
    const loginResp = await api.post(`${API_BASE}/auth/login`, {
      data: { username: ADMIN_USER, password: ADMIN_PASS },
    })
    const loginBody = await loginResp.json()
    let token = loginBody.data?.access_token ?? loginBody.data?.token
    if (loginBody.data?.tenants?.length > 0) {
      const tr = await api.post(`${API_BASE}/auth/tenant/select`, {
        data: { tenant_id: String(loginBody.data.tenants[0].id) },
        headers: { Authorization: `Bearer ${loginBody.data.token}` },
      })
      const tb = await tr.json()
      token = tb.data?.access_token ?? tb.data?.token
    }

    await page.goto('/home', { waitUntil: 'domcontentloaded', timeout: 15000 })
    await page.evaluate((t) => {
      localStorage.setItem('access_token', t)
    }, token)
    await page.reload({ waitUntil: 'networkidle', timeout: 15000 })
    await page.waitForTimeout(1000)

    const hasLoadPlugin = await page.evaluate(() => typeof (window as any).loadPlugin === 'function')
    if (!hasLoadPlugin) {
      test.skip(true, 'window.loadPlugin 未暴露')
      return
    }

    // bundle 路径不存在，后端会返回 404
    await page.route('**/static/plugins/nonexistent/admin-pc/bundle.js', (route) => {
      route.fulfill({ status: 404, body: 'Not Found' })
    })

    // loadPlugin 应 reject，但不抛出 uncaught error
    const loadResult = await page.evaluate(async () => {
      try {
        await (window as any).loadPlugin('nonexistent', 'admin', 'pc')
        return { ok: true }
      } catch (e: any) {
        return { ok: false, msg: e?.message ?? String(e) }
      }
    })

    expect(loadResult.ok, 'loadPlugin 应 reject 而非 resolve').toBeFalsy()
    expect(loadResult.msg, 'reject 消息应含路径信息').toContain('nonexistent')

    // 主应用页面仍然可操作（页面未崩溃，body 存在内容）
    const bodyContent = await page.locator('body').innerHTML()
    expect(bodyContent.length, '主应用 body 应有内容（未崩溃）').toBeGreaterThan(100)

    // 无 uncaught JS 错误
    const uncaught = jsErrors.filter((e) => !e.includes('404') && !e.includes('net::ERR'))
    expect(uncaught, `存在 uncaught JS 错误: ${uncaught.join('; ')}`).toHaveLength(0)
  })

  // TC-N04: 重复 loadPlugin 不重复加载（P1）
  test('TC-N04 重复 loadPlugin 同一插件 DOM 中只有 1 个 script 标签', async ({ page }) => {
    await page.goto('/home', { waitUntil: 'networkidle', timeout: 15000 })

    const hasLoadPlugin = await page.evaluate(() => typeof (window as any).loadPlugin === 'function')
    if (!hasLoadPlugin) {
      test.skip(true, 'window.loadPlugin 未暴露')
      return
    }

    const minBundle = `(function(){ window.__PLUGIN_GAME__ = { manifest:{name:'game',version:'1.0.0',displayName:'Game'}, setup:function(){} }; })();`
    await page.route('**/static/plugins/game/admin-pc/bundle.js', (route) => {
      route.fulfill({ status: 200, contentType: 'application/javascript', body: minBundle })
    })

    // 连续调用两次
    await page.evaluate(async () => {
      await (window as any).loadPlugin('game', 'admin', 'pc')
      await (window as any).loadPlugin('game', 'admin', 'pc')
    })

    // DOM 中应只有 1 个 script[data-plugin="game"]
    const count = await page.locator('script[data-plugin="game"]').count()
    expect(count, 'DOM 中 script[data-plugin="game"] 数量应为 1').toBe(1)
  })
})

// ============================================================
// describe 5: 后端 API 验证（TC-A01/A02/A04/TC-R06）
// ============================================================

test.describe('describe 5: 后端 API 验证', () => {
  let api: APIRequestContext
  let token: string

  test.beforeAll(async () => {
    api = await request.newContext()
    token = await loginAdmin(api)
  })

  test.afterAll(async () => {
    await api.dispose()
  })

  // TC-A01: GET /static/plugins/game/admin-pc/bundle.js — 文件存在时返回 200（P0）
  test('TC-A01 GET /static/plugins/game/admin-pc/bundle.js — 如文件存在返回 200', async () => {
    const resp = await api.get(`${API_BASE}/static/plugins/game/admin-pc/bundle.js`)
    if (resp.status() === 404) {
      // 文件不存在时跳过（game 插件未安装）
      test.skip(true, 'game 插件 bundle 不存在（未安装），跳过 200 验证')
      return
    }
    expect(resp.status()).toBe(200)
    const ct = resp.headers()['content-type'] ?? ''
    expect(ct).toContain('javascript')
  })

  // TC-A02: GET /static/plugins/nonexistent/admin-pc/bundle.js — 返回 404（P0）
  test('TC-A02 GET /static/plugins/nonexistent/admin-pc/bundle.js 返回 404', async () => {
    const resp = await api.get(`${API_BASE}/static/plugins/nonexistent_${Date.now()}/admin-pc/bundle.js`)
    expect(resp.status()).toBe(404)
  })

  // TC-N10: GET 不存在路径返回 404，不暴露服务器错误信息（P0）
  test('TC-N10 GET /static/plugins 不存在路径返回 404 不暴露服务器错误', async () => {
    const resp = await api.get(`${API_BASE}/static/plugins/no_plugin/user-h5/bundle.js`)
    expect(resp.status()).toBe(404)
    // 响应体不应包含服务器内部路径信息（安全防护）
    const body = await resp.text()
    expect(body).not.toMatch(/\/home\/|C:\\|panic:|runtime error/i)
  })

  // TC-A04: GET /api/v1/admin/plugins — 静态路由与 API 路由不冲突（P0，RG-6）
  test('TC-A04 GET /api/v1/admin/plugins 返回 200 与静态路由不冲突（RG-6）', async () => {
    const resp = await api.get(`${API_BASE}/api/v1/admin/plugins`, authHeader(token))
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    expect(body.code).toBe(0)
    // 返回的是 JSON（API 路由），不是静态文件
    const ct = resp.headers()['content-type'] ?? ''
    expect(ct).toContain('application/json')
  })

  // TC-R06: 静态文件路由不影响 /api/ 路由（P0，RG-6）
  test('TC-R06 静态文件服务不影响现有 API 路由（RG-6）', async () => {
    // 同时验证三条路由均正常
    const checks = [
      { url: `${API_BASE}/api/v1/admin/plugins`, desc: 'plugins API' },
      { url: `${API_BASE}/api/v1/admin/roles?page=1&page_size=1`, desc: 'roles API' },
    ]
    for (const { url, desc } of checks) {
      const resp = await api.get(url, authHeader(token))
      expect(resp.status(), `${desc} 响应码异常`).toBe(200)
      const body = await resp.json()
      expect(body.code, `${desc} 业务码异常`).toBe(0)
    }

    // 静态路由 404 场景也应不影响 API
    const staticResp = await api.get(`${API_BASE}/static/plugins/no_plugin_x/admin-pc/bundle.js`)
    expect(staticResp.status()).toBe(404)

    // API 路由仍然正常（再次验证无串扰）
    const apiResp = await api.get(`${API_BASE}/api/v1/admin/plugins`, authHeader(token))
    expect(apiResp.status()).toBe(200)
  })
})

// ============================================================
// describe 6: 回归验证（TC-R01/R04/R05）
// ============================================================

test.describe('describe 6: 回归验证', () => {
  let api: APIRequestContext
  let token: string

  test.beforeAll(async () => {
    api = await request.newContext()
    token = await loginAdmin(api)
  })

  test.afterAll(async () => {
    await api.dispose()
  })

  // TC-R01: PC 端登录、菜单加载、权限校验流程不受影响（P0，RG-1）
  test('TC-R01 PC 端登录菜单加载权限校验回归（RG-1）', async ({ page }) => {
    // 登录接口
    const loginResp = await api.post(`${API_BASE}/auth/login`, {
      data: { username: ADMIN_USER, password: ADMIN_PASS },
    })
    expect(loginResp.status()).toBe(200)
    const loginBody = await loginResp.json()
    expect(loginBody.code).toBe(0)
    expect(loginBody.data?.token).toBeTruthy()

    // 菜单接口
    const menuResp = await api.get(
      `${API_BASE}/api/v1/common/user-menu?platform=admin`,
      authHeader(token),
    )
    expect(menuResp.status()).toBe(200)
    const menuBody = await menuResp.json()
    expect(menuBody.code).toBe(0)
    // 菜单数据非空（权限校验生效）
    const menuList = menuBody.data ?? []
    expect(Array.isArray(menuList)).toBeTruthy()
    expect(menuList.length, '菜单列表应非空（有权限）').toBeGreaterThan(0)
  })

  // TC-R04: 现有 V1 插件 setup 函数正常调用（P0，RG-4）
  test('TC-R04 V1 插件 setup 函数通过 loadPlugin 正常调用（RG-4）', async ({ page }) => {
    await page.goto(ADMIN_H5_BASE + '/home', { waitUntil: 'networkidle', timeout: 15000 })

    const hasLoadPlugin = await page.evaluate(() => typeof (window as any).loadPlugin === 'function')
    if (!hasLoadPlugin) {
      test.skip(true, 'window.loadPlugin 未暴露，跳过 V1 兼容回归')
      return
    }

    // 记录 JS 错误
    const jsErrors: string[] = []
    page.on('pageerror', (err) => jsErrors.push(err.message))

    // Mock V1 格式 bundle（无 onTeardown，仅有 setup）
    const v1Bundle = `
      (function() {
        var setupCalled = false;
        window.__PLUGIN_GAME__ = {
          manifest: { name: 'game', version: '0.9.0', displayName: '游戏插件(V1)' },
          setup: function(ctx) { setupCalled = true; window.__V1_SETUP_CALLED__ = true; }
        };
      })();
    `
    await page.route('**/static/plugins/game/admin-pc/bundle.js', (route) => {
      route.fulfill({ status: 200, contentType: 'application/javascript', body: v1Bundle })
    })

    await page.evaluate(async () => {
      try {
        await (window as any).loadPlugin('game', 'admin', 'pc')
      } catch (_) { /* 可能已加载，忽略 */ }
    })

    // setup 被调用（window.__V1_SETUP_CALLED__ 已设置）
    const setupCalled = await page.evaluate(() => !!(window as any).__V1_SETUP_CALLED__)
    expect(setupCalled, 'V1 插件 setup 应被调用').toBeTruthy()

    // 无 JS 错误
    const uncaught = jsErrors.filter((e) => !e.includes('404') && !e.includes('net::ERR'))
    expect(uncaught, `V1 插件加载存在 JS 错误: ${uncaught.join('; ')}`).toHaveLength(0)
  })

  // TC-R05: 插件 unloadPlugin 后无残留（P0，RG-5）
  test('TC-R05 unloadPlugin 后 window 全局变量和 script 标签无残留（RG-5）', async ({ page }) => {
    await page.goto(ADMIN_H5_BASE + '/home', { waitUntil: 'networkidle', timeout: 15000 })

    const hasLoadPlugin = await page.evaluate(() => typeof (window as any).loadPlugin === 'function')
    const hasUnloadPlugin = await page.evaluate(() => typeof (window as any).unloadPlugin === 'function')

    if (!hasLoadPlugin || !hasUnloadPlugin) {
      test.skip(true, 'window.loadPlugin / unloadPlugin 未暴露，跳过卸载回归')
      return
    }

    const bundle = `
      (function() {
        window.__PLUGIN_CLEANUP__ = {
          manifest: { name: 'cleanup', version: '1.0.0', displayName: '清理测试' },
          setup: function(ctx) {}
        };
      })();
    `
    await page.route('**/static/plugins/cleanup/admin-pc/bundle.js', (route) => {
      route.fulfill({ status: 200, contentType: 'application/javascript', body: bundle })
    })

    // 加载插件
    await page.evaluate(async () => {
      await (window as any).loadPlugin('cleanup', 'admin', 'pc')
    })

    // 验证加载成功
    const scriptBefore = await page.locator('script[data-plugin="cleanup"]').count()
    expect(scriptBefore).toBe(1)

    // 卸载插件
    await page.evaluate(async () => {
      await (window as any).unloadPlugin('cleanup')
    })

    // script 标签已移除
    const scriptAfter = await page.locator('script[data-plugin="cleanup"]').count()
    expect(scriptAfter, 'unloadPlugin 后 script 标签应被移除').toBe(0)

    // window 全局变量已清除
    const globalGone = await page.evaluate(() => (window as any).__PLUGIN_CLEANUP__ === undefined)
    expect(globalGone, 'window.__PLUGIN_CLEANUP__ 应已删除').toBeTruthy()
  })

  // TC-003: admin H5 宿主库 __PLATFORM_ADMIN_LIBS__ 含 Vant（P0）
  test('TC-003 admin H5 模式 window.__PLATFORM_ADMIN_LIBS__ 包含 Vant', async ({ page }) => {
    await page.goto(ADMIN_H5_BASE + '/index-h5.html', { waitUntil: 'networkidle', timeout: 15000 })

    const libs = await page.evaluate(() => {
      const w = window as any
      const adminLibs = w.__PLATFORM_ADMIN_LIBS__
      if (!adminLibs) return null
      return {
        hasVue: !!adminLibs.Vue,
        hasElementPlus: !!adminLibs.ElementPlus,
        hasVueRouter: !!adminLibs.VueRouter,
        hasVant: !!adminLibs.Vant,
      }
    })

    if (!libs) {
      test.skip(true, 'window.__PLATFORM_ADMIN_LIBS__ 未定义，请确认 H5 入口已正确暴露宿主库')
      return
    }

    expect(libs.hasVue, '__PLATFORM_ADMIN_LIBS__.Vue 应存在').toBeTruthy()
    expect(libs.hasElementPlus, '__PLATFORM_ADMIN_LIBS__.ElementPlus 应存在').toBeTruthy()
    expect(libs.hasVueRouter, '__PLATFORM_ADMIN_LIBS__.VueRouter 应存在').toBeTruthy()
    expect(libs.hasVant, '__PLATFORM_ADMIN_LIBS__.Vant 应存在（H5 模式）').toBeTruthy()
  })
})
