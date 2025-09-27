package logic

import (
	"fmt"
	"sync"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/types"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// ScreenShareManager 录屏管理器
type ScreenShareManager struct {
	// 活跃录屏
	activeScreenShares map[string]*types.ScreenShare
	// 用户录屏映射
	userScreenShares map[string]string // userID -> screenShareID
	// 互斥锁
	mu sync.RWMutex
}

// NewScreenShareManager 创建录屏管理器
func NewScreenShareManager() *ScreenShareManager {
	return &ScreenShareManager{
		activeScreenShares: make(map[string]*types.ScreenShare),
		userScreenShares:   make(map[string]string),
	}
}

// StartScreenShare 开始录屏
func (ssm *ScreenShareManager) StartScreenShare(userID, roomID, quality string) (*types.ScreenShare, error) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	// 检查用户是否已在录屏
	if _, exists := ssm.userScreenShares[userID]; exists {
		return nil, fmt.Errorf("user %s is already screen sharing", userID)
	}

	// 创建录屏
	screenShare := &types.ScreenShare{
		ID:        generateScreenShareID(),
		UserID:    userID,
		RoomID:    roomID,
		Status:    "active",
		StartedAt: time.Now(),
		Quality:   quality,
	}

	// 保存录屏
	ssm.activeScreenShares[screenShare.ID] = screenShare
	ssm.userScreenShares[userID] = screenShare.ID

	zLog.Info("Screen share started",
		zap.String("screen_share_id", screenShare.ID),
		zap.String("user_id", userID),
		zap.String("room_id", roomID),
		zap.String("quality", quality))

	return screenShare, nil
}

// StopScreenShare 停止录屏
func (ssm *ScreenShareManager) StopScreenShare(userID string) error {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	screenShareID, exists := ssm.userScreenShares[userID]
	if !exists {
		return fmt.Errorf("user %s is not screen sharing", userID)
	}

	screenShare, exists := ssm.activeScreenShares[screenShareID]
	if !exists {
		return fmt.Errorf("screen share %s not found", screenShareID)
	}

	// 停止录屏
	ssm.stopScreenShare(screenShareID)

	zLog.Info("Screen share stopped",
		zap.String("screen_share_id", screenShareID),
		zap.String("user_id", userID),
		zap.Int64("duration", screenShare.Duration))

	return nil
}

// PauseScreenShare 暂停录屏
func (ssm *ScreenShareManager) PauseScreenShare(userID string) error {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	screenShareID, exists := ssm.userScreenShares[userID]
	if !exists {
		return fmt.Errorf("user %s is not screen sharing", userID)
	}

	screenShare, exists := ssm.activeScreenShares[screenShareID]
	if !exists {
		return fmt.Errorf("screen share %s not found", screenShareID)
	}

	if screenShare.Status != "active" {
		return fmt.Errorf("screen share %s is not active", screenShareID)
	}

	// 暂停录屏
	screenShare.Status = "paused"

	zLog.Info("Screen share paused",
		zap.String("screen_share_id", screenShareID),
		zap.String("user_id", userID))

	return nil
}

// ResumeScreenShare 恢复录屏
func (ssm *ScreenShareManager) ResumeScreenShare(userID string) error {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	screenShareID, exists := ssm.userScreenShares[userID]
	if !exists {
		return fmt.Errorf("user %s is not screen sharing", userID)
	}

	screenShare, exists := ssm.activeScreenShares[screenShareID]
	if !exists {
		return fmt.Errorf("screen share %s not found", screenShareID)
	}

	if screenShare.Status != "paused" {
		return fmt.Errorf("screen share %s is not paused", screenShareID)
	}

	// 恢复录屏
	screenShare.Status = "active"

	zLog.Info("Screen share resumed",
		zap.String("screen_share_id", screenShareID),
		zap.String("user_id", userID))

	return nil
}

