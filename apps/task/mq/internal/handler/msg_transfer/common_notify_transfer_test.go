package msg_transfer

import "testing"

func TestShouldPushCommonNotification(t *testing.T) {
	tests := []struct {
		name        string
		inserted    bool
		alreadyRead bool
		want        bool
	}{
		{name: "new unread notification", inserted: true, want: true},
		{name: "duplicate notification"},
		{name: "inserted after read intent", inserted: true, alreadyRead: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldPushCommonNotification(tt.inserted, tt.alreadyRead); got != tt.want {
				t.Fatalf("shouldPushCommonNotification(%v, %v) = %v, want %v", tt.inserted, tt.alreadyRead, got, tt.want)
			}
		})
	}
}
