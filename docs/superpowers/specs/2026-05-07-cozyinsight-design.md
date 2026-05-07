# CozyInsight 架构设计文档

**版本**: v1.0  
**日期**: 2026-05-07  
**目标**: 以 MIT 许可证完全开源、零 CE 限制、性能超越 DataEase v2.x 的企业级 BI 数据可视化平台

---

## 1. 项目定位

CozyInsight 是 DataEase v2.x 的功能级复刻与性能超越版本。

- **功能对等**: 100% 覆盖 DataEase v2.x 的 BI 核心能力
- **无 CE 限制**: 所有在 DataEase CE 中被限制的功能全部开放
- **性能超越**: 启动速度、查询响应、内存占用全面优于原版
- **许可证**: MIT（完全开源，无商业限制）
- **部署方式**: Docker Compose 一键部署，面向中小企业

---

## 2. 技术栈

### 后端

| 组件 | 选型 | 理由 |
|------|------|------|
| 语言/运行时 | Go 1.25 | 高性能、低内存、编译为单二进制 |
| Web 框架 | Gin | 轻量、高性能、生态成熟 |
| 数据库访问 | sqlx + squirrel | 比 GORM 快 30-50%，BI 场景需要精确 SQL 控制 |
| ORM/映射 | sqlx | 只做 struct 映射，零反射开销 |
| SQL 构建 | Masterminds/squirrel | 类型安全的 SQL 构建器 |
| 缓存 L1 | ristretto | 本地内存缓存（Dgraph 出品），热点查询走内存 |
| 缓存 L2 | go-redis/v9 | 分布式缓存、会话存储、任务队列后端 |
| 异步任务 | asynq | 基于 Redis 的可靠任务队列 |
| SQL 引擎 | Apache Calcite Avatica | 跨数据源 SQL 解析、路由、执行 |
| 任务调度 | robfig/cron/v3 | Cron 表达式定时任务 |
| 配置 | Viper | 支持 YAML/JSON/环境变量多源配置 |
| 日志 | Zap | 结构化高性能日志 |
| JWT | golang-jwt/jwt/v5 | 无状态认证 |
| 密码加密 | bcrypt (golang.org/x/crypto) | 行业标准 |
| Excel 导出 | excelize/v2 | 功能完整的 Go Excel 库 |
| 二维码/水印 | go-qrcode + image | 导出文件水印支持 |

### 前端

| 组件 | 选型 | 理由 |
|------|------|------|
| 框架 | React 19 | Concurrent Features、性能优化、Server Components 预留 |
| 语言 | TypeScript 5.9 | 严格模式，零 `any` |
| 构建工具 | Vite 7 | 比 Webpack 快 10 倍，HMR 极快 |
| UI 组件 | Ant Design 6 | 企业级组件库，与 DataEase 视觉一致 |
| 图表库 | ECharts 5 + @ant-design/charts | ECharts 大数据量性能更优，双库并用取优势 |
| 表格 | TanStack Table v8 | 虚拟滚动，大数据表格性能远超 Ant Design Table |
| 状态管理 | Zustand 5 | 轻量、类型安全、无样板代码 |
| 服务器状态 | TanStack Query v5 | 自动缓存、后台刷新、乐观更新、请求去重 |
| 路由 | React Router 7 | 嵌套路由、data API、代码分割 |
| 拖拽布局 | react-grid-layout | 仪表板网格拖拽，成熟稳定 |
| HTTP 客户端 | Axios 1.13 | 拦截器、取消请求、类型友好 |
| 日期处理 | dayjs | 轻量替代 moment.js |
| 工具库 | lodash-es | 按需引入 |

### 基础设施

| 组件 | 选型 | 用途 |
|------|------|------|
| 主数据库 | MySQL 8.0 | 元数据、用户、权限、配置存储 |
| 缓存/队列 | Redis 7 | 分布式缓存 + asynq 任务队列 |
| 搜索引擎 | Meilisearch | 仪表板/图表/数据源快速全文搜索 |
| 文件存储 | 本地 / MinIO | 导出文件、上传文件、模板存储 |
| SQL 引擎服务 | Apache Calcite Avatica (Docker) | 跨数据源 SQL 解析与执行 |
| 消息通知 | SMTP + Webhook | 定时报表邮件推送、告警通知 |

---

## 3. 架构设计

### 3.1 整体架构

单体服务架构，一个 Go 进程承载全部后端功能，前端为独立 SPA。

