# 公共通知通道（Common Notify Channel）

## 状态
- 创建日期: 2026-06-12
- 状态: 已完成（后端全链路 + 前端实时分发 + 通知中心面板 + 模型测试）
- 作者: iceymoss

## 实现进度（分支 feat-im-common-notify-channel）
- [x] im notifications 表 + model（MySQL，幂等唯一索引）
- [x] im RPC：Create/List/MarkRead/UnreadCount
- [x] Kafka 公共通道：mq.CommonNotify + 生产者 + 消费者（落库 + 在线推送）+ config/yaml/listen
- [x] ws push.notify → 前端 method=notify
- [x] social 6 类触点投递（friend/group apply/accept/reject，group.apply 扇出群主+管理员）
- [x] im API：v1/im/notifications 列表/未读数/标读
- [x] 前端：ws.on('notify') 实时红点 + Toast 气泡 + REST 客户端 + i18n + notificationVersion
- [x] 前端通知中心历史面板组件（铃铛+角标，挂在 chats 工具栏，读 REST + 标读 + 订阅 notificationVersion）
- [x] 去掉好友申请 10s 轮询（改挂载时拉一次 + 实时红点）
- [x] 测试：通知 model 幂等 / 多接收者 / 标读（真实 DB，AutoMigrate 校验三库 schema）

## 目标

建设一条**可扩展的公共 Kafka 通知通道**：业务方（social/im/trend...）在自己的 API/RPC 侧处理完动作后，往单一公共 topic 投递一条带 `notifyType` 的通知事件；`apps/task/mq` 消费后**先落库**（持久化到 im 服务的 `notifications` 表）**再在线推送**到 ws，前端按 `notifyType` 分发，渲染未读红点、Toast 气泡、通知中心列表、会话内系统消息。

MVP 用这条通道实现 social 的好友/群申请类实时通知；后续新增业务类型只需追加一个 `notifyType` + 前端一个分支，不动通道骨架。

## 非目标

- 不做跨 ws 节点路由（当前架构为单 ws 节点 / mq 消费者直连一个 ws 节点；多节点路由是已知遗留限制，本期沿用，单列「待定事项」）。
- 不替换/移除现有 `relationChangeTransfer`（被踢/解散）与 `trendNotifyTransfer`（动态通知）通道，二者保持原样；本通道是新增的并行公共通道。
- MVP 不做：被邀请加群、退群/被踢通知（relation 通道已部分覆盖）、通知聚合/合并、通知偏好开关（免打扰）、多端已读同步。
- 不改 `.env`，不做 schema 破坏性变更（仅 ADD COLUMN / 新建表）。

## 用户故事

- 作为**被申请人**，当有人申请加我为好友时，我想立即看到红点+1 和一条气泡提示，以便第一时间处理。
- 作为**好友申请发起人**，当对方通过或拒绝我时，我想收到结果通知，以便知道进展。
- 作为**群主/管理员**，当有人申请进我管理的群时，我想立即收到通知，以便审批。
- 作为**入群申请人**，当我的申请被通过/拒绝时，我想收到结果通知。
- 作为**离线用户**，上线后我想能拉到离线期间错过的通知（未读列表），不丢消息。
- 作为**开发者**，我想以后新增一类通知时只加一个 `notifyType` 和前端一个 case，不重复造通道。

## 核心流程（以「申请好友」为例）

1. A 在前端发起加 B 好友 → `social/api` → `social/rpc` `FriendPutIn`，写 `friend_requests`、失效气泡缓存（现有逻辑）。
2. RPC logic 成功后，**直接 Push** 一条 `mq.CommonNotify{ NotifyType: "friend.apply", ReceiverId: B, ActorId: A, BizId: <requestId>, ... }` 到公共 topic `commonNotifyTransfer`（参考撤回的直接 Push 范式）。
3. `apps/task/mq` 的 `CommonNotifyTransfer` 消费者：
   a. 幂等校验（`bizType + bizId + receiverId + notifyType` 唯一）→ 调 **im-rpc `CreateNotification`** 落库到 im 的 `notifications` 表（im 拥有该表）。
   b. 调 `WsClient.Send(Message{ Method: "push.notify", FormId: SYSTEM_ROOT_UID, Data: ws.Notify{...} })` 推送给 `ReceiverId`；ws handler 把 method 改写为 `notify` 下发前端。用户离线则静默丢弃（已落库，上线可拉）。
4. 前端 ws-client 收到 `method=notify` → 按 `notifyType` 分发：
   - 未读红点（`friendRequestUnreadCount` / `groupAppUnreadCount`）实时 +1，去掉现有 10s 轮询延迟。
   - Toast 气泡（sonner）+ 声音/振动。
   - 通知中心列表插入新项（未读）。
   - （好友通过等场景）必要时在会话内插入系统消息。
