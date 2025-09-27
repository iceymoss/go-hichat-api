package webrtc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/types"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/pion/webrtc/v3"
)

// WebRTCConnection WebRTC连接实现
type WebRTCConnection struct {
	userID         string
	peerConnection *webrtc.PeerConnection
	wsConn         types.WebSocketConnection
	connectedAt    time.Time
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewWebRTCConnection 创建新的WebRTC连接
func NewWebRTCConnection(userID string, wsConn types.WebSocketConnection, config *webrtc.Configuration) (*WebRTCConnection, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(*config)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	conn := &WebRTCConnection{
		userID:         userID,
		peerConnection: peerConnection,
		wsConn:         wsConn,
		connectedAt:    time.Now(),
		ctx:            ctx,
		cancel:         cancel,
	}

	// 设置连接状态变化处理
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		zLog.Info("WebRTC connection state changed",
			zap.String("user_id", userID),
			zap.String("state", state.String()))

		if state == webrtc.PeerConnectionStateClosed ||
			state == webrtc.PeerConnectionStateFailed ||
			state == webrtc.PeerConnectionStateDisconnected {
			conn.Close()
		}
	})

	// 设置ICE连接状态变化处理
	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		zLog.Info("ICE connection state changed",
			zap.String("user_id", userID),
			zap.String("state", state.String()))
	})

	// 设置ICE候选处理
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			candidateJSON := candidate.ToJSON()
			var sdpMLineIndex *uint16
			var sdpMid *string

			if candidateJSON.SDPMLineIndex != nil {
				sdpMLineIndex = candidateJSON.SDPMLineIndex
			}
			if candidateJSON.SDPMid != nil {
				sdpMid = candidateJSON.SDPMid
			}

			iceMsg := &types.SignalingMessage{
				Type:   types.MessageTypeIceCandidate,
				UserID: userID,
				Data: types.ICECandidateMessage{
					Candidate:     candidateJSON.Candidate,
					SDPMLineIndex: sdpMLineIndex,
					SDPMid:        sdpMid,
				},
				Timestamp: time.Now(),
			}

			if err := conn.SendMessage(iceMsg); err != nil {
				zLog.Error("Failed to send ICE candidate",
					zap.String("user_id", userID),
					zap.Error(err))
			}
		}
	})

	// 设置数据通道处理
	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		zLog.Info("Data channel received",
			zap.String("user_id", userID),
			zap.String("label", dataChannel.Label()))

		dataChannel.OnOpen(func() {
			zLog.Info("Data channel opened",
				zap.String("user_id", userID),
				zap.String("label", dataChannel.Label()))
		})

		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			zLog.Debug("Data channel message received",
				zap.String("user_id", userID),
				zap.String("label", dataChannel.Label()),
				zap.Int("size", len(msg.Data)))
		})
	})

	return conn, nil
}

// GetUserID 获取用户ID
func (c *WebRTCConnection) GetUserID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userID
}

// GetPeerConnection 获取PeerConnection
func (c *WebRTCConnection) GetPeerConnection() *webrtc.PeerConnection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.peerConnection
}

// SendMessage 发送消息
func (c *WebRTCConnection) SendMessage(message *types.SignalingMessage) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.wsConn == nil {
		return fmt.Errorf("websocket connection is nil")
	}

	return c.wsConn.SendMessage(message)
}

// Close 关闭连接
func (c *WebRTCConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cancel()

	if c.peerConnection != nil {
		if err := c.peerConnection.Close(); err != nil {
			zLog.Error("Failed to close peer connection",
				zap.String("user_id", c.userID),
				zap.Error(err))
		}
	}

	if c.wsConn != nil {
		if err := c.wsConn.Close(); err != nil {
			zLog.Error("Failed to close websocket connection",
				zap.String("user_id", c.userID),
				zap.Error(err))
		}
	}

	zLog.Info("WebRTC connection closed", zap.String("user_id", c.userID))
	return nil
}

// IsConnected 检查连接状态
func (c *WebRTCConnection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.peerConnection == nil {
		return false
	}

	state := c.peerConnection.ConnectionState()
	return state == webrtc.PeerConnectionStateConnected ||
		state == webrtc.PeerConnectionStateConnecting
}

// GetConnectedAt 获取连接时间
func (c *WebRTCConnection) GetConnectedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connectedAt
}

// CreateOffer 创建Offer
func (c *WebRTCConnection) CreateOffer() (*webrtc.SessionDescription, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.peerConnection == nil {
		return nil, fmt.Errorf("peer connection is nil")
	}

	offer, err := c.peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	// 设置本地描述
	if err := c.peerConnection.SetLocalDescription(offer); err != nil {
		return nil, fmt.Errorf("failed to set local description: %w", err)
	}

	return &offer, nil
}

// CreateAnswer 创建Answer
func (c *WebRTCConnection) CreateAnswer() (*webrtc.SessionDescription, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.peerConnection == nil {
		return nil, fmt.Errorf("peer connection is nil")
	}

	answer, err := c.peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create answer: %w", err)
	}

	// 设置本地描述
	if err := c.peerConnection.SetLocalDescription(answer); err != nil {
		return nil, fmt.Errorf("failed to set local description: %w", err)
	}

	return &answer, nil
}

// SetRemoteDescription 设置远程描述
func (c *WebRTCConnection) SetRemoteDescription(desc webrtc.SessionDescription) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.peerConnection == nil {
		return fmt.Errorf("peer connection is nil")
	}

	return c.peerConnection.SetRemoteDescription(desc)
}

// AddICECandidate 添加ICE候选
func (c *WebRTCConnection) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.peerConnection == nil {
		return fmt.Errorf("peer connection is nil")
	}

	return c.peerConnection.AddICECandidate(candidate)
}

// AddTrack 添加轨道
func (c *WebRTCConnection) AddTrack(track *webrtc.TrackLocalStaticRTP) (*webrtc.RTPSender, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.peerConnection == nil {
		return nil, fmt.Errorf("peer connection is nil")
	}

	return c.peerConnection.AddTrack(track)
}

// GetSenders 获取发送器
func (c *WebRTCConnection) GetSenders() []*webrtc.RTPSender {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.peerConnection == nil {
		return nil
	}

	return c.peerConnection.GetSenders()
}

// GetReceivers 获取接收器
func (c *WebRTCConnection) GetReceivers() []*webrtc.RTPReceiver {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.peerConnection == nil {
		return nil
	}

	return c.peerConnection.GetReceivers()
}
