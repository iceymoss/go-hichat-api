package sfu

import (
	"sync"
	"testing"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

// fakeFactory 假下行工厂：为每个订阅者建一个 fakeSink 并记录，隔离真实 pion 下行接线。
// 并发安全（newDownlink 在 pion 的 OnTrack goroutine 里被调，测试主 goroutine 读取）。
type fakeFactory struct {
	mu      sync.Mutex
	sinks   map[string]*fakeSink
	codecs  map[string]webrtc.RTPCodecCapability
	created int
}

func (f *fakeFactory) newDownlink(subUID, _ string, _ string, codec webrtc.RTPCodecCapability) (RTPSink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s := &fakeSink{}
	f.sinks[subUID] = s
	f.created++
	if f.codecs != nil {
		f.codecs[subUID] = codec
	}
	return s, nil
}

func (f *fakeFactory) createdCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.created
}

func (f *fakeFactory) sink(subUID string) *fakeSink {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sinks[subUID]
}

func (f *fakeFactory) codec(subUID string) webrtc.RTPCodecCapability {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.codecs[subUID]
}

// Test_SFU_IngestsRealTrackAndFansOut_Loopback pion 进程内 loopback 集成测试：
// 真实发布端 PeerConnection 发一路 VP8 轨 -> SFU 的 OnTrack 编排下行订阅 + 起收流泵 ->
// 断言订阅者 sink 收到经 SFU 转发的真实 RTP。验证核心媒体 ingest+fan-out 路径（不碰浏览器）。
func Test_SFU_IngestsRealTrackAndFansOut_Loopback(t *testing.T) {
	factory := &fakeFactory{sinks: map[string]*fakeSink{}}
	s := newSFUWithFactory(factory)

	// sub 先在房间里，这样 pub 的 OnTrack 会为 sub 建下行订阅
	room := s.room("call1")
	room.Join("sub")
	if err := s.SubscribeVideo("call1", "sub", "pub"); err != nil {
		t.Fatalf("SubscribeVideo: %v", err)
	}

	peer, err := s.AddPeer("call1", "pub", nil)
	if err != nil {
		t.Fatalf("AddPeer: %v", err)
	}
	defer peer.Close()

	// 发布端真实 PC + 一路 VP8 sample 轨
	pcPub, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatalf("new pub pc: %v", err)
	}
	defer pcPub.Close()

	track, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "stream")
	if err != nil {
		t.Fatalf("new track: %v", err)
	}
	if _, err := pcPub.AddTrack(track); err != nil {
		t.Fatalf("add track: %v", err)
	}

	// 非 trickle：等 ICE 收集完，SDP 内含全部候选
	offer, err := pcPub.CreateOffer(nil)
	if err != nil {
		t.Fatalf("create offer: %v", err)
	}
	gatherPub := webrtc.GatheringCompletePromise(pcPub)
	if err := pcPub.SetLocalDescription(offer); err != nil {
		t.Fatalf("set local: %v", err)
	}
	<-gatherPub

	answerSDP, err := peer.Publish(pcPub.LocalDescription().SDP)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if err := pcPub.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: answerSDP}); err != nil {
		t.Fatalf("set remote answer: %v", err)
	}

	// 持续写样本，驱动连接建立 + RTP 流动，直到订阅者收到转发或超时
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				_ = track.WriteSample(media.Sample{Data: []byte{0x00, 0x01, 0x02, 0x03}, Duration: 20 * time.Millisecond})
			}
		}
	}()
	defer close(stop)

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if sink := factory.sink("sub"); sink != nil && sink.count() > 0 {
			return // 成功：真实媒体经 SFU 转发到订阅端
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("subscriber sink received no forwarded RTP within timeout")
}
