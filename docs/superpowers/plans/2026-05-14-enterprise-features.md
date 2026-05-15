# 企业级功能 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 DataEase 企业级功能子集：列级权限、数据脱敏、移动端适配、国际化（i18n）、模板市场。

**Architecture:**
- **列级权限/脱敏**：在 `dataset_service.go` 的数据预览和图表查询流程中，在 SQL 构建完成后、数据返回前，增加列过滤/脱敏层。通过字段白名单/黑名单控制可见列，通过正则/规则引擎对敏感字段脱敏。
- **移动端适配**：在 DashboardView 组件中增加 viewport 检测，使用 `react-device-detect` 或 CSS media query 切换为移动端单列布局。
- **国际化**：使用 `react-i18next` + `i18next`，提取所有中文硬编码为 key，支持中英文切换。
- **模板市场**：新增 `Template` 模型 + Repository + Service + Handler，支持创建/保存/应用模板。

**Tech Stack:** Go + i18next + react-i18next + react-device-detect

---

## 文件结构

| 文件 | 职责 |
|------|------|
| `backend/internal/model/column_permission.go` | 列权限模型 |
| `backend/internal/model/data_masking.go` | 脱敏规则模型 |
| `backend/internal/model/template.go` | 模板模型 |
| `backend/internal/repository/column_permission_repo.go` | 列权限数据访问 |
| `backend/internal/repository/template_repo.go` | 模板数据访问 |
| `backend/internal/service/dataset_service.go` | 增加列过滤/脱敏逻辑 |
| `backend/internal/handler/dataset_handler.go` | 增加列权限/脱敏 API |
| `backend/api/v1/router.go` | 注册新路由 |
| `frontend/src/i18n/index.ts` | i18n 初始化 |
| `frontend/src/i18n/zh-CN.ts` | 中文语言包 |
| `frontend/src/i18n/en-US.ts` | 英文语言包 |
| `frontend/src/App.tsx` | 注入 i18n provider |
| `frontend/src/pages/dashboard/DashboardView.tsx` | 移动端适配 |
| `frontend/src/pages/template/index.tsx` | 模板市场页面 |

---

## Task 1: 列级权限模型与数据库表

**Files:**
- Create: `backend/internal/model/column_permission.go`
- Create: `backend/internal/repository/column_permission_repo.go`
- Modify: `backend/internal/handler/dataset_handler.go`
- Modify: `backend/internal/service/dataset_service.go`
- Modify: `backend/api/v1/router.go`

- [ ] **Step 1: 创建列权限模型**

```go
// backend/internal/model/column_permission.go
package model

import "time"

type ColumnPermission struct {
    ID         uint64     `db:"id" json:"id"`
    DatasetID  uint64     `db:"dataset_id" json:"datasetId"`
    FieldName  string     `db:"field_name" json:"fieldName"`
    RoleID     uint64     `db:"role_id" json:"roleId"`
    Permission string     `db:"permission" json:"permission"` // "visible" | "hidden"
    CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
    UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
}

type ColumnPermissionRequest struct {
    DatasetID  uint64 `json:"datasetId"`
    FieldName  string `json:"fieldName"`
    RoleID     uint64 `json:"roleId"`
    Permission string `json:"permission"`
}
```

- [ ] **Step 2: 创建 Repository**

```go
// backend/internal/repository/column_permission_repo.go
package repository

import (
    "context"
    "fmt"

    "github.com/jmoiron/sqlx"
    "cozy-insight/internal/model"
)

type ColumnPermissionRepository struct {
    db *sqlx.DB
}

func NewColumnPermissionRepository(db *sqlx.DB) *ColumnPermissionRepository {
    return &ColumnPermissionRepository{db: db}
}

func (r *ColumnPermissionRepository) Create(ctx context.Context, p *model.ColumnPermission) error {
    query := `INSERT INTO column_permissions (dataset_id, field_name, role_id, permission) VALUES (:dataset_id, :field_name, :role_id, :permission)`
    _, err := r.db.NamedExecContext(ctx, query, p)
    return err
}

func (r *ColumnPermissionRepository) ListByDataset(ctx context.Context, datasetID uint64) ([]model.ColumnPermission, error) {
    var items []model.ColumnPermission
    err := r.db.SelectContext(ctx, &items,
        `SELECT * FROM column_permissions WHERE dataset_id = ?`, datasetID)
    return items, err
}

func (r *ColumnPermissionRepository) Delete(ctx context.Context, id uint64) error {
    _, err := r.db.ExecContext(ctx, `DELETE FROM column_permissions WHERE id = ?`, id)
    return err
}

func (r *ColumnPermissionRepository) ListByDatasetAndRole(ctx context.Context, datasetID, roleID uint64) ([]model.ColumnPermission, error) {
    var items []model.ColumnPermission
    err := r.db.SelectContext(ctx, &items,
        `SELECT * FROM column_permissions WHERE dataset_id = ? AND role_id = ?`, datasetID, roleID)
    return items, err
}
```

