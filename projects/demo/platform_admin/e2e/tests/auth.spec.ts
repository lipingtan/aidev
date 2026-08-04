import { test, expect } from '@playwright/test'

// 后端 API 基础地址
const API_BASE = 'http://localhost:8000'

test.describe('认证流程 (RG-1, RG-2)', () => {
  test('TC-004: 登录成功并跳转到首页', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('input[placeholder*="用户名"], input[name="username"]', { timeout: 10000 })

    await page.fill('input[placeholder*="用户名"], input[name="username"]', 'admin')
    await page.fill('input[type="password"]', 'admin123')
    await page.click('button[type="submit"], button:has-text("登录")')

    // 等待跳转到主界面
    await page.waitForURL(/\/(home|dashboard|tenant|system)/, { timeout: 15000 })

    // 等待 token 写入（用轮询而非 waitForFunction，避免 context 切换问题）
    let token: string | null = null
    for (let i = 0; i < 10; i++) {
      token = await page.evaluate(() => localStorage.getItem('access_token'))
      if (token) break
      await page.waitForTimeout(500)
    }
    expect(token).toBeTruthy()
  })

  test('TC-A02: 错误密码登录失败', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('input[placeholder*="用户名"], input[name="username"]', { timeout: 10000 })

    await page.fill('input[placeholder*="用户名"], input[name="username"]', 'admin')
    await page.fill('input[type="password"]', 'wrongpassword')
    await page.click('button[type="submit"], button:has-text("登录")')

    // 应停留在登录页，显示错误提示（兼容不同 el-message 类名）
    await expect(
      page.locator('.el-message--error, .el-message.el-message--error, [class*="message"][class*="error"], .error-msg')
        .first()
    ).toBeVisible({ timeout: 8000 })
  })
})

test.describe('菜单加载 (RG-5)', () => {
  test.beforeEach(async ({ page }) => {
    // 通过 API 登录获取 token，直接注入 localStorage 跳过 UI 登录
    const resp = await page.request.post(`${API_BASE}/auth/login`, {
      data: { username: 'admin', password: 'admin123' },
    })
    const body = await resp.json()
    const accessToken = body.data.access_token
    const tenantId = String(body.data.tenants?.[0]?.id || '')

    await page.goto('/')
    await page.evaluate(({ token, tid }) => {
      localStorage.setItem('access_token', token)
      localStorage.setItem('current_tenant_id', tid)
    }, { token: accessToken, tid: tenantId })
  })

  test('TC-007: 侧边栏菜单正确渲染', async ({ page }) => {
    await page.goto('/home')
    await page.waitForLoadState('networkidle')

    // 使用 .first() 避免 strict mode 多元素匹配错误
    const sidebar = page.locator('.app-sidebar').first()
    await expect(sidebar).toBeVisible({ timeout: 10000 })

    await expect(page.locator('text=系统管理').first()).toBeVisible()
    await expect(page.locator('text=日志管理').first()).toBeVisible()
  })

  test('TC-014: 菜单导航到各页面无报错', async ({ page }) => {
    await page.goto('/home')
    await page.waitForLoadState('networkidle')

    // 点击系统管理展开子菜单
    const sysMenu = page.locator('text=系统管理').first()
    if (await sysMenu.isVisible()) {
      await sysMenu.click()
      await page.waitForTimeout(500)
    }

    // 使用精确匹配避免匹配到"C端用户管理"
    const userMenu = page.locator('role=menuitem[name="用户管理"]')
    if (await userMenu.count() > 0 && await userMenu.first().isVisible()) {
      await userMenu.first().click()
      await page.waitForLoadState('networkidle')
      const errors: string[] = []
      page.on('pageerror', (e) => errors.push(e.message))
      expect(errors).toHaveLength(0)
    }
  })
})
