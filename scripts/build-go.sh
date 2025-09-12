#!/bin/bash

# Go应用构建脚本
set -e

echo "🔧 构建Go应用..."

# 确保web/dist目录存在
if [ ! -d "web/dist" ]; then
    echo "❌ web/dist目录不存在，请先构建前端"
    exit 1
fi

# 检查必要文件
if [ ! -f "web/dist/index.html" ]; then
    echo "❌ web/dist/index.html不存在"
    exit 1
fi

if [ ! -d "web/dist/static" ]; then
    echo "❌ web/dist/static目录不存在"
    exit 1
fi

echo "✅ 前端文件检查通过"

# 构建Go应用
echo "🏗️ 构建Go应用..."
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o domain-max ./cmd/server

echo "✅ Go应用构建成功！"
