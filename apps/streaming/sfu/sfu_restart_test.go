package sfu

import (
	"strings"
	"testing"
	"time"
)

func Test_Peer_RequestICERestart_ServerCreatesRestartOffer(t *testing.T) {
	s := NewSFU()
	peer, client, _ := newLoopbackClient(t, s, "call1", "alice")
	defer peer.Close()
	defer client.Close()

	offers := make(chan string, 1)
	peer.OnOffer(func(sdp string) { offers <- sdp })
	before := iceUfrag(peer.pc.LocalDescription().SDP)
	peer.RequestICERestart()

	select {
	case offer := <-offers:
		after := iceUfrag(offer)
		if before == "" || after == "" || before == after {
			t.Fatalf("ICE ufrag before=%q after=%q, want different credentials", before, after)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not create an ICE restart offer")
	}
}

func iceUfrag(sdp string) string {
	for _, line := range strings.Split(sdp, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "a=ice-ufrag:") {
			return strings.TrimPrefix(line, "a=ice-ufrag:")
		}
	}
	return ""
}
