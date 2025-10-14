# 一对一通话消息数据结构

> **文档版本**: v1.0  
> **更新日期**: 2025-10-13  
> **说明**: 本文档包含一对一通话完整流程中所有消息的JSON数据结构

## 目录

- [阶段1: 通话邀请](#阶段1-通话邀请)
- [阶段2: 加入房间](#阶段2-加入房间)
- [阶段3: WebRTC连接协商](#阶段3-webrtc连接协商)
- [阶段4: ICE候选交换](#阶段4-ice候选交换)
- [阶段5: 媒体控制](#阶段5-媒体控制)
- [阶段6: 通话结束](#阶段6-通话结束)
- [错误消息](#错误消息)
- [字段说明](#字段说明)

---

## 阶段1: 通话邀请

### 1.1 发起通话邀请 (用户A → 服务器)

**消息类型**: `call_invite`  
**方向**: Client → Server  
**发送者**: 主叫方（用户A）

```json
{
  "type": "call_invite",
  "user_id": "user_a_123",
  "data": {
    "callee_id": "user_b_456"
  },
  "timestamp": "2025-10-13T10:00:00.123Z"
}
```

**字段说明**:
- `type`: 消息类型，固定为 "call_invite"
- `user_id`: 主叫方用户ID
- `data.callee_id`: 被叫方用户ID
- `timestamp`: ISO 8601 格式的时间戳

---

### 1.2 转发邀请 (服务器 → 用户B)

**消息类型**: `call_invite`  
**方向**: Server → Client  
**接收者**: 被叫方（用户B）

```json
{
  "type": "call_invite",
  "user_id": "user_a_123",
  "data": {
    "id": "call_1697184000123",
    "type": "one_to_one",
    "status": "inviting",
    "caller_id": "user_a_123",
    "caller_name": "Alice",
    "caller_avatar": "https://cdn.example.com/avatars/user_a.jpg",
    "participants": ["user_a_123", "user_b_456"],
    "room_id": "room_1697184000456",
    "created_at": "2025-10-13T10:00:00.123Z"
  },
  "timestamp": "2025-10-13T10:00:00.234Z"
}
```

**字段说明**:
- `data.id`: 通话唯一标识符（call_id）
- `data.type`: 通话类型 ("one_to_one" | "group" | "meeting")
- `data.status`: 通话状态 ("inviting" | "connected" | "ended")
- `data.caller_id`: 主叫方用户ID
- `data.caller_name`: 主叫方显示名称
- `data.caller_avatar`: 主叫方头像URL
- `data.participants`: 参与者用户ID数组
- `data.room_id`: 关联的房间ID
- `data.created_at`: 通话创建时间

---

### 1.3 接受通话 (用户B → 服务器)

**消息类型**: `call_accept`  
**方向**: Client → Server  
**发送者**: 被叫方（用户B）

```json
{
  "type": "call_accept",
  "user_id": "user_b_456",
  "data": {
    "call_id": "call_1697184000123"
  },
  "timestamp": "2025-10-13T10:00:05.678Z"
}
```

**字段说明**:
- `data.call_id`: 要接受的通话ID（从邀请消息中获取）

---

### 1.4 通知接受 (服务器 → 用户A & 用户B)

**消息类型**: `call_accept`  
**方向**: Server → Client  
**接收者**: 双方

```json
{
  "type": "call_accept",
  "user_id": "user_b_456",
  "data": {
    "id": "call_1697184000123",
    "type": "one_to_one",
    "status": "connected",
    "caller_id": "user_a_123",
    "participants": ["user_a_123", "user_b_456"],
    "room_id": "room_1697184000456",
    "created_at": "2025-10-13T10:00:00.123Z",
    "started_at": "2025-10-13T10:00:05.678Z"
  },
  "timestamp": "2025-10-13T10:00:05.789Z"
}
```

**字段说明**:
- `data.status`: 状态已更新为 "connected"
- `data.started_at`: 通话开始时间

---

### 1.5 拒绝通话 (可选，用户B → 服务器)

**消息类型**: `call_reject`  
**方向**: Client → Server  
**发送者**: 被叫方（用户B）

```json
{
  "type": "call_reject",
  "user_id": "user_b_456",
  "data": {
    "call_id": "call_1697184000123",
    "reason": "busy"
  },
  "timestamp": "2025-10-13T10:00:05.678Z"
}
```

**字段说明**:
- `data.reason`: 拒绝原因 ("busy" | "declined" | "timeout")
  - `busy`: 正忙
  - `declined`: 主动拒绝
  - `timeout`: 超时未接听

---

## 阶段2: 加入房间

### 2.1 加入房间请求 (用户A/B → 服务器)

**消息类型**: `join_room`  
**方向**: Client → Server  
**发送者**: 任一参与者

```json
{
  "type": "join_room",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "timestamp": "2025-10-13T10:00:06.123Z"
}
```

**字段说明**:
- `room_id`: 要加入的房间ID（从call对象中获取）

---

### 2.2 房间信息响应 (服务器 → 用户A/B)

**消息类型**: `room_info`  
**方向**: Server → Client  
**接收者**: 加入房间的用户

```json
{
  "type": "room_info",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": {
    "room_id": "room_1697184000456",
    "name": "Room room_1697184000456",
    "users": [
      {
        "user_id": "user_a_123",
        "username": "Alice",
        "avatar": "https://cdn.example.com/avatars/user_a.jpg",
        "joined_at": "2025-10-13T10:00:06.123Z",
        "is_muted": false,
        "is_video_on": true,
        "role": "host",
        "status": "online",
        "device": "web"
      }
    ],
    "created_at": "2025-10-13T10:00:06.123Z",
    "updated_at": "2025-10-13T10:00:06.123Z"
  },
  "timestamp": "2025-10-13T10:00:06.234Z"
}
```

**字段说明**:
- `data.users`: 房间内的用户列表
- `data.users[].user_id`: 用户ID
- `data.users[].username`: 用户显示名称
- `data.users[].avatar`: 用户头像URL
- `data.users[].joined_at`: 加入时间
- `data.users[].is_muted`: 是否静音
- `data.users[].is_video_on`: 是否开启视频
- `data.users[].role`: 用户角色 ("host" | "participant" | "viewer")
- `data.users[].status`: 在线状态 ("online" | "offline" | "busy")
- `data.users[].device`: 设备类型 ("web" | "mobile" | "desktop")

---

### 2.3 用户加入通知 (服务器 → 房间内其他用户)

**消息类型**: `user_joined`  
**方向**: Server → Client  
**接收者**: 房间内已存在的用户

```json
{
  "type": "user_joined",
  "room_id": "room_1697184000456",
  "user_id": "user_b_456",
  "data": {
    "user_id": "user_b_456",
    "username": "Bob",
    "avatar": "https://cdn.example.com/avatars/user_b.jpg",
    "joined_at": "2025-10-13T10:00:07.456Z",
    "is_muted": false,
    "is_video_on": true,
    "role": "participant",
    "status": "online",
    "device": "web"
  },
  "timestamp": "2025-10-13T10:00:07.567Z"
}
```

---

## 阶段3: WebRTC连接协商

### 3.1 发送Offer (用户A → 服务器)

**消息类型**: `offer`  
**方向**: Client → Server  
**发送者**: 主叫方（用户A）

```json
{
  "type": "offer",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": {
    "sdp": "v=0\r\no=- 4611731400430051336 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0 1\r\na=extmap-allow-mixed\r\na=msid-semantic: WMS stream_id\r\nm=audio 9 UDP/TLS/RTP/SAVPF 111 103 104 9 0 8 106 105 13 110 112 113 126\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:abc123\r\na=ice-pwd:def456ghi789\r\na=ice-options:trickle\r\na=fingerprint:sha-256 12:34:56:78:9A:BC:DE:F0:12:34:56:78:9A:BC:DE:F0:12:34:56:78:9A:BC:DE:F0:12:34:56:78:9A:BC:DE:F0\r\na=setup:actpass\r\na=mid:0\r\na=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r\na=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r\na=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r\na=sendrecv\r\na=rtcp-mux\r\na=rtpmap:111 opus/48000/2\r\na=rtcp-fb:111 transport-cc\r\na=fmtp:111 minptime=10;useinbandfec=1\r\nm=video 9 UDP/TLS/RTP/SAVPF 96 97 98 99 100 101 102 122 127 121 125 107 108 109 124 120 123 119 114 115 116\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:abc123\r\na=ice-pwd:def456ghi789\r\na=ice-options:trickle\r\na=fingerprint:sha-256 12:34:56:78:9A:BC:DE:F0:12:34:56:78:9A:BC:DE:F0:12:34:56:78:9A:BC:DE:F0:12:34:56:78:9A:BC:DE:F0\r\na=setup:actpass\r\na=mid:1\r\na=extmap:14 urn:ietf:params:rtp-hdrext:toffset\r\na=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r\na=extmap:13 urn:3gpp:video-orientation\r\na=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r\na=extmap:12 http://www.webrtc.org/experiments/rtp-hdrext/playout-delay\r\na=extmap:11 http://www.webrtc.org/experiments/rtp-hdrext/video-content-type\r\na=extmap:7 http://www.webrtc.org/experiments/rtp-hdrext/video-timing\r\na=extmap:8 http://www.webrtc.org/experiments/rtp-hdrext/color-space\r\na=sendrecv\r\na=rtcp-mux\r\na=rtcp-rsize\r\na=rtpmap:96 VP8/90000\r\na=rtcp-fb:96 goog-remb\r\na=rtcp-fb:96 transport-cc\r\na=rtcp-fb:96 ccm fir\r\na=rtcp-fb:96 nack\r\na=rtcp-fb:96 nack pli\r\na=rtpmap:97 rtx/90000\r\na=fmtp:97 apt=96\r\na=rtpmap:98 VP9/90000\r\na=rtcp-fb:98 goog-remb\r\na=rtcp-fb:98 transport-cc\r\na=rtcp-fb:98 ccm fir\r\na=rtcp-fb:98 nack\r\na=rtcp-fb:98 nack pli\r\n",
    "type": "offer"
  },
  "timestamp": "2025-10-13T10:00:08.123Z"
}
```

**字段说明**:
- `data.sdp`: SDP (Session Description Protocol) 字符串
  - 包含媒体能力描述（音频、视频编解码器）
  - 包含网络信息（ICE候选的初始信息）
  - 包含加密参数（fingerprint）
- `data.type`: 固定为 "offer"

**SDP 关键信息**:
- `m=audio`: 音频媒体描述
  - 支持的编解码器: Opus (111), PCMU (0), PCMA (8) 等
- `m=video`: 视频媒体描述
  - 支持的编解码器: VP8 (96), VP9 (98), H264 等
- `a=ice-ufrag`: ICE 用户名片段
- `a=ice-pwd`: ICE 密码
- `a=fingerprint`: DTLS 指纹（用于加密）

---

### 3.2 转发Offer (服务器 → 用户B)

**消息类型**: `offer`  
**方向**: Server → Client  
**接收者**: 被叫方（用户B）

```json
{
  "type": "offer",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": {
    "sdp": "v=0\r\no=- 4611731400430051336 2 IN IP4 127.0.0.1\r\n...(同上)",
    "type": "offer"
  },
  "timestamp": "2025-10-13T10:00:08.234Z"
}
```

---

### 3.3 发送Answer (用户B → 服务器)

**消息类型**: `answer`  
**方向**: Client → Server  
**发送者**: 被叫方（用户B）

```json
{
  "type": "answer",
  "room_id": "room_1697184000456",
  "user_id": "user_b_456",
  "data": {
    "sdp": "v=0\r\no=- 9876543210987654321 2 IN IP4 192.168.1.100\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0 1\r\na=extmap-allow-mixed\r\na=msid-semantic: WMS stream_id\r\nm=audio 9 UDP/TLS/RTP/SAVPF 111 103 104 9 0 8 106 105 13 110 112 113 126\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:xyz789\r\na=ice-pwd:uvw012abc345\r\na=ice-options:trickle\r\na=fingerprint:sha-256 AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89\r\na=setup:active\r\na=mid:0\r\na=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r\na=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r\na=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r\na=sendrecv\r\na=rtcp-mux\r\na=rtpmap:111 opus/48000/2\r\na=rtcp-fb:111 transport-cc\r\na=fmtp:111 minptime=10;useinbandfec=1\r\nm=video 9 UDP/TLS/RTP/SAVPF 96 97 98\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:xyz789\r\na=ice-pwd:uvw012abc345\r\na=ice-options:trickle\r\na=fingerprint:sha-256 AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89\r\na=setup:active\r\na=mid:1\r\na=extmap:14 urn:ietf:params:rtp-hdrext:toffset\r\na=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r\na=extmap:13 urn:3gpp:video-orientation\r\na=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r\na=sendrecv\r\na=rtcp-mux\r\na=rtcp-rsize\r\na=rtpmap:96 VP8/90000\r\na=rtcp-fb:96 goog-remb\r\na=rtcp-fb:96 transport-cc\r\na=rtcp-fb:96 ccm fir\r\na=rtcp-fb:96 nack\r\na=rtcp-fb:96 nack pli\r\n",
    "type": "answer"
  },
  "timestamp": "2025-10-13T10:00:09.456Z"
}
```

**字段说明**:
- `data.sdp`: SDP Answer 字符串（根据 Offer 生成）
- `data.type`: 固定为 "answer"

---

### 3.4 转发Answer (服务器 → 用户A)

**消息类型**: `answer`  
**方向**: Server → Client  
**接收者**: 主叫方（用户A）

```json
{
  "type": "answer",
  "room_id": "room_1697184000456",
  "user_id": "user_b_456",
  "data": {
    "sdp": "v=0\r\no=- 9876543210987654321 2 IN IP4 192.168.1.100\r\n...(同上)",
    "type": "answer"
  },
  "timestamp": "2025-10-13T10:00:09.567Z"
}
```

---

## 阶段4: ICE候选交换

### 4.1 发送ICE候选 (用户A → 服务器)

**消息类型**: `ice_candidate`  
**方向**: Client → Server  
**发送者**: 任一参与者

#### 示例1: Host候选（内网地址）

```json
{
  "type": "ice_candidate",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": {
    "candidate": "candidate:1 1 UDP 2130706431 192.168.1.100 54400 typ host",
    "sdpMLineIndex": 0,
    "sdpMid": "0"
  },
  "timestamp": "2025-10-13T10:00:10.123Z"
}
```

**ICE候选字符串解析**:
```
candidate:1 1 UDP 2130706431 192.168.1.100 54400 typ host
          │ │  │        │            │         │     │  │
          │ │  │        │            │         │     │  └─ 候选类型
          │ │  │        │            │         │     └─ 端口号
          │ │  │        │            │         └─ IP地址
          │ │  │        │            └─ 优先级(数字越大优先级越高)
          │ │  │        └─ 传输协议
          │ │  └─ 组件ID (1=RTP, 2=RTCP)
          │ └─ 基础ID
          └─ 候选序号
```

**候选类型**:
- `host`: 本地内网地址（优先级最高，延迟最低）
- `srflx`: 通过STUN服务器获取的公网地址
- `relay`: TURN中继服务器地址（优先级最低，但最可靠）

---

#### 示例2: SRFLX候选（公网地址）

```json
{
  "type": "ice_candidate",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": {
    "candidate": "candidate:2 1 UDP 1694498815 203.0.113.45 54401 typ srflx raddr 192.168.1.100 rport 54400",
    "sdpMLineIndex": 0,
    "sdpMid": "0"
  },
  "timestamp": "2025-10-13T10:00:10.234Z"
}
```

**字段说明**:
- `raddr`: 相关地址（内网地址）
- `rport`: 相关端口（内网端口）

---

#### 示例3: RELAY候选（中继地址）

```json
{
  "type": "ice_candidate",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": {
    "candidate": "candidate:3 1 UDP 16777215 turn.example.com 3478 typ relay raddr 203.0.113.45 rport 54401",
    "sdpMLineIndex": 0,
    "sdpMid": "0"
  },
  "timestamp": "2025-10-13T10:00:10.345Z"
}
```

---

### 4.2 转发ICE候选 (服务器 → 另一方)

**消息类型**: `ice_candidate`  
**方向**: Server → Client  
**接收者**: 对方用户

```json
{
  "type": "ice_candidate",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": {
    "candidate": "candidate:1 1 UDP 2130706431 192.168.1.100 54400 typ host",
    "sdpMLineIndex": 0,
    "sdpMid": "0"
  },
  "timestamp": "2025-10-13T10:00:10.456Z"
}
```

**字段说明**:
- `data.candidate`: ICE候选字符串
- `data.sdpMLineIndex`: SDP媒体行索引（0=音频, 1=视频）
- `data.sdpMid`: SDP媒体ID

---

### 4.3 ICE候选完成 (可选)

**消息类型**: `ice_candidate`  
**方向**: Client → Server  
**说明**: 表示该用户的ICE候选收集完成

```json
{
  "type": "ice_candidate",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": {
    "candidate": null,
    "sdpMLineIndex": null,
    "sdpMid": null
  },
  "timestamp": "2025-10-13T10:00:11.123Z"
}
```

**字段说明**:
- `candidate` 为 `null` 表示候选收集完成

---

## 阶段5: 媒体控制

### 5.1 静音 (用户A → 服务器)

**消息类型**: `mute`  
**方向**: Client → Server

```json
{
  "type": "mute",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": {
    "room_id": "room_1697184000456"
  },
  "timestamp": "2025-10-13T10:05:00.123Z"
}
```

---

### 5.2 静音通知 (服务器 → 用户B)

**消息类型**: `mute`  
**方向**: Server → Client

```json
{
  "type": "mute",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "timestamp": "2025-10-13T10:05:00.234Z"
}
```

---

### 5.3 取消静音 (用户A → 服务器)

**消息类型**: `unmute`  
**方向**: Client → Server

```json
{
  "type": "unmute",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": {
    "room_id": "room_1697184000456"
  },
  "timestamp": "2025-10-13T10:05:30.123Z"
}
```

---

### 5.4 关闭视频 (用户B → 服务器)

**消息类型**: `video_off`  
**方向**: Client → Server

```json
{
  "type": "video_off",
  "room_id": "room_1697184000456",
  "user_id": "user_b_456",
  "data": {
    "room_id": "room_1697184000456"
  },
  "timestamp": "2025-10-13T10:06:00.456Z"
}
```

---

### 5.5 开启视频 (用户B → 服务器)

**消息类型**: `video_on`  
**方向**: Client → Server

```json
{
  "type": "video_on",
  "room_id": "room_1697184000456",
  "user_id": "user_b_456",
  "data": {
    "room_id": "room_1697184000456"
  },
  "timestamp": "2025-10-13T10:06:30.456Z"
}
```

---

## 阶段6: 通话结束

### 6.1 结束通话 (用户A/B → 服务器)

**消息类型**: `call_end`  
**方向**: Client → Server  
**发送者**: 任一参与者

```json
{
  "type": "call_end",
  "user_id": "user_a_123",
  "data": {
    "call_id": "call_1697184000123"
  },
  "timestamp": "2025-10-13T10:15:00.123Z"
}
```

---

### 6.2 通知通话结束 (服务器 → 双方)

**消息类型**: `call_end`  
**方向**: Server → Client  
**接收者**: 双方

```json
{
  "type": "call_end",
  "user_id": "user_a_123",
  "data": {
    "id": "call_1697184000123",
    "type": "one_to_one",
    "status": "ended",
    "caller_id": "user_a_123",
    "participants": ["user_a_123", "user_b_456"],
    "room_id": "room_1697184000456",
    "created_at": "2025-10-13T10:00:00.123Z",
    "started_at": "2025-10-13T10:00:05.678Z",
    "ended_at": "2025-10-13T10:15:00.123Z",
    "duration": 894
  },
  "timestamp": "2025-10-13T10:15:00.234Z"
}
```

**字段说明**:
- `data.status`: "ended"
- `data.ended_at`: 结束时间
- `data.duration`: 通话时长（秒）

---

### 6.3 离开房间 (用户A/B → 服务器)

**消息类型**: `leave_room`  
**方向**: Client → Server

```json
{
  "type": "leave_room",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "timestamp": "2025-10-13T10:15:01.123Z"
}
```

---

### 6.4 用户离开通知 (服务器 → 另一方)

**消息类型**: `user_left`  
**方向**: Server → Client

```json
{
  "type": "user_left",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "timestamp": "2025-10-13T10:15:01.234Z"
}
```

---

## 错误消息

### 错误响应格式

**消息类型**: `error`  
**方向**: Server → Client

```json
{
  "type": "error",
  "user_id": "user_a_123",
  "data": {
    "error": "user is already in a call",
    "error_code": "CALL_ALREADY_EXISTS",
    "message": "用户正在通话中，请稍后再试",
    "details": {
      "current_call_id": "call_1234567890",
      "timestamp": "2025-10-13T10:00:00.123Z"
    }
  },
  "timestamp": "2025-10-13T10:00:00.234Z"
}
```

**常见错误码**:

#### CALL_ALREADY_EXISTS
```json
{
  "type": "error",
  "user_id": "user_a_123",
  "data": {
    "error": "user is already in a call",
    "error_code": "CALL_ALREADY_EXISTS",
    "message": "用户正在通话中"
  },
  "timestamp": "2025-10-13T10:00:00.234Z"
}
```

#### CALL_NOT_FOUND
```json
{
  "type": "error",
  "user_id": "user_a_123",
  "data": {
    "error": "call not found",
    "error_code": "CALL_NOT_FOUND",
    "message": "通话不存在"
  },
  "timestamp": "2025-10-13T10:00:00.234Z"
}
```

#### ROOM_NOT_FOUND
```json
{
  "type": "error",
  "user_id": "user_a_123",
  "data": {
    "error": "room not found",
    "error_code": "ROOM_NOT_FOUND",
    "message": "房间不存在"
  },
  "timestamp": "2025-10-13T10:00:00.234Z"
}
```

#### INVALID_DATA
```json
{
  "type": "error",
  "user_id": "user_a_123",
  "data": {
    "error": "invalid data",
    "error_code": "INVALID_DATA",
    "message": "数据格式错误",
    "details": {
      "field": "callee_id",
      "reason": "required field missing"
    }
  },
  "timestamp": "2025-10-13T10:00:00.234Z"
}
```

#### WEBRTC_CONNECTION_NOT_FOUND
```json
{
  "type": "error",
  "user_id": "user_a_123",
  "data": {
    "error": "WebRTC connection not found",
    "error_code": "WEBRTC_CONNECTION_NOT_FOUND",
    "message": "WebRTC连接不存在"
  },
  "timestamp": "2025-10-13T10:00:00.234Z"
}
```

---

## 字段说明

### 通用字段

所有消息都包含以下通用字段：

```json
{
  "type": "string",      // 消息类型
  "user_id": "string",   // 用户ID（可选）
  "room_id": "string",   // 房间ID（可选）
  "data": {},            // 消息数据（可选）
  "timestamp": "string"  // ISO 8601 时间戳
}
```

---

### 消息类型枚举

#### 通话相关
- `call_invite`: 通话邀请
- `call_accept`: 接受通话
- `call_reject`: 拒绝通话
- `call_end`: 结束通话

#### 房间相关
- `join_room`: 加入房间
- `leave_room`: 离开房间
- `room_info`: 房间信息
- `user_joined`: 用户加入通知
- `user_left`: 用户离开通知

#### WebRTC相关
- `offer`: SDP Offer
- `answer`: SDP Answer
- `ice_candidate`: ICE候选

#### 媒体控制
- `mute`: 静音
- `unmute`: 取消静音
- `video_on`: 开启视频
- `video_off`: 关闭视频

#### 系统消息
- `error`: 错误消息

---

### 通话状态枚举

```typescript
type CallStatus = 
  | "idle"       // 空闲
  | "inviting"   // 邀请中
  | "ringing"    // 响铃中
  | "connected"  // 已连接
  | "ended"      // 已结束
  | "failed";    // 失败
```

---

### 通话类型枚举

```typescript
type CallType = 
  | "one_to_one"  // 一对一通话
  | "group"       // 群组通话
  | "meeting"     // 会议
  | "live";       // 直播
```

---

### 用户角色枚举

```typescript
type UserRole = 
  | "host"        // 主持人
  | "participant" // 参与者
  | "viewer";     // 观众
```

---

### 用户状态枚举

```typescript
type UserStatus = 
  | "online"   // 在线
  | "offline"  // 离线
  | "busy";    // 忙碌
```

---

### 设备类型枚举

```typescript
type DeviceType = 
  | "web"      // 网页
  | "mobile"   // 移动端
  | "desktop"; // 桌面端
```

---

## 完整流程示例

### 用户A的视角（完整消息序列）

```json
// 1. 发起邀请
{
  "type": "call_invite",
  "user_id": "user_a_123",
  "data": { "callee_id": "user_b_456" },
  "timestamp": "2025-10-13T10:00:00.123Z"
}

// 2. 收到接受通知
{
  "type": "call_accept",
  "user_id": "user_b_456",
  "data": {
    "id": "call_1697184000123",
    "status": "connected",
    "started_at": "2025-10-13T10:00:05.678Z"
  },
  "timestamp": "2025-10-13T10:00:05.789Z"
}

// 3. 加入房间
{
  "type": "join_room",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "timestamp": "2025-10-13T10:00:06.123Z"
}

// 4. 收到房间信息
{
  "type": "room_info",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": { /* 房间详情 */ },
  "timestamp": "2025-10-13T10:00:06.234Z"
}

// 5. 收到用户B加入通知
{
  "type": "user_joined",
  "room_id": "room_1697184000456",
  "user_id": "user_b_456",
  "data": { /* 用户B信息 */ },
  "timestamp": "2025-10-13T10:00:07.567Z"
}

// 6. 发送Offer
{
  "type": "offer",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": { "sdp": "...", "type": "offer" },
  "timestamp": "2025-10-13T10:00:08.123Z"
}

// 7. 收到Answer
{
  "type": "answer",
  "room_id": "room_1697184000456",
  "user_id": "user_b_456",
  "data": { "sdp": "...", "type": "answer" },
  "timestamp": "2025-10-13T10:00:09.567Z"
}

// 8. 发送ICE候选（多个）
{
  "type": "ice_candidate",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "data": { "candidate": "..." },
  "timestamp": "2025-10-13T10:00:10.123Z"
}

// 9. 收到B的ICE候选（多个）
{
  "type": "ice_candidate",
  "room_id": "room_1697184000456",
  "user_id": "user_b_456",
  "data": { "candidate": "..." },
  "timestamp": "2025-10-13T10:00:10.789Z"
}

// ... 通话进行中 ...

// 10. 结束通话
{
  "type": "call_end",
  "user_id": "user_a_123",
  "data": { "call_id": "call_1697184000123" },
  "timestamp": "2025-10-13T10:15:00.123Z"
}

// 11. 离开房间
{
  "type": "leave_room",
  "room_id": "room_1697184000456",
  "user_id": "user_a_123",
  "timestamp": "2025-10-13T10:15:01.123Z"
}
```

---

## TypeScript 类型定义

```typescript
// 基础消息结构
interface SignalingMessage {
  type: SignalingMessageType;
  user_id?: string;
  room_id?: string;
  data?: any;
  timestamp: string; // ISO 8601
}

// 消息类型
type SignalingMessageType =
  | 'call_invite'
  | 'call_accept'
  | 'call_reject'
  | 'call_end'
  | 'join_room'
  | 'leave_room'
  | 'room_info'
  | 'user_joined'
  | 'user_left'
  | 'offer'
  | 'answer'
  | 'ice_candidate'
  | 'mute'
  | 'unmute'
  | 'video_on'
  | 'video_off'
  | 'error';

// Call对象
interface Call {
  id: string;
  type: CallType;
  status: CallStatus;
  caller_id: string;
  participants: string[];
  room_id: string;
  created_at: string;
  started_at?: string;
  ended_at?: string;
  duration?: number; // 秒
}

// Room对象
interface Room {
  room_id: string;
  name: string;
  users: User[];
  created_at: string;
  updated_at: string;
}

// User对象
interface User {
  user_id: string;
  username: string;
  avatar?: string;
  joined_at: string;
  is_muted: boolean;
  is_video_on: boolean;
  role: UserRole;
  status: UserStatus;
  device: DeviceType;
}

// WebRTC消息
interface WebRTCMessage {
  sdp: string;
  type: 'offer' | 'answer';
}

// ICE候选
interface ICECandidate {
  candidate: string | null;
  sdpMLineIndex: number | null;
  sdpMid: string | null;
}

// 错误消息
interface ErrorMessage {
  error: string;
  error_code: string;
  message: string;
  details?: any;
}
```

---

## 相关文档

- [一对一通话流程](./streaming-one-to-one-call-flow.md) - 完整流程说明
- [WebRTC API 使用](./webrtc-api-guide.md) - WebRTC 前端实现
- [服务端实现](../apps/streaming/internal/handler/signaling.go) - 源码

---

**文档结束**

