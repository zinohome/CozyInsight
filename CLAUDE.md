# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

CozyInsight 是一个企业级开源 BI 数据可视化平台，是 DataEase（Java + Vue）的**完整重构版本**，使用 **Go + React** 技术栈，目标是与原版本 100% 功能对等。代码库分为 `backend/`（Go 1.25）和 `frontend/`（React 19 / TypeScript / Vite）两部分。

> 注意：仓库根目录下的 `Archive/` 保存的是早期实现（含基于 Apache Calcite Avatica 的旧 SQL 引擎与 `start.sh`），**已废弃**。当前活动代码在 `backend/` 和 `frontend/`，查询引擎已改为原生 Go 连接器（见下文 `engine/`）。不要参照 `Archive/` 中的内容。

## 常用命令

### 后端（Go）

前置条件：MySQL 8.0+ 和 Redis 7+ 正在运行。可用根目录的 `docker-compose.yml` 启动依赖（见下）。

```bash
cd backend

go run cmd/server/main.go              # 运行服务（默认 :8100）
go build -o server cmd/server/main.go  # 构建二进制

go test ./...                          # 运行全部测试
go test -cover ./...                   # 测试 + 覆盖率
go test ./internal/service/...         # 单个包
go test -run TestFunctionName ./internal/service   # 单个测试

go fmt ./...                           # 格式化
```

### 前端（React + TypeScript）

```bash
cd frontend

npm run dev      # 开发服务器（Vite）
npm run build    # 生产构建（先 tsc -b 类型检查，再 vite build）
npm run lint     # ESLint
npm test         # 运行 Vitest 测试（vitest run）
npm run preview  # 预览生产构建
```

### Docker（启动依赖 / 部署）

根目录 `docker-compose.yml` 定义了 MySQL（自动加载 `backend/migrations/` 中的初始化 SQL）和 Redis：

```bash
docker-compose up -d
```

## 架构设计

### 后端：分层架构，职责明确

核心代码位于 `backend/internal/`，调用方向为 `handler → service → repository`：

- **`handler/`** — HTTP 请求/响应处理（Gin）。每个领域一个 handler（`chart_handler.go`、`datasource_handler.go` 等）。负责绑定 JSON 到 DTO、调用 Service、返回标准化响应。`common.go` 提供统一响应封装。
- **`dto/`** — 请求/响应数据传输对象，handler 与 service 之间的契约。
- **`service/`** — 业务逻辑层（无状态，构造函数注入 Repository）。包含数据集、图表、仪表板、RBAC、行级权限、缓存、分享链接、操作日志等。`errors.go` 定义领域错误。
- **`repository/`** — 数据访问层（GORM）。每个 Repository 封装单一模型，方法必须传入 `context.Context` 并检查 `result.Error`。
- **`model/`** — GORM 结构体，映射 DataEase 数据库 Schema。软删除模型含 `gorm.DeletedAt`。
- **`engine/`** — **跨数据源查询引擎（原生 Go，非 Avatica）**：
  - `query_engine.go` — 根据图表配置（`ChartQueryConfig`：维度/指标/过滤/排序）构建 SQL。
  - `connector.go` / `file_connector.go` — `DatasourceConnector` 接口及实现，通过 `database/sql` 驱动连接 MySQL、PostgreSQL、ClickHouse、SQLite、SQL Server，以及文件数据源。
  - `connector_pool.go` — 连接池管理。
- **`middleware/`** — JWT 认证（`auth.go`）、权限检查、CORS、操作日志（`oper_log.go`）、panic 恢复。
- **`testutil/`** — 测试辅助（内存/临时数据库等）。

公共/共享包位于 `backend/pkg/`：`config/`（Viper 配置加载）、`database/`（MySQL/GORM 初始化）、`cache/`（Redis 封装）、`jwt/`（JWT 签发/校验）、`logger/`（Zap）。

**路由**定义在 `backend/api/v1/router.go`（统一前缀 `/api/v1`，区分公开路由、`authd` 鉴权路由、`admin` 管理员路由），在 `backend/cmd/server/main.go` 中装配依赖。

