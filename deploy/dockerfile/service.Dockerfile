# 通用 go-zero 服务镜像：通过 build arg SVC_DIR 指定要编译的服务目录。
# 同一份 Dockerfile 构建所有后端服务（user/social/im/trend/task/streaming）
# 以及数据库迁移工具（deploy）。
#
#   docker build -f deploy/dockerfile/service.Dockerfile --build-arg SVC_DIR=apps/user/api -t hichat-user-api .
#
# 运行期配置文件路径由 CONFIG_FILE 环境变量给出（compose 以只读卷挂载覆盖）。

# ---------- 构建阶段 ----------
# go.mod 要求 go >= 1.25.0；官方镜像 GOTOOLCHAIN=local 不会自动下载更高版本，故基础镜像须 >= 1.25
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 国内 Go 代理，加速依赖下载
RUN go env -w GOPROXY=https://goproxy.cn,direct

# 先拷依赖清单并下载（模块缓存用 BuildKit cache mount，跨服务复用、不进镜像层）
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    go mod download

# 拷贝完整源码（所有服务共享此层）
COPY . .

# SVC_DIR 在此处声明，保证上面的依赖/源码层在各服务间复用，只有编译这一层按服务变化。
# 模块缓存 + 编译缓存用 BuildKit cache mount 并 sharing=locked：
#   - 依赖只编译一次、各服务复用，避免 12 个服务并行各编一遍把磁盘撑爆；
#   - locked 串行化对缓存的写入，进一步降低峰值磁盘占用。
ARG SVC_DIR
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    test -n "$SVC_DIR" || (echo "build arg SVC_DIR is required" && exit 1); \
    cd "$SVC_DIR" && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o /app/service

# ---------- 运行阶段 ----------
FROM alpine:3.18

# CA 证书 + 时区 + 健康检查用的 curl/wget
RUN apk add --no-cache ca-certificates tzdata curl wget && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 仅拷贝编译产物；运行期配置全部由 compose 挂载提供
COPY --from=builder /app/service /app/service

# 上传文件目录（user-api / im-api 静态资源）与日志目录，compose 用命名卷挂载持久化
RUN mkdir -p /app/temp /app/logs /app/config

# CONFIG_FILE 为空时直接运行（如迁移工具）；非空时以 -f 传入 go-zero 配置
ENV CONFIG_FILE=""
ENTRYPOINT ["/bin/sh", "-c", "if [ -n \"$CONFIG_FILE\" ]; then exec /app/service -f \"$CONFIG_FILE\"; else exec /app/service; fi"]
