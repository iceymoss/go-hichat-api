# 消息撤回 + 群 @ 功能

## 状态
- 创建日期: 2026-06-03
- 状态: 已实现（feat-im-recall-mention 分支）


## 目标
为会话模块补齐两项 IM 基础能力：
1. **消息撤回**：发送者本人（2 分钟内）和群管理员可撤回消息，撤回后所有端在原位展示"某某撤回了一条消息"，原内容不再下发。
2. **群 @**：群聊支持 @单个/多个群成员、管理员 @所有人，被 @的人在会话列表得到"[有人@你]"角标提醒。

## 非目标
- 撤回后的"重新编辑"（把原文回填到输入框）—— 后续迭代。
- 私聊 @（@只在群聊生效）。
- @ 消息的独立"提及列表"聚合页 / 全局未读 @ 汇总。
- 撤回的撤销 / 审计后台查询界面（DB 保留原文，但不做查询 UI）。
- 系统消息体系的通用化（本期仅复用一个撤回控制帧 + 前端占位渲染，不引入通用 system 消息表）。

## 用户故事

### 撤回
- 作为**消息发送者**，我想要在发错消息后 2 分钟内撤回它，以便对方/群里看不到我发错的内容。
- 作为**群管理员/群主**，我想要撤回群内任何成员的不当消息，以便维护群秩序（不受 2 分钟限制）。
- 作为**会话里的其他人**，当一条消息被撤回时，我希望气泡原位变成"某某撤回了一条消息"，并且无论我在线还是离线后再进会话，看到的都是这个状态。

### 群 @
- 作为**群成员**，我想在输入框打 `@` 时弹出群成员列表并选择一个或多个人，以便定向提醒他们。
- 作为**群管理员**，我想要 `@所有人`，以便重要通知触达全员（普通成员不可用此项）。
- 作为**被 @的人**，我希望在会话列表看到"[有人@你]"角标，进入会话后角标消失。

## 核心流程

### A. 消息撤回（happy path）
1. 用户在消息气泡上右键 / 长按 → 上下文菜单点击"撤回"（前端 `MessageContextMenu` 已有该入口）。
2. 前端调用 REST `POST /v1/im/chatlog/recall`，带 `conversationId / msgId / chatType`。
3. im-api handler 仅做参数绑定 → im-api logic 仅做编排，调 im-rpc `RecallMsg`（严格 im-api → im-rpc，业务不下沉到 api）。
4. **im-rpc `RecallMsg` 承载全部撤回业务**：load ChatLog → 校验权限与时间窗（见"权限校验"，群管理员角色经 social-rpc 查）→ **条件更新**（仅当 `Status=正常`）MongoDB：`Status=已撤回 / RecalledBy / RecalledAt`，**原 `MsgContent` 保留不清空**（审计兜底）→ 返回。
5. im-rpc 返回成功后，im-api logic 向已有**通用 topic** `MsgChatTransfer` 投递一条控制帧 `MsgType=ContentRecall`（携带 `msgId / RecalledBy`）做推送编排。状态变更（rpc，幂等可复用）与推送副作用（api 编排）分离。
6. MQ 消费者识别 `ContentRecall` 控制帧 → **跳过 addChatLog**（DB 已更新）→ 复用现有 fan-out 推一条 `push`（`frameType=FrameNoAck`，`contentType=ContentRecall`）。**通用收发通道不变，仅按类型分流**。
7. 各端收到通用 push 后，按 `contentType=ContentRecall` 分流，把本地 `msgId` 气泡原位替换为撤回占位（前端 `recalledIds` 已存在，改为由 ws 事件驱动）。
8. **离线/历史拉取**：`GetChatLog`（im-rpc）对 `Status=已撤回` 的消息**在 logic 层把 `MsgContent` 置空**再返回，并带 `status + recalledBy`，前端渲染占位。

