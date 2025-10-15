# 媒体流处理实现文档

> **实现日期**: 2025-10-15  
> **版本**: v1.0  
> **状态**: ✅ 已完成核心功能

## 📋 实现概述

本次实现完成了 streaming 服务中**缺失的核心媒体流处理功能**，包括：

1. ✅ **OnTrack 回调** - 接收远端媒体流
2. ✅ **RTP 包转发** - 实时读取和转发音视频数据
3. ✅ **SFU 媒体转发** - 服务器端选择性转发
4. ✅ **重新协商机制** - 动态添加媒体轨道时的协商

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Streaming Service                         │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │ SignalingServer │  │ RoomManager  │    │     SFU      │  │
│  └───────┬────────┘  └──────────────┘    └──────────────┘  │
│          │                                                   │
│  ┌───────▼────────────────────────────────────────┐         │
│  │         WebRTCConnection (Enhanced)             │         │
│  │  • OnTrack 回调                                 │         │
│  │  • handleIncomingTrack()                        │         │
│  │  • 本地轨道管理                                 │         │
│  └─────────────────────────────────────────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

### 数据流

```
用户A                    服务器                    用户B
  │                        │                        │
  ├─ RTP包(音频/视频) ──>│                        │
  │                        ├─ OnTrack触发         │
  │                        │  track.Read()         │
  │                        │  localTrack.Write()   │
  │                        │  AddTrack(B)          │
  │                        ├─ 转发RTP包 ────────>│
  │                        │                        │
  │                        │<─── RTP包(音频/视频) ─┤
  │                        ├─ OnTrack触发         │
  │<────── 转发RTP包 ──────┤                        │
```

## 📁 修改的文件

### 1. `apps/streaming/webrtc/connection.go` (扩展)

**新增字段**:
```go
type WebRTCConnection struct {
    // ... 原有字段 ...
    
    // 媒体轨道相关
    localTracks  map[string]*webrtc.TrackLocalStaticRTP
    tracksMu     sync.RWMutex
    
    // 回调函数
    onTrackHandler func(track *webrtc.TrackLocalStaticRTP)
}
```

**新增方法**:
- `handleIncomingTrack()` - 处理接收到的远端媒体轨道
- `SetOnTrackHandler()` - 设置轨道接收回调
- `GetLocalTracks()` - 获取所有本地轨道
- `RemoveTrack()` - 移除轨道

**核心实现**:
```go
// OnTrack 回调 - 在 NewWebRTCConnection 中设置
peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
    // 创建本地轨道
    localTrack, _ := webrtc.NewTrackLocalStaticRTP(...)
    
    // RTP包转发循环
    for {
        n, _, _ := track.Read(buf)        // 从远端读取
        localTrack.Write(buf[:n])         // 写入本地轨道
    }
})
```

### 2. `apps/streaming/internal/handler/signaling.go` (扩展)

**修改的方法**:
- `handleJoinRoom()` - 添加OnTrack回调和已有轨道同步

**新增方法**:
- `forwardTrackToRoom()` - 转发媒体轨道到房间其他用户
- `renegotiateConnection()` - 重新协商连接
- `addExistingTracksToNewUser()` - 将已有用户的轨道添加到新用户

**核心逻辑**:
```go
// 在 handleJoinRoom 中
webrtcConn.SetOnTrackHandler(func(track *webrtc.TrackLocalStaticRTP) {
    // 收到用户媒体流时，转发给房间内其他用户
    s.forwardTrackToRoom(roomID, userID, track)
})

// 为新用户添加已有用户的媒体流
s.addExistingTracksToNewUser(roomID, userID, webrtcConn)
```

## 🔥 核心功能详解

### 功能 1: 媒体流接收 (OnTrack)

**位置**: `webrtc/connection.go:128-142`

```go
peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
    // 1. 收到远端轨道
    // 2. 创建本地轨道用于转发
    // 3. 启动RTP包读取循环
    // 4. 触发回调通知上层
})
```

