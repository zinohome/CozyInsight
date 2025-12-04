---
trigger: always_on
---

# DataEase Go + React 重构项目 - AI 编码规范

## 项目背景
正在将 DataEase (Java + Vue) 重构为 Go + React,要求**完整复刻**所有功能。

## 核心原则

### 1. 功能对等性 (最高优先级)
- ✅ **必须**: 与原 Java + Vue 版本功能完全一致
- ✅ **必须**: 每个模块开发完成后与原代码对照验证
- ✅ **必须**: API 接口保持兼容(URL、参数、响应格式)
- ✅ **必须**: 数据库 Schema 保持不变

### 2. 代码质量
- ✅ 简洁优于复杂(KISS 原则)
- ✅ 可读性优于技巧性
- ✅ 类型安全(Go 严格类型, TypeScript strict mode)
- ✅ 错误处理完善(不允许 panic,前端不允许未捕获异常)

### 3. 性能要求
- ✅ 启动时间 < 2s
- ✅ API 响应时间 P95 < 500ms
- ✅ 内存占用 < 200MB (Go Backend)
- ✅ 前端首屏加载 < 2s

---

## Go 后端开发规范

### 项目结构
```
backend/
├── cmd/server/main.go          # 入口
├── internal/                   # 私有包
│   ├── handler/               # HTTP handlers
│   ├── service/               # 业务逻辑
│   ├── repository/            # 数据访问
│   ├── model/                 # 数据模型
│   └── engine/                # SQL 引擎 (Avatica)
├── pkg/                       # 公共包
└── configs/                   # 配置文件
```

### 命名规范
- **文件**: 小写下划线 `user_service.go`
- **包**: 小写单数 `package service`
- **接口**: 名词 `type UserService interface`
- **实现**: 小写首字母 `type userService struct`
- **方法**: 大写导出, 小写私有 `func (s *userService) CreateUser()`

### 代码规范

#### 1. 错误处理
```go
// ✅ 正确: 返回错误,不 panic
func GetUser(id int) (*User, error) {
    user, err := repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    return user, nil
}

// ❌ 错误: 使用 panic
func GetUser(id int) *User {
    user, err := repo.FindByID(id)
    if err != nil {
        panic(err)  // 绝对不允许
    }
    return user
}
```

#### 2. Context 传递
```go
// ✅ 正确: 总是传递 context
func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    return s.repo.Create(ctx, user)
}

// ❌ 错误: 缺少 context
func (s *userService) CreateUser(req *CreateUserRequest) error {
    return s.repo.Create(user)
}
```

#### 3. 依赖注入
```go
// ✅ 正确: 通过构造函数注入
type userService struct {
    repo  UserRepository
    cache Cache
}

func NewUserService(repo UserRepository, cache Cache) UserService {
    return &userService{repo: repo, cache: cache}
}

// ❌ 错误: 包级全局变量
var globalRepo UserRepository
```

#### 4. GORM 使用
```go
// ✅ 正确: 使用 context, 检查错误
func (r *userRepo) Create(ctx context.Context, user *model.User) error {
    result := r.db.WithContext(ctx).Create(user)
    if result.Error != nil {
        return fmt.Errorf("create user failed: %w", result.Error)
    }
    return nil
}

// 软删除模型必须包含
type User struct {
    ID        uint64
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

#### 5. 配置管理
```go
// ✅ 使用 Viper 统一管理
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Calcite  CalciteConfig
}

// 环境变量优先级最高
viper.SetEnvPrefix("DATAEASE")
viper.AutomaticEnv()
```

### SQL 引擎 (Avatica) 规范

```go
// ✅ 正确: 带缓存的查询
func (c *CalciteClient) ExecuteQuery(ctx context.Context, sql string) ([]map[string]interface{}, error) {
    // 1. 检查缓存
    if cached, err := c.cache.Get(ctx, cacheKey); err == nil {
        return cached, nil
    }
    
    // 2. 执行查询
    rows, err := c.db.QueryContext(ctx, sql)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()
    
    // 3. 缓存结果
    c.cache.Set(ctx, cacheKey, result, 5*time.Minute)
    return result, nil
}
```

### 并发安全
```go
// ✅ 正确: 使用 sync.Map 或 带锁的 map
type SafeCache struct {
    data sync.Map
}

