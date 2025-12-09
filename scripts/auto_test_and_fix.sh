#!/bin/bash

# CozyInsight 自动化测试和修复脚本
# 自动运行测试、检测错误并尝试修复

set -e

echo "🤖 CozyInsight 自动化测试和修复"
echo "======================================="

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

TEST_ATTEMPTS=0
MAX_ATTEMPTS=3
COVERAGE_TARGET=85

# 函数: 运行测试
run_tests() {
    echo -e "\n${YELLOW}📋 运行测试...${NC}"
    cd backend
    
    # 运行测试并捕获输出
    if go test -cover ./... > test_output.txt 2>&1; then
        echo -e "${GREEN}✓ 测试通过${NC}"
        return 0
    else
        echo -e "${RED}✗ 测试失败${NC}"
        cat test_output.txt
        return 1
    fi
}

# 函数: 获取覆盖率
get_coverage() {
    echo -e "\n${YELLOW}📊 检查覆盖率...${NC}"
    cd backend
    
    go test -coverprofile=coverage.out ./... > /dev/null 2>&1
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    
    echo "当前覆盖率: ${coverage}%"
    echo $coverage
}

# 函数: 自动修复编译错误
auto_fix_compile_errors() {
    echo -e "\n${YELLOW}🔧 尝试自动修复编译错误...${NC}"
    
    # 检查是否有缺少的import
    if grep -q "undefined:" test_output.txt; then
        echo "检测到undefined错误,尝试添加imports..."
        # 这里可以添加自动修复逻辑
        go mod tidy
    fi
    
    # 格式化代码
    go fmt ./...
}

# 函数: 生成覆盖率报告
generate_coverage_report() {
    echo -e "\n${YELLOW}📈 生成覆盖率报告...${NC}"
    cd backend
    
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    
    echo -e "${GREEN}✓ 覆盖率报告已生成: backend/coverage.html${NC}"
}

# 函数: 识别未覆盖的代码
identify_uncovered_code() {
    echo -e "\n${YELLOW}🔍 识别未覆盖的代码...${NC}"
    cd backend
    
    go test -coverprofile=coverage.out ./... > /dev/null 2>&1
    
    echo "低覆盖率文件:"
    go tool cover -func=coverage.out | awk '{if ($3 != "100.0%") print $1, $3}' | head -10
}

# 函数: 自动生成测试用例建议
suggest_test_cases() {
    echo -e "\n${YELLOW}💡 生成测试用例建议...${NC}"
    
    # 查找没有测试的handler
    for handler in backend/internal/handler/*_handler.go; do
        if [ -f "$handler" ]; then
            handler_name=$(basename "$handler")
            test_file="${handler%%.go}_test.go"
            
            if [ ! -f "$test_file" ]; then
                echo "  - 缺少测试: $handler_name"
            fi
        fi
    done
}

# 主流程
main() {
    echo "开始自动化测试流程..."
    
    # 第一次运行测试
    TEST_ATTEMPTS=$((TEST_ATTEMPTS + 1))
    echo -e "\n${YELLOW}尝试 #${TEST_ATTEMPTS}${NC}"
    
    if run_tests; then
        echo -e "${GREEN}✓ 测试成功!${NC}"
    else
        echo -e "${YELLOW}! 测试失败,尝试自动修复...${NC}"
        auto_fix_compile_errors
        
        # 重试
        while [ $TEST_ATTEMPTS -lt $MAX_ATTEMPTS ]; do
            TEST_ATTEMPTS=$((TEST_ATTEMPTS + 1))
            echo -e "\n${YELLOW}尝试 #${TEST_ATTEMPTS}${NC}"
            
            if run_tests; then
                echo -e "${GREEN}✓ 修复成功!${NC}"
                break
            fi
        done
    fi
    
    # 检查覆盖率
    coverage=$(get_coverage)
    
    if (( $(echo "$coverage >= $COVERAGE_TARGET" | bc -l) )); then
        echo -e "${GREEN}✓ 覆盖率达标: ${coverage}% >= ${COVERAGE_TARGET}%${NC}"
    else
        echo -e "${YELLOW}! 覆盖率不足: ${coverage}% < ${COVERAGE_TARGET}%${NC}"
        identify_uncovered_code
        suggest_test_cases
    fi
    
    # 生成报告
    generate_coverage_report
    
    echo -e "\n======================================="
    echo -e "${GREEN}✅ 自动化测试完成!${NC}"
    echo "======================================="
}

# 执行主流程
main
