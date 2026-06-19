# 音视频通话（Streaming Audio/Video Call）

## 状态
- 创建日期: 2026-06-19
- 状态: 草稿（待审阅）
- 作者: iceymoss / Claude
- 目标服务: `apps/streaming`（+ im / social / user / web 对接）

---

## 0. 背景与现状盘点（以代码为准）

现有 `apps/streaming` 是一个**信令骨架，通话功能基本不可用**：

| 模块 | 现状 | 结论 |
|------|------|------|
| 服务入口 `streaming.go` | 裸 `http.ListenAndServe(:10093)`，单 `/ws` | 未接 etcd / 不在 `hichat2.sh` |
| 鉴权 | `userID` 直接取自消息体，`CheckOrigin` 恒 true | **零鉴权，可冒充任意用户** |
| 连接表 `connections` | 仅在 `join_room` 时登记 `userID→conn` | `call_invite` 给未 join 的被叫**静默丢失** |
| 媒体转发 | 每用户建服务端 pion PC 塞进 SFU，但 SFU 的 `ForwardMediaStream/HandleMediaStream` 从未被调用，`OnTrack` 未设 | **媒体到服务端即丢，既非 P2P 也非可用 SFU** |
| 会议/录屏/直播 manager | 纯内存桩，无落库 | 重启即丢 |
| RPC client（user/social） | `servicecontext.go:27-29` 被注释 | 无好友校验、无用户资料 |
| 前端 `CallDialog.tsx` | 仅确认弹窗，`handleCall()` 是 `// TODO` | **无 WebRTC / 无通话界面 / 无来电界面** |
| 前端 `ChatDetail.tsx` | 顶部已有语音/视频按钮（`:1142/:1151`）→ 打开 CallDialog | UI 入口已就位 |

**可复用的成熟基础设施：**
- im ws：JWT 鉴权（`apps/im/ws/internal/handler/auth.go`，支持 `?token=`）、连接管理、按 uid 单推。
- 通用通知通道 `commonNotifyTransfer`（[`common-notify-channel.md`](common-notify-channel.md)）：业务方投 Kafka → `task/mq` 落库 + ws 单推 → 前端 `ws.on('notify')`。**呼叫振铃复用此通道。**
- 关系缓存 `pkg/relationcache`（[`conversation-send-authz-and-relation-cache.md`](conversation-send-authz-and-relation-cache.md)）：`frd:{uid}` Redis 集合判好友，O(1)。**通话好友校验复用。**
- 前端 ws 客户端 `web/src/lib/ws-client.ts` + `chat-store.ts`：`ws.on(method, handler)` 方法路由。
- im 消息体系：消息有 `MType`（text/image/voice/video…）。**通话记录 = 新增 `MType: call`。**

---

## 1. WebRTC 扫盲（30 秒看懂本设计在做什么）

> 给不熟悉流媒体的同学：一通音视频通话分两件事，**别混淆**。

1. **信令（Signaling）= 打电话前的“对暗号”**：A 想呼叫 B → 通知 B 有来电（振铃）→ B 接听 → 双方交换两样东西：
   - **SDP（Offer/Answer）**：我支持什么编解码、要不要视频。
   - **ICE candidate**：我在网络上的可达地址（内网 IP、公网 IP、中继地址）。
   信令只是“交换小纸条”，**走我们自己的 WebSocket**。服务器负责把纸条递给对方。

2. **媒体（Media）= 真正的音视频数据**：暗号对上后，浏览器之间（或浏览器↔服务器）建立加密的 UDP 通道直接传音视频。**这条通道由浏览器内核 + pion 库自动建立**，我们不碰字节。

3. **STUN / TURN = 帮你穿过路由器（NAT）**：
   - **STUN**：告诉你“你的公网地址是多少”，多数家庭网络靠它就能直连。免费（用 Google 的）。
   - **TURN**：直连失败时的**中继兜底**（约 10-20% 网络需要，移动网/企业网更高）。需自建 `coturn`，**本期先留接口不实现**。

4. **P2P vs SFU = 媒体怎么连**：
   - **P2P**：两个浏览器直连，服务器只递纸条。延迟最低，适合 1:1。
   - **SFU**：每个浏览器只和**服务器**连，服务器收下每个人的流再转发给其他人。多人必须用，因为 N 人 P2P 网状连接会爆炸（N×(N-1) 条连接）。

