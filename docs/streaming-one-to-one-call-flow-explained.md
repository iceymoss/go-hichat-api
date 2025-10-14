# 一对一通话流程详解（SFU架构实现）

> **文档版本**: v2.0  
> **更新日期**: 2025-10-13  
> **重要说明**: 本系统即使在一对一通话中也采用SFU架构

## 🎯 架构说明

### 实际架构（当前实现）

```
用户A ←─WebRTC─→ SFU服务器 ←─WebRTC─→ 用户B
         ↑                              ↑
      独立连接                        独立连接
```

**关键特点**:
- 每个用户都与服务器建立**独立的**WebRTC连接
- 用户A和用户B**不直接连接**
- 服务器作为中间节点转发音视频流
- Offer/Answer是用户与服务器之间协商，而非用户之间

### 为什么采用SFU？

#### 优势
1. **统一架构**: 一对一、群组、会议使用同一套代码
2. **易于扩展**: 随时可以加入第三人，变成群组通话
3. **服务器控制**: 
   - 可以录制通话
   - 可以监控质量
   - 可以审核内容
   - 可以注入媒体（背景音乐等）
4. **NAT穿透**: 避免P2P穿透失败的问题
5. **网络兼容**: 企业防火墙环境更友好

#### 劣势
1. **延迟稍高**: 增加约50-100ms（服务器中转）
2. **服务器负载**: 需要处理所有流量
3. **带宽成本**: 服务器需要上下行带宽

---

## 完整时序图

```
┌─────────────┐              ┌────────────────┐              ┌─────────────┐
│   用户A     │              │  SFU服务器      │              │   用户B     │
│  (主叫方)   │              │                │              │  (被叫方)   │
└──────┬──────┘              └────────┬───────┘              └──────┬──────┘
       │                              │                              │
       
═══════════════════════════════════════════════════════════════════════════
  阶段1: 通话邀请
═══════════════════════════════════════════════════════════════════════════

       │ ① call_invite                │                              │
       ├─────────────────────────────>│                              │
       │                              │ ② call_invite (转发)          │
       │                              ├─────────────────────────────>│
       │                              │ ③ call_accept                │
       │                              │<─────────────────────────────┤
       │ ④ call_accept (通知双方)      │                              │
       │<─────────────────────────────┼─────────────────────────────>│
       │                              │                              │

═══════════════════════════════════════════════════════════════════════════
  阶段2: 加入房间（创建服务器端WebRTC连接）
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │ ⑤ join_room                  │                              │
       ├─────────────────────────────>│                              │
       │                              │                              │
       │                              ├─ 服务器处理:                 │
       │                              │  ① 创建 webrtcConnA          │
       │                              │     = NewWebRTCConnection()  │
       │                              │  ② 保存到 connections:       │
       │                              │     connections["userA"]     │
       │                              │     = webrtcConnA           │
       │                              │  ③ 添加到 SFU:              │
       │                              │     sfu.AddUserToRoom()     │
       │                              │                              │
       │ ⑥ room_info                  │                              │
       │<─────────────────────────────┤                              │
       │                              │                              │
       │                              │ ⑦ join_room                  │
       │                              │<─────────────────────────────┤
       │                              │                              │
       │                              ├─ 服务器处理:                 │
       │                              │  ① 创建 webrtcConnB          │
       │                              │  ② 保存到 connections        │
       │                              │  ③ 添加到 SFU               │
       │                              │                              │
       │                              │ ⑧ room_info                  │
       │                              ├─────────────────────────────>│
       │                              │                              │
       │ ⑨ user_joined                │                              │
       │<─────────────────────────────┤                              │
       │                              │                              │

═══════════════════════════════════════════════════════════════════════════
  阶段3: 用户A与服务器建立WebRTC连接
═══════════════════════════════════════════════════════════════════════════

       │                              │                              │
       │ ⑩ offer                      │                              │
       ├─────────────────────────────>│                              │
       │  发送给服务器                 │                              │
       │  (不是发给B!)                │                              │
       │                              │                              │
       │                              ├─ handleOffer():              │
       │                              │  userID = "user_a_123"       │
       │                              │  webrtcConn = connections["user_a_123"]
       │                              │  ↓                           │
       │                              │  webrtcConn.SetRemoteDescription(offer)
       │                              │  "服务器接受A的offer"         │
       │                              │  ↓                           │
       │                              │  answer = webrtcConn.CreateAnswer()
       │                              │  "服务器创建自己的answer"     │
       │                              │                              │
       │ ⑪ answer                     │                              │
       │<─────────────────────────────┤                              │
       │  服务器返回给A                │                              │
       │  (不是B的answer!)            │                              │
       │                              │                              │
       │ ⑫⑬⑭ ice_candidate           │                              │
       │<════════════════════════════>│                              │
       │  A与服务器交换ICE候选          │                              │
       │                              │                              │
       │ ⑮ WebRTC连接建立:             │                              │
       │    A ↔ 服务器                │                              │
       │═══════════════════════════════╗                             │
       │                              ║                             │

═══════════════════════════════════════════════════════════════════════════
  阶段4: 用户B与服务器建立WebRTC连接（并行或之后）
═══════════════════════════════════════════════════════════════════════════

       │                              ║                             │
       │                              ║ ⑯ offer                      │
       │                              ║<────────────────────────────┤
       │                              ║  发送给服务器                │
       │                              ║  (不是发给A!)                │
       │                              ║                             │
       │                              ╠─ handleOffer():              │
       │                              ║  userID = "user_b_456"       │
       │                              ║  webrtcConn = connections["user_b_456"]
       │                              ║  ↓                           │
       │                              ║  webrtcConn.SetRemoteDescription(offer)
       │                              ║  "服务器接受B的offer"         │
       │                              ║  ↓                           │
       │                              ║  answer = webrtcConn.CreateAnswer()
       │                              ║  "服务器创建自己的answer"     │
       │                              ║                             │
       │                              ║ ⑰ answer                     │
       │                              ╠────────────────────────────>│
       │                              ║  服务器返回给B                │
       │                              ║  (不是A的answer!)            │
       │                              ║                             │
       │                              ║ ⑱⑲⑳ ice_candidate           │
       │                              ║<═══════════════════════════>│
       │                              ║  B与服务器交换ICE候选          │
       │                              ║                             │
       │                              ║ ㉑ WebRTC连接建立:            │
       │                              ║    B ↔ 服务器                │
       │                              ╚═════════════════════════════╗
       │                              ║                             ║

═══════════════════════════════════════════════════════════════════════════
  阶段5: 媒体流通过SFU转发
═══════════════════════════════════════════════════════════════════════════

       │                              ║                             ║
       │ ㉒ A上传自己的音视频流          ║                             ║
       ├═════════════════════════════>║                             ║
       │  通过 webrtcConnA            ║                             ║
       │                              ║                             ║
       │                              ║ ㉓ SFU转发A的流给B           ║
       │                              ╠════════════════════════════>║
       │                              ║  通过 webrtcConnB            ║
       │                              ║                             ║
       │                              ║ ㉔ B上传自己的音视频流        ║
       │                              ║<════════════════════════════╣
       │                              ║  通过 webrtcConnB            ║
       │                              ║                             ║
       │ ㉕ SFU转发B的流给A            ║                             ║
       │<═════════════════════════════╣                             ║
       │  通过 webrtcConnA            ║                             ║
       │                              ║                             ║
```

