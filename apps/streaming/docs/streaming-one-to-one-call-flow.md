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

一对一音视频通话采用 **WebRTC + WebSocket + SFU** 的实现方案：

- **WebSocket**: 用于信令传输（连接协商、状态同步）
- **WebRTC**: 用于音视频数据传输
- **架构模式**: SFU（Selective Forwarding Unit）架构，服务器作为媒体中继
- **连接方式**: 每个用户与服务器建立独立的WebRTC连接，服务器负责转发媒体流

**核心组件**:
- **CallManager**: 管理通话生命周期
- **RoomManager**: 管理房间和用户
- **SignalingServer**: 处理信令消息和WebRTC协商
- **SFU**: 选择性转发单元，负责媒体流转发

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
  阶段3: WebRTC连接协商 (SDP 交换 - SFU模式)
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │ 【A主动】⑩                    │                              │
       │ offer                        │                              │
       ├─────────────────────────────>│                              │
       │  发送SDP Offer               │                              │
       │  (媒体能力描述)               │                              │
       │                              │                              │
       │                              ├─【服务器处理Offer】           │
       │                              │  setRemoteDescription        │
       │                              │  createAnswer                │
       │                              │  setLocalDescription         │
       │                              │                              │
       │ ⑪ answer (服务器返回)        │                              │
       │<─────────────────────────────┤                              │
       │  服务器自动创建Answer         │                              │
       │                              │                              │
       │ 【A处理Answer】               │                              │
       │ setRemoteDescription         │                              │
       │ A与服务器的连接协商完成       │                              │
       │                              │                              │
       │                              │ 【B主动】⑫                    │
       │                              │ offer                        │
       │                              │<─────────────────────────────┤
       │                              │  发送SDP Offer               │
       │                              │                              │
       │                              ├─【服务器处理Offer】           │
       │                              │  setRemoteDescription        │
       │                              │  createAnswer                │
       │                              │                              │
       │                              │ ⑬ answer (服务器返回)        │
       │                              ├─────────────────────────────>│
       │                              │  服务器自动创建Answer         │
       │                              │                              │
       │                              │                              │ 【B处理Answer】
       │                              │                              │ setRemoteDescription
       │                              │                              │ B与服务器的连接协商完成
       │                              │                              │

═══════════════════════════════════════════════════════════════════════════
  阶段4: ICE 候选交换 (用户与服务器连接协商)
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │ 【A主动】⑭                    │                              │
       │ ice_candidate #1             │                              │
       ├─────────────────────────────>│                              │
       │  发送网络候选                 │                              │
       │  (用于A-服务器连接)           │                              │
       │                              │                              │
       │                              ├─【服务器处理】                │
       │                              │  addIceCandidate             │
       │                              │  建立A-服务器连接路径         │
       │                              │                              │
       │ 【A主动】⑯                    │                              │
       │ ice_candidate #2             │                              │
       ├─────────────────────────────>│                              │
       │  (可能有多个候选)              │                              │
       │                              │                              │
       │                              ├─【服务器处理】                │
       │                              │  继续完善连接路径             │
       │                              │                              │
       │                              │ 【B主动】⑱                    │
       │                              │ ice_candidate #1             │
       │                              │<─────────────────────────────┤
       │                              │  发送网络候选                 │
       │                              │  (用于B-服务器连接)           │
       │                              │                              │
       │                              ├─【服务器处理】                │
       │                              │  addIceCandidate             │
       │                              │  建立B-服务器连接路径         │
       │                              │                              │
       │                              │ 【B主动】⑳                    │
       │                              │ ice_candidate #2             │
       │                              │<─────────────────────────────┤
       │                              │  (可能有多个候选)              │
       │                              │                              │
       │                              ├─【服务器处理】                │
       │                              │  继续完善连接路径             │
       │                              │                              │
       
       注意：服务器也会生成自己的ICE候选发送给A和B
       
       │ ⑮ ice_candidate (服务器端)   │                              │
       │<─────────────────────────────┤                              │
       │  服务器发送候选给A            │                              │
       │                              │                              │
       │                              │ ⑲ ice_candidate (服务器端)   │
       │                              ├─────────────────────────────>│
       │                              │  服务器发送候选给B            │
       │                              │                              │

