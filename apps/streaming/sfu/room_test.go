package sfu

import (
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// fakeSink 假下行 sink：记录收到的 RTP 包，供断言 fan-out 是否正确。
type fakeSink struct {
	mu   sync.Mutex
	pkts []*rtp.Packet
}

type closableSink struct {
	fakeSink
	closed int
}

func (s *closableSink) Close() error {
	s.closed++
	return nil
}

func (f *fakeSink) WriteRTP(p *rtp.Packet) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pkts = append(f.pkts, p)
	return nil
}

func (f *fakeSink) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.pkts)
}

// Test_Room_JoinLeave_TracksMembership 房间成员进出簿记：加入幂等、离开移除、离开未知成员无副作用。
func Test_Room_JoinLeave_TracksMembership(t *testing.T) {
	tests := []struct {
		name string
		ops  func(r *Room)
		want []string
	}{
		{
			name: "join adds participants",
			ops:  func(r *Room) { r.Join("a"); r.Join("b") },
			want: []string{"a", "b"},
		},
		{
			name: "join is idempotent",
			ops:  func(r *Room) { r.Join("a"); r.Join("a") },
			want: []string{"a"},
		},
		{
			name: "leave removes participant",
			ops:  func(r *Room) { r.Join("a"); r.Join("b"); r.Leave("a") },
			want: []string{"b"},
		},
		{
			name: "leave unknown is no-op",
			ops:  func(r *Room) { r.Join("a"); r.Leave("x") },
			want: []string{"a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRoom("call1")
			tt.ops(r)
			got := r.Participants()
			sort.Strings(got)
			if !equalStrs(got, tt.want) {
				t.Errorf("Participants() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equalStrs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Test_Room_RouteRTP_FansOutToSubscribersExceptSelf 核心 fan-out：
// 某发布者的 RTP 包投递给所有订阅其该 track 的订阅者，且永不回给发布者自己；
// 无订阅者 / 未知 track 为 no-op（不 panic）。
func Test_Room_RouteRTP_FansOutToSubscribersExceptSelf(t *testing.T) {
	r := NewRoom("call1")
	r.Join("a")
	r.Join("b")
	r.Join("c")

	sinkB := &fakeSink{}
	sinkC := &fakeSink{}
	sinkSelf := &fakeSink{} // a 订阅自己的 track（异常，应被排除）

	r.Subscribe("b", "a", "a-audio", sinkB)
	r.Subscribe("c", "a", "a-audio", sinkC)
	r.Subscribe("a", "a", "a-audio", sinkSelf)

	// 路由 a 的两个包
	r.RouteRTP("a", "a-audio", &rtp.Packet{Header: rtp.Header{SequenceNumber: 1}})
	r.RouteRTP("a", "a-audio", &rtp.Packet{Header: rtp.Header{SequenceNumber: 2}})

	if got := sinkB.count(); got != 2 {
		t.Errorf("sinkB got %d pkts, want 2", got)
	}
	if got := sinkC.count(); got != 2 {
		t.Errorf("sinkC got %d pkts, want 2", got)
	}
	if got := sinkSelf.count(); got != 0 {
		t.Errorf("publisher must not receive own track: sinkSelf got %d, want 0", got)
	}

	// 未知 track / 无订阅者：no-op，不 panic
	r.RouteRTP("a", "nonexistent", &rtp.Packet{})
	r.RouteRTP("zzz", "a-audio", &rtp.Packet{})
}

func Test_Room_RouteRTP_OnlyForwardsSelectedAudioSpeakers(t *testing.T) {
	r := NewRoom("call1")
	for _, uid := range []string{"a", "b", "viewer"} {
		r.Join(uid)
	}
	r.AddPublished("a", "a-audio", "audio", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus})
	r.AddPublished("b", "b-audio", "audio", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus})
	r.AddPublished("b", "b-video", "video", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8})
	audioA := &fakeSink{}
	audioB := &fakeSink{}
	videoB := &fakeSink{}
	r.Subscribe("viewer", "a", "a-audio", audioA)
	r.Subscribe("viewer", "b", "b-audio", audioB)
	r.Subscribe("viewer", "b", "b-video", videoB)

	r.SetActiveSpeakers([]string{"a"})
	r.ManageAudio("a")
	r.ManageAudio("b")
	r.RouteRTP("a", "a-audio", &rtp.Packet{})
	r.RouteRTP("b", "b-audio", &rtp.Packet{})
	r.RouteRTP("b", "b-video", &rtp.Packet{})

	if got := audioA.count(); got != 1 {
		t.Fatalf("selected audio packets = %d, want 1", got)
	}
	if got := audioB.count(); got != 0 {
		t.Fatalf("unselected audio packets = %d, want 0", got)
	}
	if got := videoB.count(); got != 1 {
		t.Fatalf("video packets = %d, want 1", got)
	}

	r.SetActiveSpeakers([]string{"b"})
	r.RouteRTP("a", "a-audio", &rtp.Packet{})
	r.RouteRTP("b", "b-audio", &rtp.Packet{})
	if got := audioA.count(); got != 1 {
		t.Fatalf("deselected audio packets = %d, want unchanged 1", got)
	}
	if got := audioB.count(); got != 1 {
		t.Fatalf("newly selected audio packets = %d, want 1", got)
	}
}

func Test_Room_RouteRTP_ForwardsAudioWithoutLevelNegotiation(t *testing.T) {
	r := NewRoom("call1")
	r.AddPublished("legacy", "audio", "audio", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus})
	sink := &fakeSink{}
	r.Subscribe("listener", "legacy", "audio", sink)
	r.RouteRTP("legacy", "audio", &rtp.Packet{})
	if got := sink.count(); got != 1 {
		t.Fatalf("fallback audio packets = %d, want 1", got)
	}
}

func Test_Room_SimulcastLayersRemainDistinct(t *testing.T) {
	r := NewRoom("call1")
	codec := webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}
	for _, rid := range []string{"q", "h", "f"} {
		r.AddPublishedLayer("publisher", "video", rid, "video", codec)
	}
	if got := len(r.PublishedExcept("subscriber")); got != 3 {
		t.Fatalf("published layers = %d, want 3", got)
	}
	q := &fakeSink{}
	r.SubscribeLayer("subscriber", "publisher", "video", "q", q)
	r.RouteRTPLayer("publisher", "video", "h", &rtp.Packet{})
	r.RouteRTPLayer("publisher", "video", "q", &rtp.Packet{})
	if got := q.count(); got != 1 {
		t.Fatalf("selected q packets = %d, want 1", got)
	}
}

func Test_Room_SimulcastLayerReplacementClosesPreviousLayer(t *testing.T) {
	r := NewRoom("call1")
	codec := webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}
	r.AddPublishedLayer("publisher", "video", "q", "video", codec)
	r.AddPublishedLayer("publisher", "video", "h", "video", codec)
	q := &closableSink{}
	h := &closableSink{}
	r.SubscribeLayer("subscriber", "publisher", "video", "q", q)
	r.SubscribeLayer("subscriber", "publisher", "video", "h", h)

	r.RemoveVideoLayersExcept("subscriber", "publisher", "video", "h")

	if q.closed != 1 || h.closed != 0 {
		t.Fatalf("closed q/h = %d/%d, want 1/0", q.closed, h.closed)
	}
	if r.HasTrackSubscription("subscriber", "publisher", "video", "q") {
		t.Fatal("q subscription remains after selecting h")
	}
	if !r.HasTrackSubscription("subscriber", "publisher", "video", "h") {
		t.Fatal("h subscription missing after layer replacement")
	}
}

