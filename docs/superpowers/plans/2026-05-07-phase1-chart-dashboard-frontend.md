# Phase 1 子计划 C：图表管理 + 仪表板管理 + 前端页面

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** 实现图表管理（CRUD + 渲染配置）、仪表板管理（CRUD + 网格布局配置）、以及对应的前端页面（数据源列表、数据集列表、图表编辑器、仪表板编辑器）。

**Architecture:** 沿用已有的分层架构（Handler → Service → Repository），新增 chart 和 dashboard 模块。前端使用 Ant Design 6 + ECharts 5。

---

## Task 13: 图表模块 - Model + Repository + Service + Handler

**Files:**
- Create: `backend/internal/model/chart.go`
- Create: `backend/internal/repository/chart_repo.go`
- Create: `backend/internal/dto/chart.go`
- Create: `backend/internal/service/chart_service.go`
- Create: `backend/internal/handler/chart_handler.go`
- Create: `backend/migrations/004_chart.sql`
- Modify: `backend/api/v1/router.go`

### Step 1: Create chart model

Create `backend/internal/model/chart.go`:

```go
package model

import "time"

type Chart struct {
	ID          uint64     `db:"id" json:"id"`
	Title       string     `db:"title" json:"title"`
	Type        string     `db:"type" json:"type"`
	DatasetID   uint64     `db:"dataset_id" json:"datasetId"`
	Config      string     `db:"config" json:"config"`
	Status      int8       `db:"status" json:"status"`
	CreatedBy   uint64     `db:"created_by" json:"createdBy"`
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt   *time.Time `db:"deleted_at" json:"-"`
}
```

### Step 2: Create chart repository

Create `backend/internal/repository/chart_repo.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type ChartRepository struct {
	db *sqlx.DB
}

func NewChartRepository(db *sqlx.DB) *ChartRepository {
	return &ChartRepository{db: db}
}

func (r *ChartRepository) Create(ctx context.Context, chart *model.Chart) error {
	query := `INSERT INTO charts (title, type, dataset_id, config, status, created_by)
			  VALUES (:title, :type, :dataset_id, :config, :status, :created_by)`
	result, err := r.db.NamedExecContext(ctx, query, chart)
	if err != nil {
		return fmt.Errorf("create chart failed: %w", err)
	}
	id, _ := result.LastInsertId()
	chart.ID = uint64(id)
	return nil
}

func (r *ChartRepository) FindByID(ctx context.Context, id uint64) (*model.Chart, error) {
	var chart model.Chart
	query := `SELECT * FROM charts WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &chart, query, id); err != nil {
		return nil, fmt.Errorf("find chart failed: %w", err)
	}
	return &chart, nil
}

func (r *ChartRepository) List(ctx context.Context) ([]model.Chart, error) {
	var list []model.Chart
	query := `SELECT * FROM charts WHERE deleted_at IS NULL ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list charts failed: %w", err)
	}
	return list, nil
}

func (r *ChartRepository) Update(ctx context.Context, chart *model.Chart) error {
	query := `UPDATE charts SET title = :title, type = :type, dataset_id = :dataset_id, config = :config, status = :status WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, query, chart); err != nil {
		return fmt.Errorf("update chart failed: %w", err)
	}
	return nil
}

func (r *ChartRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE charts SET deleted_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete chart failed: %w", err)
	}
	return nil
}
```

### Step 3: Create chart DTO

Create `backend/internal/dto/chart.go`:

```go
package dto

// CreateChartRequest 创建图表请求
type CreateChartRequest struct {
	Title     string `json:"title" binding:"required"`
	Type      string `json:"type" binding:"required"`
	DatasetID uint64 `json:"datasetId" binding:"required"`
	Config    string `json:"config"`
}

// UpdateChartRequest 更新图表请求
type UpdateChartRequest struct {
	Title     string `json:"title"`
	Type      string `json:"type"`
	DatasetID uint64 `json:"datasetId"`
	Config    string `json:"config"`
	Status    int8   `json:"status"`
}
```

### Step 4: Create chart service

Create `backend/internal/service/chart_service.go`:

```go
package service

