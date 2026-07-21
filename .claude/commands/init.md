# /init — 初始化 go-hichat-api 的 CLAUDE.md

为本项目生成（或更新）根 `CLAUDE.md`，作为 Claude 在本仓库工作的"第一手"指引。
本命令是项目专用版本，覆盖默认 init skill。

## 工作流

### 0. 前置检查
- 如果根目录已有 `CLAUDE.md`：读取它，**不要覆盖**，进入"补全模式"——只追加缺失部分，保留已有内容；保存前 diff 给用户确认。
- 如果没有：进入"生成模式"。

### 1. 探测项目结构（自动执行）
读取以下信号，**不要让用户重复回答**：
- 根目录 `README.md`、`go.mod`、`hichat2.sh`、`docker-compose.yaml`
- `apps/*` 子目录，识别每个服务有哪几层（`api/`、`rpc/`、`ws/`、`mq/`、`cron/`、`models/`）
- 找出所有 `.api` 与 `.proto`，提炼出对外接口与 RPC 方法概览
- 扫描 `pkg/` 下的共享包，列出关键工具（如 `pkg/db`、`pkg/interceptor`、`pkg/message`）
- 探测 `web/` 是否存在 + 框架（`package.json` 里的 next/react/bun）
- 读取 `.claude/rules/*.md`，把规则文件作为权威来源**链接**进 CLAUDE.md，而不是复制内容

### 2. 与用户确认（最多 3 个问题）
仅当无法从代码推断时才问，每次只问一个，例如：
- 主开发场景（"我会主要在哪几个服务上工作"）——为了在 CLAUDE.md 里给出更针对性的"日常任务速查"
- 是否需要为每个 `apps/<svc>/` 生成子级 CLAUDE.md
- 项目是否有未提交的开发约定（沟通群约定、PR 模板）需要写进去

如果用户全部不确定就跳过，按推断写。

### 3. 生成根 `CLAUDE.md`
按以下骨架写入项目根目录。**保持简短，每节不超过 30 行**；细节用链接指向 `.claude/rules/` 或 `docs/`。

```markdown
# go-hichat-api

HiChat 2.0 — 基于 go-zero 的微服务 IM + 社交 + 动态空间。前后端分离，前端 `web/`（Next.js + Bun + Semi UI）。

## 架构总览

```
┌── apps/<svc>/api  (HTTP 入口, .api 描述)
│      └── 调用 rpcclient ──┐
├── apps/<svc>/rpc  (gRPC 服务, .proto 描述)
│      └── 通过 etcd 注册/发现
├── apps/im/ws     (WebSocket 长连接, 心跳/ack/在线状态)
└── apps/task/{mq,cron} (Kafka 消费 + 定时任务)
```

数据：MySQL（业务表）/ MongoDB（聊天记录）/ Redis（会话/在线）/ Etcd（服务注册）/ Kafka（消息队列）。

## 服务清单

| 服务 | 层 | 入口 | 用途 |
|------|----|------|------|
| user   | api + rpc | apps/user      | 注册、登录、用户资料、JWT 签发 |
| social | api + rpc | apps/social    | 好友、群、申请、管理员 |
| im     | api + rpc + ws | apps/im   | 会话、聊天记录、读未读、长连接 |
| trend  | api + rpc | apps/trend     | 动态、点赞、评论、屏蔽 |
| task   | mq + cron | apps/task      | 异步消费 + 调度任务 |

入口契约：
- HTTP: `apps/<svc>/api/<svc>.api`
- gRPC: `apps/<svc>/rpc/<svc>.proto`
- 自动汇总文档见 `docs/specs/api.md`（运行 `/sync-api-docs` 更新）。

## 启动 / 测试

- 一键起所有服务：`./hichat2.sh`（前置：MySQL/Redis/Etcd/Mongo/Kafka 已起）
- 单服务：`go run apps/<svc>/<layer>/<svc>.go -f apps/<svc>/<layer>/etc/<svc>-sample.yaml`
- 测试：`go test ./... -count=1`
- 前端：`cd web && bun dev`

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
```

### 4. 可选：生成服务级 CLAUDE.md
如果用户在第 2 步要求，给每个 `apps/<svc>/` 写一份 30 行内的 CLAUDE.md：
- 该服务的对外契约（.api / .proto 概览）
- 这个服务的特殊点（im 的 ws、task 的 mq、user 的 JWT）
- 跑/调试这个服务的最小命令

### 5. 报告
输出三件事：
- 写了哪些文件
- 哪些假设是基于推断（用户可能想覆盖）
- 建议下一步：跑 `/onboard` 验证内容、或开始第一个 `/spec`

## 严格约束

- **不要**复述 `.claude/rules/*.md` 的内容到 CLAUDE.md 里——只链接。规则有更新时不会被忘记。
- **不要**写超过骨架里的 30 行/节。CLAUDE.md 是索引，不是文档。
- **不要**编造服务、接口、命令；只写探测到的真实内容。
- **不要**在 CLAUDE.md 里写"如何安装 Go"这种通用信息——README 已经覆盖。