5. 用户打开通知中心 → `im/api` `ListNotifications` 拉取历史/未读；点击/已读 → `MarkNotificationRead`。

## 各 notifyType 的接收者与触点

| notifyType | 触发点（social/rpc logic） | 接收者 | 落库内容要点 |
|---|---|---|---|
| `friend.apply` | `friendputinlogic.go`（InvalidateCountCache 后） | 被申请人 | actor=申请人, bizId=friend_request.id, 申请留言 |
| `friend.accept` | `friendputinhandlelogic.go`（emit 成功后, result=1） | 申请人 | actor=处理人 |
| `friend.reject` | `friendputinhandlelogic.go`（Commit 后, result=2） | 申请人 | actor=处理人 |
| `group.apply` | `groupputinlogic.go`（createGroupReq, 需审核路径） | 群主/管理员（多接收者） | actor=申请人, bizId=group_request, groupId |
| `group.accept` | `groupputinhandlelogic.go`（Commit 后, result=1） | 申请人 | actor=审批人, groupId |
| `group.reject` | `groupputinhandlelogic.go`（Commit 后, result=2） | 申请人 | actor=审批人, groupId |

> `group.apply` 接收者是「群主+管理员」多人，消费者需扇出多条 ws 推送 + 多行落库（每个接收者一行）。

## 异常处理

| 场景 | 处理方式 |
|------|---------|
| Kafka 重复投递 / 消费重试 | 消费幂等：`notifications` 表对 `(receiver_id, notify_type, biz_id)` 建唯一索引，落库用 ON CONFLICT/先查后插，重复直接跳过 |
| 接收者离线 | ws 推送静默丢弃；通知已落库，上线后走 `ListNotifications` 拉未读 |
| im-rpc 落库失败 | 可重试错误 → 抛出让 Kafka 重投；不可重试（如参数非法）→ 打日志 + 不阻塞，按 mq-task 规则不无脑 panic |
| ws 推送失败（节点不可达） | 不影响落库结果；记日志。不重试 ws（在线态本就尽力而为） |
| 业务事务提交后、Push 前进程崩溃 | 极小概率丢通知（直接 Push 范式的固有取舍，已与用户确认接受）；如需零丢失再升级为 outbox |
| 申请人=接收者（自己加自己等异常数据） | 投递前过滤，actor==receiver 不发通知 |
| 多接收者部分失败（group.apply） | 逐个接收者独立处理，单个失败不影响其他；失败项记日志 |

## 技术设计

### 数据模型（im 服务 / `apps/im/immodels`）

> 注意：im-rpc 原为 Mongo-only（chatlog/conversations 均在 MongoDB），本表落 MySQL 需给 im-rpc 新增 MySQL 连接（config + svc + yaml）+ 在 `pkg/db/objects` 加表并在 `deploy/main.go` 注册项目级迁移。已与用户确认用 MySQL。

新建表 `notifications`（GORM 处理主键，TEXT 存 JSON，三库兼容）：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | uint64 | 主键（GORM 自增，勿用 AUTO_INCREMENT 字面量） |
| `receiver_id` | varchar | 接收者 uid（索引） |
| `notify_type` | varchar | 业务类型，如 `friend.apply` |
| `actor_id` | varchar | 触发者 uid |
| `biz_id` | varchar | 业务关联 id（friend_request.id / group_request 等），幂等键 |
| `group_id` | varchar | 群相关通知的群 id，可空 |
| `title` | varchar | 冗余标题（可空，前端也可用 type+i18n 渲染） |
| `content` | varchar | 冗余正文/摘要 |
| `payload` | TEXT | 扩展字段 JSON（用 common.Marshal），向前兼容 |
| `is_read` | int | 0 未读 / 1 已读（用 commonTrueVal/commonFalseVal） |
| `created_at` | datetime | 创建时间 |
| `read_at` | datetime | 已读时间，可空 |

唯一索引：`uniq_receiver_type_biz (receiver_id, notify_type, biz_id)`。普通索引：`idx_receiver_read (receiver_id, is_read)`。

> schema 变更需用户二次确认（database-model 规则）。建表/索引迁移须三库（SQLite/MySQL/PostgreSQL）兼容。

### Kafka 消息体（`apps/task/mq/mq/mq.go` 追加）

