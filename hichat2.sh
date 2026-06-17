#!/bin/bash

LOG_DIR="logs"
mkdir -p "$LOG_DIR"
declare -a PIDS=()

cleanup() {
  echo -e "\n正在停止所有服务..."
  for pid in "${PIDS[@]}"; do
    # 负号表示杀掉整个进程组（setsid 使每个服务成为组长），
    # 这样 go run 派生的实际服务二进制也会被一起终止
    kill -TERM -- "-$pid" 2>/dev/null
  done
  wait 2>/dev/null
  echo "所有服务已终止"
  exit 0
}
trap cleanup SIGINT SIGTERM

start_service() {
  local service_type=$1
  local app_dir=$2
  local name="$app_dir-$service_type"
  local main go_file cfg

  mkdir -p "$LOG_DIR/$name"
  echo "$name: 启动中..."

  case $service_type in
    rpc)  go_file="apps/$app_dir/rpc/$app_dir.go"; cfg="apps/$app_dir/rpc/etc/${app_dir}-sample.yaml" ;;
    api)  go_file="apps/$app_dir/api/$app_dir.go"; cfg="apps/$app_dir/api/etc/${app_dir}-sample.yaml" ;;
    im)   go_file="apps/$app_dir/ws/im.go";        cfg="apps/$app_dir/ws/etc/${app_dir}-sample.yaml" ;;
    task) go_file="apps/$app_dir/mq/mq.go";        cfg="apps/$app_dir/mq/etc/mq-sample.yaml" ;;
    *)    echo "$name: 未知服务类型 $service_type"; return 1 ;;
  esac

  echo go run "$go_file" -f "$cfg"
  # setsid 让服务在独立进程组中运行，$! 即为组长 PID（= PGID）
  setsid go run "$go_file" -f "$cfg" >> "$LOG_DIR/$name/$name.log" 2>&1 &

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
  "api im"
  "im im"
  "task task"
  "rpc trend"
  "api trend"
)

# 遍历数组启动服务
for service in "${SERVICES[@]}"; do
  IFS=' ' read -r type dir <<< "$service"
  start_service "$type" "$dir"

  sleep 2
done

echo "所有服务已启动，按 Ctrl+C 停止"
# 等待所有后台任务；收到 SIGINT/SIGTERM 时 wait 被中断并触发 cleanup
wait