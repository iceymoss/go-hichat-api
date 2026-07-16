package sfu

import (
	"reflect"
	"testing"
	"time"

	"github.com/pion/webrtc/v3"
)

// Test_SFU_LateJoiner_ReceivesExistingPublisher_Loopback 迟到者回填：
// A 先入房发布（房内无他人），B 后入房 -> SubscribeExisting 让 B 订阅 A 的已发布轨
// -> B 的真实 PC 收到 A 经 SFU 转发的媒体。验证 SubscribeExisting 回填 + renegotiation 全链路。
func Test_SFU_LateJoiner_ReceivesExistingPublisher_Loopback(t *testing.T) {
	s := NewSFU()

	// A 先入房发布，并持续发媒体
	_, pcA, trackA := newLoopbackClient(t, s, "call1", "a")
	defer pcA.Close()
	stop := make(chan struct{})
	go writeVP8Loop(trackA, stop)
	defer close(stop)

	// B 后入房（其自身发布不会让 B 收到 A——A 早于 B 入房）
	_, pcB, _ := newLoopbackClient(t, s, "call1", "b")
	defer pcB.Close()

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

	// 回填只自动订阅音频；可见视频由分页订阅显式请求。
	s.SubscribeExisting("call1", "b")
	if err := s.SubscribeVideo("call1", "b", "a"); err != nil {
		t.Fatalf("SubscribeVideo: %v", err)
	}

	select {
	case <-gotRTP:
		// 成功：迟到者 B 经回填 + renegotiation 收到早入房者 A 的媒体
	case <-time.After(15 * time.Second):
		t.Fatal("late joiner B did not receive existing publisher A's media within timeout")
	}
}

func Test_SFU_SubscribeExisting_OnlyAutomaticallySubscribesAudio(t *testing.T) {
	factory := &fakeFactory{sinks: map[string]*fakeSink{}, codecs: map[string]webrtc.RTPCodecCapability{}}
	s := newSFUWithFactory(factory)
	room := s.room("call1")
	room.Join("publisher")
	room.Join("subscriber")
	videoCodec := webrtc.RTPCodecCapability{
		MimeType:     webrtc.MimeTypeH264,
		ClockRate:    90000,
		SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
		RTCPFeedback: []webrtc.RTCPFeedback{{Type: "nack"}, {Type: "nack", Parameter: "pli"}},
	}
	room.AddPublished("publisher", "video", "video", videoCodec)
	audioCodec := webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2}
	room.AddPublished("publisher", "audio", "audio", audioCodec)

	s.SubscribeExisting("call1", "subscriber")

	if got := factory.createdCount(); got != 1 {
		t.Fatalf("created downlinks = %d, want only audio downlink", got)
	}
	if got := factory.codec("subscriber"); !reflect.DeepEqual(got, audioCodec) {
		t.Fatalf("downlink codec = %+v, want %+v", got, audioCodec)
	}
}

func Test_SFU_VideoSubscription_IsPendingIdempotentAndRemovable(t *testing.T) {
	factory := &fakeFactory{sinks: map[string]*fakeSink{}}
	s := newSFUWithFactory(factory)
	room := s.room("call1")
	room.Join("publisher")
	room.Join("subscriber")

	if err := s.SubscribeVideo("call1", "subscriber", "publisher"); err != nil {
		t.Fatal(err)
	}
	if got := factory.createdCount(); got != 0 {
		t.Fatalf("downlinks before publish = %d, want 0", got)
	}
	room.AddPublished("publisher", "video", "video", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8})
	if err := s.reconcileVideoSubscription("call1", "subscriber", "publisher"); err != nil {
		t.Fatal(err)
	}
	if err := s.SubscribeVideo("call1", "subscriber", "publisher"); err != nil {
		t.Fatal(err)
	}
	if got := factory.createdCount(); got != 1 {
		t.Fatalf("downlinks after repeated subscribe = %d, want 1", got)
	}

	s.UnsubscribeVideo("call1", "subscriber", "publisher")
	s.UnsubscribeVideo("call1", "subscriber", "publisher")
	if room.HasVideoSubscription("subscriber", "publisher") {
		t.Fatal("video subscription remains after unsubscribe")
	}
}
