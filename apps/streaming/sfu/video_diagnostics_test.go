package sfu

import (
	"testing"

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
