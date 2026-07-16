package sfu

import (
	"io"
	"testing"

	"github.com/pion/rtcp"
)

type fakeRTCPReader struct {
	reads [][]rtcp.Packet
	index int
}

func (r *fakeRTCPReader) ReadRTCP() ([]rtcp.Packet, error) {
	if r.index >= len(r.reads) {
		return nil, io.EOF
	}
	packets := r.reads[r.index]
	r.index++
	return packets, nil
}

func Test_DrainRTCP_AllSendersDrainAndVideoRequestsKeyframe(t *testing.T) {
	tests := []struct {
		name    string
		video   bool
		packets [][]rtcp.Packet
		wantPLI int
	}{
		{
			name:    "audio sender drains receiver reports",
			packets: [][]rtcp.Packet{{&rtcp.ReceiverReport{}}, {&rtcp.TransportLayerNack{}}},
		},
		{
			name:  "video sender forwards pli and fir",
			video: true,
			packets: [][]rtcp.Packet{{
				&rtcp.PictureLossIndication{},
				&rtcp.FullIntraRequest{},
				&rtcp.ReceiverReport{},
			}},
			wantPLI: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeRTCPReader{reads: tt.packets}
			pli := 0
			drainRTCP(reader, tt.video, func() { pli++ })
			if reader.index != len(tt.packets) {
				t.Fatalf("reads = %d, want %d", reader.index, len(tt.packets))
			}
			if pli != tt.wantPLI {
				t.Fatalf("PLI requests = %d, want %d", pli, tt.wantPLI)
			}
		})
	}
}
