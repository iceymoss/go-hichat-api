# 会话发送权限鉴权 + 关系缓存性能优化

## 状态
- 创建日期: 2026-06-10
- 状态: 已实现，待晨审 + 灰度验证（分支 `feat-im-send-authz-relation-cache`，15 commits）
- 关联讨论: 被踢出群 / 删除好友后仍能在原会话发送消息并成功投递（安全漏洞）

## 实现备注（晨审重点）
- 全部 13 步已落地，`go build ./...` 通过；`pkg/relationcache` 与 outbox model 有真实 Redis/MySQL 单测。
- 🟡 鉴权闸门默认 **OFF**（`AuthzGate.Enabled/GroupChat/SingleChat` 全 false），合入不改现状行为；验证缓存/事件链路后再分粒度放量。
- **未做端到端 Kafka 联调**（需起全套服务），靠单测 + 照搬已验证的 trend-notify 链路保证正确性。
- 已知取舍：①读穿透回源 `LoadGroupMembers` 版本用 0，依赖 Kafka 按 gid 分区有序兜底（见待定）；②好友删除走无版本门 `RemoveFriendIfLoaded`（好友事件跨分区，单版本门会误拒）；③前端 IM 组件未接入 i18n，失效横幅按组件现状用硬编码中文；④无好友/群"新增"事件，re-add 后缓存靠 TTL/刷新自愈。

## 目标
修复"会话发送链路零鉴权"导致的安全漏洞：被踢出群、退群、群解散、删除好友后，用户仍能在原会话发送消息并成功投递给对方/群成员。同时通过 **Redis 关系缓存 + 事件驱动维护**，把发消息热路径上的关系/成员校验从 DB 卸载，使 **DB QPS 与消息量解耦**（DB 只在冷加载和关系变更时被打）。

## 非目标
- 不改动消息本体的存储链路（MongoDB chatLog 落库逻辑不动）
- 不引入新的 HTTP `.api` / gRPC `.proto` 接口（鉴权拒绝走 ws 错误帧；关系变更走内部 Kafka topic）
- 不做"陌生人单聊"：本期产品语义为**单聊必须互为好友**
- 不做 MongoDB 分片 / 批量写优化 —— 拆到独立 spec [`chatlog-write-perf.md`](chatlog-write-perf.md)
- 不改动 `.env`、不改动现有 topic 协议

## 兼容性与安全上线（硬约束：全程不影响现有功能与链路）

> 第一原则：任何一步出问题，系统行为必须**自动退化回"改动前的现状"**，绝不阻断正常聊天。

1. **灰度总开关**：鉴权闸门由 config 开关控制（`AuthzGate.Enabled` + `AuthzGate.SingleChat` / `AuthzGate.GroupChat` 分粒度）。**默认 OFF**，上线先观察缓存与事件链路数据正确，再分单聊/群聊逐步打开。关掉即等于现状。
2. **全程 fail-open**：闸门只在"**确定**非成员/非好友"时拒发；缓存 miss / Redis 错误 / RPC 错误 / 任何不确定 → **一律放行**，退化为现状（现状就是不校验）。
3. **绝不移除旧路径(strangler)**：群扇出改读 Redis，但 miss/错误时**回源到现有 `GroupUsers` RPC**——旧链路保留为兜底，不是替换。
4. **纯新增组件零侵入**：新 topic `relationChangeTransfer`、新消费者、新 outbox 表、前端新监听器，都**不触碰**现有 topic / 消费者 / 表 / 事件；新增 consumer 不影响既有 consumer 注册。
5. **social 改造不破坏现有写操作**：outbox 与踢人/删好友同事务——需保证"加了 outbox 写不会让 kick/delete 本身失败"（同库本地表插入，失败概率极低；仍以测试覆盖该失败路径，必要时事件发射也挂开关）。best-effort Redis 同步**永远忽略错误、绝不阻断**主流程。
6. **回滚粒度**：每步可独立 commit、独立回滚；热路径改动（步骤 8/10）的代码即使合入，行为仍由灰度开关锁死在 OFF。
7. **可观测**：闸门每次"拒发"和每次"fail-open 放行"都打点日志，灰度期间据此确认无误拒、无漏放，再放量。

