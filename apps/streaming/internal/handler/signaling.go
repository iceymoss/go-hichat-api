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
	"github.com/iceymoss/go-hichat-api/apps/streaming/sfu"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"

	"github.com/gorilla/websocket"
	imws "github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	wsframe "github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/pion/webrtc/v3"
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
	groups   *logic.GroupCallService
	sfu      *sfu.SFU // 群通话 SFU：进程内媒体转发（1:1 不经此）

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
		svc:    svcCtx,
		auth:   NewJwtAuth(svcCtx),
		calls:  logic.NewCallService(ring),
		groups: logic.NewGroupCallService(svcCtx.Config.SFU.MaxUsersPerRoom),
		sfu: sfu.NewSFU(
			sfu.WithPublicIP(svcCtx.Config.Public.IP),
			sfu.WithUDPPortRange(svcCtx.Config.Public.UDPPortMin, svcCtx.Config.Public.UDPPortMax),
		),
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
		s.notifyUser(sess.CallerID, &wsframe.CallSignal{Event: "timeout", CallId: sess.ID, Reason: string(sess.EndReason)})
		s.notifyUser(sess.CalleeID, &wsframe.CallSignal{Event: "timeout", CallId: sess.ID, Reason: string(sess.EndReason)})
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
	case types.MessageTypeGroupInvite:
		s.handleGroupInvite(c, msg)
	case types.MessageTypeGroupJoin:
		s.handleGroupJoin(c, msg)
	case types.MessageTypeGroupLeave:
		s.handleGroupLeave(c, msg)
	case types.MessageTypeSFUPublish:
		s.handleSFUPublish(c, msg)
	case types.MessageTypeSFUAnswer:
		s.handleSFUAnswer(c, msg)
	case types.MessageTypeSFUIce:
		s.handleSFUIce(c, msg)
	case types.MessageTypeSFUMediaState:
		s.handleSFUMediaState(c, msg)
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
	s.notifyUser(sess.CallerID, &wsframe.CallSignal{Event: "accept", CallId: sess.ID, FromUid: c.uid})
}

func (s *SignalingServer) handleReject(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	sess, err := s.calls.Reject(callID, c.uid)
	if err != nil {
		s.sendError(c, err.Error())
		return
	}
	s.notifyUser(sess.CallerID, &wsframe.CallSignal{Event: "reject", CallId: sess.ID})
}

func (s *SignalingServer) handleCancel(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	sess, err := s.calls.Cancel(callID, c.uid)
	if err != nil {
		s.sendError(c, err.Error())
		return
	}
	s.notifyUser(sess.CalleeID, &wsframe.CallSignal{Event: "cancel", CallId: sess.ID})
}

func (s *SignalingServer) handleEnd(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	sess, err := s.calls.End(callID, c.uid)
	if err != nil {
		s.sendError(c, err.Error())
		return
	}
	peer := sess.Peer(c.uid)
	s.notifyUser(peer, &wsframe.CallSignal{Event: "end", CallId: sess.ID, Reason: string(sess.EndReason), Duration: sess.Duration()})
}

// ==================== 群组通话（Mesh）====================

// handleGroupInvite 发起群通话：校验群成员 -> 建群会话 -> 回执 callId -> 逐个被邀成员振铃（im ws）。
func (s *SignalingServer) handleGroupInvite(c *clientConn, msg *types.SignalingMessage) {
	groupID := dataStr(msg.Data, "group_id")
	if groupID == "" {
		s.sendError(c, "group_id required")
		return
	}
	callType := logic.CallType(dataStr(msg.Data, "call_type"))
	if callType != logic.CallVoice && callType != logic.CallVideo {
		callType = logic.CallVoice
	}
	members := dataStrSlice(msg.Data, "members")
	if len(members) == 0 {
		s.sendError(c, "members required")
		return
	}
	if !s.isGroupMember(c.uid, groupID) {
		s.sendError(c, "not a group member")
		return
	}
	// 过滤：仅保留群成员、去掉自己
	invited := make([]string, 0, len(members))
	for _, m := range members {
		if m != c.uid && s.isGroupMember(m, groupID) {
			invited = append(invited, m)
		}
	}

	callID, err := s.groups.Create(c.uid, groupID, callType, invited)
	if err != nil {
		s.sendError(c, err.Error())
		return
	}

	// 为发起人建 SFU peer（服务端 PeerConnection），接好 renegotiation offer 下发
	if err := s.setupSFUPeer(c, callID); err != nil {
		zLog.Error("sfu setup peer failed", zap.String("uid", c.uid), zap.Error(err))
		_, _, _ = s.groups.Leave(callID, c.uid)
		s.sendError(c, "sfu setup failed")
		return
	}

	// 回执发起人 callId
	s.send(c, &types.SignalingMessage{
		Type:   types.MessageTypeGroupCreated,
		RoomID: callID,
		Data:   map[string]any{"call_id": callID, "call_type": string(callType), "group_id": groupID},
	})

	// 振铃被邀成员（im ws -> call.signal，event=group.invite，带参与者名单）
	roster := append([]string{c.uid}, invited...)
	for _, m := range invited {
		s.pushSignal(&wsframe.CallSignal{
			ReceiverId: m,
			Event:      "group.invite",
			CallId:     callID,
			CallType:   string(callType),
			MediaMode:  "sfu",
			Scope:      "group",
			GroupId:    groupID,
			FromUid:    c.uid,
			Members:    roster,
		})
	}

	// 广播群通话已开始（全体群成员看到横幅/列表标识）
	s.broadcastGroupState(groupID)
}

