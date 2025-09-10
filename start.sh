#!/bin/bash

# 域名管理系统启动脚本

set -e

echo "🚀 域名管理系统启动脚本"
echo "========================"

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

# 检查Docker Compose是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose未安装，请先安装Docker Compose"
    exit 1
fi

# 检查.env文件是否存在
if [ ! -f .env ]; then
    echo "📝 创建环境配置文件..."
    cp env.example .env
    echo "⚠️  请编辑 .env 文件配置数据库密码和其他设置"
    echo "   nano .env"
    echo ""
    read -p "按回车键继续..."
fi

# 显示当前配置
echo "📋 当前配置:"
echo "   端口: $(grep PORT .env | cut -d'=' -f2 || echo '8080')"
echo "   数据库: $(grep DB_TYPE .env | cut -d'=' -f2 || echo 'postgres')"
echo ""

# 选择启动模式
echo "请选择启动模式:"
echo "1) 开发模式 (开发和测试)"
echo "2) 生产模式 (推荐用于生产环境)"
echo "3) 仅构建镜像"
echo "4) 查看服务状态"
echo "5) 停止服务"
echo "6) 查看日志"
echo "7) 备份数据库"
echo ""

read -p "请输入选项 (1-7): " choice

case $choice in
    1)
        echo "🔧 启动开发模式..."
        docker-compose up -d
        ;;
    2)
        echo "🏭 启动生产模式..."
        if [ -f docker-compose.prod.yml ]; then
            docker-compose -f docker-compose.prod.yml up -d
        else
            echo "⚠️  生产配置文件不存在，使用默认配置..."
            docker-compose up -d
        fi
        ;;
    3)
        echo "🔨 构建镜像..."
        docker-compose build
        ;;
    4)
        echo "📊 服务状态:"
        docker-compose ps
        echo ""
        echo "健康检查:"
        curl -s http://localhost:8080/api/health || echo "❌ 应用未响应"
        ;;
    5)
        echo "⏹️  停止服务..."
        docker-compose down
        ;;
    6)
        echo "📜 查看日志..."
        docker-compose logs -f
        ;;
    7)
        echo "💾 备份数据库..."
        mkdir -p backups
        timestamp=$(date +%Y%m%d_%H%M%S)
        docker-compose exec db pg_dump -U postgres domain_manager | gzip > backups/backup_$timestamp.sql.gz
        echo "✅ 备份完成: backups/backup_$timestamp.sql.gz"
        ;;
    *)
        echo "❌ 无效选项"
        exit 1
        ;;
esac

if [ $choice -eq 1 ] || [ $choice -eq 2 ]; then
    echo ""
    echo "🎉 启动完成！"
    echo ""
    echo "📍 访问地址:"
    echo "   用户端: http://localhost:8080"
    echo "   管理后台: http://localhost:8080/admin"
    echo "   API健康检查: http://localhost:8080/api/health"
    echo ""
    echo "👤 默认管理员账户:"
    echo "   邮箱: admin@example.com"
    echo "   密码: admin123"
    echo "   ⚠️  请立即修改默认密码！"
    echo ""
    echo "📚 更多信息请查看 README.md"
    echo ""
    echo "🔧 常用命令:"
    echo "   查看状态: docker-compose ps"
    echo "   查看日志: docker-compose logs -f"
    echo "   停止服务: docker-compose down"
    echo "   重启服务: docker-compose restart"
fi