═══════════════════════════════════════════════════════════════════════════
  阶段5: 媒体传输建立
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │          ┌─────────────────────────────────────┐            │
       │          │ WebRTC连接建立成功（SFU架构）       │            │
       │          │                                     │            │
       │          │ • A与服务器: ICE connected          │            │
       │          │ • B与服务器: ICE connected          │            │
       │          │ • 开始传输音视频数据                │            │
       │          │ • A发送流 → 服务器 → B              │            │
       │          │ • B发送流 → 服务器 → A              │            │
       │          │ • SFU负责选择性转发媒体流           │            │
       │          │ • 使用DTLS/SRTP加密                 │            │
       │          └─────────────────────────────────────┘            │
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
**目的**: 发送 SDP Offer，描述自己的媒体能力和配置，请求与服务器建立WebRTC连接

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

#### 消息 ⑪: answer (服务器 → A)

**发起方**: 服务器（自动生成）  
**目的**: 服务器创建Answer，与用户A完成WebRTC连接协商

**服务器处理流程**:
1. 服务器收到A的Offer
2. 调用 `setRemoteDescription(offer)` 设置远程描述
3. 调用 `createAnswer()` 自动生成Answer
4. 调用 `setLocalDescription(answer)` 设置本地描述
5. 发送Answer给A

**用户A处理流程**:
- 用户A收到服务器的Answer
- 调用 `setRemoteDescription(answer)` 设置远程描述
- A与服务器的媒体协商完成

---

#### 消息 ⑫: offer (B → 服务器)

**发起方**: 用户B  
**目的**: 发送 SDP Offer，描述自己的媒体能力和配置，请求与服务器建立WebRTC连接

**消息格式**:
```json
{
    "type": "offer",
    "room_id": "room_1697184000",
    "user_id": "userB",
    "data": {
        "sdp": "v=0\r\no=- 4611731400430051337 2 IN IP4 192.168.1.100\r\ns=-\r\nt=0 0\r\n...",
        "type": "offer"
    },
    "timestamp": "2025-10-13T10:00:09Z"
}
```

---

#### 消息 ⑬: answer (服务器 → B)

**发起方**: 服务器（自动生成）  
**目的**: 服务器创建Answer，与用户B完成WebRTC连接协商

**服务器处理流程**:
1. 服务器收到B的Offer
2. 调用 `setRemoteDescription(offer)` 设置远程描述
3. 调用 `createAnswer()` 自动生成Answer
4. 调用 `setLocalDescription(answer)` 设置本地描述
5. 发送Answer给B

**用户B处理流程**:
- 用户B收到服务器的Answer
- 调用 `setRemoteDescription(answer)` 设置远程描述
- B与服务器的媒体协商完成

**重要说明**:
- 在SFU架构中，A和B都是与服务器建立独立的WebRTC连接
- 服务器作为中间节点，负责接收和转发媒体流
- 不存在A和B之间的直接P2P连接

---

### 阶段4: ICE候选交换

#### 消息 ⑭⑯⑱⑳: ice_candidate (A & B → 服务器)

**发起方**: 用户A 和 用户B（双方都会发送多个）  
**目的**: 发送网络候选给服务器，建立用户与服务器之间的最佳连接路径

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

**服务器处理流程**:
1. 服务器收到用户的ICE候选
2. 调用 `addIceCandidate(candidate)` 添加候选到对应的PeerConnection
3. WebRTC自动尝试使用候选建立连接
4. 选择延迟最低、最稳定的路径

---

#### 消息 ⑮⑲: ice_candidate (服务器 → A & B)

**发起方**: 服务器  
**目的**: 服务器也会生成自己的ICE候选，发送给用户A和B

**说明**:
- 服务器的PeerConnection会生成自己的ICE候选
- 这些候选代表服务器的网络地址
- 用户收到后调用 `addIceCandidate()` 添加
- 最终建立用户到服务器的WebRTC连接

