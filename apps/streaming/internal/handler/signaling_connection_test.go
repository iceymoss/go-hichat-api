package handler

import "testing"

type testRoomPeer string

func (p testRoomPeer) RoomID() string { return string(p) }

func Test_SignalingServer_Unregister_ReplacedConnectionKeepsCall(t *testing.T) {
	tests := []struct {
		name       string
		registered *clientConn
		closing    *clientConn
		wantClean  bool
	}{
		{
			name:       "current connection cleans call",
			registered: &clientConn{uid: "u1"},
			wantClean:  true,
		},
		{
			name:       "replaced connection keeps call",
			registered: &clientConn{uid: "u1"},
			closing:    &clientConn{uid: "u1"},
			wantClean:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closing := tt.closing
			if closing == nil {
				closing = tt.registered
			}
			s := &SignalingServer{conns: map[string]*clientConn{"u1": tt.registered}}
			if got := s.unregister(closing); got != tt.wantClean {
				t.Fatalf("unregister() = %v, want %v", got, tt.wantClean)
			}
			_, remains := s.conns["u1"]
			if remains == tt.wantClean {
				t.Fatalf("connection remains = %v, want %v", remains, !tt.wantClean)
			}
		})
	}
}

func TestVideoUnsubscribeAllowed(t *testing.T) {
	tests := []struct {
		name             string
		callID           string
		publisherUID     string
		selfUID          string
		subscriber       testRoomPeer
		subscriberInCall bool
		want             bool
	}{
		{name: "publisher may already be gone", callID: "call-1", publisherUID: "u2", selfUID: "u1", subscriber: "call-1", subscriberInCall: true, want: true},
		{name: "subscriber peer belongs to another call", callID: "call-1", publisherUID: "u2", selfUID: "u1", subscriber: "call-2", subscriberInCall: true},
		{name: "subscriber already left", callID: "call-1", publisherUID: "u2", selfUID: "u1", subscriber: "call-1"},
		{name: "self unsubscribe is invalid", callID: "call-1", publisherUID: "u1", selfUID: "u1", subscriber: "call-1", subscriberInCall: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := videoUnsubscribeAllowed(tt.callID, tt.publisherUID, tt.selfUID, tt.subscriber, tt.subscriberInCall); got != tt.want {
				t.Fatalf("videoUnsubscribeAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
