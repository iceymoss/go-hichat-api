<p align="center">
  <img src="../assets/brand/hichat-green-lockup.svg" alt="HiChat" width="320" />
</p>

# go-hichat-api

[English](../README.md) | [简体中文](README.zh-CN.md)

go-hichat-api 是 HiChat 2.0 的后端与 Web 客户端仓库，是一个基于 go-zero 的微服务即时通讯与社交平台。项目整合 REST API、zRPC 服务、WebSocket 长连接、Kafka 异步链路、MongoDB 聊天记录、MySQL 业务数据、Redis 运行时状态，以及独立的 WebRTC 流媒体服务。

这个仓库可以作为现代 IM 系统的实践参考，覆盖清晰的服务边界、消息投递、已读回执、在线状态、动态通知、富媒体消息、社交关系、动态空间和 Next.js Web 客户端。

## 项目亮点

- 基于 go-zero REST 和 zRPC 的微服务架构。
- 通过 `.api` 和 `.proto` 文件维护 API 优先的服务契约。
- WebSocket 网关支持认证、心跳、在线状态、消息 ACK、已读回执和实时推送。
- Kafka 链路处理聊天投递、已读事件、消息撤回、动态通知和后台任务。
- MongoDB 存储聊天记录，MySQL 存储业务数据，Redis 存储会话、缓存、在线状态和运行时协调数据。
- 独立 WebRTC 流媒体服务，支持通话、会议、屏幕共享、直播、房间和 SFU 流程。
- `web/` 下提供完整 Web 客户端，技术栈为 Next.js 16、React 19、Bun、TypeScript、Tailwind CSS 和 Semi UI。

## 核心能力

| 领域 | 能力 |
| --- | --- |
| 用户与账号 | 手机号/密码登录、JWT 签发、手机/邮箱验证码、密码重置、资料管理、头像上传、账号注销、用户搜索和内部用户查询 RPC。 |
| 社交关系 | 好友申请、好友列表、备注、拉黑、朋友圈权限、消息通知设置、标签、好友举报和在线状态查询。 |
| 群组 | 建群、搜索群、入群申请、成员邀请、邀请 token、成员管理、群公告、角色、管理员操作、群主转让和群 `@`。 |
| 即时通讯 | 单聊/群聊会话、会话置顶/免打扰、MongoDB 聊天记录、文本/文件/语音/图片/视频消息、引用、提及、未读状态、已读记录、消息撤回和媒体上传。 |
| 实时网关 | WebSocket 认证、路由分发、Redis 在线状态、Kafka 消息投递、服务端推送、ACK 跟踪、重试、去重和动态通知。 |
| 动态空间 | 动态发布、可见范围、媒体资源、评论、回复、点赞、草稿、未读计数、动态消息通知和在线推送。 |
| 异步任务 | 聊天、已读、撤回和动态通知事件的 Kafka 消费，以及 cron 任务扩展点。 |
| 流媒体 | WebRTC 单聊通话、群组通话、会议、屏幕共享、直播、信令、房间和 SFU 组件。 |
| Web 客户端 | Next.js 应用、Bun 脚本、TypeScript、Tailwind CSS、Semi UI，开发服务默认运行在 `3001` 端口。 |

## 架构

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ L0 客户端层                                                                 │
│                                                                              │
│  Web Client (web/: Next.js + React)        Mobile / Third-party Clients       │
└───────────────┬────────────────────────────┬───────────────────────────┬─────┘
                │ REST                       │ WebSocket                 │ WebRTC
                ▼                            ▼                           ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│ L1 接入层                                                                   │
│                                                                              │
│  HTTP APIs                         实时接入                    媒体接入      │
│  ┌─────────────────────────────┐   ┌───────────────────────┐    ┌──────────┐ │
│  │ user/api   social/api       │   │ im/ws                 │    │streaming │ │
│  │ im/api     trend/api        │   │ auth heartbeat ack    │    │signaling │ │
│  │ REST routes + JWT context   │   │ online push routing   │    │rooms SFU │ │
│  └──────────────┬──────────────┘   └───────────┬───────────┘    └────┬─────┘ │
└─────────────────┼──────────────────────────────┼─────────────────────┼───────┘
                  │ zRPC                         │ 发布/消费           │ Redis
                  ▼                              ▼                     ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│ L2 领域服务层                                                               │
│                                                                              │
│  ┌────────────┐  ┌──────────────┐  ┌────────────┐  ┌──────────────┐          │
│  │ user/rpc   │  │ social/rpc   │  │ im/rpc     │  │ trend/rpc    │          │
│  │ 认证       │  │ 好友         │  │ 会话       │  │ 动态         │          │
│  │ 资料       │  │ 群组         │  │ 聊天记录   │  │ 评论         │          │
│  │ 验证码     │  │ 申请         │  │ 已读/撤回  │  │ 点赞/通知    │          │
│  └─────┬──────┘  └──────┬───────┘  └─────┬──────┘  └──────┬───────┘          │
└────────┼────────────────┼────────────────┼────────────────┼─────────────────┘
         │                │                │                │
         │ MySQL          │ MySQL          │ MongoDB        │ MySQL + Kafka
         ▼                ▼                ▼                ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│ L3 事件与异步层                                                             │
