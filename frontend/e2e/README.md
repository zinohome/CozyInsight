# CozyInsight E2E Tests (Playwright)

端到端测试覆盖前端 SPA 的关键用户流程：路由、表单、组件渲染。

## 跑测试

```bash
cd frontend

# 装浏览器（只第一次需要）
npx playwright install chromium

# 跑全部测试
npm run e2e

# UI 模式（边点边看）
npm run e2e:ui

# 调试模式
npm run e2e:debug

# 失败后看 trace + 报告
npm run e2e:report
```

## 测试文件

| 文件 | 覆盖内容 |
| --- | --- |
| `e2e/smoke.spec.ts` | 登录页加载、404 路由、token 注入后路由守卫 |

## 跑测试的前置条件

- **基础 E2E** (smoke)：仅需 `npm run dev` 跑起来（Vite + 前端）。
  Playwright config 会自动启动 dev server (`reuseExistingServer: true`)。
- **真接口 E2E**（未来扩展）：需要 `cd backend && go run cmd/server/main.go` 同时跑后端。
  设置 `E2E_BASE_URL=http://localhost:5173` 让 Playwright 用外部 dev server。

## CI 集成

CI 推荐步骤：

```yaml
- name: Install Playwright browsers
  run: npx playwright install --with-deps chromium

- name: Run E2E tests
  run: npm run e2e
  env:
    CI: true

- name: Upload Playwright report
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: playwright-report
    path: playwright-report/
```

## 约定

- **不** 在 E2E 测试里 mock 后端 —— 用真接口或干脆不调后端（注入 token 模拟登录态）。
- **不** 测试组件实现细节（用 vitest 单测覆盖），E2E 只测"用户能看到/能点"。
- 测试名用 `smoke:` / `flow:` / `regression:` 前缀分类。
- 每个测试独立 `page`，不依赖顺序。
