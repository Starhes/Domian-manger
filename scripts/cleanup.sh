#!/bin/bash

# Domain MAX 清理脚本
# 用于清理开发和测试过程中生成的临时文件

echo "=== Domain MAX 清理脚本 ==="
echo

# 定义需要清理的文件和目录
CLEANUP_ITEMS=(
    # 构建产物
    "domain-max"
    "domain-max.exe"
    "web/dist"
    "web/node_modules"
    
    # 测试文件
    "*.test"
    "coverage.out"
    "profile.out"
    "test_output.log"
    "debug.log"
    
    # 临时目录
    "tmp/"
    "temp/"
    "test_data/"
    ".test_cache/"
    
    # IDE文件
    ".vscode/"
    ".idea/"
    "*.swp"
    "*.swo"
    "*~"
    
    # 系统文件
    ".DS_Store"
    "Thumbs.db"
)

# 定义保留的重要文件
KEEP_FILES=(
    ".env"
    "configs/env.example"
    "README.md"
    "LICENSE"
    "go.mod"
    "go.sum"
    "web/package.json"
    "web/package-lock.json"
)

# 显示清理预览
show_preview() {
    echo "🔍 将要清理的文件和目录："
    echo
    
    found_items=0
    for item in "${CLEANUP_ITEMS[@]}"; do
        if [[ "$item" == *"/" ]]; then
            # 目录
            if [ -d "$item" ]; then
                echo "  📁 $item"
                ((found_items++))
            fi
        elif [[ "$item" == *"*"* ]]; then
            # 通配符文件
            if ls $item 1> /dev/null 2>&1; then
                for file in $item; do
                    echo "  📄 $file"
                    ((found_items++))
                done
            fi
        else
            # 普通文件
            if [ -f "$item" ]; then
                echo "  📄 $item"
                ((found_items++))
            fi
        fi
    done
    
    if [ $found_items -eq 0 ]; then
        echo "  ✨ 没有找到需要清理的文件"
    fi
    
    echo
    echo "✅ 将要保留的重要文件："
    for file in "${KEEP_FILES[@]}"; do
        if [ -f "$file" ]; then
            echo "  📄 $file"
        fi
    done
    echo
}

# 执行清理
do_cleanup() {
    echo "🧹 开始清理..."
    
    cleaned_count=0
    
    for item in "${CLEANUP_ITEMS[@]}"; do
        if [[ "$item" == *"/" ]]; then
            # 目录
            if [ -d "$item" ]; then
                echo "  删除目录: $item"
                rm -rf "$item"
                ((cleaned_count++))
            fi
        elif [[ "$item" == *"*"* ]]; then
            # 通配符文件
            if ls $item 1> /dev/null 2>&1; then
                for file in $item; do
                    echo "  删除文件: $file"
                    rm -f "$file"
                    ((cleaned_count++))
                done
            fi
        else
            # 普通文件
            if [ -f "$item" ]; then
                echo "  删除文件: $item"
                rm -f "$item"
                ((cleaned_count++))
            fi
        fi
    done
    
    echo
    echo "✅ 清理完成！共删除 $cleaned_count 个文件/目录"
}

# 备份重要文件
backup_configs() {
    echo "💾 备份重要配置文件..."
    
    backup_dir="backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    # 备份.env文件（如果存在）
    if [ -f ".env" ]; then
        cp .env "$backup_dir/.env.backup"
        echo "  ✅ 已备份 .env -> $backup_dir/.env.backup"
    fi
    
    # 备份其他重要文件
    for file in "${KEEP_FILES[@]}"; do
        if [ -f "$file" ]; then
            dest_dir="$backup_dir/$(dirname $file)"
            mkdir -p "$dest_dir"
            cp "$file" "$backup_dir/$file.backup"
            echo "  ✅ 已备份 $file -> $backup_dir/$file.backup"
        fi
    done
    
    echo "  📁 备份目录: $backup_dir"
    echo
}

# 深度清理（包括Docker相关）
deep_cleanup() {
    echo "🔥 执行深度清理..."
    
    # 清理Docker资源
    if command -v docker &> /dev/null; then
        echo "  🐳 清理Docker资源..."
        
        # 停止相关容器
        docker-compose -f deployments/docker-compose.yml down 2>/dev/null || true
        
        # 清理未使用的镜像
        docker image prune -f
        
        # 清理未使用的容器
        docker container prune -f
        
        # 清理未使用的网络
        docker network prune -f
        
        # 清理未使用的卷（谨慎使用）
        if [ "$1" = "--include-volumes" ]; then
            echo "  ⚠️  清理Docker卷..."
            docker volume prune -f
        fi
    fi
    
    # 清理Go缓存
    if command -v go &> /dev/null; then
        echo "  🐹 清理Go缓存..."
        go clean -cache
        go clean -modcache
    fi
    
    # 清理npm缓存
    if command -v npm &> /dev/null; then
        echo "  📦 清理npm缓存..."
        npm cache clean --force
    fi
    
    echo "✅ 深度清理完成"
}

# 主菜单
main() {
    case "$1" in
        "--preview"|"-p")
            show_preview
            ;;
        "--backup"|"-b")
            backup_configs
            ;;
        "--force"|"-f")
            do_cleanup
            ;;
        "--deep")
            deep_cleanup "$2"
            ;;
        "--all")
            backup_configs
            do_cleanup
            deep_cleanup
            ;;
        "--help"|"-h")
            echo "用法: $0 [选项]"
            echo
            echo "选项:"
            echo "  -p, --preview           预览将要清理的文件"
            echo "  -b, --backup            备份重要配置文件"
            echo "  -f, --force             强制清理（无确认）"
            echo "  --deep                  深度清理（包括Docker和缓存）"
            echo "  --deep --include-volumes 深度清理（包括Docker卷）"
            echo "  --all                   执行完整清理（备份+清理+深度清理）"
            echo "  -h, --help              显示此帮助信息"
            echo
            echo "交互模式（无参数）："
            echo "  显示预览并询问是否清理"
            ;;
        *)
            show_preview
            echo "❓ 确认清理这些文件吗？ (y/N): "
            read -r response
            if [[ "$response" =~ ^[Yy]$ ]]; then
                do_cleanup
            else
                echo "❌ 清理已取消"
            fi
            ;;
    esac
}

# 运行主程序
main "$@"

echo
echo "💡 提示："
echo "  - 清理前建议先备份：$0 --backup"
echo "  - 查看清理预览：$0 --preview"
echo "  - 深度清理：$0 --deep"
echo "  - 完整清理：$0 --all"
echo
echo "📚 重要文件不会被清理："
echo "  - 项目源代码和配置"
echo "  - 环境变量文件"
echo "  - 文档和许可证"