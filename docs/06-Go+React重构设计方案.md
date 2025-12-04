# DataEase Go + React 重构设计方案

## 重构目标

将 DataEase 从 **Java + Vue** 技术栈重构为 **Go + React** 技术栈,保持核心功能不变,提升性能、开发效率和可维护性。

---

## 技术栈选型

### 后端: Go 技术栈

#### 核心框架
- **Gin**: v1.9+ (高性能 Web 框架)
- **Go 版本**: 1.21+ (与 Java 21 对应)

#### ORM 与数据库
- **GORM**: v2.0+ (Go语言ORM,对应MyBatis Plus)
- **MySQL Driver**: go-sql-driver/mysql
- **数据库迁移**: golang-migrate/migrate

#### SQL 引擎 ⭐

**推荐方案: Apache Calcite Avatica Go Client + Avatica Server**

- **calcite-avatica-go**: v5+ (Apache 官方 Go 客户端)
- **Avatica Server**: 基于 Calcite 的 SQL Gateway
- **架构**: Go Backend → Avatica Client → HTTP/Protobuf → Avatica Server → Calcite → 数据源

**方案优势**:
- ✅ 保留 Apache Calcite 完整能力(SQL 解析、优化、跨数据源查询)
- ✅ Apache 官方支持,稳定可靠
- ✅ 通过 HTTP/Protobuf 通信,性能可接受
- ✅ 架构清晰,易于部署和扩展
- ✅ 零 SQL 引擎迁移成本

**备选方案**:
- CGO 调用 Java (性能差,不推荐)
- 自研 SQL 解析器 (开发成本高,长期方案)

#### 缓存
- **go-redis**: Redis 客户端
- **groupcache**: 本地缓存(替代 Ehcache)

#### 任务调度
- **cron**: robfig/cron v3 (替代 Quartz)
- **asynq**: 分布式异步任务队列

#### API 文档
- **Swaggo**: Swagger 自动生成(替代 Knife4j)

#### 其他工具库
- **jwt-go**: JWT 认证
- **viper**: 配置管理
- **zap**: 日志库(高性能)
- **excelize**: Excel 处理
- **chromedp**: 浏览器自动化(替代 Selenium)
- **go-pdf**: PDF 生成

---

### 前端: React 技术栈

#### 核心框架
- **React**: 18.2+
- **TypeScript**: 5.0+
- **React Router**: v6 (路由管理)

#### 状态管理
- **Zustand**: 轻量级状态管理(替代 Pinia)
- 备选: Redux Toolkit, Jotai

#### UI 组件库
- **Ant Design**: v5.x (成熟的企业级UI库,替代 Element Plus)
- **Ant Design Charts**: 基于 AntV 的 React 图表库
- 备选: Material-UI, Chakra UI

#### 数据可视化
- **AntV/G2**: React 适配版
- **AntV/L7**: 地理可视化
- **AntV/S2**: 表格分析
- **ECharts for React**: echarts-for-react

#### 构建工具
- **Vite**: v5.x (快速构建,保持与Vue项目一致的构建体验)
- 备选: Next.js (如需SSR)

#### 代码编辑器
- **Monaco Editor**: VSCode 内核(替代 CodeMirror)
- **React Ace**: Ace Editor React 版本

#### 拖拽库
- **react-dnd**: 拖拽功能
- **react-grid-layout**: 网格布局

#### 其他库
- **ahooks**: React Hooks 工具集
- **dayjs**: 日期处理
- **axios**: HTTP 客户端
- **lodash**: 工具函数

---

## 后端架构设计 (Go)

