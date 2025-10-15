# 流媒体服务 API 使用示例

## 概述

本文档提供了流媒体服务的完整 API 使用示例，包括一对一通话、群组通话、会议、录屏和直播功能。

## WebSocket 连接

### 连接地址
```
ws://localhost:10093/ws
```

### 消息格式
```json
{
    "type": "message_type",
    "user_id": "user_123",
    "room_id": "room_456",
    "data": {...},
    "timestamp": "2025-01-01T00:00:00Z"
}
```

## 1. 一对一通话

### 1.1 发起通话邀请

```javascript
// 发起一对一通话
const callInvite = {
    type: 'call_invite',
    user_id: 'caller_123',
    data: {
        callee_id: 'callee_456'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(callInvite));
```

### 1.2 接受通话

```javascript
// 接受通话
const callAccept = {
    type: 'call_accept',
    user_id: 'callee_456',
    data: {
        call_id: 'call_789'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(callAccept));
```

### 1.3 拒绝通话

```javascript
// 拒绝通话
const callReject = {
    type: 'call_reject',
    user_id: 'callee_456',
    data: {
        call_id: 'call_789'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(callReject));
```

### 1.4 结束通话

```javascript
// 结束通话
const callEnd = {
    type: 'call_end',
    user_id: 'caller_123',
    data: {
        call_id: 'call_789'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(callEnd));
```

## 2. 群组通话

### 2.1 发起群组通话

```javascript
// 发起群组通话
const groupInvite = {
    type: 'group_invite',
    user_id: 'host_123',
    data: {
        participants: ['user_456', 'user_789', 'user_101']
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(groupInvite));
```

### 2.2 加入群组通话

```javascript
// 加入群组通话
const groupJoin = {
    type: 'group_join',
    user_id: 'user_456',
    data: {
        call_id: 'group_call_789'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(groupJoin));
```

### 2.3 离开群组通话

```javascript
// 离开群组通话
const groupLeave = {
    type: 'group_leave',
    user_id: 'user_456',
    data: {
        call_id: 'group_call_789'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(groupLeave));
```

## 3. 会议功能

### 3.1 创建会议

```javascript
// 创建会议
const meetingCreate = {
    type: 'meeting_create',
    user_id: 'host_123',
    data: {
        title: '项目讨论会议',
        description: '讨论项目进度和下一步计划',
        settings: {
            max_participants: 20,
            allow_screen_share: true,
            allow_recording: true,
            mute_on_join: false,
            video_on_join: true,
            waiting_room: false
        }
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(meetingCreate));
```

### 3.2 加入会议

```javascript
// 加入会议
const meetingJoin = {
    type: 'meeting_join',
    user_id: 'participant_456',
    data: {
        meeting_id: 'meeting_789',
        username: '张三'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(meetingJoin));
```

### 3.3 离开会议

```javascript
// 离开会议
const meetingLeave = {
    type: 'meeting_leave',
    user_id: 'participant_456',
    data: {
        meeting_id: 'meeting_789'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(meetingLeave));
```

### 3.4 会议控制

```javascript
// 静音参与者
const meetingControl = {
    type: 'meeting_control',
    user_id: 'host_123',
    data: {
        meeting_id: 'meeting_789',
        action: 'mute_participant',
        participant_id: 'participant_456'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(meetingControl));

// 取消静音参与者
const unmuteControl = {
    type: 'meeting_control',
    user_id: 'host_123',
    data: {
        meeting_id: 'meeting_789',
        action: 'unmute_participant',
        participant_id: 'participant_456'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(unmuteControl));

// 结束会议
const endMeeting = {
    type: 'meeting_control',
    user_id: 'host_123',
    data: {
        meeting_id: 'meeting_789',
        action: 'end_meeting'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(endMeeting));
```

## 4. 录屏功能

### 4.1 开始录屏

```javascript
// 开始录屏
const screenShareStart = {
    type: 'screen_share_start',
    user_id: 'user_123',
    data: {
        room_id: 'room_456',
        quality: 'high' // high, medium, low
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(screenShareStart));
```

### 4.2 停止录屏

```javascript
// 停止录屏
const screenShareStop = {
    type: 'screen_share_stop',
    user_id: 'user_123',
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(screenShareStop));
```

### 4.3 请求录屏

```javascript
// 请求其他用户录屏
const screenShareRequest = {
    type: 'screen_share_request',
    user_id: 'requester_123',
    data: {
        target_user_id: 'target_456',
        room_id: 'room_789'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(screenShareRequest));
```

## 5. 直播功能

### 5.1 开始直播

```javascript
// 开始直播
const liveStart = {
    type: 'live_start',
    user_id: 'streamer_123',
    data: {
        title: '技术分享直播',
        description: '分享最新的技术趋势和开发经验'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(liveStart));
```

### 5.2 停止直播

```javascript
// 停止直播
const liveStop = {
    type: 'live_stop',
    user_id: 'streamer_123',
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(liveStop));
```

### 5.3 加入直播

