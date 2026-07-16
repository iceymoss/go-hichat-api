package sfu

import (
	"testing"
	"time"

	"github.com/pion/webrtc/v3"
)

func Test_Peer_TrickleICE_ConnectsWithoutWaitingForGathering(t *testing.T) {
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

	peer.OnICECandidate(func(candidate webrtc.ICECandidateInit) {
		_ = client.AddICECandidate(candidate)
	})
	client.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			_ = peer.AddICECandidate(candidate.ToJSON())
		}
	})

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
	if err := client.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: answer}); err != nil {
		t.Fatalf("SetRemoteDescription: %v", err)
	}

	connected := make(chan struct{}, 1)
	client.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if state == webrtc.PeerConnectionStateConnected {
			connected <- struct{}{}
		}
	})
	select {
	case <-connected:
	case <-time.After(10 * time.Second):
		t.Fatalf("trickle ICE connection state = %s, want connected", client.ConnectionState())
	}
}
