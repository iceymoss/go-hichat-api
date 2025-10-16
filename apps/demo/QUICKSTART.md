# 快速开始指南

## 🚀 启动服务

### 方法 1: 使用启动脚本

```bash
cd /Users/iceymoss/project/go-hichat-api/apps/demo
./start.sh
```

### 方法 2: 直接运行

```bash
cd /Users/iceymoss/project/go-hichat-api/apps/demo
go run demo.go -f etc/demo.yaml
```

## 📱 测试视频通话

1. **启动服务后，你会看到如下输出：**

```
🎥 视频通话服务启动成功!
📡 信令服务器: ws://0.0.0.0:8890/ws
🌐 访问地址: http://0.0.0.0:8890
🔧 状态接口: http://0.0.0.0:8890/status
```

2. **打开两个浏览器窗口（或使用两台设备）：**

   - 窗口 1: 打开 `http://localhost:8890`
   - 窗口 2: 打开 `http://localhost:8890`
   
   或者使用两台设备：
   - 设备 1: `http://你的IP地址:8890`
   - 设备 2: `http://你的IP地址:8890`

3. **在两个窗口中都点击 "开始通话" 按钮**

4. **允许浏览器访问摄像头和麦克风**

5. **系统会自动配对并建立视频连接！**

## 🎮 功能说明

- **开始通话**: 连接到服务器并等待配对
- **挂断**: 结束当前通话
- **静音/取消静音**: 控制麦克风
- **关闭/开启视频**: 控制摄像头

## 🔍 查看服务状态

访问: http://localhost:8890/status

返回示例：
```json
{
  "status": "running",
  "clients": 2,
  "stun_servers": [
    "stun:stun.l.google.com:19302",
    "stun:stun1.l.google.com:19302"
  ]
}
```

## 🐛 故障排查

### 问题: 无法访问摄像头
- 检查浏览器权限设置
- 确保使用 Chrome/Firefox/Safari 等现代浏览器
- 必须使用 HTTPS 或 localhost

### 问题: 无法连接到对方
- 确保两个客户端都已点击"开始通话"
- 检查浏览器控制台是否有错误
- 确认服务器正在运行

### 问题: 视频卡顿
- 检查网络带宽
- 尝试降低视频分辨率（可在代码中修改）
- 关闭其他占用带宽的应用

## 📂 项目文件说明

```
apps/demo/
├── demo.go                      # 服务入口
├── start.sh                     # 启动脚本
├── etc/demo.yaml               # 配置文件
├── internal/                    # 内部代码
│   ├── config/                 # 配置
│   ├── handler/                # HTTP 处理器
│   ├── logic/                  # 业务逻辑（信令服务器）
│   └── svc/                    # 服务上下文
└── static/                      # 前端文件
    ├── index.html              # 主页面
    └── app.js                  # 前端逻辑
```

## 🎯 下一步

- 查看 `README.md` 了解详细文档
- 修改 `etc/demo.yaml` 自定义配置
- 在 `static/app.js` 中调整视频分辨率等参数

祝你使用愉快！🎉

