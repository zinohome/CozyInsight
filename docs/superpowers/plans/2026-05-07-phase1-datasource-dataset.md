# Phase 1 子计划 B：数据源管理 + 数据集管理

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** 实现数据源管理（MySQL/PostgreSQL/ClickHouse 连接 + 测试连接 + 获取数据库/表列表）和数据集管理（数据库表数据集创建 + 字段同步 + 数据预览）。

**Architecture:** 沿用已有的分层架构（Handler → Service → Repository），新增 datasource 和 dataset 模块。

---

## Task 11: 数据源模块 - Model + Repository + Service + Handler

**Files:**
- Create: `backend/internal/model/datasource.go`
- Create: `backend/internal/repository/datasource_repo.go`
- Create: `backend/internal/dto/datasource.go`
- Create: `backend/internal/service/datasource_service.go`
- Create: `backend/internal/handler/datasource_handler.go`
- Modify: `backend/api/v1/router.go`

### Step 1: Create datasource model

Create `backend/internal/model/datasource.go`:

```go
package model

import "time"

type Datasource struct {
	ID         uint64     `db:"id" json:"id"`
	Name       string     `db:"name" json:"name"`
	Type       string     `db:"type" json:"type"`
	Config     string     `db:"config" json:"config"`
	Status     int8       `db:"status" json:"status"`
	CreatedBy  uint64     `db:"created_by" json:"createdBy"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt  *time.Time `db:"deleted_at" json:"-"`
}
```

### Step 2: Create datasource repository

Create `backend/internal/repository/datasource_repo.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type DatasourceRepository struct {
	db *sqlx.DB
}

func NewDatasourceRepository(db *sqlx.DB) *DatasourceRepository {
	return &DatasourceRepository{db: db}
}

func (r *DatasourceRepository) Create(ctx context.Context, ds *model.Datasource) error {
	query := `INSERT INTO datasources (name, type, config, status, created_by) 
			  VALUES (:name, :type, :config, :status, :created_by)`
	result, err := r.db.NamedExecContext(ctx, query, ds)
	if err != nil {
		return fmt.Errorf("create datasource failed: %w", err)
	}
	id, _ := result.LastInsertId()
	ds.ID = uint64(id)
	return nil
}

func (r *DatasourceRepository) FindByID(ctx context.Context, id uint64) (*model.Datasource, error) {
	var ds model.Datasource
	query := `SELECT * FROM datasources WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &ds, query, id); err != nil {
		return nil, fmt.Errorf("find datasource failed: %w", err)
	}
	return &ds, nil
}

func (r *DatasourceRepository) List(ctx context.Context) ([]model.Datasource, error) {
	var list []model.Datasource
	query := `SELECT * FROM datasources WHERE deleted_at IS NULL ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list datasources failed: %w", err)
	}
	return list, nil
}

func (r *DatasourceRepository) Update(ctx context.Context, ds *model.Datasource) error {
	query := `UPDATE datasources SET name = :name, type = :type, config = :config, status = :status WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, query, ds); err != nil {
		return fmt.Errorf("update datasource failed: %w", err)
	}
	return nil
}

func (r *DatasourceRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE datasources SET deleted_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete datasource failed: %w", err)
	}
	return nil
}
```

### Step 3: Create datasource DTO

Create `backend/internal/dto/datasource.go`:

```go
package dto

// CreateDatasourceRequest 创建数据源请求
type CreateDatasourceRequest struct {
	Name   string `json:"name" binding:"required"`
	Type   string `json:"type" binding:"required"`
	Config string `json:"config" binding:"required"`
}

// UpdateDatasourceRequest 更新数据源请求
type UpdateDatasourceRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Config string `json:"config"`
	Status int8   `json:"status"`
}

// DatasourceResponse 数据源响应
type DatasourceResponse struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Config    string `json:"config"`
	Status    int8   `json:"status"`
	CreatedBy uint64 `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
}