import (
	"context"
	"encoding/json"
	"fmt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type ChartService struct {
	repo *repository.ChartRepository
}

func NewChartService(repo *repository.ChartRepository) *ChartService {
	return &ChartService{repo: repo}
}

func (s *ChartService) Create(ctx context.Context, req *dto.CreateChartRequest, userID uint64) (*model.Chart, error) {
	if req.Config == "" {
		req.Config = "{}"
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(req.Config), &cfg); err != nil {
		return nil, fmt.Errorf("invalid config json: %w", err)
	}

	chart := &model.Chart{
		Title:     req.Title,
		Type:      req.Type,
		DatasetID: req.DatasetID,
		Config:    req.Config,
		Status:    1,
		CreatedBy: userID,
	}

	if err := s.repo.Create(ctx, chart); err != nil {
		return nil, err
	}
	return chart, nil
}

func (s *ChartService) GetByID(ctx context.Context, id uint64) (*model.Chart, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ChartService) List(ctx context.Context) ([]model.Chart, error) {
	return s.repo.List(ctx)
}

func (s *ChartService) Update(ctx context.Context, id uint64, req *dto.UpdateChartRequest) error {
	chart, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Title != "" {
		chart.Title = req.Title
	}
	if req.Type != "" {
		chart.Type = req.Type
	}
	if req.DatasetID != 0 {
		chart.DatasetID = req.DatasetID
	}
	if req.Config != "" {
		chart.Config = req.Config
	}
	chart.Status = req.Status

	return s.repo.Update(ctx, chart)
}

func (s *ChartService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
```

### Step 5: Create chart handler

Create `backend/internal/handler/chart_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/service"
)

type ChartHandler struct {
	service *service.ChartService
}

func NewChartHandler(service *service.ChartService) *ChartHandler {
	return &ChartHandler{service: service}
}

