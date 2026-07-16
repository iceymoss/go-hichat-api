# 群聊音视频通话 · 自建 SFU（pion/webrtc，in-process）

## 状态
- 创建日期: 2026-07-06
- 状态: **实施中（2026-07-16）— Phase 0 已完成（5 人浏览器实测）；Phase 1 代码完成、待公网验收；Phase 2 代码完成、待多人浏览器验收；Phase 3 分页完成，三层 simulcast 因浏览器 RID 暂停导致卡帧已回退为受控单层 VP8；Phase 4 路由压测完成、真实 30–50 PeerConnection 压测待完成**
- 分支: `feat-streaming-group-sfu-pion`（从 main 拉出）
- 关联: `apps/streaming`、`web/src`（IM 通话）、`deploy/`（coturn）
- 背景: 放弃"外挂 LiveKit server"路线（那让 streaming 退化成只签 token 的空壳），改为**核心媒体转发自己实现、跑在 streaming 进程内**，让开源项目的群聊音视频能力长在自己仓库里。

## 目标
在 `apps/streaming` 内用 **pion/webrtc v3** 自建 **SFU（Selective Forwarding Unit）**：每个参与者与 streaming 建**一条** PeerConnection，上行发布自己的音视频，服务端 `OnTrack` 收流后**按订阅关系把 RTP 转发**给房间内其他参与者。媒体（RTP）走 **浏览器 ↔ streaming**，不经任何外部媒体服务。目标稳定支持 **30–50 人**一通群通话，含 **受控视频质量 + 分页/按可见宫格订阅 + 服务端活跃发言人音频选路（active speaker）**。当前稳定路径为受控单层 VP8；simulcast 作为后续可选增强，不再是当前版本验收前置。系统需**公网可用**（含 coturn TURN 中继），且 **1:1 单聊 P2P 不动**。

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

### 3. 视频质量控制 + 分页/按可见宫格订阅
- 当前发布端使用受控单层 VP8，最高 `640x360@24fps/700kbps`，避免五人时每个客户端同时接收四路 720p 后超过带宽和解码临界点。
- 后端保留 RID publication generation、层选择和旧下行替换能力，但生产前端不主动发布 `f/h/q`。恢复 simulcast 前必须增加 RTP 层活跃度检测、失活降级和恢复，不能只按 RID 是否注册判断可用性。
- **分页订阅**：前端只订阅当前页可见宫格里的视频（`subscribe`/`unsubscribe` 帧），不可见的不下发；带宽随**可见数**而非总人数增长。默认仅语音进一步降载。

### 4. 公网可用（coturn / ICE / 端口）
- 新增 `coturn` 服务（docker-compose），下发到前端的 ICE 追加 `turn:` 条目（`iceserver.go`）。
- streaming 侧 pion `SettingEngine`：SFU 使用 **ICE Lite**（浏览器作为唯一 controlling agent），配置 **公网 IP（1:1 NAT）** 与 **UDP 媒体端口范围**（如 50000–50200），docker 暴露该范围；生产 `wss`。
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
5. 视频：默认关；谁开摄像头就以受控单层 VP8 上行，订阅者按可见宫格 `subscribe` 下发。后续若重新启用 simulcast，再由订阅策略选层。
6. 离开 `group_leave` / 断线：关该 PC、移除其下行轨（触发其余人 renegotiation）、会话态更新 + `broadcastGroupState`；房间空则销毁。

## 异常处理

| 场景 | 处理 |
|------|------|
| 房间满（≥上限） | `group_join` 返回满员错误 → 前端提示 |
| 忙线 | 现有 `IsBusy` 语义保留 |
| renegotiation glare | SFU 恒为 offerer，客户端只 answer；冲突时以 SFU offer 为准 |
| 新订阅者视频黑屏 | 订阅即发 PLI 请关键帧 |
| 丢包或接收端过载 | interceptor NACK 重传；当前通过单层分辨率/帧率/码率上限和分页限载。simulcast 恢复后再增加自动降层 |
| 对称 NAT / UDP 封 | coturn TCP/TLS 中继 |
| 发布者掉线 | `OnTrack` 收流泵 EOF → 关轨 → 通知订阅者移除 tile + renegotiation |
| 服务端过载 | 上限 + active-speaker 选路 + 分页订阅共同限载；超限拒绝加入 |
| 参与者只剩 1 人 | 结束会话 + `broadcastGroupState` |

