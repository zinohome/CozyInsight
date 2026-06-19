# Handoff — CozyInsight 开发接力

> 给接手这个项目的 agent:阅读本文档以快速了解**当前状态**、**刚完成什么**、**接下来要做什么**、**踩过的坑**。

最后更新:2026-06-19

---

## 🚦 当前状态(2026-06-19)

- **分支**:`main`(唯一活跃分支,临时分支 `frontend-port` / `backend-port` 已合并并清理)
- **`origin/main` 头**:`6c3ab65`(比远端 `4a7af5f` 领先 6 个本地 commit)
- **工作区**:干净,无未提交改动
- **构建状态**:`go build ./...` 0 错误,`go test ./internal/...` 全部通过

## 📜 最近 6 个 commit(全部已 push 到 `origin/main`)

```
6c3ab65 feat(backend): add RateLimit middleware + schedule system on sqlx stack
2f54203 chore(gitignore): explicitly ignore backend/server binary + coverage
ad0a71e chore: untrack backend/server binary (95MB Go build output)
b5124f5 feat(frontend): DashboardList with breadcrumb + rename modal + folder support
a5db0e7 fix(frontend): resolve all TypeScript errors (request wrapper + strict-mode sweep)
5c02e32 fix(frontend): repair broken ChartEditor, fix import paths, add column chart type
4a7af5f (origin/main base) test(frontend): bootstrap E2E test framework with Playwright + smoke tests
```

### Stage A — 前端 TypeScript / ChartEditor / DashboardList(5c02e32, a5db0e7, b5124f5)
- **`request.ts` 重写**:`Unwrap<T>` 泛型自动从 `AxiosResponse<infer U>` 解包,统一 `Promise<U>`;错误拦截器把 `code !== 200` 转为 `Promise.reject`
- **`chart.ts` types**:HEAD 27 种 + 本地 2 个别名(`column` / `map`)
- **ChartEditor**:修复 `pages/chart/...` 导入路径、补缺失的 `column` chart type
- **DashboardList**:完全重写 — 面包屑导航 + 文件夹/仪表板混合树 + Popconfirm 删除 + 重命名 modal
  - 关键:API 方法名 `dashboardAPI.delete` → `dashboardAPI.remove`(匹配后端实际接口)
  - 状态机:`breadcrumbs` 数组记录路径,`currentPid` 跟踪当前位置

### Stage C — Git 卫生(ad0a71e, 2f54203)
- **`backend/server` 95MB Go 二进制**被从 git 索引移除(`git rm --cached`),磁盘文件保留供本地运行
- **`.gitignore` 补全**:`backend/server`、`backend/coverage*.out`

### Stage B — 后端 sqlx 移植(6c3ab65)
远程 main 用 **sqlx** 替代了旧的 GORM/Avatica,所以本地基于 Avatica 写的中间件/schedule 系统需要重做。

**新增 RateLimit 中间件**:
- 文件:`backend/internal/middleware/common.go`
- Token bucket 实现,工厂 `RateLimit(rate, burst float64)`
- 单实例全局共享桶(生产环境应换 Redis);4 个单元测试

**新增 Schedule 完整系统**:
- `model.ScheduleTask` + `migrations/018_schedule_tasks.sql`
- `repository.ScheduleTaskRepository`(sqlx 模式 CRUD,`NamedExecContext` 用于插入)
- `service.ScheduleService` 接口 + **TaskHandler hook**
  - `RegisterTaskType(taskType, handler)` 运行时注册
  - 3 个默认 handler:`email_report` / `snapshot` / `data_sync`,都调用 `recordTaskRun` 写审计到 `task.Config.last_runs`(保留最近 10 条)
  - **未注册 type 走降级**(打日志,不报错)
  - **handler 报错时**:`task.Status = "error"` + 返回错误
- `handler.ScheduleHandler` HTTP 层(8 个方法)
- 8 个新路由(CRUD + enable/disable/execute)
- **20 个新测试**(9 service + 7 handler + 4 middleware),全通过
- 新依赖:`github.com/robfig/cron/v3`

---

## 🎯 下一阶段工作建议(优先级排序)

| 优先级 | 任务 | 价值 | 工作量 |
|--------|------|------|--------|
| 🔴 高 | **前端页面审计**:`frontend/src/pages/*` 找占位/未实现页面 | 揭示完成度真相 | 20 分钟 |
| 🔴 高 | **service 测试覆盖**:chart/dataset/dashboard/share_link/row_permission 缺测试 | 已有 81% handler 覆盖,service 补齐到 80%+ | 60+ 分钟 |
| 🟡 中 | **GitHub Actions CI**:`.github/workflows/ci.yml` 跑 Go test + tsc/lint | 防止退化 | 15 分钟 |
| 🟡 中 | **补齐前端页面**:基于审计结果,接好 API + UI | 把"100% 对等"推到 100% | 60+ 分钟 |
| 🟢 低 | **CLAUDE.md 同步**:补"已完成里程碑"小节 | 文档维护 | 5 分钟 |
| 🟢 低 | **Dockerfile**:根目录无,生产部署需要 | 部署 | 30 分钟 |
| 🟢 低 | **README.md**:CLAUDE.md 给 Claude,README 给人类 | 入门 | 20 分钟 |

**建议第一刀**:**前端页面审计**(决定 D5 工作量)+ **service 测试覆盖**(纯加测试,零风险高收益)。

---

## 🪤 踩过的坑(避免重复)

