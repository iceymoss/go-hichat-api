# 一对一视频通话 Demo

这是一个基于 **WebRTC** 和 **Pion** 库实现的简单一对一视频通话应用。

## ✨ 功能特性

- 🎥 **实时视频通话** - 一对一视频通话，支持高清画质
- 🎤 **音频控制** - 支持静音/取消静音
- 📹 **视频控制** - 支持开启/关闭摄像头
- 🔄 **自动配对** - 系统自动为用户匹配对方
- 💬 **实时消息** - 显示连接状态和操作日志
- 🌐 **跨平台** - 支持桌面和移动端浏览器

## 🏗️ 技术架构

### 后端
- **Go** - 主要编程语言
- **Pion WebRTC** - WebRTC 实现库
- **Go-Zero** - Web 框架
- **Gorilla WebSocket** - WebSocket 支持

### 前端
- **原生 JavaScript** - 无需任何框架
- **WebRTC API** - 浏览器原生 API
- **WebSocket** - 信令通信

## 📦 安装依赖

确保已安装以下依赖：

```bash
# 检查 Go 版本 (需要 Go 1.19+)
go version

# 安装项目依赖
cd apps/demo
go mod tidy
```

## 🚀 快速开始

### 方式一：使用启动脚本

```bash
cd apps/demo
./start.sh
```

### 方式二：直接运行

```bash
cd apps/demo
go run demo.go -f etc/demo.yaml
```

### 访问应用

启动成功后，打开浏览器访问：

```
http://localhost:8890
```

## 🎮 使用说明

1. **打开两个浏览器窗口**（或使用两台设备）
2. 在两个窗口中都访问 `http://localhost:8890`
3. 在两个窗口中点击 **"开始通话"** 按钮
4. 允许浏览器访问摄像头和麦克风
5. 系统会自动配对两个用户，建立视频连接

### 控制按钮

- **开始通话** - 连接到服务器并开始配对
- **挂断** - 结束当前通话
- **静音/取消静音** - 控制本地麦克风
- **关闭/开启视频** - 控制本地摄像头

## 📁 项目结构

```
apps/demo/
├── demo.go                    # 主程序入口
├── start.sh                   # 启动脚本
├── README.md                  # 项目文档
├── etc/
│   └── demo.yaml             # 配置文件
├── internal/
│   ├── config/
│   │   └── config.go         # 配置结构
│   ├── handler/
│   │   ├── routes.go         # 路由配置
│   │   └── websocket_handler.go  # WebSocket 处理器
│   ├── logic/
│   │   └── signaling.go      # 信令服务器逻辑
│   └── svc/
│       └── servicecontext.go # 服务上下文
└── static/
    ├── index.html            # 前端页面
    └── app.js                # 前端 JavaScript
```

## ⚙️ 配置说明

编辑 `etc/demo.yaml` 文件：

```yaml
Name: demo-api
Host: 0.0.0.0        # 监听地址
Port: 8890           # 监听端口

Log:
  ServiceName: demo-api
  Mode: console
  Level: info

# STUN 服务器配置
StunServers:
  - stun:stun.l.google.com:19302
  - stun:stun1.l.google.com:19302
```

## 🔧 API 接口

### WebSocket 信令接口

**端点**: `ws://localhost:8890/ws`

**消息格式**:

```javascript
// Offer/Answer
{
  "type": "offer" | "answer",
  "sdp": "...",
  "fromId": "client-id",
  "toId": "peer-id"
}

// ICE Candidate
{
  "type": "candidate",
  "candidate": "{...}",
  "fromId": "client-id",
  "toId": "peer-id"
}
```

### REST API

**状态接口**: `GET http://localhost:8890/status`

响应：
```json
{
  "status": "running",
  "clients": 2,
  "stun_servers": ["stun:stun.l.google.com:19302"]
}
```

## 🌍 部署说明

### 本地网络测试

如果要在局域网内测试，修改配置：

```yaml
Host: 0.0.0.0  # 监听所有网卡
Port: 8890
```

然后使用本机 IP 地址访问，例如：
```
http://192.168.1.100:8890
```

### 生产环境部署

生产环境建议：

1. **使用 HTTPS** - WebRTC 需要安全上下文
2. **配置 TURN 服务器** - 处理 NAT 穿透问题
3. **负载均衡** - 使用 Nginx 等反向代理
4. **监控和日志** - 记录连接状态和错误

## 🐛 常见问题

### Q: 无法访问摄像头/麦克风？
A: 确保：
- 浏览器已授予摄像头和麦克风权限
- 使用 HTTPS 或 localhost（Chrome 要求安全上下文）
- 没有其他应用占用摄像头

### Q: 无法连接到对方？
A: 检查：
- 两个客户端都已成功连接到信令服务器
- 防火墙设置
- STUN 服务器是否可用

### Q: 视频卡顿或延迟？
A: 可能原因：
- 网络带宽不足
- CPU 占用过高
- 可以尝试降低视频分辨率

## 📚 扩展功能

可以基于此 Demo 扩展的功能：

- [ ] 多人视频会议
- [ ] 屏幕共享
- [ ] 文字聊天
- [ ] 录制功能
- [ ] 美颜滤镜
- [ ] 虚拟背景
- [ ] 房间管理
- [ ] 用户认证

## 📖 参考资料

- [Pion WebRTC](https://github.com/pion/webrtc)
- [WebRTC API (MDN)](https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API)
- [Go-Zero](https://go-zero.dev/)

## 📝 License

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**享受视频通话吧！** 🎉

