# 流媒体服务 (Streaming Service)

这是一个基于 go-zero 框架和 WebRTC 技术构建的实时音视频通话微服务，支持多人视频会议和语音通话。

## 功能特性

- 🎥 **实时视频通话**: 基于 WebRTC 的高质量视频传输
- 🎤 **实时语音通话**: 低延迟音频传输
- 👥 **多人会议**: 支持多人同时参与的视频会议
- 🏠 **房间管理**: 灵活的房间创建和管理
- 📡 **SFU架构**: 选择性转发单元，优化带宽使用
- 🔐 **安全连接**: 支持 STUN/TURN 服务器
- 📊 **实时统计**: 连接状态和带宽监控
- 🔄 **自动清理**: 自动清理过期房间和连接

## 技术栈

- **框架**: go-zero
- **WebRTC**: pion/webrtc
- **WebSocket**: gorilla/websocket
- **信令服务器**: 自定义实现
- **SFU**: 选择性转发单元
- **房间管理**: 内存管理 + Redis 持久化

## 目录结构

```
apps/streaming/
├── etc/                           # 配置文件
│   └── streaming-local.yaml
├── internal/                      # 内部实现
│   ├── config/                   # 配置结构
│   ├── handler/                  # 处理器
│   │   └── signaling.go         # 信令服务器
│   ├── svc/                     # 服务上下文
│   └── types/                   # 类型定义
├── webrtc/                      # WebRTC 相关
│   └── connection.go            # WebRTC 连接管理
├── room/                        # 房间管理
│   ├── room.go                  # 房间实现
│   └── manager.go               # 房间管理器
├── sfu/                         # SFU 实现
│   └── sfu.go                   # 选择性转发单元
├── streaming.go                 # 主入口文件
└── README.md                    # 说明文档
```

## 快速开始

### 1. 安装依赖

确保项目中已包含以下依赖：

```go
// go.mod
require (
    github.com/pion/webrtc/v3 v3.2.40
    github.com/gorilla/websocket v1.5.0
    github.com/zeromicro/go-zero v1.8.2
)
```

### 2. 配置服务

编辑 `etc/streaming-local.yaml` 配置文件：

```yaml
Name: streaming.service
ListenOn: 0.0.0.0:10093

# WebRTC 配置
WebRTC:
  IceServers:
    - URLs:
        - "stun:stun.l.google.com:19302"
        - "stun:stun1.l.google.com:19302"
    - URLs:
        - "turn:your-turn-server.com:3478"
      Username: "username"
      Credential: "password"

# SFU 配置
SFU:
  MaxRooms: 1000
  MaxUsersPerRoom: 50
  RoomTimeout: 3600
  UserTimeout: 300
```

### 3. 启动服务

```bash
# 启动流媒体服务
go run apps/streaming/streaming.go -f apps/streaming/etc/streaming-local.yaml
```

### 4. 连接WebSocket

```javascript
// 前端连接示例
const ws = new WebSocket('ws://localhost:10093/ws');

ws.onopen = function() {
    console.log('Connected to streaming service');
    
    // 加入房间
    ws.send(JSON.stringify({
        type: 'join_room',
        room_id: 'room_123',
        user_id: 'user_456',
        timestamp: new Date().toISOString()
    }));
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('Received message:', message);
};
```

## API 接口

### WebSocket 信令协议

#### 1. 加入房间

```json
{
    "type": "join_room",
    "room_id": "room_123",
    "user_id": "user_456",
    "timestamp": "2025-01-01T00:00:00Z"
}
```

#### 2. 离开房间

```json
{
    "type": "leave_room",
    "room_id": "room_123",
    "user_id": "user_456",
    "timestamp": "2025-01-01T00:00:00Z"
}
```

#### 3. 发送 Offer

```json
{
    "type": "offer",
    "room_id": "room_123",
    "user_id": "user_456",
    "data": {
        "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n...",
        "type": "offer"
    },
    "timestamp": "2025-01-01T00:00:00Z"
}
```

#### 4. 发送 Answer

```json
{
    "type": "answer",
    "room_id": "room_123",
    "user_id": "user_456",
    "data": {
        "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n...",
        "type": "answer"
    },
    "timestamp": "2025-01-01T00:00:00Z"
}
```

#### 5. 发送 ICE 候选

```json
{
    "type": "ice_candidate",
    "room_id": "room_123",
    "user_id": "user_456",
    "data": {
        "candidate": "candidate:1 1 UDP 2113667326 192.168.1.100 54400 typ host",
        "sdpMLineIndex": 0,
        "sdpMid": "0"
    },
    "timestamp": "2025-01-01T00:00:00Z"
}
```

## 核心组件

### 1. 信令服务器 (SignalingServer)

负责处理 WebRTC 连接建立过程中的信令交换：

- 处理 WebSocket 连接
- 管理房间和用户
- 交换 SDP 和 ICE 候选
- 协调媒体流传输

### 2. SFU (Selective Forwarding Unit)

选择性转发单元，优化多人通话的带宽使用：

- 接收每个用户的媒体流
- 选择性转发给其他用户
- 支持音频混合和视频转发
- 动态调整传输质量

