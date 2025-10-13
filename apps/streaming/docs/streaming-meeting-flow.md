# 会议功能完整流程文档

> **文档版本**: v1.0  
> **更新日期**: 2025-10-13  
> **适用服务**: streaming 流媒体服务

## 📋 目录

- [概述](#概述)
- [完整时序图](#完整时序图)
- [阶段详解](#阶段详解)
- [核心数据结构](#核心数据结构)
- [会议控制功能](#会议控制功能)
- [与群组通话的区别](#与群组通话的区别)
- [关键要点](#关键要点)

---

## 概述

会议功能是**企业级的多人音视频通话解决方案**，提供完整的会议管理、参与者控制、权限管理等功能。

### 核心特点

- **预约机制**: 支持提前创建会议
- **会议管理**: 完整的会议生命周期管理
- **权限控制**: 主持人可控制所有参与者
- **高级功能**: 等候室、录制、屏幕共享等
- **大规模支持**: 最多50人同时参会

### 使用场景

- 企业视频会议
- 在线培训/网课
- 远程面试
- 商务洽谈
- 团队协作

---

## 完整时序图

```
┌─────────┐        ┌────────────┐        ┌─────────┐  ┌─────────┐  ┌─────────┐
│ 主持人H │        │ 信令服务器  │        │参会者P1 │  │参会者P2 │  │参会者P3 │
└────┬────┘        └──────┬─────┘        └────┬────┘  └────┬────┘  └────┬────┘
     │                    │                    │            │            │
     
═══════════════════════════════════════════════════════════════════════════════
  阶段1: 创建会议
═══════════════════════════════════════════════════════════════════════════════

     │                    │                    │            │            │
     │ 【H主动】①          │                    │            │            │
     │ meeting_create     │                    │            │            │
     ├───────────────────>│                    │            │            │
     │ 创建会议            │                    │            │            │
     │                    │                    │            │            │
     │                    ├─ CreateMeeting()   │            │            │
     │                    │  • 生成 meeting_id │            │            │
     │                    │  • 设置主持人       │            │            │
     │                    │  • 配置会议设置     │            │            │
     │                    │  • 状态: scheduled │            │            │
     │                    │                    │            │            │
     │ ② meeting_create   │                    │            │            │
     │<───────────────────┤                    │            │            │
     │ 返回会议信息        │                    │            │            │
     │                    │                    │            │            │

═══════════════════════════════════════════════════════════════════════════════
  阶段2: 参会者加入
═══════════════════════════════════════════════════════════════════════════════

     │                    │                    │            │            │
     │                    │ 【P1主动】③         │            │            │
     │                    │ meeting_join       │            │            │
     │                    │<───────────────────┤            │            │
     │                    │ 加入会议            │            │            │
     │                    │                    │            │            │
     │                    ├─ JoinMeeting()     │            │            │
     │                    │  • 检查会议是否存在  │            │            │
     │                    │  • 检查人数限制     │            │            │
     │                    │  • 应用入会设置     │            │            │
     │                    │  • 状态→ ongoing   │            │            │
     │                    │                    │            │            │
     │ ④ meeting_join     │                    │            │            │
     │<───────────────────┤                    │            │            │
     │ 通知主持人          │                    │            │            │
     │                    │                    │            │            │
     │                    │ ⑤ meeting_join     │            │            │
     │                    ├───────────────────>│            │            │
     │                    │ 返回会议信息        │            │            │
     │                    │                    │            │            │
     │                    │ 【P2主动】⑥         │            │            │
     │                    │ meeting_join       │            │            │
     │                    │<─────────────────────────────────┤            │
     │                    │                    │            │            │
     │ ⑦ meeting_join     │                    │            │            │
     │<───────────────────┼───────────────────>│            │            │
     │ (通知H和P1)         │                    ├───────────>│            │
     │                    │                    │            │            │
     │                    │ ⑧ meeting_join     │            │            │
     │                    ├─────────────────────────────────>│            │
     │                    │ 返回会议信息        │            │            │
     │                    │                    │            │            │
     │                    │ 【P3加入过程类似...】│            │            │
     │                    │                    │            │            │

═══════════════════════════════════════════════════════════════════════════════
  阶段3: 加入房间和WebRTC连接 (与群组通话类似)
═══════════════════════════════════════════════════════════════════════════════

     │                    │                    │            │            │
     │ join_room          │                    │            │            │
     ├───────────────────>│                    │            │            │
     │                    │                    │            │            │
     │                    │ join_room          │            │            │
     │                    │<───────────────────┤            │            │
     │                    │                    │            │            │
     │ 【每个人与SFU建立WebRTC连接】            │            │            │
     │ offer/answer/ice   │                    │            │            │
     │                    │                    │            │            │
     │ 【媒体流通过SFU转发】                    │            │            │
     │                    │                    │            │            │

═══════════════════════════════════════════════════════════════════════════════
  阶段4: 会议控制
═══════════════════════════════════════════════════════════════════════════════

     │                    │                    │            │            │
     │ 【H主动】⑨          │                    │            │            │
     │ meeting_control    │                    │            │            │
     ├───────────────────>│                    │            │            │
     │ action:            │                    │            │            │
     │ mute_participant   │                    │            │            │
     │ target: P1         │                    │            │            │
     │                    │                    │            │            │
     │                    ├─ MuteParticipant() │            │            │
     │                    │  • 验证主持人权限   │            │            │
     │                    │  • 更新参会者状态   │            │            │
     │                    │                    │            │            │
     │ ⑩ meeting_control  │                    │            │            │
     │<───────────────────┤                    │            │            │
     │ (确认操作)          │                    │            │            │
     │                    │                    │            │            │
     │                    │ ⑪ meeting_control  │            │            │
     │                    ├───────────────────>│            │            │
     │                    │ 通知P1被静音        │            │            │
     │                    │                    │            │            │
     │                    │ ⑫ meeting_control  │            │            │
     │                    ├─────────────────────────────────>│            │
     │                    │ 通知P2: P1被静音    │            ├───────────>│
     │                    │                    │            │            │

═══════════════════════════════════════════════════════════════════════════════
  阶段5: 参会者离开
═══════════════════════════════════════════════════════════════════════════════

     │                    │                    │            │            │
     │                    │ 【P2主动】⑬         │            │            │
     │                    │ meeting_leave      │            │            │
     │                    │<─────────────────────────────────┤            │
     │                    │ 离开会议            │            │            │
     │                    │                    │            │            │
     │                    ├─ LeaveMeeting()    │            │            │
     │                    │  • 从参会者移除     │            │            │
     │                    │  • 更新会议信息     │            │            │
     │                    │                    │            │            │
     │ ⑭ meeting_leave    │                    │            │            │
     │<───────────────────┼───────────────────>│            │            │
     │ (通知其他人: P2离开)│                    │            ├───────────>│
     │                    │                    │            │            │
     │                    │ leave_room         │            │            │
     │                    │<─────────────────────────────────┤            │
     │                    │                    │            │            │

═══════════════════════════════════════════════════════════════════════════════
  阶段6: 结束会议
═══════════════════════════════════════════════════════════════════════════════

     │                    │                    │            │            │
     │ 【H主动】⑮          │                    │            │            │
     │ meeting_control    │                    │            │            │
     ├───────────────────>│                    │            │            │
     │ action:            │                    │            │            │
     │ end_meeting        │                    │            │            │
     │                    │                    │            │            │
     │                    ├─ EndMeeting()      │            │            │
     │                    │  • 验证主持人权限   │            │            │
     │                    │  • 更新状态: ended │            │            │
     │                    │  • 清理所有资源     │            │            │
     │                    │                    │            │            │
     │ ⑯ meeting_control  │                    │            │            │
     │<───────────────────┼───────────────────>│            │            │
     │ (通知所有人: 会议结束)                   ├───────────>│            │
     │                    │                    │            ├───────────>│
     │                    │                    │            │            │
     │ 【所有人离开房间】  │                    │            │            │
     │ leave_room         │                    │            │            │
     │                    │                    │            │            │
```

---

## 阶段详解

### 阶段1: 创建会议

#### 消息 ①: meeting_create (H → 服务器)

**发起方**: 主持人  
**目的**: 创建新会议

**消息格式**:
```json
{
    "type": "meeting_create",
    "user_id": "hostH",
    "data": {
        "title": "项目讨论会议",
        "description": "讨论Q4项目进度和下一步计划",
        "settings": {
            "max_participants": 20,
            "allow_screen_share": true,
            "allow_recording": true,
            "mute_on_join": false,
            "video_on_join": true,
            "waiting_room": false
        }
    },
    "timestamp": "2025-10-13T10:00:00Z"
}
```

**核心字段**:
- `title`: 会议标题（必填）
- `description`: 会议描述（可选）
- `settings`: 会议设置对象

**会议设置说明**:
- `max_participants`: 最大参会人数 (1-50)
- `allow_screen_share`: 是否允许屏幕共享
- `allow_recording`: 是否允许录制
- `mute_on_join`: 加入时是否自动静音
- `video_on_join`: 加入时是否自动开启视频
- `waiting_room`: 是否启用等候室

---

#### 消息 ②: meeting_create (服务器 → H)

**发起方**: 服务器（响应）  
**目的**: 返回创建的会议信息

**消息格式**:
```json
{
    "type": "meeting_create",
    "user_id": "hostH",
    "data": {
        "id": "meeting_1697184000",
        "title": "项目讨论会议",
        "description": "讨论Q4项目进度和下一步计划",
        "host_id": "hostH",
        "participants": [],
        "status": "scheduled",
        "scheduled_at": "2025-10-13T10:00:00Z",
        "settings": {
            "max_participants": 20,
            "allow_screen_share": true,
            "allow_recording": true,
            "mute_on_join": false,
            "video_on_join": true,
            "waiting_room": false
        }
    },
    "timestamp": "2025-10-13T10:00:00Z"
}
```

**核心字段**:
- `id`: 会议唯一标识（meeting_id）
- `host_id`: 主持人用户ID
- `participants`: 参会者列表（初始为空）
- `status`: 会议状态 (`scheduled` → `ongoing` → `ended`)

---

### 阶段2: 参会者加入

#### 消息 ③⑥: meeting_join (P1, P2, P3 → 服务器)

**发起方**: 参会者  
**目的**: 加入会议

**消息格式**:
```json
{
    "type": "meeting_join",
    "user_id": "participantP1",
    "data": {
        "meeting_id": "meeting_1697184000",
        "username": "张三"
    },
    "timestamp": "2025-10-13T10:05:00Z"
}
```

**核心字段**:
- `meeting_id`: 要加入的会议ID
- `username`: 显示名称（可选，默认为用户ID）

**服务器处理**:
- 调用 `JoinMeeting(meetingID, userID, username)`
- 检查会议是否存在
- 检查人数是否达到上限
- 应用会议设置（静音、视频）
- 第一人加入时，状态更新为 `ongoing`

---

#### 消息 ④⑦: meeting_join (服务器 → 所有人)

**发起方**: 服务器（广播）  
**目的**: 通知已在会议的成员，有新人加入

**消息格式**:
```json
{
    "type": "meeting_join",
    "user_id": "participantP1",
    "data": {
        "id": "meeting_1697184000",
        "title": "项目讨论会议",
        "participants": [
            {
                "user_id": "hostH",
                "username": "主持人",
                "role": "host",
                "is_muted": false,
                "is_video_on": true
            },
            {
                "user_id": "participantP1",
                "username": "张三",
                "role": "participant",
                "is_muted": false,
                "is_video_on": true
            }
        ],
        "status": "ongoing",
        "started_at": "2025-10-13T10:05:00Z"
    },
    "timestamp": "2025-10-13T10:05:00Z"
}
```

---

#### 消息 ⑤⑧: meeting_join (服务器 → 加入者)

**发起方**: 服务器（响应）  
**目的**: 返回当前会议完整信息

**作用**: 让新加入者了解当前会议状态和所有参会者

---

### 阶段4: 会议控制

#### 消息 ⑨: meeting_control (H → 服务器)

**发起方**: 主持人  
**目的**: 控制会议和参会者

**消息格式**:
```json
{
    "type": "meeting_control",
    "user_id": "hostH",
    "data": {
        "meeting_id": "meeting_1697184000",
        "action": "mute_participant",
        "participant_id": "participantP1"
    },
    "timestamp": "2025-10-13T10:10:00Z"
}
```

**支持的控制动作**:

1. **静音参与者**:
```json
{
    "action": "mute_participant",
    "participant_id": "participantP1"
}
```

2. **取消静音参与者**:
```json
{
    "action": "unmute_participant",
    "participant_id": "participantP1"
}
```

3. **移除参与者** (可扩展):
```json
{
    "action": "remove_participant",
    "participant_id": "participantP1"
}
```

4. **全体静音** (可扩展):
```json
{
    "action": "mute_all"
}
```

5. **锁定会议** (可扩展):
```json
{
    "action": "lock_meeting"
}
```

6. **结束会议**:
```json
{
    "action": "end_meeting"
}
```

---

#### 消息 ⑩⑪⑫: meeting_control (服务器 → 所有人)

**发起方**: 服务器（广播）  
**目的**: 同步控制结果给所有参会者

**消息格式**:
```json
{
    "type": "meeting_control",
    "user_id": "hostH",
    "data": {
        "meeting": {
            "id": "meeting_1697184000",
            "participants": [...]
        },
        "action": "mute_participant",
        "target_user": "participantP1"
    },
    "timestamp": "2025-10-13T10:10:00Z"
}
```

---

### 阶段5: 参会者离开

#### 消息 ⑬: meeting_leave (P2 → 服务器)

**发起方**: 参会者  
**目的**: 主动离开会议

**消息格式**:
```json
{
    "type": "meeting_leave",
    "user_id": "participantP2",
    "data": {
        "meeting_id": "meeting_1697184000"
    },
    "timestamp": "2025-10-13T10:12:00Z"
}
```

**服务器处理**:
- 调用 `LeaveMeeting(meetingID, userID)`
- 从参会者列表移除
- 广播给其他人
- 如果是主持人离开，会议结束

---

#### 消息 ⑭: meeting_leave (服务器 → 所有人)

**发起方**: 服务器（广播）  
**目的**: 通知其他人，有参会者离开

---

### 阶段6: 结束会议

#### 消息 ⑮: meeting_control (H → 服务器)

**发起方**: 主持人  
**目的**: 结束会议（action: end_meeting）

**权限**: 仅主持人可以结束会议

---

#### 消息 ⑯: meeting_control (服务器 → 所有人)

**发起方**: 服务器（广播）  
**目的**: 通知所有参会者，会议已结束

**消息格式**:
```json
{
    "type": "meeting_control",
    "user_id": "hostH",
    "data": {
        "meeting": {
            "id": "meeting_1697184000",
            "status": "ended",
            "ended_at": "2025-10-13T10:15:00Z"
        },
        "action": "end_meeting"
    },
    "timestamp": "2025-10-13T10:15:00Z"
}
```

---

## 核心数据结构

### Meeting 对象

```json
{
    "id": "meeting_1697184000",
    "title": "项目讨论会议",
    "description": "讨论Q4项目进度和下一步计划",
    "host_id": "hostH",
    "participants": [
        {
            "user_id": "hostH",
            "username": "主持人",
            "role": "host",
            "is_muted": false,
            "is_video_on": true,
            "joined_at": "2025-10-13T10:00:00Z"
        },
        {
            "user_id": "participantP1",
            "username": "张三",
            "role": "participant",
            "is_muted": false,
            "is_video_on": true,
            "joined_at": "2025-10-13T10:05:00Z"
        }
    ],
    "status": "ongoing",
    "scheduled_at": "2025-10-13T10:00:00Z",
    "started_at": "2025-10-13T10:05:00Z",
    "ended_at": null,
    "settings": {
        "max_participants": 20,
        "allow_screen_share": true,
        "allow_recording": true,
        "mute_on_join": false,
        "video_on_join": true,
        "waiting_room": false
    }
}
```

---

### MeetingSettings 对象

```json
{
    "max_participants": 20,
    "allow_screen_share": true,
    "allow_recording": true,
    "mute_on_join": false,
    "video_on_join": true,
    "waiting_room": false
}
```

**字段说明**:
- `max_participants`: 最大参会人数
- `allow_screen_share`: 是否允许参会者共享屏幕
- `allow_recording`: 是否允许录制会议
- `mute_on_join`: 新加入者是否自动静音
- `video_on_join`: 新加入者是否自动开启视频
- `waiting_room`: 是否启用等候室（需要主持人批准）

---

### 会议状态

```
scheduled → ongoing → ended
    ↓          ↓
  取消      中断
```

**状态说明**:
- `scheduled`: 已创建，等待开始
- `ongoing`: 进行中（至少1人加入）
- `ended`: 已结束
- `cancelled`: 已取消（可扩展）

---

## 会议控制功能

### 主持人权限

| 功能 | 说明 | 消息 |
|-----|------|------|
| **静音参与者** | 强制静音某个参与者 | mute_participant |
| **取消静音** | 解除静音 | unmute_participant |
| **全体静音** | 静音所有参与者 | mute_all |
| **移除参与者** | 踢出某个参与者 | remove_participant |
| **锁定会议** | 禁止新人加入 | lock_meeting |
| **结束会议** | 结束整个会议 | end_meeting |
| **开启/关闭录制** | 控制会议录制 | start_recording/stop_recording |

### 参与者权限

| 功能 | 说明 |
|-----|------|
| **控制自己的音视频** | 可以自主静音/开关视频 |
| **举手** | 请求发言（可扩展） |
| **屏幕共享** | 共享屏幕（如果允许） |
| **发送消息** | 会议聊天（可扩展） |
| **离开会议** | 主动退出 |

---

## 与群组通话的区别

### 功能对比

| 特性 | 群组通话 | 会议功能 |
|-----|---------|---------|
| **使用场景** | 临时多人通话 | 正式会议 |
| **创建方式** | 邀请制 | 预约制 |
| **会议ID** | 动态生成 | 可提前生成 |
| **主持人权限** | 基础（仅结束） | 完整（控制所有人） |
| **参与方式** | 必须被邀请 | 可通过ID加入 |
| **会议设置** | 无 | 丰富的配置 |
| **等候室** | 不支持 | 支持 |
| **录制功能** | 不支持 | 支持 |
| **会议控制** | 最小化 | 完整控制 |
| **状态管理** | 简单 | 完整生命周期 |

---

### 消息对比

| 阶段 | 群组通话 | 会议功能 |
|-----|---------|---------|
| **发起** | group_invite | meeting_create |
| **加入** | group_join | meeting_join |
| **离开** | group_leave | meeting_leave |
| **控制** | 无 | meeting_control |
| **结束** | call_end | meeting_control(end_meeting) |

---

## 关键要点

### 1. 会议生命周期

```
创建 → 等待 → 进行中 → 结束
  ↓      ↓        ↓       ↓
保存   加入     控制    清理
```

**各阶段说明**:
- **创建**: 主持人创建会议，生成meeting_id
- **等待**: 状态为 `scheduled`，等待参会者加入
- **进行中**: 第一人加入后，状态变为 `ongoing`
- **结束**: 主持人结束或主持人离开

### 2. 权限管理

**主持人**:
- 拥有所有控制权限
- 可以强制操作任何参会者
- 离开会议则会议结束

**参会者**:
- 只能控制自己的状态
- 可以离开但不能结束会议
- 受主持人控制

### 3. 会议设置的应用

**mute_on_join**:
- 新加入者自动静音
- 适用于大型会议

**video_on_join**:
- 控制新加入者的视频状态
- 节省带宽

**waiting_room**:
- 新加入者进入等候室
- 需要主持人批准才能进入
- 增强安全性

### 4. 扩展功能

**等候室功能**:
```json
{
    "type": "meeting_admission",
    "action": "approve",
    "participant_id": "user123"
}
```

**举手发言**:
```json
{
    "type": "meeting_signal",
    "action": "raise_hand",
    "user_id": "participant123"
}
```

**会议录制**:
```json
{
    "type": "meeting_control",
    "action": "start_recording"
}
```

### 5. 最佳实践

**创建会议**:
- 提前创建会议，生成会议ID
- 合理设置参会人数上限
- 根据会议性质配置入会设置

**管理会议**:
- 大型会议启用 `mute_on_join`
- 敏感会议启用 `waiting_room`
- 及时移除干扰参会者

**结束会议**:
- 主持人主动结束会议
- 或设置超时自动结束
- 保存会议记录（可扩展）

---

## 消息统计

### 10人会议示例

| 阶段 | 消息类型 | 数量 |
|-----|---------|------|
| 创建会议 | meeting_create | 2 |
| 参会者加入 | meeting_join | 10 + 45 |
| 加入房间 | join_room/room_info | 20 |
| WebRTC连接 | offer/answer/ice | 60-120 |
| 会议控制 | meeting_control | 0-N |
| 参会者离开 | meeting_leave | 0-N |
| 结束会议 | meeting_control | 1 + 9 |

**总计**: 约 138-216 条消息

---

## 相关文档

- [一对一通话流程](./streaming-one-to-one-call-flow.md)
- [群组通话流程](streaming-group-call-flow.md)
- [等候室功能](./meeting-waiting-room.md)
- [会议录制功能](./meeting-recording.md)

---

**文档结束**

如有疑问，请参考源码：
- `apps/streaming/internal/handler/signaling.go` (994-1245行)
- `apps/streaming/internal/logic/meeting_manager.go`