```go
type CommonNotify struct {
    NotifyType string `json:"notifyType"`          // friend.apply / group.accept ...
    ReceiverId string `json:"receiverId"`          // 单接收者
    ActorId    string `json:"actorId"`             // 触发者
    BizId      string `json:"bizId,omitempty"`     // 幂等键 + 业务跳转
    GroupId    string `json:"groupId,omitempty"`
    Title      string `json:"title,omitempty"`
    Content    string `json:"content,omitempty"`
    Payload    string `json:"payload,omitempty"`   // 扩展 JSON
    CreateTime int64  `json:"createTime"`
}
```

> 多接收者（group.apply）在**生产端**展开为多条单接收者消息，保持消费端单接收者模型简单。

### ws 数据帧（`apps/im/ws/ws/ws.go` 追加）

```go
type Notify struct {
    NotifyType string `json:"notifyType"`
    ReceiverId string `json:"receiverId"`
    ActorId    string `json:"actorId"`
    BizId      string `json:"bizId,omitempty"`
    GroupId    string `json:"groupId,omitempty"`
    Title      string `json:"title,omitempty"`
    Content    string `json:"content,omitempty"`
    Payload    string `json:"payload,omitempty"`
    CreateTime int64  `json:"createTime"`
}
```

### 生产者 client（`apps/task/mq/mq_client/common_notify_client.go`）

参照 `trend_notify_client.go` / `msg_recall_client.go`：`kq.NewPusher(addr, "commonNotifyTransfer")`，暴露 `Push(ctx, *mq.CommonNotify)`。social/rpc 在 servicecontext 注入该 client。

### 消费者（`apps/task/mq/internal/handler/msg_transfer/common_notify_transfer.go`）

```go
type CommonNotifyTransfer struct{ *BaseChatTransfer }
func (m *CommonNotifyTransfer) Consume(ctx, key, value string) error {
    var in mq.CommonNotify
    common.Unmarshal([]byte(value), &in)
    // 1. 幂等落库（调 im-rpc CreateNotification 或 im notification model）
    // 2. ws 在线推送
    return m.svcCtx.WsClient.Send(websocket.Message{
        FrameType: websocket.FrameNoAck,
        Method:    "push.notify",
        FormId:    constants.SYSTEM_ROOT_UID,
        Data:      &ws.Notify{...},
    })
}
```

注册：`internal/config/config.go` 加 `CommonNotifyTransfer kq.KqConf` → `etc/mq-dev.yaml`/`mq-sample.yaml` 加 topic 块 → `internal/handler/listen.go` 加一行 `kq.MustNewQueue(l.svc.Config.CommonNotifyTransfer, commonNotifyConsumeHandle)`。

### 落库归属（im 服务）

- 表与读写 API 归 im 服务（im 本就管会话/推送/ws，通知中心最贴近）。
- 落库由消费者经 **im-rpc `CreateNotification`** 完成（task/mq 不直写 im 库表，遵守微服务边界）；需在 im.proto 追加 `CreateNotification` / `ListNotifications` / `MarkNotificationRead`，并在 task/mq servicecontext 注入 im-rpc client。
- 备选：消费者直接用 im notification model 写库（更少 RPC，但弱化边界）。**推荐走 im-rpc**，详见待定事项。

### im 读取/标读接口（`apps/im/api/im.api`）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `v1/im/notifications` | 拉取通知列表（分页 + `unreadOnly` 可选） |
| GET | `v1/im/notifications/unreadCount` | 未读总数（可分 type） |
| POST | `v1/im/notifications/read` | 标记已读（单条/批量/全部） |

`.api` 改完跑 `goctl api go`；`.proto` 改完跑 `goctl rpc protoc`；可选字段加 `optional` + 指针 + omitempty。

### ws 路由（`apps/im/ws/internal/handler/routes.go` + `push/notify.go`）

新增 `{ Method: "push.notify", Handler: push.Notify(svc) }`，handler 内把 `sendMsg.Method = "notify"` 下发前端（参考 `trend_notify.go` 把 `push.trend`→`trend.notify`）。

### 前端（web/）

- `web/src/lib/ws-client.ts`：无需改骨架，已有 `ws.on(method, handler)`。
- `web/src/store/chat-store.ts`（`initWs`）：注册 `ws.on('notify', handler)`，按 `notifyType` switch：
  - `friend.apply` → `im-store.friendRequestUnreadCount++` + toast + 通知中心插入。
  - `group.apply` → `groupAppUnreadCount++` + toast。
  - `friend.accept/reject`、`group.accept/reject` → toast + 通知中心 + 红点。
