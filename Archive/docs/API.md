# CozyInsight API 文档

> **版本**: v1.0.0  
> **基础URL**: `http://localhost:8100/api/v1`

---

## 认证

所有需要认证的接口都需要在Header中携带JWT Token:

```http
Authorization: Bearer <your_jwt_token>
```

---

## 1. 认证授权

### 1.1 用户注册

```http
POST /api/v1/auth/register
```

**请求体**:
```json
{
  "username": "admin",
  "password": "admin123",
  "email": "admin@example.com",
  "nickName": "Administrator"
}
```

**响应**:
```json
{
  "id": "user-id",
  "username": "admin",
  "email": "admin@example.com",
  "token": "eyJhbGc..."
}
```

### 1.2 用户登录

```http
POST /api/v1/auth/login
```

**请求体**:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

---

## 2. 数据源管理

### 2.1 创建数据源

```http
POST /api/v1/datasource
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "name": "MySQL测试库",
  "type": "mysql",
  "config": {
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "root123",
    "database": "test_db"
  }
}
```

### 2.2 测试连接

```http
POST /api/v1/datasource/test
Authorization: Bearer <token>
```

### 2.3 获取数据库列表

```http
GET /api/v1/datasource/:id/databases
Authorization: Bearer <token>
```

### 2.4 获取表列表

```http
GET /api/v1/datasource/:id/tables?database=test_db
Authorization: Bearer <token>
```

---

## 3. 数据集管理

### 3.1 创建数据集分组

```http
POST /api/v1/dataset/group
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "name": "分组名称",
  "pid": "0",
  "type": "folder"
}
```

### 3.2 创建数据集

```http
POST /api/v1/dataset/table
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "name": "数据集名称",
  "datasetGroupId": "group-id",
  "datasourceId": "datasource-id",
  "dbName": "test_db",
  "tableName": "users",
  "type": "db",
  "mode": 0
}
```

### 3.3 数据预览

```http
GET /api/v1/dataset/:id/preview?limit=100
Authorization: Bearer <token>
```

**响应**:
```json
{
  "fields": [
    {
      "name": "id",
      "type": "BIGINT",
      "deType": 2
    }
  ],
  "data": [
    {"id": 1, "name": "张三"}
  ]
}
```

### 3.4 字段同步

```http
POST /api/v1/dataset/:id/sync-fields
Authorization: Bearer <token>
```

### 3.5 导出数据

```http
GET /api/v1/dataset/:id/export?format=excel
Authorization: Bearer <token>
```

参数:
- `format`: `excel` 或 `csv`

---

## 4. 图表管理

### 4.1 创建图表

```http
POST /api/v1/chart
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "name": "销售趋势图",
  "sceneId": "scene-id",
  "tableId": "table-id",
  "type": "line",
  "xAxis": {
    "fields": [{"fieldId": "date"}]
  },
  "yAxis": {
    "fields": [{"fieldId": "amount", "aggregate": "SUM"}]
  }
}
```

### 4.2 获取图表数据

```http
GET /api/v1/chart/:id/data
Authorization: Bearer <token>
```

**响应**:
```json
[
  {"date": "2023-01", "amount": 10000},
  {"date": "2023-02", "amount": 15000}
]
```

### 4.3 导出图表数据

```http
GET /api/v1/chart/:id/export?format=excel
Authorization: Bearer <token>
```

---

## 5. 仪表板管理

### 5.1 创建仪表板

```http
POST /api/v1/dashboard
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "name": "销售仪表板",
  "pid": "0",
  "nodeType": "dashboard",
  "type": "dashboard"
}
```

### 5.2 保存组件

```http
POST /api/v1/dashboard/:id/components
Authorization: Bearer <token>
```

**请求体**:
```json
[
  {
    "chartId": "chart-id",
    "type": "chart",
    "x": 0,
    "y": 0,
    "w": 6,
    "h": 4
  }
]
```

### 5.3 获取组件

```http
GET /api/v1/dashboard/:id/components
Authorization: Bearer <token>
```

### 5.4 发布仪表板

```http
POST /api/v1/dashboard/:id/publish
Authorization: Bearer <token>
```

### 5.5 获取已发布仪表板(公开)

```http
GET /api/v1/dashboard/public/:id
```

---

## 6. 权限管理

### 6.1 角色管理

#### 创建角色

```http
POST /api/v1/role
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "name": "数据分析师",
  "description": "可以查看和创建图表"
}
```

#### 分配角色给用户

```http
POST /api/v1/role/:roleId/assign
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "userId": "user-id"
}
```

### 6.2 权限管理

#### 授予资源权限

```http
POST /api/v1/permission/resource/grant
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "resourceType": "dataset",
  "resourceId": "dataset-id",
  "targetType": "user",
  "targetId": "user-id",
  "permission": "read"
}
```

权限级别:
- `read` - 只读
- `write` - 读写
- `manage` - 管理(含删除)

#### 检查权限

```http
GET /api/v1/permission/check?resourceType=dataset&resourceId=xxx&action=read
Authorization: Bearer <token>
```

---

## 7. 分享管理

### 7.1 创建分享

```http
POST /api/v1/share
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "resourceType": "dashboard",
  "resourceId": "dashboard-id",
  "password": "1234",
  "expireTime": 1735660800000
}
```

**响应**:
```json
{
  "id": "share-id",
  "token": "abc123",
  "resourceType": "dashboard",
  "resourceId": "dashboard-id"
}
```

分享链接: `http://yourdomain/share/abc123`

### 7.2 验证分享(公开访问)

```http
GET /api/v1/share/validate/:token?password=1234
```

---

## 8. 定时任务

### 8.1 创建任务

```http
POST /api/v1/schedule
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "name": "每日报表",
  "type": "email_report",
  "cronExpr": "0 0 9 * * *",
  "enabled": true,
  "config": {
    "emails": ["admin@example.com"]
  }
}
```

### 8.2 启用任务

```http
POST /api/v1/schedule/:id/enable
Authorization: Bearer <token>
```

### 8.3 立即执行

```http
POST /api/v1/schedule/:id/execute
Authorization: Bearer <token>
```

---

## 通用响应格式

### 成功响应

```json
{
  "code": 200,
  "data": { ... }
}
```

### 错误响应

```json
{
  "code": 400,
  "error": "错误信息"
}
```

### HTTP状态码

- `200` - 成功
- `400` - 请求参数错误
- `401` - 未认证
- `403` - 无权限
- `404` - 资源不存在  
- `500` - 服务器错误

---

## Cron表达式示例

```
0 0 * * * *     每小时
0 0 9 * * *     每天9点
0 0 9 * * 1     每周一9点
0 0 1 1 * *     每月1号0点
```

格式: `秒 分 时 日 月 周`

---

## 数据类型映射

### deType字段类型

- `0` - 文本
- `1` - 时间
- `2` - 数值
- `3` - 地理位置
- `4` - 其他

### 图表类型

- `bar` - 横向柱状图
- `column` - 纵向柱状图
- `line` - 折线图
- `pie` - 饼图
- `scatter` - 散点图
- `radar` - 雷达图
- `heatmap` - 热力图
- `area` - 面积图
- `funnel` - 漏斗图
- `gauge` - 仪表盘
- `wordcloud` - 词云
- `table` - 表格

---

**完整API文档 - CozyInsight v1.0**
