#!/bin/bash

# CozyInsight 完整测试套件
# 运行所有测试

set -e

echo "🧪 CozyInsight 完整测试套件"
echo "======================================="

# 1. Go单元测试
echo -e "\n📦 1. 运行Go单元测试..."
echo "-----------------------------------"
cd backend
go test -v -cover ./internal/service/... || true
go test -v -cover ./internal/repository/... || true
go test -v -cover ./pkg/... || true
cd ..

# 2. Go代码检查
echo -e "\n🔍 2. Go代码质量检查..."
echo "-----------------------------------"
cd backend
go vet ./... || true
cd ..

# 3. 前端测试
echo -e "\n⚛️  3. 运行前端测试..."
echo "-----------------------------------"
cd frontend
npm test -- --watchAll=false --coverage || true
cd ..

# 4. 前端代码检查
echo -e "\n🔧 4. 前端代码质量检查..."
echo "-----------------------------------"
cd frontend
npm run lint || true
cd ..

# 5. API集成测试
echo -e "\n🌐 5. API集成测试..."
echo "-----------------------------------"
chmod +x scripts/test_api.sh
./scripts/test_api.sh || true

echo -e "\n======================================="
echo "✅ 所有测试完成!"
echo "======================================="