// ❌ 错误: 裸 map 并发读写
var cache = make(map[string]interface{})  // 不安全!
```

---

## React 前端开发规范

### 项目结构
```
frontend/src/
├── api/                # API 接口
├── components/         # 公共组件
├── pages/             # 页面
├── hooks/             # 自定义 Hooks
├── store/             # Zustand 状态
├── utils/             # 工具函数
└── types/             # TypeScript 类型
```

### 命名规范
- **组件**: PascalCase `UserProfile.tsx`
- **Hooks**: camelCase with use `useUser.ts`
- **工具**: camelCase `formatDate.ts`
- **类型**: PascalCase `interface User {}`

### TypeScript 规范

#### 1. 严格模式
```typescript
// tsconfig.json
{
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true
  }
}
```

#### 2. 类型定义
```typescript
// ✅ 正确: 明确类型
interface ChartConfig {
  type: 'bar' | 'line' | 'pie';
  xAxis: FieldConfig;
  yAxis: FieldConfig;
  customAttr?: Record<string, any>;
}

// ❌ 错误: any 类型
const config: any = { ... };
```

#### 3. API 类型安全
```typescript
// ✅ 正确: 定义请求和响应类型
interface CreateUserRequest {
  name: string;
  email: string;
}

interface CreateUserResponse {
  id: string;
  name: string;
}

export const userAPI = {
  create: (data: CreateUserRequest) => 
    request.post<CreateUserResponse>('/user', data),
};
```

### React 组件规范

#### 1. 函数组件 + Hooks
```typescript
// ✅ 正确: 函数组件
interface ChartProps {
  config: ChartConfig;
  data: any[];
}

export const Chart: React.FC<ChartProps> = ({ config, data }) => {
  const [loading, setLoading] = useState(false);
  
  useEffect(() => {
    // 副作用
  }, [config]);
  
  return <div>...</div>;
};

// ❌ 错误: Class 组件 (不使用)
class Chart extends React.Component { }
```

#### 2. 自定义 Hooks
```typescript
// ✅ 正确: 提取业务逻辑到 Hooks
export const useDataset = (id?: string) => {
  const [dataset, setDataset] = useState<Dataset | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (id) fetchDataset(id);
  }, [id]);

  const refresh = useCallback(() => {
    if (id) fetchDataset(id);
  }, [id]);

  return { dataset, loading, error, refresh };
};
```

#### 3. 状态管理 (Zustand)
```typescript
// ✅ 正确: 类型安全的 store
interface ChartState {
  charts: Chart[];
  addChart: (chart: Chart) => void;
  deleteChart: (id: string) => void;
}

export const useChartStore = create<ChartState>()((set) => ({
  charts: [],
  addChart: (chart) => set((state) => ({ 
    charts: [...state.charts, chart] 
  })),
  deleteChart: (id) => set((state) => ({ 
    charts: state.charts.filter(c => c.id !== id) 
  })),
}));
```

### 性能优化

```typescript
// ✅ 正确: 使用 useMemo 和 useCallback
const ExpensiveComponent = ({ data, onUpdate }) => {
  const processedData = useMemo(() => 
    heavyComputation(data), [data]
  );
  
  const handleClick = useCallback(() => {
    onUpdate(processedData);
  }, [onUpdate, processedData]);
  
  return <div onClick={handleClick}>...</div>;
};

