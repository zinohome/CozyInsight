#!/bin/bash

# CozyInsight 代码统计脚本

set -e

echo "========================================="
echo "  CozyInsight 代码统计"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 统计函数
count_lines() {
    local dir=$1
    local ext=$2
    local name=$3
    
    if [ -d "$dir" ]; then
        local count=$(find "$dir" -name "*.$ext" -type f -exec cat {} + | wc -l | tr -d ' ')
        local files=$(find "$dir" -name "*.$ext" -type f | wc -l | tr -d ' ')
        printf "${GREEN}%-15s${NC}: ${BLUE}%7s${NC} 行  (${YELLOW}%3s${NC} 文件)\n" "$name" "$count" "$files"
    fi
}

# 后端统计
echo "📦 后端代码 (Go)"
echo "-----------------------------------"
count_lines "backend/internal" "go" "业务代码"
count_lines "backend/pkg" "go" "公共包"
count_lines "backend/cmd" "go" "入口文件"
echo ""

# 前端统计
echo "🎨 前端代码 (React)"
echo "-----------------------------------"
count_lines "frontend/src" "ts" "TypeScript"
count_lines "frontend/src" "tsx" "TSX组件"
echo ""

# 测试统计
echo "🧪 测试代码"
echo "-----------------------------------"
count_lines "backend" "*_test.go" "Go测试"
count_lines "frontend/src" "test.ts" "React测试"
echo ""

# 文档统计
echo "📖 文档"
echo "-----------------------------------"
count_lines "docs" "md" "Markdown"
echo ""

# 配置文件
echo "⚙️  配置文件"
echo "-----------------------------------"
count_lines "." "yaml" "YAML"
count_lines "." "yml" "YML"
count_lines "deployments" "sql" "SQL"
echo ""

# 总计
echo "========================================="
echo "📊 总体统计"
echo "========================================="

total_go=$(find backend -name "*.go" -type f -exec cat {} + | wc -l | tr -d ' ')
total_ts=$(find frontend/src -name "*.ts" -o -name "*.tsx" -type f | xargs cat 2>/dev/null | wc -l | tr -d ' ')
total_md=$(find docs -name "*.md" -type f -exec cat {} + 2>/dev/null | wc -l | tr -d ' ')

total_files=$(find . -type f \( -name "*.go" -o -name "*.ts" -o -name "*.tsx" \) | wc -l | tr -d ' ')
total_lines=$((total_go + total_ts))

echo ""
printf "${GREEN}总代码行数${NC}: ${BLUE}%s${NC} 行\n" "$total_lines"
printf "${GREEN}总文件数${NC}:   ${BLUE}%s${NC} 个\n" "$total_files"
printf "${GREEN}文档字数${NC}:   ${BLUE}%s${NC} 行\n" "$total_md"
echo ""

# 模块统计
echo "========================================="
echo "🏗️  模块统计"
echo "========================================="
echo ""

models=$(find backend/internal/model -name "*.go" -type f | wc -l | tr -d ' ')
repos=$(find backend/internal/repository -name "*.go" -type f | wc -l | tr -d ' ')
services=$(find backend/internal/service -name "*.go" -type f | wc -l | tr -d ' ')
handlers=$(find backend/internal/handler -name "*.go" -type f | wc -l | tr -d ' ')

printf "${GREEN}Model${NC}:      ${BLUE}%3s${NC} 个\n" "$models"
printf "${GREEN}Repository${NC}: ${BLUE}%3s${NC} 个\n" "$repos"
printf "${GREEN}Service${NC}:    ${BLUE}%3s${NC} 个\n" "$services"
printf "${GREEN}Handler${NC}:    ${BLUE}%3s${NC} 个\n" "$handlers"
echo ""

# 前端模块
pages=$(find frontend/src/pages -name "*.tsx" -type f | wc -l | tr -d ' ')
components=$(find frontend/src/components -name "*.tsx" -type f | wc -l | tr -d ' ')
apis=$(find frontend/src/api -name "*.ts" -type f | wc -l | tr -d ' ')

printf "${GREEN}Pages${NC}:      ${BLUE}%3s${NC} 个\n" "$pages"
printf "${GREEN}Components${NC}: ${BLUE}%3s${NC} 个\n" "$components"
printf "${GREEN}API模块${NC}:   ${BLUE}%3s${NC} 个\n" "$apis"
echo ""

echo "========================================="
echo "✅ 统计完成!"
echo "========================================="
