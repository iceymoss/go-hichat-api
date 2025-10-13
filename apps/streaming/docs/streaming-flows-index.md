# Streaming 流媒体服务流程文档索引

> **最后更新**: 2025-10-13  
> **服务**: streaming 流媒体服务

## 📚 文档概览

本目录包含 streaming 服务所有功能的完整流程文档，每个文档都详细说明了消息流、数据结构和使用场景。

---

## 📖 流程文档列表

### 1. [一对一通话流程](./streaming-one-to-one-call-flow.md)

**功能**: 两人之间的音视频通话  
**架构**: P2P 直连  
**适用场景**: 
- 好友视频通话
- 远程一对一面试
- 客服视频服务

**核心特点**:
- ✅ 最低延迟
- ✅ 最佳音视频质量
- ✅ 带宽消耗最小

**消息流程**: 约 36-65 条消息

---

### 2. [群组通话流程](streaming-group-call-flow.md)

**功能**: 3-50人的多人音视频通话  
**架构**: SFU 媒体转发  
**适用场景**:
- 家庭视频聚会
- 小团队会议
- 朋友群聊

**核心特点**:
- ✅ 支持多人同时通话
- ✅ SFU 优化带宽
- ✅ 主持人管理功能

**消息流程**: 约 58-86 条消息（4人）

---

### 3. [会议功能流程](streaming-meeting-flow.md)

**功能**: 企业级视频会议解决方案  
**架构**: SFU + 完整会议管理  
**适用场景**:
- 企业视频会议
- 在线培训/网课
- 远程面试
- 商务洽谈

**核心特点**:
- ✅ 会议预约机制
- ✅ 完整的权限控制
- ✅ 等候室功能
- ✅ 会议录制
- ✅ 主持人控制所有人

**消息流程**: 约 138-216 条消息（10人）

---

### 4. [录屏分享流程](streaming-screen-share-flow.md)

**功能**: 在通话/会议中共享屏幕  
**架构**: WebRTC 屏幕流传输  
**适用场景**:
- 远程演示 PPT
- 代码协作/Code Review
- 远程技术支持
- 在线培训/教学

**核心特点**:
- ✅ 实时屏幕内容共享
- ✅ 质量可调（high/medium/low）
- ✅ 暂停/恢复功能
- ✅ 请求共享机制

**消息流程**: 约 4-8 条消息（5人房间）

---

### 5. [直播功能流程](streaming-live-stream-flow.md)

**功能**: 一对多的实时音视频广播  
**架构**: 推流 + CDN 分发  
**适用场景**:
- 电商带货直播
- 在线教育直播
- 游戏直播
- 大型活动直播

**核心特点**:
- ✅ 支持1000+并发
- ✅ CDN 全球加速
- ✅ 多码率自适应
- ✅ 弹幕点赞互动
- ✅ 直播数据统计

**消息流程**: 约 400-500 条消息（100人观众）

---

## 🔄 功能对比

### 参与人数

| 功能 | 人数限制 | 架构方式 | 延迟 |
|-----|---------|---------|------|
| 一对一通话 | 固定 2 人 | P2P 直连 | 最低 |
| 群组通话 | 3-50 人 | SFU 转发 | 低 |
| 会议功能 | 3-50 人 | SFU + 管理 | 低 |
| 录屏分享 | 不限（房间内） | SFU 转发 | 低 |
| 直播功能 | 1000+ 观众 | 推流 + CDN | 中 |

---

### 使用场景

| 功能 | 商务 | 教育 | 社交 | 娱乐 | 电商 |
|-----|-----|------|------|------|------|
| 一对一通话 | ✅ | ✅ | ✅ | ✅ | ❌ |
| 群组通话 | ✅ | ✅ | ✅ | ✅ | ❌ |
| 会议功能 | ✅✅ | ✅✅ | ❌ | ❌ | ❌ |
| 录屏分享 | ✅✅ | ✅✅ | ✅ | ❌ | ❌ |
| 直播功能 | ✅ | ✅✅ | ✅ | ✅✅ | ✅✅ |

