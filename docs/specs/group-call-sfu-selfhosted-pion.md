# 群聊音视频通话 · 自建 SFU（pion/webrtc，in-process）

## 状态
- 创建日期: 2026-07-06
- 状态: **实施中（2026-07-16 接手）— Phase 0 已完成；Phase 1 代码完成、待公网验收；Phase 2→4 待完成**
- 分支: `feat-streaming-group-sfu-pion`（从 main 拉出）
- 关联: `apps/streaming`、`web/src`（IM 通话）、`deploy/`（coturn）
- 背景: 放弃"外挂 LiveKit server"路线（那让 streaming 退化成只签 token 的空壳），改为**核心媒体转发自己实现、跑在 streaming 进程内**，让开源项目的群聊音视频能力长在自己仓库里。

## 目标
在 `apps/streaming` 内用 **pion/webrtc v3** 自建 **SFU（Selective Forwarding Unit）**：每个参与者与 streaming 建**一条** PeerConnection，上行发布自己的音视频，服务端 `OnTrack` 收流后**按订阅关系把 RTP 转发**给房间内其他参与者。媒体（RTP）走 **浏览器 ↔ streaming**，不经任何外部媒体服务。目标稳定支持 **30–50 人**一通群通话，含 **simulcast（视频多码率上行）+ 分页/按可见宫格订阅 + 服务端活跃发言人音频选路（active speaker）**。**公网可用**（含 coturn TURN 中继）。**1:1 单聊 P2P 不动。**

## 非目标
- 不外挂 LiveKit / mediasoup / Janus 等外部 SFU 服务（本 spec 的全部意义就是自建）。
- 不动 1:1 单聊：2 人继续走现有 `call-engine.ts` / `CallService` 的浏览器直连 P2P。
- 不做录制、直播推流、屏幕共享、白板（后续再议）。
- 不做通话历史落库（涉及新表，需按 `database-model.md` 单独确认）。
- 不追求 >50 人 / 多 SFU 节点级联（单节点 30–50 为本轮上限；多节点留将来）。

## 用户故事
- 作为群成员，我想发起一个 30–50 人的群语音（默认）通话，所有人的声音经服务端转发互通，而不是卡在 4 人 mesh。
- 作为参与者，我想按需打开摄像头，且在几十人里只清晰看到当前宫格里可见的几个人的视频（分页订阅），不因人多而跑满带宽。
- 作为参与者，我希望"谁在说话"自动高亮/上浮。
- 作为自部署者，我希望这套在公网真实可用（经 coturn 穿透对称 NAT），docker-compose 一键起，且**核心在本仓库里**、可读可改。

## 现状（改造基线）

| 层 | 现状文件 | 现状 | 改造 |
|----|---------|------|------|
| 后端 SFU 骨架 | `apps/streaming/sfu/sfu.go` | **空壳**：往 map 塞 `TrackLocalStaticRTP`，无 `OnTrack` 收流、无 RTP 转发泵、无 renegotiation/PLI/simulcast，从未接入信令 | **重写**为真正的 SFU 引擎 |
| 后端 WebRTC 封装 | `apps/streaming/webrtc/connection.go` | 残缺连接封装，未接入 | 重写为"每参与者一条 PeerConnection + 发布/订阅"封装 |
| 后端会话态 | `apps/streaming/internal/logic/groupcall.go` | mesh 会话态，`maxPart=4` | 复用会话/邀请/忙线/上限语义（上限提到可配 50），去掉"两两建连"含义，改挂 SFU 房间 |
| 后端信令 | `apps/streaming/internal/handler/signaling.go` | `group_invite/join/leave` + `relayToPeer`（两两 offer/answer/ice） | 保留呼叫控制；**群媒体协商改为 client↔SFU 单条 PC**：新增 SFU 上/下行信令（publish/subscribe/renegotiate），移除群 mesh 的两两 relay（1:1 relay 保留） |
| 后端 ICE 下发 | `apps/streaming/internal/handler/iceserver.go` | 下发 STUN | 追加 coturn TURN 条目（公网穿透） |
| 前端引擎 | `web/src/lib/group-call-engine.ts`、`peer-conn.ts` | N 条 PeerConn 两两 P2P | 重写为 `sfu-group-engine.ts`：**一条** PC 连 SFU，发布本地轨、按需订阅远端 |
| 前端 UI | `web/src/components/im/CallOverlay.tsx` | 群网格渲染各 peer MediaStream | 渲染 SFU 下发的各远端轨 + 分页订阅 + active speaker 高亮 |
| 部署 | `docker-compose.yaml`、`deploy/` | 无 TURN | 新增 `coturn`；streaming 暴露媒体 UDP 端口范围 + 公网 IP 配置 |