### 1. **远程 main 是 sqlx 栈,不是 GORM**
不要用 `gorm.io/gorm` 写新代码。所有 `repository/*.go` 走 `github.com/jmoiron/sqlx` 模式:`db.GetContext` / `db.SelectContext` / `db.NamedExecContext` / `db.ExecContext`。

### 2. **Archive/ 是废弃代码**
根目录 `Archive/` 保存的是基于 Apache Calcite Avatica 的旧 SQL 引擎和 `start.sh`,**已废弃**,不要参照。`CLAUDE.md` 明确写了。当前代码在 `backend/` 和 `frontend/`。

### 3. **本地 main 与 origin/main 的"假冲突"**
之前本地 main 有 13 个旧 commit 是基于旧 Avatica 引擎写的,直接 merge 会爆 34 个冲突。**正确做法**:`git reset --hard origin/main` 把 main 对齐远端 sqlx 基线,然后把有用的工作在新分支上重做(我们用 `frontend-port` / `backend-port` 做的)。**不要尝试 cherry-pick 旧 commit 到新 main** — 它们是不同栈。

### 4. **`backend/server` 二进制**
本地可能有 95MB 的 `backend/server` Go 构建产物(被 .gitignore)。**不要** `rm` 它(本地还在用),但也**不要** commit 它。

### 5. **API 方法名是 `remove` 不是 `delete`**
前端 `dashboardAPI.remove(id)`、`datasetAPI.remove(id)` 等 — 后端实际接口是 `DELETE /resource/:id`,但 API 客户端用 `remove` 命名以避免和 `delete` 关键字混淆。

### 6. **Axios 自动解包**
`request.get<T>()` 返回 `Promise<T>`,**不**是 `Promise<AxiosResponse<T>>`。`Unwrap<T>` 泛型在 `request.ts` 已经处理。调用方直接 `const data = await request.get<Foo[]>('/foo')`。

### 7. **response 格式**
后端统一响应:`gin.H{"code": 200, "data": ...}` / `gin.H{"code": 4xx, "error": "..."}`。**不要**返回裸 `gin.H{"success": true}`。

### 8. **不要 panic**
`CLAUDE.md` 明确禁止 Go `panic()`。Service/Repository 方法必须返回 `error`,用 `fmt.Errorf("...: %w", err)` 包装。

### 9. **没有 CI 保护**
当前没有 `.github/workflows/`,push 不会自动跑测试。**任何重构前先本地跑 `go test ./...` + `npx tsc --noEmit`**。建议尽快加 CI(见下一阶段 D3)。

### 10. **Service 测试已有部分**
`internal/service/` 已有覆盖:`auth_service` / `cache_service` / `chart_service` / `dashboard_service` / `dataset_service` / `datasource_service` / `message_service` / `oper_log_service` / `role_service` / `row_permission_service` / `share_link_service` / `user_service` / `workbench_service`(均带 `_test.go`)。补的应该是这些里**不充分**的分支,不是从零写。

---

## 🛠 常用命令

```bash
# 后端
cd backend
go test ./...                          # 全部测试
go test -cover ./...                   # 覆盖率
go test -run TestXxx ./internal/...    # 单测
go build ./...                         # 编译检查
go fmt ./...                           # 格式化

# 前端
cd frontend
npx tsc --noEmit                       # 类型检查(零错误基线)
npm test                               # Vitest
npm run build                          # 生产构建
npm run lint                           # ESLint

# Docker 依赖
docker-compose up -d                   # MySQL + Redis
```

---

## 📁 关键文件路径(快速定位)

**后端**:
- 入口:`backend/cmd/server/main.go`
- 路由:`backend/api/v1/router.go` ← `Setup()` 函数
- 配置:`backend/configs/app.yaml`
- 迁移:`backend/migrations/`(001-018)
- 中间件:`backend/internal/middleware/`(auth, oper_log, common)
- Service 接口:`backend/internal/service/`
- Repository:`backend/internal/repository/`
- 模型:`backend/internal/model/`
- 引擎:`backend/internal/engine/`(原生 Go connector,非 Avatica)

**前端**:
- 入口:`frontend/src/main.tsx` + `App.tsx`
- API 封装:`frontend/src/api/request.ts`(带 Unwrap) + 各域 `*.ts`
- 类型:`frontend/src/types/`
- 页面:`frontend/src/pages/`
- 路由:`frontend/src/router/`
- 状态:Zustand(`frontend/src/store/`)+ React Query(在 hooks/ 或页面里)
- E2E:`frontend/e2e/`(Playwright)

---

## 📞 联系上下文

- 项目主页:https://github.com/zinohome/CozyInsight
- 原版参考:DataEase(Java + Vue),我们的目标是 100% 功能对等
- 提交人:ZhangJun(`git config user.name` 确认)
- Co-Authored-By 标记:Claude Code 协助的 commit 已带此标记

---

**TL;DR 给下一个 agent**:
1. 状态:`main` 干净,`origin/main` 领先 6 个 commit,全栈 sqlx + React 19
2. 刚完成:前端 TS 修复 + 后端 RateLimit/Schedule 系统移植
3. 下一步:**前端页面审计** + **service 测试覆盖** 是性价比最高的工作
4. 别踩坑:Archive/ 是废弃的,远程用 sqlx 不是 GORM,API 用 `remove` 不用 `delete`,统一响应格式 `{"code", "data"|"error"}`
5. 跑测试:`go test ./internal/...` + `npx tsc --noEmit`,目前全绿
