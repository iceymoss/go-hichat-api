package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/streaming/room"
	"github.com/iceymoss/go-hichat-api/apps/streaming/sfu"
	"github.com/iceymoss/go-hichat-api/apps/streaming/webrtc"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// SignalingServer 信令服务器
type SignalingServer struct {
	svc          *svc.ServiceContext
	roomManager  *room.RoomManager
	sfu          *sfu.SFU
	connections  map[string]*webrtc.WebRTCConnection
	mu           sync.RWMutex
	upgrader     websocket.Upgrader
	messageQueue chan *SignalingMessage
	workerCount  int
	ctx          context.Context
	cancel       context.CancelFunc
}

// SignalingMessage 信令消息包装器
type SignalingMessage struct {
	Conn    *websocket.Conn
	Message *types.SignalingMessage
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
				return svc.Config.Signaling.WebSocket.CheckOrigin
			},
		},
		messageQueue: make(chan *SignalingMessage, svc.Config.Signaling.MessageQueue.BufferSize),
		workerCount:  svc.Config.Signaling.MessageQueue.WorkerCount,
		ctx:          ctx,
		cancel:       cancel,
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
				}
				return
			}

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
	zLog.Debug("Handling signaling message",
		zap.String("type", string(msg.Type)),
		zap.String("user_id", msg.UserID),
		zap.String("room_id", msg.RoomID))

	switch msg.Type {
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
	default:
		s.sendError(conn, msg.UserID, fmt.Sprintf("unknown message type: %s", msg.Type))
	}
}

// handleJoinRoom 处理加入房间
func (s *SignalingServer) handleJoinRoom(conn *websocket.Conn, msg *types.SignalingMessage) {
	userID := msg.UserID
	roomID := msg.RoomID

	if userID == "" || roomID == "" {
		s.sendError(conn, userID, "user_id and room_id are required")
		return
	}

	// 创建或获取房间
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		// 房间不存在，创建新房间
		room, err = s.roomManager.CreateRoom(roomID, fmt.Sprintf("Room %s", roomID))
		if err != nil {
			s.sendError(conn, userID, fmt.Sprintf("failed to create room: %v", err))
			return
		}
	}

	// 创建用户
	user := &types.User{
		UserID:    userID,
		Username:  fmt.Sprintf("User %s", userID),
		JoinedAt:  time.Now(),
		IsMuted:   false,
		IsVideoOn: true,
	}

	// 添加用户到房间
	if err := room.AddUser(user); err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to join room: %v", err))
		return
	}

	// 创建WebRTC连接
	webrtcConfig := s.getWebRTCConfig()
	webrtcConn, err := webrtc.NewWebRTCConnection(userID, &WebSocketConnection{conn: conn}, webrtcConfig)
	if err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to create WebRTC connection: %v", err))
		return
	}

	// 保存连接
	s.mu.Lock()
	s.connections[userID] = webrtcConn
	s.mu.Unlock()

	// 添加用户到SFU
	if err := s.sfu.AddUserToRoom(roomID, userID, webrtcConn.GetPeerConnection()); err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to add user to SFU: %v", err))
		return
	}

	// 发送房间信息
	roomInfo := room.GetRoomInfo()
	s.sendMessage(conn, &types.SignalingMessage{
		Type:      types.MessageTypeRoomInfo,
		RoomID:    roomID,
		UserID:    userID,
		Data:      roomInfo,
		Timestamp: time.Now(),
	})

	// 通知房间内其他用户
	s.broadcastToRoom(roomID, userID, &types.SignalingMessage{
		Type:      types.MessageTypeUserJoined,
		RoomID:    roomID,
		UserID:    userID,
		Data:      user,
		Timestamp: time.Now(),
	})

	zLog.Info("User joined room",
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
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
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
	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
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

	// 添加ICE候选
	iceCandidate := webrtc.ICECandidateInit{
		Candidate:     candidate,
		SDPMLineIndex: int(sdpMLineIndex),
		SDPMid:        &sdpMid,
	}

	if err := webrtcConn.AddICECandidate(iceCandidate); err != nil {
		s.sendError(conn, userID, fmt.Sprintf("failed to add ICE candidate: %v", err))
		return
	}

	zLog.Debug("ICE candidate handled",
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
func (s *SignalingServer) getWebRTCConfig() *webrtc.Configuration {
	config := &webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{},
	}

	// 添加ICE服务器
	for _, iceServer := range s.svc.Config.WebRTC.IceServers {
		server := webrtc.ICEServer{
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
}

func (w *WebSocketConnection) SendMessage(message *types.SignalingMessage) error {
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