**复用点**：呼叫控制链路（im ws `push.call` 振铃 → `group.invite` → 接听/拒接/挂断/忙线/45s 兜底 → `broadcastGroupState` 群横幅）**完全保留**，只换"媒体怎么连/转发"。

## 核心架构（SFU 数据面）

```
浏览器A ──上行(pub)──┐                        ┌── 下行(sub) ──▶ 浏览器B
   ▲  下行(sub) ◀───┤   streaming 进程内 SFU  ├── 下行(sub) ──▶ 浏览器C
   │                │                        │
浏览器A 一条 PC ────┤  OnTrack 收 A 的 RTP    │  每个订阅者一条 PC
                    │  → 写进订阅者的下行本地轨 │  上行=发布, 下行=按需订阅
                    └────────────────────────┘
```

每个参与者与 streaming 建 **1 条 `webrtc.PeerConnection`**：
- **上行（发布）**：浏览器 `addTrack(mic[, cam])` → 服务端 `pc.OnTrack` 拿到 `*TrackRemote`。
- **收流泵**：每个 `OnTrack` 起一个 goroutine 循环 `track.ReadRTP()`，把包写进"该发布者对应的下行本地轨"`*TrackLocalStaticRTP`（每个订阅者各持一份，或共享 fan-out）。
- **下行（订阅）**：给订阅者的 PC `AddTrack(下行本地轨)`；`AddTrack` 触发 `negotiationneeded` → **服务端发起 renegotiation**（新 offer → 客户端 answer）。
- **关键帧**：新订阅者加入时，服务端向发布者发 **RTCP PLI**（Picture Loss Indication）请求关键帧，避免新订阅者黑屏等待。
- **上行纠错**：启用 pion interceptor 的 **NACK/重传** 与 **RTCP 反馈**。

## 关键技术点（自建 SFU 的硬骨头）

### 1. Renegotiation（每次订阅关系变化都要重新协商）
mesh 是两两一次性协商；SFU 里**每加/减一路订阅轨都改变 SDP**，服务端必须能主动发起 renegotiation。信令需支持"服务端 → 客户端"方向的 offer。用 pion `pc.OnNegotiationNeeded` + 手动 `CreateOffer/SetLocalDescription` → 经 streaming ws 下发 `sfu_offer`，客户端 `setRemote+answer` 回 `sfu_answer`。需处理 **glare**（双方同时 offer）：约定 SFU 永远是 offerer，客户端只 answer。

### 2. 活跃发言人音频选路（方案 A · 服务端 top-N）
- 客户端发布音频时带 **RFC 6464 audio-level RTP 头扩展**（`urn:ietf:params:rtp-hdrext:ssrc-audio-level`）。
- 服务端**不解码**，直接读每包头扩展里的音量值，滑动窗口维护每个发布者的近期音量。
- 周期性（如每 200ms）选出音量 **top-N（默认 3–5，可配）** 发言人，**只把这几路音频转发**给所有订阅者；其余音频路暂停转发。
- 切换时对新上榜者发 PLI 不适用（音频无关键帧），直接开始转发即可；对下榜者停止写包。
- 产出的 top-N 名单顺带经 ws `active_speakers` 帧下发前端做高亮/上浮。
- **好处**：30–50 人每客户端只收 ~5 路音频而非 49 路，带宽/CPU 降一个数量级。

### 3. 视频 simulcast + 分页/按可见宫格订阅
- 发布端 `livekit` 不用，用浏览器原生 simulcast：`RTCRtpSender` 的 `sendEncodings`（如 f/h/q 三档）上行多码率。
- 服务端对每个订阅者、每个可见发布者，**选一档**转发（大画面选高码率、缩略选低码率），用 pion 的 simulcast track 支持（`OnTrack` 会按 rid 拿到多条）。
- **分页订阅**：前端只订阅当前页可见宫格里的视频（`subscribe`/`unsubscribe` 帧），不可见的不下发；带宽随**可见数**而非总人数增长。默认仅语音进一步降载。

### 4. 公网可用（coturn / ICE / 端口）
- 新增 `coturn` 服务（docker-compose），下发到前端的 ICE 追加 `turn:` 条目（`iceserver.go`）。
- streaming 侧 pion `SettingEngine`：配置 **公网 IP（1:1 NAT）** 与 **UDP 媒体端口范围**（如 50000–50200），docker 暴露该范围；生产 `wss`。
- 对称 NAT / UDP 被封时经 coturn TCP/TLS 中继兜底。

## 信令契约（streaming ws；附加，1:1 帧不动）