func Test_Room_VideoReplacementClosesPreviousTrackID(t *testing.T) {
	r := NewRoom("call1")
	old := &closableSink{}
	current := &closableSink{}
	r.SubscribeLayer("subscriber", "publisher", "old-video", "q", old)
	r.SubscribeLayer("subscriber", "publisher", "new-video", "q", current)

	r.RemoveVideoLayersExcept("subscriber", "publisher", "new-video", "q")

	if old.closed != 1 || current.closed != 0 {
		t.Fatalf("closed old/current = %d/%d, want 1/0", old.closed, current.closed)
	}
}

func Test_Room_PublicationGenerationRejectsStaleRTPAndCleanup(t *testing.T) {
	r := NewRoom("call1")
	codec := webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}
	first := r.AddPublishedLayer("publisher", "video", "q", "video", codec)
	sink := &closableSink{}
	r.SubscribeLayer("subscriber", "publisher", "video", "q", sink)
	second := r.AddPublishedLayer("publisher", "video", "q", "video", codec)

	if first == second {
		t.Fatal("replacement publication reused generation")
	}
	if sink.closed != 1 {
		t.Fatalf("old downlink closed = %d, want 1", sink.closed)
	}
	replacement := &fakeSink{}
	r.SubscribeLayer("subscriber", "publisher", "video", "q", replacement)
	r.RouteRTPLayerGeneration("publisher", "video", "q", first, &rtp.Packet{})
	r.RouteRTPLayerGeneration("publisher", "video", "q", second, &rtp.Packet{})
	if got := replacement.count(); got != 1 {
		t.Fatalf("replacement packets = %d, want only current generation", got)
	}
	if r.UnpublishLayerGeneration("publisher", "video", "q", first) {
		t.Fatal("stale generation removed replacement publication")
	}
	if !r.HasTrackSubscription("subscriber", "publisher", "video", "q") {
		t.Fatal("stale generation removed replacement downlink")
	}
}