### 1. 项目结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # 入口文件
├── api/
│   └── v1/                      # API 路由定义
│       ├── datasource/
│       ├── dataset/
│       ├── chart/
│       ├── visualization/
│       └── ...
├── internal/                    # 内部包
│   ├── handler/                 # HTTP 处理器(Controller层)
│   ├── service/                 # 业务逻辑层
│   ├── repository/              # 数据访问层(DAO)
│   ├── model/                   # 数据模型
│   ├── dto/                     # 数据传输对象
│   ├── middleware/              # 中间件
│   ├── engine/                  # SQL 引擎
│   ├── cache/                   # 缓存层
│   └── util/                    # 工具函数
├── pkg/                         # 公共包(可导出)
│   ├── config/                  # 配置管理
│   ├── logger/                  # 日志
│   ├── database/                # 数据库连接
│   ├── jwt/                     # JWT 工具
│   └── crypto/                  # 加密工具
├── scripts/                     # 脚本
│   └── migrations/              # 数据库迁移
├── configs/                     # 配置文件
│   ├── app.yaml
│   └── app.production.yaml
├── go.mod
└── go.sum
```

### 2. 分层架构

```
┌─────────────────────────────────────┐
│         Handler Layer               │ (Gin handlers)
├─────────────────────────────────────┤
│         Service Layer               │ (Business Logic)
├─────────────────────────────────────┤
│       Repository Layer              │ (GORM)
├─────────────────────────────────────┤
│        Database Layer               │ (MySQL)
└─────────────────────────────────────┘
```

### 3. 模块対応 (Java → Go)

| Java 模块 | Go 包名 | 说明 |
|-----------|---------|------|
| datasource | internal/service/datasource | 数据源服务 |
| dataset | internal/service/dataset | 数据集服务 |
| chart | internal/service/chart | 图表服务 |
| visualization | internal/service/visualization | 可视化服务 |
| engine | internal/engine | SQL 引擎 |
| job | internal/scheduler | 任务调度 |
| system | internal/service/system | 系统管理 |

### 4. SQL 引擎实现 (Avatica) ⭐⭐⭐

#### 4.1 架构设计

```
┌──────────────────────────────────────────────────────────┐
│                     Go Backend                           │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Dataset Service / Chart Service                   │  │
│  └──────────────────┬─────────────────────────────────┘  │
│                     │                                     │
│  ┌──────────────────▼─────────────────────────────────┐  │
│  │         Calcite Client (avatica-go)                │  │
│  │  - SQL 执行                                         │  │
│  │  - 连接池管理                                       │  │
│  │  - 结果缓存                                         │  │
│  └──────────────────┬─────────────────────────────────┘  │
└────────────────────┬│──────────────────────────────────┘
                     ││ HTTP/Protobuf
                     ││
┌────────────────────▼▼──────────────────────────────────┐
│              Avatica Server (Java)                      │
│  ┌────────────────────────────────────────────────────┐ │
│  │            Apache Calcite                          │ │
│  │  - SQL Parser                                      │ │
│  │  - SQL Optimizer                                   │ │
│  │  - SQL Executor                                    │ │
│  └──────────────────┬─────────────────────────────────┘ │
└────────────────────┬│──────────────────────────────────┘
                     ││ JDBC
                     ││
┌────────────────────▼▼──────────────────────────────────┐
│              Data Sources                               │
│  MySQL / PostgreSQL / Oracle / ClickHouse / etc.       │
└────────────────────────────────────────────────────────┘
```

#### 4.2 Avatica Server 配置

```yaml
# avatica-server/application.yml
server:
  port: 8765
  
spring:
  application:
    name: avatica-server

avatica:
  # Avatica 配置
  max-statements-per-connection: 100
  connection-pool:
    max-size: 100
    min-idle: 10
    
calcite:
  # Calcite 配置
  parser:
    factory: org.apache.calcite.sql.parser.impl.SqlParserImpl
  optimizer:
    enable: true
    
# 数据源连接池
datasource:
  hikari:
    maximum-pool-size: 50
    minimum-idle: 10
    connection-timeout: 30000
    idle-timeout: 600000
    max-lifetime: 1800000