```
┌─────────────────┐         ┌─────────────────────────────┐
│   React SPA     │◄───────▶│   Go Backend (Port 8100)    │
│  (Vite Dev/     │   HTTP  │  ┌───────────────────────┐  │
│   Nginx Prod)   │         │  │   Gin Router          │  │
└─────────────────┘         │  ├───────────────────────┤  │
                            │  │   Middleware          │  │
                            │  │   (Auth/Permission/   │  │
                            │  │    CORS/Log/Recovery) │  │
                            │  ├───────────────────────┤  │
                            │  │   Handler Layer       │  │
                            │  │   (HTTP 请求/响应)     │  │
                            │  ├───────────────────────┤  │
                            │  │   Service Layer       │  │
                            │  │   (业务逻辑)           │  │
                            │  ├───────────────────────┤  │
                            │  │   Repository Layer    │  │
                            │  │   (sqlx + squirrel)   │  │
                            │  ├───────────────────────┤  │
                            │  │   Engine Layer        │  │
                            │  │   (Calcite Client +   │  │
                            │  │    Data Connectors)   │  │
                            │  ├───────────────────────┤  │
                            │  │   Background Workers  │  │
                            │  │   (asynq + cron)      │  │
                            │  └───────────────────────┘  │
                            │                             │
                            └─────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
                 MySQL 8.0        Redis 7           Avatica
                 (元数据)       (L2缓存+队列)       (SQL引擎)
                    │                 │
                    └─────────────────┘
                          Meilisearch
                         (全文搜索)
```

### 3.2 分层职责

**Handler 层** (`internal/handler/`)
- HTTP 请求入口，JSON ↔ DTO 绑定
- 调用 Service 层，返回标准化响应 `{code, data, error}`
- 不做业务逻辑，只负责输入校验和输出格式化

**Service 层** (`internal/service/`)
- 纯业务逻辑，无状态
- 通过接口依赖 Repository 和 Engine
- 构造函数注入（`NewXxxService(...)`），禁止包级全局变量
- 每个方法必须接受并传递 `context.Context`

**Repository 层** (`internal/repository/`)
- sqlx + squirrel 执行数据库操作
- 每个 Repository 封装单一模型的 CRUD
- 所有查询必须传入 `context.Context`
- 返回错误时必须用 `fmt.Errorf("...: %w", err)` 包装

**Engine 层** (`internal/engine/`)
- `calcite_client.go`: Apache Calcite Avatica 客户端，跨数据源 SQL 执行
- `datasource_connector.go`: 各数据源原生连接器（MySQL、PostgreSQL、ClickHouse 等）
- 查询结果入双层缓存（ristretto L1 → Redis L2）

**Background Workers**
- `asynq`: 大数据导出、报表生成等耗时任务异步处理
- `robfig/cron`: 定时任务调度（定时报表、数据同步）

### 3.3 缓存策略

双层缓存设计，这是性能超越 DataEase 的核心手段：

```
查询请求
    │
    ▼
┌──────────────┐  命中?  ┌──────────────┐  命中?  ┌──────────────┐
│ ristretto L1 │────────▶│  Redis L2    │────────▶│  数据源执行  │
│  (本地内存)   │   否    │  (分布式)     │   否    │  (Calcite)  │
│  TTL: 5min   │         │  TTL: 30min   │         │             │
└──────────────┘         └──────────────┘         └──────────────┘
    │                        │                        │
    ▼                        ▼                        ▼
 返回数据                 回填 L1                  回填 L1+L2
```

缓存键设计：`cache:{datasource_id}:{dataset_id}:{sql_hash}:{user_id?}`
- 含行级权限的查询加入 `user_id` 隔离
- SQL 变更、数据源配置变更时主动失效缓存

### 3.4 数据库 Schema 策略

**策略**: 自主设计，不兼容 DataEase 原生 schema，但提供迁移脚本。

理由：
- DataEase 的 schema 是围绕 Java/Spring 设计的，有冗余字段、历史包袱
- 新 schema 针对 Go + sqlx 优化，更简洁、索引更合理
- 提供 `scripts/migrate-from-dataease.sql` 供用户从 DataEase 迁移数据

Schema 设计原则：
- 每张表有 `id`（雪花 ID）、`created_at`、`updated_at`、`deleted_at`（软删除）
- 外键关系用应用层保证（不建物理外键），便于分片和性能优化
- 频繁查询的字段建立联合索引
- JSON 字段存储变长配置（如图表样式、数据源配置），减少表数量

---

## 4. 功能模块

### 4.1 核心 BI 功能（对标 DataEase v2.x）

| 模块 | 功能 | CE 限制处理 |
|------|------|-------------|
| **数据源** | MySQL、PostgreSQL、ClickHouse、Oracle、SQL Server、Doris、StarRocks、Excel、CSV、API | 无数量限制 |
| **数据集** | 数据库表数据集、SQL 数据集、Excel 数据集、API 数据集、关联数据集 | 无数量限制 |
| **图表** | 柱状图、折线图、饼图、散点图、雷达图、热力图、漏斗图、仪表盘、词云、面积图、表格、透视表 | 无类型限制 |
| **仪表板** | 网格拖拽布局、联动、下钻、公共筛选器、跳转、预览/编辑模式 | 无数量限制 |
| **数据大屏** | 自由布局（非网格）、大屏模板、自适应分辨率 | 无数量限制 |
| **模板中心** | 仪表板模板、大屏模板、数据集模板、一键应用 | 完全开放 |

