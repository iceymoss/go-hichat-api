package handler

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/logic"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/types"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	imws "github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	wsframe "github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = 25 * time.Second
)

// clientConn 单个用户的 streaming ws 连接（按鉴权 uid 登记）。
// gorilla 连接不允许并发写，所有写经 writeMu 串行化。
type clientConn struct {
	uid     string
	conn    *websocket.Conn
	writeMu sync.Mutex
}

func (c *clientConn) sendJSON(v any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(v)
}

func (c *clientConn) ping() error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second))
}

func (c *clientConn) close() {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.conn.Close()
}

// SignalingServer 信令服务器：JWT 鉴权的 streaming ws，
// 负责呼叫控制（创建/接听/拒接/取消/挂断）与 1:1 媒体协商 relay（offer/answer/ice）。
// 控制结果通过 im ws（push.call -> 客户端 call.signal）下发；媒体帧在两端 streaming ws 之间转发。
type SignalingServer struct {
	svc      *svc.ServiceContext
	auth     *JwtAuth
	upgrader websocket.Upgrader
	calls    *logic.CallService

	mu    sync.RWMutex
	conns map[string]*clientConn // uid -> conn（单会话，重复登录顶号）

	ctx    context.Context
	cancel context.CancelFunc
}

// NewSignalingServer 创建信令服务器
func NewSignalingServer(svcCtx *svc.ServiceContext) *SignalingServer {
	ctx, cancel := context.WithCancel(context.Background())

	ring := time.Duration(svcCtx.Config.Call.RingTimeoutSeconds) * time.Second
	s := &SignalingServer{
		svc:   svcCtx,
		auth:  NewJwtAuth(svcCtx),
		calls: logic.NewCallService(ring),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  svcCtx.Config.Signaling.WebSocket.ReadBufferSize,
			WriteBufferSize: svcCtx.Config.Signaling.WebSocket.WriteBufferSize,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		conns:  make(map[string]*clientConn),
		ctx:    ctx,
		cancel: cancel,
	}

	// 振铃超时：通知双方“未接听”
	s.calls.SetTimeoutHandler(func(sess *logic.CallSession) {
		s.pushSignal(&wsframe.CallSignal{ReceiverId: sess.CallerID, Event: "timeout", CallId: sess.ID, Reason: string(sess.EndReason)})
		s.pushSignal(&wsframe.CallSignal{ReceiverId: sess.CalleeID, Event: "timeout", CallId: sess.ID, Reason: string(sess.EndReason)})
	})

	return s
}

// HandleWebSocket 处理 WebSocket 连接：JWT 鉴权 -> 升级 -> 按 uid 登记 -> 读循环 -> 断线清理。
func (s *SignalingServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	uid := s.auth.ParseUID(r)
	if uid == "" {
		zLog.Warn("streaming ws auth failed", zap.String("remote_addr", r.RemoteAddr))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		zLog.Error("Failed to upgrade websocket connection", zap.Error(err))
		return
	}

	c := &clientConn{uid: uid, conn: conn}
	s.register(c)
	zLog.Info("streaming ws connected", zap.String("user_id", uid))

	stopPing := make(chan struct{})
	go s.pingLoop(c, stopPing)

	s.readLoop(c)

	close(stopPing)
	s.unregister(c)
	s.cleanupCall(uid)
	zLog.Info("streaming ws disconnected", zap.String("user_id", uid))
}

// register 登记连接；同一 uid 已有连接则顶号（关旧）。
func (s *SignalingServer) register(c *clientConn) {
	s.mu.Lock()
	if old, ok := s.conns[c.uid]; ok {
		old.close()
	}
	s.conns[c.uid] = c
	s.mu.Unlock()
}

// unregister 仅当当前登记的就是本连接时才删除（避免顶号后误删新连接）。
func (s *SignalingServer) unregister(c *clientConn) {
	s.mu.Lock()
	if cur, ok := s.conns[c.uid]; ok && cur == c {
		delete(s.conns, c.uid)
	}
	s.mu.Unlock()
}

func (s *SignalingServer) getConn(uid string) (*clientConn, bool) {
	s.mu.RLock()
	c, ok := s.conns[uid]
	s.mu.RUnlock()
	return c, ok
}

// pingLoop 周期性发送 ws ping（浏览器自动回 pong），维持连接存活检测。
func (s *SignalingServer) pingLoop(c *clientConn, stop <-chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := c.ping(); err != nil {
				return
			}
		}
	}
}

// readLoop 读循环：心跳超时关连接；逐帧路由。
func (s *SignalingServer) readLoop(c *clientConn) {
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg types.SignalingMessage
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zLog.Info("streaming ws read closed", zap.String("user_id", c.uid), zap.Error(err))
			}
			return
		}
		s.route(c, &msg)
	}
}

