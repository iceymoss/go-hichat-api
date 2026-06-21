# HiChat 前端（Next.js 16 standalone）镜像
# 构建用 bun，产物为自带 Node 服务的 standalone server.js，运行在 3001 端口。

# ---------- 构建阶段 ----------
FROM oven/bun:1 AS builder
WORKDIR /app

# 先装依赖，利用层缓存
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

# 拷贝源码并构建（next.config.ts 已设 output: "standalone"）
COPY . .
RUN bun run build

# ---------- 运行阶段 ----------
FROM oven/bun:1-slim AS runner
WORKDIR /app

ENV NODE_ENV=production \
    HOSTNAME=0.0.0.0 \
    PORT=3001

# standalone 产物 + 静态资源 + public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

EXPOSE 3001
CMD ["bun", "server.js"]
