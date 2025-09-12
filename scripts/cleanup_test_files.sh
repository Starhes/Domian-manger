#!/bin/bash

# 域名管理系统测试文件清理脚本
# 用途：清理开发和测试过程中生成的临时文件

echo "=== 域名管理系统测试文件清理脚本 ==="
echo

# 定义需要清理的测试文件和目录
TEST_FILES=(
    "scripts/setup_smtp.go"
    "scripts/test_smtp.go"
    "scripts/check_smtp.sh"
    "smtp_config.env"
    "test_output.log"
    "debug.log"
    "*.test"
    "coverage.out"
    "profile.out"
)

# 定义需要清理的测试目录
TEST_DIRS=(
    "tmp/"
    "temp/"
    "test_data/"
    "coverage/"
    ".test_cache/"
)

# 定义生产环境保留的文件
KEEP_FILES=(
    "docs/SMTP_CONFIG.md"
    "docs/SMTP_IMPROVEMENT_SUMMARY.md" 
    "env.example"
    "internal/services/email.go"
    "internal/services/admin.go"
    "internal/models/models.go"
)

# 显示清理预览
show_preview() {
    echo "🔍 将要清理的测试文件："
    echo
    
    for file in "${TEST_FILES[@]}"; do
        if [ -f "$file" ] || [ -n "$(ls $file 2>/dev/null)" ]; then
            echo "  📄 $file"
        fi
    done
    
    for dir in "${TEST_DIRS[@]}"; do
        if [ -d "$dir" ]; then
            echo "  📁 $dir"
        fi
    done
    
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
    echo "🧹 开始清理测试文件..."
    
    cleaned_count=0
    
    # 清理测试文件
    for file in "${TEST_FILES[@]}"; do
        if [ -f "$file" ]; then
            echo "  删除文件: $file"
            rm -f "$file"
            ((cleaned_count++))
        elif [ -n "$(ls $file 2>/dev/null)" ]; then
            echo "  删除匹配文件: $file"
            rm -f $file
            ((cleaned_count++))
        fi
    done
    
    # 清理测试目录
    for dir in "${TEST_DIRS[@]}"; do
        if [ -d "$dir" ]; then
            echo "  删除目录: $dir"
            rm -rf "$dir"
            ((cleaned_count++))
        fi
    done
    
    # 清理编译产物
    if [ -f "domain-manager" ]; then
        echo "  删除编译文件: domain-manager"
        rm -f domain-manager
        ((cleaned_count++))
    fi
    
    if [ -f "domain-manager.exe" ]; then
        echo "  删除编译文件: domain-manager.exe"
        rm -f domain-manager.exe
        ((cleaned_count++))
    fi
    
    echo
    echo "✅ 清理完成！共删除 $cleaned_count 个文件/目录"
}

# 备份重要配置
backup_configs() {
    echo "💾 备份重要配置文件..."
    
    backup_dir="backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    # 备份.env文件（如果存在）
    if [ -f ".env" ]; then
        cp .env "$backup_dir/.env.backup"
        echo "  ✅ 已备份 .env -> $backup_dir/.env.backup"
    fi
    
    # 备份重要文档
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
        "--help"|"-h")
            echo "用法: $0 [选项]"
            echo
            echo "选项:"
            echo "  -p, --preview    预览将要清理的文件"
            echo "  -b, --backup     备份重要配置文件"
            echo "  -f, --force      强制清理（无确认）"
            echo "  -h, --help       显示此帮助信息"
            echo
            echo "交互模式（无参数）："
            echo "  显示预览并询问是否清理"
            ;;
        *)
            show_preview
            echo "❓ 确认清理这些测试文件吗？ (y/N): "
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
echo "  - 生产部署前应清理所有测试文件"
echo
echo "📚 保留的重要文件不会被清理："
echo "  - 项目源代码和配置"
echo "  - 文档和说明文件"
echo "  - 数据库脚本"