// route 按消息类型分发。以鉴权 uid 为准，忽略客户端自报的 UserID（防冒充）。
func (s *SignalingServer) route(c *clientConn, msg *types.SignalingMessage) {
	switch msg.Type {
	case types.MessageTypeCallInvite:
		s.handleInvite(c, msg)
	case types.MessageTypeCallAccept:
		s.handleAccept(c, msg)
	case types.MessageTypeCallReject:
		s.handleReject(c, msg)
	case types.MessageTypeCallCancel:
		s.handleCancel(c, msg)
	case types.MessageTypeCallEnd:
		s.handleEnd(c, msg)
	case types.MessageTypeOffer, types.MessageTypeAnswer, types.MessageTypeIceCandidate, types.MessageTypeMediaState:
		s.relayToPeer(c, msg)
	default:
		s.sendError(c, "unknown message type: "+string(msg.Type))
	}
}

// handleInvite 主叫发起：好友校验 -> 创建会话 -> 回执主叫 callId -> 振铃被叫（im ws）。
func (s *SignalingServer) handleInvite(c *clientConn, msg *types.SignalingMessage) {
	calleeID := dataStr(msg.Data, "callee_id")
	if calleeID == "" {
		s.sendError(c, "callee_id required")
		return
	}
	callType := logic.CallType(dataStr(msg.Data, "call_type"))
	if callType != logic.CallVoice && callType != logic.CallVideo {
		callType = logic.CallVoice
	}

	if !s.areFriends(c.uid, calleeID) {
		s.sendError(c, "not friends")
		return
	}

	sess, err := s.calls.Create(c.uid, calleeID, callType)
	if err != nil {
		if err == logic.ErrBusy {
			// 被叫忙线：回执主叫 busy
			s.send(c, &types.SignalingMessage{Type: types.MessageTypeCallReject, Data: map[string]any{"reason": "busy"}})
		} else {
			s.sendError(c, err.Error())
		}
		return
	}

	// 回执主叫：携带 callId（后续 offer/ice 用）
	s.send(c, &types.SignalingMessage{
		Type:   types.MessageTypeCallCreated,
		RoomID: sess.ID,
		Data:   map[string]any{"call_id": sess.ID, "call_type": string(callType), "callee_id": calleeID, "media_mode": "p2p"},
	})

	// 振铃被叫（im ws -> call.signal）。昵称/头像由前端用本地好友资料解析，
	// 不在此发 RPC，避免取资料阻塞关键的振铃路径。
	s.pushSignal(&wsframe.CallSignal{
		ReceiverId: calleeID,
		Event:      "invite",
		CallId:     sess.ID,
		CallType:   string(callType),
		MediaMode:  "p2p",
		Scope:      "single",
		FromUid:    c.uid,
	})
}

func (s *SignalingServer) handleAccept(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	sess, err := s.calls.Accept(callID, c.uid)
	if err != nil {
		s.sendError(c, err.Error())
		return
	}
	zLog.Info("call accepted", zap.String("call_id", sess.ID), zap.String("by", c.uid), zap.String("notify_caller", sess.CallerID))
	s.pushSignal(&wsframe.CallSignal{ReceiverId: sess.CallerID, Event: "accept", CallId: sess.ID, FromUid: c.uid})
}

func (s *SignalingServer) handleReject(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	sess, err := s.calls.Reject(callID, c.uid)
	if err != nil {
		s.sendError(c, err.Error())
		return
	}
	s.pushSignal(&wsframe.CallSignal{ReceiverId: sess.CallerID, Event: "reject", CallId: sess.ID})
}

func (s *SignalingServer) handleCancel(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	sess, err := s.calls.Cancel(callID, c.uid)
	if err != nil {
		s.sendError(c, err.Error())
		return
	}
	s.pushSignal(&wsframe.CallSignal{ReceiverId: sess.CalleeID, Event: "cancel", CallId: sess.ID})
}

func (s *SignalingServer) handleEnd(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	sess, err := s.calls.End(callID, c.uid)
	if err != nil {
		s.sendError(c, err.Error())
		return
	}
	peer := sess.Peer(c.uid)
	s.pushSignal(&wsframe.CallSignal{ReceiverId: peer, Event: "end", CallId: sess.ID, Reason: string(sess.EndReason), Duration: sess.Duration()})
}

