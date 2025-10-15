package handler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/logic"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/streaming/room"
	"github.com/iceymoss/go-hichat-api/apps/streaming/sfu"
	"github.com/iceymoss/go-hichat-api/apps/streaming/webrtc"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/gorilla/websocket"
	libRTC "github.com/pion/webrtc/v3"
)

// SignalingServer 信令服务器结构体
// 负责处理WebSocket连接、消息路由和业务逻辑协调
type SignalingServer struct {
	svc          *svc.ServiceContext                 // 服务上下文，包含配置和依赖
	roomManager  *room.RoomManager                   // 房间管理器，负责房间的创建、删除和用户管理
	sfu          *sfu.SFU                            // 选择性转发单元，负责媒体流的转发
	connections  map[string]*webrtc.WebRTCConnection // WebRTC连接映射，key为用户ID
	mu           sync.RWMutex                        // 读写锁，保护connections的并发访问
	upgrader     websocket.Upgrader                  // WebSocket升级器
	messageQueue chan *SignalingMessage              // 消息处理队列
	workerCount  int                                 // 消息处理工作协程数量
	ctx          context.Context                     // 上下文，用于优雅关闭
	cancel       context.CancelFunc                  // 取消函数

	// 业务管理器
	callManager        *logic.CallManager        // 通话管理器，处理一对一和群组通话
	meetingManager     *logic.MeetingManager     // 会议管理器，处理会议相关功能
	screenShareManager *logic.ScreenShareManager // 录屏管理器，处理屏幕共享功能
	liveStreamManager  *logic.LiveStreamManager  // 直播管理器，处理直播相关功能
}

// SignalingMessage 信令消息包装器
// 将WebSocket连接和消息绑定，用于消息处理队列
type SignalingMessage struct {
	Conn    *websocket.Conn         // WebSocket连接
	Message *types.SignalingMessage // 解析后的信令消息
}

// NewSignalingServer 创建信令服务器
func NewSignalingServer(svc *svc.ServiceContext) *SignalingServer {
	ctx, cancel := context.WithCancel(context.Background())

	server := &SignalingServer{
		svc:         svc,
		roomManager: room.NewRoomManager(svc.Config.SFU.MaxRooms, time.Duration(svc.Config.SFU.RoomTimeout)*time.Second),
		sfu:         sfu.NewSFU(svc.Config.SFU.MaxRooms, svc.Config.SFU.MaxUsersPerRoom),
		connections: make(map[string]*webrtc.WebRTCConnection),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  svc.Config.Signaling.WebSocket.ReadBufferSize,
			WriteBufferSize: svc.Config.Signaling.WebSocket.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				// 允许所有来源，用于测试
				return true
			},
		},
		messageQueue: make(chan *SignalingMessage, svc.Config.Signaling.MessageQueue.BufferSize),
		workerCount:  svc.Config.Signaling.MessageQueue.WorkerCount,
		ctx:          ctx,
		cancel:       cancel,

		// 初始化管理器
		callManager:        logic.NewCallManager(),
		meetingManager:     logic.NewMeetingManager(),
		screenShareManager: logic.NewScreenShareManager(),
		liveStreamManager:  logic.NewLiveStreamManager(),
	}

	// 启动消息处理工作协程
	for i := 0; i < server.workerCount; i++ {
		go server.messageWorker(i)
	}

	return server
}

// HandleWebSocket 处理WebSocket连接
func (s *SignalingServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		zLog.Error("Failed to upgrade websocket connection", zap.Error(err))
		return
	}
	defer conn.Close()

	zLog.Info("WebSocket connection established",
		zap.String("remote_addr", r.RemoteAddr))

	// 处理WebSocket消息
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			var msg types.SignalingMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					zLog.Error("WebSocket read error", zap.Error(err))
				} else {
					zLog.Info("WebSocket connection closed normally")
				}
				return
			}

			// 记录接收到的消息
			zLog.Info("Received WebSocket message",
				zap.String("type", string(msg.Type)),
				zap.String("user_id", msg.UserID))

			// 将消息发送到处理队列
			s.messageQueue <- &SignalingMessage{
				Conn:    conn,
				Message: &msg,
			}
		}
	}
}

// messageWorker 消息处理工作协程
func (s *SignalingServer) messageWorker(workerID int) {
	zLog.Info("Signaling message worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-s.ctx.Done():
			zLog.Info("Signaling message worker stopped", zap.Int("worker_id", workerID))
			return
		case msg := <-s.messageQueue:
			s.handleMessage(msg.Conn, msg.Message)
		}
	}
}

