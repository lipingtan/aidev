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
    headless: false,
    // slowMo: 每个底层操作（click/fill/keypress）之间插入 100ms 延迟
    // 模拟真实手速，让操作可见但不至于太慢
    slowMo: 100,
  },
  projects: [
    // 全局 setup：登录并保存状态（依赖前端 localhost:3000）
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
    // 纯 API 测试（不依赖前端，不需要 auth-setup）
    {
      name: 'api',
      use: {
        ...devices['Desktop Chrome'],
      },
      // 无 dependencies，可在没有前端的情况下运行
    },
  ],
  // webServer 已移除：前端/后端需手动启动
  // 前端: localhost:3000 (npm run dev)
  // 后端: localhost:8000 (go run)
});