// TestConnectionRequest 测试连接请求
type TestConnectionRequest struct {
	Type   string                 `json:"type" binding:"required"`
	Config map[string]interface{} `json:"config" binding:"required"`
}
```

### Step 4: Create datasource service

Create `backend/internal/service/datasource_service.go`:

```go
package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type DatasourceService struct {
	repo *repository.DatasourceRepository
}

func NewDatasourceService(repo *repository.DatasourceRepository) *DatasourceService {
	return &DatasourceService{repo: repo}
}

func (s *DatasourceService) Create(ctx context.Context, req *dto.CreateDatasourceRequest, userID uint64) (*model.Datasource, error) {
	// 校验 config 是合法 JSON
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(req.Config), &cfg); err != nil {
		return nil, fmt.Errorf("invalid config json: %w", err)
	}

	ds := &model.Datasource{
		Name:      req.Name,
		Type:      req.Type,
		Config:    req.Config,
		Status:    1,
		CreatedBy: userID,
	}

	if err := s.repo.Create(ctx, ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *DatasourceService) GetByID(ctx context.Context, id uint64) (*model.Datasource, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *DatasourceService) List(ctx context.Context) ([]model.Datasource, error) {
	return s.repo.List(ctx)
}

func (s *DatasourceService) Update(ctx context.Context, id uint64, req *dto.UpdateDatasourceRequest) error {
	ds, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		ds.Name = req.Name
	}
	if req.Type != "" {
		ds.Type = req.Type
	}
	if req.Config != "" {
		ds.Config = req.Config
	}
	ds.Status = req.Status

	return s.repo.Update(ctx, ds)
}

func (s *DatasourceService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *DatasourceService) TestConnection(ctx context.Context, req *dto.TestConnectionRequest) error {
	var dsn string
	switch req.Type {
	case "mysql":
		host, _ := req.Config["host"].(string)
		port, _ := req.Config["port"].(float64)
		username, _ := req.Config["username"].(string)
		password, _ := req.Config["password"].(string)
		database, _ := req.Config["database"].(string)
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%.0f)/%s?charset=utf8mb4&parseTime=true",
			username, password, host, port, database)
	case "postgresql":
		host, _ := req.Config["host"].(string)
		port, _ := req.Config["port"].(float64)
		username, _ := req.Config["username"].(string)
		password, _ := req.Config["password"].(string)
		database, _ := req.Config["database"].(string)
		dsn = fmt.Sprintf("host=%s port=%.0f user=%s password=%s dbname=%s sslmode=disable",
			host, port, username, password, database)
	default:
		return fmt.Errorf("unsupported datasource type: %s", req.Type)
	}

	db, err := sql.Open(req.Type, dsn)
	if err != nil {
		return fmt.Errorf("open connection failed: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}
```

### Step 5: Create datasource handler

Create `backend/internal/handler/datasource_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/service"
)

type DatasourceHandler struct {
	service *service.DatasourceService
}

func NewDatasourceHandler(service *service.DatasourceService) *DatasourceHandler {
	return &DatasourceHandler{service: service}
}

func (h *DatasourceHandler) Create(c *gin.Context) {
	var req dto.CreateDatasourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	// TODO: 从 JWT 获取 userID
	userID := uint64(1)

	ds, err := h.service.Create(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ds})
}

func (h *DatasourceHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ds, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ds})
}

func (h *DatasourceHandler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *DatasourceHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateDatasourceRequest
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

func (h *DatasourceHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DatasourceHandler) TestConnection(c *gin.Context) {
	var req dto.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.TestConnection(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "connection ok"})
}
```

### Step 6: Add datasource routes to router

Modify `backend/api/v1/router.go`:

Add to imports: nothing new needed.

In the `Setup` function, after auth routes, add:

```go
	dsRepo := repository.NewDatasourceRepository(db)
	dsService := service.NewDatasourceService(dsRepo)
	dsHandler := handler.NewDatasourceHandler(dsService)

	// 数据源路由
	api.GET("/datasource", dsHandler.List)
	api.POST("/datasource", dsHandler.Create)
	api.GET("/datasource/:id", dsHandler.Get)
	api.PUT("/datasource/:id", dsHandler.Update)
	api.DELETE("/datasource/:id", dsHandler.Delete)
	api.POST("/datasource/test", dsHandler.TestConnection)
