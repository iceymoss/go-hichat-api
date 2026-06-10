# ChatLog 写入性能优化（批量写 + 分片准备）

## 状态
- 创建日期: 2026-06-10
- 状态: 草稿
- 关联: [`conversation-send-authz-and-relation-cache.md`](conversation-send-authz-and-relation-cache.md)（关系缓存把关系读卸下 DB 后，剩下的真实落库工作量由本 spec 处理）

## 目标
降低高并发下 MongoDB `chatLog` 写入压力：把消费者**逐条 `Insert`** 改为**攒批 `InsertMany`**，在不丢消息、不破坏"回填真实 MsgId 给发送方"语义的前提下，用吞吐换可控延迟。并为后续 MongoDB 分片做好**准备方案**（shard key、索引、迁移步骤），infra 就绪即可启用。

## 非目标
- 本期**不实际搭建分片集群、不迁移数据**——分片只产出可执行的准备方案 + 必要的代码兼容性改造
- 不改消息协议、不改 ws 推送链路、不改鉴权（鉴权见关联 spec）
- 不动读路径分页查询逻辑（`ListBySendTime` / `ListAfter` 等）

## 用户故事
作为运维，在消息高并发时，我希望 `chatLog` 落库以批量方式进行，把单条 `InsertMany` 之外的 DB 往返次数显著降低，且任何一条已 ack 的消息都不丢。

## 核心流程

### 现状（逐条）
```
kq Consume(ctx, key, value)  →  addChatLog:
   ├─ ChatLogModel.Insert(单条)  → 返回 ObjectID
   ├─ data.MsgId = ObjectID.Hex()         （回填，供 echoToSender / UpdateMsg）
   ├─ ConversationModel.UpdateMsg          （会话 last msg + total++）
   ├─ ConversationsModel.Find/Update       （发送方未读）
   └─ markAtMe
→ MsgChatTransfer 扇出推送
→ echoToSender（把真实 MsgId 回推发送方）
```
每条消息 ≈ 1 次 Mongo 写(Insert) + 1 次 Mongo 写(UpdateMsg) + 发送方会话读写。

### 目标（攒批 + 阻塞至 flush 保证 ack 语义）
```
Consume:
   ├─ 构造 ChatLog → 投入 BatchWriter（带 doneCh）
   └─ 阻塞等待 doneCh  ←── 关键：Consume 未返回则 kq 不提交 offset
                            → flush 成功落库后才返回 → 崩溃不丢（at-least-once）
BatchWriter flusher goroutine（按 min(size N, interval T) 触发）:
   ├─ InsertMany(batch)            → 有序返回 InsertedIDs
   ├─ 把各自 ObjectID 写回对应 item，signal 每个 doneCh
   └─（阶段二）BulkWrite 合并会话 UpdateMsg
Consume 拿到 MsgId 后:
   → UpdateMsg / markAtMe / 扇出 / echoToSender（保持原逐条副作用）
```

## 异常处理
| 场景 | 处理方式 |
|------|---------|
| flush 前进程崩溃 | Consume 阻塞未返回 → kq 未提交 offset → 重启重投，**不丢**（攒批期间消息尚未 ack 是核心保证） |
| InsertMany 部分失败 | 用 ordered 插入；失败则对该批降级逐条重试，区分可重试/不可重试（mq-task.md：不可重试写死信不 panic） |
| flusher goroutine 卡死/退出 | 每个 doneCh 带超时；超时则 Consume 返回 error 让 kq 重投，避免消费者永久阻塞 |
| 同会话消息乱序 | 现状生产端 `Push` 无 partition key 即已无序；本期**补 partition key = conversationId**，同会话进同分区保序，批内 InsertMany 有序 |
| 延迟敏感 | flush interval 设小（建议 50ms）+ size 上限，延迟有界；可配置 |
| 优雅退出 | flusher 收到 ctx.Done() 先把缓冲区 flush 干净再退（mq-task.md：长任务可中断） |

## 技术设计

### 批量写组件 `BatchWriter`（新增，建议 `pkg/batchwriter/` 或 mq internal 下）
- 输入 channel：`chan *pendingItem{ log *ChatLog; done chan result }`
- flusher：`select { case item := <-ch: 累积; case <-ticker.C: flush; }`，累积到 N 条或 T 到则 `InsertMany`
- 回填：`InsertMany` 返回的 `InsertedIDs` 顺序与输入一致 → 逐个 `item.done <- result{oid}`
- 参数可配：`BatchSize`（默认 100）、`FlushInterval`（默认 50ms）、`Timeout`

### 模型层
- `ChatLogModel` 新增 `InsertMany(ctx, logs []*ChatLog) ([]primitive.ObjectID, error)`（mongo `collection.InsertMany`，ordered）
- 保留单条 `Insert` 不动（撤回等其它路径仍用）

