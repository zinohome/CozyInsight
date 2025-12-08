# Avatica Server 部署指南

## 概述

Apache Calcite Avatica Server 是 CozyInsight 项目的 SQL 引擎核心组件，负责：
- SQL 解析和优化
- 跨数据源查询
- 查询结果缓存

## 前置要求

- Docker & Docker Compose
- Java 21+ (如果本地运行)
- 至少 2GB 可用内存

## 快速开始

### 1. 下载 Avatica Server JAR

```bash
# 下载最新版本的 Avatica Server
wget https://repo1.maven.org/maven2/org/apache/calcite/avatica/avatica-server/1.24.0/avatica-server-1.24.0.jar \
  -O avatica-server.jar
```

### 2. 准备数据库驱动

将需要的 JDBC 驱动放入 `drivers/` 目录：

```bash
mkdir -p drivers && cd drivers

# MySQL
wget https://repo1.maven.org/maven2/com/mysql/mysql-connector-j/8.2.0/mysql-connector-j-8.2.0.jar

# PostgreSQL
wget https://repo1.maven.org/maven2/org/postgresql/postgresql/42.7.1/postgresql-42.7.1.jar

# ClickHouse (可选)
wget https://repo1.maven.org/maven2/com/clickhouse/clickhouse-jdbc/0.5.0/clickhouse-jdbc-0.5.0.jar
```

### 3. 启动服务

```bash
# 使用 Docker Compose 启动
docker-compose up -d

# 查看日志
docker-compose logs -f avatica

# 停止服务
docker-compose down
```

### 4. 验证部署

```bash
# 测试连接
curl http://localhost:8765/

# 预期输出: HTTP 200 OK
```

## 本地开发模式

如果不使用 Docker，可以直接运行：

```bash
java -Xmx2g -Xms1g \
  -jar avatica-server.jar \
  --spring.config.location=file:./config/application.yml
```

## 配置说明

配置文件：`config/application.yml`

主要配置项：
- `server.port`: Avatica 服务端口（默认 8765）
- `avatica.max-statements-per-connection`: 每个连接的最大语句数
- `datasource.hikari.*`: 连接池配置

## 监控与日志

### 查看日志

```bash
# Docker 模式
docker-compose logs -f avatica

# 本地模式
tail -f logs/avatica.log
```

### 健康检查

```bash
curl http://localhost:8765/actuator/health
```

## 性能优化

### JVM 参数调优

根据负载调整 `JAVA_OPTS`：

```bash
# 低负载（< 1000 并发）
JAVA_OPTS="-Xmx1g -Xms512m"

# 中负载（1000-5000 并发）
JAVA_OPTS="-Xmx2g -Xms1g"

# 高负载（> 5000 并发）
JAVA_OPTS="-Xmx4g -Xms2g -XX:+UseG1GC -XX:MaxGCPauseMillis=200"
```

### 连接池调优

在 `application.yml` 中调整：

```yaml
datasource:
  hikari:
    maximum-pool-size: 100  # 根据并发调整
    minimum-idle: 20
    connection-timeout: 30000
```

## 生产环境部署

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: avatica-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: avatica-server
  template:
    metadata:
      labels:
        app: avatica-server
    spec:
      containers:
      - name: avatica
        image: cozyinsight/avatica-server:latest
        ports:
        - containerPort: 8765
        env:
        - name: JAVA_OPTS
          value: "-Xmx2g -Xms1g"
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
---
apiVersion: v1
kind: Service
metadata:
  name: avatica-server
spec:
  type: LoadBalancer
  ports:
  - port: 8765
    targetPort: 8765
  selector:
    app: avatica-server
```

## 故障排查

### 常见问题

**1. 连接超时**
```bash
# 检查端口是否开放
netstat -an | grep 8765

# 检查防火墙
sudo ufw status
```

**2. 内存不足**
```bash
# 增加 JVM 堆内存
JAVA_OPTS="-Xmx4g -Xms2g"
```

**3. 数据库驱动未找到**
```bash
# 确保驱动 JAR 在 drivers/ 目录
ls -la drivers/
```

## 安全建议

1. **不要暴露到公网**：Avatica Server 应仅在内网访问
2. **启用认证**：生产环境建议配置 HTTP Basic Auth
3. **监控资源使用**：使用 Prometheus + Grafana 监控

## 参考资料

- [Apache Calcite Avatica 官方文档](https://calcite.apache.org/avatica/)
- [avatica-go 客户端](https://github.com/apache/calcite-avatica-go)