```javascript
// 加入直播观看
const liveJoin = {
    type: 'live_join',
    user_id: 'viewer_456',
    data: {
        stream_id: 'live_stream_789'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(liveJoin));
```

### 5.4 离开直播

```javascript
// 离开直播
const liveLeave = {
    type: 'live_leave',
    user_id: 'viewer_456',
    data: {
        stream_id: 'live_stream_789'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(liveLeave));
```

## 6. 媒体控制

### 6.1 静音/取消静音

```javascript
// 静音
const mute = {
    type: 'mute',
    user_id: 'user_123',
    data: {
        room_id: 'room_456'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(mute));

// 取消静音
const unmute = {
    type: 'unmute',
    user_id: 'user_123',
    data: {
        room_id: 'room_456'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(unmute));
```

### 6.2 开启/关闭视频

```javascript
// 开启视频
const videoOn = {
    type: 'video_on',
    user_id: 'user_123',
    data: {
        room_id: 'room_456'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(videoOn));

// 关闭视频
const videoOff = {
    type: 'video_off',
    user_id: 'user_123',
    data: {
        room_id: 'room_456'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(videoOff));
```

## 7. WebRTC 信令

### 7.1 加入房间

```javascript
// 加入房间
const joinRoom = {
    type: 'join_room',
    user_id: 'user_123',
    room_id: 'room_456',
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(joinRoom));
```

### 7.2 离开房间

```javascript
// 离开房间
const leaveRoom = {
    type: 'leave_room',
    user_id: 'user_123',
    room_id: 'room_456',
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(leaveRoom));
```

### 7.3 发送 Offer

```javascript
// 发送 Offer
const offer = {
    type: 'offer',
    user_id: 'user_123',
    room_id: 'room_456',
    data: {
        sdp: 'v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n...',
        type: 'offer'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(offer));
```

### 7.4 发送 Answer

```javascript
// 发送 Answer
const answer = {
    type: 'answer',
    user_id: 'user_456',
    room_id: 'room_456',
    data: {
        sdp: 'v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n...',
        type: 'answer'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(answer));
```

### 7.5 发送 ICE 候选

```javascript
// 发送 ICE 候选
const iceCandidate = {
    type: 'ice_candidate',
    user_id: 'user_123',
    room_id: 'room_456',
    data: {
        candidate: 'candidate:1 1 UDP 2113667326 192.168.1.100 54400 typ host',
        sdpMLineIndex: 0,
        sdpMid: '0'
    },
    timestamp: new Date().toISOString()
};

ws.send(JSON.stringify(iceCandidate));
```

## 8. 完整示例

### 8.1 一对一通话完整流程

```javascript
class OneToOneCall {
    constructor(ws, userId) {
        this.ws = ws;
        this.userId = userId;
        this.currentCall = null;
    }

    // 发起通话
    startCall(calleeId) {
        const callInvite = {
            type: 'call_invite',
            user_id: this.userId,
            data: {
                callee_id: calleeId
            },
            timestamp: new Date().toISOString()
        };
        
        this.ws.send(JSON.stringify(callInvite));
    }

    // 接受通话
    acceptCall(callId) {
        const callAccept = {
            type: 'call_accept',
            user_id: this.userId,
            data: {
                call_id: callId
            },
            timestamp: new Date().toISOString()
        };
        
        this.ws.send(JSON.stringify(callAccept));
    }

    // 拒绝通话
    rejectCall(callId) {
        const callReject = {
            type: 'call_reject',
            user_id: this.userId,
            data: {
                call_id: callId
            },
            timestamp: new Date().toISOString()
        };
        
        this.ws.send(JSON.stringify(callReject));
    }

    // 结束通话
    endCall(callId) {
        const callEnd = {
            type: 'call_end',
            user_id: this.userId,
            data: {
                call_id: callId
            },
            timestamp: new Date().toISOString()
        };
        
        this.ws.send(JSON.stringify(callEnd));
    }

    // 处理消息
    handleMessage(message) {
        switch (message.type) {
            case 'call_invite':
                console.log('收到通话邀请:', message.data);
                // 显示通话邀请界面
                break;
                
            case 'call_accept':
                console.log('通话被接受:', message.data);
                // 开始建立 WebRTC 连接
                break;
                
            case 'call_reject':
                console.log('通话被拒绝:', message.data);
                // 显示通话被拒绝
                break;
                
            case 'call_end':
                console.log('通话结束:', message.data);
                // 清理资源
                break;
        }
    }
}

// 使用示例
const ws = new WebSocket('ws://localhost:10093/ws');
const call = new OneToOneCall(ws, 'user_123');

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    call.handleMessage(message);
};

// 发起通话
call.startCall('user_456');
```

### 8.2 会议系统完整流程