// handleMessage 处理信令消息
func (s *SignalingServer) handleMessage(conn *websocket.Conn, msg *types.SignalingMessage) {
	zLog.Info("Handling signaling message",
		zap.String("type", string(msg.Type)),
		zap.String("user_id", msg.UserID),
		zap.String("room_id", msg.RoomID))

	switch msg.Type {
	// 基础通话
	case types.MessageTypeJoinRoom:
		s.handleJoinRoom(conn, msg)
	case types.MessageTypeLeaveRoom:
		s.handleLeaveRoom(conn, msg)
	case types.MessageTypeOffer:
		s.handleOffer(conn, msg)
	case types.MessageTypeAnswer:
		s.handleAnswer(conn, msg)
	case types.MessageTypeIceCandidate:
		s.handleIceCandidate(conn, msg)

	// 一对一通话
	case types.MessageTypeCallInvite:
		s.handleCallInvite(conn, msg)
	case types.MessageTypeCallAccept:
		s.handleCallAccept(conn, msg)
	case types.MessageTypeCallReject:
		s.handleCallReject(conn, msg)
	case types.MessageTypeCallEnd:
		s.handleCallEnd(conn, msg)

	// 群组通话
	case types.MessageTypeGroupInvite:
		s.handleGroupInvite(conn, msg)
	case types.MessageTypeGroupJoin:
		s.handleGroupJoin(conn, msg)
	case types.MessageTypeGroupLeave:
		s.handleGroupLeave(conn, msg)

	// 会议功能
	case types.MessageTypeMeetingCreate:
		s.handleMeetingCreate(conn, msg)
	case types.MessageTypeMeetingJoin:
		s.handleMeetingJoin(conn, msg)
	case types.MessageTypeMeetingLeave:
		s.handleMeetingLeave(conn, msg)
	case types.MessageTypeMeetingControl:
		s.handleMeetingControl(conn, msg)

	// 录屏功能
	case types.MessageTypeScreenShareStart:
		s.handleScreenShareStart(conn, msg)
	case types.MessageTypeScreenShareStop:
		s.handleScreenShareStop(conn, msg)
	case types.MessageTypeScreenShareRequest:
		s.handleScreenShareRequest(conn, msg)

	// 直播功能
	case types.MessageTypeLiveStart:
		s.handleLiveStart(conn, msg)
	case types.MessageTypeLiveStop:
		s.handleLiveStop(conn, msg)
	case types.MessageTypeLiveJoin:
		s.handleLiveJoin(conn, msg)
	case types.MessageTypeLiveLeave:
		s.handleLiveLeave(conn, msg)

	// 媒体控制
	case types.MessageTypeMute:
		s.handleMute(conn, msg)
	case types.MessageTypeUnmute:
		s.handleUnmute(conn, msg)
	case types.MessageTypeVideoOn:
		s.handleVideoOn(conn, msg)
	case types.MessageTypeVideoOff:
		s.handleVideoOff(conn, msg)

	default:
		s.sendError(conn, msg.UserID, fmt.Sprintf("unknown message type: %s", msg.Type))
	}
}

// handleJoinRoom 处理加入房间
// handleJoinRoom 处理用户加入房间的消息
func (s *SignalingServer) handleJoinRoom(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID
	roomID := msg.RoomID

	zLog.Info("🔵 [Step 1] 处理加入房间请求",
		zap.String("user_id", userID),
		zap.String("room_id", roomID))

	if userID == "" || roomID == "" {
		zLog.Warn("加入房间请求参数不完整",
			zap.String("user_id", userID),
			zap.String("room_id", roomID))
		s.sendError(conn, userID, "user_id and room_id are required")
		return
	}

	zLog.Info("🔵 [Step 2] 开始创建/获取房间")

	// 创建或获取房间
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		// 房间不存在，创建新房间
		zLog.Info("房间不存在，创建新房间",
			zap.String("room_id", roomID))
		room, err = s.roomManager.CreateRoom(roomID, fmt.Sprintf("Room %s", roomID))
		if err != nil {
			zLog.Error("创建房间失败",
				zap.String("room_id", roomID),
				zap.Error(err))
			s.sendError(conn, userID, fmt.Sprintf("failed to create room: %v", err))
			return
		}
		zLog.Info("✅ 房间创建成功",
			zap.String("room_id", roomID))
	} else {
		zLog.Info("✅ 获取到现有房间",
			zap.String("room_id", roomID))
	}

	zLog.Info("🔵 [Step 3] 创建用户对象")

	// 创建用户
	user := &types.User{
		UserID:    userID,
		Username:  fmt.Sprintf("User %s", userID),
		JoinedAt:  time.Now(),
		IsMuted:   false,
		IsVideoOn: true,
	}

	zLog.Info("🔵 [Step 4] 添加用户到房间")

	// 添加用户到房间
	if err := room.AddUser(user); err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to join room: %v", err))
		return
	}

	zLog.Info("✅ 用户已添加到房间")

	zLog.Info("🔵 [Step 5] 创建WebRTC连接")

	// 创建WebRTC连接
	webrtcConfig := s.getWebRTCConfig()
	webrtcConn, err := webrtc.NewWebRTCConnection(userID, &WebSocketConnection{conn: conn}, webrtcConfig)
	if err != nil {
		zLog.Error("❌ 创建WebRTC连接失败", zap.Error(err))
		s.sendError(conn, userID, fmt.Sprintf("failed to create WebRTC connection: %v", err))
		return
	}

	zLog.Info("✅ WebRTC连接创建成功")

	zLog.Info("🔵 [Step 6] 设置媒体流回调")

	// 🔥 设置媒体流接收回调 - 当收到用户的媒体流时转发给房间内其他用户
	webrtcConn.SetOnTrackHandler(func(track *libRTC.TrackLocalStaticRTP) {
		zLog.Info("收到用户媒体流，开始转发",
			zap.String("from_user", userID),
			zap.String("room_id", roomID),
			zap.String("track_id", track.ID()),
			zap.String("kind", track.Kind().String()))

		// 转发给房间内的其他用户
		s.forwardTrackToRoom(roomID, userID, track)
	})

	zLog.Info("✅ 媒体流回调设置完成")

	zLog.Info("🔵 [Step 7] 保存连接")

	// 保存连接
	s.mu.Lock()
	s.connections[userID] = webrtcConn
	s.mu.Unlock()

	zLog.Info("✅ 连接已保存")

	zLog.Info("🔵 [Step 8] 添加用户到SFU")

	// 添加用户到SFU
	if err := s.sfu.AddUserToRoom(roomID, userID, webrtcConn.GetPeerConnection()); err != nil {
		zLog.Error("❌ 添加用户到SFU失败", zap.Error(err))
		s.sendError(conn, userID, fmt.Sprintf("failed to add user to SFU: %v", err))
		return
	}

	zLog.Info("✅ 用户已添加到SFU")

	zLog.Info("🔵 [Step 9] 同步已有媒体流")

	// 🔥 将房间内已有用户的媒体流添加到新用户
	s.addExistingTracksToNewUser(roomID, userID, webrtcConn)

	zLog.Info("✅ 已有媒体流同步完成")

	zLog.Info("🔵 [Step 10] 发送房间信息")

	// 发送房间信息
	roomInfo := room.GetRoomInfo()
	s.sendMessage(conn, &types.SignalingMessage{
		Type:      types.MessageTypeRoomInfo,
		RoomID:    roomID,
		UserID:    userID,
		Data:      roomInfo,
		Timestamp: time.Now(),
	})

	zLog.Info("✅ 房间信息已发送")

	zLog.Info("🔵 [Step 11] 通知其他用户")

	// 通知房间内其他用户
	s.broadcastToRoom(roomID, userID, &types.SignalingMessage{
		Type:      types.MessageTypeUserJoined,
		RoomID:    roomID,
		UserID:    userID,
		Data:      user,
		Timestamp: time.Now(),
	})

	zLog.Info("✅✅✅ User joined room successfully",
		zap.String("user_id", userID),
		zap.String("room_id", roomID),
		zap.Int("room_user_count", room.GetUserCount()))
}