---

## 🔍 代码流程对照

### 用户A发送Offer的完整过程

#### 前端代码（用户A）
```javascript
// 1. 用户A收到 user_joined 通知后
case 'user_joined':
    // 2. 创建Offer（与本地PeerConnection）
    const offer = await this.pc.createOffer();
    await this.pc.setLocalDescription(offer);
    
    // 3. 发送Offer给服务器（注意：是给服务器！）
    this.sendSignalingMessage({
        type: 'offer',
        room_id: this.roomId,
        user_id: 'user_a_123',
        data: {
            sdp: offer.sdp,
            type: 'offer'
        }
    });
```

#### 服务器处理（signaling.go:368-428）
```go
func (s *SignalingServer) handleOffer(conn *websocket.Conn, msg *types.SignalingMessage) {
    userID := msg.UserID  // "user_a_123"
    
    // 1. 获取用户A与服务器的连接
    webrtcConn := s.connections[userID]  // 这是A与服务器的连接！
    
    // 2. 服务器接受用户A的Offer
    offer := libRTC.SessionDescription{
        Type: libRTC.SDPTypeOffer,
        SDP:  sdp,
    }
    webrtcConn.SetRemoteDescription(offer)
    // 意思：服务器说"我知道你A的媒体能力了"
    
    // 3. 服务器创建Answer
    answer := webrtcConn.CreateAnswer()
    // 意思：服务器说"这是我(服务器)的媒体能力"
    
    // 4. 服务器把Answer发回给用户A
    s.sendMessage(conn, &types.SignalingMessage{
        Type:   types.MessageTypeAnswer,
        UserID: userID,  // 还是 "user_a_123"
        Data:   answer,
    })
    // 注意：这个answer是发给A的，不是转发给B的！
}
```

#### 前端代码（用户A收到Answer）
```javascript
// 4. 用户A收到服务器的Answer
case 'answer':
    // 5. 设置远程描述（这是服务器的描述）
    await this.pc.setRemoteDescription(
        new RTCSessionDescription(message.data)
    );
    // 现在用户A和服务器完成了媒体协商
```

