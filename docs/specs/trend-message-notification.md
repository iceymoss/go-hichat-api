# 动态消息通知系统（Trend Message Notification）

## 状态
- 创建日期: 2026-06-10
- 状态: 已实现 + 后端端到端验证通过（建表已执行；浏览器实时红点待双账号手测）
- 关联模块: `apps/trend`（api + rpc + models）、`apps/task/mq`、`apps/im/ws`、`web/`

## 目标
为动态（朋友圈）模块建立**统一的消息通知中心**：点赞、评论、回复评论、发布动态 @、评论中 @ 五类互动事件，统一写入一张专门的 `trend_message` 表，并通过「Kafka → task/mq 消费 → WebSocket 推送」链路实时下发给被通知用户；前端提供消息中心 UI、未读红点、实时刷新。

替换现有**轮询式**未读机制（`GetUnreadReplies` / `GetUnreadLikes` + 扫 `trend_agree.is_read` / `trend_discuss.read` 标志），新表为唯一数据来源。

## 非目标（本期不做）
- 点赞聚合（"X 等 N 人赞了你的动态"）—— 本期每次点赞一条独立消息。
- 消息分类型 tab、消息免打扰、@ 搜索联想。
- 旧 `trend_agree.is_read` / `trend_discuss.read` 字段的物理删除（保留列，只是前端不再使用旧接口）。
- 跨服务（im/social）的统一消息中心 —— 本期只覆盖 trend 域。

## 用户故事
- 作为动态作者，当有人点赞 / 评论我的动态时，我希望**实时**收到通知并在消息中心看到。
- 作为评论者，当有人回复我的评论时，我希望收到回复通知。
- 作为用户，当有人在发动态或评论里 @我 时，我希望收到 @通知。
- 作为用户，打开 App / 刷新页面时，我希望看到动态消息的**未读总数**（及按类型明细），点进消息中心后未读清零。

## 关键设计决策（已与用户确认）
| 决策点 | 结论 |
|--------|------|
| 旧轮询未读机制 | 新 `trend_message` 表为唯一来源，旧接口废弃（保留代码不删，前端不再调用） |
| 推送链路 | 完全复用撤回模式：api logic 投 Kafka 专用 topic → task/mq 消费 → WsClient 推送 → 前端按 ws method 路由 |
| 点赞呈现 | 每次点赞一条独立消息（不聚合） |
| 未读数 | 一个接口返回 `total` 总未读 + 按类型明细 `{like, comment, reply, at_trend, at_comment}` |
| 自我操作 | 给自己动态点赞/评论、@自己 **不生成消息** |
| 级联删除 | 取消点赞 / 删评论 / 删动态时，**级联软删除**（`state=0`）对应消息，未读数随之减少 |
| 标记已读 | 进入消息列表**全部标已读**（一键清零） |

## 核心流程

### A. 生成 + 实时推送（写路径，复用撤回模式）
以「点赞」为例（评论 / 回复 / @ 同构）：

```
前端点赞
  └─ trend-api: ToggleLikeLogic
        └─ 调 trend-rpc: ToggleLike
              ├─ 写 trend_agree（既有逻辑）
              └─ 若为「点赞」且非自我操作 → 写 trend_message(type=1)，返回创建的消息信息
        └─ rpc 成功且本次真正新增消息时
              └─ 向 trendNotifyTransfer topic 投递 TrendNotifyTransfer 事件（状态变更与推送副作用分离）
                    └─ task/mq: TrendNotifyTransfer 消费者
                          └─ 按 receiverId 单推：WsClient.Send(Method="push.trend")
                                └─ im/ws: push.TrendNotify handler
                                      └─ GetConn(receiverId) → Send(method="trend.notify", data=通知体)
                                            └─ 前端 ws.on('trend.notify') → 未读+1、列表插入、红点刷新
```

