# DataEase 重构开发指南

## 目标

将 DataEase 从 **Java + Vue** 完整重构为 **Go + React**,保持100%功能对等。

---

## 开发流程

### 1. 模块开发流程

```
┌─────────────────┐
│  1. 分析原代码   │ 理解业务逻辑
└────────┬────────┘
         ↓
┌─────────────────┐
│  2. 设计 API    │ 保持接口兼容
└────────┬────────┘
         ↓
┌─────────────────┐
│  3. 实现后端    │ Go + GORM + Avatica
└────────┬────────┘
         ↓
┌─────────────────┐
│  4. 实现前端    │ React + TypeScript
└────────┬────────┘
         ↓
┌─────────────────┐
│  5. 编写测试    │ 单元测试 + 集成测试
└────────┬────────┘
         ↓
┌─────────────────┐
│  6. 功能对比    │ 与原版逐一验证
└────────┬────────┘
         ↓
┌─────────────────┐
│  7. 性能测试    │ 满足性能指标
└────────┬────────┘
         ↓
┌─────────────────┐
│  8. 代码审查    │ 通过 Checklist
└────────┬────────┘
         ↓
┌─────────────────┐
│  9. 集成合并    │ 合入主分支
└─────────────────┘
```

### 2. 每日工作流程

**上午**:
1. Pull 最新代码
2. 查看任务清单 (`task.md`)
3. 选择一个模块开始开发
4. 先阅读原 Java/Vue 代码,理解逻辑

**下午**:
1. 实现 Go 后端 或 React 前端
2. 编写单元测试
3. 本地验证功能
4. 提交代码 (符合 Commit 规范)

**晚上**:
1. 代码审查
2. 更新文档
3. 更新任务清单
4. 记录遇到的问题和解决方案

---

## 技术栈

### 后端 (Go)
- **框架**: Gin 1.9+
- **ORM**: GORM 2.0+
- **SQL 引擎**: Apache Calcite Avatica
- **缓存**: go-redis + groupcache
- **任务调度**: robfig/cron v3
- **配置**: viper
- **日志**: zap
- **测试**: testify

### 前端 (React)
- **核心**: React 18.2 + TypeScript 5.0
- **路由**: React Router v6
- **状态**: Zustand
- **UI**: Ant Design 5.x
- **图表**: AntV G2 + L7 + S2, ECharts
- **编辑器**: Monaco Editor
- **拖拽**: react-dnd, react-grid-layout
- **构建**: Vite 5.x
- **测试**: Jest + Testing Library

---

## 代码组织