## 用户故事
1. 作为群主/管理员，当我把某成员踢出群后，**该成员不能再向本群发送消息**，且其客户端会话被实时标记为只读、输入框禁用。
2. 作为用户，当我删除某好友后，**双方都不能再在该单聊会话发送消息**，客户端实时禁用输入。
3. 作为运维，**消息高并发时 DB 不应成为第一瓶颈**：正常发消息热路径不查关系表，只读 Redis O(1)。

## 核心流程

### A. 发送鉴权（热路径）
```
单聊：客户端 --chat.user--> ws(conversation.Chat)
        └─ SISMEMBER frd:{sendId} {recvId}
             命中好友 → 投 Kafka；明确非好友 → 拒发 + 回错误帧；miss → 读穿透 FriendList RPC 回填
群聊：ws 直投 Kafka → task/mq 消费者
        └─ 读 grp:mem:{gid}（扇出名单，本来就要读）
             sendId ∈ 成员集 → 落库 + 扇出；不在 → 不落库 + echo 拒绝帧给发送方
```

### B. 关系变更维护缓存（低频路径，outbox 强保证）
```
social: 一个事务里 { 改 group_members / friends 表  +  写 relation_outbox 行 }
      └─ best-effort 顺手 SREM/SADD Redis + INCR ver（乐观快路径，可失败）
relay(social 内后台 poller, Redis 单实例锁): 轮询 outbox status=0
      └─ 投 Kafka relationChangeTransfer（partition key = gid / 好友对，保证同实体有序）
      └─ 标记 outbox status=1
task/mq 消费 relationChangeTransfer:
      ├─ 版本门校验（ver > 本地 ver 才应用）→ SADD/SREM 更新 Redis 关系缓存
      └─ 推 ws 帧（method=relation.changed）给受影响用户 → 客户端禁用输入闭环
```

### C. 冷启动 / 读穿透（防缓存击穿）
```
grp:mem:{gid} miss → SetNX grp:mem:{gid}:lock EX5（单飞）
      赢者: RPC GroupUsers → 重建 SET + 写 ver（CAS：ver 被推进则放弃本次重建结果）→ 释放锁
      其余: 本条消息降级直连 RPC（fail-open 友好），不放过 DB 击穿
```

## 异常处理
| 场景 | 处理方式 |
|------|---------|
| Redis 不可用 | 闸门 **fail-open**（放行）：退化到 RPC；RPC 也失败仍放行，绝不让 Redis 抖动全员禁言 |
| 关系明确不存在（非成员/非好友） | 闸门 **fail-closed**（拒发）+ echo 错误帧 |
| social 改库后发事件前崩溃 | outbox 与业务改动**同事务提交**，relay 重启后继续投递（at-least-once） |
| 事件乱序（踢了又拉回） | Kafka 按实体分区有序 + 事件带 `version`，消费端版本门丢弃旧版本 |
| 事件重复消费 | SADD/SREM 幂等 + 版本门，重放无副作用（符合 mq-task.md 幂等要求） |
| read-repopulate 竞态（旧成员复活） | 重建抢单飞锁 + ver CAS，后发生者赢 |
| 缓存漏更（前述都漏） | key TTL 1h 自愈；上线后读穿透从 RPC 权威重建 |
| 被踢者离线时收到 relation.changed | ws 静默丢弃，上线拉会话列表时已是只读态（前端据会话状态判断） |
| relay 多副本并发 | Redis 分布式单实例锁（参考 mq-task.md 定时任务约定） |

## 技术设计

### 数据模型

**新增表 `relation_outbox`**（事务性发件箱，建在 social 库；TEXT 存 JSON，兼容三库，主键交给 ORM）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | 主键 | 单调自增，**兼作 version**（同分区单调） |
| event_type | VARCHAR | group.member.removed / group.disbanded / friend.deleted |
| payload | TEXT | 事件 JSON（用 common/json 序列化） |
| status | INT | 0 待投递 / 1 已投递 |
| group_id | VARCHAR | 分区/索引用（好友事件为空） |
| created_at | DATETIME | |
| sent_at | DATETIME | NULL |