// setupSFUPeer 为 c.uid 在 callID 房间建一条 SFU PeerConnection，并接好服务端发起的
// renegotiation offer 下发（sfu_offer）。客户端收到 group_created/group_roster 后发 sfu_publish。
func (s *SignalingServer) setupSFUPeer(c *clientConn, callID string) error {
	// 服务端 PC 也配 STUN，增强候选（Phase 1 追加 coturn TURN）
	ice := sfu.STUNServers("stun:stun.l.google.com:19302", "stun:stun1.l.google.com:19302")
	peer, err := s.sfu.AddPeer(callID, c.uid, ice)
	if err != nil {
		return err
	}
	peer.OnOffer(func(sdp string) {
		s.send(c, &types.SignalingMessage{
			Type:   types.MessageTypeSFUOffer,
			RoomID: callID,
			Data:   map[string]any{"call_id": callID, "sdp": sdp},
		})
	})
	peer.OnICECandidate(func(candidate webrtc.ICECandidateInit) {
		s.send(c, &types.SignalingMessage{
			Type:   types.MessageTypeSFUIce,
			RoomID: callID,
			Data: map[string]any{
				"call_id":       callID,
				"candidate":     candidate.Candidate,
				"sdpMid":        candidate.SDPMid,
				"sdpMLineIndex": candidate.SDPMLineIndex,
			},
		})
	})
	return nil
}

// handleSFUPublish 客户端发布 offer -> SFU 建 answer 回执 -> 回填订阅房间内已有发布者。
func (s *SignalingServer) handleSFUPublish(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	peer := s.sfu.GetPeer(c.uid)
	if peer == nil || peer.RoomID() != callID || !s.groups.HasParticipant(callID, c.uid) {
		s.sendError(c, "not a group participant")
		return
	}
	answer, err := peer.Publish(dataStr(msg.Data, "sdp"))
	if err != nil {
		zLog.Error("sfu publish failed", zap.String("uid", c.uid), zap.Error(err))
		s.sendError(c, "sfu publish failed")
		return
	}
	s.send(c, &types.SignalingMessage{
		Type:   types.MessageTypeSFUPublishAnswer,
		RoomID: callID,
		Data:   map[string]any{"call_id": callID, "sdp": answer},
	})
	// 回填：本端订阅房间内已有发布者（迟到者据此收到早入房者媒体）
	s.sfu.SubscribeExisting(callID, c.uid)
}

// handleSFUAnswer 客户端对服务端 renegotiation offer 的 answer。
func (s *SignalingServer) handleSFUAnswer(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	peer := s.sfu.GetPeer(c.uid)
	if peer == nil || peer.RoomID() != callID || !s.groups.HasParticipant(callID, c.uid) {
		return
	}
	if err := peer.HandleAnswer(dataStr(msg.Data, "sdp")); err != nil {
		zLog.Warn("sfu handle answer failed", zap.String("uid", c.uid), zap.Error(err))
	}
}

// handleSFUIce 把客户端 trickle candidate 加到其 SFU PeerConnection。
func (s *SignalingServer) handleSFUIce(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	peer := s.sfu.GetPeer(c.uid)
	if peer == nil || peer.RoomID() != callID || !s.groups.HasParticipant(callID, c.uid) {
		s.sendError(c, "not a group participant")
		return
	}
	data, ok := msg.Data.(map[string]any)
	if !ok {
		return
	}
	candidate := webrtc.ICECandidateInit{Candidate: dataStr(data, "candidate")}
	if mid, ok := data["sdpMid"].(string); ok {
		candidate.SDPMid = &mid
	}
	if index, ok := data["sdpMLineIndex"].(float64); ok {
		i := uint16(index)
		candidate.SDPMLineIndex = &i
	}
	if err := peer.AddICECandidate(candidate); err != nil {
		zLog.Warn("sfu add ice candidate failed", zap.String("uid", c.uid), zap.Error(err))
	}
}

