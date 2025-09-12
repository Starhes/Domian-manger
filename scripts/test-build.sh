#!/bin/bash

# 测试前端构建脚本
set -e

echo "🔍 测试前端构建..."

cd web

echo "📦 安装依赖..."
npm install

echo "🔧 运行TypeScript检查..."
npx tsc --noEmit

echo "🏗️ 运行构建..."
npm run build

echo "✅ 构建成功！"
echo "📁 构建输出目录："
ls -la dist/

echo "📄 检查生成的文件："
ls -la dist/static/ 2>/dev/null || echo "静态文件目录不存在（正常）"

echo "🌐 检查index.html..."
if [ -f "dist/index.html" ]; then
    echo "✅ index.html 生成成功"
else
    echo "❌ index.html 未生成"
    exit 1
fi

echo "🎉 所有测试通过！"