> ⚠️ schema 变更前需用户确认（database-model.md）。索引：`(status, id)` 供 relay 轮询。

**Redis key**
| key | 类型 | 说明 | TTL |
|-----|------|------|-----|
| `grp:mem:{gid}` | SET\<uid\> | 群成员集（authz + 扇出共用），含 `__LOADED__` 哨兵区分空群/未加载 | 1h |
| `grp:mem:{gid}:ver` | STRING int64 | 版本门 | 1h |
| `grp:mem:{gid}:lock` | STRING | 冷启动单飞锁 | 5s |
| `frd:{uid}` | SET\<friendUid\> | 好友集（双向维护） | 1h |
| `frd:{uid}:ver` | STRING int64 | 版本门 | 1h |

### Kafka topic（common 关系变更通知，参考 trendNotifyTransfer）
- topic 名：`relationChangeTransfer`（沿用现有 camelCase 约定，与 msgChatTransfer/trendNotifyTransfer 一致）
- 消息体 `mq.RelationChangeTransfer`：`{ eventType, groupId, userId, operatorId, friendA, friendB, version, timestamp }`
- partition key：群事件用 `groupId`，好友事件用 `min(a,b)+max(a,b)`，保证同实体有序

### 实现步骤（每步可独立 commit）

> 标签：🟢 纯新增/零侵入（不碰旧链路）｜🟡 改动热路径，必须灰度开关 + fail-open 兜底
> 实现顺序：先把 🟢 全部落地并验证缓存/事件链路正确，最后再合入 🟡（且默认 OFF）。

**social 侧（生产端 + outbox + best-effort Redis）**
1. [ ] 新建 `relation_outbox` 表 + `socialmodels/relationoutboxmodel.go`（参考现有 sqlx model 结构）
2. [ ] social config + yaml 接入 Kafka 生产端（`RelationChangeTransfer{Addrs,Topic}`）+ Redis（参考 trend api 的 `TrendNotifyTransferClient` 接线、`pkg/db.GetRedisConn`）
3. [ ] `groupkicklogic` / `groupquitlogic` / `groupdisbandlogic` / `frienddeletelogic`：**统一事务路径**，在删成员/好友的同事务内写 outbox 行；事务提交后 best-effort SREM/SADD Redis + INCR ver
4. [ ] social 内后台 relay poller（Redis 单实例锁）：轮询 outbox → 投 `relationChangeTransfer` → 标记 sent

**task/mq 侧（消费事件维护缓存 + 客户端闭环推送）**
5. [ ] `mq/mq.go` 加 `RelationChangeTransfer` 结构体；`mq_client/relation_change_client.go` 生产端封装（照抄 `trend_notify_client.go`）
6. [ ] `internal/handler/msg_transfer/relation_change_transfer.go` 消费者（照抄 `trend_notify_transfer.go`）：版本门校验 → 更新 Redis 关系缓存 → 推 `relation.changed` ws 帧
7. [ ] `listen.go` + config + yaml 注册新 consumer（照抄 TrendNotifyTransfer 四处接线）
8. [ ] 🟡 `base_msg_chat_transfer.go` 的 `group()`：扇出名单改读 `grp:mem:{gid}`（**miss/错误 → 回源 `GroupUsers` RPC**，旧链路保留）；落库前做 `sendId ∈ 成员集` 群聊 authz 闸门，**受 `AuthzGate.GroupChat` 开关控制，不确定即放行**
9. [ ] 新建 Redis 关系缓存封装 `pkg/relationcache/`（SISMEMBER/SMEMBERS/单飞回源/版本门/CAS），social 与 task/mq 共用

