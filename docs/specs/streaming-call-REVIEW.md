# 音视频通话 — 实现进度与审核说明（Phase 0–2 交付）

> 给明早 review 用。分支 `feat-streaming-call`。本文件汇总本次交付了什么、做了哪些关键决策、怎么验证、已知问题、以及 Phase 3 的下一步计划。

## TL;DR

- ✅ **Phase 0–2 已完成并提交**，1:1 语音/视频通话**已端到端实测跑通**（视频双向音视频、语音双向音频）。
- ✅ **通话记录**：通话结束在聊天里留一条气泡（时长/已取消/已拒绝/未接），可点击回拨；**零 MySQL schema 变更**（复用聊天链路落 MongoDB）。
- ⏳ **Phase 3（群组 SFU）未动手**——这是个需要像 Phase 1 那样**交互式联调**的大块（WebRTC 重协商），盲写过夜大概率交付的是坏代码。已写好详细落地计划，建议白天开一个联调 session 做。详见 [§Phase 3 计划](#phase-3-群组通话-sfu--下一步计划未实现)。

## 提交记录（本次 session，时间倒序）

| commit | 内容 |
|---|---|
| `09e92ee` feat | 通话记录聊天展示 + 一键回拨（Phase 2） |
| `de195e8` fix | 语音通话补隐藏 audio 元素播放远端音频（之前只有视频通话有声） |
| `4c2d097` fix | 通话控制信令改走 streaming ws 直连，绕开 im ws 顶号churn（**关键修复**） |
| `3d6c9e1` chore | streaming 启动自检日志（确认以 uid=101 接入 im ws） |
| `e74f7a2` fix | streaming 以系统 uid=101 接入 im ws，不再复用真实用户 token（**根因修复**） |
| `77afc4a` fix | ICE 接口兜底返回默认 STUN + 前端加诊断日志 |
| `564ac08` fix | ICE 接口加 CORS + 前端 ICE 拉取加超时 |
| `41cad11` fix | 去掉振铃路径上阻塞的 GetUserById RPC（修复振铃丢失） |
| `6e1ae28` feat | 1:1 通话前端（call-engine/call-store/CallOverlay/CallDialog 接线） |
| `b9ffacb` feat | 1:1 信令后端（状态机 + relay + im ws push.call 路由） |
| `7d3b1e4` feat | 生产级接入（JWT ws 鉴权 / RPC client / ICE 接口 / hichat2.sh）（Phase 0） |
| `4173cbd`/`3f126e5` docs | spec 文档 |

## Phase 1 联调依次趟平的 6 个坑（都已修，记录备查）

1. **振铃丢失**：`handleInvite` 在发振铃前调了 `GetUserById` RPC 取昵称，RPC 卡住→振铃永不发出。改成昵称由前端用本地好友资料解析。
2. **系统 token 顶号（根因）**：streaming 用 Redis 里的「系统 root token」连 im ws，但那 token 实际是某真实用户(uid 11)的→streaming 和该用户在 im ws 抢同一 uid 槽位互相顶号、丢信令。改成用 JWT 密钥现签 uid=`SYSTEM_ROOT_UID`(101)。
3. **CORS**：streaming HTTP(:10093) 的 `/v1/streaming/ice-servers` 跨端口(web :3001)被浏览器拦。加 `Access-Control-Allow-Origin`。
4. **ICE 为空**：go-zero 配置加载不认 yaml tag，IceServers 读出来是空。handler 兜底返回公共 STUN；前端 ICE fetch 加 2.5s 超时不阻塞建连。
5. **信令走稳定通道**：accept 等控制信令原来都经 im ws，受 uid churn 影响时丢。改成**优先经对端自己的 streaming ws 直连**下发（稳定），只有初始 invite 走 im ws。
6. **语音没声**：通话界面只有视频通话渲染 `<video>`（视频元素会播音轨），语音通话没有任何媒体元素播放远端音频。补了个隐藏 `<audio>`，视频元素改 `muted` 防双声。

## 架构（最终形态）

```
呼叫控制（invite/accept/reject/cancel/end/timeout）
  主叫→streaming ws 发起；streaming 下发给对端：
    对端在 streaming ws 上 → 直连下发 type:call_signal（稳定，1:1 在用）
    对端不在（初始振铃）   → 经 im ws push.call → 前端 ws.on('call.signal')
媒体协商（offer/answer/ice）
  1:1 = streaming 纯 relay，两端浏览器 P2P 直连（媒体不过服务器）
通话记录
  结束由主叫端发一条 mType=10(call) 聊天消息 → 走 msgChatTransfer → MongoDB chatLog + 推给被叫
  双方会话内渲染通话气泡（时长/取消/拒绝/未接），点击回拨
```

关键文件：
- 后端 `apps/streaming/internal/handler/signaling.go`（信令+relay+notifyUser）、`internal/logic/call.go`（状态机，11 单测）、`internal/svc/servicecontext.go`（im ws 系统 token）、`apps/im/ws/internal/handler/push/call_notify.go`（im 侧 push.call 路由）。
- 前端 `web/src/lib/call-engine.ts`（WebRTC 引擎）、`call-store.ts`（zustand + 通话记录投递）、`components/im/CallOverlay.tsx`（来电/通话界面）、`ChatDetail.tsx`（通话气泡）。

## 怎么验证（明早）

```bash
# 起全套基础设施后：
./hichat2.sh
cd web && bun dev
```
1. 两个互为好友的账号分别在两个浏览器/标签页登录。
2. 会话页顶部点 📞/🎥 → 对端弹全屏来电 → 接听 → 双向音视频。
3. 挂断后，双方会话里出现一条通话记录气泡（如「通话时长 0:23」），点它回拨。
4. 试 拒接 / 取消 / 不接(30s 超时) → 分别出现「已拒绝/已取消/未接来电」。

> 自检：`grep "im ws auth as system uid" logs/streaming-streaming/*.log` 应为 `uid=101`。

## 已知问题（不阻塞通话，但要处理）

1. **前端 im ws (user 11) 每 2~3s 重连刷屏**：im 日志大量 `关闭旧连接 uid 11` + `use of closed network connection`。疑似 user 11 在多个标签页/上下文登录，或 dev 模式 React StrictMode 重复挂载 ws。**不影响通话**（通话信令已改走 streaming ws），但影响普通聊天可靠性 + 刷日志。待单独排查 `src/app/page.tsx:66` 的 initWs effect / 是否多端登录。
2. **同机跨浏览器(Chrome↔Safari) ICE 偶发**：本质是 mDNS（Chrome 把局域网 host 候选藏成 `.local`）。同机/同局域网目前靠 STUN 的 srflx 候选能通；彻底稳定要 TURN（本期留接口未部署）。
3. **i18n**：通话相关文案（CallOverlay 来电/通话界面、CallDialog、通话气泡、引擎错误提示）**已接入 `t()`**（中英词典在 `web/src/lib/i18n.ts` 的 `call.*` 段）。仅会话列表预览 `[语音通话]/[视频通话]`（`media-message.ts`）沿用该模块既有的硬编码风格（那一族 `[图片]/[视频]` 都没 i18n），保持一致。
4. **streaming `streaming_call` 统计表未建**：Phase 2 选了「仅聊天消息(MongoDB)」零 schema 变更方案。如需通话历史/统计独立表，待你确认字段后补（spec §8.2）。

## Phase 3（群组通话 SFU）— 下一步计划（未实现）

**为什么没盲写**：群组要 SFU——服务端为每个参与者建 pion PeerConnection，`OnTrack` 收流后转发给其他人，且**每次有人进/出都要 WebRTC 重协商**（服务端加 track→重新 offer→客户端 answer）。重协商是 WebRTC 最易错的部分，Phase 1 那个简单得多的 1:1 relay 都联调了好几轮。无法运行时验证的情况下盲写一大坨 SFU，交付的基本是坏代码，反而增加明早的排查负担。建议白天开一个联调 session 来做（像 Phase 1 那样写一点测一点）。

**落地计划（additive，不动 1:1 路径）**：
1. 后端 `sfu/sfu.go` 接线：参与者 join 时建服务端 PC；`pc.OnTrack` 收到流→建 `TrackLocalStaticRTP`、起 goroutine 读 RTP 写本地 track；把该 track `AddTrack` 到房间内其他参与者的 PC。
2. **重协商**：`AddTrack` 后服务端对受影响的 PC `CreateOffer`→`SetLocalDescription`→经 streaming ws 下发 offer；客户端 answer 回来 `SetRemoteDescription`。需处理 `OnNegotiationNeeded` / signalingState 竞态。
3. 新增群组信令：`group_invite`(多被叫振铃) / `group_join` / `group_leave`；房间用 `logic` 里新建 `GroupCallService`（host + 成员集 + 房间状态），与 1:1 的 `CallService` 并存。
4. 成员进出广播 `peer-joined`/`peer-left`，触发各端 UI 增删画面 + 服务端重协商。
5. 前端：`call-engine` 加 SFU 模式（与 SFU 服务器协商，而非 P2P relay）；`CallOverlay` 加 N 路视频网格；`CallDialog` 群组选成员已就绪（`onConfirm` 已留群组分支，目前提示「即将上线」）。
6. Redis 房间状态 + 跨节点亲和（设计预留，可单节点先行）。

**参考**：pion SFU 标准范式（`pion/webrtc` example-sfu-ws）。现有 `sfu/sfu.go` 已有 RoomSFU/UserSFU/tracks 骨架，但 `OnTrack`/转发/重协商未接线。

## 验收建议

先确认 Phase 1+2 在你的环境跑顺（尤其通话记录气泡 + 回拨），再决定：
- a) Phase 3 群组：约个时间一起联调（推荐）。
- b) 先把「已知问题 1」的 im ws 重连风暴查掉（影响日常聊天）。
- c) `streaming_call` 统计表要不要建（确认字段我就补）。
