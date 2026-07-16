// Package sfu 实现 apps/streaming 进程内的群通话 SFU（Selective Forwarding Unit）核心。
//
// 设计上把「路由簿记 / fan-out 决策」（本文件，纯逻辑、可单测）与「pion PeerConnection /
// TrackRemote I/O」（后续文件）分离：Room 只维护成员与订阅路由表，下行投递抽象成 RTPSink
// 接口（*webrtc.TrackLocalStaticRTP 天然满足），媒体读写泵在上层把真实 track 接进来。
package sfu

import (
	"io"
	"sync"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// RTPSink 下行投递目标：把一个 RTP 包写给某订阅者的某条下行轨。
// *webrtc.TrackLocalStaticRTP 天然满足该接口（其 WriteRTP(*rtp.Packet) error）。
type RTPSink interface {
	WriteRTP(*rtp.Packet) error
}

// trackKey 唯一标识一条已发布轨：发布者 uid + track id。
type trackKey struct {
	pubUID  string
	trackID string
}

// PublishedTrack 已发布轨的元信息（供迟到者回填订阅）。Kind 为 "audio"/"video"。
type PublishedTrack struct {
	PubUID  string
	TrackID string
	Kind    string
	Codec   webrtc.RTPCodecCapability
}

// Room 一通群通话的路由核心：维护参与者集合与「发布者 -> 订阅者下行 sink」路由表，
// 负责把某发布者的 RTP 包 fan-out 给其余订阅者。并发安全（所有访问经 mu）。
type Room struct {
	id string

	mu           sync.RWMutex
	participants map[string]struct{}             // uid 集合
	subs         map[trackKey]map[string]RTPSink // 已发布轨 -> (订阅者 uid -> 下行 sink)
	published    map[trackKey]PublishedTrack     // 当前已发布轨注册表（供回填）
	activeAudio  map[string]struct{}             // 当前允许转发音频的发布者 uid
	managedAudio map[string]struct{}             // 已协商 audio-level、可参与 top-N 的发布者
	speakers     *ActiveSpeakers
}

// NewRoom 创建空房间。
func NewRoom(id string) *Room {
	return &Room{
		id:           id,
		participants: make(map[string]struct{}),
		subs:         make(map[trackKey]map[string]RTPSink),
		published:    make(map[trackKey]PublishedTrack),
		activeAudio:  make(map[string]struct{}),
		managedAudio: make(map[string]struct{}),
		speakers:     NewActiveSpeakers(),
	}
}

func (r *Room) ManageAudio(uid string) {
	r.mu.Lock()
	r.managedAudio[uid] = struct{}{}
	r.mu.Unlock()
}

// ObserveAudioLevel 更新发布者响度，并立即刷新音频转发选择。
func (r *Room) ObserveAudioLevel(uid string, level uint8, limit int) {
	r.speakers.Observe(uid, level)
	r.SetActiveSpeakers(r.speakers.Top(limit))
}

// ActiveSpeakers 返回当前按响度排序的发言人。
func (r *Room) ActiveSpeakers(limit int) []string {
	return r.speakers.Top(limit)
}

// SetActiveSpeakers 原子替换当前允许转发的音频发布者集合，不改变下行轨和 SDP。
func (r *Room) SetActiveSpeakers(uids []string) {
	active := make(map[string]struct{}, len(uids))
	for _, uid := range uids {
		active[uid] = struct{}{}
	}
	r.mu.Lock()
	r.activeAudio = active
	r.mu.Unlock()
}

func (r *Room) refreshActiveSpeakers(limit int) []string {
	speakers := r.speakers.Top(limit)
	r.SetActiveSpeakers(speakers)
	return speakers
}

// ID 返回房间标识（= callID）。
func (r *Room) ID() string { return r.id }

// Join 加入房间（幂等）。
func (r *Room) Join(uid string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.participants[uid] = struct{}{}
}

// Leave 离开房间（未知成员为 no-op），并清理其订阅路由：
// 作为发布者——移除其所有已发布轨的订阅表；作为订阅者——从其余轨的订阅表里移除自己。
func (r *Room) Leave(uid string) {
	r.mu.Lock()
	removed := make([]RTPSink, 0)
	delete(r.participants, uid)
	delete(r.managedAudio, uid)
	delete(r.activeAudio, uid)
	for key, subs := range r.subs {
		if key.pubUID == uid {
			for _, sink := range subs {
				removed = append(removed, sink)
			}
			delete(r.subs, key) // 作为发布者：整条轨的订阅表清掉
			continue
		}
		if sink, ok := subs[uid]; ok {
			removed = append(removed, sink)
			delete(subs, uid) // 作为订阅者：从别人轨的订阅表里移除自己
		}
		if len(subs) == 0 {
			delete(r.subs, key)
		}
	}
	for key := range r.published {
		if key.pubUID == uid {
			delete(r.published, key) // 作为发布者：从已发布注册表移除
		}
	}
	r.mu.Unlock()
	closeSinks(removed)
}

// AddPublished 登记一条已发布轨（供迟到者回填订阅）。
func (r *Room) AddPublished(pubUID, trackID, kind string, codec webrtc.RTPCodecCapability) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := trackKey{pubUID: pubUID, trackID: trackID}
	r.published[key] = PublishedTrack{PubUID: pubUID, TrackID: trackID, Kind: kind, Codec: codec}
}

