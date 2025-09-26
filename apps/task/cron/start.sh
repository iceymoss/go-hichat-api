#!/bin/bash

# 定时任务服务启动脚本

echo "Starting Cron Task Service..."

# 检查配置文件是否存在
if [ ! -f "etc/cron-local.yaml" ]; then
    echo "Error: Configuration file etc/cron-local.yaml not found!"
    exit 1
fi

# 启动服务
go run cron.go -f etc/cron-local.yaml
