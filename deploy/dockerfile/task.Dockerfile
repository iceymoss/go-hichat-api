# 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置国内Go代理
RUN go env -w GOPROXY=https://goproxy.cn,direct

# 预下载依赖 (利用Docker缓存层)
COPY go.mod go.sum ./
RUN go mod download

# 拷贝整个项目
COPY . .

# 编译task二进制文件
RUN cd apps/task/mq && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o /app/task-mq

# 运行阶段 - 使用极简镜像
FROM alpine:3.18

# 安装CA证书和必要的工具
RUN apk add --no-cache ca-certificates tzdata curl && \
    # 设置时区
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    # 创建应用用户
    adduser -D -u 10001 appuser && \
    # 创建应用目录结构
    mkdir -p /app/config /app/logs && \
    chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 从构建阶段复制编译好的二进制文件
COPY --from=builder --chown=appuser /app/task-mq /app/task-mq

# 从构建阶段复制配置文件
COPY --from=builder --chown=appuser /app/config /app/config
COPY --from=builder --chown=appuser /app/apps/task/mq/etc /app/apps/task/mq/etc

# 设置环境变量
ENV APP_NAME="task-mq" \
    CONFIG_DIR="/app/config" \
    CONFIG_PATH="apps/task/mq/etc/task-mq-local" \
    LOG_DIR="/app/logs" \
    SERVICE_PORT=10000

# 暴露服务端口
EXPOSE 10091

# 工作目录
WORKDIR /app

#apps/task/mq/etc/mq-local
# 启动应用 (使用环境变量确定配置)
ENTRYPOINT ["/app/task-mq", "-f", "apps/task/mq/etc/mq-local.yaml"]