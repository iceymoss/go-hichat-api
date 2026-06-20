package logic

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// 群组通话相关错误
var (
	ErrGroupCallNotFound = errors.New("group call not found")
	ErrGroupCallFull     = errors.New("group call is full")
)

// GroupCall 一通群组（Mesh）通话的会话状态。
// Mesh：参与者两两 P2P 直连；本结构只管成员/邀请状态，媒体连接在前端按成员两两建立。
type GroupCall struct {
	ID          string
	Type        CallType
	GroupID     string
	InitiatorID string

	participants map[string]struct{} // 已加入
	invited      map[string]struct{} // 已邀请待接听
	createdAt    time.Time
}

// GroupCallService 管理群组通话会话（与 1:1 的 CallService 并存，互不影响）。
type GroupCallService struct {
	mu       sync.Mutex
	calls    map[string]*GroupCall
	userCall map[string]string // uid -> callId（忙线占用，群内）
	maxPart  int
	now      func() time.Time
	idSeq    uint64
}

// NewGroupCallService 创建群通话服务，maxParticipants<=0 时默认 4（Mesh 上限）。
func NewGroupCallService(maxParticipants int) *GroupCallService {
	if maxParticipants <= 0 {
		maxParticipants = 4
	}
	return &GroupCallService{
		calls:    make(map[string]*GroupCall),
		userCall: make(map[string]string),
		maxPart:  maxParticipants,
		now:      time.Now,
	}
}

// Create 发起群通话：发起人即加入，invited 为待振铃成员。返回新建会话快照信息。
func (s *GroupCallService) Create(initiator, groupID string, t CallType, invited []string) (callID string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, busy := s.userCall[initiator]; busy {
		return "", ErrBusy
	}

	s.idSeq++
	g := &GroupCall{
		ID:           fmt.Sprintf("gcall_%d_%d", s.now().UnixNano(), s.idSeq),
		Type:         t,
		GroupID:      groupID,
		InitiatorID:  initiator,
		participants: map[string]struct{}{initiator: {}},
		invited:      make(map[string]struct{}),
		createdAt:    s.now(),
	}
	for _, u := range invited {
		if u != initiator {
			g.invited[u] = struct{}{}
		}
	}
	s.calls[g.ID] = g
	s.userCall[initiator] = g.ID
	return g.ID, nil
}

// Join 加入群通话。返回加入后「其他已在房间的参与者」列表（供新人 UI + 老人建连）。
func (s *GroupCallService) Join(callID, uid string) (others []string, t CallType, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.calls[callID]
	if !ok {
		return nil, "", ErrGroupCallNotFound
	}
	if _, in := g.participants[uid]; in {
		return othersLocked(g, uid), g.Type, nil // 幂等
	}
	if len(g.participants) >= s.maxPart {
		return nil, "", ErrGroupCallFull
	}
	if _, busy := s.userCall[uid]; busy {
		return nil, "", ErrBusy
	}
	g.participants[uid] = struct{}{}
	delete(g.invited, uid)
	s.userCall[uid] = callID
	return othersLocked(g, uid), g.Type, nil
}

// Leave 离开群通话。返回剩余参与者 + 是否已结束（<2 人）。
func (s *GroupCallService) Leave(callID, uid string) (remaining []string, ended bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.calls[callID]
	if !ok {
		return nil, false, ErrGroupCallNotFound
	}
	delete(g.participants, uid)
	delete(s.userCall, uid)

	ended = len(g.participants) < 2
	if ended {
		for u := range g.participants {
			delete(s.userCall, u)
		}
		delete(s.calls, callID)
	}
	return participantsLocked(g), ended, nil
}

// HasParticipant 判断 uid 是否为 callID 的已加入参与者。
func (s *GroupCallService) HasParticipant(callID, uid string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.calls[callID]
	if !ok {
		return false
	}
	_, in := g.participants[uid]
	return in
}

// Participants 返回 callID 当前参与者快照。
func (s *GroupCallService) Participants(callID string) ([]string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.calls[callID]
	if !ok {
		return nil, false
	}
	return participantsLocked(g), true
}

// Info 返回会话基本信息快照（用于振铃/记录）。
func (s *GroupCallService) Info(callID string) (groupID string, t CallType, initiator string, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, exists := s.calls[callID]
	if !exists {
		return "", "", "", false
	}
	return g.GroupID, g.Type, g.InitiatorID, true
}

// ActiveCallID 返回用户当前所在群通话 ID（断线清理用）。
func (s *GroupCallService) ActiveCallID(uid string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.userCall[uid]
	return id, ok
}

// IsBusy 判断用户是否在群通话中。
func (s *GroupCallService) IsBusy(uid string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, busy := s.userCall[uid]
	return busy
}

func othersLocked(g *GroupCall, uid string) []string {
	out := make([]string, 0, len(g.participants))
	for u := range g.participants {
		if u != uid {
			out = append(out, u)
		}
	}
	return out
}

func participantsLocked(g *GroupCall) []string {
	out := make([]string, 0, len(g.participants))
	for u := range g.participants {
		out = append(out, u)
	}
	return out
}
