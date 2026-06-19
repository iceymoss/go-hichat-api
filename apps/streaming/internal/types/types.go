// Package types 定义了流媒体服务的所有类型和接口
// 包括信令消息、房间管理、用户管理、通话管理等核心数据结构
package types

import (
	"time"

	"github.com/pion/webrtc/v3"
)

// SignalingMessageType 信令消息类型枚举
// 定义了所有可能的WebSocket信令消息类型
type SignalingMessageType string

const (
	// 基础通话
	MessageTypeJoinRoom     SignalingMessageType = "join_room"
	MessageTypeLeaveRoom    SignalingMessageType = "leave_room"
	MessageTypeOffer        SignalingMessageType = "offer"
	MessageTypeAnswer       SignalingMessageType = "answer"
	MessageTypeIceCandidate SignalingMessageType = "ice_candidate"
	MessageTypeRoomInfo     SignalingMessageType = "room_info"
	MessageTypeUserJoined   SignalingMessageType = "user_joined"
	MessageTypeUserLeft     SignalingMessageType = "user_left"
	MessageTypeError        SignalingMessageType = "error"

	// 一对一通话
	MessageTypeCallInvite  SignalingMessageType = "call_invite"
	MessageTypeCallAccept  SignalingMessageType = "call_accept"
	MessageTypeCallReject  SignalingMessageType = "call_reject"
	MessageTypeCallCancel  SignalingMessageType = "call_cancel"
	MessageTypeCallEnd     SignalingMessageType = "call_end"
	MessageTypeCallCreated SignalingMessageType = "call_created" // 服务端回执主叫：携带 callId
	MessageTypeMediaState  SignalingMessageType = "media_state"  // 静音/开关摄像头

	// 群组通话
	MessageTypeGroupInvite SignalingMessageType = "group_invite"
	MessageTypeGroupJoin   SignalingMessageType = "group_join"
	MessageTypeGroupLeave  SignalingMessageType = "group_leave"

	// 会议功能
	MessageTypeMeetingCreate  SignalingMessageType = "meeting_create"
	MessageTypeMeetingJoin    SignalingMessageType = "meeting_join"
	MessageTypeMeetingLeave   SignalingMessageType = "meeting_leave"
	MessageTypeMeetingControl SignalingMessageType = "meeting_control"

	// 录屏功能
	MessageTypeScreenShareStart   SignalingMessageType = "screen_share_start"
	MessageTypeScreenShareStop    SignalingMessageType = "screen_share_stop"
	MessageTypeScreenShareRequest SignalingMessageType = "screen_share_request"

	// 直播功能
	MessageTypeLiveStart SignalingMessageType = "live_start"
	MessageTypeLiveStop  SignalingMessageType = "live_stop"
	MessageTypeLiveJoin  SignalingMessageType = "live_join"
	MessageTypeLiveLeave SignalingMessageType = "live_leave"

	// 媒体控制
	MessageTypeMute     SignalingMessageType = "mute"
	MessageTypeUnmute   SignalingMessageType = "unmute"
	MessageTypeVideoOn  SignalingMessageType = "video_on"
	MessageTypeVideoOff SignalingMessageType = "video_off"
)

// SignalingMessage 信令消息结构体
// 用于WebSocket通信的核心消息格式
type SignalingMessage struct {
	Type      SignalingMessageType `json:"type"`              // 消息类型
	RoomID    string               `json:"room_id,omitempty"` // 房间ID，可选
	UserID    string               `json:"user_id,omitempty"` // 用户ID，可选
	Data      interface{}          `json:"data,omitempty"`    // 消息数据，根据类型不同而变化
	Timestamp time.Time            `json:"timestamp"`         // 消息时间戳
}

// WebRTCMessage WebRTC Offer/Answer 消息结构体
// 用于WebRTC连接协商的SDP消息
type WebRTCMessage struct {
	SDP  string `json:"sdp"`  // SDP描述信息
	Type string `json:"type"` // SDP类型："offer" 或 "answer"
}

// ICECandidateMessage ICE候选消息结构体
// 用于WebRTC ICE候选交换，建立P2P连接
type ICECandidateMessage struct {
	Candidate     string  `json:"candidate"`     // ICE候选字符串
	SDPMLineIndex *uint16 `json:"sdpMLineIndex"` // SDP媒体行索引，指针类型允许为nil
	SDPMid        *string `json:"sdpMid"`        // SDP媒体ID，指针类型允许为nil
}

