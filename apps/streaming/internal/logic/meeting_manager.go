package logic

import (
	"fmt"
	"sync"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/types"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// MeetingManager 会议管理器
type MeetingManager struct {
	// 活跃会议
	activeMeetings map[string]*types.Meeting
	// 用户会议映射
	userMeetings map[string]string // userID -> meetingID
	// 互斥锁
	mu sync.RWMutex
}

// NewMeetingManager 创建会议管理器
func NewMeetingManager() *MeetingManager {
	return &MeetingManager{
		activeMeetings: make(map[string]*types.Meeting),
		userMeetings:   make(map[string]string),
	}
}

// CreateMeeting 创建会议
func (mm *MeetingManager) CreateMeeting(hostID, title, description string, settings *types.MeetingSettings) (*types.Meeting, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// 检查主持人是否已在会议中
	if _, exists := mm.userMeetings[hostID]; exists {
		return nil, fmt.Errorf("user %s is already in a meeting", hostID)
	}

	// 创建会议
	meeting := &types.Meeting{
		ID:           generateMeetingID(),
		Title:        title,
		Description:  description,
		HostID:       hostID,
		Participants: []*types.User{},
		Status:       "scheduled",
		ScheduledAt:  time.Now(),
		Settings:     settings,
	}

	// 设置默认会议设置
	if meeting.Settings == nil {
		meeting.Settings = &types.MeetingSettings{
			MaxParticipants:  50,
			AllowScreenShare: true,
			AllowRecording:   true,
			MuteOnJoin:       false,
			VideoOnJoin:      true,
			WaitingRoom:      false,
		}
	}

	// 保存会议
	mm.activeMeetings[meeting.ID] = meeting
	mm.userMeetings[hostID] = meeting.ID

	zLog.Info("Meeting created",
		zap.String("meeting_id", meeting.ID),
		zap.String("host_id", hostID),
		zap.String("title", title))

	return meeting, nil
}

// JoinMeeting 加入会议
func (mm *MeetingManager) JoinMeeting(meetingID, userID, username string) (*types.Meeting, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	meeting, exists := mm.activeMeetings[meetingID]
	if !exists {
		return nil, fmt.Errorf("meeting %s not found", meetingID)
	}

	// 检查用户是否已在会议中
	if _, exists := mm.userMeetings[userID]; exists {
		return nil, fmt.Errorf("user %s is already in a meeting", userID)
	}

	// 检查会议是否已满
	if len(meeting.Participants) >= meeting.Settings.MaxParticipants {
		return nil, fmt.Errorf("meeting %s is full", meetingID)
	}

	// 创建参与者
	participant := &types.User{
		UserID:    userID,
		Username:  username,
		JoinedAt:  time.Now(),
		IsMuted:   meeting.Settings.MuteOnJoin,
		IsVideoOn: meeting.Settings.VideoOnJoin,
		Role:      "participant",
		Status:    "online",
	}

	// 添加参与者
	meeting.Participants = append(meeting.Participants, participant)
	mm.userMeetings[userID] = meetingID

	// 如果会议状态是scheduled，更新为ongoing
	if meeting.Status == "scheduled" {
		meeting.Status = "ongoing"
		now := time.Now()
		meeting.StartedAt = &now
	}

	zLog.Info("User joined meeting",
		zap.String("meeting_id", meetingID),
		zap.String("user_id", userID),
		zap.String("username", username),
		zap.Int("participant_count", len(meeting.Participants)))

	return meeting, nil
}

// LeaveMeeting 离开会议
func (mm *MeetingManager) LeaveMeeting(meetingID, userID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	meeting, exists := mm.activeMeetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting %s not found", meetingID)
	}

	// 查找并移除参与者
	for i, participant := range meeting.Participants {
		if participant.UserID == userID {
			// 移除参与者
			meeting.Participants = append(meeting.Participants[:i], meeting.Participants[i+1:]...)
			delete(mm.userMeetings, userID)

			// 如果是主持人离开，结束会议
			if userID == meeting.HostID {
				mm.endMeeting(meetingID)
			}

			zLog.Info("User left meeting",
				zap.String("meeting_id", meetingID),
				zap.String("user_id", userID),
				zap.Int("remaining_participants", len(meeting.Participants)))

			return nil
		}
	}

	return fmt.Errorf("user %s is not a participant of meeting %s", userID, meetingID)
}