---

### 用户B的独立协商过程

**重要**: 用户B也会**独立**进行同样的流程

```javascript
// 用户B也需要发送Offer给服务器
const offer = await this.pc.createOffer();
this.sendSignalingMessage({
    type: 'offer',
    room_id: this.roomId,
    user_id: 'user_b_456',
    data: { sdp: offer.sdp }
});

// 服务器处理B的Offer（独立处理）
handleOffer() {
    userID = "user_b_456"
    webrtcConn = connections["user_b_456"]  // B与服务器的连接
    webrtcConn.SetRemoteDescription(offerB)
    answerB = webrtcConn.CreateAnswer()
    sendMessage(answerB)  // 发给B
}

// 用户B收到服务器的Answer
await this.pc.setRemoteDescription(answerB);
```

---

## 🔑 关键数据结构

### 服务器端的连接映射

```go
type SignalingServer struct {
    // 每个用户都有独立的WebRTC连接
    connections map[string]*webrtc.WebRTCConnection
    //          ^                    ^
    //       userID              与服务器的连接
}

// 实例:
connections = {
    "user_a_123": webrtcConnA,  // A与服务器的PeerConnection
    "user_b_456": webrtcConnB,  // B与服务器的PeerConnection
}
```

### SFU中的用户管理

```go
type SFU struct {
    rooms map[string]*RoomSFU
}

type RoomSFU struct {
    users map[string]*UserSFU
}

type UserSFU struct {
    userID   string
    peerConn *webrtc.PeerConnection  // 与服务器的连接
    tracks   map[string]*webrtc.TrackLocalStaticRTP  // 用户的媒体轨道
}

// 实例:
sfu.rooms["room_123"] = {
    users: {
        "user_a_123": {
            peerConn: <A与服务器的连接>,
            tracks: {
                "audio_track_1": <A的音频流>,
                "video_track_1": <A的视频流>
            }
        },
        "user_b_456": {
            peerConn: <B与服务器的连接>,
            tracks: {
                "audio_track_2": <B的音频流>,
                "video_track_2": <B的视频流>
            }
        }
    }
}
```

---

## 📊 完整的Offer/Answer序列

### 用户A的协商（A ↔ 服务器）

```json
// ① A发送Offer给服务器
{
  "type": "offer",
  "room_id": "room_123",
  "user_id": "user_a_123",
  "data": {
    "sdp": "v=0\r\no=- 4611731400430051336 2 IN IP4 127.0.0.1\r\n...",
    "type": "offer"
  }
}

// ② 服务器返回Answer给A
{
  "type": "answer",
  "room_id": "room_123",
  "user_id": "user_a_123",  // 注意：还是A
  "data": {
    "sdp": "v=0\r\no=- 8888888888888888888 2 IN IP4 10.0.0.1\r\n...",
    "type": "answer"
  }
}
// 这个Answer是服务器生成的，描述的是服务器的媒体能力
```

### 用户B的协商（B ↔ 服务器）

```json
// ③ B发送Offer给服务器（独立进行）
{
  "type": "offer",
  "room_id": "room_123",
  "user_id": "user_b_456",
  "data": {
    "sdp": "v=0\r\no=- 9876543210987654321 2 IN IP4 192.168.1.100\r\n...",
    "type": "offer"
  }
}

// ④ 服务器返回Answer给B（独立的Answer）
{
  "type": "answer",
  "room_id": "room_123",
  "user_id": "user_b_456",  // 注意：是B
  "data": {
    "sdp": "v=0\r\no=- 9999999999999999999 2 IN IP4 10.0.0.1\r\n...",
    "type": "answer"
  }
}
// 这个Answer也是服务器生成的，是另一个独立的Answer
```

---

## 🎬 实际运行流程示例

### 完整的消息序列