---

## 2. 目标

为 `apps/streaming` 实现**真正可用**的一对一与多人群组音视频通话，前后端打通，并把架构设计成可平滑扩展到会议 / 直播 / 录制 / 屏幕共享。

具体到 MVP（本期交付）：
1. **1:1 语音 + 视频通话**（好友之间），P2P 直连模式，端到端跑通（STUN）。
2. **多人群组通话**（群成员），SFU 服务端转发模式。
3. **通话记录落库**并在聊天里展示（“通话时长 02:35”/“已取消”/“未接听”），可点击回拨。
4. **生产级接入**：JWT 鉴权、RPC 校验好友/群成员关系、纳入 `hichat2.sh` 统一启动。
5. **前端完整体验**：发起、来电振铃、接听/拒接、通话中界面（视频画面/语音头像、静音、开关摄像头、挂断、计时）。

---

## 3. 非目标（本期明确不做，但预留扩展点）

- ❌ **TURN 中继 + 短期凭证下发**：仅留 ICE 配置下发接口与配置项，coturn 部署后再填（[待定事项](#11-待定事项)）。
- ❌ **跨 ws 节点路由**：沿用现有“单 ws 节点”遗留限制（与 `common-notify-channel` 一致）；协议层预留 `node` 字段，多节点路由后续独立 spec。
- ❌ **会议管理**（主持人控制、等候室、全员静音）：保留 `meeting_manager` 桩，本期不接媒体，后续基于 SFU Room 扩展。
- ❌ **直播 / 屏幕共享 / 录制**：保留桩与协议占位，后续基于 SFU Room 扩展（直播 = 1 publisher + N subscriber + 可选 CDN/HLS）。
- ❌ **移动端原生**：仅 Web（Next.js）。
- ❌ 不改 `.env`；schema 变更仅新建表 / `ADD COLUMN`，三库兼容。

---

## 4. 架构决策（核心，先读这一节）

### 4.1 双通道信令：振铃走 im ws，媒体协商走 streaming ws

```
         ┌──────────────── 呼叫控制（低频，需在线/离线/跨节点）────────────────┐
         │  call.invite / cancel / accept / reject / busy / timeout / end      │
A 前端 ──┤  复用 im ws + commonNotifyTransfer（在线推送 + 落库 + 前端 ws.on）  ├── B 前端
         └────────────────────────────────────────────────────────────────────┘
         ┌──────────────── 媒体协商（高频 ICE，按通话建立）──────────────────┐
A 前端 ──┤  streaming ws（:10093，JWT）: join / offer / answer / ice / leave  ├── B 前端
         └──────────────────────────── 1:1 直连 / 群组经 SFU ─────────────────┘
```

**为什么拆两条：**
- 振铃需要“找到可能离线/在另一台机器的用户”——im ws + 通用通知通道**已经解决**了在线判断、离线落库、（未来）跨节点。重新造一遍不划算。
- 媒体协商是高频、低延迟、通话期专属的流量，且**群组 SFU 必须让客户端直接和 streaming 服务交换 SDP/ICE**（客户端和 SFU 配对）。所以媒体必须有一条直达 streaming 的通道。
- 用户**平时只挂着 im ws**；streaming ws **仅在一通通话期间按需建立**，省掉全员常驻空连接。

### 4.2 统一 Room 抽象，模式可切换 `p2p | sfu`

> 这里和已有文档 [`streaming-one-to-one-call-flow-explained.md`](streaming-one-to-one-call-flow-explained.md)（主张“1:1 也走 SFU 以统一架构”）做了一个**有意识的折中**，不是疏漏：

一通通话 = 一个 `Room`，Room 有 `mode`：
- **`p2p`（1:1 默认）**：streaming 服务**不建服务端 PC**，只在两个 peer 之间**转发** offer/answer/ice。延迟最低、服务器零媒体带宽。
- **`sfu`（群组 / ≥3 人）**：streaming 服务为每个参与者建服务端 pion PC，收流（`OnTrack`）+ 转发（`ForwardMediaStream`）。

**关键：两种模式信令协议完全相同**（都是 join/offer/answer/ice 经 streaming ws）；唯一区别是“应答方”是对端 peer（p2p）还是 SFU 服务器（sfu）。因此：
- 1:1 拿到低延迟，群组拿到可扩展，**代码主干统一**。
- 1:1 未来若需录制/审核，把该 Room 的 `mode` 切到 `sfu` 即可，协议和前端不变。
- 会议/直播 = `sfu` Room + 上层策略（角色、权限、发布/订阅约束），**复用同一媒体核心**。

### 4.3 状态归属

| 状态 | 存储 | 说明 |
|------|------|------|
| 通话生命周期 `CallSession`（inviting→ringing→connected→ended） | 内存（权威）+ 结束落 `streaming_call` 表 | 通话记录、时长、结束原因 |
| Room / Participant 媒体成员 | 内存 | 进出房间、轨道映射 |
| 用户“忙线”态 | Redis `call:busy:{uid}`（TTL 兜底） | 拒绝并发第二通来电 |
| Room→节点亲和（多节点用） | Redis `call:room:{id}=node`（预留） | 本期单节点不强制 |
| 好友/群成员校验 | 复用 `pkg/relationcache`（`frd:{uid}` / `grp:mem:{gid}`） | O(1)，miss 回源 RPC |

### 4.4 服务边界（遵守 microservice 规则）
- streaming **不直连** im/social/user 的库；要数据走 RPC（`FriendList`/`GroupUsers`/`GetUserById`）或 Redis 关系缓存。
- 通话记录写入 im：**复用 im 的消息投递链路**（投 `msgChatTransfer` 或调 im-rpc），streaming 不写 MongoDB chatLog。
- streaming 独占自己的 `streaming_call` 表。

---

## 5. 用户故事

1. 作为用户，我在和好友的聊天页点“语音/视频通话”，对方手机/网页**立即振铃**，接听后我们能实时通话。
2. 作为被叫，我在**任意页面**都能收到全屏来电界面，可接听或拒接；正在通话时被呼叫，对方收到“忙线”。
3. 作为主叫，对方 30 秒未接 → 自动结束并提示“未接听”；我可中途取消。
4. 作为通话双方，我能静音、开关摄像头、看到通话计时、随时挂断；挂断后聊天里留一条“通话时长 02:35”，点它能回拨。
5. 作为群成员，我能发起群通话、选择邀请哪些成员，多人同时音视频；有人中途加入/离开，画面网格实时更新。
6. 作为离线被叫，我上线后能在聊天里看到一条“未接来电”记录。
7. 作为开发者，我后续加“会议/直播”时复用同一 Room/SFU 核心，只加上层策略与一个 Room mode 策略。

---

## 6. 核心流程

### 6.1 一对一通话 Happy Path（p2p 模式）

```
A(主叫)                im ws / commonNotify           streaming ws(:10093)              B(被叫)
  │  1. 点视频通话                                                                        │
  │  2. streaming-rpc CreateCall(callerA, calleeB, video)                                 │
  │     ├─ 校验好友(relationcache frd:{A} 含 B)                                           │
  │     ├─ 校验 B 是否忙线(call:busy:{B})                                                 │
  │     └─ 建 CallSession(inviting) + 写 call:busy:{A}/{B}                                │
  │  3. 投 commonNotify{notifyType:"call.invite", receiver:B, payload:{callId,type,caller}}│
  │ ────────────────────────────────────────────────────► task/mq 落库+推 ws ──────────► │ 4. 全屏来电界面(振铃)
  │  5. A 也建 streaming ws 连接, join(callId) 等待                                        │
  │                                              6. B 点“接听” → call.accept (经 im ws/rpc)│
  │ ◄──────────────────── 7. 推 call.accept 给 A ───────────────────────────────────────│
  │  8. 双方 getUserMedia(拿摄像头/麦克风) → 各自连 streaming ws, join(callId)             │
  │  9. A createOffer → streaming ws[offer] ──relay──► streaming ws ───────────────────► B│
  │ ◄──────────────────────────────────── B createAnswer → [answer] ──relay──────────────│
  │ 10. 双方 onicecandidate → [ice] ──relay──► 对端 addIceCandidate（多次往返）            │
  │ 11. ICE 打通 → P2P 加密媒体直连，画面/声音出现（streaming 不经手媒体字节）             │
  │ 12. 任一方挂断 → call.end → 双方关闭 PC；streaming 写 CallSession(ended,duration)      │
  │ 13. streaming 投 msgChatTransfer 一条 MType=call 消息 → 双方会话出现通话记录气泡        │
```

> **关键修复点**（对照现状）：第 2/3 步用 RPC + 通用通知，**不再依赖“被叫已 join 同房间”**；振铃送达可靠。第 5-11 步媒体走 streaming ws 的 **relay 模式，删除现有“服务端建 PC 应答”的错误实现**。

### 6.2 群组通话（sfu 模式）

1. 发起方在群聊点通话 → `CallDialog` 勾选成员 → streaming-rpc `CreateGroupCall(host, [members], video)`，校验都是群成员（`grp:mem:{gid}`）。
2. 向选中成员各投一条 `call.invite`（群振铃）。
3. 接听者 `getUserMedia` → 连 streaming ws → `join(callId)`：streaming 为其建服务端 PC。
4. **发布**：participant 把本地音视频 track 发给 SFU（offer/answer 与 SFU 协商，SFU `OnTrack` 收流）。
5. **订阅**：SFU 把房间内其他人的 track `AddTrack` 给该 participant，触发**重协商**（renegotiation）下发新 offer。
6. 有人加入/离开 → SFU 增删 track + 重协商 → 前端网格增删画面。
7. 最后一人离开 → 关闭 Room，写通话记录（群通话记录形态见 [待定](#11-待定事项)）。

### 6.3 通话记录落库 + 聊天展示

- CallSession 结束 → streaming 落 `streaming_call` 表（权威记录）。
- streaming 复用 im 投递链路写一条 `MType:"call"` 消息（`Content` = `common.Marshal` 的 JSON）：
  ```json
  {"callType":"video","status":"completed","duration":155,"direction":"outgoing","callId":"..."}
  ```
  `status`：`completed`(已接通) / `canceled`(主叫取消) / `missed`(被叫未接/离线) / `rejected`(被叫拒接) / `no_answer`(超时)。
- 前端 `ChatDetail` 的 `MessageContent` 增加 `type==='call'` 分支，渲染通话气泡（图标+文案+时长），点击 → `handleOpenCall(callType)` 回拨。

---

## 7. 异常处理

| 场景 | 处理方式 |
|------|---------|
| 被叫离线 | commonNotify 已落库 → 上线在聊天看到“未接来电”；主叫 30s 超时收 `no_answer` |
| 被叫忙线（已在通话） | `CreateCall` 检 `call:busy:{B}` → 直接回主叫 `call.busy`，不振铃 |
| 主叫取消（响铃中挂断） | `call.cancel` → 被叫来电界面消失，记录 `canceled` |
| 被叫拒接 | `call.reject` → 主叫提示，记录 `rejected` |
| 接听超时（30s 无应答） | 服务端定时器触发 `call.timeout`→`no_answer`，清忙线态 |
| 用户授权摄像头/麦克风失败 | 前端 `getUserMedia` catch → 提示“无法访问设备”，发 `call.end(reason=media_denied)` |
| ICE 长时间不通（无 TURN 的网络） | `iceConnectionState=failed` → 前端提示“连接失败”，挂断；记录 `failed`（[待定]TURN 兜底） |
| streaming ws 断线 | 通话中断线 → 触发 `oniceconnectionstatechange` 重连/挂断；清理 Room 与忙线态 |
| 进程崩溃 / 连接泄漏 | streaming ws 加**心跳 ping/pong**，超时关连接并清 Room（现状无心跳，必补） |
| 重复 `call.invite`（Kafka 重投） | 通用通道幂等键 `bizId=callId`，前端按 callId 去重 |
| 并发：同时互拨 | CallSession 以 `callId` 为准；`call:busy` 抢占，后到者 busy |
| 群通话部分成员离线 | 在线者振铃，离线者落“未接来电”；不阻塞通话建立 |
| 好友已删除 / 已退群 | `CreateCall` 关系校验失败 → 拒绝发起（复用 relationcache，fail-open 取舍同 im） |
| 未鉴权连 streaming ws | JWT 校验失败直接拒绝升级 |

---

## 8. 技术设计

### 8.1 信令协议

#### A. 呼叫控制（im ws，复用 `commonNotifyTransfer`，前端 `ws.on('notify')` 按 `notifyType` 分支）

| notifyType | 方向 | payload 要点 |
|---|---|---|
| `call.invite` | 主叫→被叫 | callId, callType(voice/video), mediaMode(p2p/sfu), caller{uid,name,avatar}, groupId? |
| `call.cancel` | 主叫→被叫 | callId |
| `call.accept` | 被叫→主叫 | callId, callee{uid} |
| `call.reject` | 被叫→主叫 | callId |
| `call.busy`  | 系统→主叫 | callId |
| `call.timeout`| 系统→双方 | callId |
| `call.end`   | 任一方→对端 | callId, reason, duration |

> accept/reject 等也可经 streaming-rpc 触发再投通知；invite/busy/timeout 由 streaming 服务投递。

#### B. 媒体协商（streaming ws `:10093/ws?token=<jwt>`，帧沿用现有 `SignalingMessage{type,callId,userID,data,timestamp}` 但收敛类型）

| type | 含义 | data |
|---|---|---|
| `join` | 加入通话房间 | {callId} → 服务端按 Room.mode 决定 relay/SFU |
| `offer` | SDP offer | {sdp, to?} （p2p:to=对端；sfu:to=server） |
| `answer` | SDP answer | {sdp, to?} |
| `ice` | ICE candidate | {candidate, sdpMid, sdpMLineIndex, to?} |
| `media-state` | 静音/开关摄像头 | {audio:bool, video:bool} → 广播房间 |
| `leave` | 离开房间 | {callId} |
| `peer-joined`/`peer-left` | 服务端→客户端 | {uid} 群组成员变化触发重协商 |
| `ping`/`pong` | 心跳 | — |

> **删除**现有 `meeting_*/live_*/screen_share_*/group_invite` 等信令在 ws 上的处理（移到内存桩/后续），ws 只保留媒体协商必需类型，避免现状那一大坨半成品分支。

### 8.2 数据模型

> ⚠️ **schema 变更需用户二次确认（database-model 规则）**。下表为提议，确认后再建。GORM 处理主键（勿用 AUTO_INCREMENT），JSON 用 TEXT，三库（SQLite/MySQL/PostgreSQL）兼容，注册到迁移。

新建表 `streaming_call`（streaming 服务独占）：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | uint64 | 主键（GORM 自增） |
| `call_id` | varchar | 业务通话 ID（唯一索引） |
| `call_type` | varchar | voice / video |
| `media_mode` | varchar | p2p / sfu |
| `scope` | varchar | single / group |
| `caller_id` | varchar | 主叫 uid（索引） |
| `callee_id` | varchar | 被叫 uid（1:1；群组为空） |
| `group_id` | varchar | 群 ID（群组通话） |
| `participants` | TEXT | 参与者 uid 列表 JSON（群组） |
| `status` | varchar | completed/canceled/missed/rejected/no_answer/failed |
| `started_at` | datetime | 接通时间，可空 |
| `ended_at` | datetime | 结束时间，可空 |
| `duration` | int | 时长（秒） |
| `created_at` | datetime | 发起时间 |

索引：`uniq_call_id(call_id)`、`idx_caller(caller_id)`、`idx_callee(callee_id)`。

im 侧通话消息：**不新建表**，复用现有 chatLog，新增 `MType="call"`，`Content` 存上文 JSON。

### 8.3 配置与 ICE 下发

- `etc/streaming-sample.yaml`：保留 STUN（Google）；TURN 段保留但默认空；新增 `JwtAuth.AccessSecret`（与 im 一致）、RPC（user/social/im）etcd key。
- **新增 HTTP 接口 `GET /v1/streaming/ice-servers`**（JWT 鉴权）：前端**不再硬编码** ICE，启动通话前拉取。本期返回静态 STUN；接 TURN 后此处下发**短期 TURN 凭证**（HMAC，预留实现位）。

### 8.4 服务端模块重构（go-zero / 目录约定）

```
apps/streaming/
  streaming.go              # 入口：JWT ws + RPC client 注入 + 进 hichat2.sh
  internal/
    config/                 # +JwtAuth +UserRpc/SocialRpc/ImRpc +ICE
    svc/                    # 注入 user/social/im client + relationcache + redis
    handler/
      signaling.go          # 收敛为：鉴权 ws + 心跳 + join/offer/answer/ice relay
      iceserver.go          # GET /v1/streaming/ice-servers
    logic/
      call.go               # CreateCall/Accept/Reject/Cancel/End + 好友校验 + 忙线 + 振铃投递
      room.go               # Room(mode) 抽象，p2p relay / sfu 调度
      callrecord.go         # 结束落 streaming_call + 投 im 通话消息
  room/  sfu/  webrtc/      # 保留：sfu 真正接线(OnTrack/Forward/重协商)，p2p relay 复用 conn 注册
  models/                   # streaming_call model（goctl model 生成）
```

- 连接注册：**ws 升级后立即用 JWT 的 uid 登记 `connections[uid]=conn`**（修复现状“只在 join_room 登记”）。
- 心跳：客户端定时 ping，服务端超时关连接并清理 Room/忙线态（websocket-im 规则）。
- 锁：`connections` / Room map 用 `sync.RWMutex` 或 `sync.Map`（现状已有锁，补全清理路径）。

### 8.5 前端模块

```
web/src/lib/
  webrtc-call.ts            # 新增：CallManager(getUserMedia, RTCPeerConnection, perfect-negotiation,
                            #        连 streaming ws, p2p/sfu 两模式, ICE 来自 /ice-servers)
  ice-config.ts            # 拉 GET /v1/streaming/ice-servers
web/src/components/im/
  CallDialog.tsx            # 发起确认（已有）→ 接 CallManager.start()
  IncomingCallOverlay.tsx   # 新增：全屏来电（ws.on('notify',call.invite) 触发, 接听/拒接)
  CallScreen.tsx            # 新增：通话中(视频网格/语音头像, 静音, 摄像头, 挂断, 计时)
  ChatDetail.tsx            # MessageContent 加 type==='call' 气泡 + 点击回拨
  chat-store / ws-client    # ws.on('notify') 增 call.* 分支 → 驱动来电/通话状态
```
- 组件库用 Semi Design + Tailwind；新文案走 `t()` i18n（frontend 规则）。
- 来电界面**全局挂载**（IMLayout 顶层），任意页面可弹。

### 8.6 RPC / 跨服务依赖

| 调用 | 用途 |
|---|---|
| social `FriendList` / relationcache `frd:{uid}` | 1:1 好友校验 |
| social `GroupUsers` / relationcache `grp:mem:{gid}` | 群成员校验 |
| user `GetUserById` | 来电展示昵称/头像 |
| 投 `commonNotifyTransfer` | 振铃信令下发（复用通道） |
| 投 `msgChatTransfer` 或 im-rpc | 写通话记录消息 |

---

## 9. 实现步骤（每步可独立 commit；分支 `feat-streaming-call-*`）

> 风格遵循已有 spec：每阶段可独立验证、可回滚。

### Phase 0 — 生产级接入地基
1. [ ] streaming 入口接 JWT 鉴权 ws（复用 `token.TokenParser`，支持 `?token=`）；config/svc 加 `JwtAuth`。
2. [ ] 注入 user/social/im RPC client + redis + relationcache（解开 `servicecontext.go` 注释并补全）。
3. [ ] `GET /v1/streaming/ice-servers`（返回 STUN，TURN 预留）；config 整理。
4. [ ] 纳入 `hichat2.sh` 启动列表；端口/etcd key 不冲突自检。

### Phase 1 — 1:1 P2P 通话闭环（最小可用，先语音后视频）
5. [ ] **重写 signaling.go**：ws 升级即按 uid 登记连接；加心跳；信令收敛为 join/offer/answer/ice **relay 模式**（删服务端应答 PC）。
6. [ ] `logic/call.go`：CreateCall/Accept/Reject/Cancel/End + 好友校验 + 忙线态 + 30s 超时定时器 + 投 `call.invite` 等。
7. [ ] 前端 `webrtc-call.ts`（perfect-negotiation）+ `ice-config.ts`。
8. [ ] 前端 `IncomingCallOverlay` + `CallScreen` + `CallDialog` 接线；`ws.on('notify')` 加 call.* 分支。
9. [ ] 端到端：同局域网两浏览器 1:1 语音 → 视频跑通。

### Phase 2 — 通话记录落库 + 聊天展示
10. [ ] `streaming_call` 表（**先与用户确认 schema**）+ goctl model + 迁移注册（三库）。
11. [ ] `logic/callrecord.go`：结束写表 + 投 im `MType:call` 消息。
12. [ ] 前端通话气泡 + 点击回拨；i18n 文案。

### Phase 3 — 群组通话（SFU）
13. [ ] **SFU 真正接线**：服务端 PC `OnTrack` 收流、`ForwardMediaStream` 分发、订阅触发**重协商**下发 offer。
14. [ ] `logic/room.go`：Room.mode 调度（1:1 走 p2p relay，群组走 sfu）；peer-joined/left 广播。
15. [ ] 群振铃多成员；前端 SFU 网格 UI（动态增删画面）。
16. [ ] Redis Room 状态 + 忙线态完善（跨节点亲和字段预留，单节点先行）。

### Phase 4 — 预留扩展（非本期，仅占位/接口对齐）
- [ ] TURN + 短期凭证下发（`/ice-servers` 填充）。
- [ ] 会议控制 / 屏幕共享 / 直播 / 录制（基于 Room+SFU 扩展）。
- [ ] 跨 ws 节点路由。

---

## 10. 参考的现有模式

- [`common-notify-channel.md`](common-notify-channel.md) — 振铃信令复用 `commonNotifyTransfer` 全链路（投递→落库→ws 单推→前端 `notify` 分发）。
- [`conversation-send-authz-and-relation-cache.md`](conversation-send-authz-and-relation-cache.md) — 好友/群成员校验复用 `pkg/relationcache`（O(1) + fail-open + 读穿透）。
- `apps/im/ws/internal/handler/auth.go` — streaming ws 的 JWT 鉴权照搬（`?token=` 支持）。
- `apps/im/ws/internal/handler/push/relation_notify.go` — 给指定 uid 单推 + 自定义 `method` 的范式。
- `web/src/lib/ws-client.ts` / `chat-store.ts` — 前端 `ws.on(method)` 路由，加 call.* 监听。
- `apps/streaming/docs/streaming-one-to-one-call-flow-explained.md` 等 — 原 SFU 时序设计参考（本 spec 在其上改为 p2p/sfu 可切换）。

---

## 11. 测试计划

- [ ] `logic/call.go`：table-driven 单测——好友校验通过/拒绝、忙线拒绝、超时转 no_answer、各状态机迁移（真实 Redis，不 mock）。
- [ ] `streaming_call` model：增删查 + 三库 AutoMigrate 校验（真实 DB，测后清理，test-files 规则）。
- [ ] 信令 relay：单测两个假连接的 offer/answer/ice 转发正确、按 callId 隔离。
- [ ] SFU：单测 OnTrack 登记 + 订阅 AddTrack + 成员离开清理。
- [ ] 端到端手测（test-account 17585710998）：1:1 语音/视频接通、拒接、取消、超时、忙线、挂断后记录；群 3 人接通 + 中途进出。
- [ ] 弱网/拒绝授权/断线异常路径手测。

---

## 12. 待定事项

1. **`streaming_call` 表 schema**：需用户确认字段（[§8.2](#82-数据模型)）后才建。
2. **TURN**：coturn 部署时机与 `/ice-servers` 短期凭证算法（HMAC-SHA1 time-limited）。
3. **群通话记录形态**：群里展示一条“群通话 时长/人数”？还是每人各一条？默认：群会话内一条系统样式记录。
4. **跨 ws 节点路由**：沿用单节点限制；多节点时振铃与媒体的节点亲和（Redis `call:room:{id}=node` + Kafka 路由）独立 spec。
5. **1:1 是否在某些条件下强制 sfu**（录制/审核/NAT 失败兜底）：Room.mode 已支持，触发策略待定。
6. **并发上限 / 限流**：单用户同时通话数、SFU 单房间人数（现 config `MaxUsersPerRoom:50`）压测后定。

---

## 13. MVP 范围（本期交付边界）

✅ 1:1 语音+视频（p2p）端到端可用（STUN，同局域网/公网 IP）
✅ 多人群组通话（sfu）可用
✅ 通话记录落库 + 聊天气泡 + 回拨
✅ JWT 鉴权 + 好友/群成员校验 + 进 hichat2.sh
✅ 来电振铃 + 通话中完整交互（静音/摄像头/挂断/计时）
✅ Room mode（p2p/sfu）+ 信令协议为会议/直播/TURN 预留扩展点

❌ TURN 中继、会议控制、直播、屏幕共享、录制、跨节点路由、移动端（均预留，非本期）