> 与撤回一致：**消息落库在 rpc（数据所有权方），Kafka 推送在 api logic**。推送失败不回滚落库（DB 已写），仅记录日志，前端可通过下次拉取列表补偿。

### B. 拉取未读数（刷新时）
```
前端刷新 → GET /v1/trend/message/unread → trend-rpc 聚合 count(is_read=0,state=1) group by type
        → 返回 { total, like, comment, reply, at_trend, at_comment }
```

### C. 消息列表 + 标记已读
```
进入消息中心 → GET /v1/trend/message/list（分页）→ 关联 user 信息返回
            → PUT /v1/trend/message/mark-read → 该用户全部未读置已读 → 红点清零
```

### D. 级联软删除
```
取消点赞   → ToggleLike(off)  → 软删 trend_message where type=1 and trend_id and actor=自己
删除评论   → DeleteDiscuss     → 软删 该 comment_id 相关的 comment/reply/at_comment 消息
删除动态   → DeleteTrend       → 软删 该 trend_id 相关的全部消息
```

## 异常处理
| 场景 | 处理方式 |
|------|---------|
| 自我操作（点赞/评论自己、@自己） | rpc 写入前过滤 `actor == receiver` 跳过 |
| Kafka 推送失败 | 不回滚落库；`zLog.Error` 记录；前端下次拉列表补偿（最终一致） |
| 接收者离线 | `push.single` GetConn 为空直接 return（消息已落库，上线拉列表可见） |
| 重复消费（Kafka 重投） | 推送幂等（NoAck 控制帧，重复推送无副作用）；落库已在 rpc 完成，消费者**不落库** |
| 同一动态重复点赞 | `trend_agree` 已有唯一键 `uniq_user_trend`；仅状态翻转为有效时才生成消息 |
| @ 列表含非法/重复 uid | 写入前去重 + 过滤自己 + 过滤空值 |
| 删除已读消息后又恢复 | 不支持恢复，软删即终态 |

## 技术设计

### 数据模型（新增表，需 DBA / 用户确认后执行）

> 遵循 trend 模块**既有约定**（`goctl model mysql` + sqlx + cache，MySQL DDL、`AUTO_INCREMENT`、`utf8mb4`），与 `trend_agree` / `trend_discuss` 一致；**不**走 `model/main.go` 的 GORM 三库迁移体系。

`deploy/sql/trend_message.sql`：
```sql
CREATE TABLE `trend_message` (
    `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `receiver_id`       INT UNSIGNED    NOT NULL COMMENT '接收者(被通知人)uid',
    `actor_id`          INT UNSIGNED    NOT NULL COMMENT '触发者(操作人)uid',
    `type`              TINYINT UNSIGNED NOT NULL COMMENT '类型 1点赞 2评论 3回复评论 4发动态@ 5评论@',
    `trend_id`          INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '关联动态ID',
    `comment_id`        BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联评论ID(评论/回复/评论@有值)',
    `parent_comment_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '被回复的评论ID(回复评论时)',
    `content`           VARCHAR(500)    NOT NULL DEFAULT '' COMMENT '内容快照(评论/回复正文摘要)',
    `is_read`           TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0未读 1已读',
    `state`             TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '0已删除(级联软删) 1正常',
    `create_time`       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time`       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_receiver_state_read` (`receiver_id`, `state`, `is_read`),   -- 未读数聚合 + 列表
    KEY `idx_receiver_id` (`receiver_id`, `id`),                          -- 列表分页(按id倒序)
    KEY `idx_cascade_trend` (`trend_id`, `state`),                       -- 删动态级联
    KEY `idx_cascade_comment` (`comment_id`, `state`),                   -- 删评论级联
    KEY `idx_cancel_like` (`type`, `trend_id`, `actor_id`, `state`)      -- 取消点赞级联
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='动态消息通知表';
```