// 房间信息
type RoomInfo struct {
	RoomID    string    `json:"room_id"`
	Name      string    `json:"name"`
	Users     []User    `json:"users"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 用户信息
type User struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Avatar    string    `json:"avatar,omitempty"`
	JoinedAt  time.Time `json:"joined_at"`
	IsMuted   bool      `json:"is_muted"`
	IsVideoOn bool      `json:"is_video_on"`
	// 扩展字段
	Role   string `json:"role,omitempty"`   // 角色：host, participant, viewer
	Status string `json:"status,omitempty"` // 状态：online, offline, busy
	Device string `json:"device,omitempty"` // 设备类型：web, mobile, desktop
}

// 通话类型
type CallType string

const (
	CallTypeOneToOne CallType = "one_to_one"
	CallTypeGroup    CallType = "group"
	CallTypeMeeting  CallType = "meeting"
	CallTypeLive     CallType = "live"
)

// 通话状态
type CallStatus string

const (
	CallStatusIdle      CallStatus = "idle"
	CallStatusInviting  CallStatus = "inviting"
	CallStatusRinging   CallStatus = "ringing"
	CallStatusConnected CallStatus = "connected"
	CallStatusEnded     CallStatus = "ended"
	CallStatusFailed    CallStatus = "failed"
)

// 通话信息
type Call struct {
	ID           string     `json:"id"`
	Type         CallType   `json:"type"`
	Status       CallStatus `json:"status"`
	CallerID     string     `json:"caller_id"`
	Participants []string   `json:"participants"`
	RoomID       string     `json:"room_id"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	Duration     int64      `json:"duration,omitempty"` // 秒
}

// 会议信息
type Meeting struct {
	ID           string           `json:"id"`
	Title        string           `json:"title"`
	Description  string           `json:"description,omitempty"`
	HostID       string           `json:"host_id"`
	Participants []*User          `json:"participants"`
	Status       string           `json:"status"` // scheduled, ongoing, ended
	ScheduledAt  time.Time        `json:"scheduled_at"`
	StartedAt    *time.Time       `json:"started_at,omitempty"`
	EndedAt      *time.Time       `json:"ended_at,omitempty"`
	Settings     *MeetingSettings `json:"settings,omitempty"`
}

// 会议设置
type MeetingSettings struct {
	MaxParticipants  int  `json:"max_participants"`
	AllowScreenShare bool `json:"allow_screen_share"`
	AllowRecording   bool `json:"allow_recording"`
	MuteOnJoin       bool `json:"mute_on_join"`
	VideoOnJoin      bool `json:"video_on_join"`
	WaitingRoom      bool `json:"waiting_room"`
}

// 录屏信息
type ScreenShare struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	RoomID    string     `json:"room_id"`
	Status    string     `json:"status"` // active, paused, stopped
	StartedAt time.Time  `json:"started_at"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
	Duration  int64      `json:"duration,omitempty"`
	Quality   string     `json:"quality,omitempty"` // high, medium, low
}

// 直播信息
type LiveStream struct {
	ID          string     `json:"id"`
	StreamerID  string     `json:"streamer_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"` // active, paused, ended
	ViewerCount int        `json:"viewer_count"`
	StartedAt   time.Time  `json:"started_at"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	StreamURL   string     `json:"stream_url,omitempty"`
	CDNURL      string     `json:"cdn_url,omitempty"`
}

// 房间管理器接口
type RoomManager interface {
	// 创建房间
	CreateRoom(roomID, name string) (*Room, error)
	// 获取房间
	GetRoom(roomID string) (*Room, error)
	// 删除房间
	DeleteRoom(roomID string) error
	// 获取所有房间
	GetAllRooms() map[string]*Room
	// 清理过期房间
	CleanupExpiredRooms() error
}

// 房间接口
type Room interface {
	// 获取房间ID
	GetID() string
	// 获取房间名称
	GetName() string
	// 添加用户
	AddUser(user *User) error
	// 移除用户
	RemoveUser(userID string) error
	// 获取用户
	GetUser(userID string) (*User, error)
	// 获取所有用户
	GetUsers() []*User
	// 广播消息
	Broadcast(message *SignalingMessage) error
	// 发送消息给特定用户
	SendToUser(userID string, message *SignalingMessage) error
	// 获取用户数量
	GetUserCount() int
	// 检查房间是否为空
	IsEmpty() bool
	// 更新房间状态
	UpdateStatus() error
}

