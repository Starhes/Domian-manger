#!/bin/bash

# Domain MAX 构建脚本

set -e

echo "=== Domain MAX 构建脚本 ==="
echo

# 检查必要工具
check_tools() {
    echo "🔍 检查构建工具..."
    
    if ! command -v go &> /dev/null; then
        echo "❌ Go 未安装，请先安装 Go 1.23+"
        exit 1
    fi
    
    if ! command -v node &> /dev/null; then
        echo "❌ Node.js 未安装，请先安装 Node.js 18+"
        exit 1
    fi
    
    if ! command -v npm &> /dev/null; then
        echo "❌ npm 未安装，请先安装 npm"
        exit 1
    fi
    
    echo "✅ 构建工具检查完成"
}

# 构建前端
build_web() {
    echo "🏗️  构建前端..."
    
    cd web
    
    # 安装依赖
    echo "📦 安装前端依赖..."
    npm ci
    
    # 构建
    echo "🔨 构建前端应用..."
    npm run build
    
    cd ..
    
    echo "✅ 前端构建完成"
}

# 构建后端
build_server() {
    echo "🏗️  构建后端..."
    
    # 下载Go依赖
    echo "📦 下载Go依赖..."
    go mod tidy
    
    # 构建
    echo "🔨 构建后端应用..."
    CGO_ENABLED=0 go build -ldflags="-w -s" -o domain-max ./cmd/server
    
    echo "✅ 后端构建完成"
}

# 清理构建产物
clean() {
    echo "🧹 清理构建产物..."
    
    rm -f domain-max
    rm -f domain-max.exe
    rm -rf web/dist
    rm -rf web/node_modules
    
    echo "✅ 清理完成"
}

# 主函数
main() {
    case "$1" in
        "clean")
            clean
            ;;
        "web")
            check_tools
            build_web
            ;;
        "server")
            check_tools
            build_server
            ;;
        "all"|"")
            check_tools
            build_web
            build_server
            echo
            echo "🎉 构建完成！"
            echo "📁 可执行文件：./domain-max"
            echo "🚀 运行命令：./domain-max"
            ;;
        *)
            echo "用法: $0 [clean|web|server|all]"
            echo
            echo "选项:"
            echo "  clean   - 清理构建产物"
            echo "  web     - 仅构建前端"
            echo "  server  - 仅构建后端"
            echo "  all     - 构建前端和后端（默认）"
            ;;
    esac
}

main "$@"