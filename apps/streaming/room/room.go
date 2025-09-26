package room

import (
	"fmt"
	"sync"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/types"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// Room 房间实现
type Room struct {
	id        string
	name      string
	users     map[string]*types.User
	mu        sync.RWMutex
	createdAt time.Time
	updatedAt time.Time
}

// NewRoom 创建新房间
func NewRoom(id, name string) *Room {
	now := time.Now()
	return &Room{
		id:        id,
		name:      name,
		users:     make(map[string]*types.User),
		createdAt: now,
		updatedAt: now,
	}
}

// GetID 获取房间ID
func (r *Room) GetID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.id
}

// GetName 获取房间名称
func (r *Room) GetName() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.name
}

// AddUser 添加用户
func (r *Room) AddUser(user *types.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if _, exists := r.users[user.UserID]; exists {
		return fmt.Errorf("user %s already exists in room %s", user.UserID, r.id)
	}

	user.JoinedAt = time.Now()
	r.users[user.UserID] = user
	r.updatedAt = time.Now()

	zLog.Info("User added to room",
		zap.String("room_id", r.id),
		zap.String("user_id", user.UserID),
		zap.String("username", user.Username))

	return nil
}

// RemoveUser 移除用户
func (r *Room) RemoveUser(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[userID]; !exists {
		return fmt.Errorf("user %s not found in room %s", userID, r.id)
	}

	delete(r.users, userID)
	r.updatedAt = time.Now()

	zLog.Info("User removed from room",
		zap.String("room_id", r.id),
		zap.String("user_id", userID))

	return nil
}

// GetUser 获取用户
func (r *Room) GetUser(userID string) (*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, fmt.Errorf("user %s not found in room %s", userID, r.id)
	}

	return user, nil
}

// GetUsers 获取所有用户
func (r *Room) GetUsers() []*types.User {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*types.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users
}

// Broadcast 广播消息
func (r *Room) Broadcast(message *types.SignalingMessage) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	// 这里应该通过连接管理器发送消息给房间内的所有用户
	// 暂时只记录日志
	zLog.Info("Broadcasting message to room",
		zap.String("room_id", r.id),
		zap.String("message_type", string(message.Type)),
		zap.Int("user_count", len(r.users)))

	return nil
}

// SendToUser 发送消息给特定用户
func (r *Room) SendToUser(userID string, message *types.SignalingMessage) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	if _, exists := r.users[userID]; !exists {
		return fmt.Errorf("user %s not found in room %s", userID, r.id)
	}

	// 这里应该通过连接管理器发送消息给特定用户
	// 暂时只记录日志
	zLog.Info("Sending message to user in room",
		zap.String("room_id", r.id),
		zap.String("user_id", userID),
		zap.String("message_type", string(message.Type)))

	return nil
}

// GetUserCount 获取用户数量
func (r *Room) GetUserCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.users)
}

// IsEmpty 检查房间是否为空
func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.users) == 0
}

// UpdateStatus 更新房间状态
func (r *Room) UpdateStatus() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.updatedAt = time.Now()

	zLog.Debug("Room status updated",
		zap.String("room_id", r.id),
		zap.Int("user_count", len(r.users)),
		zap.Time("updated_at", r.updatedAt))

	return nil
}

// GetRoomInfo 获取房间信息
func (r *Room) GetRoomInfo() *types.RoomInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]types.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, *user)
	}

	return &types.RoomInfo{
		RoomID:    r.id,
		Name:      r.name,
		Users:     users,
		CreatedAt: r.createdAt,
		UpdatedAt: r.updatedAt,
	}
}

// UpdateUserStatus 更新用户状态
func (r *Room) UpdateUserStatus(userID string, isMuted, isVideoOn bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return fmt.Errorf("user %s not found in room %s", userID, r.id)
	}

	user.IsMuted = isMuted
	user.IsVideoOn = isVideoOn
	r.updatedAt = time.Now()

	zLog.Info("User status updated in room",
		zap.String("room_id", r.id),
		zap.String("user_id", userID),
		zap.Bool("is_muted", isMuted),
		zap.Bool("is_video_on", isVideoOn))

	return nil
}

// GetCreatedAt 获取创建时间
func (r *Room) GetCreatedAt() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.createdAt
}

// GetUpdatedAt 获取更新时间
func (r *Room) GetUpdatedAt() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.updatedAt
}
