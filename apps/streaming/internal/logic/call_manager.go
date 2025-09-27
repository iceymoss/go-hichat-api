package logic

import (
	"fmt"
	"sync"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/types"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// CallManager 通话管理器
type CallManager struct {
	// 活跃通话
	activeCalls map[string]*types.Call
	// 用户通话映射
	userCalls map[string]string // userID -> callID
	// 互斥锁
	mu sync.RWMutex
}

// NewCallManager 创建通话管理器
func NewCallManager() *CallManager {
	return &CallManager{
		activeCalls: make(map[string]*types.Call),
		userCalls:   make(map[string]string),
	}
}

// CreateOneToOneCall 创建一对一通话
func (cm *CallManager) CreateOneToOneCall(callerID, calleeID string) (*types.Call, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查用户是否已在通话中
	if _, exists := cm.userCalls[callerID]; exists {
		return nil, fmt.Errorf("user %s is already in a call", callerID)
	}
	if _, exists := cm.userCalls[calleeID]; exists {
		return nil, fmt.Errorf("user %s is already in a call", calleeID)
	}

	// 创建通话
	call := &types.Call{
		ID:           generateCallID(),
		Type:         types.CallTypeOneToOne,
		Status:       types.CallStatusInviting,
		CallerID:     callerID,
		Participants: []string{callerID, calleeID},
		RoomID:       generateRoomID(),
		CreatedAt:    time.Now(),
	}

	// 保存通话
	cm.activeCalls[call.ID] = call
	cm.userCalls[callerID] = call.ID
	cm.userCalls[calleeID] = call.ID

	zLog.Info("One-to-one call created",
		zap.String("call_id", call.ID),
		zap.String("caller_id", callerID),
		zap.String("callee_id", calleeID))

	return call, nil
}

// CreateGroupCall 创建群组通话
func (cm *CallManager) CreateGroupCall(callerID string, participants []string) (*types.Call, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查用户是否已在通话中
	if _, exists := cm.userCalls[callerID]; exists {
		return nil, fmt.Errorf("user %s is already in a call", callerID)
	}

	// 检查参与者是否已在通话中
	for _, participantID := range participants {
		if _, exists := cm.userCalls[participantID]; exists {
			return nil, fmt.Errorf("user %s is already in a call", participantID)
		}
	}

	// 创建通话
	call := &types.Call{
		ID:           generateCallID(),
		Type:         types.CallTypeGroup,
		Status:       types.CallStatusInviting,
		CallerID:     callerID,
		Participants: append([]string{callerID}, participants...),
		RoomID:       generateRoomID(),
		CreatedAt:    time.Now(),
	}

	// 保存通话
	cm.activeCalls[call.ID] = call
	for _, participantID := range call.Participants {
		cm.userCalls[participantID] = call.ID
	}

	zLog.Info("Group call created",
		zap.String("call_id", call.ID),
		zap.String("caller_id", callerID),
		zap.Int("participant_count", len(participants)))

	return call, nil
}

// AcceptCall 接受通话
func (cm *CallManager) AcceptCall(callID, userID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	call, exists := cm.activeCalls[callID]
	if !exists {
		return fmt.Errorf("call %s not found", callID)
	}

	if call.Status != types.CallStatusInviting {
		return fmt.Errorf("call %s is not in inviting status", callID)
	}

	// 检查用户是否在参与者列表中
	if !cm.isParticipant(call, userID) {
		return fmt.Errorf("user %s is not a participant of call %s", userID, callID)
	}

	// 更新通话状态
	call.Status = types.CallStatusConnected
	now := time.Now()
	call.StartedAt = &now

	zLog.Info("Call accepted",
		zap.String("call_id", callID),
		zap.String("user_id", userID))

	return nil
}

// RejectCall 拒绝通话
func (cm *CallManager) RejectCall(callID, userID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	call, exists := cm.activeCalls[callID]
	if !exists {
		return fmt.Errorf("call %s not found", callID)
	}

	if call.Status != types.CallStatusInviting {
		return fmt.Errorf("call %s is not in inviting status", callID)
	}

	// 检查用户是否在参与者列表中
	if !cm.isParticipant(call, userID) {
		return fmt.Errorf("user %s is not a participant of call %s", userID, callID)
	}

	// 结束通话
	cm.endCall(callID)

	zLog.Info("Call rejected",
		zap.String("call_id", callID),
		zap.String("user_id", userID))

	return nil
}

// EndCall 结束通话
func (cm *CallManager) EndCall(callID, userID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	call, exists := cm.activeCalls[callID]
	if !exists {
		return fmt.Errorf("call %s not found", callID)
	}

	// 检查用户是否在参与者列表中
	if !cm.isParticipant(call, userID) {
		return fmt.Errorf("user %s is not a participant of call %s", userID, callID)
	}

	// 结束通话
	cm.endCall(callID)

	zLog.Info("Call ended",
		zap.String("call_id", callID),
		zap.String("user_id", userID))

	return nil
}

// GetCall 获取通话信息
func (cm *CallManager) GetCall(callID string) (*types.Call, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	call, exists := cm.activeCalls[callID]
	if !exists {
		return nil, fmt.Errorf("call %s not found", callID)
	}

	return call, nil
}

// GetUserCall 获取用户当前通话
func (cm *CallManager) GetUserCall(userID string) (*types.Call, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	callID, exists := cm.userCalls[userID]
	if !exists {
		return nil, fmt.Errorf("user %s is not in any call", userID)
	}

	call, exists := cm.activeCalls[callID]
	if !exists {
		return nil, fmt.Errorf("call %s not found", callID)
	}

	return call, nil
}

// GetActiveCalls 获取所有活跃通话
func (cm *CallManager) GetActiveCalls() []*types.Call {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	calls := make([]*types.Call, 0, len(cm.activeCalls))
	for _, call := range cm.activeCalls {
		calls = append(calls, call)
	}

	return calls
}

// endCall 结束通话（内部方法）
func (cm *CallManager) endCall(callID string) {
	call, exists := cm.activeCalls[callID]
	if !exists {
		return
	}

	// 更新通话状态
	call.Status = types.CallStatusEnded
	now := time.Now()
	call.EndedAt = &now
	if call.StartedAt != nil {
		call.Duration = int64(now.Sub(*call.StartedAt).Seconds())
	}

	// 清理用户映射
	for _, participantID := range call.Participants {
		delete(cm.userCalls, participantID)
	}

	// 删除通话
	delete(cm.activeCalls, callID)

	zLog.Info("Call ended and cleaned up",
		zap.String("call_id", callID),
		zap.Int64("duration", call.Duration))
}

// isParticipant 检查用户是否是参与者
func (cm *CallManager) isParticipant(call *types.Call, userID string) bool {
	for _, participantID := range call.Participants {
		if participantID == userID {
			return true
		}
	}
	return false
}

// generateCallID 生成通话ID
func generateCallID() string {
	return fmt.Sprintf("call_%d", time.Now().UnixNano())
}

// generateRoomID 生成房间ID
func generateRoomID() string {
	return fmt.Sprintf("room_%d", time.Now().UnixNano())
}