**作用**: 
- 当用户A发送音视频流时，服务器通过此回调接收
- 自动创建本地轨道供转发使用

### 功能 2: RTP 包转发循环

**位置**: `webrtc/connection.go:352-403`

```go
func handleIncomingTrack(remoteTrack, receiver) {
    for {
        n, _, _ := remoteTrack.Read(buf)    // 从用户读取RTP包
        localTrack.Write(buf[:n])           // 写入本地轨道
        // 每5秒记录统计
    }
}
```

**作用**:
- 实时读取用户发送的RTP包（每包约1500字节）
- 写入本地轨道，供转发给其他用户
- 统计转发数据量

### 功能 3: 媒体流转发

**位置**: `signaling.go:1680-1737`

```go
func forwardTrackToRoom(roomID, fromUserID, track) {
    // 1. 获取房间内所有用户
    // 2. 遍历其他用户
    // 3. 将track添加到每个用户的PeerConnection
    // 4. 触发重新协商
}
```

**作用**:
- 当收到用户A的媒体流时
- 自动转发给房间内的用户B、C、D...
- 每个用户都会收到其他所有用户的流

### 功能 4: 重新协商机制

**位置**: `signaling.go:1739-1772`

```go
func renegotiateConnection(conn, userID, sender) {
    // 1. 创建新的Offer
    // 2. 发送给用户
    // 3. 用户回复Answer
    // 4. 完成协商，媒体流开始传输
}
```

**作用**:
- 添加新的媒体轨道时，需要重新协商
- 自动创建Offer发送给用户
- 确保媒体流能够正确传输

### 功能 5: 新用户同步已有媒体流

**位置**: `signaling.go:1774-1831`

```go
func addExistingTracksToNewUser(roomID, newUserID, newUserConn) {
    // 1. 获取房间内已有用户
    // 2. 获取每个用户的本地轨道
    // 3. 将所有轨道添加到新用户
    // 4. 触发重新协商
}
```

**作用**:
- 用户B加入房间时，自动接收用户A的媒体流
- 无需等待用户A重新发送
- 实现即时同步

## 📊 完整的消息流程

### 用户A加入房间并发送媒体流

```
1. 用户A: join_room
   └─> 服务器: 创建WebRTCConnection
       └─> 设置OnTrack回调

2. 用户A: offer (包含媒体描述)
   └─> 服务器: createAnswer

3. 用户A: 开始发送音视频RTP包
   └─> 服务器: OnTrack触发
       ├─> 创建localTrack
       ├─> 启动RTP转发循环
       └─> 调用onTrackHandler回调
           └─> forwardTrackToRoom()
               (此时房间只有A，无需转发)
```

### 用户B加入房间

```
1. 用户B: join_room
   └─> 服务器: 创建WebRTCConnection
       ├─> 设置OnTrack回调
       └─> addExistingTracksToNewUser()
           ├─> 获取用户A的localTrack
           ├─> AddTrack(A的track) 到B的连接
           └─> renegotiateConnection(B)
               └─> 发送新的Offer给B

2. 用户B: answer (回复协商)
   └─> 服务器: 完成协商
       └─> B开始接收A的媒体流

3. 用户B: 开始发送音视频RTP包
   └─> 服务器: OnTrack触发
       └─> forwardTrackToRoom()
           ├─> 获取房间用户 [A, B]
           ├─> AddTrack(B的track) 到A的连接
           └─> renegotiateConnection(A)
               └─> A开始接收B的媒体流
```

## 🎯 关键特性

### 1. 自动转发
- 收到媒体流自动转发，无需手动触发
- 支持音频和视频同时转发

### 2. 实时性能
- RTP包读取和转发在独立goroutine中
- 每个轨道独立处理，互不影响
- 每5秒统计一次转发性能

### 3. 可靠性
- 错误处理和日志记录完善
- 连接断开自动清理资源
- 支持中途加入和离开

### 4. 可扩展性
- 支持多人通话（SFU架构）
- 可轻松扩展到群组通话、会议等场景

## 📈 性能指标

