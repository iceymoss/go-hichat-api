#!/bin/bash

LOG_DIR="logs"
mkdir -p "$LOG_DIR"
declare -a PIDS=()
trap 'kill ${PIDS[@]} 2>/dev/null; echo -e "\n所有服务已终止"; exit' SIGINT SIGTERM

start_service() {
  local service_type=$1
  local app_dir=$2
  local name="$app_dir-$service_type"

  mkdir -p "$LOG_DIR/$name"
  echo "$name: 启动中..."

  case $service_type in
    rpc)
      go run "apps/$app_dir/rpc/$app_dir.go" -f "apps/$app_dir/rpc/etc/${app_dir}-local.yaml" >> "$LOG_DIR/$name/$name.log" 2>&1 &
      ;;
    api)
      go run "apps/$app_dir/api/$app_dir.go" -f "apps/$app_dir/api/etc/${app_dir}-local.yaml" >> "$LOG_DIR/$name/$name.log" 2>&1 &
      ;;
    im)
      go run "apps/$app_dir/ws/ws/ws.go" -f "apps/$app_dir/ws/etc/${app_dir}-local.yaml" >> "$LOG_DIR/$name/$name.log" 2>&1 &
      ;;
  esac

  PIDS+=($!)
  echo "$name: 运行中... PID: $!"
}

# 服务配置数组（类型 + 目录）
SERVICES=(
  "rpc user"
  "api user"
  "rpc social"
  "api social"
  "rpc im"
  "im im"
)

# 遍历数组启动服务
for service in "${SERVICES[@]}"; do
  IFS=' ' read -r type dir <<< "$service"
  start_service "$type" "$dir"
done

echo "所有服务已启动，按 Ctrl+C 停止"
wait ${PIDS[@]}