### 4.2 企业功能（零 CE 限制）

| 功能 | 说明 | DataEase CE 限制 |
|------|------|------------------|
| **用户管理** | 用户 CRUD、组织架构（部门/岗位） | CE 限制用户数 |
| **RBAC 权限** | 角色、菜单权限、资源权限 | CE 限制角色数 |
| **行级权限** | 基于用户属性动态注入 SQL WHERE | CE 不支持 |
| **数据脱敏** | 字段级脱敏规则（手机号、身份证等） | CE 不支持 |
| **分享功能** | 公开分享（密码/有效期）、内部分享 | CE 限制分享数 |
| **数据导出** | Excel/CSV/PDF/图片导出 | CE 限制导出次数 |
| **导出水印** | 全局水印、导出文件水印 | CE 不支持 |
| **定时任务** | 定时报表生成 + 邮件推送 | CE 限制任务数 |
| **操作日志** | 全量审计、登录日志、导出日志 | CE 限制保留期 |
| **嵌入式分析** | iframe 嵌入、JS SDK | CE 不支持 |
| **SSO/LDAP** | OAuth2、SAML、LDAP | CE 不支持 |
| **AI SQL 助手** | 自然语言转 SQL | CE 不支持 |
| **地图可视化** | L7 地理引擎、区域/热力/路径图 | CE 限制地图类型 |
| **数据血缘** | 数据源 → 数据集 → 图表 → 仪表板 血缘追踪 | CE 不支持 |

---

## 5. 部署架构

### 5.1 Docker Compose 一键部署

```yaml
# docker-compose.yml
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: cozyinsight
      MYSQL_DATABASE: cozyinsight
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

  meilisearch:
    image: getmeili/meilisearch:v1.10
    volumes:
      - meilisearch_data:/meili_data

  avatica:
    image: cozyinsight/avatica:latest
    # Apache Calcite Avatica Server

  backend:
    image: cozyinsight/backend:latest
    ports:
      - "8100:8100"
    environment:
      - DATABASE_DSN=root:cozyinsight@tcp(mysql:3306)/cozyinsight
      - REDIS_ADDR=redis:6379
      - AVATICA_URL=http://avatica:8765/
    depends_on:
      - mysql
      - redis
      - avatica

  frontend:
    image: cozyinsight/frontend:latest
    ports:
      - "80:80"
    depends_on:
      - backend
```

### 5.2 单二进制部署（可选）

Go 后端编译为单二进制文件，前端静态资源通过 `embed` 嵌入 Go 二进制中。

```bash
# 一键启动，无需 Docker
cozyinsight-server --config app.yaml
# 访问 http://localhost:8100
```

适用场景：边缘部署、快速演示、开发调试。

---

## 6. UI/UX 策略

**策略**: 完全复刻 DataEase v2.x 的 UI/UX。

### 6.1 视觉还原

- **布局**: 侧边导航 + 顶部工具栏 + 主内容区，与 DataEase 一致
- **配色**: 使用 Ant Design 6 的默认主题（与 DataEase 的 Element UI 蓝色主题接近），必要时微调至一致
- **图标**: Ant Design Icons（与 DataEase 的图标集语义对应）
- **字体**: 系统默认字体栈（PingFang SC、Microsoft YaHei）
- **组件交互**: 完全对标 DataEase 的表单、表格、弹窗、抽屉、步骤条等交互模式

### 6.2 关键页面还原

| 页面 | DataEase 特征 | 实现方式 |
|------|---------------|----------|
| **工作台** | 左侧树形导航、右侧快捷入口 | React Router 布局 + Ant Design Menu |
| **数据源管理** | 卡片列表 + 连接测试弹窗 | Ant Design Card + Form |
| **数据集编辑器** | 左侧面板（字段/数据）、右侧表格 | 三栏布局 + TanStack Table |
| **图表编辑器** | 左侧字段区、中间画布、右侧配置区 | react-grid-layout + ECharts |
| **仪表板编辑器** | 顶部工具栏、中间画布、右侧属性区 | react-grid-layout + 拖拽 |
| **数据大屏** | 自由布局、组件可重叠 | 绝对定位 + 缩放适配 |

### 6.3 交互细节还原

- 图表字段拖拽到维度/指标区
- 仪表板组件拖拽调整大小和位置
- 筛选器联动（选择后自动刷新关联图表）
- 下钻路径面包屑
- 图表数据提示（Tooltip）格式
- 导出进度条和结果下载