**ws 侧（单聊发送闸门 + 闭环推送 handler）**
10. [ ] 🟡 `conversation.Chat` 单聊分支：投 Kafka 前 `SISMEMBER frd` 闸门（miss 读穿透 FriendList RPC），非好友拒发 + 错误帧；**受 `AuthzGate.SingleChat` 开关控制，缓存/RPC 不确定即放行**
11. [ ] `ws/ws.go` 加 `RelationChanged` 推送结构体；`handler/push/relation_notify.go`（照抄 `trend_notify.go`）；`routes.go` 注册 `push.relation` → 前端 `relation.changed`

**前端侧（实时禁用输入闭环）**
12. [ ] `ws.on('relation.changed')`：定位会话 → 标记只读、禁用 composer、可选系统消息"你已不在群聊/已删除好友"
13. [ ] 用户可见文案走 `t()`，翻译写入 `web/src/i18n/locales/{lang}.json`（frontend.md）

### 参考的现有模式
- `apps/task/mq/internal/handler/msg_transfer/trend_notify_transfer.go` — 新 topic 消费者骨架（照抄）
- `apps/im/ws/internal/handler/push/trend_notify.go` + `routes.go:33` — ws 推送 handler + 路由注册（照抄）
- `apps/task/mq/mq_client/trend_notify_client.go` — Kafka 生产端封装（照抄）
- `apps/trend/api/internal/svc/servicecontext.go:29` + `config.go:20` + yaml — 跨服务 Kafka 生产端接线（social 照抄）
- `apps/im/api/internal/logic/recallmsglogic.go:60` — "状态变更与推送副作用分离"范式
- `apps/task/mq/internal/handler/msg_transfer/base_msg_chat_transfer.go:57` — 群扇出现读 GroupUsers 处，改读 Redis
- `apps/social/rpc/internal/logic/friendputinhandlelogic.go:85` — `Model.Trans(ctx, session)` 事务范式（outbox 挂同 session）
- `pkg/db/redis.go:17` — `GetRedisConn() *redis.Client`

## 测试计划（table-driven，不 mock DB，三库通过，测后清理）
- [ ] Redis 关系缓存单测：SISMEMBER 闸门判定、读穿透回源、单飞锁、版本门拒旧版本、ver CAS 防复活
- [ ] 群聊：踢人 → 事件 → 缓存 SREM → 被踢者发消息被拒、不落库、不扇出
- [ ] 群聊：正常成员发消息命中缓存、零 DB 关系查询
- [ ] 单聊：删好友（双向）→ 双方发消息均被拒
- [ ] outbox：变更与 outbox 同事务，模拟发事件前崩溃 → relay 重启补投
- [ ] fail-open：Redis 不可用时不阻断正常发送
- [ ] 闭环：relation.changed 帧正确推送给被踢/被删者
- [ ] `relation_outbox` 建表在 SQLite/MySQL/PostgreSQL 三库通过

## 待定事项
- relay 落点：social 进程内后台 goroutine vs `apps/task/cron`（后者读 outbox 会跨服务库边界，倾向前者）— **建议 social 内**
- social 事务统一：现状 disband 用 GORM `tx.Table`、kick/quit 用 sqlx `Model.Delete`，outbox 同事务要求二选一统一，需在实现时确认每个 logic 的事务载体
- 好友 id 类型：Friends 表 `UserId/FriendUid` 为 uint64，ws/Redis 用 string，需统一规范化为 string
- 是否给"被移出"附带一条系统消息落库（让历史可见），还是仅前端 banner — 倾向仅前端，不落库
- 版本门粒度：用 outbox.id 全局单调即可，还是需要 per-group 独立序列 — 倾向 outbox.id + 分区有序

## MVP 范围
按已确认决策，**本期一次性完整交付**：
1. 群聊 + 单聊发送鉴权闸门（fail-closed 于确定否定 / fail-open 于基础设施故障）
2. Redis 关系缓存（群成员集 + 好友集）+ 读穿透单飞 + 版本门
3. 四层一致性：L1 best-effort 写穿透 / L2 **outbox 强保证** / L3 版本门 / L4 TTL+读穿透
4. common `relationChangeTransfer` Kafka topic（参考 trend notify + recall 设计）
5. 客户端实时禁用输入闭环
