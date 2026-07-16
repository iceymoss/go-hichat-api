package sfu

import (
	"testing"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

// newLoopbackClient 建一个客户端 PC + SFU peer，接好服务端发起的 renegotiation，
// 并发布一路 VP8 轨（客户端作初始 offerer）。返回 SFU peer、客户端 PC、可写样本的本地轨。
func newLoopbackClient(t *testing.T, s *SFU, room, uid string) (*Peer, *webrtc.PeerConnection, *webrtc.TrackLocalStaticSample) {
	t.Helper()
	peer, err := s.AddPeer(room, uid, nil)
	if err != nil {
		t.Fatalf("AddPeer(%s): %v", uid, err)
	}
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatalf("new client pc(%s): %v", uid, err)
	}

	// 服务端发起的 renegotiation：SFU offer -> 客户端 answer -> 回 SFU
	peer.OnOffer(func(sdp string) {
		if err := pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: sdp}); err != nil {
			return
		}
		ans, err := pc.CreateAnswer(nil)
		if err != nil {
			return
		}
		g := webrtc.GatheringCompletePromise(pc)
		if err := pc.SetLocalDescription(ans); err != nil {
			return
		}
		<-g
		_ = peer.HandleAnswer(pc.LocalDescription().SDP)
	})

	// 客户端发布一路 VP8 轨
	track, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", uid)
	if err != nil {
		t.Fatalf("new track(%s): %v", uid, err)
	}
	if _, err := pc.AddTrack(track); err != nil {
		t.Fatalf("add track(%s): %v", uid, err)
	}
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		t.Fatalf("create offer(%s): %v", uid, err)
	}
	g := webrtc.GatheringCompletePromise(pc)
	if err := pc.SetLocalDescription(offer); err != nil {
		t.Fatalf("set local(%s): %v", uid, err)
	}
	<-g
	ans, err := peer.Publish(pc.LocalDescription().SDP)
	if err != nil {
		t.Fatalf("publish(%s): %v", uid, err)
	}
	if err := pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: ans}); err != nil {
		t.Fatalf("set remote answer(%s): %v", uid, err)
	}
	return peer, pc, track
}

func writeVP8Loop(track *webrtc.TrackLocalStaticSample, stop <-chan struct{}) {
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
}

// Test_SFU_TwoPeers_ForwardsViaRenegotiation_Loopback 两条真实 PC 的完整 loopback：
// B 先入房发布，A 后入房发布 -> SFU 对 A 的轨为 B 建下行（AddTrack 触发服务端 renegotiation）
// -> B 的真实 PC 收到 A 经 SFU 转发的媒体。验证真实下行 + renegotiation 全链路（不碰浏览器）。
func Test_SFU_TwoPeers_ForwardsViaRenegotiation_Loopback(t *testing.T) {
	s := NewSFU() // 真实 pion 下行工厂

	// B 先入房（发布自己；此时房内无他人，OnTrack(B) 不建下行）
	_, pcB, _ := newLoopbackClient(t, s, "call1", "b")
	defer pcB.Close()

	// B 收到下行轨即视为成功
	gotRTP := make(chan struct{}, 1)
	pcB.OnTrack(func(tr *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		for {
			if _, _, err := tr.ReadRTP(); err != nil {
				return
			}
			select {
			case gotRTP <- struct{}{}:
			default:
			}
		}
	})
	if err := s.SubscribeVideo("call1", "b", "a"); err != nil {
		t.Fatalf("SubscribeVideo: %v", err)
	}

	// A 后入房发布 -> SFU 为 B 建下行 A 的轨 -> 触发对 B 的 renegotiation
	_, pcA, trackA := newLoopbackClient(t, s, "call1", "a")
	defer pcA.Close()

	stop := make(chan struct{})
	go writeVP8Loop(trackA, stop)
	defer close(stop)

	select {
	case <-gotRTP:
		// 成功：B 的真实 PC 经 SFU renegotiation 收到 A 的转发媒体
	case <-time.After(15 * time.Second):
		t.Fatal("B did not receive A's forwarded media via renegotiation within timeout")
	}
}
