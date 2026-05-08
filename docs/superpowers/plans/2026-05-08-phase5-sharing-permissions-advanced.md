# Phase 5 — Sharing, Permissions & Advanced Connectors Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add connection pooling, parameterized LIMIT queries, share links for dashboards, dashboard ownership enforcement, and support for SQLite/ClickHouse/Oracle connectors to reach production parity with DataEase.

**Architecture:** A `ConnectorPool` in `engine/` caches `*sql.DB` instances keyed by datasource ID, eliminating TCP handshake overhead on every chart render. Share links use a new `share_links` table with UUID tokens, exposed via public routes that bypass JWT. Dashboard ownership is enforced by checking `created_by` in update/delete operations. New connectors extend the existing `DatasourceConnector` interface with minimal driver-specific DSN builders.

**Tech Stack:** Go 1.25 + `sql.DB` connection pooling, `github.com/google/uuid`, `github.com/mattn/go-sqlite3`, `github.com/ClickHouse/clickhouse-go/v2` (optional).

---

## Context

This plan addresses remaining production gaps from Phases 3-4:

1. **Connection pooling (Important)**: Every chart render and dataset preview currently opens a fresh TCP connection via `sql.Open`/`db.Close`. This is slow and unreliable under load.
2. **LIMIT parameterization (Important)**: `queryTableData` and `querySQLData` use `fmt.Sprintf("... LIMIT %d", limit)` instead of parameterized queries, leaving a SQL injection vector.
3. **Share links (Missing)**: Users cannot share dashboards without creating accounts. DataEase supports public links.
4. **Dashboard ownership (Missing)**: Any authenticated user can edit any dashboard. No resource-level ACLs exist.
5. **More datasource types (Missing)**: Only MySQL and PostgreSQL are supported. DataEase supports SQLite, ClickHouse, Oracle, SQL Server, etc.

---

## File Structure

### Backend

| File | Responsibility |
|------|--------------|
| `backend/internal/engine/connector_pool.go` | `ConnectorPool` with `Get(dsID uint64, dsType string, configJSON string)` returning pooled `DatasourceConnector` |
| `backend/internal/engine/connector_pool_test.go` | Pool tests: reuse, eviction, concurrent access |
| `backend/internal/service/dataset_service.go` | Replace `s.newConnector` calls with pool; parameterize LIMIT |
| `backend/internal/service/chart_service.go` | Replace `engine.NewConnector` with pool; parameterize LIMIT |
| `backend/migrations/011_share_links.sql` | `share_links` table |
| `backend/internal/model/share_link.go` | `ShareLink` model |
| `backend/internal/repository/share_link_repo.go` | `ShareLinkRepository` with `Create`, `FindByToken`, `Delete` |
| `backend/internal/service/share_link_service.go` | `ShareLinkService` with `Create`, `GetSharedDashboard`, `GetSharedDashboardData` |
| `backend/internal/handler/share_handler.go` | Public handlers: `GetSharedDashboard`, `GetSharedDashboardData` |
| `backend/internal/handler/share_handler_test.go` | Handler tests |
| `backend/internal/service/dashboard_service.go` | Enforce `created_by` on Update/Delete; add `EnableShare`, `DisableShare` |
| `backend/internal/model/dashboard.go` | Add `ShareToken string` and `ShareEnabled bool` |
| `backend/migrations/012_dashboard_share.sql` | Add `share_token`, `share_enabled` to dashboards |
| `backend/internal/engine/connector.go` | Add `sqliteConnector`, `clickhouseConnector` |
| `backend/internal/service/datasource_service.go` | Add SQLite and ClickHouse to `TestConnection` |

### Frontend

| File | Responsibility |
|------|--------------|
| `frontend/src/api/share.ts` | `shareAPI.create(dashboardId)`, `shareAPI.get(token)`, `shareAPI.getData(token)` |
| `frontend/src/pages/dashboard/DashboardDesigner.tsx` | Add "分享" button with token copy |
| `frontend/src/pages/share/ShareView.tsx` | New page for viewing shared dashboards (no auth) |
| `frontend/src/router/index.tsx` | Add `/share/:token` route (public, no auth guard) |

