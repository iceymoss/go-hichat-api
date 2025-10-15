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

	// 媒体轨道相关
	localTracks map[string]*webrtc.TrackLocalStaticRTP // 本地轨道（用于转发）
	tracksMu    sync.RWMutex

	// 回调函数
	onTrackHandler func(track *webrtc.TrackLocalStaticRTP) // 收到远端轨道时的回调

	// ICE候选缓存（用于处理早到的候选）
	pendingCandidates []webrtc.ICECandidateInit
	candidatesMu      sync.Mutex
}

// NewWebRTCConnection 创建新的WebRTC连接
func NewWebRTCConnection(userID string, wsConn types.WebSocketConnection, config *webrtc.Configuration) (*WebRTCConnection, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 🔥 关键：创建 SettingEngine 支持 localhost 测试
	settingEngine := webrtc.SettingEngine{}

	// 允许监听所有网络接口，包括 localhost
	settingEngine.SetNetworkTypes([]webrtc.NetworkType{
		webrtc.NetworkTypeUDP4,
		webrtc.NetworkTypeUDP6,
		webrtc.NetworkTypeTCP4,
		webrtc.NetworkTypeTCP6,
	})

	// 设置端口范围（可选）
	// settingEngine.SetEphemeralUDPPortRange(10000, 20000)

	// 创建 API 实例
	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))

	// 使用自定义 API 创建 PeerConnection
	peerConnection, err := api.NewPeerConnection(*config)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	zLog.Info("PeerConnection 创建成功",
		zap.String("user_id", userID),
		zap.Int("ice_servers", len(config.ICEServers)))

	conn := &WebRTCConnection{
		userID:            userID,
		peerConnection:    peerConnection,
		wsConn:            wsConn,
		connectedAt:       time.Now(),
		ctx:               ctx,
		cancel:            cancel,
		localTracks:       make(map[string]*webrtc.TrackLocalStaticRTP),
		pendingCandidates: make([]webrtc.ICECandidateInit, 0),
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

		// 如果连接失败，输出诊断信息
		if state == webrtc.ICEConnectionStateFailed || state == webrtc.ICEConnectionStateDisconnected {
			zLog.Error("❌ ICE 连接失败",
				zap.String("user_id", userID),
				zap.String("state", state.String()),
				zap.String("提示", "请检查网络配置和ICE服务器"))
		}
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

			// 🔍 详细的 ICE 候选日志
			zLog.Info("🧊 服务器生成ICE候选",
				zap.String("user_id", userID),
				zap.String("candidate", candidateJSON.Candidate),
				zap.String("type", candidate.Typ.String()),
				zap.String("protocol", candidate.Protocol.String()),
				zap.String("address", candidate.Address),
				zap.Uint16("port", candidate.Port))

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
			} else {
				zLog.Info("✅ ICE候选已发送给客户端",
					zap.String("user_id", userID))
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

	// 🔥 核心：设置媒体轨道接收处理
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		zLog.Info("收到远端媒体轨道",
			zap.String("user_id", userID),
			zap.String("track_id", track.ID()),
			zap.String("stream_id", track.StreamID()),
			zap.String("kind", track.Kind().String()),
			zap.Uint32("ssrc", uint32(track.SSRC())))

		// 异步处理媒体流
		go conn.handleIncomingTrack(track, receiver)
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

	// 设置远程描述
	if err := c.peerConnection.SetRemoteDescription(desc); err != nil {
		return err
	}

	// 🔥 设置完成后，添加所有pending的ICE候选
	c.candidatesMu.Lock()
	pendingCount := len(c.pendingCandidates)
	if pendingCount > 0 {
		zLog.Info("处理缓存的ICE候选",
			zap.String("user_id", c.userID),
			zap.Int("count", pendingCount))

		for _, candidate := range c.pendingCandidates {
			if err := c.peerConnection.AddICECandidate(candidate); err != nil {
				zLog.Error("添加缓存的ICE候选失败",
					zap.String("user_id", c.userID),
					zap.Error(err))
			}
		}
		// 清空缓存
		c.pendingCandidates = make([]webrtc.ICECandidateInit, 0)
	}
	c.candidatesMu.Unlock()

	return nil
}