func Test_Room_ResetParticipantMediaPreservesVideoWants(t *testing.T) {
	r := NewRoom("call1")
	r.Join("publisher")
	r.Join("subscriber")
	r.WantVideo("subscriber", "publisher", "q")
	r.AddPublishedLayer("publisher", "video", "q", "video", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8})
	sink := &closableSink{}
	r.SubscribeLayer("subscriber", "publisher", "video", "q", sink)

	r.ResetParticipantMedia("publisher")

	if _, ok := r.VideoWant("subscriber", "publisher"); !ok {
		t.Fatal("publisher media reset deleted subscriber video intent")
	}
	if sink.closed != 1 {
		t.Fatalf("old media sink closed = %d, want 1", sink.closed)
	}
	if got := len(r.Participants()); got != 2 {
		t.Fatalf("participants = %d, want 2", got)
	}
}

func Test_SelectVideoRID_FallsBackToNearestAvailableLayer(t *testing.T) {
	tests := []struct {
		preferred string
		available []string
		want      string
	}{
		{preferred: "q", available: []string{"q", "h", "f"}, want: "q"},
		{preferred: "q", available: []string{"h", "f"}, want: "h"},
		{preferred: "f", available: []string{"q", "h"}, want: "h"},
		{preferred: "h", available: []string{""}, want: ""},
	}
	for _, tt := range tests {
		if got := selectVideoRID(tt.preferred, tt.available); got != tt.want {
			t.Errorf("selectVideoRID(%q, %v) = %q, want %q", tt.preferred, tt.available, got, tt.want)
		}
	}
}

// Test_Room_Leave_CleansUpSubscriptions 离开时双向清理：
// 离开者作为订阅者的下行 sink 被移除；作为发布者其 track 的路由被清空。
func Test_Room_Leave_CleansUpSubscriptions(t *testing.T) {
	r := NewRoom("call1")
	r.Join("a")
	r.Join("b")

	sinkAFromB := &fakeSink{} // a 订阅 b 的轨
	sinkBFromA := &fakeSink{} // b 订阅 a 的轨
	r.Subscribe("a", "b", "b-audio", sinkAFromB)
	r.Subscribe("b", "a", "a-audio", sinkBFromA)

	r.Leave("a")

	// a 作为订阅者：b 的轨不应再投给 a
	r.RouteRTP("b", "b-audio", &rtp.Packet{})
	if got := sinkAFromB.count(); got != 0 {
		t.Errorf("left subscriber 'a' still receiving: got %d, want 0", got)
	}
	// a 作为发布者：a 的轨路由应已清空（b 对 a 的订阅一并移除）
	r.RouteRTP("a", "a-audio", &rtp.Packet{})
	if got := sinkBFromA.count(); got != 0 {
		t.Errorf("left publisher 'a' track still routed: got %d, want 0", got)
	}
}

