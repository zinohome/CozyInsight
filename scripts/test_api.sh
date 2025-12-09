#!/bin/bash

# CozyInsight 自动化测试脚本
# 测试所有后端API端点

set -e

BASE_URL="http://localhost:8100/api/v1"
TOKEN=""

echo "🧪 CozyInsight 自动化API测试开始..."
echo "========================================="

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 测试计数
TOTAL=0
PASSED=0
FAILED=0

# 测试函数
test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_code=${5:-200}
    
    TOTAL=$((TOTAL + 1))
    echo -n "测试 $TOTAL: $name ... "
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" == "$expected_code" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $http_code)"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC} (Expected $expected_code, Got $http_code)"
        FAILED=$((FAILED + 1))
    fi
}

# 1. 健康检查
echo -e "\n📋 1. 基础健康检查"
echo "-----------------------------------"
test_api "健康检查" "GET" "/health" "" "200"

# 2. 认证测试
echo -e "\n🔐 2. 认证API测试"
echo "-----------------------------------"
test_api "用户注册" "POST" "/auth/register" \
    '{"username":"test_user","password":"test123","email":"test@example.com"}' "200"

test_api "用户登录" "POST" "/auth/login" \
    '{"username":"test_user","password":"test123"}' "200"

# 获取Token (简化版,实际需要解析响应)
# TOKEN="your_token_here"

# 3. 数据源API测试
echo -e "\n💾 3. 数据源API测试"
echo "-----------------------------------"
test_api "创建数据源" "POST" "/datasource" \
    '{"name":"Test MySQL","type":"mysql","config":"{\"host\":\"localhost\",\"port\":3306}"}' "200"

test_api "获取数据源列表" "GET" "/datasource" "" "200"

# 4. 数据集API测试
echo -e "\n📊 4. 数据集API测试"
echo "-----------------------------------"
test_api "创建数据集" "POST" "/dataset" \
    '{"name":"Test Dataset","type":"db"}' "200"

test_api "获取数据集列表" "GET" "/dataset" "" "200"

# 5. 图表API测试
echo -e "\n📈 5. 图表API测试"
echo "-----------------------------------"
test_api "创建图表" "POST" "/chart" \
    '{"name":"Test Chart","type":"bar"}' "200"

test_api "获取图表列表" "GET" "/chart" "" "200"

# 6. 仪表板API测试
echo -e "\n📱 6. 仪表板API测试"
echo "-----------------------------------"
test_api "创建仪表板" "POST" "/dashboard" \
    '{"name":"Test Dashboard","nodeType":"dashboard"}' "200"

test_api "获取仪表板列表" "GET" "/dashboard" "" "200"

# 7. 角色权限API测试
echo -e "\n🔑 7. 角色权限API测试"
echo "-----------------------------------"
test_api "创建角色" "POST" "/role" \
    '{"name":"Test Role","description":"Test"}' "200"

test_api "获取角色列表" "GET" "/role" "" "200"

test_api "获取权限列表" "GET" "/permission" "" "200"

# 8. 定时任务API测试
echo -e "\n⏰ 8. 定时任务API测试"
echo "-----------------------------------"
test_api "创建定时任务" "POST" "/schedule" \
    '{"name":"Test Task","type":"email_report","cronExpr":"0 0 * * *"}' "200"

test_api "获取任务列表" "GET" "/schedule" "" "200"

# 9. 系统设置API测试
echo -e "\n⚙️  9. 系统设置API测试"
echo "-----------------------------------"
test_api "设置系统配置" "POST" "/setting" \
    '{"type":"system","key":"test_key","value":"test_value"}' "200"

test_api "获取系统配置" "GET" "/setting/test_key" "" "200"

# 10. 操作日志API测试
echo -e "\n📝 10. 操作日志API测试"
echo "-----------------------------------"
test_api "获取操作日志" "GET" "/log" "" "200"

# 11. 数据导出API测试
echo -e "\n💾 11. 数据导出API测试"
echo "-----------------------------------"
test_api "导出图表数据" "GET" "/export/chart/test_id" "" "200"

# 12. 分享API测试
echo -e "\n🔗 12. 分享API测试"
echo "-----------------------------------"
test_api "创建分享链接" "POST" "/share" \
    '{"resourceType":"dashboard","resourceId":"test_id"}' "200"

test_api "获取分享列表" "GET" "/share" "" "200"

# 测试总结
echo -e "\n========================================"
echo "🎯 测试完成!"
echo "========================================"
echo "总计: $TOTAL"
echo -e "通过: ${GREEN}$PASSED${NC}"
echo -e "失败: ${RED}$FAILED${NC}"
echo "通过率: $(( PASSED * 100 / TOTAL ))%"
echo "========================================"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ 所有测试通过!${NC}"
    exit 0
else
    echo -e "${RED}❌ 有 $FAILED 个测试失败!${NC}"
    exit 1
fi