func (h *ChartHandler) Create(c *gin.Context) {
	var req dto.CreateChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	userID := uint64(1)

	chart, err := h.service.Create(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": chart})
}

func (h *ChartHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	chart, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": chart})
}

func (h *ChartHandler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *ChartHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.Update(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *ChartHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}
```

### Step 6: Add chart routes to router

Modify `backend/api/v1/router.go`:

In the `Setup` function, after dataset routes, add:

```go
	chartRepo := repository.NewChartRepository(db)
	chartService := service.NewChartService(chartRepo)
	chartHandler := handler.NewChartHandler(chartService)

	// 图表路由
	api.GET("/chart", chartHandler.List)
	api.POST("/chart", chartHandler.Create)
	api.GET("/chart/:id", chartHandler.Get)
	api.PUT("/chart/:id", chartHandler.Update)
	api.DELETE("/chart/:id", chartHandler.Delete)
```

### Step 7: Add migration for charts table

Create `backend/migrations/004_chart.sql`:

```sql
CREATE TABLE IF NOT EXISTS charts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(128) NOT NULL,
    type VARCHAR(32) NOT NULL COMMENT 'bar, line, pie, table, etc',
    dataset_id BIGINT UNSIGNED NOT NULL,
    config TEXT NOT NULL COMMENT 'JSON 格式的图表配置',
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    INDEX idx_dataset_id (dataset_id),
    INDEX idx_type (type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图表表';
```

### Step 8: Verify compilation

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go build -o server cmd/server/main.go
```

### Step 9: Commit

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/
git commit -m "feat: add chart management module (CRUD)"
```

---

## Task 14: 仪表板模块 - Model + Repository + Service + Handler

**Files:**
- Create: `backend/internal/model/dashboard.go`
- Create: `backend/internal/repository/dashboard_repo.go`
- Create: `backend/internal/dto/dashboard.go`
- Create: `backend/internal/service/dashboard_service.go`
- Create: `backend/internal/handler/dashboard_handler.go`
- Create: `backend/migrations/005_dashboard.sql`
- Modify: `backend/api/v1/router.go`

### Step 1: Create dashboard model

Create `backend/internal/model/dashboard.go`:

```go
package model

import "time"

type Dashboard struct {
	ID          uint64     `db:"id" json:"id"`
	Title       string     `db:"title" json:"title"`
	Config      string     `db:"config" json:"config"`
	Status      int8       `db:"status" json:"status"`
	CreatedBy   uint64     `db:"created_by" json:"createdBy"`
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt   *time.Time `db:"deleted_at" json:"-"`
}

type DashboardChart struct {
	ID          uint64    `db:"id" json:"id"`
	DashboardID uint64    `db:"dashboard_id" json:"dashboardId"`
	ChartID     uint64    `db:"chart_id" json:"chartId"`
	PositionX   int       `db:"position_x" json:"positionX"`
	PositionY   int       `db:"position_y" json:"positionY"`
	Width       int       `db:"width" json:"width"`
	Height      int       `db:"height" json:"height"`
	Config      string    `db:"config" json:"config"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}
```

### Step 2: Create dashboard repository

Create `backend/internal/repository/dashboard_repo.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type DashboardRepository struct {
	db *sqlx.DB
}

func NewDashboardRepository(db *sqlx.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) Create(ctx context.Context, d *model.Dashboard) error {
	query := `INSERT INTO dashboards (title, config, status, created_by)
			  VALUES (:title, :config, :status, :created_by)`
	result, err := r.db.NamedExecContext(ctx, query, d)
	if err != nil {
		return fmt.Errorf("create dashboard failed: %w", err)
	}
	id, _ := result.LastInsertId()
	d.ID = uint64(id)
	return nil
}

func (r *DashboardRepository) FindByID(ctx context.Context, id uint64) (*model.Dashboard, error) {
	var d model.Dashboard
	query := `SELECT * FROM dashboards WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &d, query, id); err != nil {
		return nil, fmt.Errorf("find dashboard failed: %w", err)
	}
	return &d, nil
}

func (r *DashboardRepository) List(ctx context.Context) ([]model.Dashboard, error) {
	var list []model.Dashboard
	query := `SELECT * FROM dashboards WHERE deleted_at IS NULL ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list dashboards failed: %w", err)
	}
	return list, nil
}

func (r *DashboardRepository) Update(ctx context.Context, d *model.Dashboard) error {
	query := `UPDATE dashboards SET title = :title, config = :config, status = :status WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, query, d); err != nil {
		return fmt.Errorf("update dashboard failed: %w", err)
	}
	return nil
}

func (r *DashboardRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE dashboards SET deleted_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete dashboard failed: %w", err)
	}
	return nil
}

// DashboardChart operations

func (r *DashboardRepository) AddChart(ctx context.Context, dc *model.DashboardChart) error {
	query := `INSERT INTO dashboard_charts (dashboard_id, chart_id, position_x, position_y, width, height, config)
			  VALUES (:dashboard_id, :chart_id, :position_x, :position_y, :width, :height, :config)`
	_, err := r.db.NamedExecContext(ctx, query, dc)
	if err != nil {
		return fmt.Errorf("add chart to dashboard failed: %w", err)
	}
	return nil
}

func (r *DashboardRepository) GetCharts(ctx context.Context, dashboardID uint64) ([]model.DashboardChart, error) {
	var list []model.DashboardChart
	query := `SELECT * FROM dashboard_charts WHERE dashboard_id = ? ORDER BY id`
	if err := r.db.SelectContext(ctx, &list, query, dashboardID); err != nil {
		return nil, fmt.Errorf("get dashboard charts failed: %w", err)
	}
	return list, nil
}

func (r *DashboardRepository) RemoveChart(ctx context.Context, dashboardID, chartID uint64) error {
	query := `DELETE FROM dashboard_charts WHERE dashboard_id = ? AND chart_id = ?`
	if _, err := r.db.ExecContext(ctx, query, dashboardID, chartID); err != nil {
		return fmt.Errorf("remove chart from dashboard failed: %w", err)
	}
	return nil
}
```

### Step 3: Create dashboard DTO

Create `backend/internal/dto/dashboard.go`:

```go
package dto

// CreateDashboardRequest 创建仪表板请求
type CreateDashboardRequest struct {
	Title  string `json:"title" binding:"required"`
	Config string `json:"config"`
}