### B. 群 @（happy path）
1. 用户在群聊输入框输入 `@` → 前端调用已有 `GET /social/v1/group/atList`（已就绪）拉取可 @成员，下拉选择；管理员额外可见"所有人"项（前端主门禁）。
2. 发送时，WS `chat.user` 帧的 `msg` 里带 `atUsers: []uid`（和/或 `atAll: true`）。
3. **生产端校验（im-ws chat handler）**：仅当 `atAll=true` 时调 social-rpc 校验发送者角色，非管理员则清除 `atAll`（防御性降级，无效 @all 不进入下游链路，避免污染 Kafka / 消费端负载）。普通消息与普通 @成员零额外开销，不触发 RPC。
4. MQ 消费者 `addChatLog` 时落库 `AtUsers / AtAll`；过滤非群成员 uid；fan-out 推送时 push 帧带 `atUsers/atAll`。
5. 对每个接收者：若其 uid ∈ `atUsers` 或 `atAll=true`，在该用户的 `Conversations` 里把对应会话标记 `HasAtMe=true`。
6. 接收端会话列表展示"[有人@你]"角标；进入会话 / 触发该会话 markRead 时清除 `HasAtMe`。

## 异常处理

| 场景 | 处理方式 |
|------|---------|
| 本人撤回超过 2 分钟 | im-api 返回 xerr 业务错误（如"消息发送已超过 2 分钟，无法撤回"），不改库不推送 |
| 非本人、非管理员尝试撤回 | im-api 返回无权限业务错误 |
| 撤回的 msgId 不存在 / 已撤回 | 幂等：已撤回直接返回成功（不重复推送）；不存在返回业务错误 |
| 管理员角色校验时 social-rpc 失败 | 传播 RPC 错误（不吞），返回 5xx 业务错误，不静默放行 |
| 非管理员发送 `@所有人` | 前端不展示该项；后端在**生产端（im-ws）防御性降级**：仅当 `atAll=true` 时校验角色，非管理员丢弃 `atAll` 标记并打日志（消息本体仍送达，无效 @all 不进 Kafka，消费端不再查角色） |
| `atUsers` 含非本群成员 uid | MQ 消费端过滤掉非成员 uid（以 `GroupUsers` 为准），不报错 |
| 撤回控制帧重复消费（Kafka 重投） | 消费幂等：DB 已是撤回态则只推送不二次改库；推送本身可重复（前端按 msgId 置撤回态，幂等） |
| 被撤回的消息被别人引用（quote） | 引用预览端同步显示"消息已撤回"：前端渲染 `QuoteBlock` 时若被引消息在 `recalledIds` 或历史 status=撤回，则预览文案降级为"消息已撤回" |
| 撤回并发（本人和管理员同时撤回） | DB 更新用条件更新（仅当 Status=正常 时改为撤回），先到者生效，后到者走"已撤回"幂等分支 |

## 技术设计

### 核心设计原则（贯穿全篇）
- **不改实时通信核心链路**：WS 通用 push 通道、读写 goroutine、心跳、ack 机制一律不动。撤回与 @ 只是**新增数据结构 + 按类型/字段分类分流**到对应业务渲染。
- **撤回控制消息复用已读回执的成熟模式**：走 `MsgChatTransfer` 通用 topic，靠 `MsgType=ContentRecall` 分流（与 `ContentMakeRead=6` 完全同构），不新增 topic。
- **分层**：im-api（handler 绑参 + logic 编排）→ im-rpc（撤回业务：校验 + 条件更新）；跨服务（social 角色）一律走 RPC client。
- **权限/降级靠前**：`@所有人` 的角色校验在生产端（im-ws）完成，无效 @all 不进 Kafka。

### 数据模型（MongoDB `chat_logs`，新增字段，⚠️需用户确认 schema 变更）

MongoDB 为弱 schema，新增字段为**纯追加**，对旧文档零迁移、对三库规则无影响（聊天记录仅 Mongo）。

`apps/im/models/chatlogtypes.go` `ChatLog` 增加：

| 字段 | 类型 | bson | 说明 |
|------|------|------|------|
| `Status` | int | `status` | **已存在**，启用枚举：0=正常，1=已撤回 |
| `RecalledBy` | string | `recalledBy` | 撤回操作者 uid（本人或管理员），用于前端文案区分 |
| `RecalledAt` | int64 | `recalledAt` | 撤回时间（unix nano） |
| `AtUsers` | []string | `atUsers` | 被 @的成员 uid 列表 |
| `AtAll` | bool | `atAll` | 是否 @所有人 |

`apps/im/models/conversationtypes.go` `Conversation`（每用户视角）增加：

