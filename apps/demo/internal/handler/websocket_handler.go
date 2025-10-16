package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/demo/internal/logic"
	"github.com/iceymoss/go-hichat-api/apps/demo/internal/svc"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var signalingServer *logic.SignalingServer

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func init() {
	signalingServer = logic.NewSignalingServer()
}

// WebSocket 处理器
func WebSocketHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 升级为 WebSocket 连接
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket 升级失败: %v", err)
			return
		}

		// 创建客户端
		client := &logic.Client{
			ID:        uuid.New().String(),
			Conn:      conn,
			SendChan:  make(chan []byte, 256),
			CloseChan: make(chan struct{}),
		}

		// 添加到信令服务器
		signalingServer.AddClient(client)

		// 发送客户端 ID
		welcomeMsg := logic.SignalMessage{
			Type:   "welcome",
			FromID: client.ID,
		}
		if data, err := json.Marshal(welcomeMsg); err == nil {
			conn.WriteMessage(websocket.TextMessage, data)
		}

		// 启动读写协程
		go readPump(client, signalingServer)
		go writePump(client, signalingServer)
	}
}

// 读取消息
func readPump(client *logic.Client, server *logic.SignalingServer) {
	defer func() {
		server.RemoveClient(client.ID)
		close(client.CloseChan)
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 错误: %v", err)
			}
			break
		}

		// 解析消息
		var msg logic.SignalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("解析消息失败: %v", err)
			continue
		}

		// 转发信令消息
		if msg.Type == "offer" || msg.Type == "answer" || msg.Type == "candidate" {
			server.ForwardMessage(client.ID, msg)
		}
	}
}

// 发送消息
func writePump(client *logic.Client, server *logic.SignalingServer) {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.SendChan:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-client.CloseChan:
			return
		}
	}
}

// 获取状态
func StatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "running",
			"clients":      signalingServer.GetClientCount(),
			"stun_servers": svcCtx.Config.StunServers,
		})
	}
}
