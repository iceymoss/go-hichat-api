package logic

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"testing"
	"time"
)

// Test_TurnCredential 验证 coturn REST(use-auth-secret) 短期凭证：
// username = "<过期unix>:<uid>"（uid 为空则仅时间戳）；credential = base64(HMAC-SHA1(username, secret))。
func Test_TurnCredential(t *testing.T) {
	secret := "coturn_shared_secret"
	now := time.Unix(1_700_000_000, 0)
	ttl := time.Hour

	tests := []struct {
		name     string
		uid      string
		wantUser string
	}{
		{name: "with uid", uid: "1001", wantUser: "1700003600:1001"},
		{name: "no uid", uid: "", wantUser: "1700003600"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, cred := TurnCredential(secret, tt.uid, ttl, now)
			if user != tt.wantUser {
				t.Errorf("username = %q, want %q", user, tt.wantUser)
			}
			// credential 必须是 base64(HMAC-SHA1(username, secret))
			mac := hmac.New(sha1.New, []byte(secret))
			mac.Write([]byte(user))
			want := base64.StdEncoding.EncodeToString(mac.Sum(nil))
			if cred != want {
				t.Errorf("credential = %q, want %q", cred, want)
			}
			if cred == "" {
				t.Error("credential is empty")
			}
		})
	}
}