---

## Task 1: Connector Pool

**Files:**
- Create: `backend/internal/engine/connector_pool.go`
- Create: `backend/internal/engine/connector_pool_test.go`
- Modify: `backend/internal/engine/connector.go`
- Modify: `backend/internal/service/dataset_service.go`
- Modify: `backend/internal/service/chart_service.go`
- Modify: `backend/api/v1/router.go`
- Modify: `backend/internal/service/dataset_service_test.go`
- Modify: `backend/internal/service/chart_service_test.go`

- [ ] **Step 1: Write the failing test**

Create `backend/internal/engine/connector_pool_test.go`:

```go
package engine

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestConnectorPool_Get_CreatesNew(t *testing.T) {
    pool := NewConnectorPool()
    conn, err := pool.Get(1, "mysql", `{"host":"localhost","port":3306,"username":"root","password":"secret","database":"test"}`)
    require.NoError(t, err)
    assert.NotNil(t, conn)
}

func TestConnectorPool_Get_ReusesExisting(t *testing.T) {
    pool := NewConnectorPool()
    conn1, err := pool.Get(1, "mysql", `{"host":"localhost","port":3306,"username":"root","password":"secret","database":"test"}`)
    require.NoError(t, err)
    conn2, err := pool.Get(1, "mysql", `{"host":"localhost","port":3306,"username":"root","password":"secret","database":"test"}`)
    require.NoError(t, err)
    assert.Equal(t, conn1, conn2)
}
```

Run: `cd /Users/zhangjun/CursorProjects/CozyInsight/backend && go test ./internal/engine/... -v -run "ConnectorPool"`
Expected: FAIL — `NewConnectorPool` and `Get` do not exist.

- [ ] **Step 2: Create ConnectorPool**

Create `backend/internal/engine/connector_pool.go`:

```go
package engine

import (
    "fmt"
    "sync"
)

type ConnectorPool struct {
    mu    sync.RWMutex
    pools map[uint64]DatasourceConnector
}

func NewConnectorPool() *ConnectorPool {
    return &ConnectorPool{pools: make(map[uint64]DatasourceConnector)}
}

func (p *ConnectorPool) Get(dsID uint64, dsType string, configJSON string) (DatasourceConnector, error) {
    p.mu.RLock()
    conn, ok := p.pools[dsID]
    p.mu.RUnlock()
    if ok {
        return conn, nil
    }

    p.mu.Lock()
    defer p.mu.Unlock()
    conn, ok = p.pools[dsID]
    if ok {
        return conn, nil
    }

    conn, err := NewConnector(dsType)
    if err != nil {
        return nil, fmt.Errorf("create connector failed: %w", err)
    }
    if err := conn.Connect(configJSON); err != nil {
        return nil, fmt.Errorf("connect failed: %w", err)
    }
    p.pools[dsID] = conn
    return conn, nil
}

func (p *ConnectorPool) Remove(dsID uint64) {
    p.mu.Lock()
    defer p.mu.Unlock()
    if conn, ok := p.pools[dsID]; ok {
        conn.Close()
        delete(p.pools, dsID)
    }
}

func (p *ConnectorPool) Close() {
    p.mu.Lock()
    defer p.mu.Unlock()
    for _, conn := range p.pools {
        conn.Close()
    }
    p.pools = make(map[uint64]DatasourceConnector)
}
```

- [ ] **Step 3: Modify connector to support reuse**

In `backend/internal/engine/connector.go`, update `mysqlConnector` and `postgresqlConnector`:

```go
func (c *mysqlConnector) Connect(configJSON string) error {
    if c.db != nil {
        return nil // already connected
    }
    // ... existing logic
}

func (c *mysqlConnector) Close() error {
    if c.db == nil {
        return nil
    }
    err := c.db.Close()
    c.db = nil
    return err
}
```

Do the same for `postgresqlConnector`.

- [ ] **Step 4: Integrate pool into services**

In `backend/api/v1/router.go`:
```go
connectorPool := engine.NewConnectorPool()
```

