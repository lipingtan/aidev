import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: 0,
  workers: 1,
  reporter: [['html', { open: 'never' }], ['list']],
  timeout: 60000,
  expect: {
    timeout: 10000
  },
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10000,
    // UI 测试默认有头模式
    headless: false,
  },
  projects: [
    // 全局 setup：登录并保存状态
    {
      name: 'auth-setup',
      testMatch: /global-setup\.ts/,
    },
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        // 复用登录状态，避免每个测试重复登录
        storageState: './test-results/.auth/state.json',
      },
      dependencies: ['auth-setup'],
    },
  ],
  // webServer 已移除：前端/后端需手动启动
  // 前端: localhost:3000 (npm run dev)
  // 后端: localhost:8000 (go run)
});