### kq 并发配置
- `Processors` 需 ≥ 一定并发，否则缓冲区永远凑不满、只能等 timer——这是正确行为但要确认 size 与 Processors 匹配，size 取 ≤ Processors 量级，主要靠 timer 兜底
- 确认 go-zero kq 是 Consume 返回后自动 commit（是）→ 阻塞至 flush 的 ack 模型成立

### 生产端 partition key（保序）
- `mq_client` 的 `Push` 改为带 key 的 `pusher.Push(ctx, key, value)`，key = conversationId
- ⚠️ 需确认 `kq.Pusher` 是否支持指定 key（go-queue kq Push 签名）；不支持则评估替代

### 分片准备方案（本期只产出文档 + 兼容性改造，不执行）
- **shard key 选型**：`conversationId` 哈希分片
  - 理由：所有热查询（`ListBySendTime`/`ListAfter`/按会话拉取）都带 `conversationId` → mongos 定向路由到单 shard，无 scatter-gather；同会话数据共置
  - 取舍：超大群单会话无法跨 shard 再切分（哈希键）；若需要可改 `{conversationId, _id}` 范围分片，文档列出对比
- **scatter-gather 风险点**：按 `_id`(ObjectID) 单独操作的方法在 conversationId 分片下会全分片扫描：
  - `UpdateRecalled` / `UpdateMakeRead` / `FindByMsgId` / `ListByMsgId`
  - 兼容性改造：这些操作尽量补传 `conversationId` 进 filter（撤回/已读本就知道会话），把 scatter-gather 降为定向
- **索引**（无论是否分片都该补，本期可落地）：`{conversationId:1, sendTime:-1}`、`{conversationId:1, _id:1}`；启动时 `EnsureIndexes`
- **迁移步骤文档**：`sh.enableSharding(db)` → `sh.shardCollection("HiChat2.chatLog", {conversationId:"hashed"})` → 预分裂 chunk；附前置索引要求

### 实现步骤（每步可独立 commit）
1. [ ] `ChatLogModel.InsertMany` + 启动时 `EnsureIndexes`（索引本身就是即时收益）
2. [ ] `BatchWriter` 组件 + 单测（size/interval 触发、ID 回填顺序、超时）
3. [ ] `MsgChatTransfer.Consume` 接入 BatchWriter（阻塞至 flush），保留其余副作用逐条
4. [ ] 生产端 `Push` 补 partition key = conversationId（保序，需先确认 kq 支持）
5. [ ] BatchWriter 优雅退出：ctx.Done 时 flush 残留
6. [ ] （阶段二，可选）会话 `UpdateMsg` 改 BulkWrite + 按会话折叠 lastMsg
7. [ ] 分片准备：`_id` 类操作补 conversationId filter（兼容性）+ 输出迁移文档到 `docs/specs/`

### 参考的现有模式
- `apps/task/mq/internal/handler/msg_transfer/msg_chat_transfer.go:62` — `addChatLog` 现状，批量化入口
- `apps/im/models/chatlogmodelgen.go:96` — `Insert` 单条实现，`InsertMany` 照其结构
- `apps/task/mq/mq_client/msg_transfer_client.go:28` — `Push` 现状，补 partition key 处

## 测试计划（table-driven，不 mock DB，三库/Mongo 真实环境）
- [ ] `BatchWriter`：满 N 触发、满 T 触发、ID 与输入顺序一一对应、doneCh 超时返回 error
- [ ] 攒批落库后 `echoToSender` 仍拿到正确真实 MsgId
- [ ] 模拟 flush 前 panic：消息未 ack，重启后重投不丢、不重复（配合幂等）
- [ ] 高并发压测对比：逐条 vs 批量的 Mongo 写 QPS / P99 延迟
- [ ] partition key 生效：同会话消息落同分区、顺序正确
- [ ] 索引 `EnsureIndexes` 在 Mongo 真实库建成功

## 待定事项
- `kq.Pusher` 是否支持指定 partition key —— 决定保序方案能否照搬
- BatchWriter 落点：`pkg/` 复用 vs task/mq internal 私有
- flush interval / size 默认值需压测定标（先给 50ms / 100）
- 阶段二（会话 BulkWrite 折叠）是否本期做，还是后置
- 是否引入死信主题处理不可重试的 InsertMany 失败

## MVP 范围
1. `InsertMany` + `BatchWriter`（阻塞至 flush 的不丢 ack 模型）接入聊天消费者
2. 启动建索引 `EnsureIndexes`
3. partition key 保序（若 kq 支持）
4. 分片**准备方案文档** + `_id` 类操作的 conversationId 兼容性改造（不实际分片）