**数据库迁移**位于 `backend/migrations/`，按序号命名（`001_init.sql` … `011_share_links.sql`），由 docker-compose 的 MySQL 容器自动执行。

**核心不变量**：API 路由、请求/响应格式、数据库 Schema 必须与原版 Java DataEase 保持兼容。

### 前端：组件 + 状态管理架构

基于 Vite 的 SPA，源码位于 `frontend/src/`：

- **`api/`** — Axios 封装（`request.ts`）和领域级 API 模块（`chart.ts`、`datasource.ts` 等，每个模块配套 `.test.ts`）。函数使用强类型请求/响应接口。**注意 base URL 已含 `/api/v1`，各模块路径不要重复该前缀。**
- **`pages/`** — 路由级页面，按领域组织：`dashboard/`、`dataset/`、`chart/`、`datasource/`、`screen/`（大屏）、`workbench/`、`system/`、`login/`、`profile/`、`share/`、`404/`。
- **`components/`** — 可复用 UI 组件。
- **`store/`** — Zustand 状态存储，强类型化（禁止 `any`）。
- **`hooks/`** — 数据获取与业务逻辑 Hooks。
- **`types/`** — 集中的 TypeScript 类型定义。
- **`router/`** — React Router 7 路由配置。

技术栈：React 19、Ant Design 6 + `@ant-design/charts`（图表渲染）、TanStack React Query（服务端状态/数据获取）、Zustand（客户端状态）、`react-grid-layout` / `react-rnd`（仪表板拖拽布局）。测试用 Vitest + Testing Library（`src/test-setup.ts`）。

**核心不变量**：TypeScript 严格模式，避免 `any`。昂贵计算用 `useMemo`/`useCallback`，纯组件用 `React.memo`。

### 认证与授权

- 基于 JWT 的认证，中间件保护路由。
- RBAC（角色权限）+ 行级数据权限（`row_permission_service.go` 向查询注入 SQL WHERE 条件）。
- 操作日志中间件记录管理员操作。

### 数据流（图表渲染）

1. 前端选择数据源/数据集 → 调用 `api/dataset.ts`。
2. 后端校验权限 → `dataset_service.go` → `dataset_repo.go`。
3. 提交图表配置 → `chart_service.go` 通过 `engine/query_engine.go` 构建 SQL。
4. SQL 经 `engine/connector.go` 在目标数据源执行 → 结果经 `cache_service.go` 缓存到 Redis。
5. 数据返回前端 → `@ant-design/charts` 渲染。

## 关键编码规范

- **Go 中禁止 `panic()`**。始终返回 `error`，用 `fmt.Errorf("...: %w", err)` 包装。
- **Context 传递**：每个 Service/Repository 方法必须接受并传递 `context.Context`。
- **依赖注入**：使用构造函数（`NewXxxService(...)`），禁止包级全局变量。
- **TypeScript 中禁止 `any`**，除非确实不可避免。
- **功能对等 > 性能 > 代码优雅**：冲突时优先精确复刻原版 DataEase 行为。
- **Commit 格式**：conventional commits（`feat(scope):`、`fix(scope):`、`refactor(scope):`）。

## 配置说明

后端配置位于 `backend/configs/app.yaml`：

- 服务默认运行在 `8100` 端口。
- 含 MySQL、Redis、JWT 密钥、Zap 日志配置（由 Viper 加载）。
- 配置值支持 `${VAR:-default}` 形式的环境变量覆盖，前缀为 **`COZYINSIGHT_`**（如 `COZYINSIGHT_DATABASE_HOST`、`COZYINSIGHT_JWT_SECRET`）。

## 重要文件路径

- 后端入口：`backend/cmd/server/main.go`
- 路由定义：`backend/api/v1/router.go`
- 数据库迁移：`backend/migrations/`
- 后端配置：`backend/configs/app.yaml`
- 前端入口：`frontend/src/main.tsx`
- 前端根组件：`frontend/src/App.tsx`
- 前端 Axios 封装：`frontend/src/api/request.ts`