// handleLeaveRoom 处理离开房间
func (s *SignalingServer) handleLeaveRoom(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID
	roomID := msg.RoomID

	// 从房间移除用户
	if err := s.roomManager.LeaveRoom(roomID, userID); err != nil {
		zLog.Error("Failed to leave room",
			zap.String("user_id", userID),
			zap.String("room_id", roomID),
			zap.Error(err))
	}

	// 从SFU移除用户
	if err := s.sfu.RemoveUserFromRoom(roomID, userID); err != nil {
		zLog.Error("Failed to remove user from SFU",
			zap.String("user_id", userID),
			zap.String("room_id", roomID),
			zap.Error(err))
	}

	// 关闭WebRTC连接
	s.mu.Lock()
	if webrtcConn, exists := s.connections[userID]; exists {
		webrtcConn.Close()
		delete(s.connections, userID)
	}
	s.mu.Unlock()

	// 通知房间内其他用户
	s.broadcastToRoom(roomID, userID, &types.SignalingMessage{
		Type:      types.MessageTypeUserLeft,
		RoomID:    roomID,
		UserID:    userID,
		Timestamp: time.Now(),
	})

	zLog.Info("User left room",
		zap.String("user_id", userID),
		zap.String("room_id", roomID))
}

// handleOffer 处理Offer
func (s *SignalingServer) handleOffer(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID
	roomID := msg.RoomID

	// 获取WebRTC连接
	s.mu.RLock()
	webrtcConn, exists := s.connections[userID]
	s.mu.RUnlock()

	if !exists {
		s.sendError(conn, userID, "WebRTC connection not found")
		return
	}

	// 解析Offer数据
	offerData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid offer data")
		return
	}

	sdp, ok := offerData["sdp"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid SDP in offer")
		return
	}

	// 设置远程描述
	offer := libRTC.SessionDescription{
		Type: libRTC.SDPTypeOffer,
		SDP:  sdp,
	}

	if err := webrtcConn.SetRemoteDescription(offer); err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to set remote description: %v", err))
		return
	}

	// 创建Answer
	answer, err := webrtcConn.CreateAnswer()
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to create answer: %v", err))
		return
	}

	// 发送Answer
	s.sendMessage(conn, &types.SignalingMessage{
		Type:   types.MessageTypeAnswer,
		RoomID: roomID,
		UserID: userID,
		Data: types.WebRTCMessage{
			SDP:  answer.SDP,
			Type: "answer",
		},
		Timestamp: time.Now(),
	})

	zLog.Debug("Offer handled and answer sent",
		zap.String("user_id", userID),
		zap.String("room_id", roomID))
}

// handleAnswer 处理Answer
func (s *SignalingServer) handleAnswer(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID
	roomID := msg.RoomID

	// 获取WebRTC连接
	s.mu.RLock()
	webrtcConn, exists := s.connections[userID]
	s.mu.RUnlock()

	if !exists {
		s.sendError(conn, userID, "WebRTC connection not found")
		return
	}

	// 解析Answer数据
	answerData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid answer data")
		return
	}

	sdp, ok := answerData["sdp"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid SDP in answer")
		return
	}

	// 设置远程描述
	answer := libRTC.SessionDescription{
		Type: libRTC.SDPTypeAnswer,
		SDP:  sdp,
	}

	if err := webrtcConn.SetRemoteDescription(answer); err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to set remote description: %v", err))
		return
	}

	zLog.Debug("Answer handled",
		zap.String("user_id", userID),
		zap.String("room_id", roomID))
}

// handleIceCandidate 处理ICE候选
func (s *SignalingServer) handleIceCandidate(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID
	roomID := msg.RoomID

	// 获取WebRTC连接
	s.mu.RLock()
	webrtcConn, exists := s.connections[userID]
	s.mu.RUnlock()

	if !exists {
		s.sendError(conn, userID, "WebRTC connection not found")
		return
	}

	// 解析ICE候选数据
	candidateData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid ICE candidate data")
		return
	}

	candidate, ok := candidateData["candidate"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid candidate in ICE candidate")
		return
	}

	sdpMLineIndex, ok := candidateData["sdpMLineIndex"].(float64)
	if !ok {
		s.sendError(conn, userID, "invalid sdpMLineIndex in ICE candidate")
		return
	}

	sdpMid, ok := candidateData["sdpMid"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid sdpMid in ICE candidate")
		return
	}

	linde := uint16(sdpMLineIndex)

	// 添加ICE候选
	iceCandidate := libRTC.ICECandidateInit{
		Candidate:     candidate,
		SDPMLineIndex: &linde,
		SDPMid:        &sdpMid,
	}

	// 🔍 详细日志：收到客户端的ICE候选
	zLog.Info("🧊 收到客户端ICE候选",
		zap.String("user_id", userID),
		zap.String("candidate", candidate))

	if err := webrtcConn.AddICECandidate(iceCandidate); err != nil {
		zLog.Error("❌ 添加ICE候选失败",
			zap.String("user_id", userID),
			zap.Error(err))
		s.sendError(conn, userID, fmt.Sprintf("failed to add ICE candidate: %v", err))
		return
	}

	zLog.Info("✅ ICE候选已添加",
		zap.String("user_id", userID),
		zap.String("room_id", roomID))
}