### 后端目录结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go                # 应用入口
├── internal/
│   ├── handler/                   # HTTP 处理器
│   │   ├── datasource_handler.go
│   │   ├── dataset_handler.go
│   │   └── chart_handler.go
│   ├── service/                   # 业务逻辑
│   │   ├── datasource_service.go
│   │   ├── dataset_service.go
│   │   └── chart_service.go
│   ├── repository/                # 数据访问层
│   │   ├── datasource_repo.go
│   │   └── dataset_repo.go
│   ├── model/                     # 数据模型 (GORM)
│   │   ├── datasource.go
│   │   └── dataset.go
│   ├── dto/                       # 数据传输对象
│   ├── middleware/                # 中间件
│   │   ├── auth.go
│   │   ├── logger.go
│   │   └── cors.go
│   ├── engine/                    # SQL 引擎
│   │   └── calcite_client.go
│   └── util/                      # 工具函数
├── pkg/                           # 可导出包
│   ├── config/
│   ├── logger/
│   ├── database/
│   └── cache/
├── api/
│   └── v1/
│       └── router.go              # 路由定义
├── configs/
│   ├── app.yaml
│   └── app.production.yaml
├── scripts/
│   └── migrations/                # 数据库迁移
├── tests/                         # 集成测试
├── go.mod
└── go.sum
```

### 前端目录结构

```
frontend/
├── public/
│   └── index.html
├── src/
│   ├── api/                       # API 接口
│   │   ├── request.ts            # Axios 封装
│   │   ├── datasource.ts
│   │   ├── dataset.ts
│   │   └── chart.ts
│   ├── components/                # 公共组件
│   │   ├── Chart/
│   │   ├── DatasetSelector/
│   │   ├── FilterPanel/
│   │   └── DragCanvas/
│   ├── pages/                     # 页面组件
│   │   ├── dashboard/
│   │   ├── dataset/
│   │   ├── chart/
│   │   └── system/
│   ├── layouts/                   # 布局组件
│   │   ├── MainLayout.tsx
│   │   └── DashboardLayout.tsx
│   ├── hooks/                     # 自定义 Hooks
│   │   ├── useAuth.ts
│   │   ├── useDataset.ts
│   │   └── useChart.ts
│   ├── store/                     # Zustand 状态
│   │   ├── userStore.ts
│   │   ├── chartStore.ts
│   │   └── canvasStore.ts
│   ├── utils/                     # 工具函数
│   ├── router/                    # 路由配置
│   │   └── index.tsx
│   ├── types/                     # TypeScript 类型
│   │   ├── datasource.ts
│   │   ├── dataset.ts
│   │   └── chart.ts
│   ├── styles/                    # 全局样式
│   │   └── index.css
│   ├── App.tsx
│   └── main.tsx
├── tests/                         # 测试文件
├── vite.config.ts
├── tsconfig.json
├── package.json
└── .eslintrc.js
```

---

## 开发规范

### 1. 代码风格

**Go**:
- 使用 `gofmt` 格式化
- 使用 `golangci-lint` 检查
- 遵循 [Effective Go](https://go.dev/doc/effective_go)

**React/TypeScript**:
- 使用 `prettier` 格式化
- 使用 `eslint` 检查
- 遵循 Airbnb 风格指南

### 2. 注释规范

**Go**:
```go
// CreateDatasource 创建数据源
// 
// 参数:
//   - ctx: 上下文
//   - req: 创建请求
//
// 返回:
//   - error: 错误信息
func (s *datasourceService) CreateDatasource(ctx context.Context, req *CreateDatasourceRequest) error {
    // 1. 验证请求
    if err := req.Validate(); err != nil {
        return fmt.Errorf("invalid request: %w", err)
    }
    
    // 2. 创建数据源
    ds := &model.Datasource{...}
    
    // 3. 保存到数据库
    return s.repo.Create(ctx, ds)
}
```

**TypeScript**:
```typescript
/**
 * 创建数据源
 * @param data 创建请求数据
 * @returns 创建的数据源
 */
export const createDatasource = async (
  data: CreateDatasourceRequest
): Promise<Datasource> => {
  return request.post<Datasource>('/datasource', data);
};
```

### 3. 错误处理

**Go**:
```go
// ✅ 总是返回包装的错误
func GetUser(id int) (*User, error) {
    user, err := repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    if user == nil {
        return nil, ErrUserNotFound
    }
    return user, nil
}
```

**TypeScript**:
```typescript
// ✅ 使用 try-catch 并显示错误
const handleSubmit = async () => {
  try {
    setLoading(true);
    await createDatasource(formData);
    message.success('创建成功');
    navigate('/datasource');
  } catch (error) {
    message.error(error.message || '创建失败');
  } finally {
    setLoading(false);
  }
};
```

### 4. 日志规范

**级别**:
- `DEBUG`: 详细调试信息
- `INFO`: 常规信息
- `WARN`: 警告信息
- `ERROR`: 错误信息

**示例**:
```go
logger.Info("user created",
    zap.String("user_id", user.ID),
    zap.String("username", user.Name),
)