// relayToPeer 1:1 媒体协商 relay：把 offer/answer/ice/media_state 透传给对端（不经手媒体字节）。
func (s *SignalingServer) relayToPeer(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	sess, ok := s.calls.Get(callID)
	if !ok {
		zLog.Warn("relay: call not found", zap.String("from", c.uid), zap.String("type", string(msg.Type)), zap.String("call_id", callID))
		s.sendError(c, "call not found")
		return
	}
	if sess.CallerID != c.uid && sess.CalleeID != c.uid {
		s.sendError(c, "not a participant")
		return
	}
	peer := sess.Peer(c.uid)
	pc, ok := s.getConn(peer)
	if !ok {
		// 对端不在线/未连 streaming ws，媒体帧丢弃（控制层会处理掉线）
		zLog.Warn("relay: peer not connected to streaming ws",
			zap.String("type", string(msg.Type)), zap.String("from", c.uid), zap.String("peer", peer))
		return
	}
	// 透传原帧，标记来源 uid（覆盖客户端自报值）
	msg.UserID = c.uid
	if err := pc.sendJSON(msg); err != nil {
		zLog.Error("relay to peer failed", zap.String("peer", peer), zap.Error(err))
		return
	}
	zLog.Info("relay ok", zap.String("type", string(msg.Type)), zap.String("from", c.uid), zap.String("to", peer))
}

// cleanupCall 断线清理：若用户仍在通话中，结束并通知对端。
func (s *SignalingServer) cleanupCall(uid string) {
	callID, ok := s.calls.ActiveCallID(uid)
	if !ok {
		return
	}
	sess, err := s.calls.End(callID, uid)
	if err != nil {
		return
	}
	peer := sess.Peer(uid)
	s.pushSignal(&wsframe.CallSignal{ReceiverId: peer, Event: "end", CallId: sess.ID, Reason: string(sess.EndReason), Duration: sess.Duration()})
}

// pushSignal 通过 im ws（push.call）把通话控制信令单推给接收者；离线即丢。
func (s *SignalingServer) pushSignal(sig *wsframe.CallSignal) {
	sig.Timestamp = time.Now().Unix()
	if err := s.svc.ImWsClient.Send(imws.Message{
		FrameType: imws.FrameNoAck,
		Method:    "push.call",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      sig,
	}); err != nil {
		zLog.Error("push call signal failed",
			zap.String("receiver", sig.ReceiverId), zap.String("event", sig.Event), zap.Error(err))
	}
}

// areFriends 校验两用户是否好友：先读关系缓存，Unknown 回源 FriendList 并回填，RPC 失败 fail-open。
func (s *SignalingServer) areFriends(uid, peer string) bool {
	switch s.svc.RelationCache.IsFriend(s.ctx, uid, peer) {
	case relationcache.VerdictAllowed:
		return true
	case relationcache.VerdictDenied:
		return false
	}
	resp, err := s.svc.Social.FriendList(s.ctx, &socialclient.FriendListReq{UserId: uid})
	if err != nil || resp == nil {
		return true // fail-open：不确定时放行（与 im 发送鉴权一致）
	}
	friends := make([]string, 0, len(resp.List))
	found := false
	for _, f := range resp.List {
		friends = append(friends, f.FriendUid)
		if f.FriendUid == peer {
			found = true
		}
	}
	_ = s.svc.RelationCache.LoadFriends(s.ctx, uid, friends, 0)
	return found
}

// userInfo 取用户昵称/头像（来电界面展示），失败返回空串。
// 已从振铃路径移除调用（前端用本地资料解析）；保留供后续按需使用。

func (s *SignalingServer) send(c *clientConn, msg *types.SignalingMessage) {
	msg.Timestamp = time.Now()
	if err := c.sendJSON(msg); err != nil {
		zLog.Error("send to client failed", zap.String("user_id", c.uid), zap.Error(err))
	}
}

func (s *SignalingServer) sendError(c *clientConn, errMsg string) {
	s.send(c, &types.SignalingMessage{Type: types.MessageTypeError, Data: map[string]string{"error": errMsg}})
}

// Close 关闭信令服务器：取消上下文并关闭所有连接。
func (s *SignalingServer) Close() error {
	s.cancel()
	s.mu.Lock()
	for uid, c := range s.conns {
		c.close()
		delete(s.conns, uid)
	}
	s.mu.Unlock()
	zLog.Info("signaling server closed")
	return nil
}

// msgCallID 取 callId：优先 data.call_id，回退顶层 RoomID。
func msgCallID(msg *types.SignalingMessage) string {
	if id := dataStr(msg.Data, "call_id"); id != "" {
		return id
	}
	return msg.RoomID
}

func dataStr(data any, key string) string {
	m, ok := data.(map[string]any)
	if !ok {
		return ""
	}
	v, _ := m[key].(string)
	return v
}