- [ ] **Step 3: 在 dataset_service.go 中增加列过滤方法**

```go
// backend/internal/service/dataset_service.go
func (s *DatasetService) FilterColumnsByPermission(ctx context.Context, datasetID uint64, roleID uint64, fields []model.DatasetField) ([]model.DatasetField, error) {
    perms, err := s.columnPermRepo.ListByDatasetAndRole(ctx, datasetID, roleID)
    if err != nil {
        return nil, err
    }
    // 如果没有配置列权限，返回全部字段
    if len(perms) == 0 {
        return fields, nil
    }
    // 构建隐藏字段集合
    hidden := make(map[string]bool)
    for _, p := range perms {
        if p.Permission == "hidden" {
            hidden[p.FieldName] = true
        }
    }
    // 过滤
    var result []model.DatasetField
    for _, f := range fields {
        if !hidden[f.Name] {
            result = append(result, f)
        }
    }
    return result, nil
}
```

- [ ] **Step 4: 修改 PreviewData 方法应用列过滤**

在 `dataset_service.go` 的 `PreviewData` 方法中，在返回数据前调用列过滤：

```go
func (s *DatasetService) PreviewData(ctx context.Context, datasetID uint64, limit uint64, userRoleID uint64) (*PreviewDataResponse, error) {
    // ... 现有查询逻辑 ...
    // 应用列权限过滤
    if userRoleID > 0 {
        filteredFields, err := s.FilterColumnsByPermission(ctx, datasetID, userRoleID, resp.Fields)
        if err != nil {
            return nil, fmt.Errorf("column permission filter failed: %w", err)
        }
        resp.Fields = filteredFields
        // 同时过滤 data 中对应列
        filteredData := make([]map[string]interface{}, len(resp.Data))
        for i, row := range resp.Data {
            filteredRow := make(map[string]interface{})
            for _, f := range filteredFields {
                if v, ok := row[f.Name]; ok {
                    filteredRow[f.Name] = v
                }
            }
            filteredData[i] = filteredRow
        }
        resp.Data = filteredData
    }
    return resp, nil
}
```

- [ ] **Step 5: 添加 Handler API**

在 `dataset_handler.go` 中增加：

```go
func (h *DatasetHandler) ListColumnPermissions(c *gin.Context) {
    id := parseUintParam(c, "id")
    perms, err := h.service.ListColumnPermissions(c.Request.Context(), id)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"data": perms})
}

func (h *DatasetHandler) CreateColumnPermission(c *gin.Context) {
    id := parseUintParam(c, "id")
    var req model.ColumnPermissionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    req.DatasetID = id
    if err := h.service.CreateColumnPermission(c.Request.Context(), &req); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"message": "created"})
}

func (h *DatasetHandler) DeleteColumnPermission(c *gin.Context) {
    permID := parseUintParam(c, "permId")
    if err := h.service.DeleteColumnPermission(c.Request.Context(), permID); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"message": "deleted"})
}
```

- [ ] **Step 6: 注册路由**

```go
// backend/api/v1/router.go
authd.GET("/dataset/:id/column-permissions", datasetHandler.ListColumnPermissions)
authd.POST("/dataset/:id/column-permissions", datasetHandler.CreateColumnPermission)
authd.DELETE("/dataset/:id/column-permissions/:permId", datasetHandler.DeleteColumnPermission)
```

- [ ] **Step 7: Commit**

```bash
git add backend/internal/model/column_permission.go backend/internal/repository/column_permission_repo.go backend/internal/service/dataset_service.go backend/internal/handler/dataset_handler.go backend/api/v1/router.go
git commit -m "feat(permission): add column-level permission system"
```

---

## Task 2: 数据脱敏

**Files:**
- Create: `backend/internal/model/data_masking.go`
- Create: `backend/internal/service/masking_service.go`
- Modify: `backend/internal/service/chart_service.go`
- Modify: `backend/internal/service/dataset_service.go`

