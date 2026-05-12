# 工作台（首页）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将路由 `/` 的"建设中"占位符替换为功能完整的工作台页面，包含快捷创建入口、资源统计、最近访问列表和收藏列表。

**Architecture:** 工作台复用现有的资源列表 API（datasource/dataset/chart/dashboard）做统计，新增独立的 `workbench` 模块处理浏览记录和收藏。后端采用标准的 Repository → Service → Handler 分层。前端使用 Ant Design Card + Statistic + Tabs 组件构建页面。

**Tech Stack:** Go 1.25 + Gin + sqlx + sqlmock; React 19 + TypeScript + Ant Design 6 + Zustand + Vite

---

## 文件结构

### 后端（新建/修改）

| 文件 | 类型 | 职责 |
|------|------|------|
| `backend/migrations/015_workbench.sql` | 新建 | 创建 `recent_views` 和 `favorites` 表 |
| `backend/internal/model/recent_view.go` | 新建 | `RecentView` GORM 模型 |
| `backend/internal/model/favorite.go` | 新建 | `Favorite` GORM 模型 |
| `backend/internal/dto/workbench.go` | 新建 | 工作台 DTO |
| `backend/internal/repository/workbench_repo.go` | 新建 | 浏览记录 + 收藏的数据访问 |
| `backend/internal/service/workbench_service.go` | 新建 | 工作台业务逻辑 |
| `backend/internal/handler/workbench_handler.go` | 新建 | HTTP handler |
| `backend/internal/handler/workbench_handler_test.go` | 新建 | Handler 测试 |
| `backend/api/v1/router.go` | 修改 | 注册工作台路由 |

### 前端（新建/修改）

| 文件 | 类型 | 职责 |
|------|------|------|
| `frontend/src/types/workbench.ts` | 新建 | 工作台类型定义 |
| `frontend/src/api/workbench.ts` | 新建 | 工作台 API 封装 |
| `frontend/src/pages/workbench/index.tsx` | 新建 | 工作台页面 |
| `frontend/src/router/index.tsx` | 修改 | 替换 `/` 路由 |
| `frontend/src/pages/dashboard/DashboardView.tsx` | 修改 | 访问时记录浏览历史 |
| `frontend/src/pages/screen/ScreenView.tsx` | 修改 | 访问时记录浏览历史 |

---

## Task 1: 数据库迁移

**Files:**
- Create: `backend/migrations/015_workbench.sql`

- [ ] **Step 1: 编写迁移文件**

```sql
-- backend/migrations/015_workbench.sql

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

- [ ] **Step 2: Commit**

```bash
git add backend/migrations/015_workbench.sql
git commit -m "feat(migration): add recent_views and favorites tables for workbench"
```

---

## Task 2: Model 层

**Files:**
- Create: `backend/internal/model/recent_view.go`
- Create: `backend/internal/model/favorite.go`

- [ ] **Step 1: 创建 RecentView 模型**

```go
package model

import "time"

type RecentView struct {
	ID           uint64    `db:"id" json:"id"`
	UserID       uint64    `db:"user_id" json:"userId"`
	ResourceType string    `db:"resource_type" json:"resourceType"`
	ResourceID   uint64    `db:"resource_id" json:"resourceId"`
	VisitedAt    time.Time `db:"visited_at" json:"visitedAt"`
}
```

- [ ] **Step 2: 创建 Favorite 模型**

```go
package model

import "time"

type Favorite struct {
	ID           uint64    `db:"id" json:"id"`
	UserID       uint64    `db:"user_id" json:"userId"`
	ResourceType string    `db:"resource_type" json:"resourceType"`
	ResourceID   uint64    `db:"resource_id" json:"resourceId"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/model/recent_view.go backend/internal/model/favorite.go
git commit -m "feat(model): add RecentView and Favorite models"
```

---

## Task 3: DTO 层

**Files:**
- Create: `backend/internal/dto/workbench.go`

- [ ] **Step 1: 创建工作台 DTO**

```go
package dto

import "time"

// WorkbenchStatsResponse 工作台资源统计响应
type WorkbenchStatsResponse struct {
	DatasourceCount int64 `json:"datasourceCount"`
	DatasetCount    int64 `json:"datasetCount"`
	ChartCount      int64 `json:"chartCount"`
	DashboardCount  int64 `json:"dashboardCount"`
	ScreenCount     int64 `json:"screenCount"`
}

// RecentViewItem 最近访问项（联表查询结果）
type RecentViewItem struct {
	ID        uint64    `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	Type      string    `db:"type" json:"type"`
	VisitedAt time.Time `db:"visited_at" json:"visitedAt"`
}