// sendMessage 发送消息
func (s *SignalingServer) sendMessage(conn *websocket.Conn, msg *types.SignalingMessage) {
	if err := conn.WriteJSON(msg); err != nil {
		zLog.Error("Failed to send message",
			zap.String("user_id", msg.UserID),
			zap.String("type", string(msg.Type)),
			zap.Error(err))
	}
}

// sendError 发送错误消息
func (s *SignalingServer) sendError(conn *websocket.Conn, userID, errorMsg string) {
	s.sendMessage(conn, &types.SignalingMessage{
		Type:      types.MessageTypeError,
		UserID:    userID,
		Data:      map[string]string{"error": errorMsg},
		Timestamp: time.Now(),
	})
}

// broadcastToRoom 向房间广播消息
func (s *SignalingServer) broadcastToRoom(roomID, excludeUserID string, msg *types.SignalingMessage) {
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		zLog.Error("Failed to get room for broadcast",
			zap.String("room_id", roomID),
			zap.Error(err))
		return
	}

	users := room.GetUsers()
	for _, user := range users {
		if user.UserID != excludeUserID {
			s.mu.RLock()
			if webrtcConn, exists := s.connections[user.UserID]; exists {
				webrtcConn.SendMessage(msg)
			}
			s.mu.RUnlock()
		}
	}
}

// getWebRTCConfig 获取WebRTC配置
func (s *SignalingServer) getWebRTCConfig() *libRTC.Configuration {
	config := &libRTC.Configuration{
		ICEServers: []libRTC.ICEServer{},
	}

	// 添加ICE服务器
	for _, iceServer := range s.svc.Config.WebRTC.IceServers {
		server := libRTC.ICEServer{
			URLs: iceServer.URLs,
		}

		if iceServer.Username != "" {
			server.Username = iceServer.Username
		}

		if iceServer.Credential != "" {
			server.Credential = iceServer.Credential
		}

		config.ICEServers = append(config.ICEServers, server)
	}

	// 🔥 关键：配置 ICE 传输策略
	// 在 localhost 测试时使用 all 策略，允许所有类型的候选
	config.ICETransportPolicy = libRTC.ICETransportPolicyAll

	zLog.Info("WebRTC 配置",
		zap.Int("ice_servers", len(config.ICEServers)),
		zap.String("ice_policy", string(config.ICETransportPolicy)))

	return config
}

// Close 关闭信令服务器
func (s *SignalingServer) Close() error {
	s.cancel()

	// 关闭所有WebRTC连接
	s.mu.Lock()
	for userID, conn := range s.connections {
		conn.Close()
		delete(s.connections, userID)
	}
	s.mu.Unlock()

	// 关闭SFU
	if err := s.sfu.Close(); err != nil {
		zLog.Error("Failed to close SFU", zap.Error(err))
	}

	zLog.Info("Signaling server closed")
	return nil
}

// WebSocketConnection WebSocket连接包装器
type WebSocketConnection struct {
	conn   *websocket.Conn
	userID string
	mu     sync.Mutex // 🔥 添加写锁，防止并发写入
}

func (w *WebSocketConnection) SendMessage(message *types.SignalingMessage) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteJSON(message)
}

func (w *WebSocketConnection) ReceiveMessage() (*types.SignalingMessage, error) {
	var msg types.SignalingMessage
	err := w.conn.ReadJSON(&msg)
	return &msg, err
}

func (w *WebSocketConnection) Close() error {
	return w.conn.Close()
}

func (w *WebSocketConnection) GetUserID() string {
	return w.userID
}

func (w *WebSocketConnection) SetUserID(userID string) {
	w.userID = userID
}

func (w *WebSocketConnection) IsConnected() bool {
	return w.conn != nil
}

// ========== 一对一通话处理 ==========

// handleCallInvite 处理通话邀请
// handleCallInvite 处理一对一通话邀请消息
func (s *SignalingServer) handleCallInvite(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	zLog.Info("处理一对一通话邀请",
		zap.String("caller_id", userID))

	// 解析邀请数据
	inviteData, ok := msg.Data.(map[string]interface{})
	if !ok {
		zLog.Warn("通话邀请数据格式错误",
			zap.String("caller_id", userID),
			zap.Any("data", msg.Data))
		s.sendError(conn, userID, "invalid invite data")
		return
	}

	calleeID, ok := inviteData["callee_id"].(string)
	if !ok {
		zLog.Warn("通话邀请缺少被叫用户ID",
			zap.String("caller_id", userID),
			zap.Any("invite_data", inviteData))
		s.sendError(conn, userID, "invalid callee_id")
		return
	}

	zLog.Info("解析通话邀请数据成功",
		zap.String("caller_id", userID),
		zap.String("callee_id", calleeID))

	// 创建一对一通话
	call, err := s.callManager.CreateOneToOneCall(userID, calleeID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to create call: %v", err))
		return
	}

	// 发送邀请给被叫方
	s.sendMessageToUser(calleeID, &types.SignalingMessage{
		Type:      types.MessageTypeCallInvite,
		UserID:    userID,
		Data:      call,
		Timestamp: time.Now(),
	})

	zLog.Info("Call invite sent",
		zap.String("caller_id", userID),
		zap.String("callee_id", calleeID),
		zap.String("call_id", call.ID))
}

// handleCallAccept 处理通话接受
func (s *SignalingServer) handleCallAccept(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析接受数据
	acceptData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid accept data")
		return
	}

	callID, ok := acceptData["call_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid call_id")
		return
	}

	// 接受通话
	err := s.callManager.AcceptCall(callID, userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to accept call: %v", err))
		return
	}

	// 获取通话信息
	call, err := s.callManager.GetCall(callID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to get call: %v", err))
		return
	}

	// 通知所有参与者
	for _, participantID := range call.Participants {
		s.sendMessageToUser(participantID, &types.SignalingMessage{
			Type:      types.MessageTypeCallAccept,
			UserID:    userID,
			Data:      call,
			Timestamp: time.Now(),
		})
	}

	zLog.Info("Call accepted",
		zap.String("user_id", userID),
		zap.String("call_id", callID))
}

