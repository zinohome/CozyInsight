# CozyInsight 性能优化指南

## 后端性能优化

### 1. 数据库优化

#### 索引策略
```sql
-- 数据源表
CREATE INDEX idx_datasource_type ON datasource(type);
CREATE INDEX idx_datasource_create_time ON datasource(create_time);

-- 数据集表
CREATE INDEX idx_dataset_table_group ON dataset_table(dataset_group_id);
CREATE INDEX idx_dataset_field_table ON dataset_table_field(dataset_table_id);

-- 图表表
CREATE INDEX idx_chart_scene ON core_chart_view(scene_id);
CREATE INDEX idx_chart_table ON core_chart_view(table_id);

-- 权限表
CREATE INDEX idx_resource_perm_type_id ON sys_resource_permission(resource_type, resource_id);
CREATE INDEX idx_user_role_user ON sys_user_role(user_id);
```

#### 连接池配置
```yaml
database:
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime: 3600  # 1小时
```

### 2. 缓存策略

#### Redis缓存配置
```yaml
redis:
  host: 127.0.0.1
  port: 6379
  pool_size: 50
  min_idle_conns: 10
  max_retries: 3
```

#### 缓存层次
```
L1: 本地内存缓存 (1分钟)
  - 权限检查结果
  - 用户角色信息

L2: Redis缓存 (5-30分钟)
  - SQL查询结果
  - 数据集字段信息
  - 图表配置
```

### 3. Calcite性能优化

#### JVM参数
```bash
java -Xms1g -Xmx2g \
  -XX:+UseG1GC \
  -XX:MaxGCPauseMillis=200 \
  -jar avatica-server.jar
```

#### 连接池
```yaml
calcite:
  avatica_url: "http://localhost:8765/"
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime: 3600
```

### 4. API优化

#### 分页查询
```go
// 默认分页限制
const (
    DefaultPageSize = 20
    MaxPageSize     = 1000
)

// 使用游标分页(大数据量)
type CursorPagination struct {
    Cursor string
    Limit  int
}
```

#### 批量操作
```go
// 批量创建字段
func BatchCreateFields(ctx context.Context, fields []*model.DatasetTableField) error {
    batchSize := 500
    for i := 0; i < len(fields); i += batchSize {
        end := i + batchSize
        if end > len(fields) {
            end = len(fields)
        }
        batch := fields[i:end]
        // 批量插入
    }
}
```

---

## 前端性能优化

### 1. 代码分割

#### 路由懒加载
```typescript
const DashboardList = lazy(() => import('./pages/dashboard'));
const ChartEditor = lazy(() => import('./pages/chart/ChartEditor'));
```

#### 组件懒加载
```typescript
const ChartRenderer = lazy(() => import('./components/chart/ChartRenderer'));
```

### 2. 缓存策略

#### React Query配置
```typescript
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5分钟
      cacheTime: 10 * 60 * 1000, // 10分钟
      retry: 1,
    },
  },
});
```

### 3. 渲染优化

#### 虚拟列表
```typescript
// 大数据量表格使用虚拟滚动
import { FixedSizeList } from 'react-window';
```

#### 防抖/节流
```typescript
const debouncedSearch = useDebouncedCallback(
  (value: string) => {
    search(value);
  },
  300
);
```

### 4. 图表性能

#### 数据采样
```typescript
// 超过1000点自动采样
if (data.length > 1000) {
  data = sampleData(data, 1000);
}
```

#### 按需渲染
```typescript
const ChartContainer = ({ visible, ...props }) => {
  if (!visible) return null;
  return <Chart {...props} />;
};
```

---

## 生产部署优化

### 1. 前端构建

```bash
# 生产构建
npm run build

# 开启压缩
vite build --minify
```

### 2. Nginx配置

```nginx
# 开启gzip压缩
gzip on;
gzip_types text/plain text/css application/json application/javascript;
gzip_min_length 1000;

# 静态资源缓存
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}

# API代理
location /api/ {
    proxy_pass http://backend:8100;
    proxy_cache api_cache;
    proxy_cache_valid 200 5m;
}
```

### 3. CDN加速

```typescript
// 使用CDN加载第三方库
<script src="https://cdn.jsdelivr.net/npm/react@18/umd/react.production.min.js"></script>
```

---

## 监控指标

### 关键指标

```
后端:
- API响应时间 P95 < 500ms
- QPS > 1000
- 数据库连接池使用率 < 80%
- 缓存命中率 > 60%

前端:
- FCP < 1.5s
- LCP < 2.5s
- TTI < 3.5s
- CLS < 0.1
```

---

**性能优化是持续过程,需要根据实际情况调整!**