// GetScreenShare 获取录屏信息
func (ssm *ScreenShareManager) GetScreenShare(screenShareID string) (*types.ScreenShare, error) {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	screenShare, exists := ssm.activeScreenShares[screenShareID]
	if !exists {
		return nil, fmt.Errorf("screen share %s not found", screenShareID)
	}

	return screenShare, nil
}

// GetUserScreenShare 获取用户当前录屏
func (ssm *ScreenShareManager) GetUserScreenShare(userID string) (*types.ScreenShare, error) {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	screenShareID, exists := ssm.userScreenShares[userID]
	if !exists {
		return nil, fmt.Errorf("user %s is not screen sharing", userID)
	}

	screenShare, exists := ssm.activeScreenShares[screenShareID]
	if !exists {
		return nil, fmt.Errorf("screen share %s not found", screenShareID)
	}

	return screenShare, nil
}

// GetRoomScreenShares 获取房间内所有录屏
func (ssm *ScreenShareManager) GetRoomScreenShares(roomID string) []*types.ScreenShare {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	screenShares := make([]*types.ScreenShare, 0)
	for _, screenShare := range ssm.activeScreenShares {
		if screenShare.RoomID == roomID {
			screenShares = append(screenShares, screenShare)
		}
	}

	return screenShares
}

// GetActiveScreenShares 获取所有活跃录屏
func (ssm *ScreenShareManager) GetActiveScreenShares() []*types.ScreenShare {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	screenShares := make([]*types.ScreenShare, 0, len(ssm.activeScreenShares))
	for _, screenShare := range ssm.activeScreenShares {
		screenShares = append(screenShares, screenShare)
	}

	return screenShares
}

// UpdateScreenShareQuality 更新录屏质量
func (ssm *ScreenShareManager) UpdateScreenShareQuality(userID, quality string) error {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	screenShareID, exists := ssm.userScreenShares[userID]
	if !exists {
		return fmt.Errorf("user %s is not screen sharing", userID)
	}

	screenShare, exists := ssm.activeScreenShares[screenShareID]
	if !exists {
		return fmt.Errorf("screen share %s not found", screenShareID)
	}

	// 更新质量
	screenShare.Quality = quality

	zLog.Info("Screen share quality updated",
		zap.String("screen_share_id", screenShareID),
		zap.String("user_id", userID),
		zap.String("quality", quality))

	return nil
}

// RequestScreenShare 请求录屏
func (ssm *ScreenShareManager) RequestScreenShare(requesterID, targetUserID, roomID string) error {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	// 检查目标用户是否已在录屏
	if _, exists := ssm.userScreenShares[targetUserID]; exists {
		return fmt.Errorf("user %s is already screen sharing", targetUserID)
	}

	zLog.Info("Screen share requested",
		zap.String("requester_id", requesterID),
		zap.String("target_user_id", targetUserID),
		zap.String("room_id", roomID))

	return nil
}

// stopScreenShare 停止录屏（内部方法）
func (ssm *ScreenShareManager) stopScreenShare(screenShareID string) {
	screenShare, exists := ssm.activeScreenShares[screenShareID]
	if !exists {
		return
	}

	// 更新录屏状态
	screenShare.Status = "stopped"
	now := time.Now()
	screenShare.StoppedAt = &now
	screenShare.Duration = int64(now.Sub(screenShare.StartedAt).Seconds())

	// 清理用户映射
	delete(ssm.userScreenShares, screenShare.UserID)

	// 删除录屏
	delete(ssm.activeScreenShares, screenShareID)

	zLog.Info("Screen share stopped and cleaned up",
		zap.String("screen_share_id", screenShareID),
		zap.String("user_id", screenShare.UserID),
		zap.Int64("duration", screenShare.Duration))
}

// generateScreenShareID 生成录屏ID
func generateScreenShareID() string {
	return fmt.Sprintf("screenshare_%d", time.Now().UnixNano())
}
