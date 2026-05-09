# go-hichat-api

HiChat 2.0 — 基于 go-zero 的微服务 IM + 社交 + 动态空间。前后端分离，前端 `web/`（Next.js 16 + React 19 + Bun + Semi UI）。Go 1.23（toolchain 1.24.2）。

## 架构总览

```
┌── apps/<svc>/api  (HTTP 入口, .api 描述)
│      └── 调用 rpcclient ──┐
├── apps/<svc>/rpc  (gRPC 服务, .proto 描述)
│      └── 通过 etcd 注册/发现
├── apps/im/ws         (WebSocket 长连接, 心跳/ack/在线状态)
├── apps/streaming     (WebRTC SFU + 房间)
└── apps/task/{mq,cron}(Kafka 消费 + 定时任务)
```

数据：MySQL（业务表）/ MongoDB（聊天记录）/ Redis（会话/在线）/ Etcd（服务注册）/ Kafka（消息队列）。

## 服务清单

| 服务 | 层 | 入口 | 用途 |
|------|----|------|------|
| user      | api + rpc        | apps/user      | 注册、登录、用户资料、JWT 签发 |
| social    | api + rpc        | apps/social    | 好友、群、申请、管理员 |
| im        | api + rpc + ws   | apps/im        | 会话、聊天记录、读未读、长连接 |
| trend     | api + rpc        | apps/trend     | 动态、点赞、评论、屏蔽 |
| task      | mq + cron        | apps/task      | 异步消费 + 调度任务 |
| streaming | webrtc / sfu     | apps/streaming | 实时音视频房间 |
| demo      | (示例 / 非生产)  | apps/demo      | 内部 demo，不参与 hichat2.sh 启动 |

入口契约：
- HTTP: `apps/<svc>/api/<svc>.api`（user / social / im / trend）
- gRPC: `apps/<svc>/rpc/<svc>.proto`（user / social / im / trend）
- 自动汇总文档：`docs/specs/api.md`（运行 `/sync-api-docs` 更新）

## 启动 / 测试

- 一键起所有服务：`./hichat2.sh`（前置：MySQL/Redis/Etcd/Mongo/Kafka 已起）
- 单服务：`go run apps/<svc>/<layer>/<svc>.go -f apps/<svc>/<layer>/etc/<svc>-sample.yaml`
- 测试：`go test ./... -count=1`
- 前端：`cd web && bun dev`
- streaming 单独起：`apps/streaming/start.sh`（不在 hichat2.sh 列表内）

## 关键约定（必读）

所有约束以 `.claude/rules/` 为权威。修改前请先读对应文件：

- [Go 后端](.claude/rules/go-backend.md) — JSON、错误处理、并发、资源
- [数据库 / Model](.claude/rules/database-model.md) — schema 变更、三库兼容
- [go-zero / goctl](.claude/rules/go-zero.md) — 代码生成、目录约定
- [跨服务调用](.claude/rules/rpc-client.md) — 必须走 RPC 客户端
- [微服务边界](.claude/rules/microservice.md) — 不读对方数据库
- [WebSocket / IM](.claude/rules/websocket-im.md) — ws 连接、心跳、ack
- [MQ / 定时任务](.claude/rules/mq-task.md) — 幂等、消费失败处理
- [测试](.claude/rules/test-files.md) — table-driven、不 mock 数据库
- [前端](.claude/rules/frontend.md) — bun、Semi UI、i18n

## 常见任务速查

| 想做的事 | 用什么 |
|----------|--------|
| 全新功能从需求开始 | `/spec` |
| 新增一个微服务 | `/new-service` |
| 在已有服务加 HTTP 接口 | `/new-api` |
| 加 RPC 方法 | `/new-rpc` |
| 加数据库表 | `/new-model` |
| 启动 / 重启 / 看日志 | `/run-services` |
| 审查当前 diff（Go） | `/goreview` |
| 跑测试 | `/gotest` |
| 同步 API 文档 | `/sync-api-docs` |
| /clear 后恢复上下文 | `/catchup` |
| 新成员上手 | `/onboard` |

## Subagent

复杂或者需要独立上下文的审查工作走 subagent：

- `architecture-reviewer` — 分层职责、模式一致性
- `security-reviewer` — 注入、鉴权、密钥、SSRF
- `api-contract-reviewer` — `.api` / `.proto` 兼容性 / 命名 / optional 标记
