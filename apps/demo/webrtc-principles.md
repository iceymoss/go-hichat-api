# WebRTC 工作原理详解

本文档详细介绍 WebRTC (Web Real-Time Communication) 的核心概念、工作原理和技术架构。

---

## 📖 目录

- [WebRTC 简介](#webrtc-简介)
- [核心架构](#核心架构)
- [三大核心 API](#三大核心-api)
- [信令机制](#信令机制)
- [NAT 穿透技术](#nat-穿透技术)
- [ICE 协议详解](#ice-协议详解)
- [SDP 协议详解](#sdp-协议详解)
- [媒体协商流程](#媒体协商流程)
- [音视频编解码](#音视频编解码)
- [安全机制](#安全机制)
- [架构模式](#架构模式)
- [性能优化](#性能优化)
- [应用场景](#应用场景)

---

## 🌐 WebRTC 简介

### **什么是 WebRTC？**

WebRTC (Web Real-Time Communication) 是一个开源项目，旨在为浏览器和移动应用提供**实时通信**能力。

**核心特性：**
- 🎥 **实时音视频通信** - 无需插件，原生支持
- 🔄 **点对点连接** - P2P 直连，低延迟
- 🌍 **跨平台** - 浏览器、移动端、桌面端
- 🔒 **安全加密** - 强制使用加密传输
- 📡 **数据传输** - 支持任意数据传输

### **发展历史**

```
2011年 - Google 收购 GIPS 公司，获得音视频处理技术
2011年 - Google 开源 WebRTC 项目
2012年 - Chrome、Firefox 开始支持 WebRTC
2017年 - Safari 11 加入 WebRTC 支持
2021年 - WebRTC 1.0 成为 W3C 正式标准
```

### **主要应用**

- 📞 视频通话（Zoom、Google Meet、Microsoft Teams）
- 🎮 云游戏（Google Stadia）
- 📺 直播推流
- 🏥 远程医疗
- 📚 在线教育
- 🤝 远程协作

---

## 🏗️ 核心架构

### **整体架构图**

```
┌─────────────────────────────────────────────────────────┐
│                    应用层 (Application)                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   视频通话   │  │   屏幕共享   │  │   文件传输   │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────┐
│                  WebRTC API 层                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ RTCPeerCon-  │  │ MediaStream  │  │ RTCDataChan- │  │
│  │  nection     │  │              │  │    nel       │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────┐
│                  WebRTC 引擎层                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐ │
│  │  会话管理 │  │ 音视频引擎│  │  网络传输 │  │ 安全加密│ │
│  └──────────┘  └──────────┘  └──────────┘  └─────────┘ │
└─────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────┐
│                  协议栈层                                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐ │
│  │   SDP    │  │   ICE    │  │   DTLS   │  │  SRTP   │ │
│  └──────────┘  └──────────┘  └──────────┘  └─────────┘ │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │   STUN   │  │   TURN   │  │   SCTP   │              │
│  └──────────┘  └──────────┘  └──────────┘              │
└─────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────┐
│                  传输层                                  │
│                    UDP / TCP                             │
└─────────────────────────────────────────────────────────┘
```

### **协议栈详解**

| 协议 | 全称 | 作用 |
|------|------|------|
| **SDP** | Session Description Protocol | 会话描述，协商媒体能力 |
| **ICE** | Interactive Connectivity Establishment | 交互式连接建立，NAT 穿透 |
| **STUN** | Session Traversal Utilities for NAT | NAT 会话穿越工具 |
| **TURN** | Traversal Using Relays around NAT | 中继方式穿越 NAT |
| **DTLS** | Datagram Transport Layer Security | 数据报传输层安全协议 |
| **SRTP** | Secure Real-time Transport Protocol | 安全实时传输协议 |
| **SCTP** | Stream Control Transmission Protocol | 流控制传输协议 |
| **RTP** | Real-time Transport Protocol | 实时传输协议 |

---

## 🔧 三大核心 API

### **1. RTCPeerConnection**

**作用：** 管理对等连接，处理音视频流的发送和接收。

**核心功能：**
- 建立和维护 P2P 连接
- 处理 ICE 候选
- 管理媒体流
- 统计和监控连接质量

**生命周期：**

```javascript
// 1. 创建连接
const pc = new RTCPeerConnection({
    iceServers: [
        { urls: 'stun:stun.l.google.com:19302' }
    ]
});

// 2. 添加本地媒体流
localStream.getTracks().forEach(track => {
    pc.addTrack(track, localStream);
});

// 3. 处理远程媒体流
pc.ontrack = (event) => {
    remoteVideo.srcObject = event.streams[0];
};

// 4. 处理 ICE 候选
pc.onicecandidate = (event) => {
    if (event.candidate) {
        sendToRemotePeer(event.candidate);
    }
};

// 5. 创建 Offer (呼叫方)
const offer = await pc.createOffer();
await pc.setLocalDescription(offer);
sendToRemotePeer(offer);

// 6. 创建 Answer (被叫方)
await pc.setRemoteDescription(remoteOffer);
const answer = await pc.createAnswer();
await pc.setLocalDescription(answer);
sendToRemotePeer(answer);

// 7. 添加远程 ICE 候选
await pc.addIceCandidate(remoteCandidate);

// 8. 监听连接状态
pc.onconnectionstatechange = () => {
    console.log('连接状态:', pc.connectionState);
    // new, connecting, connected, disconnected, failed, closed
};

// 9. 关闭连接
pc.close();
```

**连接状态机：**

```
[new] 
  ↓ 开始协商
[connecting] 
  ↓ ICE 协商成功
[connected] 
  ↓ 网络问题
[disconnected] 
  ↓ 尝试重连失败
[failed]

或者主动关闭：
[connected] 
  ↓ 调用 close()
[closed]
```

---

### **2. MediaStream**

**作用：** 表示媒体流（音频/视频），可以包含多个轨道（Track）。

**核心概念：**

```
MediaStream
    ├── VideoTrack 1 (摄像头)
    ├── VideoTrack 2 (屏幕共享)
    ├── AudioTrack 1 (麦克风)
    └── AudioTrack 2 (系统音频)
```

**获取媒体流：**

```javascript
// 1. 获取摄像头和麦克风
const stream = await navigator.mediaDevices.getUserMedia({
    video: {
        width: { ideal: 1280 },
        height: { ideal: 720 },
        frameRate: { ideal: 30 }
    },
    audio: {
        echoCancellation: true,  // 回声消除
        noiseSuppression: true,  // 噪声抑制
        autoGainControl: true    // 自动增益
    }
});

// 2. 获取屏幕共享
const screenStream = await navigator.mediaDevices.getDisplayMedia({
    video: {
        cursor: 'always',
        displaySurface: 'monitor'
    },
    audio: false
});

// 3. 获取可用设备列表
const devices = await navigator.mediaDevices.enumerateDevices();
devices.forEach(device => {
    console.log(device.kind, device.label);
    // videoinput, audioinput, audiooutput
});
```

**操作媒体轨道：**

```javascript
// 获取所有轨道
stream.getTracks().forEach(track => {
    console.log(track.kind, track.label);
});

// 获取视频轨道
const videoTracks = stream.getVideoTracks();

// 获取音频轨道
const audioTracks = stream.getAudioTracks();

// 添加轨道
stream.addTrack(newTrack);

// 移除轨道
stream.removeTrack(track);

// 停止轨道
track.stop();

// 启用/禁用轨道
track.enabled = false;  // 静音/关闭视频
track.enabled = true;   // 取消静音/开启视频

// 监听轨道结束
track.onended = () => {
    console.log('轨道已结束');
};
```

---

### **3. RTCDataChannel**

**作用：** 提供点对点的数据通道，用于传输任意数据。

**特性：**
- 低延迟（< 100ms）
- 可靠或不可靠传输
- 有序或无序传输
- 流量控制
- 支持二进制数据

**创建数据通道：**

```javascript
// 呼叫方创建数据通道
const dataChannel = pc.createDataChannel('chat', {
    ordered: true,           // 有序传输
    maxRetransmits: 3,       // 最大重传次数
    // 或使用 maxPacketLifeTime: 3000  // 最大生命周期（毫秒）
});

// 被叫方接收数据通道
pc.ondatachannel = (event) => {
    const dataChannel = event.channel;
    setupDataChannel(dataChannel);
};

// 设置数据通道
function setupDataChannel(channel) {
    channel.onopen = () => {
        console.log('数据通道已打开');
        channel.send('Hello!');
    };
    
    channel.onmessage = (event) => {
        console.log('收到消息:', event.data);
    };
    
    channel.onerror = (error) => {
        console.error('数据通道错误:', error);
    };
    
    channel.onclose = () => {
        console.log('数据通道已关闭');
    };
}

// 发送文本数据
dataChannel.send('Hello World');

// 发送二进制数据
const buffer = new ArrayBuffer(8);
dataChannel.send(buffer);

// 发送 Blob
const blob = new Blob(['Hello'], { type: 'text/plain' });
dataChannel.send(blob);

// 检查缓冲区
if (dataChannel.bufferedAmount < threshold) {
    dataChannel.send(data);
}
```

**应用场景：**
- 💬 实时聊天
- 📁 文件传输
- 🎮 游戏数据同步
- 📊 实时数据可视化
- 🤝 协同编辑

---

## 📡 信令机制

### **什么是信令？**

WebRTC **不包含**信令协议，需要开发者自己实现。信令用于：

1. **会话控制** - 建立、修改、终止会话
2. **能力协商** - 交换 SDP 信息
3. **网络信息** - 交换 ICE 候选

### **信令流程**

```
客户端 A                信令服务器              客户端 B
    │                      │                      │
    │──── 连接 ───────────►│                      │
    │◄─── 分配 ID ─────────│                      │
    │                      │◄──── 连接 ───────────│
    │                      │──── 分配 ID ─────────►│
    │                      │                      │
    │──── offer ──────────►│                      │
    │                      │──── offer ──────────►│
    │                      │                      │
    │                      │◄─── answer ──────────│
    │◄─── answer ──────────│                      │
    │                      │                      │
    │──── candidate ──────►│──── candidate ──────►│
    │◄─── candidate ───────│◄─── candidate ───────│
    │                      │                      │
    └──────────────── P2P 连接建立 ───────────────┘
```

### **常用信令方案**

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **WebSocket** | 双向通信、实时性好 | 需要服务器支持 | 通用场景 |
| **Socket.IO** | 自动降级、易用 | 包体积大 | 快速开发 |
| **SIP** | 标准协议、成熟 | 复杂度高 | 企业级应用 |
| **XMPP** | 可扩展、标准化 | XML 开销大 | IM 系统 |
| **自定义 HTTP** | 简单、兼容性好 | 非实时 | 简单应用 |

### **信令消息示例**

**Offer 消息：**
```json
{
  "type": "offer",
  "from": "client-a",
  "to": "client-b",
  "sdp": "v=0\r\no=- 4611731400430051336 2 IN IP4 127.0.0.1\r\n..."
}
```

**Answer 消息：**
```json
{
  "type": "answer",
  "from": "client-b",
  "to": "client-a",
  "sdp": "v=0\r\no=- 9876543210987654321 2 IN IP4 127.0.0.1\r\n..."
}
```

**ICE Candidate 消息：**
```json
{
  "type": "candidate",
  "from": "client-a",
  "to": "client-b",
  "candidate": {
    "candidate": "candidate:1 1 udp 2130706431 192.168.1.100 54321 typ host",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  }
}
```

---

## 🌉 NAT 穿透技术

### **什么是 NAT？**

NAT (Network Address Translation) 网络地址转换，允许多个设备共享一个公网 IP。

**NAT 类型：**

```
┌─────────────────────────────────────────────────┐
│ 1. Full Cone NAT (完全锥形 NAT)                 │
│    最宽松，任何外部主机都可以访问                │
│    穿透难度: ★☆☆☆☆                              │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│ 2. Restricted Cone NAT (限制锥形 NAT)           │
│    只有之前通信过的 IP 可以访问                  │
│    穿透难度: ★★☆☆☆                              │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│ 3. Port Restricted Cone NAT (端口限制锥形 NAT)  │
│    只有之前通信过的 IP:Port 可以访问             │
│    穿透难度: ★★★☆☆                              │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│ 4. Symmetric NAT (对称 NAT)                     │
│    每个外部地址分配不同的端口映射                │
│    穿透难度: ★★★★★ (需要 TURN)                  │
└─────────────────────────────────────────────────┘
```

### **STUN 服务器**

**STUN (Session Traversal Utilities for NAT)**

**作用：** 帮助客户端发现自己的公网 IP 和端口。

**工作流程：**

```
┌──────────────┐         ┌──────────────┐
│   客户端     │         │ STUN 服务器  │
│  (内网 IP:   │         │ (公网 IP:    │
│ 192.168.1.10)│         │ 1.2.3.4)     │
└──────┬───────┘         └──────┬───────┘
       │                        │
       │  1. STUN 请求          │
       │  来自: 192.168.1.10    │
       │───────────────────────►│
       │                        │
       │  2. 从 NAT 看到请求    │
       │  来自: 5.6.7.8:54321   │
       │                        │
       │  3. STUN 响应          │
       │  你的公网地址是:       │
       │  5.6.7.8:54321         │
       │◄───────────────────────│
       │                        │
```

**常用公共 STUN 服务器：**
```javascript
const iceServers = [
    { urls: 'stun:stun.l.google.com:19302' },
    { urls: 'stun:stun1.l.google.com:19302' },
    { urls: 'stun:stun2.l.google.com:19302' },
    { urls: 'stun:stun3.l.google.com:19302' },
    { urls: 'stun:stun4.l.google.com:19302' }
];
```

### **TURN 服务器**

**TURN (Traversal Using Relays around NAT)**

**作用：** 当 P2P 无法建立时，通过中继服务器转发数据。

**工作流程：**

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│   客户端 A   │         │ TURN 服务器  │         │   客户端 B   │
│  (对称 NAT)  │         │ (公网服务器) │         │  (对称 NAT)  │
└──────┬───────┘         └──────┬───────┘         └──────┬───────┘
       │                        │                        │
       │  1. 分配请求           │                        │
       │───────────────────────►│                        │
       │  2. 分配地址           │                        │
       │◄───────────────────────│                        │
       │                        │                        │
       │                        │  3. 分配请求           │
       │                        │◄───────────────────────│
       │                        │  4. 分配地址           │
       │                        │───────────────────────►│
       │                        │                        │
       │  5. 发送数据           │                        │
       │───────────────────────►│  6. 转发数据           │
       │                        │───────────────────────►│
       │                        │                        │
       │                        │  7. 发送数据           │
       │  8. 转发数据           │◄───────────────────────│
       │◄───────────────────────│                        │
```

**TURN 配置：**
```javascript
const iceServers = [
    { urls: 'stun:stun.example.com:3478' },
    {
        urls: 'turn:turn.example.com:3478',
        username: 'user',
        credential: 'password'
    }
];
```

**成本考虑：**
- STUN: 免费或低成本（只用于初始协商）
- TURN: 高成本（转发所有媒体流）
  - 带宽消耗：每个通话约 2-5 Mbps
  - 通常只有 5-10% 的连接需要 TURN

---

## 🧊 ICE 协议详解

### **ICE (Interactive Connectivity Establishment)**

**作用：** 寻找客户端之间的最佳连接路径。

### **ICE 候选类型**

```
┌─────────────────────────────────────────────────┐
│ 1. Host Candidate (主机候选)                    │
│    ├─ 本地网卡的 IP 地址                        │
│    ├─ 优先级: 最高                              │
│    └─ 示例: 192.168.1.100:54321                │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│ 2. Server Reflexive Candidate (服务器反射候选)  │
│    ├─ 通过 STUN 获取的公网地址                  │
│    ├─ 优先级: 中等                              │
│    └─ 示例: 5.6.7.8:54321                       │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│ 3. Relay Candidate (中继候选)                   │
│    ├─ 通过 TURN 服务器中继                      │
│    ├─ 优先级: 最低                              │
│    └─ 示例: 1.2.3.4:3478                        │
└─────────────────────────────────────────────────┘
```

### **ICE 状态机**

```
[new] 
  ↓ 开始收集候选
[gathering] 
  ↓ 收集完成
[complete] 

并行进行连接检查：

[checking] 
  ↓ 找到可用路径
[connected] 
  ↓ 完成所有检查
[completed]
```

### **ICE 连接检查**

ICE 会尝试所有可能的候选对组合：

```
客户端 A 候选:
  - Host: 192.168.1.10:54321
  - Srflx: 5.6.7.8:54321
  - Relay: 1.2.3.4:3478

客户端 B 候选:
  - Host: 192.168.1.20:54322
  - Srflx: 9.10.11.12:54322
  - Relay: 1.2.3.4:3479

可能的候选对 (9种组合):
  ✓ A-Host    ↔ B-Host     (同一局域网，最优)
  ✗ A-Host    ↔ B-Srflx    (不可达)
  ✗ A-Host    ↔ B-Relay    (不可达)
  ✗ A-Srflx   ↔ B-Host     (不可达)
  ✓ A-Srflx   ↔ B-Srflx    (P2P 公网，次优)
  ✗ A-Srflx   ↔ B-Relay    (不必要)
  ✗ A-Relay   ↔ B-Host     (不可达)
  ✗ A-Relay   ↔ B-Srflx    (不必要)
  ✓ A-Relay   ↔ B-Relay    (通过 TURN，保底方案)
```

**选择策略：**
1. 优先使用 Host 候选（局域网直连）
2. 其次使用 Server Reflexive（公网 P2P）
3. 最后使用 Relay（TURN 中继）

### **ICE Trickle**

**传统 ICE：** 等待所有候选收集完成后再开始连接检查

**Trickle ICE：** 一边收集候选，一边进行连接检查

```javascript
// 传统方式
pc.onicecandidate = (event) => {
    if (!event.candidate) {
        // 所有候选收集完成
        sendAllCandidates();
    }
};

// Trickle ICE (推荐)
pc.onicecandidate = (event) => {
    if (event.candidate) {
        // 立即发送每个候选
        sendCandidate(event.candidate);
    }
};
```

**优势：**
- 更快建立连接（节省 1-2 秒）
- 更好的用户体验

---

## 📄 SDP 协议详解

### **SDP (Session Description Protocol)**

**作用：** 描述多媒体会话的参数。

### **SDP 结构**

```
v=0                                    // 版本
o=- 4611731400430051336 2 IN IP4 127.0.0.1  // 会话源
s=-                                    // 会话名
t=0 0                                  // 时间描述
a=group:BUNDLE 0 1                     // 媒体分组
a=msid-semantic: WMS stream_id         // 媒体流语义

m=audio 9 UDP/TLS/RTP/SAVPF 111 103 104  // 音频媒体描述
c=IN IP4 0.0.0.0                       // 连接信息
a=rtcp:9 IN IP4 0.0.0.0                // RTCP 端口
a=ice-ufrag:abc123                     // ICE 用户名片段
a=ice-pwd:def456                       // ICE 密码
a=fingerprint:sha-256 XX:XX:XX:...     // DTLS 指纹
a=setup:actpass                        // DTLS 角色
a=mid:0                                // 媒体 ID
a=sendrecv                             // 媒体方向
a=rtcp-mux                             // RTCP 复用
a=rtpmap:111 opus/48000/2              // 编解码器映射
a=fmtp:111 minptime=10;useinbandfec=1  // 格式参数
a=ssrc:123456789 cname:user@host       // 同步源

m=video 9 UDP/TLS/RTP/SAVPF 96 97 98   // 视频媒体描述
c=IN IP4 0.0.0.0
a=rtcp:9 IN IP4 0.0.0.0
a=ice-ufrag:abc123
a=ice-pwd:def456
a=fingerprint:sha-256 XX:XX:XX:...
a=setup:actpass
a=mid:1
a=sendrecv
a=rtcp-mux
a=rtcp-fb:96 nack                      // RTCP 反馈
a=rtcp-fb:96 nack pli                  // 关键帧请求
a=rtcp-fb:96 ccm fir                   // 完整帧请求
a=rtpmap:96 VP8/90000                  // VP8 编解码器
a=rtpmap:97 H264/90000                 // H.264 编解码器
a=fmtp:97 profile-level-id=42e01f      // H.264 配置
a=ssrc:987654321 cname:user@host
```

### **关键字段说明**

| 字段 | 说明 | 示例 |
|------|------|------|
| `v=` | 版本号 | `v=0` |
| `o=` | 会话源标识 | `o=- 123 2 IN IP4 127.0.0.1` |
| `s=` | 会话名称 | `s=-` |
| `t=` | 会话时间 | `t=0 0` (永久会话) |
| `m=` | 媒体描述 | `m=video 9 UDP/TLS/RTP/SAVPF 96` |
| `a=` | 属性 | `a=sendrecv` |
| `c=` | 连接信息 | `c=IN IP4 0.0.0.0` |

### **媒体方向**

```javascript
// sendrecv - 发送和接收 (默认)
a=sendrecv

// sendonly - 仅发送 (直播推流)
a=sendonly

// recvonly - 仅接收 (直播观看)
a=recvonly

// inactive - 不活动 (暂时禁用)
a=inactive
```

### **SDP Offer/Answer 交换**

```javascript
// Caller 创建 Offer
const offer = await pc.createOffer({
    offerToReceiveAudio: true,
    offerToReceiveVideo: true,
    iceRestart: false  // 是否重启 ICE
});

await pc.setLocalDescription(offer);

// 发送到对端
sendToRemote(pc.localDescription);

// Callee 处理 Offer，创建 Answer
await pc.setRemoteDescription(remoteOffer);

const answer = await pc.createAnswer();
await pc.setLocalDescription(answer);

// 发送到对端
sendToRemote(pc.localDescription);

// Caller 处理 Answer
await pc.setRemoteDescription(remoteAnswer);
```

---

## 🎵 媒体协商流程

### **完整协商流程**

```
客户端 A (Caller)                   客户端 B (Callee)
      │                                    │
      │ 1. 创建 PeerConnection             │
      │ 2. 添加本地媒体流                  │
      │                                    │
      │ 3. createOffer()                   │
      │    - 生成 SDP Offer                │
      │    - 包含支持的编解码器            │
      │    - 包含媒体方向                  │
      │                                    │
      │ 4. setLocalDescription(offer)      │
      │    - 开始 ICE 收集                 │
      │                                    │
      │ 5. 发送 Offer ──────────────────►  │
      │                                    │ 6. setRemoteDescription(offer)
      │                                    │    - 解析对端能力
      │                                    │
      │                                    │ 7. 创建 PeerConnection
      │                                    │ 8. 添加本地媒体流
      │                                    │
      │                                    │ 9. createAnswer()
      │                                    │    - 根据 Offer 生成 Answer
      │                                    │    - 选择兼容的编解码器
      │                                    │
      │                                    │ 10. setLocalDescription(answer)
      │                                    │     - 开始 ICE 收集
      │                                    │
      │  ◄──────────────── 11. 发送 Answer │
      │                                    │
      │ 12. setRemoteDescription(answer)   │
      │     - 确认媒体参数                 │
      │                                    │
      ├───── 13. ICE 候选交换 ──────────────┤
      │                                    │
      ├────── 14. DTLS 握手 ───────────────┤
      │                                    │
      ├═══════ 15. 媒体流传输 ═════════════┤
```

### **编解码器协商**

**Offer 中的编解码器列表（优先级递减）：**
```
m=video 9 UDP/TLS/RTP/SAVPF 96 97 98 99
a=rtpmap:96 VP8/90000
a=rtpmap:97 VP9/90000
a=rtpmap:98 H264/90000
a=rtpmap:99 AV1/90000
```

**Answer 中的选择（只保留双方都支持的）：**
```
m=video 9 UDP/TLS/RTP/SAVPF 96 98
a=rtpmap:96 VP8/90000
a=rtpmap:98 H264/90000
```

---

## 🎬 音视频编解码

### **视频编解码器**

| 编解码器 | 质量 | 性能 | 浏览器支持 | 适用场景 |
|----------|------|------|------------|----------|
| **VP8** | 中等 | 高 | ✅ 全支持 | 通用场景 |
| **VP9** | 高 | 中等 | ✅ 较好 | 高质量场景 |
| **H.264** | 高 | 高 | ✅ 全支持 | 硬件加速 |
| **AV1** | 很高 | 低 | ⚠️ 部分 | 未来趋势 |

**视频参数：**
```javascript
const constraints = {
    video: {
        width: { min: 640, ideal: 1280, max: 1920 },
        height: { min: 480, ideal: 720, max: 1080 },
        frameRate: { ideal: 30, max: 60 },
        facingMode: 'user',  // 'user' 前置, 'environment' 后置
    }
};
```

### **音频编解码器**

| 编解码器 | 质量 | 延迟 | 带宽 | 适用场景 |
|----------|------|------|------|----------|
| **Opus** | 很高 | 低 | 低-高 | WebRTC 首选 |
| **G.711** | 中等 | 很低 | 固定 | 电话质量 |
| **PCMU/PCMA** | 低 | 很低 | 高 | 传统电话 |
| **iSAC** | 高 | 低 | 自适应 | 弱网环境 |

**Opus 特性：**
- 采样率：48 kHz
- 比特率：6-510 kbps（自适应）
- 延迟：5-66.5 ms
- 支持立体声
- 前向纠错 (FEC)

**音频参数：**
```javascript
const constraints = {
    audio: {
        echoCancellation: true,      // 回声消除
        noiseSuppression: true,      // 噪声抑制
        autoGainControl: true,       // 自动增益
        sampleRate: 48000,           // 采样率
        channelCount: 2,             // 声道数
        latency: 0.01                // 延迟（秒）
    }
};
```

### **带宽控制**

```javascript
// 设置发送参数
const sender = pc.getSenders().find(s => s.track.kind === 'video');
const parameters = sender.getParameters();

if (!parameters.encodings) {
    parameters.encodings = [{}];
}

// 限制最大比特率
parameters.encodings[0].maxBitrate = 1000000;  // 1 Mbps

// 设置缩放因子
parameters.encodings[0].scaleResolutionDownBy = 2;  // 分辨率减半

await sender.setParameters(parameters);
```

---

## 🔒 安全机制

### **强制加密**

WebRTC **强制使用加密**，无法关闭：

```
音视频流:
  RTP → SRTP (Secure RTP)
  使用 AES 加密

数据通道:
  SCTP → DTLS (Datagram TLS)
  使用 TLS 1.2+
```

### **DTLS 握手**

```
客户端 A                         客户端 B
    │                                │
    │  1. DTLS ClientHello           │
    │───────────────────────────────►│
    │                                │
    │  2. DTLS ServerHello           │
    │◄───────────────────────────────│
    │                                │
    │  3. Certificate                │
    │◄───────────────────────────────│
    │                                │
    │  4. Certificate                │
    │───────────────────────────────►│
    │                                │
    │  5. Finished                   │
    │◄──────────────────────────────►│
    │                                │
    │  加密通道建立                  │
```

### **证书指纹验证**

SDP 中包含证书指纹，防止中间人攻击：

```
a=fingerprint:sha-256 XX:YY:ZZ:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66
```

### **安全最佳实践**

1. **使用 HTTPS** - 避免中间人攻击信令
2. **验证信令** - 对信令消息进行签名/加密
3. **用户认证** - 在应用层实现用户认证
4. **权限控制** - 检查摄像头/麦克风权限
5. **隐私保护** - 提示用户正在进行录制/传输

---

## 🏛️ 架构模式

### **1. Mesh (网状)**

**特点：** 每个客户端与其他所有客户端建立 P2P 连接

```
        A
       /│\
      / │ \
     /  │  \
    B───┼───C
     \  │  /
      \ │ /
       \│/
        D

连接数 = n(n-1)/2
4 人 = 6 条连接
10 人 = 45 条连接
```

**优点：**
- 延迟最低
- 无服务器压力
- 实现简单

**缺点：**
- 客户端上行带宽消耗大
- 不适合大规模会议
- CPU 占用高（多路编码）

**适用场景：**
- 小规模会议（2-4 人）
- 对延迟要求高的场景

---

### **2. SFU (Selective Forwarding Unit)**

**特点：** 服务器转发，但不解码

```
    A ──┐
        │
    B ──┤
        ├── SFU 服务器
    C ──┤
        │
    D ──┘

每个客户端:
  上行: 1 路流
  下行: n-1 路流
```

**优点：**
- 客户端上行带宽低
- 支持大规模会议（100+ 人）
- 服务器 CPU 占用低

**缺点：**
- 客户端下行带宽消耗大
- 需要部署 SFU 服务器

**适用场景：**
- 中大规模会议（5-100 人）
- 教育、培训场景

**流行的 SFU 实现：**
- Janus
- Mediasoup
- Jitsi Videobridge

---

### **3. MCU (Multipoint Control Unit)**

**特点：** 服务器混流，解码后重新编码

```
    A ──┐
        │
    B ──┤         ┌──► 混合后的单路流
        ├── MCU ──┤
    C ──┤         └──► 所有客户端收到相同的流
        │
    D ──┘

每个客户端:
  上行: 1 路流
  下行: 1 路流
```

**优点：**
- 客户端带宽消耗最低
- 可实现布局控制
- 兼容性好（支持旧设备）

**缺点：**
- 服务器 CPU 占用极高
- 延迟较高
- 成本高

**适用场景：**
- 大规模会议（100+ 人）
- 直播场景
- 弱网环境

---

### **架构对比**

| 特性 | Mesh | SFU | MCU |
|------|------|-----|-----|
| **延迟** | 最低 | 低 | 中等 |
| **客户端上行** | 高 | 低 | 低 |
| **客户端下行** | 高 | 高 | 低 |
| **服务器 CPU** | 无 | 低 | 极高 |
| **服务器带宽** | 无 | 高 | 中等 |
| **适用人数** | 2-4 | 5-100 | 100+ |
| **成本** | 低 | 中 | 高 |

---

## ⚡ 性能优化

### **1. 带宽优化**

```javascript
// 动态调整视频质量
pc.getSenders().forEach(sender => {
    if (sender.track.kind === 'video') {
        const params = sender.getParameters();
        
        // 根据网络状况调整
        if (networkSpeed === 'slow') {
            params.encodings[0].maxBitrate = 500000;  // 500 kbps
            params.encodings[0].scaleResolutionDownBy = 2;
        } else if (networkSpeed === 'fast') {
            params.encodings[0].maxBitrate = 2000000;  // 2 Mbps
            params.encodings[0].scaleResolutionDownBy = 1;
        }
        
        sender.setParameters(params);
    }
});
```

### **2. 模拟编码（Simulcast）**

同时发送多个不同质量的视频流：

```javascript
const pc = new RTCPeerConnection();

const sender = pc.addTrack(videoTrack, stream);

// 配置模拟编码
const params = sender.getParameters();
params.encodings = [
    { rid: 'h', maxBitrate: 1200000, scaleResolutionDownBy: 1.0 },  // 高清
    { rid: 'm', maxBitrate: 600000, scaleResolutionDownBy: 2.0 },   // 标清
    { rid: 'l', maxBitrate: 200000, scaleResolutionDownBy: 4.0 }    // 低清
];

await sender.setParameters(params);
```

### **3. 网络质量监控**

```javascript
// 定时获取统计信息
setInterval(async () => {
    const stats = await pc.getStats();
    
    stats.forEach(report => {
        if (report.type === 'inbound-rtp' && report.kind === 'video') {
            console.log('丢包率:', report.packetsLost / report.packetsReceived);
            console.log('抖动:', report.jitter);
            console.log('接收比特率:', report.bytesReceived * 8 / report.timestamp);
        }
        
        if (report.type === 'outbound-rtp' && report.kind === 'video') {
            console.log('发送比特率:', report.bytesSent * 8 / report.timestamp);
            console.log('编码帧率:', report.framesPerSecond);
        }
    });
}, 5000);
```

### **4. 自动降级策略**

```javascript
// 监听网络质量
pc.oniceconnectionstatechange = () => {
    if (pc.iceConnectionState === 'disconnected') {
        // 降低视频质量
        reduceVideoQuality();
    }
};

// 根据丢包率调整
async function adjustQualityBasedOnPacketLoss() {
    const stats = await pc.getStats();
    
    stats.forEach(report => {
        if (report.type === 'inbound-rtp') {
            const packetLoss = report.packetsLost / report.packetsReceived;
            
            if (packetLoss > 0.05) {  // 丢包率 > 5%
                reduceVideoQuality();
            } else if (packetLoss < 0.01) {  // 丢包率 < 1%
                increaseVideoQuality();
            }
        }
    });
}
```

---

## 🎯 应用场景

### **1. 视频会议**

**特点：**
- 多人音视频
- 屏幕共享
- 文字聊天
- 录制回放

**技术选型：**
- 2-4 人：Mesh
- 5-50 人：SFU
- 50+ 人：MCU 或 SFU + 观众模式

---

### **2. 在线教育**

**特点：**
- 一对多直播
- 互动白板
- 举手发言
- 课程录制

**技术方案：**
- 教师端：高质量推流
- 学生端：接收流 + 偶尔发言
- 架构：SFU 或混合模式

---

### **3. 远程医疗**

**特点：**
- 高清视频
- 隐私保护
- 稳定可靠
- 低延迟

**技术要求：**
- 强制加密
- 专用 TURN 服务器
- 质量监控
- 录像留存

---

### **4. 在线客服**

**特点：**
- 一对一沟通
- 快速接入
- 屏幕共享
- 文件传输

**技术方案：**
- P2P 直连
- DataChannel 传输文件
- 自动排队系统

---

### **5. 云游戏**

**特点：**
- 超低延迟
- 高帧率视频
- 输入实时性
- 大量并发

**技术挑战：**
- 延迟 < 50ms
- 60fps 视频编码
- 带宽优化
- 服务器性能

---

### **6. IoT 视频监控**

**特点：**
- 单向视频流
- 长时间运行
- 低功耗
- 移动端查看

**技术方案：**
- 设备端推流
- 服务器转发
- 移动端拉流
- 录像存储

---

## 📊 性能指标

### **关键性能指标 (KPI)**

| 指标 | 优秀 | 良好 | 可接受 | 差 |
|------|------|------|--------|-----|
| **延迟** | < 150ms | 150-300ms | 300-500ms | > 500ms |
| **丢包率** | < 1% | 1-3% | 3-5% | > 5% |
| **抖动** | < 30ms | 30-50ms | 50-100ms | > 100ms |
| **连接成功率** | > 99% | 95-99% | 90-95% | < 90% |
| **视频帧率** | 30 fps | 24 fps | 15 fps | < 15 fps |
| **音频质量** | MOS > 4.0 | 3.5-4.0 | 3.0-3.5 | < 3.0 |

### **带宽需求**

| 场景 | 上行 | 下行 |
|------|------|------|
| 音频通话 | 50 kbps | 50 kbps |
| 视频通话 (SD) | 1 Mbps | 1 Mbps |
| 视频通话 (HD) | 2.5 Mbps | 2.5 Mbps |
| 视频通话 (Full HD) | 5 Mbps | 5 Mbps |
| 屏幕共享 | 500 kbps | 500 kbps |
| 4 人视频会议 | 2.5 Mbps | 7.5 Mbps (Mesh) |
| 10 人视频会议 | 2.5 Mbps | 25 Mbps (Mesh) |
| 10 人视频会议 (SFU) | 2.5 Mbps | 7.5 Mbps |

---

## 🔍 调试技巧

### **Chrome 内置工具**

**1. chrome://webrtc-internals**
- 查看所有 PeerConnection
- 实时统计信息
- ICE 候选详情
- SDP 内容
- 媒体流信息

**2. 关键指标：**
```
Video:
- packetsSent / packetsReceived
- packetsLost
- framesPerSecond
- frameWidth / frameHeight
- bytesReceived / bytesSent

Audio:
- audioLevel
- jitter
- roundTripTime
```

### **常用调试代码**

```javascript
// 1. 监听所有事件
pc.addEventListener('icecandidate', e => console.log('ICE候选:', e.candidate));
pc.addEventListener('iceconnectionstatechange', e => console.log('ICE连接状态:', pc.iceConnectionState));
pc.addEventListener('icegatheringstatechange', e => console.log('ICE收集状态:', pc.iceGatheringState));
pc.addEventListener('signalingstatechange', e => console.log('信令状态:', pc.signalingState));
pc.addEventListener('connectionstatechange', e => console.log('连接状态:', pc.connectionState));
pc.addEventListener('track', e => console.log('远程轨道:', e.track));
pc.addEventListener('negotiationneeded', e => console.log('需要重新协商'));

// 2. 获取详细统计
async function logStats() {
    const stats = await pc.getStats();
    stats.forEach(report => {
        console.log(report.type, report);
    });
}

// 3. 检查连接质量
async function checkQuality() {
    const stats = await pc.getStats();
    
    stats.forEach(report => {
        if (report.type === 'inbound-rtp' && report.mediaType === 'video') {
            const packetLoss = (report.packetsLost / (report.packetsReceived + report.packetsLost)) * 100;
            console.log(`丢包率: ${packetLoss.toFixed(2)}%`);
            console.log(`抖动: ${report.jitter}ms`);
            console.log(`帧率: ${report.framesPerSecond}fps`);
        }
    });
}
```

---

## 📚 参考资料

### **官方文档**

- [WebRTC 官网](https://webrtc.org/)
- [W3C WebRTC 规范](https://www.w3.org/TR/webrtc/)
- [MDN WebRTC API](https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API)

### **RFC 文档**

- [RFC 8825 - WebRTC Overview](https://tools.ietf.org/html/rfc8825)
- [RFC 8826 - WebRTC Security](https://tools.ietf.org/html/rfc8826)
- [RFC 8827 - WebRTC Data Channels](https://tools.ietf.org/html/rfc8827)
- [RFC 5245 - ICE](https://tools.ietf.org/html/rfc5245)
- [RFC 3264 - SDP Offer/Answer](https://tools.ietf.org/html/rfc3264)
- [RFC 5389 - STUN](https://tools.ietf.org/html/rfc5389)
- [RFC 5766 - TURN](https://tools.ietf.org/html/rfc5766)

### **开源项目**

- [Janus Gateway](https://github.com/meetecho/janus-gateway) - SFU 服务器
- [Mediasoup](https://github.com/versatica/mediasoup) - SFU 库
- [Jitsi](https://github.com/jitsi/jitsi-meet) - 完整视频会议解决方案
- [Pion WebRTC](https://github.com/pion/webrtc) - Go 语言实现

### **学习资源**

- [WebRTC for the Curious](https://webrtcforthecurious.com/) - 深入理解 WebRTC
- [WebRTC Samples](https://webrtc.github.io/samples/) - 示例代码集合
- [Kranky Geek](https://krankygeek.com/) - WebRTC 技术大会

---

## 🎓 总结

### **WebRTC 核心要点**

1. **P2P 架构** - 媒体流直接传输，延迟低
2. **强制加密** - DTLS + SRTP，安全性高
3. **NAT 穿透** - ICE + STUN/TURN，连接成功率高
4. **媒体协商** - SDP Offer/Answer，灵活配置
5. **自适应编码** - 根据网络状况调整质量

### **选择合适的架构**

```
场景决策树:

参与人数?
├─ 2-4 人 → Mesh (P2P)
├─ 5-50 人 → SFU
└─ 50+ 人 → MCU 或 SFU + 观众模式

延迟要求?
├─ 极低 (< 100ms) → Mesh 或 SFU
└─ 可接受 (< 500ms) → MCU

客户端性能?
├─ 高性能设备 → Mesh 或 SFU
└─ 低性能设备 → MCU

成本预算?
├─ 低成本 → Mesh
├─ 中等成本 → SFU
└─ 高成本可接受 → MCU
```

### **最佳实践**

✅ **务必做：**
- 使用 HTTPS
- 实现信令服务器
- 配置 STUN/TURN
- 监控连接质量
- 处理异常断线
- 优化带宽使用

❌ **避免：**
- 依赖特定浏览器特性
- 忽略网络质量监控
- 不处理 ICE 失败情况
- 硬编码媒体参数
- 忽略移动端优化

---

**文档版本:** 1.0  
**最后更新:** 2025-10-18  
**作者:** go-hichat-api team

---

**继续探索 WebRTC 的精彩世界！** 🚀