| 字段 | 类型 | bson | 说明 |
|------|------|------|------|
| `HasAtMe` | bool | `hasAtMe` | 该会话是否有未读的 @我，进会话后清除 |

### 常量与配置

`pkg/constants/im.go`：
```go
// 消息状态
MsgStatusNormal   = 0
MsgStatusRecalled = 1

// 撤回控制帧（push 的 contentType，复用 MsgChatTransfer 链路；编号续 8 之后）
ContentRecall MType = 9
```

**撤回时间窗做成配置项**（决策：可配，默认 120s）：
- `apps/im/rpc/internal/config/config.go` 加 `RecallWindowSeconds int64`
- `apps/im/rpc/etc/im-sample.yaml` 加 `RecallWindowSeconds: 120`
- 两处字段须保持一致（go-zero 约定）；im-rpc `RecallMsg` 校验本人撤回时读此配置，`0` 视为不限时间（保留逃生口）。

### API 接口

| 方法 | 路径 | 服务 | 说明 |
|------|------|------|------|
| POST | `/v1/im/chatlog/recall` | im-api（新增） | 撤回消息：body `{conversationId, msgId, chatType}` |
| GET | `/social/v1/group/atList` | social-api（**已就绪**） | @成员列表，前端直接调用 |
| GET | `/v1/im/chatlog` | im-api（**改造**） | 历史拉取：撤回消息置空 content + 带 status/recalledBy |
| WS | `chat.user`（**扩展帧**） | im-ws | `msg` 增加 `atUsers/atAll` 字段 |
| WS | `push`（**扩展帧**） | im-ws | 撤回事件 `contentType=ContentRecall`；普通消息带 `atUsers/atAll` |

### RPC 接口（`apps/im/rpc/im.proto`，字段编号只增不复用）

- 新增 `RecallMsg(RecallMsgReq) returns (RecallMsgResp)`：入参 `conversationId / msgId / operatorUid / chatType`。**撤回的权限校验、时间窗、条件更新全部在此 logic 内完成**（im-api 只调用，不下沉业务）。
- **角色校验走新增的轻量 RPC `social.GetMemberRole`（优化方案，见下）**，im-rpc svc 注入 social-rpc client。
- 推送不在 rpc 内触发：由 im-api logic 在 rpc 成功后投 Kafka 控制帧（见 Kafka/MQ），保持 rpc 幂等纯粹、便于其他调用方复用。

### 角色校验优化方案（`apps/social/rpc/social.proto` 新增）

**问题**：撤回（管理员）和 @所有人 都需判定"某 uid 在某群是否群主/管理员"。复用现有 `GroupUsers` 会拉**全量成员列表**，大群下浪费且拖慢热路径。

**方案**：新增点查 RPC，单行命中，两个调用方共用：
```proto
GetMemberRole(GetMemberRoleReq) returns (GetMemberRoleResp)
// Req:  { string groupId = 1; string userId = 2; }
// Resp: { int64 roleLevel = 1; bool isMember = 2; }
```
- 实现：`group_members` 按 `(groupId, userId)` 单行查询，返回 `roleLevel`（非成员 `isMember=false`）。
- 调用方：im-rpc `RecallMsg`（管理员撤回）+ im-ws `chat.user`（@all 降级）。
- 后续可在 social 内部对该点查加 Redis 缓存（群成员角色变更时失效），对调用方透明。

### 权限校验（im-rpc `RecallMsg` logic）

```
load ChatLog by msgId
if ChatLog.Status == 已撤回:  return ok          // 幂等
if operator == ChatLog.SendId:
    require now - ChatLog.SendTime <= RecallWindowSeconds(配置,默认120s; 0=不限)   // 本人限时
elif chatType == 群聊:
    role = social.GetMemberRole(groupId, operator).RoleLevel   // 点查，非全量
    require role in (群主, 管理员)                              // 管理员不限时
else:
    reject 无权限
```

### 生产端 @all 校验评审结论（针对反馈 #3）

**结论：可行，但必须带以下约束**——把"同步 RPC 进 im-ws 热路径"的风险关进笼子：