| 帧 | 方向 | 用途 |
|----|------|------|
| `group_invite` / `group_join` / `group_leave` | 上行 | **不变**（复用现有呼叫控制） |
| `group_created` / `group_roster` | 下行 | 复用；不再带媒体，仅会话名单 |
| `sfu_publish` | 上行 | 客户端声明要发布（含 SDP offer 或作为 renegotiation 触发） |
| `sfu_offer` | 下行 | **服务端**发起的 offer（renegotiation：新增/移除下行轨时） |
| `sfu_answer` | 上行 | 客户端对 `sfu_offer` 的 answer |
| `sfu_ice` | 双向 | ICE candidate 交换（trickle） |
| `subscribe` / `unsubscribe` | 上行 | 分页订阅：订阅/退订某发布者的视频 |
| `active_speakers` | 下行 | 当前 top-N 发言人名单（高亮） |
| 群 `offer/answer/ice`（旧两两 relay） | — | **群通话下线**（1:1 保留） |

## 核心流程

1. A `group_invite` → 后端建 SFU 房间（room=callID）+ 现有振铃逻辑不变。
2. A 建一条 PC → `sfu_publish`（offer，含 mic[+cam]）→ 服务端 answer；`OnTrack` 起收流泵。
3. B 接听 `group_join` → 建 PC 发布自己 → 服务端把 A 的下行轨 `AddTrack` 到 B、B 的加到 A → 各自 renegotiation。
4. 音频：服务端持续算 top-N，只转发活跃发言人；`active_speakers` 高亮。
5. 视频：默认关；谁开摄像头就 simulcast 上行，订阅者按可见宫格 `subscribe` 选档下发。
6. 离开 `group_leave` / 断线：关该 PC、移除其下行轨（触发其余人 renegotiation）、会话态更新 + `broadcastGroupState`；房间空则销毁。

## 异常处理

| 场景 | 处理 |
|------|------|
| 房间满（≥上限） | `group_join` 返回满员错误 → 前端提示 |
| 忙线 | 现有 `IsBusy` 语义保留 |
| renegotiation glare | SFU 恒为 offerer，客户端只 answer；冲突时以 SFU offer 为准 |
| 新订阅者视频黑屏 | 订阅即发 PLI 请关键帧 |
| 丢包 | interceptor NACK 重传；持续丢包降级到低 simulcast 层 |
| 对称 NAT / UDP 封 | coturn TCP/TLS 中继 |
| 发布者掉线 | `OnTrack` 收流泵 EOF → 关轨 → 通知订阅者移除 tile + renegotiation |
| 服务端过载 | 上限 + active-speaker 选路 + 分页订阅共同限载；超限拒绝加入 |
| 参与者只剩 1 人 | 结束会话 + `broadcastGroupState` |

## 可行性 / 负载分析（诚实）
- **音频**（默认场景）：30–50 人，方案 A 每客户端下行 ~5 路 opus（~150–250kbps），服务端转发流数 ≈ N×topN（线性，非 O(N²)）。**可行**。
- **视频**（按需开）：simulcast + 分页订阅后，带宽随可见宫格数（如 9–16）而非总人数增长。若少数人开摄像头，压力可控；若大量人同时开 + 大量订阅，则受单节点 CPU（转发+PLI）限制，需靠分页 + 层选择压住。
- **单节点 30–50 是本轮现实上限**；再往上要多 SFU 节点级联，明确留将来。
- **风险**：自建 SFU 的 renegotiation/simulcast/PLI 风暴/带宽治理是真难点（业界 3–6 人月的来源）；本 spec 用"降到 30–50 + 默认语音 + top-N 音频 + 分页视频"把范围收敛到可交付。

## 实现步骤（分阶段，每阶段可用里程碑，走 `/tdd`）

- **Phase 0 — SFU 数据面打通（最核心）**
   1. [x] 重写 `sfu/`：房间/参与者模型 + 每 PC 的发布(`OnTrack`)→收流泵→订阅者下行轨 fan-out。
   2. [x] renegotiation：`sfu_offer/answer/ice` 信令 + 服务端 offerer；PLI 关键帧。
   3. [x] 前端 `sfu-group-engine.ts`：一条 PC 连 SFU、发布、订阅、renegotiation；对齐现有回调契约。
  4. [ ] 本地验证 ~8–12 人**全订阅**（音频 + 按需视频）双向通。
- **Phase 1 — 公网可用**
   5. [ ] `coturn` 部署（docker-compose）+ `iceserver.go` 下发 TURN；pion `SettingEngine` 公网 IP + UDP 端口范围；`wss`。代码与 Compose 已完成，跨 NAT/TURN TLS 实测待验收。
