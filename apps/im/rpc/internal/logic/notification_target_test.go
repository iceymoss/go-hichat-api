package logic

import "testing"

func TestCanonicalSocialNotificationTarget(t *testing.T) {
	tests := []struct {
		notifyType string
		bizID      string
		want       bool
	}{
		{notifyType: "friend.apply", bizID: "friend:1:apply", want: true},
		{notifyType: "group.invalidated", bizID: "group:18446744073709551615:invalidated", want: true},
		{notifyType: "group.invite", bizID: "group_invite:9:invite", want: true},
		{notifyType: "group.request.resolved", bizID: "group:9:resolved", want: true},
		{notifyType: "group.invite.invalidated", bizID: "group_invite:9:invalidated", want: true},
		{notifyType: "friend.apply", bizID: "friend:1:accept"},
		{notifyType: "group.accept", bizID: "friend:1:accept"},
		{notifyType: "friend.apply", bizID: "friend:01:apply"},
		{notifyType: "friend.apply", bizID: "friend:0:apply"},
		{notifyType: "group.removed", bizID: "group:1:removed"},
	}
	for _, tt := range tests {
		if got := canonicalSocialNotificationTarget(tt.notifyType, tt.bizID); got != tt.want {
			t.Errorf("canonicalSocialNotificationTarget(%q, %q) = %v, want %v", tt.notifyType, tt.bizID, got, tt.want)
		}
	}
}