```javascript
class MeetingSystem {
    constructor(ws, userId) {
        this.ws = ws;
        this.userId = userId;
        this.currentMeeting = null;
    }

    // 创建会议
    createMeeting(title, description, settings) {
        const meetingCreate = {
            type: 'meeting_create',
            user_id: this.userId,
            data: {
                title: title,
                description: description,
                settings: settings
            },
            timestamp: new Date().toISOString()
        };
        
        this.ws.send(JSON.stringify(meetingCreate));
    }

    // 加入会议
    joinMeeting(meetingId, username) {
        const meetingJoin = {
            type: 'meeting_join',
            user_id: this.userId,
            data: {
                meeting_id: meetingId,
                username: username
            },
            timestamp: new Date().toISOString()
        };
        
        this.ws.send(JSON.stringify(meetingJoin));
    }

    // 离开会议
    leaveMeeting(meetingId) {
        const meetingLeave = {
            type: 'meeting_leave',
            user_id: this.userId,
            data: {
                meeting_id: meetingId
            },
            timestamp: new Date().toISOString()
        };
        
        this.ws.send(JSON.stringify(meetingLeave));
    }

    // 会议控制
    controlMeeting(meetingId, action, participantId = null) {
        const meetingControl = {
            type: 'meeting_control',
            user_id: this.userId,
            data: {
                meeting_id: meetingId,
                action: action,
                participant_id: participantId
            },
            timestamp: new Date().toISOString()
        };
        
        this.ws.send(JSON.stringify(meetingControl));
    }

    // 处理消息
    handleMessage(message) {
        switch (message.type) {
            case 'meeting_create':
                console.log('会议创建成功:', message.data);
                this.currentMeeting = message.data;
                break;
                
            case 'meeting_join':
                console.log('用户加入会议:', message.data);
                // 更新会议参与者列表
                break;
                
            case 'meeting_leave':
                console.log('用户离开会议:', message.data);
                // 更新会议参与者列表
                break;
                
            case 'meeting_control':
                console.log('会议控制:', message.data);
                // 处理会议控制事件
                break;
        }
    }
}

// 使用示例
const ws = new WebSocket('ws://localhost:10093/ws');
const meeting = new MeetingSystem(ws, 'host_123');

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    meeting.handleMessage(message);
};

// 创建会议
meeting.createMeeting('项目讨论', '讨论项目进度', {
    max_participants: 20,
    allow_screen_share: true,
    allow_recording: true,
    mute_on_join: false,
    video_on_join: true,
    waiting_room: false
});
```

## 9. 错误处理

### 9.1 错误消息格式

```json
{
    "type": "error",
    "user_id": "user_123",
    "data": {
        "error": "错误描述"
    },
    "timestamp": "2025-01-01T00:00:00Z"
}
```

### 9.2 错误处理示例

```javascript
ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    
    if (message.type === 'error') {
        console.error('收到错误消息:', message.data.error);
        // 处理错误
        handleError(message.data.error);
    } else {
        // 处理正常消息
        handleMessage(message);
    }
};

function handleError(error) {
    switch (error) {
        case 'user is already in a call':
            alert('用户正在通话中');
            break;
        case 'meeting is full':
            alert('会议已满');
            break;
        case 'permission denied':
            alert('权限不足');
            break;
        default:
            alert('未知错误: ' + error);
    }
}
```

## 10. 最佳实践

### 10.1 连接管理

```javascript
class StreamingClient {
    constructor(url, userId) {
        this.url = url;
        this.userId = userId;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectInterval = 1000;
    }

    connect() {
        this.ws = new WebSocket(this.url);
        
        this.ws.onopen = () => {
            console.log('WebSocket 连接已建立');
            this.reconnectAttempts = 0;
        };
        
        this.ws.onclose = () => {
            console.log('WebSocket 连接已关闭');
            this.reconnect();
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket 错误:', error);
        };
        
        this.ws.onmessage = (event) => {
            this.handleMessage(JSON.parse(event.data));
        };
    }

    reconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            
            setTimeout(() => {
                this.connect();
            }, this.reconnectInterval * this.reconnectAttempts);
        } else {
            console.error('重连失败，已达到最大重连次数');
        }
    }

    send(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            console.error('WebSocket 连接未建立');
        }
    }

    handleMessage(message) {
        // 处理消息
        console.log('收到消息:', message);
    }
}
```

### 10.2 状态管理

```javascript
class CallState {
    constructor() {
        this.state = 'idle'; // idle, calling, ringing, connected, ended
        this.currentCall = null;
        this.participants = [];
    }

    updateState(newState, data = null) {
        this.state = newState;
        if (data) {
            this.currentCall = data;
        }
        this.notifyStateChange();
    }

    notifyStateChange() {
        // 通知状态变化
        console.log('状态变化:', this.state, this.currentCall);
    }
}
```

## 总结

本文档提供了流媒体服务的完整 API 使用示例，包括：

1. **一对一通话**: 完整的通话流程
2. **群组通话**: 多人通话管理
3. **会议系统**: 会议创建、加入、控制
4. **录屏功能**: 屏幕共享和录制
5. **直播功能**: 直播推流和观看
6. **媒体控制**: 静音、视频控制
7. **WebRTC 信令**: 连接建立和媒体传输
8. **错误处理**: 异常情况处理
9. **最佳实践**: 连接管理和状态管理

通过这些示例，你可以快速集成和使用流媒体服务的各种功能。
