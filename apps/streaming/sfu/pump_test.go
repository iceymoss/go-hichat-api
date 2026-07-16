package sfu

import (
	"io"
	"testing"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// fakeReader 假上行 RTP 源：按序吐包，读完返回 io.EOF。
type fakeReader struct {
	pkts []*rtp.Packet
	i    int
}

func (f *fakeReader) ReadRTP() (*rtp.Packet, error) {
	if f.i >= len(f.pkts) {
		return nil, io.EOF
	}
	p := f.pkts[f.i]
	f.i++
	return p, nil
}

// Test_pump_ForwardsAllPacketsUntilEOF 收流泵：把上行源的每个 RTP 包路由进房间，
// 直到源返回错误（EOF/关闭）才退出。
func Test_pump_ForwardsAllPacketsUntilEOF(t *testing.T) {
	r := NewRoom("call1")
	r.Join("a")
	r.Join("b")
	sinkB := &fakeSink{}
	r.Subscribe("b", "a", "a-video", sinkB)

	src := &fakeReader{pkts: []*rtp.Packet{
		{Header: rtp.Header{SequenceNumber: 1}},
		{Header: rtp.Header{SequenceNumber: 2}},
		{Header: rtp.Header{SequenceNumber: 3}},
	}}

	pump(r, "a", "a-video", "", 0, src, 0, defaultActiveSpeakerLimit) // 同步跑到 EOF（生产中在 goroutine 里）

	if got := sinkB.count(); got != 3 {
		t.Errorf("pump forwarded %d packets, want 3", got)
	}
}

func Test_pump_ObservesAudioLevelBeforeRouting(t *testing.T) {
	r := NewRoom("call1")
	r.Join("speaker")
	r.Join("listener")
	r.AddPublished("speaker", "audio", "audio", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus})
	sink := &fakeSink{}
	r.Subscribe("listener", "speaker", "audio", sink)
	pkt := &rtp.Packet{}
	if err := pkt.Header.SetExtension(3, []byte{0x80 | 12}); err != nil {
		t.Fatal(err)
	}

	pump(r, "speaker", "audio", "", 0, &fakeReader{pkts: []*rtp.Packet{pkt}}, 3, defaultActiveSpeakerLimit)

	if got := sink.count(); got != 1 {
		t.Fatalf("forwarded audio packets = %d, want 1", got)
	}
	if got := r.ActiveSpeakers(4); len(got) != 1 || got[0] != "speaker" {
		t.Fatalf("ActiveSpeakers() = %v, want [speaker]", got)
	}
}