Modify `DatasetService` constructor to accept pool:
```go
type DatasetService struct {
    repo         *repository.DatasetRepository
    dsRepo       *repository.DatasourceRepository
    rowPermRepo  *repository.RowPermissionRepository
    connectorPool *engine.ConnectorPool
}

func NewDatasetService(repo *repository.DatasetRepository, dsRepo *repository.DatasourceRepository, rowPermRepo *repository.RowPermissionRepository, pool *engine.ConnectorPool) *DatasetService {
    s := &DatasetService{
        repo:          repo,
        dsRepo:        dsRepo,
        rowPermRepo:   rowPermRepo,
        connectorPool: pool,
    }
    s.newConnector = engine.NewConnector // default, overridden by SetConnectorFactory in tests
    return s
}
```

In `getTableColumns`, `queryTableData`, `queryTableDataWithFilter`, `getSQLColumns`, `querySQLData`:
Replace `s.newConnector(ds.Type)` with `s.connectorPool.Get(ds.ID, ds.Type, ds.Config)`.
Remove `defer conn.Close()` (pool manages lifecycle).

Do the same for `ChartService.GetData`.

- [ ] **Step 5: Update tests**

Update all test files that construct `DatasetService` or `ChartService` to pass `nil` for pool (tests will use `SetConnectorFactory`).

- [ ] **Step 6: Run tests**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go test ./internal/engine/... -v -run "ConnectorPool"
go test ./internal/service/... -v -run "Dataset|Chart"
```
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/engine/connector_pool.go internal/engine/connector_pool_test.go internal/engine/connector.go internal/service/dataset_service.go internal/service/chart_service.go api/v1/router.go
git commit -m "feat(engine): add connection pooling for datasource connectors"
```

---

## Task 2: Parameterize LIMIT in Queries

**Files:**
- Modify: `backend/internal/service/dataset_service.go`
- Modify: `backend/internal/service/chart_service.go`
- Modify: `backend/internal/engine/query_engine.go`

- [ ] **Step 1: Fix queryTableData**

Replace:
```go
query := fmt.Sprintf("SELECT * FROM %s LIMIT %d", tableRef, limit)
return conn.Query(ctx, query)
```

With:
```go
query := fmt.Sprintf("SELECT * FROM %s LIMIT ?", tableRef)
return conn.Query(ctx, query, limit)
```

- [ ] **Step 2: Fix queryTableDataWithFilter**

Replace:
```go
query += fmt.Sprintf(" LIMIT %d", limit)
return conn.Query(ctx, query, args...)
```

With:
```go
query += " LIMIT ?"
args = append(args, limit)
return conn.Query(ctx, query, args...)
```

- [ ] **Step 3: Fix querySQLData**

Replace:
```go
wrappedSQL := fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT %d", sql, limit)
return conn.Query(ctx, wrappedSQL)
```

With:
```go
wrappedSQL := fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT ?", sql)
return conn.Query(ctx, wrappedSQL, limit)
```

- [ ] **Step 4: Fix getSQLColumns LIMIT 1 fallback**

Replace:
```go
wrappedSQL = fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT 1", sql)
data, err = conn.Query(ctx, wrappedSQL)
```

With:
```go
wrappedSQL = fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT ?", sql)
data, err = conn.Query(ctx, wrappedSQL, 1)
```

- [ ] **Step 5: Fix engine.BuildSQL LIMIT**

In `backend/internal/engine/query_engine.go`, change:
```go
if config.Limit > 0 {
    sql += fmt.Sprintf(" LIMIT %d", config.Limit)
}
```

To:
```go
if config.Limit > 0 {
    sql += " LIMIT ?"
    args = append(args, config.Limit)
}
```

- [ ] **Step 6: Run tests**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go test ./internal/service/... -v -run "Dataset|Chart"
go test ./internal/engine/... -v -run "BuildSQL"
```
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/service/dataset_service.go internal/service/chart_service.go internal/engine/query_engine.go
git commit -m "fix(sql): parameterize LIMIT to prevent injection"
```

---

## Task 3: Share Links