✅✅ = 非常适合  
✅ = 适合  
❌ = 不适合

---

### 功能特性

| 功能 | 屏幕共享 | 会议控制 | 录制 | 互动 | CDN加速 |
|-----|---------|---------|------|------|---------|
| 一对一通话 | ❌ | ❌ | ❌ | ❌ | ❌ |
| 群组通话 | ❌ | 基础 | ❌ | ❌ | ❌ |
| 会议功能 | ✅ | 完整 | ✅ | ✅ | ❌ |
| 录屏分享 | ✅✅ | - | - | - | ❌ |
| 直播功能 | ❌ | 基础 | ✅ | ✅✅ | ✅✅ |

---

## 📋 消息类型总览

### 通话相关

| 消息类型 | 一对一 | 群组 | 会议 | 说明 |
|---------|-------|------|------|------|
| `call_invite` | ✅ | ❌ | ❌ | 一对一邀请 |
| `call_accept` | ✅ | ❌ | ❌ | 接受通话 |
| `call_reject` | ✅ | ❌ | ❌ | 拒绝通话 |
| `call_end` | ✅ | ✅ | ❌ | 结束通话 |
| `group_invite` | ❌ | ✅ | ❌ | 群组邀请 |
| `group_join` | ❌ | ✅ | ❌ | 加入群组 |
| `group_leave` | ❌ | ✅ | ❌ | 离开群组 |

---

### 会议相关

| 消息类型 | 说明 | 权限 |
|---------|------|------|
| `meeting_create` | 创建会议 | 任何人 |
| `meeting_join` | 加入会议 | 参会者 |
| `meeting_leave` | 离开会议 | 参会者 |
| `meeting_control` | 会议控制 | 主持人 |

**控制动作**: 
- `mute_participant` - 静音参与者
- `unmute_participant` - 取消静音
- `remove_participant` - 移除参与者
- `end_meeting` - 结束会议

---

### 录屏相关

| 消息类型 | 说明 | 发起方 |
|---------|------|--------|
| `screen_share_start` | 开始共享 | 分享者 |
| `screen_share_stop` | 停止共享 | 分享者 |
| `screen_share_pause` | 暂停共享 | 分享者 |
| `screen_share_resume` | 恢复共享 | 分享者 |
| `screen_share_request` | 请求共享 | 观看者 |

---

### 直播相关

| 消息类型 | 说明 | 发起方 |
|---------|------|--------|
| `live_start` | 开始直播 | 主播 |
| `live_stop` | 停止直播 | 主播 |
| `live_join` | 加入直播间 | 观众 |
| `live_leave` | 离开直播间 | 观众 |
| `live_comment` | 发送弹幕 | 观众 |
| `live_like` | 点赞 | 观众 |
| `live_pause` | 暂停直播 | 主播 |
| `live_resume` | 恢复直播 | 主播 |

---

### 基础房间相关

| 消息类型 | 说明 | 所有功能 |
|---------|------|---------|
| `join_room` | 加入房间 | ✅ |
| `leave_room` | 离开房间 | ✅ |
| `room_info` | 房间信息 | ✅ |
| `user_joined` | 用户加入通知 | ✅ |
| `user_left` | 用户离开通知 | ✅ |

---

### WebRTC 信令

| 消息类型 | 说明 | 所有功能 |
|---------|------|---------|
| `offer` | SDP Offer | ✅ |
| `answer` | SDP Answer | ✅ |
| `ice_candidate` | ICE 候选 | ✅ |

---

### 媒体控制

| 消息类型 | 说明 | 所有功能 |
|---------|------|---------|
| `mute` | 静音 | ✅ |
| `unmute` | 取消静音 | ✅ |
| `video_on` | 开启视频 | ✅ |
| `video_off` | 关闭视频 | ✅ |

---

## 🎯 快速导航

### 按使用场景

**企业办公**:
- [会议功能](streaming-meeting-flow.md) - 正式会议
- [录屏分享](streaming-screen-share-flow.md) - 演示和培训

