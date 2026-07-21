package logic

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// CallType 通话媒体类型
type CallType string

const (
	CallVoice CallType = "voice"
	CallVideo CallType = "video"
)

// CallState 通话生命周期状态
type CallState string

const (
	CallStateInviting  CallState = "inviting"  // 振铃中
	CallStateConnected CallState = "connected" // 已接通
	CallStateEnded     CallState = "ended"     // 已结束
)

// EndReason 通话结束原因（落库 + 聊天记录展示用）
type EndReason string

const (
	EndCompleted EndReason = "completed" // 正常接通后挂断
	EndCanceled  EndReason = "canceled"  // 主叫响铃中取消
	EndRejected  EndReason = "rejected"  // 被叫拒接
	EndNoAnswer  EndReason = "no_answer" // 振铃超时未接
	EndBusy      EndReason = "busy"      // 被叫忙线
	EndFailed    EndReason = "failed"    // 媒体连接失败
)

var (
	ErrBusy           = errors.New("user busy")
	ErrSelfCall       = errors.New("cannot call self")
	ErrCallNotFound   = errors.New("call not found")
	ErrNotParticipant = errors.New("not a participant")
	ErrBadState       = errors.New("invalid call state")
)

// CallSession 一通通话的权威状态（Phase 1 内存态；Phase 2 结束落 streaming_call 表）。
type CallSession struct {
	ID        string
	Type      CallType
	Scope     string // single（群组留作 Phase 3）
	CallerID  string
	CalleeID  string
	State     CallState
	EndReason EndReason
	CreatedAt time.Time
	StartedAt time.Time
	EndedAt   time.Time

	timer *time.Timer
}

// Peer 返回 uid 在本通话里的对端。
func (s *CallSession) Peer(uid string) string {
	if uid == s.CallerID {
		return s.CalleeID
	}
	return s.CallerID
}

func (s *CallSession) hasParticipant(uid string) bool {
	return uid == s.CallerID || uid == s.CalleeID
}

// Duration 返回接通时长（秒）；未接通返回 0。
func (s *CallSession) Duration() int64 {
	if s.StartedAt.IsZero() || s.EndedAt.IsZero() {
		return 0
	}
	if d := int64(s.EndedAt.Sub(s.StartedAt).Seconds()); d > 0 {
		return d
	}
	return 0
}

// CallService 管理 1:1 通话状态机 + 忙线占用 + 振铃超时。
// 纯内存、可单测；好友关系校验由调用方（信令 handler）在 Create 前完成。
type CallService struct {
	mu          sync.Mutex
	calls       map[string]*CallSession
	userCall    map[string]string // uid -> callId（忙线占用）
	ringTimeout time.Duration
	onTimeout   func(*CallSession)
	idSeq       uint64
	now         func() time.Time
}

// NewCallService 创建通话服务，ringTimeout<=0 时默认 30s。
func NewCallService(ringTimeout time.Duration) *CallService {
	if ringTimeout <= 0 {
		ringTimeout = 30 * time.Second
	}
	return &CallService{
		calls:       make(map[string]*CallSession),
		userCall:    make(map[string]string),
		ringTimeout: ringTimeout,
		now:         time.Now,
	}
}

// SetTimeoutHandler 设置振铃超时回调（锁外调用，用于通知双方“未接听”）。
func (cs *CallService) SetTimeoutHandler(f func(*CallSession)) {
	cs.mu.Lock()
	cs.onTimeout = f
	cs.mu.Unlock()
}

