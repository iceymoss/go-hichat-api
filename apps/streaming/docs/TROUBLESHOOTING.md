# 故障排查指南

## 🔧 常见问题解决方案

### 问题 1: "Requested device not found" 错误

**错误信息**: `无法获取媒体设备: Requested device not found`

**可能原因**:
1. 没有连接摄像头或麦克风
2. 设备被其他程序占用
3. 权限设置问题
4. 请求的设备参数不支持

**解决方案**:

#### 方案 A: 检查设备连接
```bash
# Windows: 检查设备管理器
Win + X -> 设备管理器 -> 查看"摄像头"和"音频输入"

# Mac: 系统偏好设置
系统偏好设置 -> 安全性与隐私 -> 摄像头/麦克风

# Linux: 
ls /dev/video*  # 查看摄像头设备
arecord -l      # 查看音频设备
```

#### 方案 B: 检查浏览器权限
1. **Chrome**:
   - 地址栏左侧点击锁图标/摄像头图标
   - 检查摄像头和麦克风权限
   - 设置为"允许"

2. **Firefox**:
   - 地址栏左侧点击锁图标
   - 权限 -> 使用摄像头/麦克风 -> 允许

3. **Edge**:
   - 设置 -> Cookie 和网站权限
   - 摄像头/麦克风 -> 管理权限

#### 方案 C: 关闭占用设备的程序
```
常见占用设备的程序：
- Zoom、Teams、Skype
- OBS、XSplit
- 其他视频通话应用
```

#### 方案 D: 使用改进的测试页面
我已经更新了 `test_media_streaming.html`，现在会：
- 🔄 自动降级重试（高清→标清→基础→仅音频）
- 📊 显示详细的设备检测信息
- ⚠️ 提供明确的错误提示

---

### 问题 2: 权限被拒绝

**错误信息**: `NotAllowedError` 或 `PermissionDeniedError`

**解决方案**:

1. **清除网站权限并重试**:
   ```
   Chrome: 设置 -> 隐私和安全 -> 网站设置 -> 重置权限
   ```

2. **检查操作系统权限**:
   - **Windows 10/11**: 设置 -> 隐私 -> 摄像头/麦克风 -> 允许应用访问
   - **Mac**: 系统偏好设置 -> 安全性与隐私 -> 隐私 -> 摄像头/麦克风
   - **Linux**: 检查 `/etc/security/` 配置

3. **使用 localhost 或 HTTPS**:
   ```
   ✅ 正确: http://localhost:8888/test_media_streaming.html
   ✅ 正确: http://127.0.0.1:8888/test_media_streaming.html
   ✅ 正确: https://yourdomain.com/test_media_streaming.html
   ❌ 错误: http://192.168.1.100:8888/... (非安全上下文)
   ```

---

### 问题 3: 无视频设备但有音频

**现象**: 日志显示 "检测到 0 个视频设备"

**解决方案**:

1. **使用仅音频模式**: 
   - 新版测试页面会自动降级到仅音频模式
   - 这是正常的，音频通话仍然可以工作

2. **连接外置摄像头**:
   - USB摄像头
   - 手机作为摄像头（DroidCam、EpocCam等）

3. **虚拟摄像头**:
   - OBS Virtual Camera
   - ManyCam
   - XSplit VCam

---

### 问题 4: WebSocket 连接失败

**错误信息**: `WebSocket connection to 'ws://localhost:8888/ws' failed`

**解决方案**:

1. **确认服务器正在运行**:
   ```bash
   cd apps/streaming
   go run streaming.go
   
   # 应该看到输出:
   # Streaming service started successfully!
   # WebSocket endpoint: ws://0.0.0.0:8888/ws
   ```

2. **检查端口是否被占用**:
   ```bash
   # Windows
   netstat -ano | findstr :8888
   
   # Mac/Linux
   lsof -i :8888
   ```

3. **修改配置文件**:
   ```yaml
   # apps/streaming/etc/streaming-local.yaml
   ListenOn: 0.0.0.0:8888  # 确保是这个地址
   ```

4. **检查防火墙**:
   ```bash
   # Windows: 允许端口 8888
   # Mac: 系统偏好设置 -> 安全性与隐私 -> 防火墙 -> 防火墙选项
   # Linux: sudo ufw allow 8888
   ```

---

### 问题 5: 看不到对方的视频

**现象**: 连接成功但远端视频区域是黑屏

**检查步骤**:

1. **查看日志**:
   ```
   应该看到这些日志：
   ✅ "收到远端媒体轨道: video"
   ✅ "媒体流转发成功"
   ```

2. **检查服务器日志**:
   ```bash
   # 应该看到：
   # 收到远端媒体轨道
   # 本地轨道创建成功，准备转发
   # 开始转发媒体流
   # 媒体流转发成功
   ```