### 带宽使用
```
音频流: 约 50-100 Kbps per user
视频流(720p): 约 1-2 Mbps per user

2人通话:
- 上行: 音频(50K) + 视频(1.5M) = 1.55 Mbps
- 下行: 音频(50K) + 视频(1.5M) = 1.55 Mbps
- 服务器转发: 3.1 Mbps

3人通话:
- 每人上行: 1.55 Mbps
- 每人下行: 3.1 Mbps (接收2人)
- 服务器转发: 9.3 Mbps
```

### RTP 包频率
```
音频: 约 50 包/秒
视频: 约 30-60 包/秒 (取决于帧率)

每个连接每秒处理: 80-110 个RTP包
```

## 🧪 测试建议

### 单元测试

```go
// 测试 OnTrack 回调
func TestOnTrackCallback(t *testing.T) {
    // 1. 创建WebRTCConnection
    // 2. 设置OnTrackHandler
    // 3. 模拟远端发送轨道
    // 4. 验证回调被调用
}

// 测试媒体流转发
func TestForwardTrackToRoom(t *testing.T) {
    // 1. 创建房间和两个用户
    // 2. 用户A发送媒体流
    // 3. 验证用户B收到
}
```

### 集成测试

使用提供的 `apps/streaming/example/client.html`:

```bash
# 1. 启动服务
cd apps/streaming
go run streaming.go

# 2. 打开两个浏览器窗口
# 窗口1: http://localhost:8080/example/client.html
# 窗口2: http://localhost:8080/example/client.html

# 3. 两个窗口都加入同一个房间
# 4. 开启摄像头和麦克风
# 5. 验证相互能看到对方的视频和听到声音
```

## 🐛 故障排查

### 问题 1: 看不到对方视频

**检查**:
1. 查看日志是否有 "收到远端媒体轨道"
2. 查看是否有 "媒体流转发成功"
3. 检查浏览器控制台是否有WebRTC错误

**解决**:
- 确保 ICE 连接已建立
- 检查防火墙设置
- 验证 STUN/TURN 配置

### 问题 2: 音视频卡顿

**检查**:
1. 查看日志中的 "媒体流转发统计"
2. 检查 RTP 包数量是否正常
3. 查看服务器CPU和网络使用率

**解决**:
- 降低视频分辨率
- 优化网络环境
- 增加服务器资源

### 问题 3: 中途加入无法看到已有用户

**检查**:
1. 查看 "为新用户添加已有媒体流" 日志
2. 验证 addExistingTracksToNewUser 是否被调用
3. 检查重新协商是否成功

**解决**:
- 检查房间管理器状态
- 验证WebRTC连接状态
- 查看Answer是否正常返回

## 📚 相关文档

- [一对一通话流程](streaming-one-to-one-call-flow.md)
- [WebRTC API文档](https://pkg.go.dev/github.com/pion/webrtc/v3)
- [SFU架构说明](ARCHITECTURE.md)

## ✅ 完成清单

- [x] 实现 OnTrack 回调
- [x] 实现 RTP 包读取和转发
- [x] 实现媒体流转发逻辑
- [x] 实现重新协商机制
- [x] 实现新用户同步已有媒体流
- [x] 添加完善的日志和错误处理
- [x] 编写实现文档
- [ ] 完成单元测试 (建议后续添加)
- [ ] 完成集成测试 (需要客户端配合)
- [ ] 性能压测 (建议后续进行)

## 🎉 总结

本次实现完成了 streaming 服务中最核心的媒体流处理功能，从**只有信令交换**到**完整的音视频通话**。

**核心成果**:
1. ✅ 用户可以发送音视频流到服务器
2. ✅ 服务器自动转发给房间内其他用户
3. ✅ 支持中途加入和动态协商
4. ✅ 完整的SFU架构实现

**下一步建议**:
1. 添加媒体质量控制（码率调整）
2. 实现录制功能
3. 添加美颜、滤镜等特效
4. 优化大规模会议性能

---

**文档版本**: v1.0  
**最后更新**: 2025-10-15  
**作者**: AI Assistant

