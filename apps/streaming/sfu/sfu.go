package sfu

import (
	"sync"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// downlinkFactory 为「订阅者 subUID 订阅发布者 pubUID 的 trackID」创建一条下行投递 sink。
// 生产实现：建 *webrtc.TrackLocalStaticRTP，AddTrack 到 subUID 的 PeerConnection（触发 renegotiation），
// 返回该 track 作为 RTPSink。测试用假实现，隔离 pion I/O。
type downlinkFactory interface {
	newDownlink(subUID, pubUID, trackID string, kind webrtc.RTPCodecType) (RTPSink, error)
}

// SFU 协调器：管理房间与其 pion peer，编排「某人发布 -> 为房间内其余人建下行订阅 + 起收流泵」。
type SFU struct {
	mu      sync.Mutex
	rooms   map[string]*Room
	factory downlinkFactory
}

// NewSFU 创建协调器；downlinkFactory 负责下行轨的真实 pion 接线（生产）或假实现（测试）。
func NewSFU(factory downlinkFactory) *SFU {
	return &SFU{rooms: make(map[string]*Room), factory: factory}
}

// room 取或建房间。
func (s *SFU) room(roomID string) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rooms[roomID]
	if !ok {
		r = NewRoom(roomID)
		s.rooms[roomID] = r
	}
	return r
}

// Peer 一个参与者在房间内的 pion 连接。
type Peer struct {
	uid  string
	room *Room
	pc   *webrtc.PeerConnection
}

// AddPeer 为 uid 建一条新 PeerConnection 加入房间；OnTrack 里为其余人建下行订阅并起收流泵。
func (s *SFU) AddPeer(roomID, uid string, iceServers []webrtc.ICEServer) (*Peer, error) {
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{ICEServers: iceServers})
	if err != nil {
		return nil, err
	}
	room := s.room(roomID)
	room.Join(uid)
	p := &Peer{uid: uid, room: room, pc: pc}

	pc.OnTrack(func(tr *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		trackID := tr.ID()
		// 为房间内其余人建下行订阅（下行 I/O 由 factory 负责）
		for _, sub := range room.Participants() {
			if sub == uid {
				continue
			}
			sink, err := s.factory.newDownlink(sub, uid, trackID, tr.Kind())
			if err != nil {
				continue
			}
			room.Subscribe(sub, uid, trackID, sink)
		}
		// 起收流泵：把上行 RTP 持续 fan-out，直到该轨关闭
		go pump(room, uid, trackID, trackRemoteReader{t: tr})
	})

	return p, nil
}

// Publish 处理客户端发布 offer，返回 SFU answer（非 trickle：answer 内含全部 ICE 候选）。
func (p *Peer) Publish(offerSDP string) (string, error) {
	if err := p.pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: offerSDP}); err != nil {
		return "", err
	}
	answer, err := p.pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	gatherComplete := webrtc.GatheringCompletePromise(p.pc)
	if err := p.pc.SetLocalDescription(answer); err != nil {
		return "", err
	}
	<-gatherComplete
	return p.pc.LocalDescription().SDP, nil
}

// Close 关闭连接并从房间移除（触发订阅路由清理）。
func (p *Peer) Close() error {
	p.room.Leave(p.uid)
	return p.pc.Close()
}

// trackRemoteReader 把 *webrtc.TrackRemote 适配成 rtpReader（丢弃 interceptor.Attributes）。
type trackRemoteReader struct{ t *webrtc.TrackRemote }

func (r trackRemoteReader) ReadRTP() (*rtp.Packet, error) {
	pkt, _, err := r.t.ReadRTP()
	return pkt, err
}