消息类型常量（新增 `pkg/constants/trend.go` 或 trend models `vars.go`）：
```
TrendMsgLike      = 1  // 点赞
TrendMsgComment   = 2  // 评论动态
TrendMsgReply     = 3  // 回复评论
TrendMsgAtTrend   = 4  // 发动态 @
TrendMsgAtComment = 5  // 评论 @
```

生成 model：`goctl model mysql ddl -src deploy/sql/trend_message.sql -dir apps/trend/models -c`
自定义方法（写 `trendmessagemodel.go`，**不**改 `_gen`）：
- `ListByReceiver(ctx, receiverId, lastId, limit)` —— 列表分页（state=1 倒序）
- `CountUnreadByType(ctx, receiverId)` —— 返回 `map[type]int64`，供未读聚合
- `MarkAllRead(ctx, receiverId)` —— 全部未读置已读
- `SoftDeleteByTrend(ctx, trendId)` / `SoftDeleteByComment(ctx, commentId)` / `SoftDeleteLike(ctx, trendId, actorId)` —— 级联软删

### RPC 接口（`apps/trend/rpc/trend.proto`，方法只追加、字段号不复用）
| 方法 | 说明 |
|------|------|
| `ListTrendMessages(ListTrendMessagesReq) returns (ListTrendMessagesResp)` | 消息列表分页 |
| `GetTrendMessageUnread(GetTrendMessageUnreadReq) returns (GetTrendMessageUnreadResp)` | 未读总数 + 按类型明细 |
| `MarkTrendMessagesRead(MarkTrendMessagesReadReq) returns (MarkTrendMessagesReadResp)` | 全部标已读 |

> 生成消息**不新增 rpc**，而是在既有 `ToggleLike` / `CreateRootDiscuss` / `CreateChildDiscuss` / `CreateTrend` 的 rpc resp 里**追加返回**「本次新生成的消息列表」（receiver/actor/type/trendId/commentId/content/messageId），供 api 层投 Kafka。级联软删在既有 `ToggleLike(off)` / `DeleteDiscuss` / `DeleteTrend` rpc logic 内联完成。

`GetTrendMessageUnreadResp` 字段：`total`, `like`, `comment`, `reply`, `at_trend`, `at_comment`（零值保留语义）。

### API 接口（`apps/trend/api/trend.api`，新增 `group: message`，前缀 `v1/trend`）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/v1/trend/message/list` | 消息列表（关联 actor 用户信息），分页 `last_id` |
| GET | `/v1/trend/message/unread` | 未读总数 + 类型明细 |
| PUT | `/v1/trend/message/mark-read` | 全部标已读 |

> 列表 logic 参考 `getunreadreplieslogic.go`：批量取 `actor_id` → 调 `user-rpc FindUser` 填充昵称/头像，再组装返回。

改 `.api` / `.proto` 后必须 goctl 重新生成（只填业务逻辑，不手写骨架）。

### 推送链路（Kafka + task/mq + ws，镜像撤回实现）

**1) Kafka 事件结构** `apps/task/mq/mq/mq.go` 新增：
```go
type TrendNotifyTransfer struct {
    ReceiverId string `json:"receiverId"` // 被通知人 uid
    ActorId    string `json:"actorId"`    // 操作人 uid
    MsgType    int32  `json:"msgType"`    // 1点赞 2评论 3回复 4发动态@ 5评论@
    TrendId    uint64 `json:"trendId"`
    CommentId  uint64 `json:"commentId,omitempty"`
    MessageId  uint64 `json:"messageId"`  // trend_message.id
    Content    string `json:"content,omitempty"`
    CreateTime int64  `json:"createTime"`
}
```