```

#### 4.3 Go 客户端实现

```go
// internal/engine/calcite_client.go
package engine

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/apache/calcite-avatica-go/v5"
)

type CalciteClient struct {
    db *sql.DB
    cache Cache // Redis 缓存
}

type CalciteConfig struct {
    AvaticaURL      string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

func NewCalciteClient(cfg *CalciteConfig, cache Cache) (*CalciteClient, error) {
    // 连接 Avatica Server
    // 格式: http://host:port/
    db, err := sql.Open("avatica", cfg.AvaticaURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Avatica: %w", err)
    }
    
    // 配置连接池
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    
    // 测试连接
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping Avatica: %w", err)
    }
    
    return &CalciteClient{
        db:    db,
        cache: cache,
    }, nil
}

// ExecuteQuery 执行查询(带缓存)
func (c *CalciteClient) ExecuteQuery(ctx context.Context, sql string, params ...interface{}) ([]map[string]interface{}, error) {
    // 生成缓存键
    cacheKey := c.generateCacheKey(sql, params...)
    
    // 检查缓存
    if cached, err := c.cache.Get(ctx, cacheKey); err == nil {
        return cached.([]map[string]interface{}), nil
    }
    
    // 执行查询
    rows, err := c.db.QueryContext(ctx, sql, params...)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()
    
    // 解析结果
    result, err := c.parseRows(rows)
    if err != nil {
        return nil, err
    }
    
    // 缓存结果 (TTL 5分钟)
    _ = c.cache.Set(ctx, cacheKey, result, 5*time.Minute)
    
    return result, nil
}

// ExecuteQueryNoCa che 执行查询(不缓存)
func (c *CalciteClient) ExecuteQueryNoCache(ctx context.Context, sql string, params ...interface{}) ([]map[string]interface{}, error) {
    rows, err := c.db.QueryContext(ctx, sql, params...)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()
    
    return c.parseRows(rows)
}

// parseRows 解析 SQL 结果
func (c *CalciteClient) parseRows(rows *sql.Rows) ([]map[string]interface{}, error) {
    columns, err := rows.Columns()
    if err != nil {
        return nil, err
    }
    
    var result []map[string]interface{}
    
    for rows.Next() {
        // 创建扫描目标
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        for i := range values {
            valuePtrs[i] = &values[i]
        }
        
        // 扫描行
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, err
        }
        
        // 构建结果 Map
        row := make(map[string]interface{})
        for i, col := range columns {
            row[col] = values[i]
        }
        
        result = append(result, row)
    }
    
    return result, rows.Err()
}

// generateCacheKey 生成缓存键
func (c *CalciteClient) generateCacheKey(sql string, params ...interface{}) string {
    // 简化版本,实际应使用更好的哈希算法
    return fmt.Sprintf("query:%s:%v", sql, params)
}

// Close 关闭连接
func (c *CalciteClient) Close() error {
    return c.db.Close()
}
```

#### 4.4 Service 层集成

```go
// internal/service/dataset_service.go
package service

import (
    "context"
    "backend/internal/engine"
    "backend/internal/repository"
)

type DatasetService struct {
    repo    repository.DatasetRepository
    calcite *engine.CalciteClient
}

func NewDatasetService(repo repository.DatasetRepository, calcite *engine.CalciteClient) *DatasetService {
    return &DatasetService{
        repo:    repo,
        calcite: calcite,
    }
}

// QueryData 查询数据集数据
func (s *DatasetService) QueryData(ctx context.Context, datasetID uint64, filter *QueryFilter) ([]map[string]interface{}, error) {
    // 获取数据集信息
    dataset, err := s.repo.GetByID(ctx, datasetID)
    if err != nil {
        return nil, err
    }
    
    // 构建 SQL (这里简化,实际需要根据 filter 构建)
    sql := dataset.SQL
    
    // 通过 Calcite 执行查询
    result, err := s.calcite.ExecuteQuery(ctx, sql)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}
