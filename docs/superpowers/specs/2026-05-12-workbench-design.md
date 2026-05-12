# 工作台（首页）页面设计文档

> 目标：将当前路由 `/` 的"建设中"替换为功能完整的工作台页面，对标 DataEase 工作台。

## 1. 概述

工作台是用户登录后的落地页，提供：
- 快捷创建入口 — 一键新建各类资源
- 个人资源统计 — 实时展示用户拥有的资源数量
- 最近访问 — 按时间倒序展示最近查看的仪表板/大屏
- 我的收藏 — 用户收藏的资源列表

## 2. 页面布局

```
┌─────────────────────────────────────────────────────────────┐
│  欢迎回来，{昵称}，祝您开心每一天！                              │
│  [新建数据源] [新建数据集] [新建图表] [新建仪表板] [新建数据大屏]  │
├─────────────────────────────────────────────────────────────┤
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐                │
│  │数据源 │ │数据集 │ │ 图表 │ │仪表板│ │数据大屏│              │
│  │  12  │ │  8   │ │  15  │ │  3   │ │  2   │              │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘                │
├─────────────────────────────────────────────────────────────┤
│  [最近访问] [我的收藏]                                        │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  标题           │ 类型      │ 最后访问时间              │ │
│  │  销售月报        │ 仪表板    │ 2026-05-12 10:30         │ │
│  │  实时监控大屏     │ 数据大屏  │ 2026-05-11 18:20         │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 2.1 欢迎区 + 快捷创建

- 左侧：问候语，使用用户 `nickName`，默认显示"欢迎回来，祝您开心每一天！"
- 右侧：5个快捷创建按钮，图标 + 文字，点击直接跳转对应创建页面
  - 数据源 → `/datasource`
  - 数据集 → `/dataset`
  - 图表 → `/chart`
  - 仪表板 → `/dashboard`（打开创建弹窗）
  - 数据大屏 → `/dashboard`（打开创建弹窗，类型为 screen）

### 2.2 统计卡片区

5个 `Card` 组件，使用 Ant Design `Statistic` 展示数量：
- 数据源数量（调用 `datasourceAPI.list()` 取总数）
- 数据集数量（调用 `datasetAPI.list()` 取总数）
- 图表数量（调用 `chartAPI.list()` 取总数）
- 仪表板数量（调用 `dashboardAPI.list()` 筛选 `type=dashboard`）
- 数据大屏数量（调用 `dashboardAPI.list()` 筛选 `type=screen`）

每个卡片可点击，跳转到对应列表页。

### 2.3 资源标签页

使用 Ant Design `Tabs` 组件：

**Tab 1: 最近访问**
- `Table` 展示最近访问的资源
- 列：标题、类型（仪表板/数据大屏）、最后访问时间
- 点击行跳转查看页面
- 空状态提示"暂无最近访问记录"

**Tab 2: 我的收藏**
- `Table` 展示收藏的资源
- 列：标题、类型、收藏时间、操作（取消收藏）
- 点击行跳转查看页面
- 空状态提示"暂无收藏资源"

## 3. 后端设计

### 3.1 数据库表

```sql
-- 浏览记录表
CREATE TABLE IF NOT EXISTS recent_views (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    resource_type VARCHAR(32) NOT NULL COMMENT 'dashboard|screen',
    resource_id BIGINT UNSIGNED NOT NULL,
    visited_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_time (user_id, visited_at),
    INDEX idx_user_resource (user_id, resource_type, resource_id),
    UNIQUE KEY uk_user_resource (user_id, resource_type, resource_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='浏览记录表';

-- 收藏表
CREATE TABLE IF NOT EXISTS favorites (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    resource_type VARCHAR(32) NOT NULL COMMENT 'dashboard|screen|chart|dataset|datasource',
    resource_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_time (user_id, created_at),
    UNIQUE KEY uk_user_resource (user_id, resource_type, resource_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收藏表';
```

### 3.2 API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/workbench/stats` | 获取当前用户资源统计 |
| GET | `/api/v1/workbench/recent` | 获取最近访问列表（最近20条）|
| POST | `/api/v1/workbench/recent` | 记录一次访问（dashboard/view 和 screen/view 时调用）|
| GET | `/api/v1/workbench/favorites` | 获取收藏列表 |
| POST | `/api/v1/workbench/favorites` | 添加收藏 |
| DELETE | `/api/v1/workbench/favorites/:type/:id` | 取消收藏 |

### 3.3 DTO

```go
// WorkbenchStatsResponse 资源统计响应
type WorkbenchStatsResponse struct {
	DatasourceCount int64 `json:"datasourceCount"`
	DatasetCount    int64 `json:"datasetCount"`
	ChartCount      int64 `json:"chartCount"`
	DashboardCount  int64 `json:"dashboardCount"`
	ScreenCount     int64 `json:"screenCount"`
}

// RecentViewItem 最近访问项
type RecentViewItem struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	VisitedAt   time.Time `json:"visitedAt"`
}

// FavoriteItem 收藏项
type FavoriteItem struct {
	ID        uint64    `json:"id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
}

// RecordVisitRequest 记录访问请求
type RecordVisitRequest struct {
	ResourceType string `json:"resourceType" binding:"required,oneof=dashboard screen"`
	ResourceID   uint64 `json:"resourceId" binding:"required"`
}
```

### 3.4 后端文件

| 文件 | 职责 |
|------|------|
| `backend/migrations/015_workbench.sql` | 创建 `recent_views` 和 `favorites` 表 |
| `backend/internal/dto/workbench.go` | 工作台相关 DTO |
| `backend/internal/model/recent_view.go` | `RecentView` 模型 |
| `backend/internal/model/favorite.go` | `Favorite` 模型 |
| `backend/internal/repository/workbench_repo.go` | `WorkbenchRepository` — 统计查询、浏览记录、收藏管理 |
| `backend/internal/service/workbench_service.go` | `WorkbenchService` — 业务逻辑 |
| `backend/internal/handler/workbench_handler.go` | `WorkbenchHandler` — HTTP 处理 |
| `backend/internal/handler/workbench_handler_test.go` | 测试 |
| `backend/api/v1/router.go` | 注册工作台路由 |

## 4. 前端设计

### 4.1 文件

| 文件 | 职责 |
|------|------|
| `frontend/src/api/workbench.ts` | 工作台 API 封装 |
| `frontend/src/types/workbench.ts` | 工作台类型定义 |
| `frontend/src/pages/workbench/index.tsx` | 工作台页面主组件 |
| `frontend/src/pages/dashboard/DashboardView.tsx` | 修改：访问时记录浏览历史 |
| `frontend/src/pages/screen/ScreenView.tsx` | 修改：访问时记录浏览历史 |

### 4.2 类型定义

```typescript
// frontend/src/types/workbench.ts

export interface WorkbenchStats {
  datasourceCount: number
  datasetCount: number
  chartCount: number
  dashboardCount: number
  screenCount: number
}

export interface RecentViewItem {
  id: number
  title: string
  type: 'dashboard' | 'screen'
  visitedAt: string
}

export interface FavoriteItem {
  id: number
  title: string
  type: string
  createdAt: string
}

export interface RecordVisitRequest {
  resourceType: 'dashboard' | 'screen'
  resourceId: number
}
```

### 4.3 API 封装

```typescript
// frontend/src/api/workbench.ts
import request from './request'
import type { WorkbenchStats, RecentViewItem, FavoriteItem, RecordVisitRequest } from '@/types/workbench'

export const workbenchAPI = {
  getStats: () => request.get<WorkbenchStats>('/workbench/stats'),
  getRecent: () => request.get<RecentViewItem[]>('/workbench/recent'),
  recordVisit: (data: RecordVisitRequest) => request.post<void>('/workbench/recent', data),
  getFavorites: () => request.get<FavoriteItem[]>('/workbench/favorites'),
  addFavorite: (type: string, resourceId: number) =>
    request.post<void>('/workbench/favorites', { resourceType: type, resourceId }),
  removeFavorite: (type: string, resourceId: number) =>
    request.delete<void>(`/workbench/favorites/${type}/${resourceId}`),
}
```

## 5. 数据流

### 5.1 记录浏览历史

当用户访问仪表板/大屏查看页面时：

```
DashboardView.tsx / ScreenView.tsx
  → useEffect 中调用 workbenchAPI.recordVisit({ resourceType, resourceId })
  → POST /api/v1/workbench/recent
  → WorkbenchHandler.RecordVisit
  → WorkbenchService.RecordVisit(userID, type, resourceID)
  → WorkbenchRepository.UpsertRecentView (INSERT ... ON DUPLICATE KEY UPDATE)
```

### 5.2 工作台页面加载

```
WorkbenchPage.tsx
  → useEffect 并行发起三个请求：
    - workbenchAPI.getStats() → 统计卡片数据
    - workbenchAPI.getRecent() → 最近访问列表
    - workbenchAPI.getFavorites() → 收藏列表
```

## 6. 与现有功能的集成点

1. **Layout 菜单选中状态**：`Layout/index.tsx` 中 `selectedKeys` 当前使用 `location.pathname`，工作台路由 `/` 需要正确高亮。
2. **DashboardView / ScreenView**：在 `useEffect` 中增加 `recordVisit` 调用。
3. **router**：替换 `/` 路由的占位符为 `WorkbenchPage` 组件。

## 7. 范围边界

**本次实现包含**：
- 工作台页面 UI（欢迎区 + 快捷创建 + 统计 + 最近访问 + 收藏）
- 后端 API（统计、浏览记录、收藏）
- 查看页面自动记录浏览历史

**本次不包含**：
- 仪表盘上的图表预览（缩略图）
- "我的分享"标签页（已有独立的分享管理）
- 模板市场功能
- 消息通知中心