---

## 7. 性能目标

| 指标 | DataEase v2.x (Java) | CozyInsight 目标 | 达成手段 |
|------|----------------------|------------------|----------|
| **启动时间** | ~60s | **< 3s** | Go 编译为单二进制、无 JVM 预热 |
| **内存占用** | ~2GB | **< 200MB** | Go 原生内存管理、双层缓存 |
| **API P95 响应** | ~800ms | **< 200ms** | sqlx + ristretto L1 + Redis L2 |
| **并发 QPS** | ~500 | **> 2000** | Goroutine + 连接池 + 缓存 |
| **首屏加载** | ~5s | **< 2s** | Vite 代码分割 + 懒加载 + TanStack Query |
| **图表渲染** | ~2s (1000 点) | **< 500ms** | ECharts 大数据优化 + Web Worker |
| **大数据导出** | 同步阻塞 | **异步队列** | asynq + 流式写入 Excel |

---

## 8. 分阶段交付计划

### 阶段 1：核心骨架 + 数据链路（4-6 周）

**目标**: 数据源 → 数据集 → 图表 → 仪表板 最小闭环

- 基础设施（Docker Compose、数据库迁移、项目骨架）
- JWT 认证（登录/注册）
- 数据源管理（MySQL/PostgreSQL/ClickHouse 连接 + 测试）
- 数据集管理（数据库表数据集、SQL 数据集、Excel 上传、字段同步、数据预览）
- 图表（4 种基础图表 + 拖拽式字段配置）
- 仪表板（空画布 + 拖拽布局 + 保存/加载）
- 前端页面（登录、工作台、数据源、数据集、图表编辑器、仪表板编辑器）

### 阶段 2：图表丰富 + 交互增强（3-4 周）

**目标**: 图表类型齐全，仪表板具备完整交互

- 图表扩展至 12 种
- 图表联动、下钻、公共筛选器
- 双层缓存上线（ristretto + Redis）
- TanStack Query 集成、虚拟滚动表格

### 阶段 3：企业功能 + 权限体系（3-4 周）

**目标**: 团队协作能力，零 CE 限制

- 用户/角色/组织架构
- RBAC + 行级权限
- 分享、导出（带水印）、定时任务、操作日志

### 阶段 4：高级特性 + 性能极致（4-6 周）

**目标**: 全面超越 DataEase CE

- 数据大屏、嵌入式分析、AI SQL 助手
- 地图可视化、模板中心、数据血缘
- OLAP 直连、全局搜索、SSO/LDAP

---

## 9. 关键设计决策

### 9.1 为什么不用 GORM？

BI 场景需要精确控制 SQL（子查询、CTE、窗口函数、动态 WHERE）。GORM 的抽象层会隐藏这些细节，且性能不如 sqlx + squirrel。

### 9.2 为什么用双层缓存？

DataEase 使用单层 Redis 缓存。ristretto 作为 L1 缓存可以将热点查询响应时间从 ~5ms（Redis 网络往返）降到 ~500ns（本地内存），提升 10 倍。

### 9.3 为什么用 asynq？

大数据导出（10 万行 Excel）如果同步处理会阻塞 HTTP 连接数分钟。asynq 将导出任务放入 Redis 队列，后台 Worker 异步处理，前端通过 WebSocket 或轮询获取进度。

### 9.4 为什么前端用 ECharts + AntV 双库？

- ECharts 大数据量渲染性能更优（Canvas 渲染、数据下采样）
- Ant Design Charts（基于 G2）在统计图表、交互细节上更丰富
- 双库并用，取各自优势，不限制图表类型

---

## 10. 风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| **Calcite Avatica 集成复杂度** | 高 | 参考 Archive 中的旧实现，优先接入 MySQL/PostgreSQL |
| **前端 UI 完全复刻难度** | 中 | 先复刻核心布局，交互细节逐步逼近 |
| **大数据量查询性能** | 中 | 流式查询 + 异步导出 + 查询超时熔断 |
| **阶段 1 时间估算偏差** | 中 | 每周检查进度，必要时调整范围 |

---

## 11. 非功能需求

- **代码质量**: Go 单元测试覆盖率 ≥ 70%，前端组件测试覆盖公共组件
- **文档**: 每个模块必须有 README + API 文档（Swagger/OpenAPI）
- **安全**: SQL 注入防护（参数化查询）、XSS 防护（前端自动转义）、CSRF Token、JWT 过期刷新、敏感数据加密
- **可观测性**: Zap 结构化日志、Prometheus 指标端点、健康检查 `/health`
- **国际化**: 预留 i18n 架构（react-i18next），第一阶段只支持中文

---

*本文档经用户确认后，下一步将使用 writing-plans 技能编写阶段 1 的详细实现计划。*