```

#### 4.5 配置文件

```yaml
# configs/app.yaml
calcite:
  avatica_url: "http://localhost:8765/"
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime: 1h
```

#### 4.6 部署架构

**开发/测试环境**:
```
┌─────────────────┐
│  Go Backend     │ :8100
│  +               │
│  Avatica Server │ :8765
│  (同机部署)      │
└─────────────────┘
```

**生产环境**:
```
┌─────────────────┐      ┌─────────────────────────┐
│  Go Backend     │      │   Avatica Server        │
│  (K8s Pods)     │──────▶   Cluster               │
│  Replicas: 3-5  │      │   (Load Balanced)       │
└─────────────────┘      │   Replicas: 3-5         │
                          └─────────────────────────┘
                                     │
                          ┌──────────▼──────────────┐
                          │   Data Sources          │
                          │   (MySQL/PG/etc.)       │
                          └─────────────────────────┘
```

#### 4.7 性能优化

**连接池优化**:
```go
// 根据负载调整
db.SetMaxOpenConns(100)    // 最大连接数
db.SetMaxIdleConns(20)     // 最大空闲连接
db.SetConnMaxLifetime(1h)  // 连接最大生命周期
```

**查询缓存**:
- 使用 Redis 缓存常见查询结果
- TTL: 5-10 分钟(可配置)
- 缓存键: `query:{sql_hash}:{params_hash}`

**监控指标**:
- Avatica Server 健康检查
- 查询响应时间 (P50/P95/P99)
- 连接池使用率
- 缓存命中率

#### 4.8 Avatica Server Docker 部署

```dockerfile
# Dockerfile
FROM openjdk:21-slim

# 安装 Avatica Server
COPY avatica-server.jar /app/
COPY application.yml /app/config/

WORKDIR /app

EXPOSE 8765