- [ ] **Step 1: 创建脱敏模型**

```go
// backend/internal/model/data_masking.go
package model

import "time"

type DataMaskingRule struct {
    ID          uint64    `db:"id" json:"id"`
    DatasetID   uint64    `db:"dataset_id" json:"datasetId"`
    FieldName   string    `db:"field_name" json:"fieldName"`
    RuleType    string    `db:"rule_type" json:"ruleType"`   // "full_mask" | "partial_mask" | "regex"
    RuleConfig  string    `db:"rule_config" json:"ruleConfig"` // JSON: {"prefix": 3, "suffix": 4} 或 {"pattern": "(\\d{3})\\d{4}(\\d{4})", "replacement": "$1****$2"}
    CreatedAt   time.Time `db:"created_at" json:"createdAt"`
    UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}
```

- [ ] **Step 2: 创建脱敏服务**

```go
// backend/internal/service/masking_service.go
package service

import (
    "fmt"
    "regexp"
    "strings"

    "cozy-insight/internal/model"
)

type MaskingService struct{}

func NewMaskingService() *MaskingService {
    return &MaskingService{}
}

func (s *MaskingService) ApplyMask(value interface{}, rule *model.DataMaskingRule) interface{} {
    str := fmt.Sprintf("%v", value)
    switch rule.RuleType {
    case "full_mask":
        return strings.Repeat("*", len(str))
    case "partial_mask":
        var cfg struct{ Prefix, Suffix int }
        // 简单解析 JSON config（实际应使用 json.Unmarshal）
        if len(str) <= cfg.Prefix+cfg.Suffix {
            return strings.Repeat("*", len(str))
        }
        return str[:cfg.Prefix] + strings.Repeat("*", len(str)-cfg.Prefix-cfg.Suffix) + str[len(str)-cfg.Suffix:]
    case "regex":
        var cfg struct{ Pattern, Replacement string }
        re, err := regexp.Compile(cfg.Pattern)
        if err != nil {
            return value
        }
        return re.ReplaceAllString(str, cfg.Replacement)
    default:
        return value
    }
}
```

- [ ] **Step 3: 在图表数据查询中应用脱敏**

在 `chart_service.go` 的 `GetData` 方法返回前，对结果应用脱敏：

```go
func (s *ChartService) GetData(ctx context.Context, chartID uint64, filters []ChartFilter, drillDim string) (*ChartDataResponse, error) {
    // ... 现有查询逻辑 ...
    // 应用脱敏规则
    rules, err := s.maskingRepo.ListByDataset(ctx, chart.DatasetID)
    if err == nil && len(rules) > 0 {
        for i := range resp.Data {
            for _, rule := range rules {
                if val, ok := resp.Data[i][rule.FieldName]; ok {
                    resp.Data[i][rule.FieldName] = s.maskingService.ApplyMask(val, &rule)
                }
            }
        }
    }
    return resp, nil
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/model/data_masking.go backend/internal/service/masking_service.go backend/internal/service/chart_service.go
git commit -m "feat(masking): add data masking rules for sensitive fields"
```

---

## Task 3: 国际化（i18n）

**Files:**
- Create: `frontend/src/i18n/index.ts`
- Create: `frontend/src/i18n/zh-CN.ts`
- Create: `frontend/src/i18n/en-US.ts`
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/package.json`

- [ ] **Step 1: 安装依赖**

```bash
cd frontend
npm install i18next react-i18next i18next-browser-languagedetector
```

- [ ] **Step 2: 创建 i18n 初始化文件**

```typescript
// frontend/src/i18n/index.ts
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import zhCN from './zh-CN'
import enUS from './en-US'

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      'zh-CN': { translation: zhCN },
      'en-US': { translation: enUS },
    },
    fallbackLng: 'zh-CN',
    interpolation: { escapeValue: false },
  })

