import { test, expect } from '@playwright/test'

/**
 * E2E Smoke: 加载登录页 + 验证关键 UI 元素存在。
 *
 * 这是最简的端到端冒烟测试：证明 dev server 能起来、SPA 能路由、
 * 登录页能渲染。后端不在时不点提交按钮，只验证静态 UI。
 */
test('smoke: 登录页可加载并显示表单', async ({ page }) => {
  await page.goto('/login')

  // 等待登录卡片渲染（避免 React 没挂载完）
  await page.waitForSelector('input[placeholder="用户名"]', { state: 'visible', timeout: 10_000 })

  // 关键 UI 元素（Antd Input 用 placeholder 标识，Form 也用 username/password form item）
  const usernameInputs = page.getByPlaceholder('用户名')
  await expect(usernameInputs.first()).toBeVisible()
  await expect(page.getByPlaceholder('密码').first()).toBeVisible()
  // 登录按钮（页面有"登录"/"注册"两个 tab，登录 tab 上的提交按钮文字包含"登录"）
  await expect(page.getByRole('button', { name: '登 录' })).toBeVisible()
})

test('smoke: 404 路由返回 404 页面', async ({ page }) => {
  await page.goto('/some-route-that-does-not-exist-xyz')
  // Antd Result 组件的 404 文本（不依赖后端，纯静态）
  await expect(page.getByText('404').first()).toBeVisible({ timeout: 10_000 })
})

/**
 * 登录后路由跳转：通过 localStorage 注入一个假 token，验证 Layout 守卫
 * 不会把已登录用户踢回登录页。后端的 /api/v1/user/profile 调用失败会被
 * fetchUser 静默忽略（见 store/auth.ts），所以不依赖后端。
 */
test('smoke: 注入 token 后访问受保护页面不跳登录', async ({ page, context }) => {
  // 先访问站点以建立 localStorage origin
  await page.goto('/login')

  // 注入 token（任意值即可，auth.isAuthenticated 走 !!token）
  await page.evaluate(() => {
    localStorage.setItem('token', 'fake-jwt-token-for-e2e-smoke')
  })

  // 再次访问 /datasource：应该看到 Layout 而不是被踢回 login
  await page.goto('/datasource')

  // 不应跳回 /login
  await page.waitForLoadState('networkidle')
  expect(page.url()).not.toContain('/login')
})