// UpdateDashboardRequest 更新仪表板请求
type UpdateDashboardRequest struct {
	Title  string `json:"title"`
	Config string `json:"config"`
	Status int8   `json:"status"`
}

// AddChartToDashboardRequest 添加图表到仪表板请求
type AddChartToDashboardRequest struct {
	ChartID   uint64 `json:"chartId" binding:"required"`
	PositionX int    `json:"positionX"`
	PositionY int    `json:"positionY"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Config    string `json:"config"`
}
```

### Step 4: Create dashboard service

Create `backend/internal/service/dashboard_service.go`:

```go
package service

import (
	"context"
	"encoding/json"
	"fmt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type DashboardService struct {
	repo *repository.DashboardRepository
}

func NewDashboardService(repo *repository.DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) Create(ctx context.Context, req *dto.CreateDashboardRequest, userID uint64) (*model.Dashboard, error) {
	if req.Config == "" {
		req.Config = "{}"
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(req.Config), &cfg); err != nil {
		return nil, fmt.Errorf("invalid config json: %w", err)
	}

	d := &model.Dashboard{
		Title:     req.Title,
		Config:    req.Config,
		Status:    1,
		CreatedBy: userID,
	}

	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DashboardService) GetByID(ctx context.Context, id uint64) (*model.Dashboard, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *DashboardService) List(ctx context.Context) ([]model.Dashboard, error) {
	return s.repo.List(ctx)
}

func (s *DashboardService) Update(ctx context.Context, id uint64, req *dto.UpdateDashboardRequest) error {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Title != "" {
		d.Title = req.Title
	}
	if req.Config != "" {
		d.Config = req.Config
	}
	d.Status = req.Status

	return s.repo.Update(ctx, d)
}

func (s *DashboardService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *DashboardService) AddChart(ctx context.Context, dashboardID uint64, req *dto.AddChartToDashboardRequest) error {
	if req.Config == "" {
		req.Config = "{}"
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(req.Config), &cfg); err != nil {
		return fmt.Errorf("invalid config json: %w", err)
	}

	dc := &model.DashboardChart{
		DashboardID: dashboardID,
		ChartID:     req.ChartID,
		PositionX:   req.PositionX,
		PositionY:   req.PositionY,
		Width:       req.Width,
		Height:      req.Height,
		Config:      req.Config,
	}

	return s.repo.AddChart(ctx, dc)
}

func (s *DashboardService) GetCharts(ctx context.Context, dashboardID uint64) ([]model.DashboardChart, error) {
	return s.repo.GetCharts(ctx, dashboardID)
}

func (s *DashboardService) RemoveChart(ctx context.Context, dashboardID, chartID uint64) error {
	return s.repo.RemoveChart(ctx, dashboardID, chartID)
}
```

### Step 5: Create dashboard handler

Create `backend/internal/handler/dashboard_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/service"
)

type DashboardHandler struct {
	service *service.DashboardService
}

func NewDashboardHandler(service *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) Create(c *gin.Context) {
	var req dto.CreateDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	userID := uint64(1)

	d, err := h.service.Create(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": d})
}

func (h *DashboardHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	d, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": d})
}

func (h *DashboardHandler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *DashboardHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.Update(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DashboardHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DashboardHandler) AddChart(c *gin.Context) {
	dashboardID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.AddChartToDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.AddChart(c.Request.Context(), dashboardID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DashboardHandler) GetCharts(c *gin.Context) {
	dashboardID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, err := h.service.GetCharts(c.Request.Context(), dashboardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *DashboardHandler) RemoveChart(c *gin.Context) {
	dashboardID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	chartID, _ := strconv.ParseUint(c.Param("chartId"), 10, 64)
	if err := h.service.RemoveChart(c.Request.Context(), dashboardID, chartID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}
```

### Step 6: Add dashboard routes to router

Modify `backend/api/v1/router.go`:

In the `Setup` function, after chart routes, add:

```go
	dashboardRepo := repository.NewDashboardRepository(db)
	dashboardService := service.NewDashboardService(dashboardRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	// 仪表板路由
	api.GET("/dashboard", dashboardHandler.List)
	api.POST("/dashboard", dashboardHandler.Create)
	api.GET("/dashboard/:id", dashboardHandler.Get)
	api.PUT("/dashboard/:id", dashboardHandler.Update)
	api.DELETE("/dashboard/:id", dashboardHandler.Delete)
	api.POST("/dashboard/:id/charts", dashboardHandler.AddChart)
	api.GET("/dashboard/:id/charts", dashboardHandler.GetCharts)
	api.DELETE("/dashboard/:id/charts/:chartId", dashboardHandler.RemoveChart)
```

### Step 7: Add migration for dashboards table

Create `backend/migrations/005_dashboard.sql`:

```sql
CREATE TABLE IF NOT EXISTS dashboards (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(128) NOT NULL,
    config TEXT NOT NULL COMMENT 'JSON 格式的仪表板布局配置',
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仪表板表';

CREATE TABLE IF NOT EXISTS dashboard_charts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    dashboard_id BIGINT UNSIGNED NOT NULL,
    chart_id BIGINT UNSIGNED NOT NULL,
    position_x INT DEFAULT 0,
    position_y INT DEFAULT 0,
    width INT DEFAULT 6,
    height INT DEFAULT 4,
    config TEXT NOT NULL COMMENT 'JSON 格式的图表在仪表板上的配置',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_dashboard_id (dashboard_id),
    INDEX idx_chart_id (chart_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仪表板图表关联表';
```

### Step 8: Verify compilation

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go build -o server cmd/server/main.go
```

### Step 9: Commit

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/
git commit -m "feat: add dashboard management module (CRUD + chart layout)"
```

---

## Task 15: 前端页面 - 数据源列表 + 数据集列表 + 图表编辑器 + 仪表板编辑器

**Files:**
- Create: `frontend/src/types/datasource.ts`
- Create: `frontend/src/api/datasource.ts`
- Create: `frontend/src/pages/datasource/index.tsx`
- Create: `frontend/src/types/dataset.ts`
- Create: `frontend/src/api/dataset.ts`
- Create: `frontend/src/pages/dataset/index.tsx`
- Create: `frontend/src/types/chart.ts`
- Create: `frontend/src/api/chart.ts`
- Create: `frontend/src/pages/chart/index.tsx`
- Create: `frontend/src/types/dashboard.ts`
- Create: `frontend/src/api/dashboard.ts`
- Create: `frontend/src/pages/dashboard/index.tsx`
- Modify: `frontend/src/router/index.tsx`

### Step 1: 数据源类型 + API + 列表页面

Create `frontend/src/types/datasource.ts`:

```typescript
export interface Datasource {
  id: number
  name: string
  type: string
  config: string
  status: number
  createdBy: number
  createdAt: string
}

export interface CreateDatasourceRequest {
  name: string
  type: string
  config: string
}

export interface TestConnectionRequest {
  type: string
  config: Record<string, unknown>
}
```

Create `frontend/src/api/datasource.ts`:

```typescript
import request from './request'
import type { Datasource, CreateDatasourceRequest, TestConnectionRequest } from '@/types/datasource'

export const datasourceAPI = {
  list: () => request.get<{ data: Datasource[] }>('/api/v1/datasource'),
  create: (data: CreateDatasourceRequest) => request.post<{ data: Datasource }>('/api/v1/datasource', data),
  get: (id: number) => request.get<{ data: Datasource }>(`/api/v1/datasource/${id}`),
  update: (id: number, data: Partial<CreateDatasourceRequest>) => request.put(`/api/v1/datasource/${id}`, data),
  remove: (id: number) => request.delete(`/api/v1/datasource/${id}`),
  testConnection: (data: TestConnectionRequest) => request.post('/api/v1/datasource/test', data),
}
```

Create `frontend/src/pages/datasource/index.tsx`:

```tsx
import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, message } from 'antd'
import { datasourceAPI } from '@/api/datasource'
import type { Datasource } from '@/types/datasource'

export default function DatasourcePage() {
  const [list, setList] = useState<Datasource[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [form] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const res = await datasourceAPI.list()
      setList(res.data.data)
    } catch {
      message.error('获取数据源列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: { name: string; type: string; config: string }) => {
    try {
      await datasourceAPI.create(values)
      message.success('创建成功')
      setModalVisible(false)
      form.resetFields()
      fetchList()
    } catch {
      message.error('创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await datasourceAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const columns = [
    { title: '名称', dataIndex: 'name' },
    { title: '类型', dataIndex: 'type', render: (type: string) => <Tag>{type}</Tag> },
    { title: '状态', dataIndex: 'status', render: (status: number) => (status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    { title: '创建时间', dataIndex: 'createdAt' },
    {
      title: '操作',
      render: (_: unknown, record: Datasource) => (
        <Space>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建数据源</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title="新建数据源" open={modalVisible} onCancel={() => setModalVisible(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[{ value: 'mysql', label: 'MySQL' }, { value: 'postgresql', label: 'PostgreSQL' }]} />
          </Form.Item>
          <Form.Item name="config" label="配置 (JSON)" rules={[{ required: true }]}>
            <Input.TextArea rows={4} placeholder='{"host":"localhost","port":3306,...}' />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
```

### Step 2: 数据集类型 + API + 列表页面

Create `frontend/src/types/dataset.ts`:

```typescript
export interface Dataset {
  id: number
  name: string
  datasourceId: number
  databaseName: string
  tableName: string
  type: string
  mode: number
  status: number
  createdBy: number
  createdAt: string
}

export interface DatasetField {
  id: number
  datasetId: number
  name: string
  type: string
  deType: number
  length: number
  precision: number
  scale: number
  originName: string
}

export interface CreateDatasetRequest {
  name: string
  datasourceId: number
  databaseName: string
  tableName: string
  type: string
  mode: number
}

export interface PreviewDataResponse {
  fields: Array<{
    id: number
    name: string
    type: string
    deType: number
    length: number
    precision: number
    scale: number
    originName: string
  }>
  data: Array<Record<string, unknown>>
}
```

Create `frontend/src/api/dataset.ts`:

```typescript
import request from './request'
import type { Dataset, CreateDatasetRequest, PreviewDataResponse } from '@/types/dataset'

export const datasetAPI = {
  list: () => request.get<{ data: Dataset[] }>('/api/v1/dataset'),
  create: (data: CreateDatasetRequest) => request.post<{ data: Dataset }>('/api/v1/dataset', data),
  get: (id: number) => request.get<{ data: Dataset }>(`/api/v1/dataset/${id}`),
  update: (id: number, data: Partial<CreateDatasetRequest>) => request.put(`/api/v1/dataset/${id}`, data),
  remove: (id: number) => request.delete(`/api/v1/dataset/${id}`),
  syncFields: (id: number) => request.post(`/api/v1/dataset/${id}/sync-fields`),
  preview: (id: number, limit?: number) => request.get<{ data: PreviewDataResponse }>(`/api/v1/dataset/${id}/preview`, { params: { limit } }),
}
```

Create `frontend/src/pages/dataset/index.tsx`:

```tsx
import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, message } from 'antd'
import { datasetAPI } from '@/api/dataset'
import { datasourceAPI } from '@/api/datasource'
import type { Dataset, DatasetField } from '@/types/dataset'
import type { Datasource } from '@/types/datasource'

export default function DatasetPage() {
  const [list, setList] = useState<Dataset[]>([])
  const [datasources, setDatasources] = useState<Datasource[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [previewModal, setPreviewModal] = useState(false)
  const [previewData, setPreviewData] = useState<{ fields: DatasetField[]; data: Array<Record<string, unknown>> } | null>(null)
  const [form] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const [dsRes, listRes] = await Promise.all([datasourceAPI.list(), datasetAPI.list()])
      setDatasources(dsRes.data.data)
      setList(listRes.data.data)
    } catch {
      message.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: { name: string; datasourceId: number; databaseName: string; tableName: string; type: string }) => {
    try {
      await datasetAPI.create({ ...values, mode: 0 })
      message.success('创建成功')
      setModalVisible(false)
      form.resetFields()
      fetchList()
    } catch {
      message.error('创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await datasetAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const handlePreview = async (id: number) => {
    try {
      const res = await datasetAPI.preview(id, 10)
      setPreviewData(res.data.data)
      setPreviewModal(true)
    } catch {
      message.error('预览失败')
    }
  }

  const columns = [
    { title: '名称', dataIndex: 'name' },
    { title: '数据源ID', dataIndex: 'datasourceId' },
    { title: '表名', dataIndex: 'tableName' },
    { title: '类型', dataIndex: 'type' },
    { title: '状态', dataIndex: 'status', render: (status: number) => (status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    {
      title: '操作',
      render: (_: unknown, record: Dataset) => (
        <Space>
          <Button type="link" onClick={() => handlePreview(record.id)}>预览</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建数据集</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title="新建数据集" open={modalVisible} onCancel={() => setModalVisible(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="datasourceId" label="数据源" rules={[{ required: true }]}>
            <Select options={datasources.map(d => ({ value: d.id, label: d.name }))} />
          </Form.Item>
          <Form.Item name="databaseName" label="数据库名">
            <Input />
          </Form.Item>
          <Form.Item name="tableName" label="表名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[{ value: 'db', label: '数据库表' }]} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>
      <Modal title="数据预览" open={previewModal} onCancel={() => setPreviewModal(false)} footer={null} width={800}>
        {previewData && (
          <Table
            size="small"
            columns={previewData.fields.map(f => ({ title: f.name, dataIndex: f.name, key: f.name }))}
            dataSource={previewData.data.map((row, idx) => ({ ...row, key: idx }))}
            pagination={false}
          />
        )}
      </Modal>
    </div>
  )
}
```

### Step 3: 图表类型 + API + 列表页面

Create `frontend/src/types/chart.ts`:

```typescript
export interface Chart {
  id: number
  title: string
  type: string
  datasetId: number
  config: string
  status: number
  createdBy: number
  createdAt: string
}

export interface CreateChartRequest {
  title: string
  type: string
  datasetId: number
  config: string
}
```

Create `frontend/src/api/chart.ts`:

```typescript
import request from './request'
import type { Chart, CreateChartRequest } from '@/types/chart'

export const chartAPI = {
  list: () => request.get<{ data: Chart[] }>('/api/v1/chart'),
  create: (data: CreateChartRequest) => request.post<{ data: Chart }>('/api/v1/chart', data),
  get: (id: number) => request.get<{ data: Chart }>(`/api/v1/chart/${id}`),
  update: (id: number, data: Partial<CreateChartRequest>) => request.put(`/api/v1/chart/${id}`, data),
  remove: (id: number) => request.delete(`/api/v1/chart/${id}`),
}
```

Create `frontend/src/pages/chart/index.tsx`:

```tsx
import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, message } from 'antd'
import { chartAPI } from '@/api/chart'
import { datasetAPI } from '@/api/dataset'
import type { Chart } from '@/types/chart'
import type { Dataset } from '@/types/dataset'

export default function ChartPage() {
  const [list, setList] = useState<Chart[]>([])
  const [datasets, setDatasets] = useState<Dataset[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [form] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const [dsRes, listRes] = await Promise.all([datasetAPI.list(), chartAPI.list()])
      setDatasets(dsRes.data.data)
      setList(listRes.data.data)
    } catch {
      message.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: { title: string; type: string; datasetId: number; config: string }) => {
    try {
      await chartAPI.create(values)
      message.success('创建成功')
      setModalVisible(false)
      form.resetFields()
      fetchList()
    } catch {
      message.error('创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await chartAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const columns = [
    { title: '标题', dataIndex: 'title' },
    { title: '类型', dataIndex: 'type', render: (type: string) => <Tag>{type}</Tag> },
    { title: '数据集ID', dataIndex: 'datasetId' },
    { title: '状态', dataIndex: 'status', render: (status: number) => (status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    {
      title: '操作',
      render: (_: unknown, record: Chart) => (
        <Space>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建图表</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title="新建图表" open={modalVisible} onCancel={() => setModalVisible(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="title" label="标题" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[{ value: 'bar', label: '柱状图' }, { value: 'line', label: '折线图' }, { value: 'pie', label: '饼图' }, { value: 'table', label: '表格' }]} />
          </Form.Item>
          <Form.Item name="datasetId" label="数据集" rules={[{ required: true }]}>
            <Select options={datasets.map(d => ({ value: d.id, label: d.name }))} />
          </Form.Item>
          <Form.Item name="config" label="配置 (JSON)">
            <Input.TextArea rows={4} placeholder='{}' />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
```

### Step 4: 仪表板类型 + API + 列表页面

Create `frontend/src/types/dashboard.ts`:

```typescript
export interface Dashboard {
  id: number
  title: string
  config: string
  status: number
  createdBy: number
  createdAt: string
}

export interface DashboardChart {
  id: number
  dashboardId: number
  chartId: number
  positionX: number
  positionY: number
  width: number
  height: number
  config: string
}

export interface CreateDashboardRequest {
  title: string
  config: string
}
```

Create `frontend/src/api/dashboard.ts`:

```typescript
import request from './request'
import type { Dashboard, DashboardChart, CreateDashboardRequest } from '@/types/dashboard'

export const dashboardAPI = {
  list: () => request.get<{ data: Dashboard[] }>('/api/v1/dashboard'),
  create: (data: CreateDashboardRequest) => request.post<{ data: Dashboard }>('/api/v1/dashboard', data),
  get: (id: number) => request.get<{ data: Dashboard }>(`/api/v1/dashboard/${id}`),
  update: (id: number, data: Partial<CreateDashboardRequest>) => request.put(`/api/v1/dashboard/${id}`, data),
  remove: (id: number) => request.delete(`/api/v1/dashboard/${id}`),
  addChart: (dashboardId: number, data: { chartId: number; positionX: number; positionY: number; width: number; height: number; config?: string }) => request.post(`/api/v1/dashboard/${dashboardId}/charts`, data),
  getCharts: (dashboardId: number) => request.get<{ data: DashboardChart[] }>(`/api/v1/dashboard/${dashboardId}/charts`),
  removeChart: (dashboardId: number, chartId: number) => request.delete(`/api/v1/dashboard/${dashboardId}/charts/${chartId}`),
}
```

Create `frontend/src/pages/dashboard/index.tsx`:

```tsx
import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, message } from 'antd'
import { dashboardAPI } from '@/api/dashboard'
import type { Dashboard } from '@/types/dashboard'

export default function DashboardPage() {
  const [list, setList] = useState<Dashboard[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [form] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const res = await dashboardAPI.list()
      setList(res.data.data)
    } catch {
      message.error('获取仪表板列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: { title: string; config?: string }) => {
    try {
      await dashboardAPI.create({ title: values.title, config: values.config || '{}' })
      message.success('创建成功')
      setModalVisible(false)
      form.resetFields()
      fetchList()
    } catch {
      message.error('创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await dashboardAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const columns = [
    { title: '标题', dataIndex: 'title' },
    { title: '状态', dataIndex: 'status', render: (status: number) => (status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    { title: '创建时间', dataIndex: 'createdAt' },
    {
      title: '操作',
      render: (_: unknown, record: Dashboard) => (
        <Space>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建仪表板</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title="新建仪表板" open={modalVisible} onCancel={() => setModalVisible(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="title" label="标题" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
```

### Step 5: 更新路由

Modify `frontend/src/router/index.tsx`:

```tsx
import { Routes, Route } from 'react-router-dom'
import LoginPage from '@/pages/login'
import DatasourcePage from '@/pages/datasource'
import DatasetPage from '@/pages/dataset'
import ChartPage from '@/pages/chart'
import DashboardPage from '@/pages/dashboard'

export default function Router() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<div>工作台（建设中）</div>} />
      <Route path="/datasource" element={<DatasourcePage />} />
      <Route path="/dataset" element={<DatasetPage />} />
      <Route path="/chart" element={<ChartPage />} />
      <Route path="/dashboard" element={<DashboardPage />} />
    </Routes>
  )
}
```

### Step 6: 验证前端编译

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/frontend
npm run build
```

### Step 7: Commit

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add frontend/
git commit -m "feat: add frontend pages for datasource, dataset, chart, dashboard"
```

---

*Phase 1 全部完成：项目骨架 + JWT + 数据源 + 数据集 + 图表 + 仪表板 + 前端页面。*
