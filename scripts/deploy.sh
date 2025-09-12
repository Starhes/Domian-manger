#!/bin/bash

# Domain MAX 部署脚本
set -e

echo "🚀 开始部署 Domain MAX..."

# 检查必要的环境变量
check_env_vars() {
    echo "📋 检查环境变量..."
    
    required_vars=("DB_PASSWORD" "JWT_SECRET" "ENCRYPTION_KEY")
    missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -ne 0 ]; then
        echo "❌ 缺少必要的环境变量:"
        for var in "${missing_vars[@]}"; do
            echo "   - $var"
        done
        echo ""
        echo "请设置这些环境变量或创建 .env 文件"
        exit 1
    fi
    
    echo "✅ 环境变量检查通过"
}

# 构建应用
build_app() {
    echo "🔨 构建应用..."
    
    # 构建前端
    echo "构建前端..."
    cd web
    npm install
    npm run build
    cd ..
    
    # 构建后端
    echo "构建后端..."
    go build -o domain-max ./cmd/server
    
    echo "✅ 应用构建完成"
}

# 使用 Docker 部署
deploy_with_docker() {
    echo "🐳 使用 Docker 部署..."
    
    # 停止现有容器
    docker-compose -f deployments/docker-compose.yml down || true
    
    # 构建并启动服务
    docker-compose -f deployments/docker-compose.yml up -d --build
    
    echo "✅ Docker 部署完成"
}

# 直接部署
deploy_direct() {
    echo "🚀 直接部署..."
    
    # 停止现有进程
    pkill -f domain-max || true
    
    # 启动应用
    nohup ./domain-max > app.log 2>&1 &
    
    echo "✅ 直接部署完成"
    echo "📝 日志文件: app.log"
}

# 检查部署状态
check_deployment() {
    echo "🔍 检查部署状态..."
    
    # 等待服务启动
    sleep 5
    
    # 检查健康状态
    if curl -f http://localhost:8080/api/health > /dev/null 2>&1; then
        echo "✅ 应用运行正常"
        echo "🌐 访问地址: http://localhost:8080"
    else
        echo "❌ 应用启动失败"
        echo "📝 查看日志: tail -f app.log"
        exit 1
    fi
}

# 主函数
main() {
    echo "Domain MAX 部署脚本"
    echo "===================="
    
    # 检查参数
    if [ "$1" = "docker" ]; then
        check_env_vars
        build_app
        deploy_with_docker
    elif [ "$1" = "direct" ]; then
        build_app
        deploy_direct
        check_deployment
    else
        echo "用法: $0 [docker|direct]"
        echo ""
        echo "  docker  - 使用 Docker 部署"
        echo "  direct  - 直接部署（需要手动设置环境变量）"
        exit 1
    fi
}

# 运行主函数
main "$@"