## 可行性 / 负载分析（诚实）
- **音频**（默认场景）：30–50 人，方案 A 每客户端下行 ~5 路 opus（~150–250kbps），服务端转发流数 ≈ N×topN（线性，非 O(N²)）。**可行**。
- **视频**（按需开）：当前依靠受控单层 VP8 + 分页订阅，使带宽随可见宫格数而非总人数增长。若大量人同时开摄像头，仍受客户端总下行、解码能力和单节点带宽限制，需靠分页、发布上限和后续自适应继续压住。
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
  6. [x] audio-level 头扩展读取 + top-N 选路 + `active_speakers` 下发 + 前端高亮。代码与单元/race 测试完成，待真实多人持续说话切换验收。
- **Phase 3 — 视频质量控制 + 分页订阅**
  7. [ ] `subscribe/unsubscribe` 分页、前端每页本地 + 8 路远端视频、音频跨页持续播放已完成；发布端曾实现 VP8 `q/h/f` simulcast，但真实 Chrome 多人测试发现 inactive RID 可能持续发送 padding 且不结束 `TrackRemote`，导致 SFU 永久选择失活层。当前生产路径回退为受控单层 VP8（最高 `640x360@24fps/700kbps`），待 8–12 人翻页验收后决定是否以单层方案关闭本阶段，或另行设计带 RTP 活跃度检测的 simulcast。
- **Phase 4 — 压测到 30–50 + 调优**
  8. [ ] 路由热路径 2→8→16→30→50 阶梯脚本已完成；真实多 PeerConnection 的带宽/CPU/丢包/PLI 与 TURN 压测待完成。房间上限已落到可配 `MaxUsersPerRoom`（默认 50）。

**MVP = Phase 0 + Phase 1**：SFU 打通 + 公网可用（全订阅、默认语音+按需视频，验证到中等规模）。30–50 人 + active-speaker + 受控视频质量 + 分页是 Phase 2–4；simulcast 调整为非阻塞的后续增强。

## 参考的现有模式
- `apps/streaming/internal/handler/signaling.go` — 呼叫控制信令 + relay 收发框架（在其上加 SFU 信令；1:1 relay 保留）。
- `apps/streaming/internal/logic/groupcall.go` / `call.go` — 会话态/忙线/超时（复用语义）。
- `apps/streaming/sfu/sfu.go`、`webrtc/connection.go` — 现有骨架（重写，非复用）。
- `web/src/lib/group-call-engine.ts` / `peer-conn.ts` — 回调契约 + phase 状态机（新引擎对齐；PeerConn 收敛为一条）。
- `apps/streaming/internal/handler/iceserver.go` — ICE 下发（接 coturn）。

## 测试计划
- 诊断日志：浏览器通过鉴权后的 `sfu_diagnostic` 帧上报 WS/SDP/ICE/PC/restart/track 状态摘要，streaming 与服务端事件合并写入 `logs/sfu-diagnostics.jsonl`；按 `call_id`、`uid`、`session_id` 复盘多人时间线。禁止记录 JWT、完整 SDP、完整 ICE candidate 和媒体内容；默认 20 MiB 轮转为 `.1`。
- [x] SFU 单元：房间加/退、收流泵转发正确性、订阅关系变更后下行轨增减（table-driven，不依赖真实网络的部分用假 track/RTP 包）。
- [x] `GroupCallService` 上限/忙线/幂等（沿用 `groupcall_test.go` 风格）。
- [x] active-speaker：给定一串带 audio-level 的 RTP，top-N 选路正确、切换正确。
- [ ] 集成：本地起 streaming，用 pion 客户端（或浏览器）2→8→16 人加入，验证发布/订阅/renegotiation/PLI/离开清理。
- [ ] 公网：无公网 IP 环境经 coturn 中继成功。
- [x] 前端目标 SFU 文件测试与 eslint 无新增错误；全量 `tsc` 仅剩两个无关既有错误，详见下方进展记录。
- [x] streaming 范围 `go test`、race、vet、build 全绿；仓库全量测试仍在最终交付前执行。