// 连接管理器接口
type ConnectionManager interface {
	// 添加连接
	AddConnection(userID string, conn WebRTCConnection) error
	// 移除连接
	RemoveConnection(userID string) error
	// 获取连接
	GetConnection(userID string) (WebRTCConnection, error)
	// 获取所有连接
	GetAllConnections() map[string]WebRTCConnection
	// 清理过期连接
	CleanupExpiredConnections() error
}

// WebRTC连接接口
type WebRTCConnection interface {
	// 获取用户ID
	GetUserID() string
	// 获取PeerConnection
	GetPeerConnection() *webrtc.PeerConnection
	// 发送消息
	SendMessage(message *SignalingMessage) error
	// 关闭连接
	Close() error
	// 检查连接状态
	IsConnected() bool
	// 获取连接时间
	GetConnectedAt() time.Time
}

// SFU接口
type SFU interface {
	// 处理媒体流
	HandleMediaStream(roomID, userID string, track *webrtc.TrackLocalStaticRTP) error
	// 转发媒体流
	ForwardMediaStream(roomID, fromUserID, toUserID string, track *webrtc.TrackLocalStaticRTP) error
	// 停止转发
	StopForwarding(roomID, userID string) error
	// 获取房间统计信息
	GetRoomStats(roomID string) (*RoomStats, error)
}

// 房间统计信息
type RoomStats struct {
	RoomID        string    `json:"room_id"`
	UserCount     int       `json:"user_count"`
	ActiveStreams int       `json:"active_streams"`
	Bandwidth     int64     `json:"bandwidth"` // bytes per second
	UpdatedAt     time.Time `json:"updated_at"`
}

// 信令服务器接口
type SignalingServer interface {
	// 启动服务器
	Start() error
	// 停止服务器
	Stop() error
	// 处理WebSocket连接
	HandleWebSocketConnection(conn WebSocketConnection) error
	// 处理信令消息
	HandleSignalingMessage(conn WebSocketConnection, message *SignalingMessage) error
}

// WebSocket连接接口
type WebSocketConnection interface {
	// 发送消息
	SendMessage(message *SignalingMessage) error
	// 接收消息
	ReceiveMessage() (*SignalingMessage, error)
	// 关闭连接
	Close() error
	// 获取用户ID
	GetUserID() string
	// 设置用户ID
	SetUserID(userID string)
	// 检查连接状态
	IsConnected() bool
}

// 媒体处理器接口
type MediaProcessor interface {
	// 处理音频流
	ProcessAudioStream(stream *webrtc.TrackLocalStaticRTP) error
	// 处理视频流
	ProcessVideoStream(stream *webrtc.TrackLocalStaticRTP) error
	// 混合音频流
	MixAudioStreams(streams []*webrtc.TrackLocalStaticRTP) (*webrtc.TrackLocalStaticRTP, error)
	// 录制流
	RecordStream(roomID, userID string, stream *webrtc.TrackLocalStaticRTP) error
	// 停止录制
	StopRecording(roomID, userID string) error
}

// 认证接口
type Authenticator interface {
	// 验证用户
	AuthenticateUser(token string) (*User, error)
	// 验证房间权限
	ValidateRoomAccess(userID, roomID string) error
	// 生成访问令牌
	GenerateAccessToken(userID string) (string, error)
}

// 事件处理器接口
type EventHandler interface {
	// 处理用户加入事件
	OnUserJoined(roomID, userID string) error
	// 处理用户离开事件
	OnUserLeft(roomID, userID string) error
	// 处理房间创建事件
	OnRoomCreated(roomID string) error
	// 处理房间删除事件
	OnRoomDeleted(roomID string) error
	// 处理连接错误事件
	OnConnectionError(userID string, err error) error
}

// 配置接口
type ConfigProvider interface {
	// 获取WebRTC配置
	GetWebRTCConfig() *webrtc.Configuration
	// 获取SFU配置
	GetSFUConfig() *SFUConfig
	// 获取信令配置
	GetSignalingConfig() *SignalingConfig
}

// SFU配置
type SFUConfig struct {
	MaxRooms        int
	MaxUsersPerRoom int
	RoomTimeout     time.Duration
	UserTimeout     time.Duration
}

// 信令配置
type SignalingConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	CheckOrigin     bool
	BufferSize      int
	WorkerCount     int
}
