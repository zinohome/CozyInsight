#!/bin/bash

# CozyInsight 一键启动脚本

set -e

echo "🚀 CozyInsight 快速启动..."

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查依赖
echo -e "${YELLOW}1. 检查依赖...${NC}"

check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}错误: 未找到 $1, 请先安装${NC}"
        exit 1
    else
        echo -e "${GREEN}✓ $1 已安装${NC}"
    fi
}

check_command go
check_command node
check_command mysql
echo -e "${GREEN}✓ 所有依赖已满足${NC}\n"

# 启动MySQL (如果需要)
echo -e "${YELLOW}2. 检查MySQL...${NC}"
if ! mysql -u root -e "SELECT 1" &> /dev/null; then
    echo -e "${YELLOW}请确保MySQL正在运行并且可以连接${NC}"
fi

# 创建数据库
echo -e "${YELLOW}3. 准备数据库...${NC}"
mysql -u root -e "CREATE DATABASE IF NOT EXISTS dataease DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || true
echo -e "${GREEN}✓ 数据库已准备${NC}\n"

# 启动Avatica Server (后台)
echo -e "${YELLOW}4. 启动Avatica Server...${NC}"
cd backend/deployments/avatica
if docker-compose ps | grep -q "Up"; then
    echo -e "${GREEN}✓ Avatica Server 已在运行${NC}"
else
    docker-compose up -d
    echo -e "${GREEN}✓ Avatica Server 已启动${NC}"
fi
cd ../../..

# 等待Avatica就绪
echo "等待Avatica Server启动..."
for i in {1..30}; do
    if curl -s http://localhost:8765/ > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Avatica Server 已就绪${NC}\n"
        break
    fi
    sleep 1
    echo -n "."
done

# 启动后端 (后台)
echo -e "${YELLOW}5. 启动Go后端...${NC}"
cd backend
if [ ! -f "server" ]; then
    go build -o server cmd/server/main.go
fi
./server --config configs/app.yaml > ../logs/backend.log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > ../logs/backend.pid
echo -e "${GREEN}✓ 后端已启动 (PID: $BACKEND_PID)${NC}\n"
cd ..

# 等待后端就绪
echo "等待后端就绪..."
for i in {1..30}; do
    if curl -s http://localhost:8100/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ 后端已就绪${NC}\n"
        break
    fi
    sleep 1
    echo -n "."
done

# 启动前端
echo -e "${YELLOW}6. 启动React前端...${NC}"
cd frontend
npm run dev

# 清理函数
cleanup() {
    echo -e "\n${YELLOW}正在关闭服务...${NC}"
    if [ -f "../logs/backend.pid" ]; then
        kill $(cat ../logs/backend.pid) 2>/dev/null || true
        rm ../logs/backend.pid
    fi
    cd ../backend/deployments/avatica
    docker-compose down
    echo -e "${GREEN}✓ 服务已关闭${NC}"
}

trap cleanup EXIT INT TERM
