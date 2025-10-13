# 一对一音视频通话完整流程文档

> **文档版本**: v1.0  
> **更新日期**: 2025-10-13  
> **适用服务**: streaming 流媒体服务

## 📋 目录

- [概述](#概述)
- [完整时序图](#完整时序图)
- [阶段详解](#阶段详解)
  - [阶段1: 通话邀请](#阶段1-通话邀请)
  - [阶段2: 加入房间](#阶段2-加入房间)
  - [阶段3: WebRTC连接协商](#阶段3-webrtc连接协商)
  - [阶段4: ICE候选交换](#阶段4-ice候选交换)
  - [阶段5: 媒体传输建立](#阶段5-媒体传输建立)
  - [阶段6: 媒体控制](#阶段6-媒体控制)
  - [阶段7: 通话结束](#阶段7-通话结束)
- [核心数据结构](#核心数据结构)
- [消息统计](#消息统计)
- [关键要点](#关键要点)

---

## 概述

一对一音视频通话采用 **WebRTC + WebSocket** 的标准实现方案：

- **WebSocket**: 用于信令传输（连接协商、状态同步）
- **WebRTC**: 用于音视频数据的 P2P 传输
- **架构模式**: "先邀请后连接"的标准流程

**核心组件**:
- **CallManager**: 管理通话生命周期
- **RoomManager**: 管理房间和用户
- **SignalingServer**: 转发所有信令消息
- **SFU**: 选择性转发单元（可选，用于群组通话）

---

## 完整时序图

```
┌─────────────┐              ┌────────────────┐              ┌─────────────┐
│   用户A     │              │  信令服务器     │              │   用户B     │
│  (主叫方)   │              │ SignalingServer│              │  (被叫方)   │
└──────┬──────┘              └────────┬───────┘              └──────┬──────┘
       │                              │                              │
       
═══════════════════════════════════════════════════════════════════════════
  阶段1: 通话邀请 (CallManager 管理)
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │ 【A主动】①                    │                              │
       │ call_invite                  │                              │
       ├─────────────────────────────>│                              │
       │  发起通话邀请                 │                              │
       │                              │                              │
       │                              ├─ CreateOneToOneCall()        │
       │                              │  • 检查双方是否在通话中         │
       │                              │  • 生成 call_id               │
       │                              │  • 生成 room_id               │
       │                              │  • 状态: inviting             │
       │                              │                              │
       │                              │ 【服务器转发】②               │
       │                              │ call_invite                  │
       │                              ├─────────────────────────────>│
       │                              │  转发邀请(含call_id/room_id)  │
       │                              │                              │
       │                              │                              │ 【B收到邀请】
       │                              │                              │ 显示来电界面
       │                              │                              │ 用户点击"接受"
       │                              │                              │
       │                              │ 【B主动】③                    │
       │                              │ call_accept                  │
       │                              │<─────────────────────────────┤
       │                              │  接受通话                     │
       │                              │                              │
       │                              ├─ AcceptCall()                │
       │                              │  • 更新状态: connected        │
       │                              │  • 记录 started_at            │
       │                              │                              │
       │ 【服务器通知】④               │                              │
       │ call_accept                  │                              │
       │<─────────────────────────────┼─────────────────────────────>│
       │  通知双方通话已建立            │                              │
       │                              │                              │

═══════════════════════════════════════════════════════════════════════════
  阶段2: 加入房间 (RoomManager 管理)
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │ 【A主动】⑤                    │                              │
       │ join_room                    │                              │
       ├─────────────────────────────>│                              │
       │  加入通话房间                 │                              │
       │                              │                              │
       │                              ├─ handleJoinRoom()            │
       │                              │  ① 创建/获取房间              │
       │                              │  ② 创建User对象               │
       │                              │  ③ 添加用户到房间             │
       │                              │  ④ 创建WebRTC连接            │
       │                              │  ⑤ 保存连接映射              │
       │                              │  ⑥ 添加到SFU                │
       │                              │                              │
       │ ⑥ room_info                  │                              │
       │<─────────────────────────────┤                              │
       │  返回房间信息                 │                              │
       │                              │                              │
       │                              │ 【B主动】⑦                    │
       │                              │ join_room                    │
       │                              │<─────────────────────────────┤
       │                              │  加入通话房间                 │
       │                              │                              │
       │                              │ ⑧ room_info                  │
       │                              ├─────────────────────────────>│
       │                              │  返回房间信息                 │
       │                              │                              │
       │ ⑨ user_joined                │                              │
       │<─────────────────────────────┤                              │
       │  通知A: B加入了房间           │                              │
       │                              │                              │

═══════════════════════════════════════════════════════════════════════════
  阶段3: WebRTC连接协商 (SDP 交换)
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │ 【A主动】⑩                    │                              │
       │ offer                        │                              │
       ├─────────────────────────────>│                              │
       │  发送SDP Offer               │                              │
       │  (媒体能力描述)               │                              │
       │                              │                              │
       │                              │ ⑪ offer (转发)               │
       │                              ├─────────────────────────────>│
       │                              │                              │
       │                              │                              │ 【B处理Offer】
       │                              │                              │ setRemoteDescription
       │                              │                              │ createAnswer
       │                              │                              │
       │                              │ 【B主动】⑫                    │
       │                              │ answer                       │
       │                              │<─────────────────────────────┤
       │                              │  发送SDP Answer              │
       │                              │  (确认媒体配置)               │
       │                              │                              │
       │ ⑬ answer (转发)              │                              │
       │<─────────────────────────────┤                              │
       │                              │                              │
       │ 【A处理Answer】               │                              │
       │ setRemoteDescription         │                              │
       │                              │                              │

═══════════════════════════════════════════════════════════════════════════
  阶段4: ICE 候选交换 (网络路径协商)
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │ 【A主动】⑭                    │                              │
       │ ice_candidate #1             │                              │
       ├─────────────────────────────>│                              │
       │  发送网络候选                 │                              │
       │                              │                              │
       │                              │ ⑮ ice_candidate #1 (转发)    │
       │                              ├─────────────────────────────>│
       │                              │                              │
       │ 【A主动】⑯                    │                              │
       │ ice_candidate #2             │                              │
       ├─────────────────────────────>│                              │
       │  (可能有多个候选)              │                              │
       │                              │                              │
       │                              │ ⑰ ice_candidate #2 (转发)    │
       │                              ├─────────────────────────────>│
       │                              │                              │
       │                              │ 【B主动】⑱                    │
       │                              │ ice_candidate #1             │
       │                              │<─────────────────────────────┤
       │                              │  发送网络候选                 │
       │                              │                              │
       │ ⑲ ice_candidate #1 (转发)    │                              │
       │<─────────────────────────────┤                              │
       │                              │                              │
       │                              │ 【B主动】⑳                    │
       │                              │ ice_candidate #2             │
       │                              │<─────────────────────────────┤
       │                              │  (可能有多个候选)              │
       │                              │                              │
       │ ㉑ ice_candidate #2 (转发)    │                              │
       │<─────────────────────────────┤                              │
       │                              │                              │

═══════════════════════════════════════════════════════════════════════════
  阶段5: 媒体传输建立
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │═══════════════════════════════════════════════════════════│
       │          WebRTC P2P 连接建立成功                            │
       │          • ICE 连接状态: connected                          │
       │          • 开始传输音视频数据                                │
       │          • 数据不经过信令服务器                              │
       │          • 使用 DTLS/SRTP 加密                              │
       │═══════════════════════════════════════════════════════════│
       │                              │                              │

═══════════════════════════════════════════════════════════════════════════
  阶段6: 媒体控制 (可选)
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │ 【A主动】㉒                    │                              │
       │ mute                         │                              │
       ├─────────────────────────────>│                              │
       │  静音                        │                              │
       │                              │                              │
       │                              │ ㉓ mute (广播)               │
       │                              ├─────────────────────────────>│
       │                              │  通知B: A静音了              │
       │                              │                              │
       │                              │ 【B主动】㉔                    │
       │                              │ video_off                    │
       │                              │<─────────────────────────────┤
       │                              │  关闭视频                     │
       │                              │                              │
       │ ㉕ video_off (广播)           │                              │
       │<─────────────────────────────┤                              │
       │  通知A: B关闭视频             │                              │
       │                              │                              │

═══════════════════════════════════════════════════════════════════════════
  阶段7: 通话结束
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │                              │ 【B主动】㉖                    │
       │                              │ call_end                     │
       │                              │<─────────────────────────────┤
       │                              │  结束通话                     │
       │                              │                              │
       │                              ├─ EndCall()                   │
       │                              │  • 更新状态: ended            │
       │                              │  • 计算通话时长                │
       │                              │  • 清理用户映射                │
       │                              │                              │
       │ ㉗ call_end (通知双方)        │                              │
       │<─────────────────────────────┼─────────────────────────────>│
       │  通话已结束                   │                              │
       │                              │                              │
       │ 【A主动】㉘                    │                              │
       │ leave_room                   │                              │
       ├─────────────────────────────>│                              │
       │  离开房间                     │                              │
       │                              │                              │
       │                              ├─ handleLeaveRoom()           │
       │                              │  • 从房间移除用户              │
       │                              │  • 从SFU移除                  │
       │                              │  • 关闭WebRTC连接             │
       │                              │  • 删除连接映射                │
       │                              │                              │
       │                              │ 【B主动】㉙                    │
       │                              │ leave_room                   │
       │                              │<─────────────────────────────┤
       │                              │  离开房间                     │
       │                              │                              │
```

---

## 阶段详解

### 阶段1: 通话邀请

#### 消息 ①: call_invite (A → 服务器)

**发起方**: 用户A  
**目的**: 向用户B发起通话邀请

**消息格式**:
```json
{
    "type": "call_invite",
    "user_id": "userA",
    "data": {
        "callee_id": "userB"
    },
    "timestamp": "2025-10-13T10:00:00Z"
}
```

**核心字段**:
- `type`: 消息类型，固定为 `call_invite`
- `user_id`: 主叫方用户ID
- `callee_id`: 被叫方用户ID

---

#### 消息 ②: call_invite (服务器 → B)

**发起方**: 服务器（转发）  
**目的**: 将邀请转发给被叫方，包含通话和房间信息

**消息格式**:
```json
{
    "type": "call_invite",
    "user_id": "userA",
    "data": {
        "id": "call_1697184000",
        "type": "one_to_one",
        "status": "inviting",
        "caller_id": "userA",
        "participants": ["userA", "userB"],
        "room_id": "room_1697184000",
        "created_at": "2025-10-13T10:00:00Z"
    },
    "timestamp": "2025-10-13T10:00:00Z"
}
```

**核心字段**:
- `id`: 通话唯一标识（call_id）
- `type`: 通话类型 (`one_to_one` | `group` | `meeting`)
- `status`: 通话状态 (`inviting` | `connected` | `ended`)
- `caller_id`: 主叫方用户ID
- `participants`: 参与者列表
- `room_id`: 房间ID（用于后续加入房间）

---

#### 消息 ③: call_accept (B → 服务器)

**发起方**: 用户B  
**目的**: 接受通话邀请

**消息格式**:
```json
{
    "type": "call_accept",
    "user_id": "userB",
    "data": {
        "call_id": "call_1697184000"
    },
    "timestamp": "2025-10-13T10:00:05Z"
}
```

**核心字段**:
- `call_id`: 通话ID（从邀请消息中获取）

---

#### 消息 ④: call_accept (服务器 → A & B)

**发起方**: 服务器（通知）  
**目的**: 通知双方通话已被接受，可以开始连接

**消息格式**:
```json
{
    "type": "call_accept",
    "user_id": "userB",
    "data": {
        "id": "call_1697184000",
        "status": "connected",
        "started_at": "2025-10-13T10:00:05Z"
    },
    "timestamp": "2025-10-13T10:00:05Z"
}
```

**核心字段**:
- `status`: 状态已更新为 `connected`
- `started_at`: 通话开始时间

---

### 阶段2: 加入房间

#### 消息 ⑤⑦: join_room (A & B → 服务器)

**发起方**: 用户A 和 用户B（双方都需要发送）  
**目的**: 加入通话房间，建立信令通道

**消息格式**:
```json
{
    "type": "join_room",
    "room_id": "room_1697184000",
    "user_id": "userA",
    "timestamp": "2025-10-13T10:00:06Z"
}
```

**核心字段**:
- `room_id`: 房间ID（从邀请消息中获取）
- `user_id`: 加入的用户ID

**服务器处理**:
1. 创建或获取房间
2. 创建User对象并添加到房间
3. 创建WebRTC PeerConnection
4. 保存连接映射 (userID → WebRTCConnection)
5. 添加用户到SFU

---

#### 消息 ⑥⑧: room_info (服务器 → A & B)

**发起方**: 服务器（响应）  
**目的**: 返回房间信息，包括当前在线用户

**消息格式**:
```json
{
    "type": "room_info",
    "room_id": "room_1697184000",
    "user_id": "userA",
    "data": {
        "room_id": "room_1697184000",
        "name": "Room room_1697184000",
        "users": [
            {
                "user_id": "userA",
                "username": "User userA",
                "is_muted": false,
                "is_video_on": true,
                "joined_at": "2025-10-13T10:00:06Z"
            }
        ],
        "created_at": "2025-10-13T10:00:06Z",
        "updated_at": "2025-10-13T10:00:06Z"
    },
    "timestamp": "2025-10-13T10:00:06Z"
}
```

**核心字段**:
- `users`: 房间内的用户列表
- `is_muted`: 是否静音
- `is_video_on`: 是否开启视频
- `joined_at`: 加入时间

---

#### 消息 ⑨: user_joined (服务器 → A)

**发起方**: 服务器（通知）  
**目的**: 通知已在房间的用户，有新用户加入

**消息格式**:
```json
{
    "type": "user_joined",
    "room_id": "room_1697184000",
    "user_id": "userB",
    "data": {
        "user_id": "userB",
        "username": "User userB",
        "is_muted": false,
        "is_video_on": true,
        "joined_at": "2025-10-13T10:00:07Z"
    },
    "timestamp": "2025-10-13T10:00:07Z"
}
```

**关键作用**: 用户A收到此消息后，会主动创建 Offer 开始 WebRTC 连接协商

---

### 阶段3: WebRTC连接协商

#### 消息 ⑩: offer (A → 服务器)

**发起方**: 用户A  
**目的**: 发送 SDP Offer，描述自己的媒体能力和配置

**消息格式**:
```json
{
    "type": "offer",
    "room_id": "room_1697184000",
    "user_id": "userA",
    "data": {
        "sdp": "v=0\r\no=- 4611731400430051336 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0 1\r\n...",
        "type": "offer"
    },
    "timestamp": "2025-10-13T10:00:08Z"
}
```

**核心字段**:
- `sdp`: 会话描述协议，包含媒体格式、编解码器、网络信息等
- `type`: 固定为 `offer`

**SDP内容说明**:
- 支持的音频编解码器（Opus, PCMU等）
- 支持的视频编解码器（H.264, VP8, VP9等）
- 媒体流的数量和类型（音频track、视频track）
- 传输协议和端口信息

---

#### 消息 ⑪: offer (服务器 → B)

**发起方**: 服务器（转发）  
**目的**: 将 Offer 转发给用户B

---

#### 消息 ⑫: answer (B → 服务器)

**发起方**: 用户B  
**目的**: 发送 SDP Answer，确认媒体配置

**消息格式**:
```json
{
    "type": "answer",
    "room_id": "room_1697184000",
    "user_id": "userB",
    "data": {
        "sdp": "v=0\r\no=- 4611731400430051337 2 IN IP4 192.168.1.100\r\ns=-\r\nt=0 0\r\n...",
        "type": "answer"
    },
    "timestamp": "2025-10-13T10:00:09Z"
}
```

**核心字段**:
- `sdp`: 根据 Offer 生成的应答描述
- `type`: 固定为 `answer`

**处理流程**:
1. 用户B收到 Offer
2. 调用 `setRemoteDescription(offer)` 设置远程描述
3. 调用 `createAnswer()` 生成 Answer
4. 调用 `setLocalDescription(answer)` 设置本地描述
5. 发送 Answer 给服务器

---

#### 消息 ⑬: answer (服务器 → A)

**发起方**: 服务器（转发）  
**目的**: 将 Answer 转发给用户A

**处理流程**:
- 用户A收到 Answer
- 调用 `setRemoteDescription(answer)` 设置远程描述
- 媒体协商完成

---

### 阶段4: ICE候选交换

#### 消息 ⑭⑯⑱⑳: ice_candidate (A & B → 服务器)

**发起方**: 用户A 和 用户B（双方都会发送多个）  
**目的**: 交换网络候选，寻找最佳连接路径

**消息格式**:
```json
{
    "type": "ice_candidate",
    "room_id": "room_1697184000",
    "user_id": "userA",
    "data": {
        "candidate": "candidate:1 1 UDP 2113667326 192.168.1.100 54400 typ host",
        "sdpMLineIndex": 0,
        "sdpMid": "0"
    },
    "timestamp": "2025-10-13T10:00:10Z"
}
```

**核心字段**:
- `candidate`: ICE候选字符串，包含网络信息
- `sdpMLineIndex`: 媒体行索引
- `sdpMid`: 媒体ID

**ICE候选类型**:
1. **host**: 本地内网地址（优先级最高）
   - 示例: `192.168.1.100:54400`
2. **srflx**: 通过STUN服务器获取的公网地址
   - 示例: `123.45.67.89:54400`
3. **relay**: TURN中继服务器地址（优先级最低，但最可靠）
   - 示例: `turn.example.com:3478`

**候选数量**: 每个用户通常会发送 3-10 个候选，取决于网络环境

---

#### 消息 ⑮⑰⑲㉑: ice_candidate (服务器转发)

**发起方**: 服务器（转发）  
**目的**: 将收到的 ICE 候选转发给对方

**处理流程**:
- 收到候选后立即调用 `addIceCandidate(candidate)`
- WebRTC 自动尝试建立连接
- 选择延迟最低、最稳定的路径

---

### 阶段5: 媒体传输建立

**状态**: WebRTC 连接建立完成

**特征**:
- ICE 连接状态变为 `connected`
- 开始传输实际的音视频数据
- 数据通过 P2P 通道传输，**不经过信令服务器**
- 使用 DTLS/SRTP 协议加密，确保安全

**数据流**:
```
用户A ←────── P2P 直连 ──────→ 用户B
  ↑                              ↑
  │                              │
音频流 (48kHz, 立体声)        音频流
视频流 (720p, 30fps)          视频流
```

---

### 阶段6: 媒体控制

#### 消息: mute / unmute

**发起方**: 任意用户  
**目的**: 控制麦克风状态

**消息格式**:
```json
{
    "type": "mute",
    "room_id": "room_1697184000",
    "user_id": "userA",
    "timestamp": "2025-10-13T10:05:00Z"
}
```

**广播**: 服务器会将此消息广播给房间内其他用户

---

#### 消息: video_on / video_off

**发起方**: 任意用户  
**目的**: 控制摄像头状态

**消息格式**:
```json
{
    "type": "video_off",
    "room_id": "room_1697184000",
    "user_id": "userB",
    "timestamp": "2025-10-13T10:06:00Z"
}
```

**广播**: 服务器会将此消息广播给房间内其他用户

---

### 阶段7: 通话结束

#### 消息 ㉖: call_end (任意用户 → 服务器)

**发起方**: 任意用户（A或B均可）  
**目的**: 结束通话

**消息格式**:
```json
{
    "type": "call_end",
    "user_id": "userB",
    "data": {
        "call_id": "call_1697184000"
    },
    "timestamp": "2025-10-13T10:15:00Z"
}
```

**服务器处理**:
1. 更新通话状态为 `ended`
2. 记录结束时间 `ended_at`
3. 计算通话时长 `duration`
4. 清理 CallManager 中的用户映射
5. 删除通话记录

---

#### 消息 ㉗: call_end (服务器 → A & B)

**发起方**: 服务器（通知）  
**目的**: 通知双方通话已结束

**消息格式**:
```json
{
    "type": "call_end",
    "user_id": "userB",
    "data": {
        "id": "call_1697184000",
        "status": "ended",
        "ended_at": "2025-10-13T10:15:00Z",
        "duration": 900
    },
    "timestamp": "2025-10-13T10:15:00Z"
}
```

**核心字段**:
- `status`: `ended`
- `ended_at`: 结束时间
- `duration`: 通话时长（秒）

---

#### 消息 ㉘㉙: leave_room (A & B → 服务器)

**发起方**: 用户A 和 用户B（双方都需要发送）  
**目的**: 离开房间，清理资源

**消息格式**:
```json
{
    "type": "leave_room",
    "room_id": "room_1697184000",
    "user_id": "userA",
    "timestamp": "2025-10-13T10:15:01Z"
}
```

**服务器处理**:
1. 从房间移除用户
2. 从 SFU 移除用户
3. 关闭 WebRTC 连接
4. 删除连接映射
5. 如果房间为空，标记房间可清理

---

## 核心数据结构

### Call 对象

通话的完整生命周期信息：

```json
{
    "id": "call_1697184000",
    "type": "one_to_one",
    "status": "connected",
    "caller_id": "userA",
    "participants": ["userA", "userB"],
    "room_id": "room_1697184000",
    "created_at": "2025-10-13T10:00:00Z",
    "started_at": "2025-10-13T10:00:05Z",
    "ended_at": "2025-10-13T10:15:00Z",
    "duration": 900
}
```

**字段说明**:
- `id`: 通话唯一标识
- `type`: 通话类型
  - `one_to_one`: 一对一通话
  - `group`: 群组通话
  - `meeting`: 会议
- `status`: 通话状态
  - `idle`: 空闲
  - `inviting`: 邀请中
  - `ringing`: 响铃中
  - `connected`: 已连接
  - `ended`: 已结束
  - `failed`: 失败
- `participants`: 参与者用户ID列表
- `room_id`: 关联的房间ID
- `duration`: 通话时长（秒）

---

### Room 对象

房间和用户信息：

```json
{
    "room_id": "room_1697184000",
    "name": "Room room_1697184000",
    "users": [
        {
            "user_id": "userA",
            "username": "User userA",
            "is_muted": false,
            "is_video_on": true,
            "joined_at": "2025-10-13T10:00:06Z"
        },
        {
            "user_id": "userB",
            "username": "User userB",
            "is_muted": false,
            "is_video_on": true,
            "joined_at": "2025-10-13T10:00:07Z"
        }
    ],
    "created_at": "2025-10-13T10:00:06Z",
    "updated_at": "2025-10-13T10:00:07Z"
}
```

---

### User 对象

用户在房间中的状态：

```json
{
    "user_id": "userA",
    "username": "User userA",
    "avatar": "https://example.com/avatar/userA.jpg",
    "joined_at": "2025-10-13T10:00:06Z",
    "is_muted": false,
    "is_video_on": true,
    "role": "host",
    "status": "online",
    "device": "web"
}
```

**字段说明**:
- `is_muted`: 是否静音
- `is_video_on`: 是否开启视频
- `role`: 角色（host=主持人, participant=参与者）
- `status`: 在线状态（online, offline, busy）
- `device`: 设备类型（web, mobile, desktop）

---

### WebRTC 配置

ICE 服务器配置：

```json
{
    "iceServers": [
        {
            "urls": [
                "stun:stun.l.google.com:19302",
                "stun:stun1.l.google.com:19302"
            ]
        },
        {
            "urls": "turn:turn.example.com:3478",
            "username": "user123",
            "credential": "password123"
        }
    ]
}
```

---

## 消息统计

### 按发起方分类

| 发起方 | 消息类型 | 数量 |
|-------|---------|------|
| **用户A** | call_invite | 1 |
| | join_room | 1 |
| | offer | 1 |
| | ice_candidate | 3-10 |
| | mute/unmute (可选) | 0-N |
| | video_on/off (可选) | 0-N |
| | call_end (可选) | 0-1 |
| | leave_room | 1 |
| **用户B** | call_accept | 1 |
| | join_room | 1 |
| | answer | 1 |
| | ice_candidate | 3-10 |
| | mute/unmute (可选) | 0-N |
| | video_on/off (可选) | 0-N |
| | call_end (可选) | 0-1 |
| | leave_room | 1 |
| **服务器** | 转发和通知 | 20-35 |

### 典型通话消息总数

- **最少**: 约 36 条（不包含媒体控制）
- **典型**: 约 45 条（包含少量媒体控制）
- **最多**: 约 65 条（包含大量ICE候选和媒体控制）

---

## 关键要点

### 1. 消息发起方

- **用户A和B是对等的**: 除了邀请环节，双方都会主动发送大量消息
- **Answer必须由B创建**: 服务器不能代替，这是 WebRTC 协议要求
- **ICE候选双向发送**: 双方都会发送 3-10 个候选

### 2. 时序关键点

- **必须先join_room再offer**: 确保WebRTC连接已创建
- **Offer/Answer必须按顺序**: Answer必须在收到Offer后创建
- **ICE可以并行**: ICE候选可以在Offer/Answer之前或之后发送

### 3. 连接建立

- **信令通道**: WebSocket，用于协商和控制
- **媒体通道**: WebRTC P2P，用于实际数据传输
- **加密**: 信令用WSS，媒体用DTLS/SRTP

### 4. 错误处理

- 任意环节失败都应清理资源
- 超时机制：邀请超时、连接超时、房间超时
- 状态同步：确保双方状态一致

### 5. 优化建议

- **ICE候选聚合**: 可以批量发送候选，减少消息数量
- **心跳机制**: 定期检测连接状态
- **重连机制**: 网络断开时自动重连

---

## 相关文档

- [WebRTC连接协商详解](./webrtc-connection-negotiation.md)
- [ICE候选类型说明](./ice-candidate-types.md)
- [SDP格式解析](./sdp-format-guide.md)
- [群组通话流程](./streaming-group-call-flow.md)
- [会议功能说明](./streaming-meeting-flow.md)

---

## 附录

### 消息类型定义

所有支持的信令消息类型：

**基础通话**:
- `join_room`: 加入房间
- `leave_room`: 离开房间
- `offer`: SDP Offer
- `answer`: SDP Answer
- `ice_candidate`: ICE候选
- `room_info`: 房间信息
- `user_joined`: 用户加入通知
- `user_left`: 用户离开通知
- `error`: 错误消息

**一对一通话**:
- `call_invite`: 通话邀请
- `call_accept`: 接受通话
- `call_reject`: 拒绝通话
- `call_end`: 结束通话

**媒体控制**:
- `mute`: 静音
- `unmute`: 取消静音
- `video_on`: 开启视频
- `video_off`: 关闭视频

---

**文档结束**

如有疑问，请参考源码：
- `apps/streaming/internal/handler/signaling.go`
- `apps/streaming/internal/logic/call_manager.go`
- `apps/streaming/room/manager.go`

