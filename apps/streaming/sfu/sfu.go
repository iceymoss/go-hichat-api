package sfu

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pion/ice/v2"
	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// downlinkFactory 为「订阅者 subUID 订阅发布者 pubUID 的 trackID」创建一条下行投递 sink。
// 生产实现 realDownlink：建 *webrtc.TrackLocalStaticRTP，AddTrack 到 subUID 的 PeerConnection
// （触发服务端 renegotiation），返回该 track 作为 RTPSink。测试可注入假实现隔离 pion I/O。
type downlinkFactory interface {
	newDownlink(subUID, pubUID, trackID string, kind webrtc.RTPCodecType) (RTPSink, error)
}

// SFU 协调器：管理房间、pion peer 注册表，编排「某人发布 -> 为其余人建下行订阅 + 起收流泵」。
type SFU struct {
	mu      sync.Mutex
	rooms   map[string]*Room
	peers   map[string]*Peer  // uid -> peer（单会话，uid 全局唯一）
	pubSSRC map[string]uint32 // "pubUID|trackID" -> 上行轨 SSRC（PLI 转发用）
	factory downlinkFactory
	api     *webrtc.API // 带接口过滤的 pion API（滤掉虚拟接口候选）
}

// Option 配置 SFU 的 pion 网络参数（Phase 1 公网化）。
type Option func(*sfuConfig)

type sfuConfig struct {
	publicIP       string
	udpMin, udpMax int
}

// WithPublicIP 让 SFU 对外宣告公网 IP（NAT1To1），跨 NAT 时客户端才连得到。
func WithPublicIP(ip string) Option { return func(c *sfuConfig) { c.publicIP = ip } }

// WithUDPPortRange 把媒体 UDP 限制在指定端口范围（需在防火墙/docker 放行）。
func WithUDPPortRange(min, max int) Option {
	return func(c *sfuConfig) { c.udpMin, c.udpMax = min, max }
}

// NewSFU 创建协调器，使用真实 pion 下行工厂（生产）。opts 配置公网 IP / UDP 端口范围。
func NewSFU(opts ...Option) *SFU {
	cfg := sfuConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	s := &SFU{rooms: make(map[string]*Room), peers: make(map[string]*Peer), pubSSRC: make(map[string]uint32), api: buildAPI(cfg)}
	s.factory = realDownlink{sfu: s}
	return s
}

// newSFUWithFactory 注入自定义下行工厂（测试用，隔离 pion 下行 I/O）。
func newSFUWithFactory(f downlinkFactory) *SFU {
	return &SFU{rooms: make(map[string]*Room), peers: make(map[string]*Peer), pubSSRC: make(map[string]uint32), api: buildAPI(sfuConfig{}), factory: f}
}

// buildAPI 构造带 SettingEngine 的 pion API：注册默认编解码 + 拦截器（NACK/RTCP 等），
// 并过滤 ICE 采集接口——只留可用的物理/回环接口，滤掉 VPN/Docker/AWDL 等虚拟接口，
// 否则它们的垃圾候选会拖慢 ICE 甚至判失败（表现为发起即报错、加载很慢）。
// Phase 1：可选宣告公网 IP（NAT1To1）+ 限定媒体 UDP 端口范围。
func buildAPI(cfg sfuConfig) *webrtc.API {
	me := &webrtc.MediaEngine{}
	if err := me.RegisterDefaultCodecs(); err != nil {
		panic(err)
	}
	ir := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(me, ir); err != nil {
		panic(err)
	}
	se := webrtc.SettingEngine{}
	se.SetInterfaceFilter(usableInterface)
	// 解析浏览器发来的 mDNS(.local) 主机候选：Chrome 默认把 host IP 藏成 .local，
	// 不解析就只能靠反射候选连通，本地/局域网易慢或连不上。
	se.SetICEMulticastDNSMode(ice.MulticastDNSModeQueryOnly)
	if cfg.publicIP != "" {
		se.SetNAT1To1IPs([]string{cfg.publicIP}, webrtc.ICECandidateTypeHost)
	}
	if cfg.udpMin > 0 && cfg.udpMax >= cfg.udpMin {
		if err := se.SetEphemeralUDPPortRange(uint16(cfg.udpMin), uint16(cfg.udpMax)); err != nil {
			panic(err)
		}
	}
	return webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithInterceptorRegistry(ir), webrtc.WithSettingEngine(se))
}

// usableInterface 判断某网卡是否用于 ICE 候选采集：排除已知虚拟/隧道接口。
func usableInterface(name string) bool {
	for _, bad := range []string{"utun", "awdl", "llw", "bridge", "vmenet", "docker", "veth", "tap", "tun", "ipsec", "ppp"} {
		if strings.HasPrefix(name, bad) {
			return false
		}
	}
	return true
}

func ssrcKey(pubUID, trackID string) string { return pubUID + "|" + trackID }

// STUNServers 构造 pion ICE 配置（服务端 PeerConnection 用）。给服务端也配 STUN 可增强候选，
// 缓解 Chrome 把 host 候选藏成 mDNS 时只能靠反射候选连通的情况。
func STUNServers(urls ...string) []webrtc.ICEServer {
	if len(urls) == 0 {
		return nil
	}
	return []webrtc.ICEServer{{URLs: urls}}
}