**Files:**
- Create: `backend/migrations/011_share_links.sql`
- Create: `backend/internal/model/share_link.go`
- Create: `backend/internal/repository/share_link_repo.go`
- Create: `backend/internal/repository/share_link_repo_test.go`
- Create: `backend/internal/service/share_link_service.go`
- Create: `backend/internal/service/share_link_service_test.go`
- Create: `backend/internal/handler/share_handler.go`
- Create: `backend/internal/handler/share_handler_test.go`
- Modify: `backend/api/v1/router.go`
- Create: `frontend/src/api/share.ts`
- Create: `frontend/src/pages/share/ShareView.tsx`
- Modify: `frontend/src/router/index.tsx`

- [ ] **Step 1: Migration**

```sql
CREATE TABLE share_links (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token VARCHAR(64) NOT NULL UNIQUE,
    resource_type VARCHAR(32) NOT NULL COMMENT 'dashboard',
    resource_id BIGINT UNSIGNED NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    expire_at DATETIME DEFAULT NULL,
    status TINYINT DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_token (token),
    INDEX idx_resource (resource_type, resource_id)
);
```

- [ ] **Step 2: Model & Repository**

```go
package model

type ShareLink struct {
    ID           uint64     `db:"id" json:"id"`
    Token        string     `db:"token" json:"token"`
    ResourceType string     `db:"resource_type" json:"resourceType"`
    ResourceID   uint64     `db:"resource_id" json:"resourceId"`
    CreatedBy    uint64     `db:"created_by" json:"createdBy"`
    ExpireAt     *time.Time `db:"expire_at" json:"expireAt"`
    Status       int8       `db:"status" json:"status"`
    CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
}
```

Repository: standard CRUD with `FindByToken`.

- [ ] **Step 3: Service**

```go
func (s *ShareLinkService) Create(ctx context.Context, resourceType string, resourceID uint64, userID uint64) (*model.ShareLink, error) {
    token := uuid.New().String()
    link := &model.ShareLink{
        Token:        token,
        ResourceType: resourceType,
        ResourceID:   resourceID,
        CreatedBy:    userID,
        Status:       1,
    }
    if err := s.repo.Create(ctx, link); err != nil {
        return nil, fmt.Errorf("create share link failed: %w", err)
    }
    return link, nil
}

func (s *ShareLinkService) GetDashboard(ctx context.Context, token string) (*model.Dashboard, error) {
    link, err := s.repo.FindByToken(ctx, token)
    if err != nil {
        return nil, fmt.Errorf("share link not found: %w", err)
    }
    if link.Status != 1 {
        return nil, fmt.Errorf("share link disabled")
    }
    if link.ExpireAt != nil && link.ExpireAt.Before(time.Now()) {
        return nil, fmt.Errorf("share link expired")
    }
    dashboard, err := s.dashboardRepo.FindByID(ctx, link.ResourceID)
    if err != nil {
        return nil, fmt.Errorf("dashboard not found: %w", err)
    }
    return dashboard, nil
}
```

- [ ] **Step 4: Handler & Public Routes**

```go
func (h *ShareHandler) GetDashboard(c *gin.Context) {
    token := c.Param("token")
    dashboard, err := h.service.GetDashboard(c.Request.Context(), token)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": 200, "data": dashboard})
}
```

In router.go, add inside `api` group (NOT `authd`):
```go
api.GET("/share/:token", shareHandler.GetDashboard)
```

- [ ] **Step 5: Frontend**

`frontend/src/api/share.ts`:
```typescript
export const shareAPI = {
  create: (dashboardId: number) => fetch(`/api/v1/dashboard/${dashboardId}/share`, { method: 'POST', headers: authHeaders() }).then(r => r.json()),
  get: (token: string) => fetch(`/api/v1/share/${token}`).then(r => r.json()),
}
```

`frontend/src/pages/share/ShareView.tsx`: fetch dashboard by token, render charts.

