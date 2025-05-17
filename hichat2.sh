#!/bin/bash

# 定义日志目录
LOG_DIR="logs"
mkdir -p "$LOG_DIR"

# 存储所有后台进程 PID
declare -a PIDS=()

# 捕获终止信号（Ctrl+C），杀死所有子进程
trap 'kill ${PIDS[@]} 2>/dev/null; echo -e "\n所有服务已终止"; exit' SIGINT SIGTERM

# 启动服务函数
start_service() {
    local service_type=$1  # rpc 或 api
    local app_dir=$2       # 例如 user, social
    local name="$app_dir-$service_type"

    # 创建服务日志目录
    local log_path="$LOG_DIR/$name"
    mkdir -p "$log_path"

    # 启动服务并记录 PID
    echo "$name: 启动中..."
    case $service_type in
        rpc)
            go run "apps/$app_dir/rpc/$app_dir.go" -f "apps/$app_dir/rpc/etc/${app_dir}-local.yaml" >> "$log_path/$name.log" 2>&1 &
            ;;
        api)
            go run "apps/$app_dir/api/$app_dir.go" -f "apps/$app_dir/api/etc/${app_dir}-local.yaml" >> "$log_path/$name.log" 2>&1 &
            ;;
    esac
    PIDS+=($!)
    echo "$name: 运行中... PID: $!"
}

# 并行启动所有服务
start_service rpc user
start_service api user
start_service rpc social
start_service api social

# 阻塞主进程，直到收到终止信号
echo "所有服务已启动，按 Ctrl+C 停止"
wait ${PIDS[@]}