// FavoriteItem 收藏项（联表查询结果）
type FavoriteItem struct {
	ID        uint64    `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	Type      string    `db:"type" json:"type"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

// RecordVisitRequest 记录访问请求
type RecordVisitRequest struct {
	ResourceType string `json:"resourceType" binding:"required,oneof=dashboard screen"`
	ResourceID   uint64 `json:"resourceId" binding:"required"`
}

// AddFavoriteRequest 添加收藏请求
type AddFavoriteRequest struct {
	ResourceType string `json:"resourceType" binding:"required"`
	ResourceID   uint64 `json:"resourceId" binding:"required"`
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/dto/workbench.go
git commit -m "feat(dto): add workbench DTOs"
```

---

## Task 4: Repository 层

**Files:**
- Create: `backend/internal/repository/workbench_repo.go`

- [ ] **Step 1: 创建 Repository**

```go
package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/dto"
)

type WorkbenchRepository struct {
	db *sqlx.DB
}

func NewWorkbenchRepository(db *sqlx.DB) *WorkbenchRepository {
	return &WorkbenchRepository{db: db}
}

// CountByCreatedBy 统计某用户创建的资源数量
func (r *WorkbenchRepository) CountByCreatedBy(ctx context.Context, table string, userID uint64, extraWhere string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE created_by = ? AND deleted_at IS NULL", table)
	if extraWhere != "" {
		query += " AND " + extraWhere
	}
	var count int64
	if err := r.db.GetContext(ctx, &count, query, userID); err != nil {
		return 0, fmt.Errorf("count %s failed: %w", table, err)
	}
	return count, nil
}

// UpsertRecentView 插入或更新浏览记录（ON DUPLICATE KEY UPDATE）
func (r *WorkbenchRepository) UpsertRecentView(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	query := `INSERT INTO recent_views (user_id, resource_type, resource_id, visited_at)
			  VALUES (?, ?, ?, NOW())
			  ON DUPLICATE KEY UPDATE visited_at = NOW()`
	if _, err := r.db.ExecContext(ctx, query, userID, resourceType, resourceID); err != nil {
		return fmt.Errorf("upsert recent view failed: %w", err)
	}
	return nil
}

// ListRecentViews 查询用户最近访问的资源（联表查 dashboards 取标题）
func (r *WorkbenchRepository) ListRecentViews(ctx context.Context, userID uint64, limit int) ([]dto.RecentViewItem, error) {
	query := `SELECT d.id, d.title, d.type, rv.visited_at
			  FROM recent_views rv
			  JOIN dashboards d ON rv.resource_id = d.id AND d.deleted_at IS NULL
			  WHERE rv.user_id = ? AND rv.resource_type = d.type
			  ORDER BY rv.visited_at DESC
			  LIMIT ?`
	var list []dto.RecentViewItem
	if err := r.db.SelectContext(ctx, &list, query, userID, limit); err != nil {
		return nil, fmt.Errorf("list recent views failed: %w", err)
	}
	return list, nil
}

// AddFavorite 添加收藏
func (r *WorkbenchRepository) AddFavorite(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	query := `INSERT INTO favorites (user_id, resource_type, resource_id) VALUES (?, ?, ?)
			  ON DUPLICATE KEY UPDATE created_at = NOW()`
	if _, err := r.db.ExecContext(ctx, query, userID, resourceType, resourceID); err != nil {
		return fmt.Errorf("add favorite failed: %w", err)
	}
	return nil
}

// DeleteFavorite 取消收藏
func (r *WorkbenchRepository) DeleteFavorite(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	query := `DELETE FROM favorites WHERE user_id = ? AND resource_type = ? AND resource_id = ?`
	if _, err := r.db.ExecContext(ctx, query, userID, resourceType, resourceID); err != nil {
		return fmt.Errorf("delete favorite failed: %w", err)
	}
	return nil
}

// ListFavorites 查询用户收藏列表（联表查 dashboards 取标题）
func (r *WorkbenchRepository) ListFavorites(ctx context.Context, userID uint64) ([]dto.FavoriteItem, error) {
	query := `SELECT d.id, d.title, d.type, f.created_at
			  FROM favorites f
			  JOIN dashboards d ON f.resource_id = d.id AND d.deleted_at IS NULL
			  WHERE f.user_id = ? AND f.resource_type IN ('dashboard', 'screen') AND d.type = f.resource_type
			  ORDER BY f.created_at DESC`
	var list []dto.FavoriteItem
	if err := r.db.SelectContext(ctx, &list, query, userID); err != nil {
		return nil, fmt.Errorf("list favorites failed: %w", err)
	}
	return list, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/repository/workbench_repo.go
git commit -m "feat(repo): add WorkbenchRepository with recent views and favorites"
```

---

## Task 5: Service 层

**Files:**
- Create: `backend/internal/service/workbench_service.go`

- [ ] **Step 1: 创建 Service**

```go
package service

import (
	"context"
	"fmt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/repository"
)

type WorkbenchService struct {
	repo *repository.WorkbenchRepository
}

func NewWorkbenchService(repo *repository.WorkbenchRepository) *WorkbenchService {
	return &WorkbenchService{repo: repo}
}

// GetStats 获取当前用户的资源统计
func (s *WorkbenchService) GetStats(ctx context.Context, userID uint64) (*dto.WorkbenchStatsResponse, error) {
	dsCount, err := s.repo.CountByCreatedBy(ctx, "datasources", userID, "")
	if err != nil {
		return nil, fmt.Errorf("get datasource count: %w", err)
	}
	datasetCount, err := s.repo.CountByCreatedBy(ctx, "datasets", userID, "")
	if err != nil {
		return nil, fmt.Errorf("get dataset count: %w", err)
	}
	chartCount, err := s.repo.CountByCreatedBy(ctx, "charts", userID, "")
	if err != nil {
		return nil, fmt.Errorf("get chart count: %w", err)
	}
	dbCount, err := s.repo.CountByCreatedBy(ctx, "dashboards", userID, "type = 'dashboard'")
	if err != nil {
		return nil, fmt.Errorf("get dashboard count: %w", err)
	}
	screenCount, err := s.repo.CountByCreatedBy(ctx, "dashboards", userID, "type = 'screen'")
	if err != nil {
		return nil, fmt.Errorf("get screen count: %w", err)
	}
	return &dto.WorkbenchStatsResponse{
		DatasourceCount: dsCount,
		DatasetCount:    datasetCount,
		ChartCount:      chartCount,
		DashboardCount:  dbCount,
		ScreenCount:     screenCount,
	}, nil
}

// RecordVisit 记录一次资源访问
func (s *WorkbenchService) RecordVisit(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	return s.repo.UpsertRecentView(ctx, userID, resourceType, resourceID)
}

// ListRecentViews 获取最近访问列表
func (s *WorkbenchService) ListRecentViews(ctx context.Context, userID uint64) ([]dto.RecentViewItem, error) {
	return s.repo.ListRecentViews(ctx, userID, 20)
}

// AddFavorite 添加收藏
func (s *WorkbenchService) AddFavorite(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	return s.repo.AddFavorite(ctx, userID, resourceType, resourceID)
}

// DeleteFavorite 取消收藏
func (s *WorkbenchService) DeleteFavorite(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	return s.repo.DeleteFavorite(ctx, userID, resourceType, resourceID)
}

// ListFavorites 获取收藏列表
func (s *WorkbenchService) ListFavorites(ctx context.Context, userID uint64) ([]dto.FavoriteItem, error) {
	return s.repo.ListFavorites(ctx, userID)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/service/workbench_service.go
git commit -m "feat(service): add WorkbenchService with stats, recent views and favorites"
```

---

## Task 6: Handler 层

**Files:**
- Create: `backend/internal/handler/workbench_handler.go`

- [ ] **Step 1: 创建 Handler**

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/middleware"
	"cozy-insight/internal/service"
)

type WorkbenchHandler struct {
	service *service.WorkbenchService
}

func NewWorkbenchHandler(service *service.WorkbenchService) *WorkbenchHandler {
	return &WorkbenchHandler{service: service}
}

// GetStats 获取工作台统计
func (h *WorkbenchHandler) GetStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	stats, err := h.service.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetRecentViews 获取最近访问列表
func (h *WorkbenchHandler) GetRecentViews(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.service.ListRecentViews(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

// RecordVisit 记录访问
func (h *WorkbenchHandler) RecordVisit(c *gin.Context) {
	var req dto.RecordVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.service.RecordVisit(c.Request.Context(), userID, req.ResourceType, req.ResourceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200})
}

// GetFavorites 获取收藏列表
func (h *WorkbenchHandler) GetFavorites(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.service.ListFavorites(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

// AddFavorite 添加收藏
func (h *WorkbenchHandler) AddFavorite(c *gin.Context) {
	var req dto.AddFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.service.AddFavorite(c.Request.Context(), userID, req.ResourceType, req.ResourceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200})
}

// DeleteFavorite 取消收藏
func (h *WorkbenchHandler) DeleteFavorite(c *gin.Context) {
	resourceType := c.Param("type")
	resourceID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.service.DeleteFavorite(c.Request.Context(), userID, resourceType, resourceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200})
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/handler/workbench_handler.go
git commit -m "feat(handler): add WorkbenchHandler for stats, recent views and favorites"
```

---

## Task 7: Handler 测试

**Files:**
- Create: `backend/internal/handler/workbench_handler_test.go`

- [ ] **Step 1: 编写测试（TDD 风格）**

```go
package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/middleware"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/service"
)

func setupWorkbenchHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewWorkbenchRepository(sqlxDB)
	svc := service.NewWorkbenchService(repo)
	h := NewWorkbenchHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(1))
		c.Next()
	})
	r.GET("/workbench/stats", h.GetStats)
	r.GET("/workbench/recent", h.GetRecentViews)
	r.POST("/workbench/recent", h.RecordVisit)
	r.GET("/workbench/favorites", h.GetFavorites)
	r.POST("/workbench/favorites", h.AddFavorite)
	r.DELETE("/workbench/favorites/:type/:id", h.DeleteFavorite)

	return r, mock
}

func TestWorkbenchHandler_GetStats(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM datasources").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM datasets").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM charts").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM dashboards").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM dashboards").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/workbench/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(200), resp["code"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(3), data["datasourceCount"])
	assert.Equal(t, float64(5), data["datasetCount"])
	assert.Equal(t, float64(8), data["chartCount"])
	assert.Equal(t, float64(2), data["dashboardCount"])
	assert.Equal(t, float64(1), data["screenCount"])
}

func TestWorkbenchHandler_GetRecentViews(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	now := time.Now()
	cols := []string{"id", "title", "type", "visited_at"}
	mock.ExpectQuery("SELECT d.id, d.title, d.type, rv.visited_at").
		WithArgs(1, 20).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, "Sales", "dashboard", now))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/workbench/recent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(200), resp["code"])
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
	item := data[0].(map[string]interface{})
	assert.Equal(t, "Sales", item["title"])
	assert.Equal(t, "dashboard", item["type"])
}

func TestWorkbenchHandler_RecordVisit(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	mock.ExpectExec("INSERT INTO recent_views").
		WithArgs(1, "dashboard", uint64(5)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(dto.RecordVisitRequest{ResourceType: "dashboard", ResourceID: 5})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/workbench/recent", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkbenchHandler_AddFavorite(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	mock.ExpectExec("INSERT INTO favorites").
		WithArgs(1, "dashboard", uint64(10)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(dto.AddFavoriteRequest{ResourceType: "dashboard", ResourceID: 10})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/workbench/favorites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkbenchHandler_DeleteFavorite(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	mock.ExpectExec("DELETE FROM favorites").
		WithArgs(1, "screen", uint64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/workbench/favorites/screen/3", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkbenchHandler_GetFavorites(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	now := time.Now()
	cols := []string{"id", "title", "type", "created_at"}
	mock.ExpectQuery("SELECT d.id, d.title, d.type, f.created_at").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(2, "KPI", "dashboard", now))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/workbench/favorites", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(200), resp["code"])
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
}
```

- [ ] **Step 2: 运行测试**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go test ./internal/handler/... -v -run "Workbench"
```

Expected: All tests PASS.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/workbench_handler_test.go
git commit -m "test(handler): add workbench handler tests"
```

---

## Task 8: 注册路由

**Files:**
- Modify: `backend/api/v1/router.go`

- [ ] **Step 1: 在工作台 handler 和 service 的创建位置添加代码**

在 `backend/api/v1/router.go` 中 `dashboardHandler` 创建之后、第一个 `api.Group` 之前，添加：

```go
	workbenchRepo := repository.NewWorkbenchRepository(db)
	workbenchService := service.NewWorkbenchService(workbenchRepo)
	workbenchHandler := handler.NewWorkbenchHandler(workbenchService)
```

在 `authd` 路由组内，添加：

```go
				authd.GET("/workbench/stats", workbenchHandler.GetStats)
				authd.GET("/workbench/recent", workbenchHandler.GetRecentViews)
				authd.POST("/workbench/recent", workbenchHandler.RecordVisit)
				authd.GET("/workbench/favorites", workbenchHandler.GetFavorites)
				authd.POST("/workbench/favorites", workbenchHandler.AddFavorite)
				authd.DELETE("/workbench/favorites/:type/:id", workbenchHandler.DeleteFavorite)
```

- [ ] **Step 2: Commit**

```bash
git add backend/api/v1/router.go
git commit -m "feat(router): register workbench routes"
```

---

## Task 9: 前端类型定义

**Files:**
- Create: `frontend/src/types/workbench.ts`

- [ ] **Step 1: 创建类型文件**

```typescript
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

- [ ] **Step 2: Commit**

```bash
git add frontend/src/types/workbench.ts
git commit -m "feat(types): add workbench types"
```

---

## Task 10: 前端 API 封装

**Files:**
- Create: `frontend/src/api/workbench.ts`

- [ ] **Step 1: 创建 API 文件**

```typescript
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

- [ ] **Step 2: Commit**

```bash
git add frontend/src/api/workbench.ts
git commit -m "feat(api): add workbenchAPI"
```

---

## Task 11: 工作台页面

**Files:**
- Create: `frontend/src/pages/workbench/index.tsx`

- [ ] **Step 1: 创建工作台页面**

```tsx
import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Card,
  Statistic,
  Tabs,
  Table,
  Button,
  Space,
  Tag,
  Empty,
  message,
} from 'antd'
import {
  DatabaseOutlined,
  TableOutlined,
  BarChartOutlined,
  LayoutOutlined,
  DesktopOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import { workbenchAPI } from '@/api/workbench'
import { useAuthStore } from '@/store/auth'
import type { WorkbenchStats, RecentViewItem, FavoriteItem } from '@/types/workbench'

const { TabPane } = Tabs

const statCards = [
  { key: 'datasource', label: '数据源', icon: <DatabaseOutlined />, color: '#1890ff', route: '/datasource' },
  { key: 'dataset', label: '数据集', icon: <TableOutlined />, color: '#52c41a', route: '/dataset' },
  { key: 'chart', label: '图表', icon: <BarChartOutlined />, color: '#faad14', route: '/chart' },
  { key: 'dashboard', label: '仪表板', icon: <LayoutOutlined />, color: '#722ed1', route: '/dashboard' },
  { key: 'screen', label: '数据大屏', icon: <DesktopOutlined />, color: '#eb2f96', route: '/dashboard' },
]

export default function WorkbenchPage() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const [stats, setStats] = useState<WorkbenchStats | null>(null)
  const [recentList, setRecentList] = useState<RecentViewItem[]>([])
  const [favList, setFavList] = useState<FavoriteItem[]>([])
  const [loading, setLoading] = useState(false)

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [s, r, f] = await Promise.all([
        workbenchAPI.getStats(),
        workbenchAPI.getRecent(),
        workbenchAPI.getFavorites(),
      ])
      setStats(s)
      setRecentList(r)
      setFavList(f)
    } catch {
      message.error('获取工作台数据失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleQuickCreate = (key: string) => {
    if (key === 'datasource') navigate('/datasource')
    else if (key === 'dataset') navigate('/dataset')
    else if (key === 'chart') navigate('/chart')
    else if (key === 'dashboard') navigate('/dashboard')
    else if (key === 'screen') navigate('/dashboard')
  }

  const handleStatClick = (route: string) => {
    navigate(route)
  }

  const handleRemoveFavorite = async (type: string, id: number) => {
    try {
      await workbenchAPI.removeFavorite(type, id)
      message.success('取消收藏成功')
      setFavList((prev) => prev.filter((f) => f.id !== id))
    } catch {
      message.error('取消收藏失败')
    }
  }

  const recentColumns = [
    {
      title: '标题',
      dataIndex: 'title',
      render: (text: string, record: RecentViewItem) => (
        <a onClick={() => navigate(record.type === 'screen' ? `/screen/view/${record.id}` : `/dashboard/view/${record.id}`)}>
          {text}
        </a>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      render: (type: string) => (type === 'screen' ? <Tag color="purple">数据大屏</Tag> : <Tag color="blue">仪表板</Tag>),
    },
    {
      title: '最后访问时间',
      dataIndex: 'visitedAt',
      width: 200,
    },
  ]

  const favColumns = [
    {
      title: '标题',
      dataIndex: 'title',
      render: (text: string, record: FavoriteItem) => (
        <a onClick={() => navigate(record.type === 'screen' ? `/screen/view/${record.id}` : `/dashboard/view/${record.id}`)}>
          {text}
        </a>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      render: (type: string) => (type === 'screen' ? <Tag color="purple">数据大屏</Tag> : <Tag color="blue">仪表板</Tag>),
    },
    {
      title: '收藏时间',
      dataIndex: 'createdAt',
      width: 200,
    },
    {
      title: '操作',
      width: 100,
      render: (_: unknown, record: FavoriteItem) => (
        <Button type="link" danger onClick={() => handleRemoveFavorite(record.type, record.id)}>
          取消收藏
        </Button>
      ),
    },
  ]

  const statMap: Record<string, number | undefined> = {
    datasource: stats?.datasourceCount,
    dataset: stats?.datasetCount,
    chart: stats?.chartCount,
    dashboard: stats?.dashboardCount,
    screen: stats?.screenCount,
  }

  return (
    <div style={{ padding: 24 }}>
      {/* 欢迎区 + 快捷创建 */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <h2 style={{ margin: 0, fontWeight: 500 }}>
            欢迎回来，{user?.nickName || user?.username || '用户'}，祝您开心每一天！
          </h2>
        </div>
        <Space>
          {statCards.map((card) => (
            <Button key={card.key} icon={<PlusOutlined />} onClick={() => handleQuickCreate(card.key)}>
              新建{card.label}
            </Button>
          ))}
        </Space>
      </div>

      {/* 统计卡片 */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 16, marginBottom: 24 }}>
        {statCards.map((card) => (
          <Card
            key={card.key}
            hoverable
            onClick={() => handleStatClick(card.route)}
            bodyStyle={{ textAlign: 'center', padding: 24 }}
          >
            <div style={{ fontSize: 32, color: card.color, marginBottom: 8 }}>{card.icon}</div>
            <Statistic
              title={card.label}
              value={statMap[card.key] ?? 0}
              valueStyle={{ color: card.color, fontSize: 28, fontWeight: 600 }}
            />
          </Card>
        ))}
      </div>

      {/* 资源标签页 */}
      <Card loading={loading}>
        <Tabs defaultActiveKey="recent">
          <TabPane tab="最近访问" key="recent">
            <Table
              rowKey="id"
              columns={recentColumns}
              dataSource={recentList}
              pagination={false}
              locale={{ emptyText: <Empty description="暂无最近访问记录" /> }}
            />
          </TabPane>
          <TabPane tab="我的收藏" key="favorites">
            <Table
              rowKey="id"
              columns={favColumns}
              dataSource={favList}
              pagination={false}
              locale={{ emptyText: <Empty description="暂无收藏资源" /> }}
            />
          </TabPane>
        </Tabs>
      </Card>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/workbench/index.tsx
git commit -m "feat(frontend): add WorkbenchPage with stats, recent views and favorites"
```

---

## Task 12: 修改路由

**Files:**
- Modify: `frontend/src/router/index.tsx`

- [ ] **Step 1: 替换路由占位符**

将 `frontend/src/router/index.tsx` 中的：

```tsx
import LoginPage from '@/pages/login'
```

下方添加：

```tsx
import WorkbenchPage from '@/pages/workbench'
```

将：

```tsx
<Route path="/" element={<div style={{ padding: 24 }}>工作台（建设中）</div>} />
```

替换为：

```tsx
<Route path="/" element={<WorkbenchPage />} />
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/router/index.tsx
git commit -m "feat(router): replace workbench placeholder with WorkbenchPage"
```

---

## Task 13: 查看页面记录浏览历史

**Files:**
- Modify: `frontend/src/pages/dashboard/DashboardView.tsx`
- Modify: `frontend/src/pages/screen/ScreenView.tsx`

- [ ] **Step 1: DashboardView 添加浏览记录**

在 `frontend/src/pages/dashboard/DashboardView.tsx` 的 imports 中添加：

```tsx
import { workbenchAPI } from '@/api/workbench'
```

在 `fetchDashboard` 的 `setDashboard(d)` 之后，添加浏览记录调用。找到这段代码：

```tsx
      const d = await dashboardAPI.get(numericId)
      if (d.type === 'screen') {
        navigate(`/screen/view/${id}`)
        return
      }
      setDashboard(d)
```

在其后添加：

```tsx
      // 记录浏览历史
      try {
        await workbenchAPI.recordVisit({ resourceType: 'dashboard', resourceId: numericId })
      } catch {
        // ignore
      }
```

- [ ] **Step 2: ScreenView 添加浏览记录**

在 `frontend/src/pages/screen/ScreenView.tsx` 的 imports 中添加：

```tsx
import { workbenchAPI } from '@/api/workbench'
```

找到 `useScreenData` hook 被调用后数据加载成功的地方，或者直接在最外层 `useEffect` 中添加。在组件主逻辑中，找到 `numericId` 的定义后，添加：

```tsx
  // 记录浏览历史
  useEffect(() => {
    if (isValidId) {
      workbenchAPI.recordVisit({ resourceType: 'screen', resourceId: numericId }).catch(() => {})
    }
  }, [numericId, isValidId])
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/dashboard/DashboardView.tsx frontend/src/pages/screen/ScreenView.tsx
git commit -m "feat(frontend): record visit history on dashboard and screen view"
```

---

## Task 14: 验证构建

- [ ] **Step 1: 后端编译检查**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go build ./...
```

Expected: 编译成功，无错误。

- [ ] **Step 2: 后端测试检查**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go test ./internal/handler/... -v -run "Workbench"
```

Expected: 5 个测试全部通过。

- [ ] **Step 3: 前端构建检查**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/frontend
npm run build
```

Expected: 构建成功，无 TypeScript 错误。

- [ ] **Step 4: Commit（如有未提交的变更）**

```bash
git status
git add -A
git commit -m "feat(workbench): complete workbench page implementation"
```

---

## Spec Coverage Checklist

| Spec 需求 | 对应 Task |
|-----------|-----------|
| 数据库迁移（recent_views + favorites） | Task 1 |
| Model 层（RecentView + Favorite） | Task 2 |
| DTO 层（Stats + RecentViewItem + FavoriteItem） | Task 3 |
| Repository 层（统计/浏览记录/收藏） | Task 4 |
| Service 层（业务逻辑） | Task 5 |
| Handler + 测试 | Task 6, 7 |
| 路由注册 | Task 8 |
| 前端类型/API | Task 9, 10 |
| 工作台页面（欢迎区/快捷创建/统计/标签页） | Task 11 |
| 替换路由占位符 | Task 12 |
| 查看页面记录浏览历史 | Task 13 |
| 构建验证 | Task 14 |

**Gap:** None — 所有 spec 需求已覆盖。

---

## Placeholder Scan

- No "TBD", "TODO", "implement later" found.
- All code blocks contain complete, runnable code.
- All test steps include exact expected output.
- All file paths are exact.

---

## Type Consistency Check

- `RecentViewItem.Type` / `FavoriteItem.Type`: 前后端统一为 `string`，值域 `dashboard` | `screen`。
- `WorkbenchStatsResponse` / `WorkbenchStats`: 字段名 `datasourceCount` 等前后端一致。
- `RecordVisitRequest`: 前后端字段名 `resourceType` + `resourceId` 一致。
- 后端 `CountByCreatedBy` 的 `extraWhere` 参数：在 Task 4 中定义为函数参数，在 Task 5 的调用中使用 `"type = 'dashboard'"` 和 `"type = 'screen'"`。

All consistent.
