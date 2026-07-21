package sfu

import (
	"testing"
	"time"

	"github.com/pion/webrtc/v3"
)

func Test_SFU_EmitsVideoMediaDiagnostics(t *testing.T) {
	events := make(chan MediaEvent, 4)
	s := NewSFU()
	s.OnMediaEvent(func(event MediaEvent) { events <- event })

	s.emitMediaEvent(MediaEvent{Name: "track_published", PubUID: "publisher", TrackID: "video", Kind: webrtc.RTPCodecTypeVideo.String()})
	s.emitMediaEvent(MediaEvent{Name: "downlink_created", PubUID: "publisher", SubUID: "subscriber", TrackID: "video", Kind: webrtc.RTPCodecTypeVideo.String()})
	s.emitMediaEvent(MediaEvent{Name: "keyframe_requested", PubUID: "publisher", TrackID: "video", Kind: webrtc.RTPCodecTypeVideo.String(), Reason: "downlink_created"})
	s.emitMediaEvent(MediaEvent{Name: "track_ended", PubUID: "publisher", TrackID: "video", Kind: webrtc.RTPCodecTypeVideo.String()})

	want := []string{"track_published", "downlink_created", "keyframe_requested", "track_ended"}
	for _, name := range want {
		if got := <-events; got.Name != name || got.Kind != "video" || got.PubUID != "publisher" {
			t.Fatalf("event = %+v, want name=%s video publisher", got, name)
		}
	}
}

func Test_SFU_ThrottlesSubscriberKeyframeRequests(t *testing.T) {
	s := NewSFU()
	now := time.Unix(100, 0)
	if !s.allowKeyframeRequest("publisher", "video", "subscriber_feedback", now) {
		t.Fatal("first subscriber feedback should request a keyframe")
	}
	if s.allowKeyframeRequest("publisher", "video", "subscriber_feedback", now.Add(100*time.Millisecond)) {
		t.Fatal("feedback inside throttle window should be suppressed")
	}
	if !s.allowKeyframeRequest("publisher", "video", "subscriber_feedback", now.Add(keyframeRequestInterval)) {
		t.Fatal("feedback after throttle window should request a keyframe")
	}
	if !s.allowKeyframeRequest("publisher", "video", "downlink_created", now.Add(110*time.Millisecond)) {
		t.Fatal("new downlink should request a keyframe immediately")
	}
}