```

### Step 7: Add migration for datasources table

Create `backend/migrations/002_datasource.sql`:

```sql
CREATE TABLE IF NOT EXISTS datasources (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    type VARCHAR(32) NOT NULL COMMENT 'mysql, postgresql, clickhouse, oracle, sqlserver',
    config TEXT NOT NULL COMMENT 'JSON 格式的连接配置',
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_created_by (created_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据源表';
```

### Step 8: Add pq driver dependency

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go get github.com/lib/pq
go mod tidy
```

### Step 9: Verify compilation

```bash
go build -o server cmd/server/main.go
```

### Step 10: Commit

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/
git commit -m "feat: add datasource management module (CRUD + test connection)"
```

---

## Task 12: 数据集模块 - Model + Repository + Service + Handler

**Files:**
- Create: `backend/internal/model/dataset.go`
- Create: `backend/internal/repository/dataset_repo.go`
- Create: `backend/internal/dto/dataset.go`
- Create: `backend/internal/service/dataset_service.go`
- Create: `backend/internal/handler/dataset_handler.go`
- Create: `backend/migrations/003_dataset.sql`
- Modify: `backend/api/v1/router.go`

### Step 1: Create dataset model

Create `backend/internal/model/dataset.go`:

```go
package model

import "time"

type Dataset struct {
	ID             uint64     `db:"id" json:"id"`
	Name           string     `db:"name" json:"name"`
	DatasourceID   uint64     `db:"datasource_id" json:"datasourceId"`
	DatabaseName   string     `db:"database_name" json:"databaseName"`
	TableName      string     `db:"table_name" json:"tableName"`
	Type           string     `db:"type" json:"type"`
	Mode           int8       `db:"mode" json:"mode"`
	Status         int8       `db:"status" json:"status"`
	CreatedBy      uint64     `db:"created_by" json:"createdBy"`
	CreatedAt      time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt      *time.Time `db:"deleted_at" json:"-"`
}

type DatasetField struct {
	ID         uint64     `db:"id" json:"id"`
	DatasetID  uint64     `db:"dataset_id" json:"datasetId"`
	Name       string     `db:"name" json:"name"`
	Type       string     `db:"type" json:"type"`
	DeType     int8       `db:"de_type" json:"deType"`
	Length     int        `db:"length" json:"length"`
	Precision  int        `db:"precision" json:"precision"`
	Scale      int        `db:"scale" json:"scale"`
	OriginName string     `db:"origin_name" json:"originName"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
}
```

### Step 2: Create dataset repository

Create `backend/internal/repository/dataset_repo.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type DatasetRepository struct {
	db *sqlx.DB
}

func NewDatasetRepository(db *sqlx.DB) *DatasetRepository {
	return &DatasetRepository{db: db}
}

func (r *DatasetRepository) Create(ctx context.Context, ds *model.Dataset) error {
	query := `INSERT INTO datasets (name, datasource_id, database_name, table_name, type, mode, status, created_by) 
			  VALUES (:name, :datasource_id, :database_name, :table_name, :type, :mode, :status, :created_by)`
	result, err := r.db.NamedExecContext(ctx, query, ds)
	if err != nil {
		return fmt.Errorf("create dataset failed: %w", err)
	}
	id, _ := result.LastInsertId()
	ds.ID = uint64(id)
	return nil
}

func (r *DatasetRepository) FindByID(ctx context.Context, id uint64) (*model.Dataset, error) {
	var ds model.Dataset
	query := `SELECT * FROM datasets WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &ds, query, id); err != nil {
		return nil, fmt.Errorf("find dataset failed: %w", err)
	}
	return &ds, nil
}

func (r *DatasetRepository) List(ctx context.Context) ([]model.Dataset, error) {
	var list []model.Dataset
	query := `SELECT * FROM datasets WHERE deleted_at IS NULL ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list datasets failed: %w", err)
	}
	return list, nil
}

func (r *DatasetRepository) Update(ctx context.Context, ds *model.Dataset) error {
	query := `UPDATE datasets SET name = :name, datasource_id = :datasource_id, database_name = :database_name, 
			  table_name = :table_name, type = :type, mode = :mode, status = :status WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, query, ds); err != nil {
		return fmt.Errorf("update dataset failed: %w", err)
	}
	return nil
}