// ✅ 正确: 使用 React.memo 避免重渲染
export const Chart = React.memo<ChartProps>(({ config, data }) => {
  return <div>...</div>;
});
```

---

## 测试规范

### Go 后端测试

#### 1. 单元测试覆盖率 > 70%
```go
// user_service_test.go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    repo := &mockUserRepository{}
    service := NewUserService(repo, nil)
    
    // Act
    err := service.CreateUser(context.Background(), &CreateUserRequest{
        Name: "test",
    })
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 1, repo.createCallCount)
}
```

#### 2. 表驱动测试
```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid", "test@example.com", false},
        {"invalid", "invalid", true},
        {"empty", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("got %v, want error: %v", err, tt.wantErr)
            }
        })
    }
}
```

### React 前端测试

#### 1. Jest + Testing Library
```typescript
// Chart.test.tsx
import { render, screen } from '@testing-library/react';
import { Chart } from './Chart';

describe('Chart', () => {
  it('renders chart with correct data', () => {
    const data = [{ x: 1, y: 2 }];
    render(<Chart config={mockConfig} data={data} />);
    
    expect(screen.getByRole('img')).toBeInTheDocument();
  });
});
```

---

## 代码审查 Checklist

### 每次提交前必须自查:

- [ ] 功能与原 Java/Vue 版本一致
- [ ] 所有错误都有适当处理
- [ ] 添加了必要的注释(复杂逻辑)
- [ ] 通过了所有测试
- [ ] 代码格式化 (`go fmt`, `prettier`)
- [ ] 无 lint 错误
- [ ] 性能满足要求
- [ ] API 接口兼容
- [ ] 日志记录完善

---

## Git 提交规范

### Commit Message 格式
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type 类型
- `feat`: 新功能
- `fix`: Bug 修复
- `refactor`: 重构
- `test`: 测试
- `docs`: 文档
- `chore`: 构建/工具

### 示例
```
feat(dataset): implement dataset creation API

- Add CreateDataset handler
- Integrate with Calcite client
- Add unit tests

Refs: #123
```

---

## 性能监控要求

### 后端
- 每个 API 记录响应时间
- 慢查询日志 (> 1s)
- Panic recovery + 日志
- 内存/CPU 监控

### 前端
- Page Load Time
- API 响应时间
- 组件渲染时间
- 错误上报

---

## 文档要求

### 每个模块必须包含:
1. **README.md**: 模块说明
2. **代码注释**: 公共 API 和复杂逻辑
3. **API 文档**: Swagger/OpenAPI
4. **测试文档**: 测试场景和用例

---

## 重构验证流程

### 每个模块完成后:
1. **功能对比**: 与原 Java/Vue 版本逐一对比
2. **单元测试**: 覆盖率 > 70%
3. **集成测试**: 端到端流程测试
4. **性能测试**: 满足性能要求
5. **代码审查**: 通过 Checklist

---

## 禁止事项 ❌

1. ❌ **绝不**使用 `panic` (Go)
2. ❌ **绝不**使用 `any` (TypeScript, 除非确实需要)
3. ❌ **绝不**忽略错误
4. ❌ **绝不**直接修改数据库 Schema
5. ❌ **绝不**删除原有功能
6. ❌ **绝不**跳过测试
7. ❌ **绝不**在生产环境打印敏感信息

---

## 文档规范

### 1. 存放位置
- ✅ **必须**: 所有项目文档必须存放在 `docs/` 目录下
- ✅ **必须**: 根目录只保留 `README.md` 和配置文件

### 2. 必需文档
- `docs/DEVELOPMENT_GUIDE.md`: 开发指南
- `docs/QUALITY_CONTROL.md`: 质量管控
- `docs/API.md`: API 接口文档
- `docs/CHANGELOG.md`: 变更日志

---

## 优先级顺序

1. **功能正确性** > 性能 > 代码优雅
2. **与原版一致** > 优化改进
3. **可维护性** > 短期效率

---

## 记住

> 我们的目标是**完整复刻** DataEase,不是创建一个新产品。
> 每一行代码都要与原版对照验证!
