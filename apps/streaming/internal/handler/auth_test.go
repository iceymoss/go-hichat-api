package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
)

const testSecret = "iceymoss.hichat.hichat2.com"

func newTestSvc() *svc.ServiceContext {
	var c config.Config
	c.JwtAuth.AccessSecret = testSecret
	c.WebRTC.IceServers = []struct {
		URLs       []string `yaml:"URLs"`
		Username   string   `yaml:"Username,omitempty"`
		Credential string   `yaml:"Credential,omitempty"`
	}{
		{URLs: []string{"stun:stun.l.google.com:19302"}},
		{URLs: []string{"turn:turn.example.com:3478"}, Username: "u", Credential: "p"},
	}
	// 直接构造字面量，避开 NewServiceContext 的 MustNewClient（单测不依赖 etcd/redis）
	return &svc.ServiceContext{Config: c}
}

func mustToken(t *testing.T, uid string) string {
	t.Helper()
	tok, err := ctxdata.GetJwtToken(testSecret, time.Now().Unix(), 3600, uid)
	if err != nil {
		t.Fatalf("build token: %v", err)
	}
	return tok
}

func TestJwtAuth_ParseUID(t *testing.T) {
	auth := NewJwtAuth(newTestSvc())
	valid := mustToken(t, "u123")

	tests := []struct {
		name      string
		setReq    func(r *http.Request)
		wantUID   string
	}{
		{
			name:    "valid token in query",
			setReq:  func(r *http.Request) { r.URL.RawQuery = "token=" + valid },
			wantUID: "u123",
		},
		{
			name:    "valid token in header",
			setReq:  func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+valid) },
			wantUID: "u123",
		},
		{
			name:    "no token",
			setReq:  func(r *http.Request) {},
			wantUID: "",
		},
		{
			name:    "garbage token",
			setReq:  func(r *http.Request) { r.URL.RawQuery = "token=not-a-jwt" },
			wantUID: "",
		},
		{
			name: "wrong secret",
			setReq: func(r *http.Request) {
				bad, _ := ctxdata.GetJwtToken("wrong-secret", time.Now().Unix(), 3600, "u123")
				r.URL.RawQuery = "token=" + bad
			},
			wantUID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/ws", nil)
			tt.setReq(r)
			if got := auth.ParseUID(r); got != tt.wantUID {
				t.Errorf("ParseUID() = %q, want %q", got, tt.wantUID)
			}
		})
	}
}

func TestICEServersHandler(t *testing.T) {
	svcCtx := newTestSvc()
	auth := NewJwtAuth(svcCtx)
	h := ICEServersHandler(svcCtx, auth)

	t.Run("unauthorized without token", func(t *testing.T) {
		w := httptest.NewRecorder()
		h(w, httptest.NewRequest(http.MethodGet, "/v1/streaming/ice-servers", nil))
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", w.Code)
		}
	})

	t.Run("returns ice servers with valid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/v1/streaming/ice-servers?token="+mustToken(t, "u1"), nil)
		h(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		var resp ICEServersResp
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(resp.IceServers) != 2 {
			t.Fatalf("iceServers len = %d, want 2", len(resp.IceServers))
		}
		if resp.IceServers[1].Username != "u" || resp.IceServers[1].Credential != "p" {
			t.Errorf("turn credential not passed through: %+v", resp.IceServers[1])
		}
	})
}