// handleCallReject 处理通话拒绝
func (s *SignalingServer) handleCallReject(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析拒绝数据
	rejectData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid reject data")
		return
	}

	callID, ok := rejectData["call_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid call_id")
		return
	}

	// 获取通话信息
	call, err := s.callManager.GetCall(callID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to get call: %v", err))
		return
	}

	// 更新通话状态
	call.Status = types.CallStatusEnded
	now := time.Now()
	call.EndedAt = &now
	if call.StartedAt != nil {
		call.Duration = int64(now.Sub(*call.StartedAt).Seconds())
	}

	// 拒绝通话
	err = s.callManager.RejectCall(callID, userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to reject call: %v", err))
		return
	}

	// 通知所有参与者
	for _, participantID := range call.Participants {
		s.sendMessageToUser(participantID, &types.SignalingMessage{
			Type:      types.MessageTypeCallReject,
			UserID:    userID,
			Data:      call,
			Timestamp: time.Now(),
		})
	}

	zLog.Info("Call rejected",
		zap.String("user_id", userID),
		zap.String("call_id", callID))
}

// handleCallEnd 处理通话结束
func (s *SignalingServer) handleCallEnd(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析结束数据
	endData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid end data")
		return
	}

	callID, ok := endData["call_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid call_id")
		return
	}

	// 结束通话
	err := s.callManager.EndCall(callID, userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to end call: %v", err))
		return
	}

	// 获取通话信息
	call, err := s.callManager.GetCall(callID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to get call: %v", err))
		return
	}

	// 通知所有参与者
	for _, participantID := range call.Participants {
		s.sendMessageToUser(participantID, &types.SignalingMessage{
			Type:      types.MessageTypeCallEnd,
			UserID:    userID,
			Data:      call,
			Timestamp: time.Now(),
		})
	}

	zLog.Info("Call ended",
		zap.String("user_id", userID),
		zap.String("call_id", callID))
}

// ========== 群组通话处理 ==========

// handleGroupInvite 处理群组邀请
func (s *SignalingServer) handleGroupInvite(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析邀请数据
	inviteData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid invite data")
		return
	}

	participants, ok := inviteData["participants"].([]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid participants")
		return
	}

	// 转换参与者列表
	participantList := make([]string, len(participants))
	for i, p := range participants {
		if participantID, ok := p.(string); ok {
			participantList[i] = participantID
		} else {
			s.sendError(conn, userID, "invalid participant ID")
			return
		}
	}

	// 创建群组通话
	call, err := s.callManager.CreateGroupCall(userID, participantList)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to create group call: %v", err))
		return
	}

	// 发送邀请给所有参与者
	for _, participantID := range participantList {
		s.sendMessageToUser(participantID, &types.SignalingMessage{
			Type:      types.MessageTypeGroupInvite,
			UserID:    userID,
			Data:      call,
			Timestamp: time.Now(),
		})
	}

	zLog.Info("Group call invite sent",
		zap.String("caller_id", userID),
		zap.Int("participant_count", len(participantList)),
		zap.String("call_id", call.ID))
}

// handleGroupJoin 处理群组加入
func (s *SignalingServer) handleGroupJoin(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析加入数据
	joinData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid join data")
		return
	}

	callID, ok := joinData["call_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid call_id")
		return
	}

	// 接受通话
	err := s.callManager.AcceptCall(callID, userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to join group call: %v", err))
		return
	}

	// 获取通话信息
	call, err := s.callManager.GetCall(callID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to get call: %v", err))
		return
	}

	// 通知所有参与者
	for _, participantID := range call.Participants {
		s.sendMessageToUser(participantID, &types.SignalingMessage{
			Type:      types.MessageTypeGroupJoin,
			UserID:    userID,
			Data:      call,
			Timestamp: time.Now(),
		})
	}

	zLog.Info("User joined group call",
		zap.String("user_id", userID),
		zap.String("call_id", callID))
}

// handleGroupLeave 处理群组离开
func (s *SignalingServer) handleGroupLeave(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析离开数据
	leaveData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid leave data")
		return
	}

	callID, ok := leaveData["call_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid call_id")
		return
	}

	// 结束通话
	err := s.callManager.EndCall(callID, userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to leave group call: %v", err))
		return
	}

	// 获取通话信息
	call, err := s.callManager.GetCall(callID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to get call: %v", err))
		return
	}

	// 通知所有参与者
	for _, participantID := range call.Participants {
		s.sendMessageToUser(participantID, &types.SignalingMessage{
			Type:      types.MessageTypeGroupLeave,
			UserID:    userID,
			Data:      call,
			Timestamp: time.Now(),
		})
	}

	zLog.Info("User left group call",
		zap.String("user_id", userID),
		zap.String("call_id", callID))
}

// ========== 会议功能处理 ==========

// handleMeetingCreate 处理会议创建
func (s *SignalingServer) handleMeetingCreate(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析创建数据
	createData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid create data")
		return
	}

	title, ok := createData["title"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid title")
		return
	}

	description, _ := createData["description"].(string)

	// 解析会议设置
	var settings *types.MeetingSettings
	if settingsData, exists := createData["settings"]; exists {
		if settingsMap, ok := settingsData.(map[string]interface{}); ok {
			settings = &types.MeetingSettings{
				MaxParticipants:  50,
				AllowScreenShare: true,
				AllowRecording:   true,
				MuteOnJoin:       false,
				VideoOnJoin:      true,
				WaitingRoom:      false,
			}

			if maxParticipants, ok := settingsMap["max_participants"].(float64); ok {
				settings.MaxParticipants = int(maxParticipants)
			}
			if allowScreenShare, ok := settingsMap["allow_screen_share"].(bool); ok {
				settings.AllowScreenShare = allowScreenShare
			}
			if allowRecording, ok := settingsMap["allow_recording"].(bool); ok {
				settings.AllowRecording = allowRecording
			}
			if muteOnJoin, ok := settingsMap["mute_on_join"].(bool); ok {
				settings.MuteOnJoin = muteOnJoin
			}
			if videoOnJoin, ok := settingsMap["video_on_join"].(bool); ok {
				settings.VideoOnJoin = videoOnJoin
			}
			if waitingRoom, ok := settingsMap["waiting_room"].(bool); ok {
				settings.WaitingRoom = waitingRoom
			}
		}
	}

	// 创建会议
	meeting, err := s.meetingManager.CreateMeeting(userID, title, description, settings)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to create meeting: %v", err))
		return
	}

	// 发送会议信息
	s.sendMessage(conn, &types.SignalingMessage{
		Type:      types.MessageTypeMeetingCreate,
		UserID:    userID,
		Data:      meeting,
		Timestamp: time.Now(),
	})

	zLog.Info("Meeting created",
		zap.String("user_id", userID),
		zap.String("meeting_id", meeting.ID),
		zap.String("title", title))
}