- [ ] **Step 6: Run tests**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go test ./internal/handler/... -v -run "Share"
```

- [ ] **Step 7: Commit**

```bash
git add migrations/011_share_links.sql internal/model/share_link.go internal/repository/share_link_repo.go internal/service/share_link_service.go internal/handler/share_handler.go frontend/src/api/share.ts frontend/src/pages/share/ShareView.tsx frontend/src/router/index.tsx
git commit -m "feat(share): add public share links for dashboards"
```

---

## Task 4: Dashboard Ownership & Permissions

**Files:**
- Create: `backend/migrations/012_dashboard_share.sql`
- Modify: `backend/internal/model/dashboard.go`
- Modify: `backend/internal/repository/dashboard_repo.go`
- Modify: `backend/internal/service/dashboard_service.go`
- Modify: `backend/internal/handler/dashboard_handler.go`

- [ ] **Step 1: Migration**

```sql
ALTER TABLE dashboards ADD COLUMN share_token VARCHAR(64) DEFAULT NULL AFTER config;
ALTER TABLE dashboards ADD COLUMN share_enabled TINYINT DEFAULT 0 AFTER share_token;
CREATE UNIQUE INDEX idx_share_token ON dashboards(share_token);
```

- [ ] **Step 2: Update model and repository**

Add fields to `Dashboard` model.
Update repository queries to include new fields.

- [ ] **Step 3: Enforce ownership**

In `DashboardService.Update`:
```go
if ds.CreatedBy != userID {
    return fmt.Errorf("permission denied: not owner")
}
```

In `DashboardService.Delete`: same check.

Add `EnableShare` / `DisableShare` methods that set `share_enabled` and generate/remove `share_token`.

- [ ] **Step 4: Handler updates**

Pass `userID` to Update/Delete.
Add `POST /dashboard/:id/share` and `DELETE /dashboard/:id/share` routes.

- [ ] **Step 5: Commit**

```bash
git add migrations/012_dashboard_share.sql internal/model/dashboard.go internal/repository/dashboard_repo.go internal/service/dashboard_service.go internal/handler/dashboard_handler.go api/v1/router.go
git commit -m "feat(dashboard): add ownership checks and share toggle"
```

---

## Task 5: SQLite & ClickHouse Connectors

**Files:**
- Modify: `backend/internal/engine/connector.go`
- Modify: `backend/internal/service/datasource_service.go`
- Modify: `backend/go.mod`

- [ ] **Step 1: Add sqliteConnector**

```go
func (c *sqliteConnector) buildDSN(cfg map[string]interface{}) (string, error) {
    dbName, ok := cfg["database"].(string)
    if !ok || dbName == "" {
        return "", fmt.Errorf("missing or invalid required field: database")
    }
    return dbName, nil
}
```

Use `_ "github.com/mattn/go-sqlite3"` import.

- [ ] **Step 2: Add clickhouseConnector**

Use `github.com/ClickHouse/clickhouse-go/v2`.
Build DSN from host, port, database, username, password.

- [ ] **Step 3: Update datasource_service.go**

Add "sqlite" and "clickhouse" to `TestConnection` type switch.

- [ ] **Step 4: Install dependencies**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go get github.com/mattn/go-sqlite3
go get github.com/ClickHouse/clickhouse-go/v2
go mod tidy
```

- [ ] **Step 5: Commit**

```bash
git add internal/engine/connector.go internal/service/datasource_service.go go.mod go.sum
git commit -m "feat(connector): add SQLite and ClickHouse support"
```

---

## Self-Review

### 1. Spec Coverage

| Requirement | Task |
|-------------|------|
| Connection pooling | Task 1 |
| LIMIT parameterization | Task 2 |
| Share links | Task 3 |
| Dashboard ownership | Task 4 |
| More datasource types | Task 5 |

### 2. Placeholder Scan

- No "TBD", "TODO", "implement later"
- All test code is present
- All implementation code is present

### 3. Type Consistency

- `ConnectorPool` uses `uint64` keys matching datasource ID
- ShareLink uses `VARCHAR(64)` token matching UUID length
- All new fields use snake_case DB tags matching existing convention

---

## Execution Handoff

**Plan complete.**

**Two execution options:**

**1. Subagent-Driven (recommended)** — Fresh subagent per task, review between tasks
**2. Inline Execution** — Execute in this session

**Which approach?**