**社交通讯**:
- [一对一通话](./streaming-one-to-one-call-flow.md) - 私密通话
- [群组通话](streaming-group-call-flow.md) - 朋友聚会

**内容分发**:
- [直播功能](streaming-live-stream-flow.md) - 大规模直播

---

### 按技术架构

**P2P 直连**:
- [一对一通话](./streaming-one-to-one-call-flow.md)

**SFU 转发**:
- [群组通话](streaming-group-call-flow.md)
- [会议功能](streaming-meeting-flow.md)
- [录屏分享](streaming-screen-share-flow.md)

**推流 + CDN**:
- [直播功能](streaming-live-stream-flow.md)

---

## 🔧 技术要点

### WebRTC 基础

所有实时音视频功能都基于 WebRTC 技术：

1. **信令协商** (Signaling)
   - Offer/Answer 交换
   - ICE 候选交换
   - 媒体能力协商

2. **媒体传输** (Media Transport)
   - 音频: Opus 编码
   - 视频: H.264/VP8/VP9
   - 加密: DTLS/SRTP

3. **网络穿透** (NAT Traversal)
   - STUN: 获取公网地址
   - TURN: 中继服务器
   - ICE: 最佳路径选择

---

### SFU 架构

群组通话和会议使用 SFU（选择性转发单元）：

**优势**:
- 每个用户只上传一次流
- 服务器负责转发
- 支持更多参与者

**工作原理**:
```
用户A → SFU → 用户B, C, D
用户B → SFU → 用户A, C, D
用户C → SFU → 用户A, B, D
```

---

### CDN 分发

直播使用 CDN 实现大规模分发：

**流程**:
```
主播 → 流媒体服务器 → CDN → 观众
      (推流)      (转码)  (分发)
```

**优势**:
- 支持百万级并发
- 全球加速
- 降低延迟

---

## 📚 相关资源

### 源码位置

- **Handler**: `apps/streaming/internal/handler/signaling.go`
- **Call Manager**: `apps/streaming/internal/logic/call_manager.go`
- **Meeting Manager**: `apps/streaming/internal/logic/meeting_manager.go`
- **Screen Share Manager**: `apps/streaming/internal/logic/screen_share_manager.go`
- **Live Stream Manager**: `apps/streaming/internal/logic/live_stream_manager.go`
- **Room Manager**: `apps/streaming/room/manager.go`
- **SFU**: `apps/streaming/sfu/sfu.go`
- **WebRTC**: `apps/streaming/webrtc/connection.go`

---

### 配置文件

- `apps/streaming/etc/streaming-local.yaml` - 本地开发配置

---

### 测试文件

- `apps/streaming/example/client.html` - 前端示例
- `apps/streaming/quick_test.html` - 快速测试
- `apps/streaming/comprehensive_test.html` - 完整测试

---

## 💡 最佳实践

### 1. 选择合适的功能

**人数规模**:
- 2人: 使用一对一通话（最佳质量）
- 3-10人: 使用群组通话
- 10-50人: 使用会议功能
- 50+人: 使用直播功能

**场景需求**:
- 需要会议管理: 使用会议功能
- 只需基础通话: 使用群组通话
- 需要大规模观看: 使用直播功能

---

### 2. 网络优化

**带宽需求**:
- 一对一: 上行1-2Mbps, 下行1-2Mbps
- 群组(5人): 上行1-2Mbps, 下行4-8Mbps
- 直播观看: 下行1-3Mbps

**优化建议**:
- 使用自适应码率
- 启用网络质量检测
- 提供降级方案

---

### 3. 用户体验

**关键指标**:
- 延迟: < 500ms (实时通话)
- 丢包率: < 2%
- 音视频同步: < 100ms

**体验优化**:
- 显示网络状态
- 提供重连机制
- 优雅处理异常

---

## 🤝 贡献指南

如发现文档问题或需要补充，请：

1. 提交 Issue 说明问题
2. 或直接提交 PR 修改文档
3. 保持文档格式一致

---

## 📮 反馈

如有任何问题或建议，欢迎反馈！

---

**文档索引结束**

返回 [项目主页](../../../README.md)