// handleSFUMediaState 广播开关麦/摄像头状态给同一通话内其余参与者（SFU 不经手媒体，状态经控制面同步）。
func (s *SignalingServer) handleSFUMediaState(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	participants, ok := s.groups.Participants(callID)
	if !ok || !s.groups.HasParticipant(callID, c.uid) {
		return
	}
	data, _ := msg.Data.(map[string]any)
	audio, _ := data["audio"].(bool)
	video, _ := data["video"].(bool)
	for _, p := range participants {
		if p == c.uid {
			continue
		}
		if pc, ok := s.getConn(p); ok {
			s.send(pc, &types.SignalingMessage{
				Type:   types.MessageTypeSFUMediaState,
				RoomID: callID,
				Data:   map[string]any{"call_id": callID, "uid": c.uid, "audio": audio, "video": video},
			})
		}
	}
}

// handleGroupJoin 加入群通话：建 SFU peer -> 回执当前名单。媒体由 SFU 转发，无需两两建连。
func (s *SignalingServer) handleGroupJoin(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	groupID, _, _, ok := s.groups.Info(callID)
	if !ok || !s.isGroupMember(c.uid, groupID) {
		s.sendError(c, "not a group member")
		return
	}
	others, callType, err := s.groups.Join(callID, c.uid)
	if err != nil {
		s.sendError(c, err.Error())
		return
	}
	// 为新加入者建 SFU peer（收 sfu_publish 前需就绪）
	if err := s.setupSFUPeer(c, callID); err != nil {
		zLog.Error("sfu setup peer failed", zap.String("uid", c.uid), zap.Error(err))
		_, _, _ = s.groups.Leave(callID, c.uid)
		s.sendError(c, "sfu setup failed")
		return
	}
	// 回执新加入者：当前其他参与者（tile 随 SFU 轨到达出现）
	s.send(c, &types.SignalingMessage{
		Type:   types.MessageTypeGroupRoster,
		RoomID: callID,
		Data:   map[string]any{"call_id": callID, "call_type": string(callType), "participants": others},
	})

	// 广播更新后的参与者名单（全体群成员，含未在通话中的）
	if groupID, _, _, ok := s.groups.Info(callID); ok {
		s.broadcastGroupState(groupID)
	}
}

// handleGroupLeave 离开群通话：关 SFU peer + 通知剩余参与者移除 tile + 广播群横幅状态。
func (s *SignalingServer) handleGroupLeave(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	groupID, _, _, _ := s.groups.Info(callID) // 结束前捕获 groupID（结束后会话被删）
	remaining, ended, err := s.groups.Leave(callID, c.uid)
	if err != nil {
		return
	}
	if ended {
		s.sfu.RemoveRoom(callID)
	} else {
		s.sfu.RemovePeer(c.uid)
	}
	s.broadcastSFUPeerLeft(callID, c.uid, remaining)
	s.broadcastGroupState(groupID)
}

// broadcastSFUPeerLeft 通知剩余参与者某人离开（前端移除其 tile）。
// SFU Phase 0 不主动移除下行轨，故用显式帧驱动 tile 清理。
func (s *SignalingServer) broadcastSFUPeerLeft(callID, uid string, remaining []string) {
	for _, p := range remaining {
		if pc, ok := s.getConn(p); ok {
			s.send(pc, &types.SignalingMessage{
				Type:   types.MessageTypeSFUPeerLeft,
				RoomID: callID,
				Data:   map[string]any{"call_id": callID, "uid": uid},
			})
		}
	}
}

// isGroupMember 校验 uid 是否群成员：读关系缓存，Unknown 时 fail-open（与项目鉴权一致；前端已先按群成员过滤）。
func (s *SignalingServer) isGroupMember(uid, groupID string) bool {
	switch s.svc.RelationCache.IsGroupMember(s.ctx, groupID, uid) {
	case relationcache.VerdictAllowed:
		return true
	case relationcache.VerdictDenied:
		return false
	default:
		return true // Unknown / 冷缓存 → 放行（TODO: 回源 GroupUsers RPC 收紧）
	}
}