1. **只在 `atAll=true` 时触发**：普通消息、普通 @成员（`atUsers` 非空但无 atAll）一律不调 RPC，热路径零额外开销。@all 本身是低频管理员操作。
2. **fail-closed 且不阻断消息本体**：social-rpc 超时/报错时，按"非管理员"处理——**丢弃 atAll 标记，但消息正文照常发**。绝不因 social 不可用而阻塞或丢弃消息（社交服务抖动最多让 @all 临时降级为普通群消息）。
3. **必须带超时的 ctx**：复用 zrpc 内置超时（不在 handler 手写 retry），且 ctx 跟随连接生命周期，连接关闭即取消，避免 goroutine 悬挂（符合 websocket-im 规则）。
4. **用点查 `GetMemberRole`**：单行查询，避免大群拉全量，进一步压低热路径成本。
5. **边界收敛**：im-ws 对 social 的依赖**仅限这一个角色点查**，不扩散其它 social 调用，保持 ws 层轻量。

> 与微服务规则一致：im-ws 经 social-rpc client 跨服务调用是允许的（MQ 消费者已有先例 `social.GroupUsers`），不算越界；通用收发链路、ack、心跳均不改，仅在 atAll 分支插入一次校验 = "按业务分流"而非"改核心链路"。

### Kafka / MQ（通用通道不变，仅加类型分流）

- **不新增 topic / 不改通用收发链路**：撤回控制消息复用 `MsgChatTransfer`，靠 `MsgType=ContentRecall` 区分（与已读回执 `ContentMakeRead=6` 复用同一链路的模式完全一致）。
- 生产端：im-api 在 rpc 成功后投递撤回控制帧；im-ws 在打包发消息前完成 `atAll` 角色降级（见流程 B-3）。
- MQ 消费者 `msg_chat_transfer.go` 增加分支：
  - `MsgType==ContentRecall` → 跳过 `addChatLog`（DB 已更新），复用现有 fan-out 仅推送。
  - 普通群消息 → 落库 `AtUsers/AtAll`；过滤非群成员 uid；对命中成员置 `Conversation.HasAtMe`。（`atAll` 权限已在生产端把关，**消费端不再查角色**）

### 实现步骤（每步可独立 commit）

1. [ ] **常量 + 数据模型**：`pkg/constants/im.go` 加状态/控制帧常量；`ChatLog` 加 `RecalledBy/RecalledAt/AtUsers/AtAll`，`Conversation` 加 `HasAtMe`（纯追加，旧文档零值兼容，不影响现有读写）；ChatLogModel 加条件更新方法 `UpdateRecalled`（仅当 Status=正常）。
2. [ ] **social GetMemberRole**：改 `social.proto` 加点查 RPC + goctl 生成 + logic（`group_members` 单行查 roleLevel）。im-rpc 与 im-ws 共用。
3. [ ] **撤回时间窗配置**：im-rpc `config.go` + `etc/im-sample.yaml` 加 `RecallWindowSeconds`（默认 120）。
4. [ ] **im-rpc RecallMsg**：改 `im.proto` + goctl 生成 + logic（**权限/时间窗校验经 `GetMemberRole` 查角色 + 条件更新撤回态**）；im-rpc svc 注入 social-rpc client。
5. [ ] **im-api 撤回接口**：改 `im.api` 加 `POST /chatlog/recall` + goctl 生成；handler 仅绑参，logic 仅编排（调 im-rpc → 成功后投 Kafka `ContentRecall` 控制帧）；im-api svc 注入 MsgChatTransfer 生产者（**无需 social client，校验在 rpc**）。
6. [ ] **MQ 消费端 - 撤回**：`msg_chat_transfer.go` 加 `ContentRecall` 控制帧分支（跳过落库，复用 fan-out 只推送）。
7. [ ] **历史拉取过滤**：im-rpc `GetChatLog` logic 对撤回消息置空 content、回传 status/recalledBy。
8. [ ] **@ 生产端 + 落库**：im-ws `chat.user` handler 解析 `atUsers/atAll`，**仅当 atAll 时调 `GetMemberRole` 校验角色降级**（im-ws svc 注入 social-rpc client，带超时 ctx + fail-closed）；MQ `addChatLog` 落库 @字段 + 过滤非成员；fan-out 命中成员置 `HasAtMe`。
9. [ ] **markRead 清除 @角标**：MarkRead 链路清 `Conversation.HasAtMe`。
10. [ ] **前端 - 撤回**：`handleRecallMessage` 改为调用 `/chatlog/recall`；监听 ws `ContentRecall` 事件驱动 `recalledIds`；历史消息按 status 渲染占位；文案按 `recalledBy` 区分（你/对方/管理员）；i18n 文案入 `locales`。
11. [ ] **前端 - @**：输入框 `@` 触发 `GroupAtList`（管理员才显示@所有人）；消息气泡 @高亮渲染；会话列表"[有人@你]"角标 + 进会话清除。
12. [ ] **前端 - 引用降级**：`QuoteBlock` 对被撤回的被引消息显示"消息已撤回"。