```
时间线    用户A                    服务器                      用户B
----------------------------------------------------------------------
10:00:00  发送 call_invite    →
10:00:00                      ← 接收处理
10:00:00                        转发 call_invite         →
10:00:05                      ← 接收 call_accept
10:00:05  接收 call_accept    ←   通知双方                →  接收 call_accept

10:00:06  发送 join_room      →
10:00:06                      ← 创建 webrtcConnA
10:00:06                        添加到 connections["user_a_123"]
10:00:06                        添加到 SFU
10:00:06  接收 room_info      ←

10:00:07                      ← 接收 join_room
10:00:07                        创建 webrtcConnB
10:00:07                        添加到 connections["user_b_456"]
10:00:07                        添加到 SFU
10:00:07                        发送 room_info           →
10:00:07  接收 user_joined    ←

10:00:08  发送 offer          →
          (offerA)
10:00:08                      ← handleOffer(userA)
                                webrtcConnA.SetRemoteDescription
                                answerA = webrtcConnA.CreateAnswer
10:00:08  接收 answer         ←
          (answerA from 服务器)
10:00:08  setRemoteDescription

10:00:09  发送 ice_candidate  →
10:00:09  发送 ice_candidate  →
10:00:09  接收 ice_candidate  ←  (服务器的候选)
10:00:10  接收 ice_candidate  ←

10:00:10  [连接建立: A ↔ 服务器]

10:00:11                      ← 接收 offer
                                (offerB)
10:00:11                        handleOffer(userB)
                                webrtcConnB.SetRemoteDescription
                                answerB = webrtcConnB.CreateAnswer
10:00:11                        发送 answer              →
                                                            (answerB from 服务器)
10:00:11                      ← 接收 ice_candidate
10:00:11                      ← 接收 ice_candidate
10:00:11                        发送 ice_candidate        →
10:00:12                        发送 ice_candidate        →

10:00:12                        [连接建立: B ↔ 服务器]

10:00:13  上传媒体流          →   SFU接收
10:00:13  接收对方流          ←   SFU转发B的流
10:00:13                          SFU接收              ←  上传媒体流
10:00:13                          SFU转发A的流          →  接收对方流

10:00:13  [双向媒体流转发建立完成]
```

---

## 💡 理解要点

### 1. 没有用户间的信令交换

❌ **不是这样**:
```
A的offer → 服务器转发 → B
B的answer → 服务器转发 → A
```

✅ **而是这样**:
```
A的offer → 服务器接受 → 服务器的answer → A
B的offer → 服务器接受 → 服务器的answer → B
```

---

### 2. 两个独立的WebRTC连接

```
连接1: 用户A ↔ 服务器
  - offerA: A的媒体能力
  - answerA: 服务器的媒体能力
  - ICE候选: A与服务器之间的网络路径

连接2: 用户B ↔ 服务器  
  - offerB: B的媒体能力
  - answerB: 服务器的媒体能力
  - ICE候选: B与服务器之间的网络路径
```

---

### 3. 服务器的角色

**不是**: 信令转发服务器（纯中继）  
**而是**: WebRTC对端（媒体处理节点）

服务器做的事情：
1. 与用户A建立WebRTC连接，接收A的流
2. 与用户B建立WebRTC连接，接收B的流
3. 将A的流转发给B
4. 将B的流转发给A

---

### 4. connections 映射的含义

```go
s.connections[userID] = webrtcConn
```

这个映射存储的是：**每个用户与服务器的WebRTC连接**

不是：用户之间的连接映射

---

## ⚖️ SFU vs P2P 对比

### P2P模式（传统WebRTC）

```
消息流:
A发offer → 服务器 → 转发给B
B收offer → 创建answer → 服务器 → 转发给A
A收answer → 完成协商

连接:
A ←───────直接连接───────→ B

服务器只做信令转发，不处理媒体
```

### SFU模式（当前实现）

```
消息流:
A发offer → 服务器处理 → 服务器的answer → A
B发offer → 服务器处理 → 服务器的answer → B

连接:
A ←─→ 服务器 ←─→ B

服务器处理所有媒体流
```

---

## 🎯 总结

### 你的问题的答案

**问**: offer如何从A转给B的？  
**答**: **不转发！** 当前实现中：
- A的offer只给服务器
- 服务器处理后返回answer给A
- B也会单独发送自己的offer给服务器
- 服务器返回另一个answer给B
- A和B之间没有直接的offer/answer交换

### handleOffer代码的含义

```go
func (s *SignalingServer) handleOffer(conn *websocket.Conn, msg *types.SignalingMessage) {
    // conn: 发送者的WebSocket连接
    // msg.UserID: 发送者的ID（可能是A或B）
    
    // 获取发送者与服务器的WebRTC连接
    webrtcConn := s.connections[msg.UserID]
    
    // 服务器作为对端处理Offer
    webrtcConn.SetRemoteDescription(offer)
    answer := webrtcConn.CreateAnswer()
    
    // 把Answer返回给发送者（不是转发给对方！）
    s.sendMessage(conn, answer)
}
```

### 为什么这样设计？

1. **架构统一**: 一对一和群组用同一套逻辑
2. **服务器可控**: 可以录制、审核、注入
3. **易于扩展**: 第三人随时可以加入
4. **避免穿透**: 不需要担心NAT穿透失败

---

现在清楚了吗？关键是理解：**即使是一对一，也是双星架构（两个用户都连接到服务器），而不是P2P直连**。🌟
