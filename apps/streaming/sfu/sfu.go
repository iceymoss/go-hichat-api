package sfu

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

// SFU SFU实现
type SFU struct {
	rooms           map[string]*RoomSFU
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	maxRooms        int
	maxUsersPerRoom int
}

// RoomSFU 房间SFU
type RoomSFU struct {
	roomID    string
	users     map[string]*UserSFU
	mu        sync.RWMutex
	createdAt time.Time
	updatedAt time.Time
}

// UserSFU 用户SFU
type UserSFU struct {
	userID      string
	peerConn    *webrtc.PeerConnection
	tracks      map[string]*webrtc.TrackLocalStaticRTP
	mu          sync.RWMutex
	connectedAt time.Time
}

// NewSFU 创建新的SFU
func NewSFU(maxRooms, maxUsersPerRoom int) *SFU {
	ctx, cancel := context.WithCancel(context.Background())

	sfu := &SFU{
		rooms:           make(map[string]*RoomSFU),
		ctx:             ctx,
		cancel:          cancel,
		maxRooms:        maxRooms,
		maxUsersPerRoom: maxUsersPerRoom,
	}

	// 启动清理协程
	go sfu.cleanupRoutine()

	return sfu
}

// HandleMediaStream 处理媒体流
func (s *SFU) HandleMediaStream(roomID, userID string, track *webrtc.TrackLocalStaticRTP) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		room = &RoomSFU{
			roomID:    roomID,
			users:     make(map[string]*UserSFU),
			createdAt: time.Now(),
			updatedAt: time.Now(),
		}
		s.rooms[roomID] = room
	}

	user, exists := room.users[userID]
	if !exists {
		user = &UserSFU{
			userID:      userID,
			tracks:      make(map[string]*webrtc.TrackLocalStaticRTP),
			connectedAt: time.Now(),
		}
		room.users[userID] = user
	}

	user.mu.Lock()
	user.tracks[track.ID()] = track
	user.mu.Unlock()

	room.updatedAt = time.Now()

	zLog.Info("Media stream handled",
		zap.String("room_id", roomID),
		zap.String("user_id", userID),
		zap.String("track_id", track.ID()),
		zap.String("kind", track.Kind().String()))

	return nil
}

// ForwardMediaStream 转发媒体流
func (s *SFU) ForwardMediaStream(roomID, fromUserID, toUserID string, track *webrtc.TrackLocalStaticRTP) error {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	room.mu.RLock()
	fromUser, exists := room.users[fromUserID]
	room.mu.RUnlock()

	if !exists {
		return fmt.Errorf("user %s not found in room %s", fromUserID, roomID)
	}

	room.mu.RLock()
	toUser, exists := room.users[toUserID]
	room.mu.RUnlock()

	if !exists {
		return fmt.Errorf("user %s not found in room %s", toUserID, roomID)
	}

	// 检查发送者是否有该轨道
	fromUser.mu.RLock()
	_, hasTrack := fromUser.tracks[track.ID()]
	fromUser.mu.RUnlock()

	if !hasTrack {
		return fmt.Errorf("user %s does not have track %s", fromUserID, track.ID())
	}

	// 转发轨道到接收者
	if toUser.peerConn != nil {
		_, err := toUser.peerConn.AddTrack(track)
		if err != nil {
			return fmt.Errorf("failed to add track to user %s: %w", toUserID, err)
		}
	}

	zLog.Debug("Media stream forwarded",
		zap.String("room_id", roomID),
		zap.String("from_user", fromUserID),
		zap.String("to_user", toUserID),
		zap.String("track_id", track.ID()))

	return nil
}

// StopForwarding 停止转发
func (s *SFU) StopForwarding(roomID, userID string) error {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	room.mu.Lock()
	user, exists := room.users[userID]
	if exists {
		// 停止所有轨道
		user.mu.Lock()
		for trackID, track := range user.tracks {
			if err := track.Close(); err != nil {
				zLog.Error("Failed to close track",
					zap.String("track_id", trackID),
					zap.Error(err))
			}
		}
		user.tracks = make(map[string]*webrtc.TrackLocalStaticRTP)
		user.mu.Unlock()

		// 关闭PeerConnection
		if user.peerConn != nil {
			if err := user.peerConn.Close(); err != nil {
				zLog.Error("Failed to close peer connection",
					zap.String("user_id", userID),
					zap.Error(err))
			}
		}

		delete(room.users, userID)
		room.updatedAt = time.Now()
	}
	room.mu.Unlock()

	zLog.Info("Forwarding stopped for user",
		zap.String("room_id", roomID),
		zap.String("user_id", userID))

	return nil
}