- **Phase 2 — 活跃发言人音频（方案 A）**
  6. [ ] audio-level 头扩展读取 + top-N 选路 + `active_speakers` 下发 + 前端高亮。
- **Phase 3 — 视频 simulcast + 分页订阅**
  7. [ ] 发布端 simulcast；服务端按 rid 选层；`subscribe/unsubscribe` 分页；前端按可见宫格订阅。
- **Phase 4 — 压测到 30–50 + 调优**
  8. [ ] 多端/压测脚本，带宽/CPU/丢包/PLI 治理达标；上限落到房间层。

**MVP = Phase 0 + Phase 1**：SFU 打通 + 公网可用（全订阅、默认语音+按需视频，验证到中等规模）。30–50 人 + active-speaker + simulcast + 分页是 Phase 2–4。

## 参考的现有模式
- `apps/streaming/internal/handler/signaling.go` — 呼叫控制信令 + relay 收发框架（在其上加 SFU 信令；1:1 relay 保留）。
- `apps/streaming/internal/logic/groupcall.go` / `call.go` — 会话态/忙线/超时（复用语义）。
- `apps/streaming/sfu/sfu.go`、`webrtc/connection.go` — 现有骨架（重写，非复用）。
- `web/src/lib/group-call-engine.ts` / `peer-conn.ts` — 回调契约 + phase 状态机（新引擎对齐；PeerConn 收敛为一条）。
- `apps/streaming/internal/handler/iceserver.go` — ICE 下发（接 coturn）。

## 测试计划
- 诊断日志：浏览器通过鉴权后的 `sfu_diagnostic` 帧上报 WS/SDP/ICE/PC/restart/track 状态摘要，streaming 与服务端事件合并写入 `logs/sfu-diagnostics.jsonl`；按 `call_id`、`uid`、`session_id` 复盘多人时间线。禁止记录 JWT、完整 SDP、完整 ICE candidate 和媒体内容；默认 20 MiB 轮转为 `.1`。
- [ ] SFU 单元：房间加/退、收流泵转发正确性、订阅关系变更后下行轨增减（table-driven，不依赖真实网络的部分用假 track/RTP 包）。
- [ ] `GroupCallService` 上限/忙线/幂等（沿用 `groupcall_test.go` 风格）。
- [ ] active-speaker：给定一串带 audio-level 的 RTP，top-N 选路正确、切换正确。
- [ ] 集成：本地起 streaming，用 pion 客户端（或浏览器）2→8→16 人加入，验证发布/订阅/renegotiation/PLI/离开清理。
- [ ] 公网：无公网 IP 环境经 coturn 中继成功。
- [ ] 前端 `bun run tsc` / eslint 无新增错误。
- [ ] 后端 `go test ./... -count=1` 全绿、`go build ./...` 无死引用。

## 风险
- 自建 SFU 的 renegotiation glare、simulcast 层选择、PLI 风暴、带宽自适应是硬难点，Phase 3/4 存在返工风险。
- 公网 coturn/端口/防火墙配置是常见踩坑点（Phase 1 预留联调时间）。
- 单节点 CPU 是 30–50 人视频大量开摄像头时的天花板，靠分页 + 默认语音缓解。

## 待定事项（已确认默认值，yaml 仍可配）
- top-N 发言人：**默认 N=4**（3–5 区间取中），`ActiveSpeakers` yaml 可配。
- 群通话上限：**默认 50**，`MaxParticipants` yaml 可配。
- simulcast 档位：**默认 f/h/q 三档**（full/half/quarter），默认订阅层由可见宫格大小选（大画面 f、缩略 q）。
- 是否需要通话历史落库（涉及新表，按 `database-model.md` 单独确认）——**留后续，本 spec 不做**。
- 前端 active speaker UI 形态（上浮 / 边框 / 大小画中画）——实现到 Phase 2 时前端定。

## 交付范围（全量）
**本 spec 要求 Phase 0→4 全部完成**（自建 pion SFU、公网可用、30–50 人、active-speaker 音频、simulcast、分页订阅）。分阶段**只是中途测试检查点**，不是交付到 Phase 0+1 就停：
- 每个 Phase 独立 TDD + commit + 阶段性联调，作为可测里程碑。
- Phase 0+1（SFU 打通 + 公网可用）是**第一个可跑通的里程碑**，之后继续推进到 30–50 人达标的完整形态。
- 分阶段隔离返工面：某阶段（如 simulcast 层选择）需调整时，影响限在该 Phase，不推翻已测好的收流泵 / 信令底座。
