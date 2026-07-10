package sfu

import (
	"sort"
	"sync"
	"testing"

	"github.com/pion/rtp"
)

// fakeSink 假下行 sink：记录收到的 RTP 包，供断言 fan-out 是否正确。
type fakeSink struct {
	mu   sync.Mutex
	pkts []*rtp.Packet
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
