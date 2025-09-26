package types

import (
	"context"
	"time"

	"github.com/pion/webrtc/v3"
)

// 信令消息类型
type SignalingMessageType string

const (
	// 加入房间
	MessageTypeJoinRoom SignalingMessageType = "join_room"
	// 离开房间
	MessageTypeLeaveRoom SignalingMessageType = "leave_room"
	// 发送Offer
	MessageTypeOffer SignalingMessageType = "offer"
	// 发送Answer
	MessageTypeAnswer SignalingMessageType = "answer"
	// 发送ICE候选
	MessageTypeIceCandidate SignalingMessageType = "ice_candidate"
	// 房间信息
	MessageTypeRoomInfo SignalingMessageType = "room_info"
	// 用户加入
	MessageTypeUserJoined SignalingMessageType = "user_joined"
	// 用户离开
	MessageTypeUserLeft SignalingMessageType = "user_left"
	// 错误消息
	MessageTypeError SignalingMessageType = "error"
)

// 信令消息
type SignalingMessage struct {
	Type      SignalingMessageType `json:"type"`
	RoomID    string               `json:"room_id,omitempty"`
	UserID    string               `json:"user_id,omitempty"`
	Data      interface{}          `json:"data,omitempty"`
	Timestamp time.Time            `json:"timestamp"`
}

// WebRTC Offer/Answer 消息
type WebRTCMessage struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"` // "offer" or "answer"
}

// ICE 候选消息
type ICECandidateMessage struct {
	Candidate     string `json:"candidate"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
	SDPMid        string `json:"sdpMid"`
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