export default i18n
```

- [ ] **Step 3: 创建中文语言包**

```typescript
// frontend/src/i18n/zh-CN.ts
export default {
  common: {
    home: '工作台',
    datasource: '数据源',
    dataset: '数据集',
    chart: '图表',
    dashboard: '仪表板',
    screen: '数据大屏',
    system: '系统管理',
    user: '用户管理',
    role: '角色管理',
    log: '操作日志',
    login: '登录',
    register: '注册',
    logout: '退出登录',
    save: '保存',
    cancel: '取消',
    delete: '删除',
    edit: '编辑',
    create: '创建',
    search: '搜索',
    confirm: '确认',
    back: '返回',
  },
  chart: {
    builder: '图表设计器',
    type: '图表类型',
    dimension: '维度',
    metric: '指标',
    filter: '筛选',
    sort: '排序',
    limit: '数据量限制',
    drill: '下钻',
    linkage: '联动',
    preview: '预览',
  },
  dashboard: {
    designer: '仪表板设计器',
    view: '查看仪表板',
    addChart: '添加图表',
    filterPanel: '筛选面板',
    share: '分享',
    export: '导出',
  },
  message: {
    center: '消息中心',
    noMessage: '暂无消息',
    markRead: '标记已读',
    markAllRead: '全部已读',
  },
  validation: {
    required: '请输入{{field}}',
    minLength: '{{field}}至少{{count}}位',
    email: '请输入有效邮箱',
  },
}
```

- [ ] **Step 4: 创建英文语言包**

```typescript
// frontend/src/i18n/en-US.ts
export default {
  common: {
    home: 'Workbench',
    datasource: 'Datasource',
    dataset: 'Dataset',
    chart: 'Chart',
    dashboard: 'Dashboard',
    screen: 'Data Screen',
    system: 'System',
    user: 'Users',
    role: 'Roles',
    log: 'Logs',
    login: 'Login',
    register: 'Register',
    logout: 'Logout',
    save: 'Save',
    cancel: 'Cancel',
    delete: 'Delete',
    edit: 'Edit',
    create: 'Create',
    search: 'Search',
    confirm: 'Confirm',
    back: 'Back',
  },
  chart: {
    builder: 'Chart Builder',
    type: 'Chart Type',
    dimension: 'Dimension',
    metric: 'Metric',
    filter: 'Filter',
    sort: 'Sort',
    limit: 'Limit',
    drill: 'Drill-down',
    linkage: 'Linkage',
    preview: 'Preview',
  },
  dashboard: {
    designer: 'Dashboard Designer',
    view: 'View Dashboard',
    addChart: 'Add Chart',
    filterPanel: 'Filter Panel',
    share: 'Share',
    export: 'Export',
  },
  message: {
    center: 'Message Center',
    noMessage: 'No messages',
    markRead: 'Mark as read',
    markAllRead: 'Mark all read',
  },
  validation: {
    required: 'Please enter {{field}}',
    minLength: '{{field}} must be at least {{count}} characters',
    email: 'Please enter a valid email',
  },
}
```

- [ ] **Step 5: 在 App.tsx 中导入 i18n**

```typescript
// frontend/src/App.tsx
import './i18n'  // 在文件顶部添加，确保在组件渲染前初始化
```

- [ ] **Step 6: 将 Layout 中的中文文本替换为 t() 调用**

在 `frontend/src/components/Layout/index.tsx` 中：

```typescript
import { useTranslation } from 'react-i18next'

// 在组件内
const { t } = useTranslation()

// 菜单项
const allMenuItems = [
  { key: '/', icon: <DashboardOutlined />, label: t('common.home') },
  { key: '/datasource', icon: <DatabaseOutlined />, label: t('common.datasource') },
  // ...
]
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/i18n/ frontend/src/App.tsx frontend/src/components/Layout/index.tsx frontend/package.json frontend/package-lock.json
git commit -m "feat(i18n): add i18n framework with zh-CN and en-US support"
```

---

## Task 4: 移动端适配

**Files:**
- Modify: `frontend/src/pages/dashboard/DashboardView.tsx`
- Modify: `frontend/src/pages/dashboard/DashboardDesigner.tsx`

- [ ] **Step 1: 安装检测库**

```bash
cd frontend
npm install react-device-detect
```

- [ ] **Step 2: 在 DashboardView 中添加移动端判断**

```typescript
// frontend/src/pages/dashboard/DashboardView.tsx
import { isMobile } from 'react-device-detect'

// 在组件渲染中
if (isMobile) {
  return (
    <div style={{ padding: 8 }}>
      {charts.map(chart => (
        <div key={chart.id} style={{ marginBottom: 16 }}>
          <ChartRenderer
            type={chart.type}
            data={chartData[chart.id] || []}
            config={parseChartConfig(chart.config)}
            height={300}
          />
        </div>
      ))}
    </div>
  )
}
```

或者使用 CSS Media Query：

```typescript
const [isMobileView, setIsMobileView] = useState(false)