### 参考的现有模式

- `apps/task/mq/internal/handler/msg_transfer/msg_read_transfer.go` — 已读回执复用 `MsgChatTransfer` + 控制帧 `ContentMakeRead` 的模式，**撤回控制帧完全照此实现**。
- `apps/task/mq/internal/handler/msg_transfer/msg_chat_transfer.go` — `addChatLog` / fan-out / 群成员推送（调 `social.GroupUsers`）。
- `apps/social/api/internal/logic/group/groupatlistlogic.go` — @成员列表查询（含 RoleLevel），前端 @直接复用。
- `apps/im/ws/websocket/message.go` + `apps/im/ws/ws/ws.go` — WS 帧 / `ws.Chat` 数据结构，扩展 `atUsers/atAll`。
- `web/src/components/im/ChatDetail.tsx` — `recalledIds`(L209) / `handleRecallMessage`(L730) / 撤回占位渲染(L1390) / `QuoteBlock`(L139) 已有骨架，改为服务端驱动。
- `web/src/lib/ws-client.ts` — MType ↔ 前端 type 映射表，加 `ContentRecall` 处理。

## 测试计划
- [ ] 本人 2 分钟内撤回成功；超 2 分钟返回业务错误。
- [ ] 群管理员撤回他人消息成功且不受时间限制；普通成员撤回他人消息被拒。
- [ ] 撤回幂等：重复撤回同一消息不重复推送、不报错。
- [ ] 撤回并发：本人与管理员同时撤回，仅一方生效，另一方走幂等分支。
- [ ] 离线用户上线拉历史：撤回消息 content 为空、带 status，渲染占位。
- [ ] 被引用消息撤回后，引用预览显示"消息已撤回"。
- [ ] @单个/多个成员：被 @者会话角标出现，进会话清除。
- [ ] 管理员 @所有人成功；普通成员 @所有人被降级（atAll 丢弃，消息正文仍送达）。
- [ ] `atUsers` 含非群成员 uid 被过滤。
- [ ] 三库环境无关（聊天记录仅 Mongo，但涉及 social MySQL 角色查询需在三库通过）。
- [ ] 测试后清理测试产生的 chat_logs / conversation 数据。

## 已决策（来自评审）
- ✅ 撤回时间窗**做成配置**：`RecallWindowSeconds`，默认 120s，`0` 不限时。
- ✅ 角色校验**采用优化方案**：新增轻量点查 `social.GetMemberRole`，im-rpc（撤回）与 im-ws（@all）共用，不拉全量。
- ✅ **im-ws 注入 social-rpc client 评审通过**，约束见"生产端 @all 校验评审结论"。

## 待定事项
- 撤回文案细节：管理员撤回展示"管理员撤回了一条消息"还是"群主撤回了 X 的一条消息"。
- @所有人是否也允许群主之外的"管理员"角色（取决于 social 角色枚举的粒度）。
- `HasAtMe` 是否需要细化为"未读 @消息计数 / 定位到具体消息"，MVP 先用 bool。
- `GetMemberRole` 是否一并加 Redis 缓存（MVP 可先不加，留接口）。

## MVP 范围
**包含**：
- 本人 2 分钟内撤回 + 群管理员撤回（私聊/群聊），ws 实时 + 离线拉取一致，原内容 DB 保留、下发置空，前端原位占位渲染。
- @单个/多个群成员 + 管理员 @所有人 + 被 @者会话列表"[有人@你]"角标。
- 被引用消息撤回后引用预览降级。

**后续迭代**：撤回后"重新编辑"回填、全局未读 @汇总页、撤回审计后台、私聊 @、@消息精确定位跳转。