logger.Error("failed to create datasource",
    zap.Error(err),
    zap.String("type", dsType),
)
```

---

## 测试规范

### 1. 单元测试

**覆盖率要求**: 70%+

**Go**:
```go
// datasource_service_test.go
func TestDatasourceService_Create(t *testing.T) {
    // Arrange
    mockRepo := &MockDatasourceRepository{}
    service := NewDatasourceService(mockRepo, nil)
    
    req := &CreateDatasourceRequest{
        Name: "test",
        Type: "mysql",
    }
    
    // Act
    err := service.CreateDatasource(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 1, mockRepo.CreateCallCount)
}
```

**React**:
```typescript
// DatasetList.test.tsx
describe('DatasetList', () => {
  it('renders dataset list correctly', async () => {
    const mockDatasets = [
      { id: '1', name: 'Dataset 1' },
      { id: '2', name: 'Dataset 2' },
    ];
    
    jest.spyOn(datasetAPI, 'list').mockResolvedValue(mockDatasets);
    
    render(<DatasetList />);
    
    await waitFor(() => {
      expect(screen.getByText('Dataset 1')).toBeInTheDocument();
      expect(screen.getByText('Dataset 2')).toBeInTheDocument();
    });
  });
});
```

### 2. 集成测试

**后端**:
- 测试完整的 API 流程
- 使用真实数据库(测试环境)
- 测试 Avatica 连接

**前端**:
- 测试用户交互流程
- 测试 API 调用
- 测试路由跳转

### 3. 端到端测试 (E2E)

使用 Playwright 或 Cypress 测试完整用户流程。

---

## 性能要求

### 后端
- **启动时间**: < 2s
- **API 响应** (P95): < 500ms
- **内存占用**: < 200MB (空闲), < 1GB (负载)
- **并发**: > 10,000 req/s

### 前端
- **首屏加载**: < 2s
- **路由切换**: < 300ms
- **图表渲染**: < 1s (1000 数据点)

---

## Git 工作流

### 分支策略

```
main (生产)
  ↑
develop (开发)
  ↑
feature/datasource-api (功能分支)
```

### 提交规范

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type**:
- `feat`: 新功能
- `fix`: Bug 修复
- `refactor`: 重构
- `test`: 测试
- `docs`: 文档
- `style`: 格式
- `chore`: 构建/工具

**示例**:
```
feat(dataset): implement dataset query API

- Add QueryDataset handler
- Integrate with Calcite for SQL execution
- Add Redis caching for query results
- Add unit tests with 85% coverage

Closes #42
```

---

## 质量检查清单

### 每次提交前

- [ ] 代码格式化 (`go fmt`, `prettier`)
- [ ] Lint 检查通过 (`golangci-lint`, `eslint`)
- [ ] 单元测试通过
- [ ] 覆盖率 > 70%
- [ ] 与原版功能对比验证
- [ ] API 接口兼容性检查
- [ ] 性能测试通过
- [ ] 文档已更新

### 每周

- [ ] 集成测试通过
- [ ] 代码审查
- [ ] 技术债务清理
- [ ] 依赖更新

---

## 文档要求

### 必须文档
1. **README.md**: 每个模块的说明
2. **API.md**: API 接口文档 (Swagger)
3. **TESTING.md**: 测试用例文档
4. **CHANGELOG.md**: 变更日志

### 注释要求
- 所有公共 API 必须有注释
- 复杂逻辑必须有注释
- 算法必须注明时间/空间复杂度

---

## 监控和日志

### 后端监控
- **健康检查**: `/health`
- **指标**: `/metrics` (Prometheus 格式)
- **日志**: 结构化日志 (JSON)

### 前端监控
- **错误上报**: Sentry
- **性能监控**: Web Vitals
- **用户行为**: 埋点统计

---

## 安全规范

### 后端
- [ ] SQL 注入防护 (使用参数化查询)
- [ ] XSS 防护
- [ ] CSRF 防护
- [ ] 敏感数据加密
- [ ] JWT Token 验证
- [ ] Rate Limiting

### 前端
- [ ] 输入验证
- [ ] XSS 防护 (React 自动转义)
- [ ] HTTPS Only
- [ ] Token 存储安全
- [ ] 敏感信息不存本地

---

## 进度管理

### 任务清单 (`task.md`)
- 每天更新进度
- 标记完成状态: `[ ]` → `[/]` → `[x]`
- 记录遇到的问题

### 周报
每周五提交周报:
- 本周完成的模块
- 遇到的技术难点
- 下周计划

---

## 问题管理

### Issue 模板
```markdown
## 问题描述
简要描述问题

## 复现步骤
1. 
2. 
3. 

## 期望行为
应该如何

## 实际行为
实际如何

## 环境信息
- Go/Node 版本:
- 浏览器:
- 操作系统:

## 相关代码
```language
code here
```
```

---

## 参考资料

### Go
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [GORM 文档](https://gorm.io/)

### React
- [React 官方文档](https://react.dev/)
- [TypeScript 手册](https://www.typescriptlang.org/docs/)
- [Ant Design 组件库](https://ant.design/)

### 测试
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
- [React Testing Library](https://testing-library.com/react)

---

记住:我们的目标是**100%功能复刻**,每一行代码都要与原版对照验证!