// GetRoomStats 获取房间统计信息
func (s *SFU) GetRoomStats(roomID string) (*types.RoomStats, error) {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("room %s not found", roomID)
	}

	room.mu.RLock()
	userCount := len(room.users)
	activeStreams := 0
	bandwidth := int64(0)

	for _, user := range room.users {
		user.mu.RLock()
		activeStreams += len(user.tracks)
		// 这里可以计算实际的带宽使用
		// 简化处理，假设每个轨道使用固定带宽
		bandwidth += int64(len(user.tracks) * 1000000) // 1Mbps per track
		user.mu.RUnlock()
	}
	room.mu.RUnlock()

	return &types.RoomStats{
		RoomID:        roomID,
		UserCount:     userCount,
		ActiveStreams: activeStreams,
		Bandwidth:     bandwidth,
		UpdatedAt:     time.Now(),
	}, nil
}

// AddUserToRoom 添加用户到房间
func (s *SFU) AddUserToRoom(roomID, userID string, peerConn *webrtc.PeerConnection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		room = &RoomSFU{
			roomID:    roomID,
			users:     make(map[string]*UserSFU),
			createdAt: time.Now(),
			updatedAt: time.Now(),
		}
		s.rooms[roomID] = room
	}

	if len(room.users) >= s.maxUsersPerRoom {
		return fmt.Errorf("room %s is full (max %d users)", roomID, s.maxUsersPerRoom)
	}

	user := &UserSFU{
		userID:      userID,
		peerConn:    peerConn,
		tracks:      make(map[string]*webrtc.TrackLocalStaticRTP),
		connectedAt: time.Now(),
	}

	room.users[userID] = user
	room.updatedAt = time.Now()

	zLog.Info("User added to SFU room",
		zap.String("room_id", roomID),
		zap.String("user_id", userID),
		zap.Int("user_count", len(room.users)))

	return nil
}

// RemoveUserFromRoom 从房间移除用户
func (s *SFU) RemoveUserFromRoom(roomID, userID string) error {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	room.mu.Lock()
	user, exists := room.users[userID]
	if exists {
		// 停止所有轨道
		user.mu.Lock()
		for trackID, track := range user.tracks {
			if err := track.Close(); err != nil {
				zLog.Error("Failed to close track",
					zap.String("track_id", trackID),
					zap.Error(err))
			}
		}
		user.mu.Unlock()

		// 关闭PeerConnection
		if user.peerConn != nil {
			if err := user.peerConn.Close(); err != nil {
				zLog.Error("Failed to close peer connection",
					zap.String("user_id", userID),
					zap.Error(err))
			}
		}

		delete(room.users, userID)
		room.updatedAt = time.Now()
	}
	room.mu.Unlock()

	zLog.Info("User removed from SFU room",
		zap.String("room_id", roomID),
		zap.String("user_id", userID))

	return nil
}

// GetRoomUserCount 获取房间用户数量
func (s *SFU) GetRoomUserCount(roomID string) (int, error) {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("room %s not found", roomID)
	}

	room.mu.RLock()
	count := len(room.users)
	room.mu.RUnlock()

	return count, nil
}

// cleanupRoutine 清理协程
func (s *SFU) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup 清理过期房间和用户
func (s *SFU) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiredRooms := make([]string, 0)

	for roomID, room := range s.rooms {
		room.mu.RLock()
		userCount := len(room.users)
		room.mu.RUnlock()

		// 如果房间为空且超过5分钟未更新，则删除
		if userCount == 0 && now.Sub(room.updatedAt) > 5*time.Minute {
			expiredRooms = append(expiredRooms, roomID)
		}
	}

	// 删除过期房间
	for _, roomID := range expiredRooms {
		delete(s.rooms, roomID)
		zLog.Info("Expired SFU room cleaned up", zap.String("room_id", roomID))
	}

	if len(expiredRooms) > 0 {
		zLog.Info("SFU cleanup completed",
			zap.Int("cleaned_rooms", len(expiredRooms)),
			zap.Int("remaining_rooms", len(s.rooms)))
	}
}

// Close 关闭SFU
func (s *SFU) Close() error {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	// 关闭所有房间
	for roomID, room := range s.rooms {
		room.mu.Lock()
		for userID, user := range room.users {
			// 关闭所有轨道
			user.mu.Lock()
			for trackID, track := range user.tracks {
				if err := track.Close(); err != nil {
					zLog.Error("Failed to close track",
						zap.String("room_id", roomID),
						zap.String("user_id", userID),
						zap.String("track_id", trackID),
						zap.Error(err))
				}
			}
			user.mu.Unlock()

			// 关闭PeerConnection
			if user.peerConn != nil {
				if err := user.peerConn.Close(); err != nil {
					zLog.Error("Failed to close peer connection",
						zap.String("room_id", roomID),
						zap.String("user_id", userID),
						zap.Error(err))
				}
			}
		}
		room.mu.Unlock()
	}

	s.rooms = make(map[string]*RoomSFU)

	zLog.Info("SFU closed")
	return nil
}