// handleMeetingJoin 处理会议加入
func (s *SignalingServer) handleMeetingJoin(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析加入数据
	joinData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid join data")
		return
	}

	meetingID, ok := joinData["meeting_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid meeting_id")
		return
	}

	username, _ := joinData["username"].(string)
	if username == "" {
		username = fmt.Sprintf("User %s", userID)
	}

	// 加入会议
	meeting, err := s.meetingManager.JoinMeeting(meetingID, userID, username)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to join meeting: %v", err))
		return
	}

	// 通知所有参与者
	for _, participant := range meeting.Participants {
		s.sendMessageToUser(participant.UserID, &types.SignalingMessage{
			Type:      types.MessageTypeMeetingJoin,
			UserID:    userID,
			Data:      meeting,
			Timestamp: time.Now(),
		})
	}

	zLog.Info("User joined meeting",
		zap.String("user_id", userID),
		zap.String("meeting_id", meetingID),
		zap.String("username", username))
}

// handleMeetingLeave 处理会议离开
func (s *SignalingServer) handleMeetingLeave(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析离开数据
	leaveData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid leave data")
		return
	}

	meetingID, ok := leaveData["meeting_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid meeting_id")
		return
	}

	// 离开会议
	err := s.meetingManager.LeaveMeeting(meetingID, userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to leave meeting: %v", err))
		return
	}

	// 获取会议信息
	meeting, err := s.meetingManager.GetMeeting(meetingID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to get meeting: %v", err))
		return
	}

	// 通知所有参与者
	for _, participant := range meeting.Participants {
		s.sendMessageToUser(participant.UserID, &types.SignalingMessage{
			Type:      types.MessageTypeMeetingLeave,
			UserID:    userID,
			Data:      meeting,
			Timestamp: time.Now(),
		})
	}

	zLog.Info("User left meeting",
		zap.String("user_id", userID),
		zap.String("meeting_id", meetingID))
}

// handleMeetingControl 处理会议控制
func (s *SignalingServer) handleMeetingControl(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析控制数据
	controlData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid control data")
		return
	}

	meetingID, ok := controlData["meeting_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid meeting_id")
		return
	}

	action, ok := controlData["action"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid action")
		return
	}

	// 处理不同的控制动作
	switch action {
	case "mute_participant":
		participantID, ok := controlData["participant_id"].(string)
		if !ok {
			s.sendError(conn, userID, "invalid participant_id")
			return
		}

		err := s.meetingManager.MuteParticipant(meetingID, userID, participantID)
		if err != nil {
			s.sendError(conn, userID, fmt.Sprintf("failed to mute participant: %v", err))
			return
		}

	case "unmute_participant":
		participantID, ok := controlData["participant_id"].(string)
		if !ok {
			s.sendError(conn, userID, "invalid participant_id")
			return
		}

		err := s.meetingManager.UnmuteParticipant(meetingID, userID, participantID)
		if err != nil {
			s.sendError(conn, userID, fmt.Sprintf("failed to unmute participant: %v", err))
			return
		}

	case "end_meeting":
		err := s.meetingManager.EndMeeting(meetingID, userID)
		if err != nil {
			s.sendError(conn, userID, fmt.Sprintf("failed to end meeting: %v", err))
			return
		}

	default:
		s.sendError(conn, userID, fmt.Sprintf("unknown action: %s", action))
		return
	}

	// 获取会议信息
	meeting, err := s.meetingManager.GetMeeting(meetingID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to get meeting: %v", err))
		return
	}

	// 通知所有参与者
	for _, participant := range meeting.Participants {
		s.sendMessageToUser(participant.UserID, &types.SignalingMessage{
			Type:   types.MessageTypeMeetingControl,
			UserID: userID,
			Data: map[string]interface{}{
				"meeting": meeting,
				"action":  action,
			},
			Timestamp: time.Now(),
		})
	}

	zLog.Info("Meeting control action executed",
		zap.String("user_id", userID),
		zap.String("meeting_id", meetingID),
		zap.String("action", action))
}

// ========== 录屏功能处理 ==========

// handleScreenShareStart 处理录屏开始
func (s *SignalingServer) handleScreenShareStart(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析开始数据
	startData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid start data")
		return
	}

	roomID, ok := startData["room_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid room_id")
		return
	}

	quality, _ := startData["quality"].(string)
	if quality == "" {
		quality = "medium"
	}

	// 开始录屏
	screenShare, err := s.screenShareManager.StartScreenShare(userID, roomID, quality)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to start screen share: %v", err))
		return
	}

	// 通知房间内其他用户
	s.broadcastToRoom(roomID, userID, &types.SignalingMessage{
		Type:      types.MessageTypeScreenShareStart,
		RoomID:    roomID,
		UserID:    userID,
		Data:      screenShare,
		Timestamp: time.Now(),
	})

	zLog.Info("Screen share started",
		zap.String("user_id", userID),
		zap.String("room_id", roomID),
		zap.String("quality", quality))
}

