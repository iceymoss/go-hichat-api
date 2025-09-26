package room

import (
	"fmt"
	"sync"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/types"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// RoomManager 房间管理器实现
type RoomManager struct {
	rooms       map[string]*Room
	mu          sync.RWMutex
	maxRooms    int
	roomTimeout time.Duration
}

// NewRoomManager 创建房间管理器
func NewRoomManager(maxRooms int, roomTimeout time.Duration) *RoomManager {
	return &RoomManager{
		rooms:       make(map[string]*Room),
		maxRooms:    maxRooms,
		roomTimeout: roomTimeout,
	}
}

// CreateRoom 创建房间
func (rm *RoomManager) CreateRoom(roomID, name string) (*Room, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if roomID == "" {
		return nil, fmt.Errorf("room ID cannot be empty")
	}

	if name == "" {
		name = fmt.Sprintf("Room %s", roomID)
	}

	if _, exists := rm.rooms[roomID]; exists {
		return nil, fmt.Errorf("room %s already exists", roomID)
	}

	if len(rm.rooms) >= rm.maxRooms {
		return nil, fmt.Errorf("maximum number of rooms (%d) reached", rm.maxRooms)
	}

	room := NewRoom(roomID, name)
	rm.rooms[roomID] = room

	zLog.Info("Room created",
		zap.String("room_id", roomID),
		zap.String("name", name),
		zap.Int("total_rooms", len(rm.rooms)))

	return room, nil
}

// GetRoom 获取房间
func (rm *RoomManager) GetRoom(roomID string) (*Room, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room %s not found", roomID)
	}

	return room, nil
}

// DeleteRoom 删除房间
func (rm *RoomManager) DeleteRoom(roomID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	// 检查房间是否为空
	if !room.IsEmpty() {
		return fmt.Errorf("cannot delete non-empty room %s", roomID)
	}

	delete(rm.rooms, roomID)

	zLog.Info("Room deleted",
		zap.String("room_id", roomID),
		zap.Int("total_rooms", len(rm.rooms)))

	return nil
}

// GetAllRooms 获取所有房间
func (rm *RoomManager) GetAllRooms() map[string]*Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// 返回房间的副本
	rooms := make(map[string]*Room)
	for id, room := range rm.rooms {
		rooms[id] = room
	}

	return rooms
}

// CleanupExpiredRooms 清理过期房间
func (rm *RoomManager) CleanupExpiredRooms() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()
	expiredRooms := make([]string, 0)

	for roomID, room := range rm.rooms {
		// 检查房间是否过期且为空
		if room.IsEmpty() && now.Sub(room.GetUpdatedAt()) > rm.roomTimeout {
			expiredRooms = append(expiredRooms, roomID)
		}
	}

	// 删除过期房间
	for _, roomID := range expiredRooms {
		delete(rm.rooms, roomID)
		zLog.Info("Expired room cleaned up", zap.String("room_id", roomID))
	}

	if len(expiredRooms) > 0 {
		zLog.Info("Room cleanup completed",
			zap.Int("cleaned_rooms", len(expiredRooms)),
			zap.Int("remaining_rooms", len(rm.rooms)))
	}

	return nil
}

// GetRoomCount 获取房间数量
func (rm *RoomManager) GetRoomCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.rooms)
}

// GetTotalUserCount 获取总用户数量
func (rm *RoomManager) GetTotalUserCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	totalUsers := 0
	for _, room := range rm.rooms {
		totalUsers += room.GetUserCount()
	}

	return totalUsers
}

// GetRoomStats 获取房间统计信息
func (rm *RoomManager) GetRoomStats() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := map[string]interface{}{
		"total_rooms":  len(rm.rooms),
		"total_users":  0,
		"empty_rooms":  0,
		"active_rooms": 0,
		"max_rooms":    rm.maxRooms,
		"room_timeout": rm.roomTimeout.String(),
	}

	totalUsers := 0
	emptyRooms := 0
	activeRooms := 0

	for _, room := range rm.rooms {
		userCount := room.GetUserCount()
		totalUsers += userCount

		if userCount == 0 {
			emptyRooms++
		} else {
			activeRooms++
		}
	}

	stats["total_users"] = totalUsers
	stats["empty_rooms"] = emptyRooms
	stats["active_rooms"] = activeRooms

	return stats
}

// JoinRoom 用户加入房间
func (rm *RoomManager) JoinRoom(roomID string, user *types.User) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	return room.AddUser(user)
}

// LeaveRoom 用户离开房间
func (rm *RoomManager) LeaveRoom(roomID, userID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	return room.RemoveUser(userID)
}

// GetUserRoom 获取用户所在房间
func (rm *RoomManager) GetUserRoom(userID string) (*Room, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for _, room := range rm.rooms {
		if _, err := room.GetUser(userID); err == nil {
			return room, nil
		}
	}

	return nil, fmt.Errorf("user %s not found in any room", userID)
}

// UpdateRoomStatus 更新房间状态
func (rm *RoomManager) UpdateRoomStatus(roomID string) error {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	return room.UpdateStatus()
}

// GetRoomInfo 获取房间信息
func (rm *RoomManager) GetRoomInfo(roomID string) (*types.RoomInfo, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room %s not found", roomID)
	}

	return room.GetRoomInfo(), nil
}