// Create 发起通话：双方均空闲才创建 inviting 会话，标记忙线并启动振铃超时。
func (cs *CallService) Create(callerID, calleeID string, t CallType) (*CallSession, error) {
	if callerID == calleeID {
		return nil, ErrSelfCall
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if _, busy := cs.userCall[callerID]; busy {
		return nil, ErrBusy
	}
	if _, busy := cs.userCall[calleeID]; busy {
		return nil, ErrBusy
	}

	s := &CallSession{
		ID:        cs.nextID(),
		Type:      t,
		Scope:     "single",
		CallerID:  callerID,
		CalleeID:  calleeID,
		State:     CallStateInviting,
		CreatedAt: cs.now(),
	}
	cs.calls[s.ID] = s
	cs.userCall[callerID] = s.ID
	cs.userCall[calleeID] = s.ID

	id := s.ID
	s.timer = time.AfterFunc(cs.ringTimeout, func() { cs.fireTimeout(id) })
	return s, nil
}

// Accept 被叫接听：inviting -> connected。
func (cs *CallService) Accept(callID, uid string) (*CallSession, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	s, err := cs.getActive(callID, uid)
	if err != nil {
		return nil, err
	}
	if uid != s.CalleeID {
		return nil, ErrNotParticipant // 只有被叫能接听
	}
	if s.State != CallStateInviting {
		return nil, ErrBadState
	}

	s.State = CallStateConnected
	s.StartedAt = cs.now()
	cs.stopTimer(s)
	return s, nil
}

// Reject 被叫拒接（仅响铃中）。
func (cs *CallService) Reject(callID, uid string) (*CallSession, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	s, err := cs.getActive(callID, uid)
	if err != nil {
		return nil, err
	}
	if uid != s.CalleeID {
		return nil, ErrNotParticipant
	}
	if s.State != CallStateInviting {
		return nil, ErrBadState
	}
	cs.endLocked(s, EndRejected)
	return s, nil
}

// Cancel 主叫取消（仅响铃中）。
func (cs *CallService) Cancel(callID, uid string) (*CallSession, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	s, err := cs.getActive(callID, uid)
	if err != nil {
		return nil, err
	}
	if uid != s.CallerID {
		return nil, ErrNotParticipant
	}
	if s.State != CallStateInviting {
		return nil, ErrBadState
	}
	cs.endLocked(s, EndCanceled)
	return s, nil
}

// End 挂断（任一方）：已接通=completed；响铃中=主叫 canceled / 被叫 rejected。
func (cs *CallService) End(callID, uid string) (*CallSession, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	s, err := cs.getActive(callID, uid)
	if err != nil {
		return nil, err
	}

	reason := EndCompleted
	if s.State == CallStateInviting {
		if uid == s.CallerID {
			reason = EndCanceled
		} else {
			reason = EndRejected
		}
	}
	cs.endLocked(s, reason)
	return s, nil
}

// Get 返回会话快照（仍存活才有）。
func (cs *CallService) Get(callID string) (*CallSession, bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	s, ok := cs.calls[callID]
	return s, ok
}

// IsBusy 判断用户当前是否在通话中。
func (cs *CallService) IsBusy(uid string) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	_, busy := cs.userCall[uid]
	return busy
}

// ActiveCallID 返回用户当前所在通话 ID（断线清理用）。
func (cs *CallService) ActiveCallID(uid string) (string, bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	id, ok := cs.userCall[uid]
	return id, ok
}

// fireTimeout 振铃超时：仍在 inviting 才结束为 no_answer，并锁外回调通知双方。
func (cs *CallService) fireTimeout(callID string) {
	cs.mu.Lock()
	s, ok := cs.calls[callID]
	if !ok || s.State != CallStateInviting {
		cs.mu.Unlock()
		return
	}
	cs.endLocked(s, EndNoAnswer)
	cb := cs.onTimeout
	cs.mu.Unlock()

	if cb != nil {
		cb(s)
	}
}

// getActive 取存活会话并校验 uid 为参与者（持锁）。
func (cs *CallService) getActive(callID, uid string) (*CallSession, error) {
	s, ok := cs.calls[callID]
	if !ok {
		return nil, ErrCallNotFound
	}
	if !s.hasParticipant(uid) {
		return nil, ErrNotParticipant
	}
	return s, nil
}

// endLocked 置终态、清忙线、移除会话、停表（持锁）。
func (cs *CallService) endLocked(s *CallSession, reason EndReason) {
	if s.State == CallStateEnded {
		return
	}
	s.State = CallStateEnded
	s.EndReason = reason
	s.EndedAt = cs.now()
	cs.stopTimer(s)
	delete(cs.userCall, s.CallerID)
	delete(cs.userCall, s.CalleeID)
	delete(cs.calls, s.ID)
}

func (cs *CallService) stopTimer(s *CallSession) {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
}

// nextID 生成通话 ID（持锁调用）。
func (cs *CallService) nextID() string {
	cs.idSeq++
	return fmt.Sprintf("call_%d_%d", cs.now().UnixNano(), cs.idSeq)
}