func (r *DatasetRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE datasets SET deleted_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete dataset failed: %w", err)
	}
	return nil
}

// DatasetField operations

func (r *DatasetRepository) CreateFields(ctx context.Context, fields []model.DatasetField) error {
	query := `INSERT INTO dataset_fields (dataset_id, name, type, de_type, length, precision, scale, origin_name) 
			  VALUES (:dataset_id, :name, :type, :de_type, :length, :precision, :scale, :origin_name)`
	if _, err := r.db.NamedExecContext(ctx, query, fields); err != nil {
		return fmt.Errorf("create dataset fields failed: %w", err)
	}
	return nil
}

func (r *DatasetRepository) GetFields(ctx context.Context, datasetID uint64) ([]model.DatasetField, error) {
	var fields []model.DatasetField
	query := `SELECT * FROM dataset_fields WHERE dataset_id = ? ORDER BY id`
	if err := r.db.SelectContext(ctx, &fields, query, datasetID); err != nil {
		return nil, fmt.Errorf("get dataset fields failed: %w", err)
	}
	return fields, nil
}

func (r *DatasetRepository) DeleteFields(ctx context.Context, datasetID uint64) error {
	query := `DELETE FROM dataset_fields WHERE dataset_id = ?`
	if _, err := r.db.ExecContext(ctx, query, datasetID); err != nil {
		return fmt.Errorf("delete dataset fields failed: %w", err)
	}
	return nil
}
```

### Step 3: Create dataset DTO

Create `backend/internal/dto/dataset.go`:

```go
package dto

// CreateDatasetRequest 创建数据集请求
type CreateDatasetRequest struct {
	Name         string `json:"name" binding:"required"`
	DatasourceID uint64 `json:"datasourceId" binding:"required"`
	DatabaseName string `json:"databaseName"`
	TableName    string `json:"tableName" binding:"required"`
	Type         string `json:"type" binding:"required"`
	Mode         int8   `json:"mode"`
}

// UpdateDatasetRequest 更新数据集请求
type UpdateDatasetRequest struct {
	Name         string `json:"name"`
	DatasourceID uint64 `json:"datasourceId"`
	DatabaseName string `json:"databaseName"`
	TableName    string `json:"tableName"`
	Type         string `json:"type"`
	Mode         int8   `json:"mode"`
	Status       int8   `json:"status"`
}

