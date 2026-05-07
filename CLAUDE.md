# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

CozyInsight 是一个企业级开源 BI 数据可视化平台。它是 DataEase（Java + Vue）的**完整重构版本**，使用 **Go + React** 技术栈，要求与原版本 100% 功能对等。代码库分为 `backend/`（Go）和 `frontend/`（React/TypeScript）两部分。

## 常用命令

### 全栈开发

一键启动所有依赖和服务：
```bash
./start.sh
```

### 后端（Go）

前置条件：MySQL 8.0+、Redis 7+ 和 Avatica Server 必须正在运行。

```bash
cd backend

# 运行服务
go run cmd/server/main.go

# 构建二进制文件
go build -o server cmd/server/main.go

# 运行全部测试
go test ./...

# 运行测试并生成覆盖率报告
go test -cover ./...

# 运行单个包的测试
go test ./internal/service/...

# 运行单个测试
go test -run TestFunctionName ./path/to/package

# 格式化代码
go fmt ./...
```

### 前端（React + TypeScript）

```bash
cd frontend

# 开发服务器
npm run dev

# 生产构建
npm run build

# 代码检查
npm run lint
```

### Docker（生产部署）

```bash
cd deployments
docker-compose up -d
```

### Avatica Server（SQL 引擎）

```bash
cd backend/deployments/avatica
docker-compose up -d
```

## 架构设计

### 后端：分层架构，职责明确

后端采用严格的分层架构，核心代码位于 `backend/internal/`：

- **`handler/`** — HTTP 请求/响应处理（Gin handlers）。每个领域独立一个 handler（如 `chart_handler.go`、`datasource_handler.go`）。Handler 负责绑定 JSON 到 DTO、调用 Service、返回标准化响应。
- **`service/`** — 业务逻辑层。包含数据集、图表、仪表板、权限等核心业务规则。Service 为无状态设计，通过构造函数注入 Repository。
- **`repository/`** — 数据访问层，使用 GORM。每个 Repository 封装单一模型的数据库操作。所有查询必须传入 `context.Context` 并检查 `result.Error`。
- **`model/`** — GORM 结构体，直接映射原版 DataEase 的数据库 Schema。软删除模型必须包含 `gorm.DeletedAt`。
- **`engine/`** — Apache Calcite Avatica 客户端（`calcite_client.go`）和数据源连接器（`datasource_connector.go`）。这是跨数据源 SQL 查询的执行引擎。
- **`middleware/`** — 认证（`auth.go`）、权限检查、CORS、操作日志（`oper_log.go`）、panic 恢复。

公共/共享包位于 `backend/pkg/`：
- `config/` — 基于 Viper 的配置加载器
- `database/` — MySQL 连接初始化
- `cache/` — Redis 客户端封装
- `jwt/` — JWT 生成与验证
- `logger/` — Zap 结构化日志

**路由**定义在 `backend/api/v1/router.go`，在 `backend/cmd/server/main.go` 中组装。

**核心不变量**：API 路由、请求/响应格式、数据库 Schema 必须与原版 Java DataEase 保持兼容。

### 前端：组件 + 状态管理架构

前端是基于 Vite 的单页应用，源码位于 `frontend/src/`：

- **`api/`** — Axios 封装和领域级 API 模块（如 `datasource.ts`、`chart.ts`）。API 函数必须使用强类型的请求/响应接口。
- **`pages/`** — 路由级页面组件，按领域组织（`dashboard/`、`dataset/`、`chart/`、`system/`）。
- **`components/`** — 可复用 UI 组件（如 `Chart/`、`FilterPanel/`、`DragCanvas/`）。
- **`store/`** — Zustand 状态存储。每个 store 必须强类型化（禁止 `any`）。
- **`hooks/`** — 自定义 React Hooks，用于数据获取和业务逻辑（如 `useDataset`、`useChart`）。
- **`types/`** — 集中管理的 TypeScript 类型定义。
- **`router/`** — React Router 7 路由配置。

**核心不变量**：TypeScript 开启严格模式。避免 `any` 类型。昂贵计算使用 `useMemo`/`useCallback`，纯组件使用 `React.memo`。

### 认证与授权

- 基于 JWT 的认证，中间件保护路由。
- RBAC（基于角色的访问控制）+ 行级数据权限（通过 `row_permission_service.go` 注入 SQL WHERE 条件）。
- 操作日志中间件记录所有管理员操作。

### 数据流（图表渲染）

1. 前端选择数据源/数据集 → 调用 `datasetAPI`
2. 后端校验权限 → `dataset_service.go` → `dataset_repo.go`
3. 提交图表配置 → `chart_service.go` 通过 `calcite_client.go` 构建 SQL
4. SQL 通过数据源连接器执行 → 结果缓存到 Redis
5. 数据返回前端 → `@ant-design/charts` 渲染

## 关键规范（来自 `.cursorrules`）

- **Go 中禁止使用 `panic()`**。始终返回 `error`，并用 `fmt.Errorf(...: %w)` 包装。
- **Context 传递**：每个 Service/Repository 方法必须接受并传递 `context.Context`。
- **依赖注入**：使用构造函数（`NewXxxService(...)`），禁止包级全局变量。
- **TypeScript 中禁止使用 `any`**，除非确实不可避免。
- **功能对等 > 性能 > 代码优雅**：有冲突时，优先精确复刻原版 DataEase 行为。
- **Commit 格式**：使用 conventional commits（`feat(scope):`、`fix(scope):`、`refactor(scope):`）。

## 配置说明

后端配置位于 `backend/configs/app.yaml`：
- 服务默认运行在 `8100` 端口。
- MySQL、Redis、Avatica URL、JWT 密钥、Zap 日志配置均在此定义。
- 环境变量以 `DATAEASE_` 为前缀可覆盖 YAML 值（由 Viper 管理）。

## 重要文件路径

- 后端入口：`backend/cmd/server/main.go`
- 路由定义：`backend/api/v1/router.go`
- 前端入口：`frontend/src/main.tsx`
- 前端根组件：`frontend/src/App.tsx`
- 开发文档：`docs/DEVELOPMENT_GUIDE.md`、`docs/API.md`