func (s *SFU) setPubSSRC(pubUID, trackID string, ssrc uint32) {
	s.mu.Lock()
	s.pubSSRC[ssrcKey(pubUID, trackID)] = ssrc
	s.mu.Unlock()
}

func (s *SFU) getPubSSRC(pubUID, trackID string) (uint32, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ssrc, ok := s.pubSSRC[ssrcKey(pubUID, trackID)]
	return ssrc, ok
}

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

func (s *SFU) getPeer(uid string) *Peer {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.peers[uid]
}

// GetPeer 返回 uid 的 peer（信令层路由 sfu_publish/answer 用），不存在返回 nil。
func (s *SFU) GetPeer(uid string) *Peer { return s.getPeer(uid) }

// RemovePeer 关闭并移除某参与者的 peer（离开/断线时调用）：从注册表删除并关闭其 PeerConnection，
// 触发房间路由清理（其上行轨与作为订阅者的下行订阅一并移除）。
func (s *SFU) RemovePeer(uid string) {
	s.mu.Lock()
	p := s.peers[uid]
	delete(s.peers, uid)
	s.mu.Unlock()
	if p != nil {
		_ = p.Close()
	}
}

// Peer 一个参与者在房间内的 pion 连接（含服务端发起 renegotiation 的能力）。
type Peer struct {
	uid  string
	room *Room
	pc   *webrtc.PeerConnection

	negMu      sync.Mutex
	negPending bool             // 有 offer 在途、期间又有 track 变化时标记待补
	negTimer   *time.Timer      // 去抖定时器（合并一批 AddTrack 成一次 renegotiation）
	onOffer    func(sdp string) // 服务端发起的 renegotiation offer -> 客户端（由信令层接线）
}

// AddPeer 为 uid 建一条新 PeerConnection 加入房间；OnTrack 里为其余人建下行订阅并起收流泵；
// AddTrack 触发的重协商由服务端作为 offerer 驱动（避 glare）。
func (s *SFU) AddPeer(roomID, uid string, iceServers []webrtc.ICEServer) (*Peer, error) {
	pc, err := s.api.NewPeerConnection(webrtc.Configuration{ICEServers: iceServers})
	if err != nil {
		return nil, err
	}
	room := s.room(roomID)
	room.Join(uid)
	p := &Peer{uid: uid, room: room, pc: pc}

	pc.OnNegotiationNeeded(p.scheduleNegotiation)
	pc.OnSignalingStateChange(func(st webrtc.SignalingState) {
		if st == webrtc.SignalingStateStable {
			p.flushPendingNegotiation()
		}
	})
	pc.OnTrack(func(tr *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		trackID := tr.ID()
		s.setPubSSRC(uid, trackID, uint32(tr.SSRC())) // 记录上行 SSRC（PLI 转发用）
		room.AddPublished(uid, trackID, tr.Kind().String())
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
		go func() {
			pump(room, uid, trackID, trackRemoteReader{t: tr})
			room.Unpublish(uid, trackID) // 轨结束：从注册表 + 订阅表清理
		}()
	})

	s.mu.Lock()
	s.peers[uid] = p
	s.mu.Unlock()
	return p, nil
}

// SubscribeExisting 为迟到入房者 uid 回填订阅：订阅房间内已有的全部发布轨，
// 使其收到早于自己入房者的媒体。信令层在该 peer 发布/就绪后调用。
func (s *SFU) SubscribeExisting(roomID, uid string) {
	room := s.room(roomID)
	for _, pt := range room.PublishedExcept(uid) {
		sink, err := s.factory.newDownlink(uid, pt.PubUID, pt.TrackID, kindFromString(pt.Kind))
		if err != nil {
			continue
		}
		room.Subscribe(uid, pt.PubUID, pt.TrackID, sink)
	}
}

// kindFromString 把 "audio"/"video" 还原为 webrtc.RTPCodecType。
func kindFromString(s string) webrtc.RTPCodecType {
	if s == webrtc.RTPCodecTypeAudio.String() {
		return webrtc.RTPCodecTypeAudio
	}
	return webrtc.RTPCodecTypeVideo
}

// Publish 处理客户端发布 offer（客户端作 offerer），返回 SFU answer（非 trickle）。
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

// OnOffer 注册服务端发起的 renegotiation offer 回调（信令层把 sdp 送给客户端）。
func (p *Peer) OnOffer(cb func(sdp string)) {
	p.negMu.Lock()
	p.onOffer = cb
	p.negMu.Unlock()
}

// HandleAnswer 应用客户端对服务端 renegotiation offer 的 answer。
func (p *Peer) HandleAnswer(sdp string) error {
	return p.pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: sdp})
}

// negotiationDebounce 去抖窗口：把一批 AddTrack（如迟到者回填 4 条轨）合并成一次 renegotiation，
// 大幅减少重协商竞争与偶发丢轨（"少一个人的视频"的根因）。
const negotiationDebounce = 120 * time.Millisecond

