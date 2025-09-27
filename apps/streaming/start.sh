#!/bin/bash

# 流媒体服务启动脚本

echo "Starting Streaming Service..."

# 检查配置文件是否存在
if [ ! -f "etc/streaming-local.yaml" ]; then
    echo "Error: Configuration file etc/streaming-local.yaml not found!"
    exit 1
fi

# 启动服务
go run streaming.go -f etc/streaming-local.yaml
