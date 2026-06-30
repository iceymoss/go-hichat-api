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

## 产品截图

截取自基于演示数据集运行的 Web 客户端（14 个预置用户，含好友、群组、会话和动态，由 [`scripts/mockdata`](../scripts/mockdata) 生成）。

### 账号

<table>
  <tr>
    <td width="33%" align="center"><img src="screenshots/login.png" alt="登录"/><br/><sub><b>登录</b></sub></td>
    <td width="33%" align="center"><img src="screenshots/register.png" alt="注册"/><br/><sub><b>注册</b></sub></td>
    <td width="33%" align="center"><img src="screenshots/forgot-password.png" alt="忘记密码"/><br/><sub><b>忘记密码</b></sub></td>
  </tr>
</table>

### 即时通讯

<table>
  <tr>
    <td width="50%" align="center"><img src="screenshots/conversation-list.png" alt="会话列表"/><br/><sub><b>会话列表</b></sub></td>
    <td width="50%" align="center"><img src="screenshots/single-chat.png" alt="单聊会话"/><br/><sub><b>单聊会话</b></sub></td>
  </tr>
  <tr>
    <td width="50%" align="center"><img src="screenshots/group-chat.png" alt="群会话"/><br/><sub><b>群会话</b></sub></td>
    <td width="50%" align="center"><img src="screenshots/create-group.png" alt="创建群聊"/><br/><sub><b>创建群聊</b></sub></td>
  </tr>
  <tr>
    <td width="50%" align="center"><img src="screenshots/conversation-batch-actions.png" alt="多选会话、免打扰、置顶"/><br/><sub><b>多选 · 免打扰 · 置顶</b></sub></td>
    <td width="50%" align="center"><img src="screenshots/chat-user-card.png" alt="对话框用户资料卡片"/><br/><sub><b>对话内资料卡片</b></sub></td>
  </tr>
</table>

### 好友

<table>
  <tr>
    <td width="50%" align="center"><img src="screenshots/friend-list-and-settings.png" alt="好友列表、详情、设置"/><br/><sub><b>好友列表 · 详情 · 设置</b></sub></td>
    <td width="50%" align="center"><img src="screenshots/friend-request-detail.png" alt="好友申请详情"/><br/><sub><b>好友申请详情</b></sub></td>
  </tr>
  <tr>
    <td width="50%" align="center"><img src="screenshots/friend-requests-received.png" alt="收到的好友申请"/><br/><sub><b>收到的申请</b></sub></td>
    <td width="50%" align="center"><img src="screenshots/friend-requests-sent.png" alt="我发起的好友申请"/><br/><sub><b>我发起的申请</b></sub></td>
  </tr>
</table>

### 动态空间（朋友圈）

<table>
  <tr>
    <td width="50%" align="center"><img src="screenshots/publish-moment.png" alt="发布动态"/><br/><sub><b>发布动态</b></sub></td>
    <td width="50%" align="center"><img src="screenshots/moments-feed-and-detail.png" alt="动态列表、详情、评论、点赞"/><br/><sub><b>列表 · 详情 · 评论 · 点赞</b></sub></td>
  </tr>
  <tr>
    <td width="50%" align="center"><img src="screenshots/moments.png" alt="我的朋友圈"/><br/><sub><b>我的朋友圈</b></sub></td>
    <td width="50%" align="center"><img src="screenshots/moments-space.png" alt="朋友圈空间"/><br/><sub><b>朋友圈空间</b></sub></td>
  </tr>
  <tr>
    <td width="50%" align="center"><img src="screenshots/moments-notifications.png" alt="点赞与评论列表"/><br/><sub><b>点赞 / 评论列表</b></sub></td>
    <td width="50%"></td>
  </tr>
</table>

### 个人主页与设置

<table>
  <tr>
    <td width="33%" align="center"><img src="screenshots/profile-home.png" alt="我的主页"/><br/><sub><b>我的主页</b></sub></td>
    <td width="33%" align="center"><img src="screenshots/favorites.png" alt="我的收藏"/><br/><sub><b>我的收藏</b></sub></td>
    <td width="33%" align="center"><img src="screenshots/settings.png" alt="系统设置"/><br/><sub><b>系统设置</b></sub></td>
  </tr>
  <tr>
    <td width="33%" align="center"><img src="screenshots/settings-more.png" alt="更多设置"/><br/><sub><b>更多设置</b></sub></td>
    <td width="33%"></td>
    <td width="33%"></td>
  </tr>
</table>

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
  participant C as 发送方
  participant WS as im/ws
  participant K as Kafka
  participant MQ as task/mq
  participant DB as MongoDB
  participant R as 接收方
  C->>WS: chat.user
  WS->>K: msgChatTransfer
  K->>MQ: 消费
  MQ->>DB: 落库 chatlog
  MQ->>WS: push
  WS->>R: 消息帧
  R-->>WS: ACK
  WS-->>C: echo（真实 msgId）
```

### 已读回执

```mermaid
sequenceDiagram
  participant C as 阅读方
  participant WS as im/ws
  participant K as Kafka
  participant MQ as task/mq
  participant DB as MongoDB
  participant API as im/api
  participant S as 发送方
  C->>WS: chat.markChat
  WS->>K: msgReadTransfer
  K->>MQ: 消费
  MQ->>DB: 更新已读 bitmap + 时间
  MQ->>WS: 推送已读回执
  WS->>S: readRecords 更新
  S->>API: GET /v1/im/chatlog（已读详情）
  API->>DB: 查询已读 / 未读用户