- 通知中心组件：新增（或复用 `FriendRequestList`）一个可下拉/弹层的通知列表，接 `ListNotifications` + `MarkNotificationRead`。
- i18n：在 `web/src/i18n/locales/{zh-CN,en}.json` 新增 `notify.friend.apply` 等 key（当前 social 通知文案缺失，全部硬编码，需补 t() 包裹）。

### 实现步骤（每步可独立 commit）

1. [ ] **数据模型**：im 新建 `notifications` 表 + model（goctl model，三库兼容，唯一/普通索引）— 需用户确认 schema。
2. [ ] **im RPC**：im.proto 加 `CreateNotification/ListNotifications/MarkNotificationRead` + logic（落库幂等、列表分页、标读）。
3. [ ] **Kafka 通道骨架**：`mq.CommonNotify` 消息体 + `common_notify_client.go` 生产者 + config/yaml/listen 注册 + `common_notify_transfer.go` 消费者（落库 via im-rpc + ws 推送）。
4. [ ] **ws 下行**：`ws.Notify` 帧 + `push.notify` 路由/handler（method 改写为 `notify`）。
5. [ ] **social 投递**：social/rpc 注入 common-notify client；在 friend/group 各 logic 触点 Push 通知（含 group.apply 多接收者展开、actor==receiver 过滤）。
6. [ ] **im 读取 API**：im.api 加 notifications 列表/未读数/标读接口 + handler/logic。
7. [ ] **前端 ws 分发**：`ws.on('notify')` + 按 type 更新红点/toast/通知中心。
8. [ ] **前端通知中心 + i18n**：通知列表组件 + zh-CN/en 文案，去掉好友申请 10s 轮询延迟（改实时）。

### 参考的现有模式

- `apps/im/api/internal/logic/recallmsglogic.go` — 业务处理后直接 `Push` 到 Kafka 的范式（本期投递照此）。
- `apps/task/mq/mq_client/trend_notify_client.go` — 生产者 client 封装模板。
- `apps/task/mq/internal/handler/msg_transfer/trend_notify_transfer.go` — 单接收者消费→ws 推送模板。
- `apps/im/ws/internal/handler/push/trend_notify.go` — `push.trend`→`trend.notify` method 改写模板（本期 `push.notify`→`notify`）。
- `apps/task/mq/internal/handler/listen.go` — 消费者注册点。
- `apps/social/rpc/internal/logic/friendputinhandlelogic.go` / `groupputinhandlelogic.go` — 投递触点。
- `web/src/store/chat-store.ts`（`ws.on('relation.changed')`）— 前端 ws method 分发模板。

## 测试计划

- [ ] 消费者幂等：同一条 CommonNotify 重复消费只落一行、只推一次（table-driven，覆盖三库）。
- [ ] 多接收者：group.apply 展开后每个群主/管理员各一行 + 各一次推送。
- [ ] 离线落库：接收者离线时 ws 推送失败不影响落库；ListNotifications 能拉到未读。
- [ ] 标读：MarkNotificationRead 单条/批量/全部，unreadCount 正确递减。
- [ ] actor==receiver 过滤、空 receiver 等边界。
- [ ] 三库（SQLite/MySQL/PostgreSQL）下建表与读写通过；测试后清理数据。
- [ ] 前端：模拟收到各 notifyType 帧，红点/toast/通知中心渲染正确。

## 待定事项

- 落库由「消费者调 im-rpc」还是「消费者直写 im model」？规范推荐走 im-rpc（守边界），但多一次 RPC；请确认。
- 跨 ws 多节点路由缺失（现为单节点）：在线推送在多节点部署下会漏推（落库仍在，上线可拉）。本期是否接受？是否排期单独解决？
- 通知中心 UI：复用 `FriendRequestList` 还是新建独立通知列表组件？
- `group.apply` 接收者范围：仅群主，还是群主+全部管理员？
- 通知是否需要「免打扰/通知偏好」开关（沿用 `friends.notify_enabled`？）— 倾向后续迭代。
- 通知保留策略（TTL / 数量上限 / 清理任务）？

## MVP 范围

**包含**：
- 公共 Kafka 通道（topic `commonNotifyTransfer` + 消息体 + 生产者 + 消费者 + 注册）。
- im `notifications` 表 + 落库（幂等） + List/UnreadCount/MarkRead 接口。
- ws `push.notify`→`notify` 下行通道。
- social 6 类触点投递：`friend.apply/accept/reject`、`group.apply/accept/reject`。
- 前端 4 形态：实时红点、Toast、通知中心列表、（必要时）会话内系统消息 + i18n。

**不包含（后续迭代）**：被邀请加群、退群/被踢通知整合、通知聚合、免打扰偏好、多端已读同步、跨 ws 多节点路由、通知清理任务。
