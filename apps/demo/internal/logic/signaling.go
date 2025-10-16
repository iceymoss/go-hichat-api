package logic

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// 信令消息类型
type SignalMessage struct {
	Type      string          `json:"type"` // offer, answer, candidate, join, leave
	SDP       string          `json:"sdp,omitempty"`
	Candidate json.RawMessage `json:"candidate,omitempty"`
	FromID    string          `json:"fromId,omitempty"`
	ToID      string          `json:"toId,omitempty"`
}

// 客户端连接
type Client struct {
	ID        string
	Conn      *websocket.Conn
	PeerID    string // 连接的对端 ID
	SendChan  chan []byte
	CloseChan chan struct{}
	IsMatched bool
}

// 信令服务器
type SignalingServer struct {
	clients       map[string]*Client
	waitingClient *Client // 等待配对的客户端
	mutex         sync.RWMutex
}

func NewSignalingServer() *SignalingServer {
	return &SignalingServer{
		clients: make(map[string]*Client),
	}
}

// 添加客户端
func (s *SignalingServer) AddClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.clients[client.ID] = client
	log.Printf("客户端已连接: %s, 当前在线: %d", client.ID, len(s.clients))

	// 自动配对逻辑
	if s.waitingClient == nil {
		// 当前客户端进入等待
		s.waitingClient = client
		log.Printf("客户端 %s 正在等待配对...", client.ID)
	} else if !s.waitingClient.IsMatched {
		// 配对两个客户端
		client1 := s.waitingClient
		client2 := client

		client1.PeerID = client2.ID
		client2.PeerID = client1.ID
		client1.IsMatched = true
		client2.IsMatched = true

		log.Printf("配对成功: %s <-> %s", client1.ID, client2.ID)

		// 通知双方开始连接
		s.notifyMatch(client1, client2)

		s.waitingClient = nil
	}
}

// 通知配对成功
func (s *SignalingServer) notifyMatch(client1, client2 *Client) {
	// 通知 client1 - 作为呼叫方（caller），需要创建 offer
	msg1 := SignalMessage{
		Type: "matched",
		ToID: client2.ID,
		SDP:  "caller", // 标记为呼叫方
	}
	if data, err := json.Marshal(msg1); err == nil {
		select {
		case client1.SendChan <- data:
		default:
		}
	}

	// 通知 client2 - 作为被叫方（callee），等待 offer
	msg2 := SignalMessage{
		Type: "matched",
		ToID: client1.ID,
		SDP:  "callee", // 标记为被叫方
	}
	if data, err := json.Marshal(msg2); err == nil {
		select {
		case client2.SendChan <- data:
		default:
		}
	}
}

// 移除客户端
func (s *SignalingServer) RemoveClient(clientID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client, exists := s.clients[clientID]
	if !exists {
		return
	}

	// 通知对端断开
	if client.PeerID != "" {
		if peer, ok := s.clients[client.PeerID]; ok {
			peer.PeerID = ""
			peer.IsMatched = false
			msg := SignalMessage{
				Type: "peer-disconnected",
			}
			if data, err := json.Marshal(msg); err == nil {
				select {
				case peer.SendChan <- data:
				default:
				}
			}
		}
	}

	// 如果是等待中的客户端，清空等待
	if s.waitingClient != nil && s.waitingClient.ID == clientID {
		s.waitingClient = nil
	}

	delete(s.clients, clientID)
	log.Printf("客户端已断开: %s, 当前在线: %d", clientID, len(s.clients))
}

// 转发信令消息
func (s *SignalingServer) ForwardMessage(fromID string, msg SignalMessage) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	fromClient, exists := s.clients[fromID]
	if !exists || fromClient.PeerID == "" {
		return
	}

	toClient, exists := s.clients[fromClient.PeerID]
	if !exists {
		return
	}

	// 设置发送者 ID
	msg.FromID = fromID
	msg.ToID = toClient.ID

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化消息失败: %v", err)
		return
	}

	select {
	case toClient.SendChan <- data:
		log.Printf("转发消息: %s -> %s, 类型: %s", fromID, toClient.ID, msg.Type)
	default:
		log.Printf("发送通道已满，丢弃消息")
	}
}

// 获取客户端数量
func (s *SignalingServer) GetClientCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.clients)
}

// 创建 WebRTC PeerConnection 配置
func CreatePeerConnectionConfig(stunServers []string) webrtc.Configuration {
	var iceServers []webrtc.ICEServer

	if len(stunServers) > 0 {
		urls := make([]string, len(stunServers))
		copy(urls, stunServers)
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs: urls,
		})
	}

	return webrtc.Configuration{
		ICEServers: iceServers,
	}
}