**重要说明**:
- ICE候选交换是双向的：用户 ↔ 服务器
- 不存在用户A和用户B之间的直接ICE候选交换
- 每个用户只需要与服务器建立连接即可

---

### 阶段5: 媒体传输建立

**状态**: WebRTC 连接建立完成（SFU架构）

**特征**:
- A与服务器的ICE连接状态变为 `connected`
- B与服务器的ICE连接状态变为 `connected`
- 开始传输实际的音视频数据
- 数据通过**SFU服务器转发**，不是P2P直连
- 使用 DTLS/SRTP 协议加密，确保安全

**数据流（SFU架构）**:
```
用户A                    SFU服务器                    用户B
  │                          │                          │
  ├─ 音频流 (48kHz) ────────>│                          │
  │  视频流 (720p)           │                          │
  │                          ├─ 转发音频/视频 ────────>│
  │                          │                          │
  │                          │<──── 音频流 (48kHz) ─────┤
  │                          │      视频流 (720p)       │
  │<──── 转发音频/视频 ──────┤                          │
  │                          │                          │
```

**SFU工作原理**:
1. 用户A发送媒体流到服务器
2. 用户B发送媒体流到服务器  
3. SFU服务器选择性地转发流：
   - 接收A的流，转发给B
   - 接收B的流，转发给A
4. 每个用户只维护一个上行和一个下行连接
5. 服务器可以控制转发质量和带宽

**优势**:
- 适合多人通话（每个用户只需一个连接）
- 服务器可以进行流控制和质量调整
- 降低客户端带宽和CPU压力

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

### 1. 架构特点

- **SFU架构**: 服务器作为媒体中继，而非P2P直连
- **独立连接**: 每个用户与服务器建立独立的WebRTC连接
- **服务器协商**: Offer/Answer在用户和服务器之间完成，服务器自动生成Answer
- **媒体转发**: SFU负责接收和转发所有媒体流

### 2. 消息发起方

- **用户A和B是对等的**: 除了邀请环节，双方都会主动发送消息
- **Answer由服务器创建**: 服务器收到Offer后自动生成Answer，不需要转发给其他用户
- **ICE候选双向交换**: 用户发送候选给服务器，服务器也发送候选给用户

### 3. 时序关键点

- **必须先join_room再offer**: 确保WebRTC连接已创建
- **Offer/Answer必须按顺序**: 服务器在收到Offer后立即创建Answer返回
- **ICE可以并行**: ICE候选可以在Offer/Answer之前或之后发送
- **独立协商**: A和B分别与服务器完成协商，互不影响

### 4. 连接建立

- **信令通道**: WebSocket，用于协商和控制
- **媒体通道**: WebRTC连接（用户-服务器），用于实际数据传输
- **加密**: 信令用WSS，媒体用DTLS/SRTP
- **连接拓扑**: 星型拓扑（用户 ↔ 服务器 ↔ 用户）

### 5. 错误处理

- 任意环节失败都应清理资源
- 超时机制：邀请超时、连接超时、房间超时
- 状态同步：确保双方状态一致

### 6. 优化建议

- **ICE候选聚合**: 可以批量发送候选，减少消息数量
- **心跳机制**: 定期检测连接状态
- **重连机制**: 网络断开时自动重连
- **媒体质量调整**: SFU可以根据网络状况动态调整转发质量
- **带宽管理**: 服务器端统一管理和优化带宽分配

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
- `apps/streaming/internal/handler/signaling.go` - 信令服务器和消息处理
- `apps/streaming/internal/logic/call_manager.go` - 通话生命周期管理
- `apps/streaming/room/manager.go` - 房间管理
- `apps/streaming/sfu/sfu.go` - SFU媒体转发实现
- `apps/streaming/webrtc/connection.go` - WebRTC连接封装

**注意**: 本文档描述的是**SFU架构**实现，不是P2P架构。如需P2P实现，需要修改服务器端代码逻辑。