**2) 生产者** `apps/task/mq/mq_client/trend_notify_client.go`（镜像 `msg_recall_client.go`）：
- `TrendNotifyTransferClient` 接口 + `Push(*mq.TrendNotifyTransfer)`
- 在 **trend-api** 接线：
  - `config.go` 加 `TrendNotifyTransfer struct { Addrs []string; Topic string }`
  - `servicecontext.go` 加 `TrendNotifyTransferClient: mq_client.NewTrendNotifyTransferClient(c.TrendNotifyTransfer.Addrs, c.TrendNotifyTransfer.Topic)`
  - `etc/*-sample.yaml` / `*-dev.yaml` 配 brokers + topic `trendNotifyTransfer`

**3) 消费者** `apps/task/mq/internal/handler/msg_transfer/trend_notify_transfer.go`（镜像 `msg_recall_transfer.go`）：
- `Consume`：反序列化 → 组装 `ws.TrendNotify` 推送体 → `WsClient.Send(Method="push.trend", Data=...)`
- **不落库**（已在 trend-rpc 落库）
- `apps/task/mq/internal/handler/listen.go` 注册：`kq.MustNewQueue(l.svc.Config.TrendNotifyTransfer, trendNotifyConsumeHandle)`
- `apps/task/mq/internal/config/config.go` 加 `TrendNotifyTransfer kq.KqConf` + yaml 配置

**4) WS 推送帧（与聊天帧解耦）** —— 撤回复用了 `ws.Chat`（因撤回属于会话事件）；动态通知**不属于任何会话**，故新增独立帧与路由：
- `apps/im/ws/ws/` 新增 `TrendNotify` struct（receiverId/actorId/msgType/trendId/commentId/messageId/content/createTime）
- `apps/im/ws/internal/handler/routes.go` 新增路由 `Method:"push.trend"` → `push.TrendNotify(svc)`
- `apps/im/ws/internal/handler/push/trend_notify.go`：`GetConn(receiverId)` → 构造客户端帧并**显式设置 `Method="trend.notify"`**（让前端用独立 handler 接收，不污染聊天的 `ws.on('')`）。需要 `NewMessage` 后 `msg.Method = "trend.notify"`，或新增 `NewMessageWithMethod`。

### 前端（`web/`）
- `web/src/lib/trend-api.ts`：
  - 新增 `listTrendMessages(token, lastId)` / `getTrendMessageUnread(token)` / `markTrendMessagesRead(token)`
  - 旧 `getUnread` / `getUnreadLikes` / `unreadToNotifications` 标注 `@deprecated`，前端不再调用
- `web/src/lib/ws-client.ts`：`ContentType` 无需改；新增对 `method="trend.notify"` 的支持（已有按 method 路由能力）
- WS 接收：在动态相关 store（参考 `chat-store.ts` 的 `ws.on('')`）注册 `ws.on('trend.notify', handler)` → 未读 +1、列表头插、红点刷新
- 消息中心 UI：在 `MomentsFeed.tsx` / 动态入口加「消息」入口 + 未读红点；新增消息列表面板（按 `type` 渲染不同文案：赞了你 / 评论了你 / 回复了你 / @了你），点击跳转对应动态详情 `TrendDetailPanel.tsx`
- i18n：`web/src/i18n/locales/{zh,en}.json` 加 `trend.message.*` key（用 `t()` 包裹，不硬编码中文）
- UI 用 Semi Design 组件 + Tailwind，包管理用 bun

## 实现步骤（每步可独立 commit）

### 后端
1. [ ] `deploy/sql/trend_message.sql` + `goctl model` 生成 + 自定义方法（`ListByReceiver`/`CountUnreadByType`/`MarkAllRead`/级联软删）
2. [ ] 消息类型常量（`pkg/constants` 或 models `vars.go`）
3. [ ] `.proto` 加 3 个查询/标记方法 + 在既有 ToggleLike/CreateRootDiscuss/CreateChildDiscuss/CreateTrend resp 追加「新生成消息」字段；goctl 生成
4. [ ] trend-rpc logic：四类事件写 `trend_message`（过滤自我操作、@ 去重）+ 三个查询 logic + 级联软删（unlike/删评论/删动态）
5. [ ] `.api` 加 message 组 3 接口；goctl 生成；api logic（列表关联 user 信息）
6. [ ] Kafka：`mq.TrendNotifyTransfer` 结构 + 生产者 client + trend-api 接线（config/svc/yaml）
7. [ ] trend-api logic（toggleLike/createDiscuss/createTrend）rpc 成功后投 Kafka
8. [ ] task/mq：消费者 + listen 注册 + config/yaml
9. [ ] im/ws：`ws.TrendNotify` 结构 + `push.trend` 路由 + `trend_notify.go` handler（显式 method）