```

### 消息撤回

```mermaid
sequenceDiagram
  participant C as 操作者
  participant API as im/api
  participant RPC as im/rpc
  participant DB as MongoDB
  participant K as Kafka
  participant MQ as task/mq
  participant WS as im/ws
  C->>API: POST /v1/im/chatlog/recall
  API->>RPC: RecallMsg
  RPC->>DB: 标记已撤回
  API->>K: msgRecallTransfer
  K->>MQ: 消费
  MQ->>WS: 推送撤回
  WS->>C: 撤回帧（在线客户端）
```

### 动态通知

```mermaid
sequenceDiagram
  participant C as 操作者
  participant API as trend/api
  participant RPC as trend/rpc
  participant DB as MySQL
  participant K as Kafka
  participant MQ as task/mq
  participant WS as im/ws
  participant R as 接收方
  C->>API: 提及 / 评论 / 回复 / 点赞
  API->>RPC: 执行业务逻辑
  RPC->>DB: 写入动态 + 通知
  RPC->>K: trendNotifyTransfer
  K->>MQ: 消费
  MQ->>WS: push.trend
  WS->>R: trend.notify（在线时）
```

## 音视频通话

独立的 `streaming` 服务提供 WebRTC 实时音视频。通话**控制**信令复用 im ws 通道（`push.call` → 客户端 `call.signal`）；**媒体协商**（offer/answer/ICE）走 streaming 服务自有的 WebSocket relay，而媒体数据本身**点对点直连、永不经过服务器**。1:1 为直接 P2P；群组采用**全连接 Mesh** 拓扑（两两直连，上限 4 人）。SFU 与 TURN 作为会议、直播、更强 NAT 穿透的预留扩展点。

### 单聊通话（1:1）

```mermaid
sequenceDiagram
  participant A as 主叫
  participant ST as streaming
  participant SR as social/rpc
  participant WS as im/ws
  participant B as 被叫
  A->>ST: call_invite（被叫、类型）
  ST->>SR: 校验好友关系
  ST-->>A: 会话已创建（callId）
  ST->>WS: push.call invite
  WS->>B: call.signal invite（振铃）
  B->>ST: call_accept
  ST->>A: call.signal accept
  A->>ST: offer / ICE
  ST->>B: relay offer / ICE
  B->>ST: answer / ICE
  ST->>A: relay answer / ICE
  Note over A,B: P2P 媒体直连，不经服务器
  A->>ST: call_end
  ST->>B: call.signal end
  Note over A,B: 向会话投递一条通话记录
```

### 群聊通话（Mesh）

```mermaid
sequenceDiagram
  participant I as 发起人
  participant ST as streaming
  participant SR as social/rpc
  participant WS as im/ws
  participant M as 成员
  I->>ST: group_invite（群、成员、类型）
  ST->>SR: 校验群成员
  ST-->>I: group_created（callId）
  ST->>WS: push.call group.invite（逐个成员）
  WS->>M: call.signal group.invite（振铃）
  ST->>M: group.state 广播（横幅 / 标识）
  M->>ST: group_join（callId）
  ST->>I: peer_joined（新成员 uid）
  Note over I,M: 由老成员向新人发 offer（避免 glare）
  I->>ST: offer（to = 成员）
  ST->>M: relay offer
  M->>ST: answer（to = 发起人）
  ST->>I: relay answer
  Note over I,M: 两两 P2P 直连（全连接 Mesh）
  M->>ST: group_leave
  ST->>I: peer_left
  ST->>WS: group.state 广播（更新 / 清除）
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

- 后端：Go 1.25、go-zero、zRPC、gRPC、goctl。
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
│   ├── screenshots
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

```

## 快速开始

### 一键部署（Docker Compose）

无需本地工具链，只要有 Docker，克隆后一条命令即可拉起整套服务（6 个微服务 + 中间件 + 前端）：

```bash
git clone https://github.com/iceymoss/go-hichat-api.git
cd go-hichat-api
docker compose up -d --build
```

随后访问 **http://localhost:2470** 。首次使用点「注册」即可——演示模式下验证码会自动填入输入框（无需真实短信），填昵称密码即可注册登录。

```bash
docker compose ps            # 查看各服务状态
docker compose logs -f web   # 跟踪某个服务日志
docker compose down          # 停止（保留数据）
docker compose down -v       # 停止并清空所有数据卷

# 一键清理：删数据卷 + 删本项目构建的镜像
docker compose down -v --remove-orphans && docker images 'hichat-*' -q | xargs -r docker rmi
```

架构、端口、清理卸载、服务器域名（反向代理 + HTTPS）部署、音视频（TURN）说明详见 [Docker 部署指南](../deploy/docker/README.zh-CN.md)。

### 填充演示数据（可选）

部署完想立刻有数据可点、可截图？内置的数据生成器会注册 **14 个中文示例用户**，并灌入好友、群组、单聊/群聊、动态及评论点赞——即 [产品截图](#产品截图) 里展示的那套数据。

用 Docker Compose 部署的（无需本机 Go 工具链），直接跑随仓库附带的一次性服务：

```bash
docker compose --profile mock run --rm mockdata
```

从源码运行的（已装 Go）：

```bash
go run ./scripts/mockdata              # 完整数据集
go run ./scripts/mockdata -trends-only # 只重灌动态/评论/点赞
```

随后访问 **http://localhost:2470**，用主角账号登录：

- 手机号 `13800138000`，密码 `hichat2024`。14 个账号同密码，手机号为 `13800138000`–`13800138013`。

> 仅在「全新/空库」上**跑一次**；重复运行会产生重复的好友申请与群。它只插入演示数据，不会删除任何东西。人设与内容脚本见 [`scripts/mockdata`](../scripts/mockdata)。

### 前置依赖

- Go 1.25 或更高版本（与 `go.mod` 的 `go` 指令一致）。
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