│                                                                              │
│  Kafka Topics                                                                │
│  ┌────────────────┐ ┌───────────────┐ ┌────────────────┐ ┌───────────────┐  │
│  │ chat-transfer  │ │ read-transfer │ │ recall-transfer│ │ trend-notify  │  │
│  └───────┬────────┘ └──────┬────────┘ └───────┬────────┘ └──────┬────────┘  │
│          └─────────────────┴──────────┬───────┴─────────────────┘           │
│                                        ▼                                     │
│  apps/task/mq: 持久化聊天、更新已读、推送撤回和动态通知                      │
│  apps/task/cron: 统计、清理和可扩展定时任务                                  │
└────────────────────────────────────────┬─────────────────────────────────────┘
                                         │ 持久化 / 更新 / 推送
                                         ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│ L4 数据与运行时基础设施                                                     │
│                                                                              │
│  MySQL: 用户、好友、群组、动态、评论、点赞、通知                             │
│  MongoDB: 聊天记录、已读记录、撤回状态                                      │
│  Redis: session/JWT 状态、在线状态、缓存、WS 运行态、房间状态                │
│  Etcd: go-zero RPC 服务注册与发现                                           │
└──────────────────────────────────────────────────────────────────────────────┘

关键连接
1. Web/Mobile -> HTTP APIs -> zRPC -> Domain RPC -> MySQL/MongoDB。
2. Web/Mobile -> im/ws -> Kafka -> task/mq -> MongoDB + im/ws 推送。
3. trend/rpc -> Kafka trend-notify -> task/mq -> im/ws -> 在线客户端。
4. RPC 服务注册到 Etcd；API 服务从 Etcd 发现 RPC 节点。
5. im/ws 和 streaming 使用 Redis 存储在线状态、会话、缓存和房间状态。
```

## 消息链路

### 聊天投递

```mermaid
sequenceDiagram
  SenderClient->>IMWS: WebSocket chat.user
  IMWS->>KafkaMsgChatTransfer: 发布聊天消息
  KafkaMsgChatTransfer->>TaskMQ: 消费 MsgChatTransfer
  TaskMQ->>MongoDBChatLog: 持久化聊天记录
  TaskMQ->>IMWS: 通过 push 路由推送
  IMWS->>ReceiverClient: WebSocket 消息帧
  ReceiverClient-->>IMWS: ACK 帧
  IMWS-->>SenderClient: 可选发送方回响和服务端 msgId
```

### 已读回执

```mermaid
sequenceDiagram
  ReaderClient->>IMWS: WebSocket chat.markChat
  IMWS->>KafkaMsgReadTransfer: 发布已读事件
  KafkaMsgReadTransfer->>TaskMQ: 消费 MsgReadTransfer
  TaskMQ->>MongoDBChatLog: 更新已读 bitmap 和已读时间
  TaskMQ->>IMWS: 推送已读回执控制消息
  IMWS->>SenderClient: WebSocket readRecords 更新
  SenderClient->>IMAPI: GET /v1/im/chatlog/readRecords
  IMAPI->>MongoDBChatLog: 查询详细已读和未读用户
```

### 消息撤回

```mermaid
sequenceDiagram
  OperatorClient->>IMAPI: POST /v1/im/chatlog/recall
  IMAPI->>IMRPC: RecallMsg
  IMRPC->>MongoDBChatLog: 标记消息为已撤回
  IMAPI->>KafkaMsgRecallTransfer: 发布撤回事件
  KafkaMsgRecallTransfer->>TaskMQ: 消费 MsgRecallTransfer
  TaskMQ->>IMWS: 推送撤回控制帧
  IMWS->>OnlineClients: WebSocket 撤回帧
```

### 动态通知

```mermaid
sequenceDiagram
  ActorClient->>TrendAPI: 创建提及、评论、回复或点赞
  TrendAPI->>TrendRPC: 执行业务逻辑
  TrendRPC->>MySQLTrendData: 写入动态与通知数据
  TrendRPC->>KafkaTrendNotifyTransfer: 发布 TrendNotifyTransfer
  KafkaTrendNotifyTransfer->>TaskMQ: 消费动态通知事件
  TaskMQ->>IMWS: push.trend 携带 TrendNotify payload
  IMWS->>ReceiverClient: 在线时推送 trend.notify