### 前端
10. [ ] `trend-api.ts` 新增 3 个接口封装，旧接口标 deprecated
11. [ ] `ws.on('trend.notify')` 接收 + 动态消息 store（未读数 + 列表）
12. [ ] 消息中心 UI（入口 + 红点 + 列表 + 跳转）+ i18n

## 参考的现有模式
| 文件 | 参考了什么 |
|------|-----------|
| `apps/im/api/internal/logic/recallmsglogic.go` | api logic 调 rpc 后投 Kafka、推送失败不回滚 |
| `apps/task/mq/mq_client/msg_recall_client.go` | Kafka 生产者 client 结构 |
| `apps/task/mq/internal/handler/msg_transfer/msg_recall_transfer.go` | 消费者翻译事件 → WsClient.Send，不落库 |
| `apps/task/mq/internal/handler/msg_transfer/base_msg_chat_transfer.go` | `single()` 按 RecvId 单推 |
| `apps/im/ws/internal/handler/push/pusher.go` | ws push handler 按 uid GetConn → Send |
| `apps/trend/api/internal/logic/comment/getunreadreplieslogic.go` | 批量取 uid → 调 user-rpc 填充用户信息 |
| `apps/im/api/internal/svc/servicecontext.go` | 生产者 client 接线方式 |
| `web/src/lib/chat-store.ts` (`ws.on('')`) | 前端 ws 事件接收 + store 更新 |

## 测试计划
- [ ] `trend_message` model：table-driven，覆盖 ListByReceiver / CountUnreadByType / MarkAllRead / 三种级联软删（不 mock DB，测试后清理数据）
- [ ] 写路径：点赞/评论/回复/发动态@/评论@ 各生成正确类型消息；自我操作不生成；@ 去重
- [ ] 级联：取消点赞 / 删评论 / 删动态后对应消息 state=0、未读数减少
- [ ] 未读聚合：total = 各类型之和（仅 state=1 且 is_read=0）
- [ ] 标记已读：调用后 total 归零
- [ ] 幂等：Kafka 事件重复消费不产生重复推送副作用
- [ ] 手动验证：双账号（含测试账号 17585710998）实时收到推送 + 红点 + 列表

## 待定事项
- [ ] **数据库 schema 变更需用户最终确认**（建表语句、索引）后才执行（遵循 `database-model.md`）。
- [ ] `ListTrendMessages` 分页游标用 `last_id`（自增 id 倒序）还是 `create_time`？建议 `last_id`，与现有 unread 分页一致。
- [ ] 「评论同时 @ 作者」是否去重为一条（避免作者既收 comment 又收 at_comment）？建议：被回复人/作者命中 comment/reply 时，从 @列表剔除，避免重复通知。
- [ ] ws 是否新增 `NewMessageWithMethod` 还是构造后赋值 `msg.Method`？建议后者，改动最小。

## MVP 范围
**包含**：
- 后端 `trend_message` 表 + 列表/未读数/标记已读接口
- 五类事件（点赞/评论/回复/发动态@/评论@）生成消息 + Kafka 实时推送
- 前端消息中心 UI + 未读红点 + ws 实时刷新

**不包含**（后续迭代）：点赞聚合、分类型 tab、消息免打扰、@ 搜索联想、跨服务统一消息中心。