### 2026-07-16 五人浏览器稳定性进展

当前已部署并验证受控单层 VP8，替代主动发布 `q/h/f` 三层 simulcast。关键提交：

- `19b9103 fix(web/call): keep simulcast transceiver bidirectional`：共享 SFU PeerConnection 的视频 transceiver 保持 `sendrecv`。
- `b206d24 fix(streaming/sfu): isolate video publication generations`：旧 publication generation 不能清理或污染替代媒体源。
- `545844a fix(web/call): rebuild camera tracks after toggle`：摄像头关闭使用 `replaceTrack(null)` 并停止旧轨，重新开启时获取新轨并在原 sender 上替换。
- `583718f fix(web/call): publish stable single-layer group video`：停止主动发布 `q/h/f`，恢复单层 VP8。
- `ed0b24e fix(streaming/sfu): stabilize five-participant video calls`：群视频限制为最高 `640x360@24fps/700kbps`；退订已离开的发布者改为幂等；新增 `server_error` 诊断。

已完成的真实浏览器验证：

- 最新单层五人 call：`gcall_1784210847031654000_1`。
- 每名发布者只有一条无 RID 的 `video/VP8` publication，证明单层版本已部署。
- 第五人加入时所有 SFU offer/answer 均在约 18–87 ms 内完成，ICE/PeerConnection 保持 `connected`，无新的 `negotiation_error`、ICE restart 或媒体连接重建。
- 旧三层 call `gcall_1784208963578564000_1` 证明 inactive `q` 可在 `TrackRemote.ReadRTP()` 不结束时只增加 RTP packet、媒体字节和 `frames_decoded` 不再增长；PLI 无法恢复该层。这是停止主动 simulcast 的直接证据。
- 单层初版仍允许摄像头升到 `1280x720@30fps`。第五人加入后，每个浏览器最多同时解码四路 720p，单路可达约 4–5 Mbps，总下行约 15–25 Mbps；日志出现 `frames_per_second` 降至 1–9、`jitter_ms` 升至 35–75、`freeze_count` 增长而 `packets_lost=0`，表明客户端带宽/解码队列达到临界点。
- 当前限制为单层 `640x360`、最高 24 FPS、最高 700 kbps；用户反馈相比此前已有质的稳定性提升。该限制仍需一次新部署后的五人日志复核，并继续扩展到 8–12 人。

错误弹窗结论：

- 最新五人 call 没有与“通话失败/通话出错”对应的协商或 ICE 故障。
- 旧客户端将服务端所有 `error` 帧无条件映射为 `call.err.generic`，且不记录服务端原始原因。
- 已确认一个正常清理竞争：发布者离开后，其他客户端发送迟到的 `unsubscribe`，服务端因发布者已删除返回 `not a group participant`，前端误报整通电话失败。
- `ed0b24e` 已将该 unsubscribe 路径改为幂等，并将后续服务端错误记录为 `server_error`，字段包含错误摘要、通话 phase 和 signaling state。若再出现弹窗，应优先按 `call_id` 查该事件，不再从 ICE/SDP 猜测。

当前自动验证结果：

- `bun test src/lib/sfu-group-engine.test.ts`：17 pass，0 fail。
- 目标 SFU 前端文件 ESLint 通过。
- `go test ./apps/streaming/... -count=1` 通过。
- `go test -race ./apps/streaming/sfu ./apps/streaming/internal/handler -count=1` 通过。
- `go vet ./apps/streaming/...`、`go build ./apps/streaming/...`、`git diff --check` 通过。
- 全量 `bun x tsc --noEmit` 仅被两个与 SFU 无关的既有错误阻塞：`MomentsFeed.tsx` 的 `aspectSquare` 和 `ProfilePage.tsx` 的 `ringColor`。

下一次从本 spec 继续时按以下顺序执行：

