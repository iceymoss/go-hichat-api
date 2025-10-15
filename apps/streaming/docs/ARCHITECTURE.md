# 流媒体服务架构说明

## 通信架构

### 1. 信令层 (WebSocket)
- **协议**: WebSocket over TCP
- **用途**: 连接协商、房间管理、用户管理
- **端口**: 10093
- **消息类型**: JSON格式的信令消息

### 2. 媒体层 (WebRTC)
- **协议**: WebRTC over UDP
- **用途**: 实际音视频数据传输
- **特点**: P2P直连，低延迟

## 与主流产品对比

| 产品 | 信令协议 | 媒体协议 | 优势 |
|------|----------|----------|------|
| 微信/QQ | 私有协议 | 私有RTP/WebRTC | 优化控制、低延迟 |
| Zoom | WebSocket + 私有 | WebRTC/SFU | 稳定性好 |
| 我们的服务 | WebSocket | WebRTC | 标准协议、易开发 |

## 技术栈

- **后端**: Go + WebSocket + WebRTC (Pion)
- **前端**: WebRTC API + WebSocket
- **架构**: SFU (选择性转发单元)

## 核心组件

1. **SignalingServer**: WebSocket信令服务器
2. **RoomManager**: 房间管理
3. **SFU**: 媒体流转发
4. **CallManager**: 通话管理
5. **MeetingManager**: 会议管理

## 消息流程

```
客户端 --WebSocket--> 信令服务器 --业务逻辑--> 房间/通话管理
   |                                                    |
   --WebRTC P2P----------------------------------------|
```

## 扩展性

- 支持1对1通话
- 支持群组通话
- 支持会议功能
- 支持录屏分享
- 支持直播功能
