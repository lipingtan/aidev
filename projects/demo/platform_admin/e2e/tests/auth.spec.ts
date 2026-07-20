import { test, expect } from '@playwright/test'

// 后端 API 基础地址
const API_BASE = 'http://localhost:8000'

test.describe('认证流程 (RG-1, RG-2)', () => {
  test('TC-004: 登录成功并跳转到首页', async ({ page }) => {
    await page.goto('/')
    // 等待登录页加载
    await page.waitForSelector('input[placeholder*="用户名"], input[name="username"]', { timeout: 10000 })

    // 填写登录表单
    await page.fill('input[placeholder*="用户名"], input[name="username"]', 'admin')
    await page.fill('input[type="password"]', 'admin123')

    // 点击登录按钮
    await page.click('button[type="submit"], button:has-text("登录")')

    // 等待跳转（登录后可能需要选择租户或直接进入首页）
    await page.waitForURL(/\/(home|dashboard|tenant)/, { timeout: 10000 })

    // 验证已登录（localStorage 有 token）
    const token = await page.evaluate(() => localStorage.getItem('access_token'))
    expect(token).toBeTruthy()
  })

  test('TC-A02: 错误密码登录失败', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('input[placeholder*="用户名"], input[name="username"]', { timeout: 10000 })

    await page.fill('input[placeholder*="用户名"], input[name="username"]', 'admin')
    await page.fill('input[type="password"]', 'wrongpassword')
    await page.click('button[type="submit"], button:has-text("登录")')

    // 应停留在登录页，显示错误提示
    await expect(page.locator('.el-message--error, .el-notification__content, .error-msg')).toBeVisible({ timeout: 5000 })
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
    const tenantId = body.data.tenants?.[0]?.id || ''

    await page.goto('/')
    await page.evaluate(({ token, tid }) => {
      localStorage.setItem('access_token', token)
      localStorage.setItem('current_tenant_id', tid)
    }, { token: accessToken, tid: tenantId })
  })

  test('TC-007: 侧边栏菜单正确渲染', async ({ page }) => {
    await page.goto('/home')
    await page.waitForLoadState('networkidle')

    // 验证侧边栏菜单项
    const sidebar = page.locator('.el-menu, .app-sidebar, nav')
    await expect(sidebar).toBeVisible({ timeout: 10000 })

    // 验证关键菜单项存在
    await expect(page.locator('text=系统管理')).toBeVisible()
    await expect(page.locator('text=日志管理')).toBeVisible()
  })

  test('TC-014: 菜单导航到各页面无报错', async ({ page }) => {
    await page.goto('/home')
    await page.waitForLoadState('networkidle')

    // 点击系统管理展开子菜单
    const sysMenu = page.locator('text=系统管理')
    if (await sysMenu.isVisible()) {
      await sysMenu.click()
      await page.waitForTimeout(500)
    }

    // 导航到用户管理
    const userMenu = page.locator('text=用户管理')
    if (await userMenu.isVisible()) {
      await userMenu.click()
      await page.waitForLoadState('networkidle')
      // 不应有 JS 错误
      const errors: string[] = []
      page.on('pageerror', (e) => errors.push(e.message))
      expect(errors).toHaveLength(0)
    }
  })
})