// Test_Room_PublishedRegistry_ExceptAndCleanup published 轨注册表：
// PublishedExcept 返回除指定 uid 外的已发布轨（供迟到者回填）；Unpublish 与 Leave 正确清理。
func Test_Room_PublishedRegistry_ExceptAndCleanup(t *testing.T) {
	r := NewRoom("c")
	r.Join("a")
	r.Join("b")
	videoCodec := webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264, ClockRate: 90000, SDPFmtpLine: "profile-level-id=42e01f"}
	r.AddPublished("a", "a-vid", "video", videoCodec)
	r.AddPublished("a", "a-aud", "audio", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus})
	r.AddPublished("b", "b-aud", "audio", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus})

	// PublishedExcept("b") 只含 a 的两条
	if got := trackIDs(r.PublishedExcept("b")); !sameSet(got, []string{"a-vid", "a-aud"}) {
		t.Errorf("PublishedExcept(b) = %v, want a-vid,a-aud", got)
	}
	for _, track := range r.PublishedExcept("b") {
		if track.TrackID == "a-vid" && !reflect.DeepEqual(track.Codec, videoCodec) {
			t.Fatalf("video codec = %+v, want %+v", track.Codec, videoCodec)
		}
	}

	// Unpublish 移除单条
	r.Unpublish("a", "a-vid")
	if got := trackIDs(r.PublishedExcept("b")); !sameSet(got, []string{"a-aud"}) {
		t.Errorf("after Unpublish, PublishedExcept(b) = %v, want a-aud", got)
	}

	// Leave 移除该发布者全部已发布轨
	r.Leave("a")
	if got := r.PublishedExcept("b"); len(got) != 0 {
		t.Errorf("after 'a' leaves, PublishedExcept(b) = %v, want empty", got)
	}
}

func Test_Room_SubscriptionRemoval_ClosesDownlinks(t *testing.T) {
	tests := []struct {
		name string
		act  func(r *Room, first, second *closableSink)
		want [2]int
	}{
		{
			name: "publisher leaves",
			act: func(r *Room, _, _ *closableSink) {
				r.Leave("publisher")
			},
			want: [2]int{1, 0},
		},
		{
			name: "subscriber leaves",
			act: func(r *Room, _, _ *closableSink) {
				r.Leave("subscriber")
			},
			want: [2]int{1, 0},
		},
		{
			name: "track ends",
			act: func(r *Room, _, _ *closableSink) {
				r.AddPublished("publisher", "video", "video", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8})
				// AddPublished replaces any existing downlink for this logical publication.
				r.Unpublish("publisher", "video")
			},
			want: [2]int{1, 0},
		},
		{
			name: "subscription replaced",
			act: func(r *Room, _, second *closableSink) {
				r.Subscribe("subscriber", "publisher", "video", second)
			},
			want: [2]int{1, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRoom("call1")
			r.Join("publisher")
			r.Join("subscriber")
			first := &closableSink{}
			second := &closableSink{}
			r.Subscribe("subscriber", "publisher", "video", first)

			tt.act(r, first, second)

			if got := [2]int{first.closed, second.closed}; got != tt.want {
				t.Fatalf("closed counts = %v, want %v", got, tt.want)
			}
		})
	}
}

func trackIDs(pts []PublishedTrack) []string {
	out := make([]string, 0, len(pts))
	for _, p := range pts {
		out = append(out, p.TrackID)
	}
	return out
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := map[string]int{}
	for _, x := range a {
		m[x]++
	}
	for _, x := range b {
		m[x]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}