// Unpublish 移除一条已发布轨及其订阅表（轨结束/关闭时）。
func (r *Room) Unpublish(pubUID, trackID string) {
	r.mu.Lock()
	key := trackKey{pubUID: pubUID, trackID: trackID}
	removed := make([]RTPSink, 0, len(r.subs[key]))
	for _, sink := range r.subs[key] {
		removed = append(removed, sink)
	}
	delete(r.published, key)
	delete(r.subs, key)
	r.mu.Unlock()
	closeSinks(removed)
}

// PublishedExcept 返回除 uid 外的全部已发布轨（迟到者 uid 入房时据此回填订阅）。
func (r *Room) PublishedExcept(uid string) []PublishedTrack {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]PublishedTrack, 0, len(r.published))
	for _, pt := range r.published {
		if pt.PubUID == uid {
			continue
		}
		out = append(out, pt)
	}
	return out
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

// Empty 报告房间当前是否没有参与者。
func (r *Room) Empty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.participants) == 0
}

// Subscribe 登记：订阅者 subUID 订阅发布者 pubUID 的 trackID，下行经 sink 投递。
// 订阅自己的轨会被记录但在 RouteRTP 时排除（不回声）。
func (r *Room) Subscribe(subUID, pubUID, trackID string, sink RTPSink) {
	key := trackKey{pubUID: pubUID, trackID: trackID}
	r.mu.Lock()
	subs, ok := r.subs[key]
	if !ok {
		subs = make(map[string]RTPSink)
		r.subs[key] = subs
	}
	old := subs[subUID]
	subs[subUID] = sink
	r.mu.Unlock()
	if old != nil && old != sink {
		closeSinks([]RTPSink{old})
	}
}

func closeSinks(sinks []RTPSink) {
	for _, sink := range sinks {
		if closer, ok := sink.(io.Closer); ok {
			_ = closer.Close()
		}
	}
}

// RouteRTP 把发布者 pubUID 的 trackID 的一个 RTP 包 fan-out 给所有订阅者，
// 但永不回给发布者自己。未知轨 / 无订阅者为 no-op。
func (r *Room) RouteRTP(pubUID, trackID string, pkt *rtp.Packet) {
	key := trackKey{pubUID: pubUID, trackID: trackID}
	r.mu.RLock()
	if published, ok := r.published[key]; ok && published.Kind == webrtc.RTPCodecTypeAudio.String() {
		_, managed := r.managedAudio[pubUID]
		_, active := r.activeAudio[pubUID]
		if managed && !active {
			r.mu.RUnlock()
			return
		}
	}
	// 拷出目标 sink，避免持锁写 I/O
	targets := make([]RTPSink, 0, len(r.subs[key]))
	for subUID, sink := range r.subs[key] {
		if subUID == pubUID {
			continue // 排除自己
		}
		targets = append(targets, sink)
	}
	r.mu.RUnlock()

	for _, sink := range targets {
		_ = sink.WriteRTP(pkt) // 单个订阅者写失败不影响其他订阅者；错误由上层 I/O 层处理
	}
}