// handleScreenShareStop 处理录屏停止
func (s *SignalingServer) handleScreenShareStop(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 停止录屏
	err := s.screenShareManager.StopScreenShare(userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to stop screen share: %v", err))
		return
	}

	// 获取录屏信息
	screenShare, err := s.screenShareManager.GetUserScreenShare(userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to get screen share: %v", err))
		return
	}

	// 通知房间内其他用户
	s.broadcastToRoom(screenShare.RoomID, userID, &types.SignalingMessage{
		Type:      types.MessageTypeScreenShareStop,
		RoomID:    screenShare.RoomID,
		UserID:    userID,
		Data:      screenShare,
		Timestamp: time.Now(),
	})

	zLog.Info("Screen share stopped",
		zap.String("user_id", userID),
		zap.String("room_id", screenShare.RoomID))
}

// handleScreenShareRequest 处理录屏请求
func (s *SignalingServer) handleScreenShareRequest(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析请求数据
	requestData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid request data")
		return
	}

	targetUserID, ok := requestData["target_user_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid target_user_id")
		return
	}

	roomID, ok := requestData["room_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid room_id")
		return
	}

	// 请求录屏
	err := s.screenShareManager.RequestScreenShare(userID, targetUserID, roomID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to request screen share: %v", err))
		return
	}

	// 发送请求给目标用户
	s.sendMessageToUser(targetUserID, &types.SignalingMessage{
		Type:   types.MessageTypeScreenShareRequest,
		UserID: userID,
		RoomID: roomID,
		Data: map[string]interface{}{
			"requester_id": userID,
			"room_id":      roomID,
		},
		Timestamp: time.Now(),
	})

	zLog.Info("Screen share requested",
		zap.String("requester_id", userID),
		zap.String("target_user_id", targetUserID),
		zap.String("room_id", roomID))
}

// ========== 直播功能处理 ==========

// handleLiveStart 处理直播开始
func (s *SignalingServer) handleLiveStart(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析开始数据
	startData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid start data")
		return
	}

	title, ok := startData["title"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid title")
		return
	}

	description, _ := startData["description"].(string)

	// 开始直播
	liveStream, err := s.liveStreamManager.StartLiveStream(userID, title, description)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to start live stream: %v", err))
		return
	}

	// 发送直播信息
	s.sendMessage(conn, &types.SignalingMessage{
		Type:      types.MessageTypeLiveStart,
		UserID:    userID,
		Data:      liveStream,
		Timestamp: time.Now(),
	})

	zLog.Info("Live stream started",
		zap.String("user_id", userID),
		zap.String("stream_id", liveStream.ID),
		zap.String("title", title))
}

// handleLiveStop 处理直播停止
func (s *SignalingServer) handleLiveStop(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 停止直播
	err := s.liveStreamManager.StopLiveStream(userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to stop live stream: %v", err))
		return
	}

	// 获取直播信息
	liveStream, err := s.liveStreamManager.GetUserLiveStream(userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to get live stream: %v", err))
		return
	}

	// 通知所有观众
	viewers, err := s.liveStreamManager.GetStreamViewers(liveStream.ID)
	if err == nil {
		for _, viewerID := range viewers {
			s.sendMessageToUser(viewerID, &types.SignalingMessage{
				Type:      types.MessageTypeLiveStop,
				UserID:    userID,
				Data:      liveStream,
				Timestamp: time.Now(),
			})
		}
	}

	zLog.Info("Live stream stopped",
		zap.String("user_id", userID),
		zap.String("stream_id", liveStream.ID))
}

// handleLiveJoin 处理直播加入
func (s *SignalingServer) handleLiveJoin(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析加入数据
	joinData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid join data")
		return
	}

	streamID, ok := joinData["stream_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid stream_id")
		return
	}

	// 加入直播
	liveStream, err := s.liveStreamManager.JoinLiveStream(streamID, userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to join live stream: %v", err))
		return
	}

	// 发送直播信息
	s.sendMessage(conn, &types.SignalingMessage{
		Type:      types.MessageTypeLiveJoin,
		UserID:    userID,
		Data:      liveStream,
		Timestamp: time.Now(),
	})

	zLog.Info("User joined live stream",
		zap.String("user_id", userID),
		zap.String("stream_id", streamID),
		zap.Int("viewer_count", liveStream.ViewerCount))
}

// handleLiveLeave 处理直播离开
func (s *SignalingServer) handleLiveLeave(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析离开数据
	leaveData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid leave data")
		return
	}

	streamID, ok := leaveData["stream_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid stream_id")
		return
	}

	// 离开直播
	err := s.liveStreamManager.LeaveLiveStream(streamID, userID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to leave live stream: %v", err))
		return
	}

	// 获取直播信息
	liveStream, err := s.liveStreamManager.GetLiveStream(streamID)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to get live stream: %v", err))
		return
	}

	zLog.Info("User left live stream",
		zap.String("user_id", userID),
		zap.String("stream_id", streamID),
		zap.Int("remaining_viewer_count", liveStream.ViewerCount))
}

// ========== 媒体控制处理 ==========

// handleMute 处理静音
func (s *SignalingServer) handleMute(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析静音数据
	muteData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid mute data")
		return
	}

	roomID, ok := muteData["room_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid room_id")
		return
	}

	// 通知房间内其他用户
	s.broadcastToRoom(roomID, userID, &types.SignalingMessage{
		Type:      types.MessageTypeMute,
		RoomID:    roomID,
		UserID:    userID,
		Timestamp: time.Now(),
	})

	zLog.Info("User muted",
		zap.String("user_id", userID),
		zap.String("room_id", roomID))
}

// handleUnmute 处理取消静音
func (s *SignalingServer) handleUnmute(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析取消静音数据
	unmuteData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid unmute data")
		return
	}

	roomID, ok := unmuteData["room_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid room_id")
		return
	}

	// 通知房间内其他用户
	s.broadcastToRoom(roomID, userID, &types.SignalingMessage{
		Type:      types.MessageTypeUnmute,
		RoomID:    roomID,
		UserID:    userID,
		Timestamp: time.Now(),
	})

	zLog.Info("User unmuted",
		zap.String("user_id", userID),
		zap.String("room_id", roomID))
}

