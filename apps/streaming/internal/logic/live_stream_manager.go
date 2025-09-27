package logic

import (
	"fmt"
	"sync"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/types"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// LiveStreamManager 直播管理器
type LiveStreamManager struct {
	// 活跃直播
	activeStreams map[string]*types.LiveStream
	// 用户直播映射
	userStreams map[string]string // streamerID -> streamID
	// 观众映射
	viewers map[string][]string // streamID -> []viewerID
	// 互斥锁
	mu sync.RWMutex
}

// NewLiveStreamManager 创建直播管理器
func NewLiveStreamManager() *LiveStreamManager {
	return &LiveStreamManager{
		activeStreams: make(map[string]*types.LiveStream),
		userStreams:   make(map[string]string),
		viewers:       make(map[string][]string),
	}
}

// StartLiveStream 开始直播
func (lsm *LiveStreamManager) StartLiveStream(streamerID, title, description string) (*types.LiveStream, error) {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	// 检查用户是否已在直播
	if _, exists := lsm.userStreams[streamerID]; exists {
		return nil, fmt.Errorf("user %s is already streaming", streamerID)
	}

	// 创建直播
	liveStream := &types.LiveStream{
		ID:          generateLiveStreamID(),
		StreamerID:  streamerID,
		Title:       title,
		Description: description,
		Status:      "active",
		ViewerCount: 0,
		StartedAt:   time.Now(),
		StreamURL:   generateStreamURL(streamerID),
		CDNURL:      generateCDNURL(streamerID),
	}

	// 保存直播
	lsm.activeStreams[liveStream.ID] = liveStream
	lsm.userStreams[streamerID] = liveStream.ID
	lsm.viewers[liveStream.ID] = make([]string, 0)

	zLog.Info("Live stream started",
		zap.String("stream_id", liveStream.ID),
		zap.String("streamer_id", streamerID),
		zap.String("title", title))

	return liveStream, nil
}

// StopLiveStream 停止直播
func (lsm *LiveStreamManager) StopLiveStream(streamerID string) error {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	streamID, exists := lsm.userStreams[streamerID]
	if !exists {
		return fmt.Errorf("user %s is not streaming", streamerID)
	}

	liveStream, exists := lsm.activeStreams[streamID]
	if !exists {
		return fmt.Errorf("live stream %s not found", streamID)
	}

	// 停止直播
	lsm.stopLiveStream(streamID)

	zLog.Info("Live stream stopped",
		zap.String("stream_id", streamID),
		zap.String("streamer_id", streamerID),
		zap.Int("final_viewer_count", liveStream.ViewerCount))

	return nil
}

// JoinLiveStream 加入直播
func (lsm *LiveStreamManager) JoinLiveStream(streamID, viewerID string) (*types.LiveStream, error) {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	liveStream, exists := lsm.activeStreams[streamID]
	if !exists {
		return nil, fmt.Errorf("live stream %s not found", streamID)
	}

	if liveStream.Status != "active" {
		return nil, fmt.Errorf("live stream %s is not active", streamID)
	}

	// 检查观众是否已在观看
	viewerList, exists := lsm.viewers[streamID]
	if !exists {
		viewerList = make([]string, 0)
		lsm.viewers[streamID] = viewerList
	}

	// 检查是否已在观看列表中
	for _, existingViewer := range viewerList {
		if existingViewer == viewerID {
			return liveStream, nil // 已在观看
		}
	}

	// 添加观众
	lsm.viewers[streamID] = append(viewerList, viewerID)
	liveStream.ViewerCount = len(lsm.viewers[streamID])

	zLog.Info("Viewer joined live stream",
		zap.String("stream_id", streamID),
		zap.String("viewer_id", viewerID),
		zap.Int("viewer_count", liveStream.ViewerCount))

	return liveStream, nil
}

// LeaveLiveStream 离开直播
func (lsm *LiveStreamManager) LeaveLiveStream(streamID, viewerID string) error {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	liveStream, exists := lsm.activeStreams[streamID]
	if !exists {
		return fmt.Errorf("live stream %s not found", streamID)
	}

	viewerList, exists := lsm.viewers[streamID]
	if !exists {
		return fmt.Errorf("viewer list for stream %s not found", streamID)
	}

	// 查找并移除观众
	for i, existingViewer := range viewerList {
		if existingViewer == viewerID {
			// 移除观众
			lsm.viewers[streamID] = append(viewerList[:i], viewerList[i+1:]...)
			liveStream.ViewerCount = len(lsm.viewers[streamID])

			zLog.Info("Viewer left live stream",
				zap.String("stream_id", streamID),
				zap.String("viewer_id", viewerID),
				zap.Int("remaining_viewer_count", liveStream.ViewerCount))

			return nil
		}
	}

	return fmt.Errorf("viewer %s is not watching stream %s", viewerID, streamID)
}

// GetLiveStream 获取直播信息
func (lsm *LiveStreamManager) GetLiveStream(streamID string) (*types.LiveStream, error) {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	liveStream, exists := lsm.activeStreams[streamID]
	if !exists {
		return nil, fmt.Errorf("live stream %s not found", streamID)
	}

	return liveStream, nil
}

// GetUserLiveStream 获取用户当前直播
func (lsm *LiveStreamManager) GetUserLiveStream(streamerID string) (*types.LiveStream, error) {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	streamID, exists := lsm.userStreams[streamerID]
	if !exists {
		return nil, fmt.Errorf("user %s is not streaming", streamerID)
	}

	liveStream, exists := lsm.activeStreams[streamID]
	if !exists {
		return nil, fmt.Errorf("live stream %s not found", streamID)
	}

	return liveStream, nil
}

// GetActiveLiveStreams 获取所有活跃直播
func (lsm *LiveStreamManager) GetActiveLiveStreams() []*types.LiveStream {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	streams := make([]*types.LiveStream, 0, len(lsm.activeStreams))
	for _, stream := range lsm.activeStreams {
		streams = append(streams, stream)
	}

	return streams
}

// GetStreamViewers 获取直播观众列表
func (lsm *LiveStreamManager) GetStreamViewers(streamID string) ([]string, error) {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	viewers, exists := lsm.viewers[streamID]
	if !exists {
		return nil, fmt.Errorf("viewer list for stream %s not found", streamID)
	}

	// 返回观众列表的副本
	viewerList := make([]string, len(viewers))
	copy(viewerList, viewers)

	return viewerList, nil
}

// UpdateLiveStreamInfo 更新直播信息
func (lsm *LiveStreamManager) UpdateLiveStreamInfo(streamerID, title, description string) error {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	streamID, exists := lsm.userStreams[streamerID]
	if !exists {
		return fmt.Errorf("user %s is not streaming", streamerID)
	}

	liveStream, exists := lsm.activeStreams[streamID]
	if !exists {
		return fmt.Errorf("live stream %s not found", streamID)
	}

	// 更新信息
	liveStream.Title = title
	liveStream.Description = description

	zLog.Info("Live stream info updated",
		zap.String("stream_id", streamID),
		zap.String("streamer_id", streamerID),
		zap.String("title", title))

	return nil
}

// PauseLiveStream 暂停直播
func (lsm *LiveStreamManager) PauseLiveStream(streamerID string) error {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	streamID, exists := lsm.userStreams[streamerID]
	if !exists {
		return fmt.Errorf("user %s is not streaming", streamerID)
	}

	liveStream, exists := lsm.activeStreams[streamID]
	if !exists {
		return fmt.Errorf("live stream %s not found", streamID)
	}

	if liveStream.Status != "active" {
		return fmt.Errorf("live stream %s is not active", streamID)
	}

	// 暂停直播
	liveStream.Status = "paused"

	zLog.Info("Live stream paused",
		zap.String("stream_id", streamID),
		zap.String("streamer_id", streamerID))

	return nil
}

// ResumeLiveStream 恢复直播
func (lsm *LiveStreamManager) ResumeLiveStream(streamerID string) error {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	streamID, exists := lsm.userStreams[streamerID]
	if !exists {
		return fmt.Errorf("user %s is not streaming", streamerID)
	}

	liveStream, exists := lsm.activeStreams[streamID]
	if !exists {
		return fmt.Errorf("live stream %s not found", streamID)
	}

	if liveStream.Status != "paused" {
		return fmt.Errorf("live stream %s is not paused", streamID)
	}

	// 恢复直播
	liveStream.Status = "active"

	zLog.Info("Live stream resumed",
		zap.String("stream_id", streamID),
		zap.String("streamer_id", streamerID))

	return nil
}

// stopLiveStream 停止直播（内部方法）
func (lsm *LiveStreamManager) stopLiveStream(streamID string) {
	liveStream, exists := lsm.activeStreams[streamID]
	if !exists {
		return
	}

	// 更新直播状态
	liveStream.Status = "ended"
	now := time.Now()
	liveStream.EndedAt = &now

	// 清理用户映射
	delete(lsm.userStreams, liveStream.StreamerID)

	// 清理观众映射
	delete(lsm.viewers, streamID)

	// 删除直播
	delete(lsm.activeStreams, streamID)

	zLog.Info("Live stream stopped and cleaned up",
		zap.String("stream_id", streamID),
		zap.String("streamer_id", liveStream.StreamerID),
		zap.Int("final_viewer_count", liveStream.ViewerCount))
}

// generateLiveStreamID 生成直播ID
func generateLiveStreamID() string {
	return fmt.Sprintf("live_%d", time.Now().UnixNano())
}

// generateStreamURL 生成推流URL
func generateStreamURL(streamerID string) string {
	return fmt.Sprintf("rtmp://streaming.example.com/live/%s", streamerID)
}

// generateCDNURL 生成CDN分发URL
func generateCDNURL(streamerID string) string {
	return fmt.Sprintf("https://cdn.example.com/live/%s.m3u8", streamerID)
}