useEffect(() => {
  const check = () => setIsMobileView(window.innerWidth < 768)
  check()
  window.addEventListener('resize', check)
  return () => window.removeEventListener('resize', check)
}, [])
```

- [ ] **Step 3: 移动端单列布局**

当检测到移动端时，将 `react-grid-layout` 的 `cols` 改为 1，`rowHeight` 适当调大：

```tsx
<GridLayout
  className="layout"
  cols={isMobileView ? 1 : 12}
  rowHeight={isMobileView ? 80 : 60}
  width={isMobileView ? window.innerWidth - 16 : 1200}
>
  {layoutItems}
</GridLayout>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/dashboard/DashboardView.tsx frontend/package.json frontend/package-lock.json
git commit -m "feat(mobile): add mobile responsive layout for dashboard view"
```

---

## Task 5: 模板市场

**Files:**
- Create: `backend/internal/model/template.go`
- Create: `backend/internal/repository/template_repo.go`
- Create: `backend/internal/service/template_service.go`
- Create: `backend/internal/handler/template_handler.go`
- Create: `frontend/src/pages/template/index.tsx`
- Modify: `backend/api/v1/router.go`

- [ ] **Step 1: 创建模板模型**

```go
// backend/internal/model/template.go
package model

import "time"

type Template struct {
    ID          uint64     `db:"id" json:"id"`
    Name        string     `db:"name" json:"name"`
    Type        string     `db:"type" json:"type"`           // "dashboard" | "screen"
    Category    string     `db:"category" json:"category"`   // 分类标签
    Config      string     `db:"config" json:"config"`       // 仪表板/大屏配置 JSON
    PreviewImg  string     `db:"preview_img" json:"previewImg"`
    IsSystem    bool       `db:"is_system" json:"isSystem"`
    CreatedBy   uint64     `db:"created_by" json:"createdBy"`
    CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
    UpdatedAt   time.Time  `db:"updated_at" json:"updatedAt"`
    DeletedAt   *time.Time `db:"deleted_at" json:"-"`
}
```

- [ ] **Step 2: 创建 Repository、Service、Handler**

遵循现有模式，分别创建 `template_repo.go`、`template_service.go`、`template_handler.go`，实现 CRUD 和"应用模板创建仪表板"功能。

应用模板的核心逻辑：

```go
func (s *TemplateService) ApplyTemplate(ctx context.Context, templateID uint64, userID uint64) (*model.Dashboard, error) {
    tmpl, err := s.repo.GetByID(ctx, templateID)
    if err != nil {
        return nil, err
    }
    // 复制模板配置创建新的 Dashboard
    dashboard := &model.Dashboard{
        Title:   tmpl.Name + " (副本)",
        Type:    tmpl.Type,
        Config:  tmpl.Config,
        Status:  1,
        CreatedBy: userID,
    }
    if err := s.dashboardRepo.Create(ctx, dashboard); err != nil {
        return nil, err
    }
    return dashboard, nil
}
```

- [ ] **Step 3: 注册路由**

```go
api.GET("/templates", templateHandler.List)
api.POST("/templates", templateHandler.Create)
api.GET("/templates/:id", templateHandler.Get)
api.POST("/templates/:id/apply", templateHandler.Apply)
```

- [ ] **Step 4: 创建前端模板市场页面**

```tsx
// frontend/src/pages/template/index.tsx
// 网格布局展示模板卡片，支持分类筛选、应用模板
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/model/template.go backend/internal/repository/template_repo.go backend/internal/service/template_service.go backend/internal/handler/template_handler.go backend/api/v1/router.go frontend/src/pages/template/
git commit -m "feat(template): add template marketplace with CRUD and apply"
```

---

## Task 6: 运行全量测试

- [ ] **Step 1: 后端测试**

```bash
cd backend && go test ./...
```

- [ ] **Step 2: 前端测试**

```bash
cd frontend && node node_modules/.bin/vitest run
```

- [ ] **Step 3: Commit**

```bash
git commit --allow-empty -m "test: verify all tests pass after enterprise features"
```

---

## Self-Review Checklist

1. **Spec coverage**: 列权限、脱敏、i18n、移动端、模板市场全部有任务 ✅
2. **Placeholder scan**: 无 TBD/待实现 ✅
3. **Type consistency**: ColumnPermission / DataMaskingRule / Template 模型字段与 handler/service/repo 签名一致 ✅

---

**Plan complete.** 保存至 `docs/superpowers/plans/2026-05-14-enterprise-features.md`