```

## 服务列表

| 服务 | 层 | 职责 |
| --- | --- | --- |
| `user` | `api`, `rpc`, `models` | 账号、认证、资料、验证码、用户查询 |
| `social` | `api`, `rpc`, `socialmodels` | 好友、好友申请、群组、群成员、邀请链接、群公告 |
| `im` | `api`, `rpc`, `ws`, `models`, `immodels` | 会话、聊天记录、已读回执、消息撤回、WebSocket 网关 |
| `trend` | `api`, `rpc`, `models` | 动态、评论、点赞、草稿、媒体、动态通知 |
| `task` | `mq`, `cron` | Kafka 消费者和定时任务 |
| `streaming` | `internal`, `room`, `sfu`, `webrtc` | WebRTC 通话、房间、会议、屏幕共享、直播 |
| `demo` | 独立示例 | 内部 demo 服务，不属于主启动脚本 |

## 技术栈

- 后端：Go 1.23、toolchain Go 1.24.2、go-zero、zRPC、gRPC、goctl。
- 实时通信：WebSocket、Kafka、WebRTC、Pion。
- 存储：MySQL、MongoDB、Redis。
- 服务发现：Etcd。
- 前端：Next.js 16、React 19、Bun、TypeScript、Tailwind CSS、Semi UI。

## 仓库结构

由 `tree -L 2` 生成。

```text
.
├── CLAUDE.md
├── LICENSE
├── README.md
├── apps
│   ├── demo
│   ├── im
│   ├── social
│   ├── streaming
│   ├── task
│   ├── trend
│   └── user
├── cmd
├── common
├── config
│   ├── config-local.yaml
│   └── config-sample.yaml
├── deploy
│   ├── dockerfile
│   ├── sql
│   ├── sql_init.go
│   └── trendmig
├── docker-compose.yaml
├── docs
│   ├── README.zh-CN.md
│   ├── api.md
│   ├── development-guide.md
│   ├── imgs
│   └── specs
├── go.mod
├── go.sum
├── hichat2.sh
├── logs
│   ├── im-api
│   ├── im-im
│   ├── im-rpc
│   ├── im-ws
│   ├── social-api
│   ├── social-rpc
│   ├── task-mq
│   ├── task-task
│   ├── trend-api
│   ├── trend-rpc
│   ├── user-api
│   └── user-rpc
├── pkg
│   ├── 2fa
│   ├── bitmap
│   ├── config
│   ├── constants
│   ├── ctxdata
│   ├── db
│   ├── encrypt
│   ├── errors
│   ├── http
│   ├── interceptor
│   ├── logger
│   ├── message
│   ├── relationcache
│   ├── sensitive
│   ├── storage
│   ├── systemconfig
│   ├── test
│   ├── transaction
│   ├── utils
│   ├── wuid
│   └── xerr
├── resources
│   └── sensitive
├── temp
│   ├── avatar
│   ├── emoji
│   ├── favorite
│   ├── im
│   ├── img.png
│   ├── img_1.png
│   ├── img_2.png
│   ├── img_3.png
│   ├── img_4.png
│   └── trend
├── test.sh
└── web
    ├── Caddyfile
    ├── bun.lock
    ├── components.json
    ├── dev.log
    ├── dist
    ├── download
    ├── eslint.config.mjs
    ├── examples
    ├── next-env.d.ts
    ├── next.config.ts
    ├── node_modules
    ├── package.json
    ├── postcss.config.mjs
    ├── public
    ├── scripts
    ├── src
    ├── tailwind.config.ts
    ├── tsconfig.json
    ├── tsconfig.tsbuildinfo
    ├── upload
    └── worklog.md

71 directories, 32 files
```

## 快速开始

### 前置依赖

- Go 1.23 或更高版本，使用 toolchain Go 1.24.2。
- Web 客户端需要 Bun。
- MySQL、Redis、Etcd、MongoDB 和 Kafka。
- go-zero 工具链：`goctl`、`protoc`、`protoc-gen-go` 和 `protoc-gen-go-grpc`。

本地依赖安装和代码生成说明见 [开发指南](development-guide.md)。

### 启动后端服务

先启动所需基础设施，再运行主后端服务：

```bash
./hichat2.sh
```

该脚本会启动 user、social、IM、task 和 trend 服务，并将日志写入 `logs/`。

按需手动启动单个服务：

```bash
go run apps/<service>/<layer>/<service>.go -f apps/<service>/<layer>/etc/<service>-sample.yaml
```

单独启动 streaming 服务：

```bash
apps/streaming/start.sh
```

### 启动 Web 客户端

```bash
cd web
bun install
bun dev
```

Web 开发服务默认运行在 `3001` 端口。

## 开发

- [开发指南](development-guide.md)：本地中间件配置、go-zero 工具链、代码生成、启动说明和 Docker 示例。
- [API 文档](api.md)：生成的 REST 和 gRPC 契约汇总。
- [功能规格](specs)：功能分析、设计说明和实现记录。

## 测试

在仓库根目录运行后端测试：

```bash
go test ./... -count=1
```

在 `web/` 下运行前端 lint：

```bash
bun lint
```

## 贡献

请参考 [贡献指南](https://github.com/iceymoss/go-hichat-api/issues/207) 了解贡献规范。

## 许可证

本项目基于 [Apache License 2.0](../LICENSE) 开源。
