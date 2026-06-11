import { defineConfig, devices } from '@playwright/test'

/**
 * Playwright E2E configuration for CozyInsight.
 *
 * 默认只跑 chromium + 1 worker（CI 友好），启用失败重试、HTML 报告。
 * 启动 Vite dev server（端口 5173，代理到后端 8100），前端单页应用直接访问。
 *
 * Run with: npx playwright test
 * 调试: npx playwright test --debug
 * 报告:  npx playwright show-report
 *
 * 前置条件：后端服务在 :8100 启动（`cd backend && go run cmd/server/main.go`）。
 * 如果只想跑前端 + mock 后端，可用 `npm run dev` 启动 Vite，Playwright 会自动复用。
 */
export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: process.env.E2E_BASE_URL ?? 'http://localhost:5173',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  // 跑测试时自动启动 Vite dev server；已有实例则跳过。
  webServer: process.env.E2E_BASE_URL
    ? undefined
    : {
        command: 'npm run dev',
        url: 'http://localhost:5173',
        reuseExistingServer: true,
        timeout: 60_000,
        stdout: 'ignore',
        stderr: 'pipe',
      },
  timeout: 30_000,
  expect: {
    timeout: 5_000,
  },
})