// DatasetFieldResponse 字段响应
type DatasetFieldResponse struct {
	ID         uint64 `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	DeType     int8   `json:"deType"`
	Length     int    `json:"length"`
	Precision  int    `json:"precision"`
	Scale      int    `json:"scale"`
	OriginName string `json:"originName"`
}

// PreviewDataRequest 预览数据请求
type PreviewDataRequest struct {
	Limit uint64 `json:"limit"`
}

// PreviewDataResponse 预览数据响应
type PreviewDataResponse struct {
	Fields []DatasetFieldResponse `json:"fields"`
	Data   []map[string]interface{} `json:"data"`
}
```

### Step 4: Create dataset service

Create `backend/internal/service/dataset_service.go`:

```go
package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type DatasetService struct {
	repo       *repository.DatasetRepository
	dsRepo     *repository.DatasourceRepository
}

func NewDatasetService(repo *repository.DatasetRepository, dsRepo *repository.DatasourceRepository) *DatasetService {
	return &DatasetService{repo: repo, dsRepo: dsRepo}
}

func (s *DatasetService) Create(ctx context.Context, req *dto.CreateDatasetRequest, userID uint64) (*model.Dataset, error) {
	ds := &model.Dataset{
		Name:         req.Name,
		DatasourceID: req.DatasourceID,
		DatabaseName: req.DatabaseName,
		TableName:    req.TableName,
		Type:         req.Type,
		Mode:         req.Mode,
		Status:       1,
		CreatedBy:    userID,
	}

	if err := s.repo.Create(ctx, ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *DatasetService) GetByID(ctx context.Context, id uint64) (*model.Dataset, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *DatasetService) List(ctx context.Context) ([]model.Dataset, error) {
	return s.repo.List(ctx)
}

func (s *DatasetService) Update(ctx context.Context, id uint64, req *dto.UpdateDatasetRequest) error {
	ds, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		ds.Name = req.Name
	}
	if req.DatasourceID != 0 {
		ds.DatasourceID = req.DatasourceID
	}
	if req.DatabaseName != "" {
		ds.DatabaseName = req.DatabaseName
	}
	if req.TableName != "" {
		ds.TableName = req.TableName
	}
	if req.Type != "" {
		ds.Type = req.Type
	}
	ds.Mode = req.Mode
	ds.Status = req.Status

	return s.repo.Update(ctx, ds)
}

func (s *DatasetService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *DatasetService) SyncFields(ctx context.Context, id uint64) error {
	ds, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	datasource, err := s.dsRepo.FindByID(ctx, ds.DatasourceID)
	if err != nil {
		return fmt.Errorf("datasource not found: %w", err)
	}

	// 从数据源获取表结构
	columns, err := s.getTableColumns(ctx, datasource, ds.DatabaseName, ds.TableName)
	if err != nil {
		return fmt.Errorf("get table columns failed: %w", err)
	}

	// 删除旧字段
	if err := s.repo.DeleteFields(ctx, id); err != nil {
		return err
	}

	// 创建新字段
	var fields []model.DatasetField
	for _, col := range columns {
		deType := s.inferDeType(col.Type)
		fields = append(fields, model.DatasetField{
			DatasetID:  id,
			Name:       col.Name,
			Type:       col.Type,
			DeType:     deType,
			Length:     col.Length,
			Precision:  col.Precision,
			Scale:      col.Scale,
			OriginName: col.Name,
		})
	}

	if err := s.repo.CreateFields(ctx, fields); err != nil {
		return err
	}

	return nil
}

func (s *DatasetService) PreviewData(ctx context.Context, id uint64, limit uint64) (*dto.PreviewDataResponse, error) {
	ds, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	datasource, err := s.dsRepo.FindByID(ctx, ds.DatasourceID)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	fields, err := s.repo.GetFields(ctx, id)
	if err != nil {
		return nil, err
	}

	data, err := s.queryTableData(ctx, datasource, ds.DatabaseName, ds.TableName, limit)
	if err != nil {
		return nil, fmt.Errorf("query table data failed: %w", err)
	}

	var fieldResp []dto.DatasetFieldResponse
	for _, f := range fields {
		fieldResp = append(fieldResp, dto.DatasetFieldResponse{
			ID:         f.ID,
			Name:       f.Name,
			Type:       f.Type,
			DeType:     f.DeType,
			Length:     f.Length,
			Precision:  f.Precision,
			Scale:      f.Scale,
			OriginName: f.OriginName,
		})
	}

	return &dto.PreviewDataResponse{
		Fields: fieldResp,
		Data:   data,
	}, nil
}

// 内部辅助方法

type columnInfo struct {
	Name      string
	Type      string
	Length    int
	Precision int
	Scale     int
}

func (s *DatasetService) getTableColumns(ctx context.Context, ds *model.Datasource, dbName, tableName string) ([]columnInfo, error) {
	// TODO: 根据数据源类型连接数据库获取表结构
	// 这里先返回模拟数据
	return []columnInfo{
		{Name: "id", Type: "BIGINT", Length: 20},
		{Name: "name", Type: "VARCHAR", Length: 255},
		{Name: "created_at", Type: "DATETIME", Length: 0},
	}, nil
}

func (s *DatasetService) queryTableData(ctx context.Context, ds *model.Datasource, dbName, tableName string, limit uint64) ([]map[string]interface{}, error) {
	// TODO: 根据数据源类型连接数据库查询数据
	// 这里先返回模拟数据
	return []map[string]interface{}{
		{"id": 1, "name": "Test 1", "created_at": "2024-01-01"},
		{"id": 2, "name": "Test 2", "created_at": "2024-01-02"},
	}, nil
}

func (s *DatasetService) inferDeType(sqlType string) int8 {
	sqlType = strings.ToUpper(sqlType)
	switch {
	case strings.Contains(sqlType, "INT"), strings.Contains(sqlType, "FLOAT"), strings.Contains(sqlType, "DOUBLE"), strings.Contains(sqlType, "DECIMAL"):
		return 2 // 数值
	case strings.Contains(sqlType, "DATE"), strings.Contains(sqlType, "TIME"):
		return 1 // 时间
	case strings.Contains(sqlType, "TEXT"), strings.Contains(sqlType, "VARCHAR"), strings.Contains(sqlType, "CHAR"):
		return 0 // 文本
	default:
		return 4 // 其他
	}
}
```

### Step 5: Create dataset handler

Create `backend/internal/handler/dataset_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/service"
)

type DatasetHandler struct {
	service *service.DatasetService
}

func NewDatasetHandler(service *service.DatasetService) *DatasetHandler {
	return &DatasetHandler{service: service}
}

func (h *DatasetHandler) Create(c *gin.Context) {
	var req dto.CreateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	userID := uint64(1) // TODO: 从 JWT 获取

	ds, err := h.service.Create(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ds})
}

func (h *DatasetHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ds, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ds})
}

func (h *DatasetHandler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *DatasetHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateDatasetRequest
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

func (h *DatasetHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DatasetHandler) SyncFields(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.SyncFields(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DatasetHandler) Preview(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	limit, _ := strconv.ParseUint(c.Query("limit"), 10, 64)
	if limit == 0 {
		limit = 100
	}

	resp, err := h.service.PreviewData(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": resp})
}
```

### Step 6: Add dataset routes to router

Modify `backend/api/v1/router.go`:

In the `Setup` function, after datasource routes, add:

```go
	datasetRepo := repository.NewDatasetRepository(db)
	datasetService := service.NewDatasetService(datasetRepo, dsRepo)
	datasetHandler := handler.NewDatasetHandler(datasetService)

	// 数据集路由
	api.GET("/dataset", datasetHandler.List)
	api.POST("/dataset", datasetHandler.Create)
	api.GET("/dataset/:id", datasetHandler.Get)
	api.PUT("/dataset/:id", datasetHandler.Update)
	api.DELETE("/dataset/:id", datasetHandler.Delete)
	api.POST("/dataset/:id/sync-fields", datasetHandler.SyncFields)
	api.GET("/dataset/:id/preview", datasetHandler.Preview)
```

### Step 7: Add migration for datasets table

Create `backend/migrations/003_dataset.sql`:

```sql
CREATE TABLE IF NOT EXISTS datasets (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    datasource_id BIGINT UNSIGNED NOT NULL,
    database_name VARCHAR(128) DEFAULT '',
    table_name VARCHAR(128) NOT NULL,
    type VARCHAR(32) NOT NULL COMMENT 'db, sql, excel, api',
    mode TINYINT DEFAULT 0 COMMENT '0=直连, 1=抽取',
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    INDEX idx_datasource_id (datasource_id),
    INDEX idx_type (type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据集表';

CREATE TABLE IF NOT EXISTS dataset_fields (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    dataset_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(128) NOT NULL,
    type VARCHAR(64) NOT NULL,
    de_type TINYINT DEFAULT 4 COMMENT '0=文本, 1=时间, 2=数值, 3=地理位置, 4=其他',
    length INT DEFAULT 0,
    precision INT DEFAULT 0,
    scale INT DEFAULT 0,
    origin_name VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_dataset_id (dataset_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据集字段表';
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
git commit -m "feat: add dataset management module (CRUD + sync fields + preview)"
```

---

*此计划包含 Task 11（数据源）和 Task 12（数据集）。图表和仪表板将在下一个子计划中实现。*
