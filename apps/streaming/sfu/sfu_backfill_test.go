package sfu

import (
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

	// 回填：B 订阅房间内已有发布者（A）
	s.SubscribeExisting("call1", "b")

	select {
	case <-gotRTP:
		// 成功：迟到者 B 经回填 + renegotiation 收到早入房者 A 的媒体
	case <-time.After(15 * time.Second):
		t.Fatal("late joiner B did not receive existing publisher A's media within timeout")
	}
}