1. 同时部署包含 `ed0b24e` 的 web 和 streaming，所有测试浏览器硬刷新，创建全新五人视频 call。
2. 确认 `track_published` 仍为每人一条无 RID 的 `video/VP8`，`video_inbound_stats` 分辨率不超过 `640x360`，FPS 接近 24 且不再持续增长 `video_stalled`。
3. 每人至少执行两轮摄像头关闭/开启，确认 `camera_toggle_requested`、`camera_state_changed` 成对出现且无 `camera_toggle_failed`，远端解码帧恢复增长。
4. 若出现通话错误弹窗，直接检查同一 `call_id` 的 `server_error`；根据原始错误区分房间上限、publish、subscribe 或 reconnect，不把订阅维护错误当成整通电话失败。
5. 扩展到 8–12 人并验证翻页：当前页视频创建、上一页视频退订、跨页音频持续、快速翻页无协商错误。Phase 3 是否以受控单层 VP8 关闭应以该结果决定。
6. 多人持续交替讲话，验收 top-4 active speaker 音频连续性、排序和前端高亮。
7. 在真实公网环境验收跨 NAT、TURN UDP/TCP/TLS、生产 `wss` 和防火墙媒体端口。
8. 最后执行真实 8→16→30→50 条 PeerConnection 阶梯压测，记录 CPU、内存、上下行带宽、丢包、NACK/PLI 和 TURN 开销；不得用 Room benchmark 代替端到端结论。

### 2026-07-16 路由热路径基准

命令：`./apps/streaming/loadtest.sh`。模型为每个参与者发布一路音频和一路视频，服务端只转发 top-4 音频，每个视频发布者仅有 8 个可见订阅者；测试同时断言 50 人每媒体周期恰好产生 `4×49 + 50×8 = 596` 次下行写入，防止策略回退为全订阅。

Apple M1 Pro / darwin arm64 实测：

| 参与者 | 约 ns/op | B/op | allocs/op |
|-------:|----------:|-----:|----------:|
| 2 | 662–674 | 6 | 2 |
| 8 | 3,438–3,699 | 1,370 | 20 |
| 16 | 7,123–7,485 | 3,060 | 36 |
| 30 | 13,006–13,529 | 5,858 | 64 |
| 50 | 21,596–22,863 | 10,148 | 104 |

该结果仅证明 Room RTP 路由、top-N 和分页 fan-out 随人数近似线性，不能替代真实 DTLS/SRTP、ICE/TURN、浏览器编解码、网络丢包和 30–50 条 PeerConnection 的端到端验收。

## 风险
- 自建 SFU 的 renegotiation glare、simulcast 层选择、PLI 风暴、带宽自适应是硬难点，Phase 3/4 存在返工风险。
- 公网 coturn/端口/防火墙配置是常见踩坑点（Phase 1 预留联调时间）。
- 单节点 CPU 是 30–50 人视频大量开摄像头时的天花板，靠分页 + 默认语音缓解。

## 待定事项（已确认默认值，yaml 仍可配）
- top-N 发言人：**默认 N=4**（3–5 区间取中），`ActiveSpeakers` yaml 可配。
- 群通话上限：**默认 50**，`MaxParticipants` yaml 可配。
- 视频发布策略：当前默认使用受控单层 VP8（最高 `640x360@24fps/700kbps`）。`f/h/q` 三层代码曾完成但已从生产发布路径撤下；只有在实现 RTP 层活跃度检测、失活降级和恢复策略并通过多人浏览器验收后，才重新启用 simulcast。
- 是否需要通话历史落库（涉及新表，按 `database-model.md` 单独确认）——**留后续，本 spec 不做**。
- 前端 active speaker UI 形态（上浮 / 边框 / 大小画中画）——实现到 Phase 2 时前端定。

## 交付范围（全量）
**本 spec 要求 Phase 0→4 全部完成**（自建 pion SFU、公网可用、30–50 人、active-speaker 音频、受控视频质量、分页订阅）。分阶段**只是中途测试检查点**，不是交付到 Phase 0+1 就停。三层 simulcast 不再是当前版本关闭 spec 的硬前置条件，但若重新启用，必须先解决 inactive RID 生命周期和自动降级/恢复：
- 每个 Phase 独立 TDD + commit + 阶段性联调，作为可测里程碑。
- Phase 0+1（SFU 打通 + 公网可用）是**第一个可跑通的里程碑**，之后继续推进到 30–50 人达标的完整形态。
- 分阶段隔离返工面：某阶段（如 simulcast 层选择）需调整时，影响限在该 Phase，不推翻已测好的收流泵 / 信令底座。