// handleVideoOn 处理开启视频
func (s *SignalingServer) handleVideoOn(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析开启视频数据
	videoOnData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid video_on data")
		return
	}

	roomID, ok := videoOnData["room_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid room_id")
		return
	}

	// 通知房间内其他用户
	s.broadcastToRoom(roomID, userID, &types.SignalingMessage{
		Type:      types.MessageTypeVideoOn,
		RoomID:    roomID,
		UserID:    userID,
		Timestamp: time.Now(),
	})

	zLog.Info("User video on",
		zap.String("user_id", userID),
		zap.String("room_id", roomID))
}

// handleVideoOff 处理关闭视频
func (s *SignalingServer) handleVideoOff(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID

	// 解析关闭视频数据
	videoOffData, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, userID, "invalid video_off data")
		return
	}

	roomID, ok := videoOffData["room_id"].(string)
	if !ok {
		s.sendError(conn, userID, "invalid room_id")
		return
	}

	// 通知房间内其他用户
	s.broadcastToRoom(roomID, userID, &types.SignalingMessage{
		Type:      types.MessageTypeVideoOff,
		RoomID:    roomID,
		UserID:    userID,
		Timestamp: time.Now(),
	})

	zLog.Info("User video off",
		zap.String("user_id", userID),
		zap.String("room_id", roomID))
}

// sendMessageToUser 发送消息给特定用户
func (s *SignalingServer) sendMessageToUser(userID string, msg *types.SignalingMessage) {
	s.mu.RLock()
	if webrtcConn, exists := s.connections[userID]; exists {
		webrtcConn.SendMessage(msg)
	}
	s.mu.RUnlock()
}

// 🔥 核心方法：转发媒体轨道到房间内其他用户
func (s *SignalingServer) forwardTrackToRoom(roomID, fromUserID string, track *libRTC.TrackLocalStaticRTP) {
	// 获取房间信息
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		zLog.Error("Failed to get room for forwarding track",
			zap.String("room_id", roomID),
			zap.Error(err))
		return
	}

	// 获取房间内的所有用户
	users := room.GetUsers()

	zLog.Info("开始转发媒体流",
		zap.String("from_user", fromUserID),
		zap.String("room_id", roomID),
		zap.String("track_id", track.ID()),
		zap.Int("target_users", len(users)-1))

	// 遍历房间内的其他用户
	for _, user := range users {
		if user.UserID == fromUserID {
			continue // 不转发给自己
		}

		// 获取目标用户的WebRTC连接
		s.mu.RLock()
		targetConn, exists := s.connections[user.UserID]
		s.mu.RUnlock()

		if !exists {
			zLog.Warn("目标用户连接不存在",
				zap.String("target_user", user.UserID))
			continue
		}

		// 🎯 关键：将track添加到目标用户的PeerConnection
		sender, err := targetConn.AddTrack(track)
		if err != nil {
			zLog.Error("Failed to add track to target user",
				zap.String("from_user", fromUserID),
				zap.String("to_user", user.UserID),
				zap.String("track_id", track.ID()),
				zap.Error(err))
			continue
		}

		zLog.Info("媒体流转发成功",
			zap.String("from_user", fromUserID),
			zap.String("to_user", user.UserID),
			zap.String("track_id", track.ID()),
			zap.String("kind", track.Kind().String()))

		// 🔄 重新协商：添加track后需要创建新的offer
		go s.renegotiateConnection(targetConn, user.UserID, sender)
	}
}

// 🔄 重新协商连接（添加track后）
func (s *SignalingServer) renegotiateConnection(conn *webrtc.WebRTCConnection, userID string, sender *libRTC.RTPSender) {
	// 创建新的Offer
	offer, err := conn.CreateOffer()
	if err != nil {
		zLog.Error("Failed to create offer for renegotiation",
			zap.String("user_id", userID),
			zap.Error(err))
		return
	}

	// 发送新的Offer给用户
	s.mu.RLock()
	if webrtcConn, exists := s.connections[userID]; exists {
		err := webrtcConn.SendMessage(&types.SignalingMessage{
			Type:   types.MessageTypeOffer,
			UserID: userID,
			Data: types.WebRTCMessage{
				SDP:  offer.SDP,
				Type: "offer",
			},
			Timestamp: time.Now(),
		})
		if err != nil {
			zLog.Error("Failed to send renegotiation offer",
				zap.String("user_id", userID),
				zap.Error(err))
		} else {
			zLog.Info("发送重新协商Offer",
				zap.String("user_id", userID))
		}
	}
	s.mu.RUnlock()
}

// 🔥 将房间内已有用户的媒体流添加到新加入的用户
func (s *SignalingServer) addExistingTracksToNewUser(roomID, newUserID string, newUserConn *webrtc.WebRTCConnection) {
	// 获取房间信息
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		return
	}

	// 获取房间内的所有用户
	users := room.GetUsers()

	zLog.Info("为新用户添加已有媒体流",
		zap.String("new_user", newUserID),
		zap.String("room_id", roomID),
		zap.Int("existing_users", len(users)-1))

	// 遍历房间内的其他用户
	for _, user := range users {
		if user.UserID == newUserID {
			continue // 跳过自己
		}

		// 获取已有用户的WebRTC连接
		s.mu.RLock()
		existingConn, exists := s.connections[user.UserID]
		s.mu.RUnlock()

		if !exists {
			continue
		}

		// 获取已有用户的所有本地轨道
		tracks := existingConn.GetLocalTracks()

		// 将每个轨道添加到新用户
		for _, track := range tracks {
			_, err := newUserConn.AddTrack(track)
			if err != nil {
				zLog.Error("Failed to add existing track to new user",
					zap.String("existing_user", user.UserID),
					zap.String("new_user", newUserID),
					zap.String("track_id", track.ID()),
					zap.Error(err))
				continue
			}

			zLog.Info("已有媒体流添加成功",
				zap.String("from_user", user.UserID),
				zap.String("to_new_user", newUserID),
				zap.String("track_id", track.ID()))
		}
	}

	// 如果添加了track，需要重新协商
	if len(users) > 1 {
		go s.renegotiateConnection(newUserConn, newUserID, nil)
	}
}