// EndMeeting 结束会议
func (mm *MeetingManager) EndMeeting(meetingID, userID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	meeting, exists := mm.activeMeetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting %s not found", meetingID)
	}

	// 检查是否是主持人
	if userID != meeting.HostID {
		return fmt.Errorf("user %s is not the host of meeting %s", userID, meetingID)
	}

	// 结束会议
	mm.endMeeting(meetingID)

	zLog.Info("Meeting ended by host",
		zap.String("meeting_id", meetingID),
		zap.String("host_id", userID))

	return nil
}

// GetMeeting 获取会议信息
func (mm *MeetingManager) GetMeeting(meetingID string) (*types.Meeting, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	meeting, exists := mm.activeMeetings[meetingID]
	if !exists {
		return nil, fmt.Errorf("meeting %s not found", meetingID)
	}

	return meeting, nil
}

// GetUserMeeting 获取用户当前会议
func (mm *MeetingManager) GetUserMeeting(userID string) (*types.Meeting, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	meetingID, exists := mm.userMeetings[userID]
	if !exists {
		return nil, fmt.Errorf("user %s is not in any meeting", userID)
	}

	meeting, exists := mm.activeMeetings[meetingID]
	if !exists {
		return nil, fmt.Errorf("meeting %s not found", meetingID)
	}

	return meeting, nil
}

// GetActiveMeetings 获取所有活跃会议
func (mm *MeetingManager) GetActiveMeetings() []*types.Meeting {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	meetings := make([]*types.Meeting, 0, len(mm.activeMeetings))
	for _, meeting := range mm.activeMeetings {
		meetings = append(meetings, meeting)
	}

	return meetings
}

// UpdateMeetingSettings 更新会议设置
func (mm *MeetingManager) UpdateMeetingSettings(meetingID, userID string, settings *types.MeetingSettings) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	meeting, exists := mm.activeMeetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting %s not found", meetingID)
	}

	// 检查是否是主持人
	if userID != meeting.HostID {
		return fmt.Errorf("user %s is not the host of meeting %s", userID, meetingID)
	}

	// 更新设置
	meeting.Settings = settings

	zLog.Info("Meeting settings updated",
		zap.String("meeting_id", meetingID),
		zap.String("host_id", userID))

	return nil
}

// MuteParticipant 静音参与者
func (mm *MeetingManager) MuteParticipant(meetingID, hostID, participantID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	meeting, exists := mm.activeMeetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting %s not found", meetingID)
	}

	// 检查是否是主持人
	if hostID != meeting.HostID {
		return fmt.Errorf("user %s is not the host of meeting %s", hostID, meetingID)
	}

	// 查找并静音参与者
	for _, participant := range meeting.Participants {
		if participant.UserID == participantID {
			participant.IsMuted = true
			zLog.Info("Participant muted",
				zap.String("meeting_id", meetingID),
				zap.String("participant_id", participantID),
				zap.String("host_id", hostID))
			return nil
		}
	}

	return fmt.Errorf("participant %s not found in meeting %s", participantID, meetingID)
}

// UnmuteParticipant 取消静音参与者
func (mm *MeetingManager) UnmuteParticipant(meetingID, hostID, participantID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	meeting, exists := mm.activeMeetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting %s not found", meetingID)
	}

	// 检查是否是主持人
	if hostID != meeting.HostID {
		return fmt.Errorf("user %s is not the host of meeting %s", hostID, meetingID)
	}

	// 查找并取消静音参与者
	for _, participant := range meeting.Participants {
		if participant.UserID == participantID {
			participant.IsMuted = false
			zLog.Info("Participant unmuted",
				zap.String("meeting_id", meetingID),
				zap.String("participant_id", participantID),
				zap.String("host_id", hostID))
			return nil
		}
	}

	return fmt.Errorf("participant %s not found in meeting %s", participantID, meetingID)
}

// endMeeting 结束会议（内部方法）
func (mm *MeetingManager) endMeeting(meetingID string) {
	meeting, exists := mm.activeMeetings[meetingID]
	if !exists {
		return
	}

	// 更新会议状态
	meeting.Status = "ended"
	now := time.Now()
	meeting.EndedAt = &now

	// 清理用户映射
	for _, participant := range meeting.Participants {
		delete(mm.userMeetings, participant.UserID)
	}

	// 删除会议
	delete(mm.activeMeetings, meetingID)

	zLog.Info("Meeting ended and cleaned up",
		zap.String("meeting_id", meetingID),
		zap.Int("participant_count", len(meeting.Participants)))
}

// generateMeetingID 生成会议ID
func generateMeetingID() string {
	return fmt.Sprintf("meeting_%d", time.Now().UnixNano())
}