### 3. 房间管理器 (RoomManager)

管理房间和用户的生命周期：

- 创建和删除房间
- 用户加入和离开
- 房间状态监控
- 自动清理过期房间

### 4. WebRTC 连接管理

管理每个用户的 WebRTC 连接：

- 创建和管理 PeerConnection
- 处理媒体轨道
- 监控连接状态
- 错误恢复和重连

## 配置说明

### WebRTC 配置

```yaml
WebRTC:
  IceServers:          # ICE 服务器配置
    - URLs:            # STUN/TURN 服务器地址
      - "stun:stun.l.google.com:19302"
    - URLs:
      - "turn:your-turn-server.com:3478"
      Username: "username"      # TURN 服务器用户名
      Credential: "password"    # TURN 服务器密码
  
  Media:               # 媒体配置
    Video:
      Width: 1280      # 视频宽度
      Height: 720      # 视频高度
      FrameRate: 30    # 帧率
      Bitrate: 1000000 # 比特率 (bps)
    
    Audio:
      SampleRate: 48000 # 采样率
      Channels: 2       # 声道数
      Bitrate: 128000   # 比特率 (bps)
```

### SFU 配置

```yaml
SFU:
  MaxRooms: 1000           # 最大房间数
  MaxUsersPerRoom: 50      # 每个房间最大用户数
  RoomTimeout: 3600        # 房间超时时间(秒)
  UserTimeout: 300         # 用户连接超时时间(秒)
```

### 信令配置

```yaml
Signaling:
  WebSocket:
    ReadBufferSize: 4096   # 读缓冲区大小
    WriteBufferSize: 4096  # 写缓冲区大小
    CheckOrigin: true      # 检查 Origin 头
  
  MessageQueue:
    BufferSize: 1000       # 消息队列缓冲区大小
    WorkerCount: 10        # 消息处理工作协程数
```

## 部署建议

### 1. 服务器要求

- **CPU**: 2核心以上
- **内存**: 4GB以上
- **网络**: 高带宽，低延迟
- **操作系统**: Linux/Windows/macOS

### 2. 网络配置

- 配置 STUN 服务器用于 NAT 穿透
- 配置 TURN 服务器用于中继传输
- 开放必要的端口 (WebSocket, STUN/TURN)
- 配置防火墙规则

### 3. 监控和日志

- 监控连接数量和带宽使用
- 记录错误和性能指标
- 设置告警和自动恢复
- 定期清理过期数据

## 扩展功能

### 1. 录制功能

```go
// 实现媒体录制
type MediaRecorder interface {
    StartRecording(roomID string) error
    StopRecording(roomID string) error
    GetRecording(roomID string) ([]byte, error)
}
```

### 2. 屏幕共享

```go
// 支持屏幕共享
type ScreenShare interface {
    StartScreenShare(userID string) error
    StopScreenShare(userID string) error
    IsScreenSharing(userID string) bool
}
```

### 3. 虚拟背景

```go
// 虚拟背景处理
type VirtualBackground interface {
    SetBackground(userID string, background string) error
    RemoveBackground(userID string) error
}
```

### 4. 美颜滤镜

```go
// 美颜和滤镜
type BeautyFilter interface {
    ApplyFilter(userID string, filter string) error
    RemoveFilter(userID string) error
}
```

## 性能优化

### 1. 带宽优化

- 使用 SFU 架构减少带宽消耗
- 动态调整视频质量
- 音频优先传输策略
- 网络自适应编码

### 2. 延迟优化

- 选择就近的 TURN 服务器
- 优化信令传输
- 减少媒体处理延迟
- 使用硬件加速

### 3. 扩展性优化

- 水平扩展支持
- 负载均衡
- 分布式房间管理
- 缓存优化

## 故障排除

### 1. 连接问题

- 检查 STUN/TURN 服务器配置
- 验证网络连接
- 查看防火墙设置
- 检查 WebSocket 连接

### 2. 音视频问题

- 检查媒体设备权限
- 验证编解码器支持
- 查看带宽限制
- 检查音频/视频轨道

### 3. 性能问题

- 监控 CPU 和内存使用
- 检查网络延迟和丢包
- 优化媒体处理
- 调整并发连接数

## 安全考虑

### 1. 认证授权

- 实现用户认证
- 房间访问控制
- 权限管理
- 会话管理

### 2. 数据安全

- 加密媒体传输
- 保护用户隐私
- 防止数据泄露
- 安全日志记录

### 3. 网络安全

- 防止 DDoS 攻击
- 限制连接频率
- 验证消息格式
- 监控异常行为

## 总结

这个流媒体服务提供了一个完整的实时音视频通话解决方案，具有以下优势：

- **高性能**: 基于 WebRTC 的低延迟传输
- **可扩展**: SFU 架构支持大规模并发
- **易集成**: 标准 WebSocket 接口
- **功能丰富**: 支持多种音视频功能
- **稳定可靠**: 完善的错误处理和恢复机制

通过合理的配置和部署，可以满足各种实时通信场景的需求。
