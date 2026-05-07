#!/bin/bash
# CozyInsight Avatica Server 快速启动脚本

set -e

echo "======================================"
echo "   CozyInsight Avatica Server Setup   "
echo "======================================"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AVATICA_DIR="$SCRIPT_DIR"

# 检查依赖
command -v docker >/dev/null 2>&1 || { echo "错误: Docker 未安装。请先安装 Docker。"; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo "错误: Docker Compose 未安装。请先安装 Docker Compose。"; exit 1; }

echo "✓ Docker 和 Docker Compose 已安装"
echo ""

# 检查 JAR 文件是否存在
if [ ! -f "$AVATICA_DIR/avatica-server.jar" ]; then
    echo "未找到 avatica-server.jar，开始下载..."
    echo ""
    
    # Avatica 版本
    AVATICA_VERSION="1.24.0"
    DOWNLOAD_URL="https://repo1.maven.org/maven2/org/apache/calcite/avatica/avatica-server/$AVATICA_VERSION/avatica-server-$AVATICA_VERSION.jar"
    
    echo "下载地址: $DOWNLOAD_URL"
    wget -O "$AVATICA_DIR/avatica-server.jar" "$DOWNLOAD_URL" || {
        echo "下载失败！请手动下载 Avatica Server JAR 文件。"
        echo "下载地址: $DOWNLOAD_URL"
        echo "保存为: $AVATICA_DIR/avatica-server.jar"
        exit 1
    }
    
    echo "✓ Avatica Server JAR 下载完成"
    echo ""
fi

# 检查数据库驱动
echo "检查 JDBC 驱动..."
mkdir -p "$AVATICA_DIR/drivers"

if [ ! -f "$AVATICA_DIR/drivers/mysql-connector-j-8.2.0.jar" ]; then
    echo "下载 MySQL 驱动..."
    wget -P "$AVATICA_DIR/drivers" https://repo1.maven.org/maven2/com/mysql/mysql-connector-j/8.2.0/mysql-connector-j-8.2.0.jar
fi

if [ ! -f "$AVATICA_DIR/drivers/postgresql-42.7.1.jar" ]; then
    echo "下载 PostgreSQL 驱动..."
    wget -P "$AVATICA_DIR/drivers" https://repo1.maven.org/maven2/org/postgresql/postgresql/42.7.1/postgresql-42.7.1.jar
fi

echo "✓ JDBC 驱动准备完成"
echo ""

# 构建并启动 Docker 容器
echo "启动 Avatica Server..."
echo ""

cd "$AVATICA_DIR"
docker-compose up -d

echo ""
echo "等待服务启动..."
sleep 10

# 健康检查
MAX_RETRIES=10
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8765/ > /dev/null 2>&1; then
        echo ""
        echo "======================================"
        echo "   ✓ Avatica Server 启动成功！"
        echo "======================================"
        echo ""
        echo "服务地址: http://localhost:8765"
        echo ""
        echo "查看日志: docker-compose logs -f avatica"
        echo "停止服务: docker-compose down"
        echo ""
        exit 0
    fi
    
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "等待服务就绪... ($RETRY_COUNT/$MAX_RETRIES)"
    sleep 3
done

echo ""
echo "警告: 服务可能未正常启动"
echo "请检查日志: docker-compose logs avatica"
echo ""
exit 1