// scheduleNegotiation 由 OnNegotiationNeeded 触发：去抖排程，多次触发合并成一次 offer。
func (p *Peer) scheduleNegotiation() {
	p.negMu.Lock()
	defer p.negMu.Unlock()
	if p.negTimer != nil {
		return // 已排程，等它触发时会反映当时的全部 track 变化
	}
	p.negTimer = time.AfterFunc(negotiationDebounce, p.doNegotiation)
}

// doNegotiation 去抖窗口到期：stable 就发一次合并的 offer；有 offer 在途则标记待补。
func (p *Peer) doNegotiation() {
	p.negMu.Lock()
	p.negTimer = nil
	if p.pc.SignalingState() != webrtc.SignalingStateStable {
		p.negPending = true // 有 offer 在途，待 answer 回 stable 后再补
		p.negMu.Unlock()
		return
	}
	sdp, ok := p.createOfferLocked()
	cb := p.onOffer
	p.negMu.Unlock()
	if ok && cb != nil {
		cb(sdp) // 锁外回调，避免重入死锁
	}
}

// flushPendingNegotiation 信令回到 stable 时，若期间又有 track 变化则再排一次去抖协商。
func (p *Peer) flushPendingNegotiation() {
	p.negMu.Lock()
	if !p.negPending || p.negTimer != nil {
		p.negMu.Unlock()
		return
	}
	p.negPending = false
	p.negTimer = time.AfterFunc(negotiationDebounce, p.doNegotiation)
	p.negMu.Unlock()
}

// createOfferLocked 生成并设置本端 offer，返回其 SDP（含已收集候选）。调用方须持 negMu。
func (p *Peer) createOfferLocked() (string, bool) {
	offer, err := p.pc.CreateOffer(nil)
	if err != nil {
		return "", false
	}
	if err := p.pc.SetLocalDescription(offer); err != nil {
		return "", false
	}
	return p.pc.LocalDescription().SDP, true
}

// Close 关闭连接并从房间 + 注册表移除（触发订阅路由清理）。
func (p *Peer) Close() error {
	p.negMu.Lock()
	if p.negTimer != nil {
		p.negTimer.Stop()
		p.negTimer = nil
	}
	p.negMu.Unlock()
	p.room.Leave(p.uid)
	return p.pc.Close()
}

// realDownlink 真实 pion 下行：为订阅者 PC 建下行本地轨并 AddTrack（触发其重协商）。
type realDownlink struct{ sfu *SFU }

func (d realDownlink) newDownlink(subUID, pubUID, trackID string, kind webrtc.RTPCodecType) (RTPSink, error) {
	sub := d.sfu.getPeer(subUID)
	if sub == nil {
		return nil, fmt.Errorf("sfu: subscriber peer %s not found", subUID)
	}
	local, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: mimeForKind(kind)},
		pubUID+"_"+trackID, // 下行 track id（含发布者，便于订阅端区分来源）
		pubUID,             // stream id = 发布者
	)
	if err != nil {
		return nil, err
	}
	sender, err := sub.pc.AddTrack(local) // 触发 sub 的 OnNegotiationNeeded
	if err != nil {
		return nil, err
	}
	// 视频：转发订阅者的关键帧请求(PLI/FIR)给发布者，避免新订阅者黑屏等待下一个关键帧
	if kind == webrtc.RTPCodecTypeVideo {
		go d.forwardKeyframeRequests(sender, pubUID, trackID)
	}
	return local, nil
}

// forwardKeyframeRequests 读订阅者下行 sender 的 RTCP，遇 PLI/FIR 就向发布者 PC 发 PLI 请关键帧。
func (d realDownlink) forwardKeyframeRequests(sender *webrtc.RTPSender, pubUID, trackID string) {
	buf := make([]byte, 1500)
	for {
		n, _, err := sender.Read(buf)
		if err != nil {
			return // sender 关闭
		}
		pkts, err := rtcp.Unmarshal(buf[:n])
		if err != nil {
			continue
		}
		for _, pkt := range pkts {
			switch pkt.(type) {
			case *rtcp.PictureLossIndication, *rtcp.FullIntraRequest:
				pub := d.sfu.getPeer(pubUID)
				ssrc, ok := d.sfu.getPubSSRC(pubUID, trackID)
				if pub == nil || !ok {
					continue
				}
				_ = pub.pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: ssrc}})
			}
		}
	}
}

// mimeForKind Phase 0 固定编解码：音频 Opus、视频 VP8（simulcast/协商留 Phase 3）。
func mimeForKind(kind webrtc.RTPCodecType) string {
	if kind == webrtc.RTPCodecTypeAudio {
		return webrtc.MimeTypeOpus
	}
	return webrtc.MimeTypeVP8
}

// trackRemoteReader 把 *webrtc.TrackRemote 适配成 rtpReader（丢弃 interceptor.Attributes）。
type trackRemoteReader struct{ t *webrtc.TrackRemote }

func (r trackRemoteReader) ReadRTP() (*rtp.Packet, error) {
	pkt, _, err := r.t.ReadRTP()
	return pkt, err
}