// relayToPeer 媒体协商 relay：把 offer/answer/ice/media_state 透传给对端（不经手媒体字节）。
// 带 data.to=目标uid 时走群组 Mesh 定向；否则走 1:1 对端推导。
func (s *SignalingServer) relayToPeer(c *clientConn, msg *types.SignalingMessage) {
	callID := msgCallID(msg)
	to := dataStr(msg.Data, "to")
	var target string
	if to != "" {
		s.sendError(c, "group media uses sfu signaling")
		return
	} else {
		// 1:1：按会话推导对端（保持原行为）
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
		target = sess.Peer(c.uid)
	}

	pc, ok := s.getConn(target)
	if !ok {
		zLog.Warn("relay: target not connected to streaming ws",
			zap.String("type", string(msg.Type)), zap.String("from", c.uid), zap.String("to", target))
		return
	}
	// 透传原帧，标记来源 uid（覆盖客户端自报值）
	msg.UserID = c.uid
	if err := pc.sendJSON(msg); err != nil {
		zLog.Error("relay to peer failed", zap.String("to", target), zap.Error(err))
		return
	}
	zLog.Info("relay ok", zap.String("type", string(msg.Type)), zap.String("from", c.uid), zap.String("to", target))
}

// cleanupCall 断线清理：若用户仍在通话中，结束并通知对端。
func (s *SignalingServer) cleanupCall(uid string) {
	// 1:1
	if callID, ok := s.calls.ActiveCallID(uid); ok {
		if sess, err := s.calls.End(callID, uid); err == nil {
			peer := sess.Peer(uid)
			s.notifyUser(peer, &wsframe.CallSignal{Event: "end", CallId: sess.ID, Reason: string(sess.EndReason), Duration: sess.Duration()})
		}
	}
	// 群组：离开并关 SFU peer（其上行轨与下行订阅经房间路由清理）
	if callID, ok := s.groups.ActiveCallID(uid); ok {
		groupID, _, _, _ := s.groups.Info(callID) // 结束前捕获 groupID
		if remaining, ended, err := s.groups.Leave(callID, uid); err == nil {
			if ended {
				s.sfu.RemoveRoom(callID)
			} else {
				s.sfu.RemovePeer(uid)
			}
			s.broadcastSFUPeerLeft(callID, uid, remaining)
			s.broadcastGroupState(groupID)
		}
	}
}

// notifyUser 下发通话控制信令给某用户：
// 优先走该用户的 streaming ws 直连（稳定、低延迟、不受 im ws 顶号churn 影响）；
// 不在 streaming ws（如初始振铃时被叫还没连）则回退 im ws 的 push.call。
func (s *SignalingServer) notifyUser(uid string, sig *wsframe.CallSignal) {
	sig.ReceiverId = uid
	sig.Timestamp = time.Now().Unix()
	if c, ok := s.getConn(uid); ok {
		s.send(c, &types.SignalingMessage{Type: types.MessageTypeCallSignal, Data: sig})
		return
	}
	s.pushSignal(sig)
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

// groupMemberUids 取群全体成员 uid（用于群通话状态广播）。RPC 失败返回空（不广播）。
func (s *SignalingServer) groupMemberUids(groupID string) []string {
	resp, err := s.svc.Social.GroupUsers(s.ctx, &socialclient.GroupUsersReq{GroupId: groupID})
	if err != nil || resp == nil {
		zLog.Error("group call state: GroupUsers failed", zap.String("group", groupID), zap.Error(err))
		return nil
	}
	uids := make([]string, 0, len(resp.List))
	for _, m := range resp.List {
		uids = append(uids, m.UserId)
	}
	return uids
}

// broadcastGroupState 向群全体成员广播某群当前通话状态（event=group.state）。
// 用于「进群聊看到通话中横幅 / 会话列表通话标识 / 加入通话」。participants 为空表示通话已结束（前端清除）。
// callType 仅在仍活跃时有意义；已结束时取查询快照（可能为空，前端凭 participants 判空即可）。
func (s *SignalingServer) broadcastGroupState(groupID string) {
	if groupID == "" {
		return
	}
	callID, t, participants, ok := s.groups.ActiveByGroup(groupID)
	if !ok {
		participants = []string{} // 已结束：广播空名单让前端清除横幅/标识
	}
	members := s.groupMemberUids(groupID)
	for _, m := range members {
		s.pushSignal(&wsframe.CallSignal{
			ReceiverId: m,
			Event:      "group.state",
			Scope:      "group",
			GroupId:    groupID,
			CallId:     callID,
			CallType:   string(t),
			Members:    participants,
		})
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

func dataStrSlice(data any, key string) []string {
	m, ok := data.(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := m[key].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
