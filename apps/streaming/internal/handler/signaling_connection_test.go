package handler

import "testing"

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
