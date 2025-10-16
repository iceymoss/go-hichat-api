#!/bin/bash

echo "🚀 启动视频通话 Demo 服务..."
echo ""

# 检查配置文件
if [ ! -f "etc/demo.yaml" ]; then
    echo "❌ 错误: 找不到配置文件 etc/demo.yaml"
    exit 1
fi

# 启动服务
go run demo.go -f etc/demo.yaml