3. **验证 ICE 连接**:
   ```
   页面统计区域应该显示：
   ICE 状态: connected
   连接状态: connected
   ```

4. **检查重新协商**:
   ```
   日志中应该有：
   "发送重新协商Offer"
   "收到Offer，创建Answer"
   ```

**解决方案**:

- 刷新页面重新加入
- 确保两个用户在同一个房间
- 检查网络连接
- 查看浏览器控制台是否有错误

---

### 问题 6: 音视频卡顿或延迟

**解决方案**:

1. **降低视频质量**:
   - 测试页面会自动选择合适的质量
   - 手动修改约束：`{ width: 640, height: 480 }`

2. **检查网络**:
   ```bash
   # 测试延迟
   ping 8.8.8.8
   
   # 测试带宽
   # 使用 speedtest.net
   ```

3. **优化服务器**:
   ```bash
   # 检查CPU使用率
   top
   
   # 检查内存
   free -m
   ```

4. **使用 TURN 服务器**:
   ```yaml
   # apps/streaming/etc/streaming-local.yaml
   WebRTC:
     IceServers:
       - URLs:
           - "stun:stun.l.google.com:19302"
       - URLs:
           - "turn:your-turn-server:3478"
         Username: "username"
         Credential: "password"
   ```

---

## 🔍 诊断工具

### 1. 浏览器诊断页面

**Chrome**:
```
chrome://webrtc-internals/
```
- 查看所有WebRTC连接
- 实时统计信息
- ICE候选详情

**Firefox**:
```
about:webrtc
```

**Edge**:
```
edge://webrtc-internals/
```

### 2. 测试 WebRTC 支持

访问: https://test.webrtc.org/

### 3. 设备测试

```javascript
// 在浏览器控制台运行
navigator.mediaDevices.enumerateDevices()
  .then(devices => {
    console.log('设备列表:', devices);
    devices.forEach(device => {
      console.log(`${device.kind}: ${device.label}`);
    });
  });
```

### 4. 测试媒体获取

```javascript
// 在浏览器控制台运行
navigator.mediaDevices.getUserMedia({ video: true, audio: true })
  .then(stream => {
    console.log('✅ 媒体流获取成功');
    console.log('视频轨道:', stream.getVideoTracks());
    console.log('音频轨道:', stream.getAudioTracks());
    stream.getTracks().forEach(track => track.stop());
  })
  .catch(err => {
    console.error('❌ 错误:', err.name, err.message);
  });
```

---

## 📱 不同环境的注意事项

### Windows

```
常见问题：
1. 摄像头驱动过时 -> 更新驱动
2. 隐私设置限制 -> 检查 Windows 隐私设置
3. 防病毒软件阻止 -> 添加浏览器到白名单
```

### Mac

```
常见问题：
1. 系统权限未授予 -> 系统偏好设置 -> 安全性与隐私
2. Safari 限制严格 -> 建议使用 Chrome
3. M1/M2 芯片兼容性 -> 确保使用最新浏览器
```

### Linux

```
常见问题：
1. 视频设备权限 -> sudo usermod -a -G video $USER
2. PulseAudio/ALSA 配置 -> 检查音频服务
3. 浏览器 Snap 版本限制 -> 使用 deb 或 rpm 版本
```

### 虚拟机

```
常见问题：
1. USB 设备透传 -> 配置 USB 设备共享
2. 性能不足 -> 分配更多CPU/内存
3. 网络配置 -> 使用桥接模式而非NAT
```

---

## 🆘 获取帮助

如果以上方案都无法解决问题：

1. **查看完整日志**:
   - 打开浏览器开发者工具 (F12)
   - 查看 Console 标签页
   - 复制所有错误信息

2. **检查服务器日志**:
   ```bash
   # 运行服务时查看完整输出
   go run streaming.go
   ```

3. **提供诊断信息**:
   ```
   - 操作系统和版本
   - 浏览器和版本
   - 错误截图
   - 完整的错误日志
   - chrome://webrtc-internals/ 的截图
   ```

4. **测试简化场景**:
   - 只测试本地视频（不加入房间）
   - 测试 localhost 而非 IP 地址
   - 使用不同的浏览器

---

## ✅ 快速检查清单

在报告问题前，请确认：

- [ ] 浏览器支持 WebRTC（Chrome、Firefox、Edge）
- [ ] 使用 localhost 或 HTTPS 访问
- [ ] 摄像头和麦克风已连接
- [ ] 已授予浏览器权限
- [ ] 设备未被其他程序占用
- [ ] 服务器正在运行 (localhost:8888)
- [ ] 防火墙允许连接
- [ ] 查看了浏览器控制台错误
- [ ] 查看了 chrome://webrtc-internals/
- [ ] 尝试了降级重试（测试页面自动进行）

---

**文档版本**: v1.0  
**最后更新**: 2025-10-15