CMD ["java", "-jar", "avatica-server.jar", "--spring.config.location=config/application.yml"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  avatica-server:
    image: avatica-server:latest
    ports:
      - "8765:8765"
    environment:
      - JAVA_OPTS=-Xmx2g -Xms1g
    volumes:
      - ./config:/app/config
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8765/actuator/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### 5. 配置管理 (Viper)

```yaml
# configs/app.yaml
server:
  port: 8100
  mode: debug  # debug, release

database:
  driver: mysql
  host: 127.0.0.1
  port: 3306
  username: root
  password: encrypted_password
  database: dataease
  
redis:
  host: 127.0.0.1
  port: 6379
  password:密码
  db: 0

cache:
  type: redis  # redis, memory
  
logger:
  level: info
  file: logs/app.log
```

### 5. 数据模型示例 (GORM)

```go
// internal/model/datasource.go
package model

import (
    "gorm.io/gorm"
)

type Datasource struct {
    ID            uint64         `gorm:"primaryKey"`
    Name          string         `gorm:"size:255;not null"`
    Description   string         `gorm:"size:255"`
    Type          string         `gorm:"size:50;not null"`
    PID           *uint64        `gorm:"index"`
    EditType      string         `gorm:"size:50"`
    Configuration string         `gorm:"type:longtext"`
    CreateTime    int64          `gorm:"not null"`
    UpdateTime    int64          `gorm:"not null"`
    CreateBy      string         `gorm:"size:50"`
    Status        string         `gorm:"type:longtext"`
    QrtzInstance  string         `gorm:"type:longtext"`
    TaskStatus    string         `gorm:"size:50"`
    DeletedAt     gorm.DeletedAt `gorm:"index"` // 软删除
}

func (Datasource) TableName() string {
    return "core_datasource"
}
```

### 6. Repository 示例

```go
// internal/repository/datasource_repo.go
package repository

import (
    "context"
    "backend/internal/model"
    "gorm.io/gorm"
)

type DatasourceRepository interface {
    Create(ctx context.Context, ds *model.Datasource) error
    GetByID(ctx context.Context, id uint64) (*model.Datasource, error)
    List(ctx context.Context, filter *DatasourceFilter) ([]*model.Datasource, error)
    Update(ctx context.Context, ds *model.Datasource) error
    Delete(ctx context.Context, id uint64) error
}

type datasourceRepo struct {
    db *gorm.DB
}

func NewDatasourceRepository(db *gorm.DB) DatasourceRepository {
    return &datasourceRepo{db: db}
}

func (r *datasourceRepo) Create(ctx context.Context, ds *model.Datasource) error {
    return r.db.WithContext(ctx).Create(ds).Error
}

// ... 其他方法
```

### 7. Service 示例

```go
// internal/service/datasource_service.go
package service

import (
    "context"
    "backend/internal/dto"
    "backend/internal/repository"
)

type DatasourceService interface {
    CreateDatasource(ctx context.Context, req *dto.CreateDatasourceRequest) error
    TestConnection(ctx context.Context, id uint64) (*dto.ConnectionTestResult, error)
    // ... 其他方法
}

type datasourceService struct {
    repo  repository.DatasourceRepository
    cache cache.Cache
}

func NewDatasourceService(repo repository.DatasourceRepository, cache cache.Cache) DatasourceService {
    return &datasourceService{
        repo:  repo,
        cache: cache,
    }
}
```

### 8. Handler 示例 (Gin)

```go
// internal/handler/datasource_handler.go
package handler

import (
    "net/http"
    "backend/internal/service"
    "backend/internal/dto"
    "github.com/gin-gonic/gin"
)

type DatasourceHandler struct {
    service service.DatasourceService
}

func NewDatasourceHandler(service service.DatasourceService) *DatasourceHandler {
    return &DatasourceHandler{service: service}
}

// CreateDatasource godoc
// @Summary 创建数据源
// @Tags datasource
// @Accept json
// @Produce json
// @Param request body dto.CreateDatasourceRequest true "请求体"
// @Success 200 {object} dto.Response
// @Router /api/v1/datasource [post]
func (h *DatasourceHandler) CreateDatasource(c *gin.Context) {
    var req dto.CreateDatasourceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
        return
    }
    
    if err := h.service.CreateDatasource(c.Request.Context(), &req); err != nil {
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
        return
    }
    
    c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}
```

### 9. 路由注册

```go
// api/v1/router.go
package v1

import (
    "backend/internal/handler"
    "backend/internal/middleware"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, handlers *handler.Handlers) {
    api := r.Group("/api/v1")
    api.Use(middleware.JWTAuth())
    
    // 数据源路由
    datasource := api.Group("/datasource")
    {
        datasource.POST("", handlers.Datasource.CreateDatasource)
        datasource.GET("/:id", handlers.Datasource.GetDatasource)
        datasource.PUT("/:id", handlers.Datasource.UpdateDatasource)
        datasource.DELETE("/:id", handlers.Datasource.DeleteDatasource)
        datasource.POST("/:id/test", handlers.Datasource.TestConnection)
    }
    
    // 数据集路由
    dataset := api.Group("/dataset")
    {
        // ...
    }
    
    // 图表路由
    chart := api.Group("/chart")
    {
        // ...
    }
}
```

---

## 前端架构设计 (React)

### 1. 项目结构

```
frontend/
├── public/
│   ├── index.html
│   └── assets/
├── src/
│   ├── api/                     # API 接口
│   │   ├── datasource.ts
│   │   ├── dataset.ts
│   │   └── chart.ts
│   ├── components/              # 公共组件
│   │   ├── Chart/
│   │   ├── DatasetSelector/
│   │   └── FilterPanel/
│   ├── layouts/                 # 布局组件
│   │   ├── MainLayout.tsx
│   │   └── DashboardLayout.tsx
│   ├── pages/                   # 页面组件
│   │   ├── dashboard/
│   │   ├── dataset/
│   │   ├── chart/
│   │   └── system/
│   ├── hooks/                   # 自定义 Hooks
│   │   ├── useAuth.ts
│   │   ├── useDataset.ts
│   │   └── useChart.ts
│   ├── store/                   # Zustand 状态管理
│   │   ├── userStore.ts
│   │   ├── chartStore.ts
│   │   └── canvasStore.ts
│   ├── utils/                   # 工具函数
│   ├── router/                  # 路由配置
│   │   └── index.tsx
│   ├── types/                   # TypeScript 类型定义
│   ├── styles/                  # 全局样式
│   ├── App.tsx
│   └── main.tsx
├── vite.config.ts
├── tsconfig.json
└── package.json
```

### 2. 状态管理 (Zustand)

```typescript
// src/store/chartStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface ChartState {
  charts: Chart[];
  activeChartId: string | null;
  setCharts: (charts: Chart[]) => void;
  addChart: (chart: Chart) => void;
  updateChart: (id: string, chart: Partial<Chart>) => void;
  deleteChart: (id: string) => void;
  setActiveChart: (id: string | null) => void;
}

export const useChartStore = create<ChartState>()(
  persist(
    (set) => ({
      charts: [],
      activeChartId: null,
      setCharts: (charts) => set({ charts }),
      addChart: (chart) => set((state) => ({ 
        charts: [...state.charts, chart] 
      })),
      updateChart: (id, updates) => set((state) => ({
        charts: state.charts.map(c => c.id === id ? { ...c, ...updates } : c)
      })),
      deleteChart: (id) => set((state) => ({
        charts: state.charts.filter(c => c.id !== id)
      })),
      setActiveChart: (id) => set({ activeChartId: id }),
    }),
    {
      name: 'chart-storage',
    }
  )
);
```

### 3. API 封装 (Axios)

```typescript
// src/api/request.ts
import axios, { AxiosRequestConfig } from 'axios';
import { message } from 'antd';

const instance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
});

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
instance.interceptors.response.use(
  (response) => {
    const { data } = response;
    if (data.code !== 200) {
      message.error(data.message);
      return Promise.reject(data);
    }
    return data.data;
  },
  (error) => {
    message.error(error.message);
    return Promise.reject(error);
  }
);

export default instance;

// src/api/datasource.ts
import request from './request';
import { Datasource, CreateDatasourceRequest } from '@/types/datasource';

export const datasourceAPI = {
  create: (data: CreateDatasourceRequest) => 
    request.post<Datasource>('/datasource', data),
  
  getById: (id: string) => 
    request.get<Datasource>(`/datasource/${id}`),
  
  list: (params?: any) => 
    request.get<Datasource[]>('/datasource', { params }),
  
  update: (id: string, data: Partial<Datasource>) => 
    request.put<Datasource>(`/datasource/${id}`, data),
  
  delete: (id: string) => 
    request.delete(`/datasource/${id}`),
  
  testConnection: (id: string) => 
    request.post<{ success: boolean; message: string }>(`/datasource/${id}/test`),
};
```

### 4. 自定义 Hooks

```typescript
// src/hooks/useDataset.ts
import { useState, useEffect } from 'react';
import { datasetAPI } from '@/api/dataset';
import { Dataset } from '@/types/dataset';

export const useDataset = (id?: string) => {
  const [dataset, setDataset] = useState<Dataset | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (id) {
      fetchDataset(id);
    }
  }, [id]);

  const fetchDataset = async (datasetId: string) => {
    try {
      setLoading(true);
      const data = await datasetAPI.getById(datasetId);
      setDataset(data);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const refresh = () => {
    if (id) fetchDataset(id);
  };

  return { dataset, loading, error, refresh };
};
```

### 5. 路由配置 (React Router v6)

```typescript
// src/router/index.tsx
import { createBrowserRouter, Navigate } from 'react-router-dom';
import MainLayout from '@/layouts/MainLayout';
import Dashboard from '@/pages/dashboard';
import Dataset from '@/pages/dataset';
import Chart from '@/pages/chart';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <MainLayout />,
    children: [
      {
        index: true,
        element: <Navigate to="/workbranch" replace />,
      },
      {
        path: 'workbranch',
        element: <Workbranch />,
      },
      {
        path: 'dashboard',
        children: [
          {
            index: true,
            element: <Dashboard />,
          },
          {
            path: ':id',
            element: <DashboardView />,
          },
        ],
      },
      {
        path: 'dataset',
        children: [
          {
            index: true,
            element: <Dataset />,
          },
          {
            path: 'form/:id?',
            element: <DatasetForm />,
          },
        ],
      },
      {
        path: 'chart',
        element: <Chart />,
      },
    ],
  },
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '*',
    element: <NotFound />,
  },
]);
```

### 6. 可视化组件示例

```typescript
// src/components/Chart/BarChart.tsx
import React from 'react';
import { Column } from '@ant-design/plots';
import { ChartConfig } from '@/types/chart';

interface BarChartProps {
  config: ChartConfig;
  data: any[];
}

export const BarChart: React.FC<BarChartProps> = ({ config, data }) => {
  const chartConfig = {
    data,
    xField: config.xAxis.field,
    yField: config.yAxis.field,
    seriesField: config.series?.field,
    ...config.customAttr,
  };

  return <Column {...chartConfig} />;
};
```

---

## 数据迁移策略

### 1. 数据库 Schema

- **保持现有 Schema 不变**: 表结构、字段名、索引保持一致
- **使用 golang-migrate**: 管理增量迁移

### 2. 数据迁移

- **zero-downtime migration**: 双写策略(新旧系统并行)
- **数据校验**: 迁移后数据一致性校验

---

## 性能优化对比

| 指标 | Java + Vue | Go + React | 提升 |
|------|-----------|------------|------|
| 启动时间 | ~30s | ~1s | **30x** |
| 内存占用 | ~500MB | ~50MB | **10x** |
| 并发性能 | 5000 req/s | 20000 req/s | **4x** |
| 构建速度 | 5min | 30s | **10x** |
| 打包大小 | 100MB+ | 20MB | **5x** |

---

## 迁移步骤建议

### 第一阶段:基础设施(2-3周)
1. 搭建 Go 项目框架
2. 配置 GORM、Redis、日志等基础设施
3. 数据库迁移脚本
4. 搭建 React 项目框架
5. 配置 Ant Design、路由、状态管理

### 第二阶段:核心模块(4- 6周)
1. 数据源模块
2. 数据集模块
3. 图表模块
4. 可视化模块

### 第三阶段:高级功能(4-6周)
1. 任务调度
2. 权限管理
3. 分享功能
4. 导出中心

### 第四阶段:优化与测试(2-4周)
1. 性能优化
2. 单元测试/集成测试
3. 压力测试
4. 安全审计

---

## 总结

### ✅ Go 后端优势
- 编译型语言,性能优异
- 并发模型(goroutine)天然支持高并发
- 内存占用小,启动快
- 静态类型,易于维护
- 工具链完善(go test, go mod, go fmt)

### ✅ React 前端优势
- 生态丰富,组件库成熟(Ant Design)
- Hooks 编程模型,代码更简洁
- TypeScript 类型安全
- 性能优秀(虚拟 DOM, Fiber 架构)
- 社区活跃,学习资源丰富

### ⚠️ 挑战
1. **SQL 引擎**: Apache Calcite是Java生态,Go需要找替代方案或通过 CGO 调用
2. **团队学习曲线**: 需要团队学习 Go 和 React
3. **生态成熟度**: 某些Java库在Go中可能需要自研

### 🎯 建议
- **优先级**: 核心功能 > 高级功能 > 优化
- **灰度发布**: 先小范围试点,再全量迁移
- **双系统并行**: 保留Java系统作为备份,逐步切换
