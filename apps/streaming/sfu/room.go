// Package sfu 实现 apps/streaming 进程内的群通话 SFU（Selective Forwarding Unit）核心。
//
// 设计上把「路由簿记 / fan-out 决策」（本文件，纯逻辑、可单测）与「pion PeerConnection /
// TrackRemote I/O」（后续文件）分离：Room 只维护成员与订阅路由表，下行投递抽象成 RTPSink
// 接口（*webrtc.TrackLocalStaticRTP 天然满足），媒体读写泵在上层把真实 track 接进来。
package sfu

import "sync"

// Room 一通群通话的路由核心：维护参与者集合与「发布者 -> 订阅者下行 sink」路由表，
// 负责把某发布者的 RTP 包 fan-out 给其余订阅者。并发安全（所有访问经 mu）。
type Room struct {
	id string

	mu           sync.RWMutex
	participants map[string]struct{} // uid 集合
}

// NewRoom 创建空房间。
func NewRoom(id string) *Room {
	return &Room{
		id:           id,
		participants: make(map[string]struct{}),
	}
}

// ID 返回房间标识（= callID）。
func (r *Room) ID() string { return r.id }

// Join 加入房间（幂等）。
func (r *Room) Join(uid string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.participants[uid] = struct{}{}
}

// Leave 离开房间（未知成员为 no-op）。
func (r *Room) Leave(uid string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.participants, uid)
}

// Participants 返回当前成员 uid 列表（顺序不保证）。
func (r *Room) Participants() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.participants))
	for uid := range r.participants {
		out = append(out, uid)
	}
	return out
}