// AddICECandidate 添加ICE候选
func (c *WebRTCConnection) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.peerConnection == nil {
		return fmt.Errorf("peer connection is nil")
	}

	// 🔥 关键：检查是否已设置远程描述
	if c.peerConnection.RemoteDescription() == nil {
		// 还没设置RemoteDescription，缓存候选
		c.candidatesMu.Lock()
		c.pendingCandidates = append(c.pendingCandidates, candidate)
		c.candidatesMu.Unlock()

		zLog.Info("⏳ ICE候选已缓存（等待RemoteDescription）",
			zap.String("user_id", c.userID),
			zap.Int("pending_count", len(c.pendingCandidates)))
		return nil
	}

	// 已设置RemoteDescription，直接添加
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

// 🔥 核心方法：处理收到的远端媒体轨道
func (c *WebRTCConnection) handleIncomingTrack(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	// 创建本地轨道用于转发
	localTrack, err := webrtc.NewTrackLocalStaticRTP(
		remoteTrack.Codec().RTPCodecCapability,
		remoteTrack.ID(),
		remoteTrack.StreamID(),
	)
	if err != nil {
		zLog.Error("Failed to create local track",
			zap.String("user_id", c.userID),
			zap.Error(err))
		return
	}

	// 保存本地轨道
	c.tracksMu.Lock()
	c.localTracks[remoteTrack.ID()] = localTrack
	c.tracksMu.Unlock()

	zLog.Info("本地轨道创建成功，准备转发",
		zap.String("user_id", c.userID),
		zap.String("track_id", localTrack.ID()),
		zap.String("kind", remoteTrack.Kind().String()))

	// 触发回调，通知上层有新的轨道可以转发
	if c.onTrackHandler != nil {
		c.onTrackHandler(localTrack)
	}

	// 🎯 核心循环：读取RTP包并写入本地轨道
	buf := make([]byte, 1500) // MTU大小
	rtpPacketCount := 0
	lastLogTime := time.Now()

	for {
		select {
		case <-c.ctx.Done():
			zLog.Info("停止处理媒体流（上下文取消）",
				zap.String("user_id", c.userID),
				zap.String("track_id", remoteTrack.ID()))
			return
		default:
			// 从远端轨道读取RTP包
			n, _, err := remoteTrack.Read(buf)
			if err != nil {
				if err.Error() == "EOF" || err.Error() == "io: read/write on closed pipe" {
					zLog.Info("媒体轨道已关闭",
						zap.String("user_id", c.userID),
						zap.String("track_id", remoteTrack.ID()))
					return
				}
				zLog.Error("Failed to read from remote track",
					zap.String("user_id", c.userID),
					zap.Error(err))
				return
			}

			// 写入本地轨道（供转发使用）
			if _, err := localTrack.Write(buf[:n]); err != nil {
				if err.Error() != "io: read/write on closed pipe" {
					zLog.Error("Failed to write to local track",
						zap.String("user_id", c.userID),
						zap.Error(err))
				}
				return
			}

			rtpPacketCount++

			// 每5秒记录一次统计信息
			if time.Since(lastLogTime) > 5*time.Second {
				zLog.Debug("媒体流转发统计",
					zap.String("user_id", c.userID),
					zap.String("track_id", remoteTrack.ID()),
					zap.String("kind", remoteTrack.Kind().String()),
					zap.Int("packets", rtpPacketCount))
				rtpPacketCount = 0
				lastLogTime = time.Now()
			}
		}
	}
}

// SetOnTrackHandler 设置轨道接收回调
func (c *WebRTCConnection) SetOnTrackHandler(handler func(track *webrtc.TrackLocalStaticRTP)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onTrackHandler = handler
}

// GetLocalTracks 获取所有本地轨道
func (c *WebRTCConnection) GetLocalTracks() []*webrtc.TrackLocalStaticRTP {
	c.tracksMu.RLock()
	defer c.tracksMu.RUnlock()

	tracks := make([]*webrtc.TrackLocalStaticRTP, 0, len(c.localTracks))
	for _, track := range c.localTracks {
		tracks = append(tracks, track)
	}
	return tracks
}

// RemoveTrack 移除轨道
func (c *WebRTCConnection) RemoveTrack(trackID string) {
	c.tracksMu.Lock()
	defer c.tracksMu.Unlock()
	delete(c.localTracks, trackID)
}
