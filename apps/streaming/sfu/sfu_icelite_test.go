package sfu

import (
	"strings"
	"testing"

	"github.com/pion/webrtc/v3"
)

func Test_Peer_Publish_AnswerAdvertisesICELite(t *testing.T) {
	s := NewSFU()
	peer, err := s.AddPeer("call1", "alice", nil)
	if err != nil {
		t.Fatalf("AddPeer: %v", err)
	}
	defer peer.Close()

	client, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatalf("NewPeerConnection: %v", err)
	}
	defer client.Close()
	if _, err := client.CreateDataChannel("probe", nil); err != nil {
		t.Fatalf("CreateDataChannel: %v", err)
	}
	offer, err := client.CreateOffer(nil)
	if err != nil {
		t.Fatalf("CreateOffer: %v", err)
	}
	if err := client.SetLocalDescription(offer); err != nil {
		t.Fatalf("SetLocalDescription: %v", err)
	}
	answer, err := peer.Publish(client.LocalDescription().SDP)
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if !strings.Contains(answer, "a=ice-lite") {
		t.Fatal("SFU answer does not advertise ICE Lite")
	}
}
